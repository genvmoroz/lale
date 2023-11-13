package core

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/genvmoroz/lale/service/internal/repo/dictionary"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/pkg/logger"
	"github.com/genvmoroz/lale/service/pkg/speech"
	"github.com/google/uuid"
	"github.com/samber/lo"
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
		GenStory(words []string, lang language.Tag) (string, error)
	}

	Dictionary interface {
		GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error)
	}

	TextToSpeechRepo interface {
		ToSpeech(ctx context.Context, req speech.ToSpeechRequest) ([]byte, error)
	}

	Service interface {
		InspectCard(ctx context.Context, req InspectCardRequest) (entity.Card, error)
		PromptCard(ctx context.Context, req PromptCardRequest) (PromptCardResponse, error)
		CreateCard(ctx context.Context, req CreateCardRequest) (entity.Card, error)
		GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error)
		UpdateCard(ctx context.Context, req UpdateCardRequest) (entity.Card, error)
		UpdateCardPerformance(ctx context.Context, req UpdateCardPerformanceRequest) (UpdateCardPerformanceResponse, error)
		GetCardsToLearn(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error)
		GetCardsToRepeat(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error)
		GetSentences(ctx context.Context, req GetSentencesRequest) (GetSentencesResponse, error)
		GenerateStory(ctx context.Context, req GenerateStoryRequest) (GenerateStoryResponse, error)
		DeleteCard(ctx context.Context, req DeleteCardRequest) (entity.Card, error)
	}

	service struct {
		cardRepo         CardRepo
		sessionRepo      SessionRepo
		aiHelper         AIHelper
		ankiAlgo         AnkiAlgo
		dictionary       Dictionary
		textToSpeechRepo TextToSpeechRepo

		validator validator
	}
)

func NewService(
	cardRepo CardRepo,
	sessionRepo SessionRepo,
	aiHelper AIHelper,
	anki AnkiAlgo,
	dictionary Dictionary,
	textToSpeechRepo TextToSpeechRepo) (Service, error) {
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
		validator:        validator{},
	}, nil
}

