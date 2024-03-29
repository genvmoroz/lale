package dependency

import (
	"context"
	"fmt"
	"time"

	"github.com/genvmoroz/lale/service/internal/algo"
	"github.com/genvmoroz/lale/service/internal/core"
	"github.com/genvmoroz/lale/service/internal/options"
	"github.com/genvmoroz/lale/service/internal/repo/card"
	"github.com/genvmoroz/lale/service/internal/repo/dictionary"
	"github.com/genvmoroz/lale/service/internal/repo/session"
	"github.com/genvmoroz/lale/service/pkg/openai"
	"github.com/genvmoroz/lale/service/pkg/speech"
	"github.com/genvmoroz/lale/service/pkg/speech/google"
)

type Dependency struct {
	service *core.Service
}

func NewDependency(ctx context.Context, cfg options.Config) (*Dependency, error) {
	openaiHelper, err := openai.NewHelper(cfg.OpenAI) // TODO: move it to internal/repo package and name it AI
	if err != nil {
		return nil, fmt.Errorf("create openai helper: %w", err)
	}

	userSessionRepo, err := session.NewRepo(cfg.Session)
	if err != nil {
		return nil, fmt.Errorf("create user session client: %w", err)
	}
	cardRepo, err := card.NewRepo(ctx, cfg.CardRepo)
	if err != nil {
		return nil, fmt.Errorf("create card repo: %w", err)
	}

	dictionaryRepo, err := dictionary.NewRepo(
		dictionary.Config{
			Host:    cfg.Dictionary.Host,
			Retries: cfg.Dictionary.Retries,
			Timeout: cfg.Dictionary.Timeout,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create dictionary client: %w", err)
	}

	googleTextToSpeechClient, err := google.NewTextToSpeechClient(ctx, cfg.Google)
	if err != nil {
		return nil, fmt.Errorf("new google text-to-speech client: %w", err)
	}

	textToSpeechRepo := speech.NewRepo(googleTextToSpeechClient)

	service, err := core.NewService(
		cardRepo,
		userSessionRepo,
		openaiHelper,
		algo.NewAnki(time.Now),
		dictionaryRepo,
		textToSpeechRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("create core service: %w", err)
	}

	return &Dependency{service: service}, nil
}

func (d *Dependency) BuildService() *core.Service {
	return d.service
}
