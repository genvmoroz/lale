//nolint:copyloopvar // it's ok
package core

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
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
		CalculateNextDueDate(performance uint32, consecutiveCorrectAnswersNumber uint32) time.Time
	}

	AIHelper interface {
		GenerateSentences(word string, size int) ([]string, error)
		GetFamilyWordsWithTranslation(word string, lang language.Tag) (map[string]string, error)
		GenStory(words []string, lang language.Tag) (string, error)
	}

	Dictionary interface {
		GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error)
	}

	TextToSpeechRepo interface {
		ToSpeech(ctx context.Context, req speech.ToSpeechRequest) ([]byte, error)
	}

	Service struct {
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
	textToSpeechRepo TextToSpeechRepo,
) (*Service, error) {
	if lo.IsNil(cardRepo) {
		return nil, errors.New("card repo is required")
	}
	if lo.IsNil(sessionRepo) {
		return nil, errors.New("session repo is required")
	}
	if lo.IsNil(aiHelper) {
		return nil, errors.New("aiHelper is required")
	}
	if lo.IsNil(anki) {
		return nil, errors.New("anki algo is required")
	}
	if lo.IsNil(dictionary) {
		return nil, errors.New("dictionary is required")
	}
	if lo.IsNil(textToSpeechRepo) {
		return nil, errors.New("textToSpeechRepo is required")
	}

	return &Service{
		cardRepo:         cardRepo,
		sessionRepo:      sessionRepo,
		aiHelper:         aiHelper,
		ankiAlgo:         anki,
		dictionary:       dictionary,
		textToSpeechRepo: textToSpeechRepo,
		validator:        validator{},
	}, nil
}

// TODO: add ability to change speech voice, it's going to be an additional endpoint,
// the client will be able to choose the voice from the list of available voices.
// Create a new collection with words and their pronunciation.
// Language + Word - different voices - audio data.

