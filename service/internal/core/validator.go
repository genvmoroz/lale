package core

import (
	"errors"
	"fmt"
	"strings"
)

type validator struct{}

func (validator) ValidateInspectCardRequest(req InspectCardRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Language.String())) == 0 {
		return errors.New("language is required")
	}
	if len(strings.TrimSpace(req.Word)) == 0 {
		return errors.New("word is required")
	}

	return nil
}

func (validator) ValidatePromptCardRequest(req PromptCardRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.WordLanguage.String())) == 0 {
		return errors.New("word language is required")
	}
	if len(strings.TrimSpace(req.TranslationLanguage.String())) == 0 {
		return errors.New("translation language is required")
	}
	if len(strings.TrimSpace(req.Word)) == 0 {
		return errors.New("word is required")
	}

	return nil
}

func (validator) ValidateCreateCardRequest(req CreateCardRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Language.String())) == 0 {
		return errors.New("language is required")
	}
	if len(req.WordInformationList) == 0 {
		return errors.New("wordInformationList are required, specify one at least")
	}

	return nil
}

func (validator) ValidateDeleteCardRequest(req DeleteCardRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.CardID)) == 0 {
		return errors.New("cardID is required")
	}

	return nil
}

func (validator) ValidateGetCardsForReviewRequest(req GetCardsForReviewRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Language.String())) == 0 {
		return errors.New("language is required")
	}

	return nil
}

func (validator) ValidateUpdateCardPerformanceRequest(req UpdateCardPerformanceRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.CardID)) == 0 {
		return errors.New("cardID is required")
	}
	if req.PerformanceRating > 5 {
		return fmt.Errorf("performance rating %d is out of range [0:5]", req.PerformanceRating)
	}

	return nil
}

func (validator) ValidateGetCardsRequest(req GetCardsRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}

	return nil
}

func (validator) ValidateGetSentencesRequest(req GetSentencesRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Word)) == 0 {
		return errors.New("word is required")
	}

	return nil
}

func (validator) ValidateGenerateStoryRequest(req GenerateStoryRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Language.String())) == 0 {
		return errors.New("language is required")
	}

	return nil
}
