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

	PromptCardRequest struct {
		UserID              string
		Word                string
		WordLanguage        language.Tag
		TranslationLanguage language.Tag
	}

	PromptCardResponse struct {
		Words []string
	}

	Parameters struct {
		EnrichWordInformationFromDictionary bool
	}

	CreateCardRequest struct {
		UserID              string
		Language            language.Tag
		WordInformationList []entity.WordInformation
		Params              Parameters
	}

	DeleteCardRequest struct {
		UserID string
		CardID string
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

	UpdateCardRequest struct {
		UserID              string
		CardID              string
		WordInformationList []entity.WordInformation
		Params              Parameters
	}

	UpdateCardPerformanceRequest struct {
		UserID         string
		CardID         string
		IsInputCorrect bool
	}

	UpdateCardPerformanceResponse struct {
		NextDueDate time.Time
	}

	GetSentencesRequest struct {
		UserID         string
		Word           string
		SentencesCount int
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
