package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/genvmoroz/lale/service/internal/repo/dictionary"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/pkg/logger"
	"github.com/genvmoroz/lale/service/pkg/speech"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/text/language"
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

	AIHelper interface {
		GenerateSentences(word string, size uint32) ([]string, error)
		GetFamilyWordsWithTranslation(word string, lang language.Tag) (map[string]string, error)
	}

	Dictionary interface {
		GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error)
	}

	TextToSpeechRepo interface {
		ToSpeech(ctx context.Context, req speech.ToSpeechRequest) ([]byte, error)
	}

	Service interface {
		InspectCard(ctx context.Context, req InspectCardRequest) (InspectCardResponse, error)
		PromptCard(ctx context.Context, req PromptCardRequest) (PromptCardResponse, error)
		CreateCard(ctx context.Context, req CreateCardRequest) (CreateCardResponse, error)
		GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error)
		UpdateCardPerformance(ctx context.Context, req UpdateCardPerformanceRequest) (UpdateCardPerformanceResponse, error)
		GetCardsToReview(ctx context.Context, req GetCardsForReviewRequest) (GetCardsResponse, error)
		GetSentences(ctx context.Context, req GetSentencesRequest) (GetSentencesResponse, error)
		DeleteCard(ctx context.Context, req DeleteCardRequest) (DeleteCardResponse, error)
	}

	service struct {
		cardRepo         CardRepo
		sessionRepo      SessionRepo
		aiHelper         AIHelper
		ankiAlgo         AnkiAlgo
		dictionary       Dictionary
		textToSpeechRepo TextToSpeechRepo
		validator        Validator
	}
)

func NewService(
	cardRepo CardRepo,
	sessionRepo SessionRepo,
	aiHelper AIHelper,
	anki AnkiAlgo,
	dictionary Dictionary,
	textToSpeechRepo TextToSpeechRepo,
	validator Validator) (Service, error) {
	if cardRepo == nil {
		return nil, errors.New("card repo is required")
	}
	if sessionRepo == nil {
		return nil, errors.New("session repo is required")
	}
	if aiHelper == nil {
		return nil, errors.New("aiHelper is required")
	}
	if anki == nil {
		return nil, errors.New("anki algo is required")
	}
	if dictionary == nil {
		return nil, errors.New("dictionary is required")
	}
	if textToSpeechRepo == nil {
		return nil, errors.New("textToSpeechRepo is required")
	}

	return &service{
		cardRepo:         cardRepo,
		sessionRepo:      sessionRepo,
		aiHelper:         aiHelper,
		ankiAlgo:         anki,
		dictionary:       dictionary,
		textToSpeechRepo: textToSpeechRepo,
		validator:        validator,
	}, nil
}

func (s *service) InspectCard(ctx context.Context, req InspectCardRequest) (InspectCardResponse, error) {
	resp := InspectCardResponse{}

	if err := s.validator.ValidateInspectCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":   req.UserID,
					"Language": req.Language.String(),
					"Word":     req.Word,
					"Request":  "InspectCard",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	for _, card := range cards {
		card := card

		if strings.EqualFold(card.Language.String(), req.Language.String()) {
			for _, wordInfo := range card.WordInformationList {
				if strings.EqualFold(wordInfo.Word, req.Word) {
					logger.FromContext(ctx).
						Debug("card found")
					resp.Card = card
					return resp, nil
				}
			}
		}
	}

	logger.FromContext(ctx).
		Debug("card not found")

	return resp, NewCardNotFoundError().WithWord(req.Word)
}

