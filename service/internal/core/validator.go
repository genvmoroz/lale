package core

import (
	"errors"
	"fmt"
	"strings"
)

type Validator struct{}

var DefaultValidator = Validator{}

func (Validator) ValidateInspectCardRequest(req InspectCardRequest) error {
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

func (Validator) ValidatePromptCardRequest(req PromptCardRequest) error {
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

func (Validator) ValidateCreateCardRequest(req CreateCardRequest) error {
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

func (Validator) ValidateDeleteCardRequest(req DeleteCardRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.CardID)) == 0 {
		return errors.New("cardID is required")
	}

	return nil
}

func (Validator) ValidateGetCardsForReviewRequest(req GetCardsForReviewRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Language.String())) == 0 {
		return errors.New("language is required")
	}

	return nil
}

func (Validator) ValidateUpdateCardPerformanceRequest(req UpdateCardPerformanceRequest) error {
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

func (Validator) ValidateGetCardsRequest(req GetCardsRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}

	return nil
}

func (Validator) ValidateGetSentencesRequest(req GetSentencesRequest) error {
	if len(strings.TrimSpace(req.UserID)) == 0 {
		return errors.New("userID is required")
	}
	if len(strings.TrimSpace(req.Word)) == 0 {
		return errors.New("word is required")
	}

	return nil
}
