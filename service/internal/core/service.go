package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/genvmoroz/lale/service/internal/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	CardRepo interface {
		GetCardsForUser(ctx context.Context, userID string) ([]entity.Card, error)
		SaveCards(ctx context.Context, cards []entity.Card) error
		DeleteCard(ctx context.Context, cardID string) error
	}

	SessionRepo interface {
		CreateSession(userID string) error
		CloseSession(userID string) error
	}

	AnkiAlgo interface {
		CalculateNextDueDate(uint32, uint32) time.Time
	}

	SentenceScraper interface {
		ScrapeSentences(word string, size uint32) ([]string, error)
	}

	Dictionary interface {
		GetWordInformation(language lang.Language, word string) (entity.WordInformation, error)
	}

	Service interface {
		InspectCard(ctx context.Context, req InspectCardRequest) (InspectCardResponse, error)
		CreateCard(ctx context.Context, req CreateCardRequest) (CreateCardResponse, error)
		GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error)
		UpdateCardPerformance(ctx context.Context, req UpdateCardPerformanceRequest) (UpdateCardPerformanceResponse, error)
		GetCardsToReview(ctx context.Context, req GetCardsForReviewRequest) (GetCardsResponse, error)
		DeleteCard(ctx context.Context, req DeleteCardRequest) (DeleteCardResponse, error)
	}

	service struct {
		cardRepo         CardRepo
		sessionRepo      SessionRepo
		sentenceScrapers []SentenceScraper
		ankiAlgo         AnkiAlgo
		dictionary       Dictionary
		validator        Validator
	}
)

func NewService(
	cardRepo CardRepo,
	sessionRepo SessionRepo,
	sentenceScrapers []SentenceScraper,
	anki AnkiAlgo,
	dictionary Dictionary,
	validator Validator) (Service, error) {
	if cardRepo == nil {
		return nil, errors.New("card repo is required")
	}
	if sessionRepo == nil {
		return nil, errors.New("session repo is required")
	}
	if len(sentenceScrapers) == 0 {
		return nil, errors.New("sentence scrapers are required")
	}
	if anki == nil {
		return nil, errors.New("anki algo is required")
	}
	if dictionary == nil {
		return nil, errors.New("dictionary is required")
	}

	return &service{
		cardRepo:         cardRepo,
		sessionRepo:      sessionRepo,
		sentenceScrapers: sentenceScrapers,
		ankiAlgo:         anki,
		dictionary:       dictionary,
		validator:        validator,
	}, nil
}

func (s *service) InspectCard(ctx context.Context, req InspectCardRequest) (InspectCardResponse, error) {
	resp := InspectCardResponse{}

	if err := s.validator.ValidateInspectCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	for _, card := range cards {
		if card.Language.Equal(req.Language) {
			for _, wordInfo := range card.WordInformationList {
				if strings.EqualFold(wordInfo.Word, req.Word) {

					resp.Card = card
					return resp, nil
				}
			}
		}
	}

	return resp, NewCardNotFoundError().WithWord(req.Word)
}

func (s *service) CreateCard(ctx context.Context, req CreateCardRequest) (CreateCardResponse, error) {
	resp := CreateCardResponse{}

	if err := s.validator.ValidateCreateCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	for _, wordInfo := range req.WordInformationList {
		for _, card := range cards {
			if card.Language.Equal(req.Language) {
				for _, val := range card.WordInformationList {
					if strings.EqualFold(val.Word, wordInfo.Word) {
						return resp, NewCardAlreadyExistsError(wordInfo.Word)
					}
				}
			}
		}
	}

	card := entity.Card{
		ID:       uuid.NewString(),
		UserID:   req.UserID,
		Language: req.Language,
	}

	if req.Params.EnrichWordInformationFromDictionary {
		enrichedWordInformationList, err := s.enrichWordInformationListFromDictionary(req.Language, req.WordInformationList)
		if err != nil {
			return resp, logAndReturnError(
				fmt.Sprintf("failed to get words from dictionary: %s", err.Error()),
				map[string]interface{}{"UserID": req.UserID},
			)
		}
		card.WordInformationList = enrichedWordInformationList
	}

	if err = s.cardRepo.SaveCards(ctx, []entity.Card{card}); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	resp.Card = card

	return resp, nil
}