func (s *service) PromptCard(ctx context.Context, req PromptCardRequest) (PromptCardResponse, error) {
	if err := s.validator.ValidatePromptCardRequest(req); err != nil {
		return PromptCardResponse{}, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":   req.UserID,
					"Language": req.Language.String(),
					"Word":     req.Word,
					"Request":  "PromptCard",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return PromptCardResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("request words from the ai helper")
	wordsWithTranslationMap, err := s.aiHelper.GetFamilyWordsWithTranslation(req.Word, req.Language)
	if err != nil {
		return PromptCardResponse{}, fmt.Errorf("get family words with translation for word (%s): %w", req.Word, err)
	}

	logger.FromContext(ctx).
		Debug("filter not found words out")
	maps.DeleteFunc(wordsWithTranslationMap, s.notFoundInDictionary(req.Language))

	wordsWithTranslationSlice := lo.MapToSlice[string, string](
		wordsWithTranslationMap,
		func(key string, value string) string {
			return fmt.Sprintf("%s - %s", key, value)
		},
	)

	return PromptCardResponse{
		Words: wordsWithTranslationSlice,
	}, nil
}

func (s *service) notFoundInDictionary(lang language.Tag) func(word, _ string) bool {
	return func(word, _ string) bool {
		_, err := s.dictionary.GetWordInformation(word, lang)
		return err != nil && errors.Is(err, dictionary.ErrNotFound)
	}
}

func (s *service) CreateCard(ctx context.Context, req CreateCardRequest) (CreateCardResponse, error) {
	resp := CreateCardResponse{}

	if err := s.validator.ValidateCreateCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":   req.UserID,
					"Language": req.Language.String(),
					"Words":    extractWords(req.WordInformationList),
					"Request":  "CreateCard",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	for _, wordInfo := range req.WordInformationList {
		for _, card := range cards {
			if strings.EqualFold(card.Language.String(), req.Language.String()) {
				for _, val := range card.WordInformationList {
					if strings.EqualFold(val.Word, wordInfo.Word) {
						logger.FromContext(ctx).
							Debug("card already exists")
						return resp, NewCardAlreadyExistsError(wordInfo.Word)
					}
				}
			}
		}
	}

	card := entity.Card{
		ID:                  uuid.NewString(),
		UserID:              req.UserID,
		Language:            req.Language,
		WordInformationList: req.WordInformationList,
	}

	if req.Params.EnrichWordInformationFromDictionary {
		logger.FromContext(ctx).
			Debug("enrich card with info from dictionary")
		enrichedWordInformationList, err := s.enrichWordInformationListFromDictionary(card.Language, card.WordInformationList)
		if err != nil {
			return resp, logAndReturnError(
				ctx,
				fmt.Sprintf("get words from dictionary: %s", err.Error()),
				map[string]interface{}{"UserID": req.UserID},
			)
		}
		card.WordInformationList = enrichedWordInformationList
	}

	logger.FromContext(ctx).
		Debug("enrich card with audio")
	if err = s.enrichWordInformationListWithAudio(ctx, card.Language, card.WordInformationList); err != nil {
		return CreateCardResponse{}, fmt.Errorf("enrich card with audio: %w", err)
	}

	logger.
		FromContext(ctx).
		Debug("save card")
	if err = s.cardRepo.SaveCards(ctx, []entity.Card{card}); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	resp.Card = card

	return resp, nil
}

func (s *service) enrichWordInformationListFromDictionary(language language.Tag, wordInformationLists []entity.WordInformation) ([]entity.WordInformation, error) {
	var enrichedWords []entity.WordInformation
	for _, wordInfo := range wordInformationLists {
		enrichedWordInfo, err := s.dictionary.GetWordInformation(wordInfo.Word, language)
		if err != nil {
			return nil, fmt.Errorf("request word [%s] from dictionary: %w", wordInfo.Word, err)
		}

		enrichedWordInfo.Word = wordInfo.Word
		enrichedWordInfo.Translation = wordInfo.Translation

		enrichedWords = append(enrichedWords, enrichedWordInfo)
	}

	return enrichedWords, nil
}

func (s *service) enrichWordInformationListWithAudio(ctx context.Context, lang language.Tag, infoList []entity.WordInformation) error {
	for i := 0; i < len(infoList); i++ {
		if err := s.textToSpeech(ctx, lang, &infoList[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) textToSpeech(ctx context.Context, lang language.Tag, info *entity.WordInformation) error {
	if info == nil {
		return nil
	}

	req := speech.ToSpeechRequest{
		Input: info.Word,
		Voice: speech.VoiceSelectionParams{
			Language:             lang,
			Name:                 "en-US-Standard-C",
			PreferredVoiceGender: speech.Female,
		},
		AudioConfig: speech.AudioConfig{AudioEncoding: speech.Mp3},
	}
	audio, err := s.textToSpeechRepo.ToSpeech(ctx, req)
	if err != nil {
		return fmt.Errorf("text to speech: %w", err)
	}

	info.Audio = audio

	return nil
}

func (s *service) GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":   req.UserID,
					"Language": req.Language.String(),
					"Request":  "GetAllCards",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return GetCardsResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				WithField("UserID", req.UserID).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	apiCards := make([]entity.Card, 0, len(cards))

	logger.FromContext(ctx).
		Debug("filter cards out by language")
	for _, card := range cards {
		if len(strings.TrimSpace(req.Language.String())) == 0 || strings.EqualFold(card.Language.String(), req.Language.String()) {
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

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":            req.UserID,
					"CardID":            req.CardID,
					"PerformanceRating": req.PerformanceRating,
					"Request":           "UpdateCardPerformance",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	var card *entity.Card
	for _, c := range cards {
		c := c
		if c.ID == req.CardID {
			logger.FromContext(ctx).
				Debug("card found")
			card = &c
			break
		}
	}

	if card == nil {
		logger.FromContext(ctx).
			Debug("card not found")
		return resp, NewCardNotFoundError().WithID(req.CardID)
	}

	logger.FromContext(ctx).
		Debug("calculate next due date")
	nextDueDate := s.ankiAlgo.CalculateNextDueDate(req.PerformanceRating, card.GetAnswer(req.PerformanceRating > 2))
	card.NextDueDate = nextDueDate

	logger.FromContext(ctx).
		Debug("save card")
	if err = s.cardRepo.SaveCards(ctx, []entity.Card{*card}); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("save card: %s", err.Error()),
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

	ctx = logger.ContextWithLogger(ctx,
		logger.FromContext(ctx).
			WithFields(
				logrus.Fields{
					"UserID":   req.UserID,
					"Language": req.Language.String(),
					"Request":  "GetCardsToReview",
				},
			),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return GetCardsResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	logger.FromContext(ctx).
		Debug("filter cards out by next due date")
	cardsToReview := make([]entity.Card, 0, len(cards))
	for _, card := range cards {
		if strings.EqualFold(card.Language.String(), req.Language.String()) && card.NeedToReview() {
			cardsToReview = append(cardsToReview, card)
		}
	}

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    cardsToReview,
	}, nil
}

func (s *service) GetSentences(ctx context.Context, req GetSentencesRequest) (GetSentencesResponse, error) {
	if err := s.validator.ValidateGetSentencesRequest(req); err != nil {
		return GetSentencesResponse{}, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx, logger.FromContext(ctx).
		WithFields(
			logrus.Fields{
				"UserID":         req.UserID,
				"Word":           req.Word,
				"SentencesCount": req.SentencesCount,
				"Request":        "GetSentences",
			},
		),
	)

	logger.FromContext(ctx).
		Debug("get sentences")
	sentences, err := s.generateSentences(req.Word, req.SentencesCount)
	if err != nil {
		return GetSentencesResponse{}, fmt.Errorf("get sentences for word [%s]: %w", req.Word, err)
	}

	return GetSentencesResponse{Sentences: sentences}, nil
}

func (s *service) generateSentences(word string, size uint32) ([]string, error) {
	if len(strings.TrimSpace(word)) == 0 {
		return nil, nil
	}

	sentences, err := s.aiHelper.GenerateSentences(word, size)
	if err != nil {
		return nil, fmt.Errorf("generate sentences for word [%s]: %w", word, err)
	}

	return sentences, nil
}

func (s *service) DeleteCard(ctx context.Context, req DeleteCardRequest) (DeleteCardResponse, error) {
	resp := DeleteCardResponse{}

	if err := s.validator.ValidateDeleteCardRequest(req); err != nil {
		return resp, NewRequestValidationError(err)
	}

	ctx = logger.ContextWithLogger(ctx, logger.FromContext(ctx).
		WithFields(
			logrus.Fields{
				"UserID":  req.UserID,
				"CardID":  req.CardID,
				"Request": "DeleteCard",
			},
		),
	)

	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(req.UserID); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}
	defer func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(req.UserID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	var card *entity.Card
	for _, c := range cards {
		c := c
		if c.ID == req.CardID {
			logger.FromContext(ctx).
				Debug("card found")
			card = &c
			break
		}
	}

	if card == nil {
		logger.FromContext(ctx).
			Debug("card not found")
		return resp, NewCardNotFoundError().WithID(req.CardID)
	}

	logger.FromContext(ctx).
		Debug("delete card")
	if err = s.cardRepo.DeleteCard(ctx, req.CardID); err != nil {
		return resp, logAndReturnError(
			ctx,
			fmt.Sprintf("delete card: %s", err.Error()),
			map[string]interface{}{
				"UserID": req.UserID,
				"CardID": req.CardID,
			},
		)
	}

	resp.Card = *card

	return resp, nil
}

func logAndReturnError(ctx context.Context, msg string, fields logrus.Fields) error {
	logger.FromContext(ctx).
		WithFields(fields).
		Errorf(msg)

	return errors.New(msg)
}

func extractWords(list []entity.WordInformation) []string {
	words := make([]string, len(list))

	for index, info := range list {
		words[index] = info.Word
	}

	return words
}
