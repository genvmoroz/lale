package core

import (
	"errors"
	"strings"
)

type (
	Environment struct {
		Users []User
	}

	User struct {
		Name  string
		Cards []Card
	}

	Card struct {
		ID    string
		Words []Word
	}

	Word struct {
		Word        string
		Translation []string
	}
)

type LoadRequest struct {
	LaleServiceHost                    string
	LaleServicePort                    uint32
	ParallelUsers                      uint32
	CardsPerUser                       uint32
	WordsPerCard                       uint32
	ActionCreateCardEnabled            bool
	ActionInspectCardEnabled           bool
	ActionPromptCardEnabled            bool
	ActionGetAllCardsEnabled           bool
	ActionUpdateCardEnabled            bool
	ActionUpdateCardPerformanceEnabled bool
	ActionGetCardsToLearnEnabled       bool
	ActionGetCardsToRepeatEnabled      bool
	ActionGetSentencesEnabled          bool
	ActionGenerateStoryEnabled         bool
	ActionDeleteCardEnabled            bool
}

func (req *LoadRequest) Validate() error {
	if strings.TrimSpace(req.LaleServiceHost) == "" {
		return errors.New("lale service host is required")
	}
	if req.LaleServicePort == 0 {
		return errors.New("lale service port is required")
	}
	if req.ParallelUsers == 0 {
		return errors.New("parallel users must be greater than 0")
	}
	if req.CardsPerUser == 0 {
		return errors.New("cards per user must be greater than 0")
	}
	if req.WordsPerCard == 0 {
		return errors.New("words per card must be greater than 0")
	}
	return nil
}

type PerformerConfig struct {
	LaleServiceHost string
	LaleServicePort uint32
}