func (s *service) InspectCard(ctx context.Context, req InspectCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateInspectCardRequest(req); err != nil {
		return entity.Card{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Word":     req.Word,
			"Request":  "InspectCard",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	card, found := lo.Find(cards,
		func(item entity.Card) bool {
			_, wordFound := lo.Find(item.WordInformationList,
				func(item entity.WordInformation) bool {
					return strings.EqualFold(item.Word, req.Word)
				},
			)
			return wordFound
		},
	)
	if found {
		logger.FromContext(ctx).
			Debug("card found")
		return card, nil
	}

	logger.FromContext(ctx).
		Debug("card not found")
	return entity.Card{}, NewCardNotFoundError().WithWord(req.Word)
}

func (s *service) PromptCard(ctx context.Context, req PromptCardRequest) (PromptCardResponse, error) {
	if err := s.validator.ValidatePromptCardRequest(req); err != nil {
		return PromptCardResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":              req.UserID,
			"WordLanguage":        req.WordLanguage.String(),
			"TranslationLanguage": req.TranslationLanguage.String(),
			"Word":                req.Word,
			"Request":             "PromptCard",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return PromptCardResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("request words from the AI helper")
	wordsWithTranslationMap, err := s.aiHelper.GetFamilyWordsWithTranslation(req.Word, req.TranslationLanguage)
	if err != nil {
		return PromptCardResponse{}, fmt.Errorf("get family words with translation for word (%s): %w", req.Word, err)
	}

	logger.FromContext(ctx).
		Debug("filter not found words out")
	maps.DeleteFunc(wordsWithTranslationMap, s.notFoundInDictionary(req.WordLanguage))

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

func (s *service) CreateCard(ctx context.Context, req CreateCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateCreateCardRequest(req); err != nil {
		return entity.Card{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Words":    extractWords(req.WordInformationList),
			"Request":  "CreateCard",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, logAndReturnError(
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
						return entity.Card{}, NewCardAlreadyExistsError(wordInfo.Word)
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
			return entity.Card{}, logAndReturnError(
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
		return entity.Card{}, fmt.Errorf("enrich card with audio: %w", err)
	}

	logger.
		FromContext(ctx).
		Debug("save card")
	if err = s.cardRepo.SaveCards(ctx, []entity.Card{card}); err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	return card, nil
}

func (s *service) GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Request":  "GetAllCards",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

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
	if err := s.validator.ValidateUpdateCardPerformanceRequest(req); err != nil {
		return UpdateCardPerformanceResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":            req.UserID,
			"CardID":            req.CardID,
			"PerformanceRating": req.PerformanceRating,
			"Request":           "UpdateCardPerformance",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return UpdateCardPerformanceResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return UpdateCardPerformanceResponse{}, logAndReturnError(
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
		return UpdateCardPerformanceResponse{}, NewCardNotFoundError().WithID(req.CardID)
	}

	logger.FromContext(ctx).
		Debug("calculate next due date")
	nextDueDate := s.ankiAlgo.CalculateNextDueDate(req.PerformanceRating, card.GetAnswer(req.PerformanceRating > 2))
	card.NextDueDate = nextDueDate

	logger.FromContext(ctx).
		Debug("save card")
	if err = s.cardRepo.SaveCards(ctx, []entity.Card{*card}); err != nil {
		return UpdateCardPerformanceResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	return UpdateCardPerformanceResponse{
		NextDueDate: nextDueDate,
	}, nil
}

func (s *service) UpdateCard(ctx context.Context, req UpdateCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateUpdateCardRequest(req); err != nil {
		return entity.Card{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":  req.UserID,
			"CardID":  req.CardID,
			"Request": "UpdateCard",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")

	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	card, found := lo.Find(cards,
		func(item entity.Card) bool {
			return item.ID == req.CardID
		},
	)
	if !found {
		logger.FromContext(ctx).
			Debug("card not found")
		return entity.Card{}, NewCardNotFoundError().WithID(req.CardID)
	}

	card.WordInformationList = req.WordInformationList
	if req.Params.EnrichWordInformationFromDictionary {
		err = s.enrichCardFromDictionary(&card)
		if err != nil {
			return entity.Card{}, fmt.Errorf("enrich card words from dictionary: %w", err)
		}
	}

	logger.FromContext(ctx).
		Debug("save card")
	if err = s.cardRepo.SaveCards(ctx, []entity.Card{card}); err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("save card: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	return card, nil
}

func (s *service) GetCardsToLearn(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	return s.getCardsByFilter(
		ctx,
		req,
		"GetCardsToLearn",
		func(card entity.Card) bool {
			return strings.EqualFold(card.Language.String(), req.Language.String()) && card.NeedToLearn()
		},
	)
}

func (s *service) GetCardsToRepeat(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	return s.getCardsByFilter(
		ctx,
		req,
		"GetCardsToRepeat",
		func(card entity.Card) bool {
			return strings.EqualFold(card.Language.String(), req.Language.String()) && card.NeedToRepeat()
		},
	)
}

func (s *service) getCardsByFilter(
	ctx context.Context,
	req GetCardsRequest,
	requestName string,
	predicate func(card entity.Card) bool,
) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Request":  requestName,
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

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
		Debug("filter cards out")
	filtered := lo.Filter[entity.Card](cards,
		func(item entity.Card, _ int) bool {
			return predicate(item)
		},
	)

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    filtered,
	}, nil
}

func (s *service) GetSentences(ctx context.Context, req GetSentencesRequest) (GetSentencesResponse, error) {
	if err := s.validator.ValidateGetSentencesRequest(req); err != nil {
		return GetSentencesResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":         req.UserID,
			"Word":           req.Word,
			"SentencesCount": req.SentencesCount,
			"Request":        "GetSentences",
		},
	)

	logger.FromContext(ctx).
		Debug("get sentences")
	sentences, err := s.generateSentences(req.Word, req.SentencesCount)
	if err != nil {
		return GetSentencesResponse{}, fmt.Errorf("get sentences for word [%s]: %w", req.Word, err)
	}

	return GetSentencesResponse{
		Sentences: sentences,
	}, nil
}

func (s *service) GenerateStory(ctx context.Context, req GenerateStoryRequest) (GenerateStoryResponse, error) {
	if err := s.validator.ValidateGenerateStoryRequest(req); err != nil {
		return GenerateStoryResponse{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Request":  "GenerateStory",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return GenerateStoryResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return GenerateStoryResponse{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	logger.FromContext(ctx).
		Debug("filter cards out by next due date")
	cardsForStory := make([]entity.Card, 0, len(cards))
	for _, card := range cards {
		if strings.EqualFold(card.Language.String(), req.Language.String()) && !reflect.DeepEqual(card.NextDueDate, time.Time{}) {
			cardsForStory = append(cardsForStory, card)
		}
	}

	words := mapCardsToWords(cardsForStory)

	story, err := s.aiHelper.GenStory(lo.Shuffle[string](words), req.Language)
	if err != nil {
		return GenerateStoryResponse{}, fmt.Errorf("generate story: %w", err)
	}

	return GenerateStoryResponse{Story: story}, nil
}

func (s *service) DeleteCard(ctx context.Context, req DeleteCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateDeleteCardRequest(req); err != nil {
		return entity.Card{}, NewRequestValidationError(err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":  req.UserID,
			"CardID":  req.CardID,
			"Request": "DeleteCard",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	logger.FromContext(ctx).
		Debug("get all cards for user")
	cards, err := s.cardRepo.GetCardsForUser(ctx, req.UserID)
	if err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("get cards: %s", err.Error()),
			map[string]interface{}{"UserID": req.UserID},
		)
	}

	card, found := lo.Find(cards,
		func(item entity.Card) bool {
			return item.ID == req.CardID
		},
	)
	if !found {
		logger.FromContext(ctx).
			Debug("card not found")
		return entity.Card{}, NewCardNotFoundError().WithID(req.CardID)
	}

	logger.FromContext(ctx).
		Debug("delete card")
	if err = s.cardRepo.DeleteCard(ctx, req.CardID); err != nil {
		return entity.Card{}, logAndReturnError(
			ctx,
			fmt.Sprintf("delete card: %s", err.Error()),
			map[string]interface{}{
				"UserID": req.UserID,
				"CardID": req.CardID,
			},
		)
	}

	return card, nil
}

func mapCardsToWords(cards []entity.Card) []string {
	return lo.FlatMap[entity.Card, string](
		cards,
		func(item entity.Card, _ int) []string {
			return lo.Map[entity.WordInformation, string](
				item.WordInformationList,
				func(item entity.WordInformation, _ int) string {
					return item.Word
				},
			)
		},
	)
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

func logAndReturnError(ctx context.Context, msg string, fields map[string]any) error {
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

func (s *service) createUserSession(ctx context.Context, userID string) (func(), error) {
	logger.FromContext(ctx).
		Debug("create user session")
	if err := s.sessionRepo.CreateSession(userID); err != nil {
		return nil, logAndReturnError(
			ctx,
			fmt.Sprintf("create user session: %s", err.Error()),
			map[string]interface{}{"UserID": userID},
		)
	}
	return func() {
		logger.FromContext(ctx).
			Debug("close user session")
		if closeErr := s.sessionRepo.CloseSession(userID); closeErr != nil {
			logger.FromContext(ctx).
				Errorf("close user session: %s", closeErr.Error())
		}
	}, nil
}

func (s *service) enrichCardFromDictionary(card *entity.Card) (err error) {
	if card == nil {
		return fmt.Errorf("card is nil")
	}

	card.WordInformationList, err = s.enrichWordInformationListFromDictionary(card.Language, card.WordInformationList)
	if err != nil {
		return fmt.Errorf("enrich word information list from dictionary: %w", err)
	}

	return nil
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

func (s *service) enrichWordInformationListWithAudio(ctx context.Context, _ language.Tag, infoList []entity.WordInformation) error {
	for i := 0; i < len(infoList); i++ {
		audio, err := s.textToAudio(ctx, infoList[i].Word)
		if err != nil {
			return fmt.Errorf("text (%s) to speech: %w", infoList[i].Word, err)
		}
		infoList[i].Audio = audio
	}

	return nil
}

func (s *service) textToAudio(ctx context.Context, text string) ([]byte, error) {
	req := speech.ToSpeechRequest{
		Input: text,
		Voice: speech.VoiceSelectionParams{
			Language:             "en-US",
			Name:                 "en-US-Standard-C",
			PreferredVoiceGender: speech.Female,
		},
		AudioConfig: speech.AudioConfig{AudioEncoding: speech.Mp3},
	}
	return s.textToSpeechRepo.ToSpeech(ctx, req)
}

func (s *service) notFoundInDictionary(lang language.Tag) func(word, _ string) bool {
	return func(word, _ string) bool {
		_, err := s.dictionary.GetWordInformation(word, lang)
		return err != nil && errors.Is(err, dictionary.ErrNotFound)
	}
}

func createContextWithCorrelationLogger(ctx context.Context, fields map[string]any) context.Context {
	return logger.ContextWithLogger(ctx,
		logger.WithCorrelationID().
			WithFields(fields),
	)
}