func (s *Service) InspectCard(ctx context.Context, req InspectCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateInspectCardRequest(req); err != nil {
		return entity.Card{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
	return entity.Card{}, fmt.Errorf("%w: word %s", NewNotFoundError(), req.Word)
}

func (s *Service) PromptCard(ctx context.Context, req PromptCardRequest) (PromptCardResponse, error) {
	if err := s.validator.ValidatePromptCardRequest(req); err != nil {
		return PromptCardResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
		return PromptCardResponse{}, fmt.Errorf(
			"get family words with translation for word (%s): %w",
			req.Word, err,
		)
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

func (s *Service) CreateCard(ctx context.Context, req CreateCardRequest) (entity.Card, error) { //nolint:gocognit,lll // it's ok
	if err := s.validator.ValidateCreateCardRequest(req); err != nil {
		return entity.Card{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
						return entity.Card{}, fmt.Errorf("%w: word %s", NewAlreadyExistsError(), wordInfo.Word)
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
		var enriched []entity.WordInformation
		enriched, err = s.enrichWordInformationListFromDictionary(card.Language, card.WordInformationList)
		if err != nil {
			return entity.Card{}, logAndReturnError(
				ctx,
				fmt.Sprintf("get words from dictionary: %s", err.Error()),
				map[string]interface{}{"UserID": req.UserID},
			)
		}
		card.WordInformationList = enriched
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

func (s *Service) GetAllCards(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
		if len(strings.TrimSpace(req.Language.String())) == 0 ||
			strings.EqualFold(card.Language.String(), req.Language.String()) {
			apiCards = append(apiCards, card)
		}
	}

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    apiCards,
	}, nil
}

func (s *Service) UpdateCardPerformance(
	ctx context.Context,
	req UpdateCardPerformanceRequest,
) (UpdateCardPerformanceResponse, error) {
	if err := s.validator.ValidateUpdateCardPerformanceRequest(req); err != nil {
		return UpdateCardPerformanceResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
	}

	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":         req.UserID,
			"CardID":         req.CardID,
			"IsInputCorrect": req.IsInputCorrect,
			"Request":        "UpdateCardPerformance",
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
		return UpdateCardPerformanceResponse{}, fmt.Errorf("%w: card ID %s", NewNotFoundError(), req.CardID)
	}

	logger.FromContext(ctx).
		Debug("calculate next due date")

	card.AddAnswer(req.IsInputCorrect)
	consecutiveCorrectAnswersNumber := card.GetConsecutiveCorrectAnswersNumber()

	// todo: remove this line after testing
	// const performanceDivider = 2
	// performance := consecutiveCorrectAnswersNumber / performanceDivider
	performance := consecutiveCorrectAnswersNumber
	if performance > MaxAllowedPerformanceRating {
		performance = MaxAllowedPerformanceRating
	}
	nextDueDate := s.ankiAlgo.CalculateNextDueDate(performance, consecutiveCorrectAnswersNumber)
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

func (s *Service) UpdateCard(ctx context.Context, req UpdateCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateUpdateCardRequest(req); err != nil {
		return entity.Card{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
		return entity.Card{}, fmt.Errorf("%w: card ID %s", NewNotFoundError(), req.CardID)
	}

	card.WordInformationList = req.WordInformationList
	if req.Params.EnrichWordInformationFromDictionary {
		err = s.enrichCardFromDictionary(&card)
		if err != nil {
			return entity.Card{}, fmt.Errorf("enrich card words from dictionary: %w", err)
		}
	}

	err = s.enrichCardWithAudio(ctx, &card)
	if err != nil {
		return entity.Card{}, fmt.Errorf("enrich card with audio: %w", err)
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

func (s *Service) GetCardsToLearn(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Request":  "GetCardsToLearn",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	toLearnFilter := func(card entity.Card) bool {
		return strings.EqualFold(card.Language.String(), req.Language.String()) && card.NeedToLearn()
	}

	resp, err := s.getCardsByFilter(ctx, req, toLearnFilter)
	if err != nil {
		return GetCardsResponse{}, err
	}

	var newCards []entity.Card
	for _, card := range slices.Backward(resp.Cards) {
		newCards = append(newCards, card)
	}
	resp.Cards = newCards

	return resp, nil
}

func (s *Service) GetCardsToRepeat(ctx context.Context, req GetCardsRequest) (GetCardsResponse, error) {
	ctx = createContextWithCorrelationLogger(ctx,
		map[string]any{
			"UserID":   req.UserID,
			"Language": req.Language.String(),
			"Request":  "GetCardsToRepeat",
		},
	)

	closeSession, err := s.createUserSession(ctx, req.UserID)
	if err != nil {
		return GetCardsResponse{}, fmt.Errorf("create user session: %w", err)
	}
	defer closeSession()

	toRepeatFilter := func(card entity.Card) bool {
		return strings.EqualFold(card.Language.String(), req.Language.String()) && card.NeedToRepeat()
	}

	resp, err := s.getCardsByFilter(ctx, req, toRepeatFilter)
	if err != nil {
		return GetCardsResponse{}, err
	}

	sortByConsecutiveCorrectAnswersAndShuffleInChunks(resp.Cards, 5) //nolint:mnd // it's ok, will be removed later

	return resp, nil
}

func (s *Service) getCardsByFilter(
	ctx context.Context, req GetCardsRequest, filter func(card entity.Card) bool,
) (GetCardsResponse, error) {
	if err := s.validator.ValidateGetCardsRequest(req); err != nil {
		return GetCardsResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
	}

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

	// todo: add iter package here to iterate over cards with a custom iterator
	logger.FromContext(ctx).
		Debug("filter cards out")
	filtered := lo.Filter[entity.Card](cards,
		func(item entity.Card, _ int) bool {
			return filter(item)
		},
	)

	return GetCardsResponse{
		UserID:   req.UserID,
		Language: req.Language,
		Cards:    filtered,
	}, nil
}

func (s *Service) GetSentences(ctx context.Context, req GetSentencesRequest) (GetSentencesResponse, error) {
	if err := s.validator.ValidateGetSentencesRequest(req); err != nil {
		return GetSentencesResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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

func (s *Service) GenerateStory(ctx context.Context, req GenerateStoryRequest) (GenerateStoryResponse, error) {
	if err := s.validator.ValidateGenerateStoryRequest(req); err != nil {
		return GenerateStoryResponse{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
		if strings.EqualFold(card.Language.String(), req.Language.String()) &&
			!reflect.DeepEqual(card.NextDueDate, time.Time{}) {
			cardsForStory = append(cardsForStory, card)
		}
	}

	words := mapCardsToWords(cardsForStory)

	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	story, err := s.aiHelper.GenStory(words, req.Language)
	if err != nil {
		return GenerateStoryResponse{}, fmt.Errorf("generate story: %w", err)
	}

	return GenerateStoryResponse{Story: story}, nil
}

func (s *Service) DeleteCard(ctx context.Context, req DeleteCardRequest) (entity.Card, error) {
	if err := s.validator.ValidateDeleteCardRequest(req); err != nil {
		return entity.Card{}, fmt.Errorf("%w: %w", NewValidationError(), err)
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
		return entity.Card{}, fmt.Errorf("%w: card ID %s", NewNotFoundError(), req.CardID)
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

func (s *Service) generateSentences(word string, size int) ([]string, error) {
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
		Error(msg)

	return errors.New(msg)
}

func extractWords(list []entity.WordInformation) []string {
	words := make([]string, len(list))

	for index, info := range list {
		words[index] = info.Word
	}

	return words
}

func (s *Service) createUserSession(ctx context.Context, userID string) (func(), error) {
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

func (s *Service) enrichCardFromDictionary(card *entity.Card) error {
	if card == nil {
		return fmt.Errorf("card is nil")
	}

	enriched, err := s.enrichWordInformationListFromDictionary(card.Language, card.WordInformationList)
	if err != nil {
		return fmt.Errorf("enrich word information list from dictionary: %w", err)
	}

	card.WordInformationList = enriched

	return nil
}

func (s *Service) enrichWordInformationListFromDictionary(
	language language.Tag,
	wordInformationLists []entity.WordInformation,
) ([]entity.WordInformation, error) {
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

func (s *Service) enrichCardWithAudio(ctx context.Context, card *entity.Card) error {
	if card == nil {
		return fmt.Errorf("card is nil")
	}

	err := s.enrichWordInformationListWithAudio(ctx, card.Language, card.WordInformationList)
	if err != nil {
		return fmt.Errorf("enrich words with audio: %w", err)
	}

	return nil
}

func (s *Service) enrichWordInformationListWithAudio(
	ctx context.Context,
	_ language.Tag,
	infoList []entity.WordInformation,
) error {
	for i := 0; i < len(infoList); i++ {
		audio, err := s.textToAudio(ctx, infoList[i].Word)
		if err != nil {
			return fmt.Errorf("text (%s) to speech: %w", infoList[i].Word, err)
		}
		infoList[i].Audio = audio
	}

	return nil
}

func (s *Service) textToAudio(ctx context.Context, text string) ([]byte, error) {
	req := speech.ToSpeechRequest{
		Input: text,
		Voice: speech.VoiceSelectionParams{
			Language:             "en-GB",
			Name:                 "en-GB-Standard-C",
			PreferredVoiceGender: speech.Female,
		},
		AudioConfig: speech.AudioConfig{AudioEncoding: speech.Mp3},
	}
	return s.textToSpeechRepo.ToSpeech(ctx, req)
}

func (s *Service) notFoundInDictionary(lang language.Tag) func(word, _ string) bool {
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

// sortByConsecutiveCorrectAnswersAndShuffleInChunks sorts cards by consecutive correct answers number (desc)
// and shuffles items within consecutive chunks of size chunkSize.
// todo: reimplement this to determine the chunk size based on the similarity of the consecutive correct answers numberю
func sortByConsecutiveCorrectAnswersAndShuffleInChunks(cards []entity.Card, chunkSize uint8) {
	if chunkSize < 1 {
		chunkSize = 1
	}

	slices.SortFunc(cards, func(a, b entity.Card) int {
		// Descending by ConsecutiveCorrectAnswersNumber
		return cmp.Compare(b.ConsecutiveCorrectAnswersNumber, a.ConsecutiveCorrectAnswersNumber)
	})

	for chunk := range slices.Chunk(cards, int(chunkSize)) {
		if len(chunk) <= 1 {
			continue
		}

		rand.Shuffle(len(chunk), func(i, j int) {
			chunk[i], chunk[j] = chunk[j], chunk[i]
		})
	}
}
