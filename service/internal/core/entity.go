package core

import (
	"time"

	"github.com/genvmoroz/lale/service/pkg/entity"
	"golang.org/x/text/language"
)

type (
	InspectCardRequest struct {
		UserID   string
		Language language.Tag
		Word     string
	}

	InspectCardResponse struct {
		Card entity.Card
	}

	PromptCardRequest struct {
		UserID              string
		Word                string
		WordLanguage        language.Tag
		TranslationLanguage language.Tag
	}

	PromptCardResponse struct {
		Words []string
	}

	CreateCardParameters struct {
		EnrichWordInformationFromDictionary bool
	}

	CreateCardRequest struct {
		UserID              string
		Language            language.Tag
		WordInformationList []entity.WordInformation
		Params              CreateCardParameters
	}

	CreateCardResponse struct {
		Card entity.Card
	}

	DeleteCardRequest struct {
		UserID string
		CardID string
	}

	DeleteCardResponse struct {
		Card entity.Card
	}

	GetCardsRequest struct {
		UserID   string
		Language language.Tag
	}

	GetCardsResponse struct {
		UserID   string
		Language language.Tag
		Cards    []entity.Card
	}

	UpdateCardPerformanceRequest struct {
		UserID            string
		CardID            string
		PerformanceRating uint32
	}

	UpdateCardPerformanceResponse struct {
		NextDueDate time.Time
	}

	GetSentencesRequest struct {
		UserID         string
		Word           string
		SentencesCount uint32
	}

	GetSentencesResponse struct {
		Sentences []string
	}

	GenerateStoryRequest struct {
		UserID   string
		Language language.Tag
	}

	GenerateStoryResponse struct {
		Story string
	}
)
