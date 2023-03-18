package core

import (
	"time"

	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	InspectCardRequest struct {
		UserID         string
		Language       lang.Language
		Word           string
		SentencesCount uint32
	}

	InspectCardResponse struct {
		Card entity.Card
	}

	CreateCardParameters struct {
		EnrichWordInformationFromDictionary bool
	}

	CreateCardRequest struct {
		UserID              string
		Language            lang.Language
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

	GetCardsForReviewRequest struct {
		UserID         string
		Language       lang.Language
		SentencesCount uint32
	}

	GetCardsRequest struct {
		UserID   string
		Language lang.Language
	}

	GetCardsResponse struct {
		UserID   string
		Language lang.Language
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
)