func (s *service) enrichWordInformationListFromDictionary(language lang.Language, wordInformationLists []entity.WordInformation) ([]entity.WordInformation, error) {
	var enrichedWords []entity.WordInformation
	for _, wordInfo := range wordInformationLists {
		enrichedWordInfo, err := s.dictionary.GetWordInformation(language, wordInfo.Word)
		if err != nil {
			return nil, fmt.Errorf("failed to request word [%s] from dictionary: %w", wordInfo.Word, err)
		}

		enrichedWordInfo.Word = wordInfo.Word
		enrichedWordInfo.Translation = wordInfo.Translation

		enrichedWords = append(enrichedWords, enrichedWordInfo)
	}

	return enrichedWords, nil
}

func (s *service) GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return GetCardsResponse{}, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	apiCards := make([]entity.Card, 0, len(cards))

	for _, card := range cards {
		if len(strings.TrimSpace(req.Language.String())) == 0 || req.Language.Equal(card.Language) {
			apiCards = append(apiCards, card)
		}
	}

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    apiCards,
	}, nil
}

func (s *service) UpdateCardPerformance(ctx context.Context, req UpdateCardPerformanceRequest) (UpdateCardPerformanceResponse, error) {
	resp := UpdateCardPerformanceResponse{}

	if err := s.validator.ValidateUpdateCardPerformanceRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	var card *entity.Card
	for _, c := range cards {
		if c.ID == req.CardID {
			card = &c
		}
	}

	if card == nil {
		return resp, NewCardNotFoundError().WithID(req.CardID)
	}

	nextDueDate := s.ankiAlgo.CalculateNextDueDate(req.PerformanceRating, card.GetAnswer(req.PerformanceRating > 2))
	card.NextDueDate = nextDueDate

	if err = s.cardRepo.SaveCards(ctx, []entity.Card{*card}); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	resp.NextDueDate = nextDueDate

	return resp, nil
}

func (s *service) GetCardsToReview(ctx context.Context, req GetCardsForReviewRequest) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsForReviewRequest(req); err != nil {
		return GetCardsResponse{}, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return GetCardsResponse{}, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	cardsToReview := make([]entity.Card, 0, len(cards))
	for _, card := range cards {
		if card.Language.Equal(req.Language) && card.NeedToReview() {
			if err = s.enrichCardWithSentences(&card, req.SentencesCount); err != nil {
				return GetCardsResponse{}, logAndReturnError(
					fmt.Sprintf("failed to enrich card with sentences: %s", err.Error()),
					map[string]interface{}{"UserID": req.UserID, "CardID": card.ID},
				)
			}

			cardsToReview = append(cardsToReview, card)
		}
	}

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    cardsToReview,
	}, nil
}

func (s *service) enrichCardWithSentences(card *entity.Card, size uint32) error {
	if card == nil {
		return nil
	}

	for i := 0; i < len(card.WordInformationList); i++ {
		if err := s.enrichWordInformationWithSentences(&card.WordInformationList[i], size); err != nil {
			return fmt.Errorf("failed to scrape sentences for word [%s]: %w", card.WordInformationList[i].Word, err)
		}
	}

	return nil
}

func (s *service) enrichWordInformationWithSentences(wordInfo *entity.WordInformation, size uint32) error {
	if wordInfo == nil {
		return nil
	}

	for sizeLeft, scraperIndex := size, 0; sizeLeft > 0 && scraperIndex < len(s.sentenceScrapers); scraperIndex++ {
		sentences, err := s.sentenceScrapers[scraperIndex].ScrapeSentences(wordInfo.Word, sizeLeft)
		if err != nil {
			return fmt.Errorf("failed to scrape sentences for word [%s]: %w", wordInfo.Word, err)
		}
		wordInfo.Sentences = sentences
		sizeLeft -= uint32(len(sentences))
	}

	return nil
}

func (s *service) DeleteCard(ctx context.Context, req DeleteCardRequest) (DeleteCardResponse, error) {
	resp := DeleteCardResponse{}

	if err := s.validator.ValidateDeleteCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logrus.
				WithField("UserID", req.UserID).
				Errorf("failed to close user session: %s", closeErr.Error())
		}
	}()

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	var card *entity.Card
	for _, c := range cards {
		if c.ID == req.CardID {
			card = &c
		}
	}

	if card == nil {
		return resp, NewCardNotFoundError().WithID(req.CardID)
	}

	if err = s.cardRepo.DeleteCard(ctx, req.CardID); err != nil {
		return resp, logAndReturnError(
			fmt.Sprintf("failed to delete card: %s", err.Error()),
			map[string]interface{}{
				"UserID": req.UserID,
				"CardID": req.CardID,
			},
		)
	}

	resp.Card = *card

	return resp, nil
}

func logAndReturnError(msg string, fields logrus.Fields) error {
	logrus.
		WithFields(fields).
		Errorf(msg)

	return errors.New(msg)
}
