package dependency

import (
	"context"
	"fmt"
	"github.com/genvmoroz/lale/service/pkg/speech"
	"github.com/genvmoroz/lale/service/pkg/speech/google"
	"time"

	"github.com/genvmoroz/lale/service/internal/algo"
	"github.com/genvmoroz/lale/service/internal/core"
	"github.com/genvmoroz/lale/service/internal/options"
	"github.com/genvmoroz/lale/service/internal/repo/card"
	"github.com/genvmoroz/lale/service/internal/repo/dictionary"
	"github.com/genvmoroz/lale/service/internal/repo/redis"
	"github.com/genvmoroz/lale/service/internal/repo/session"
	"github.com/genvmoroz/lale/service/pkg/sentence/hippo"
	"github.com/genvmoroz/lale/service/pkg/sentence/openai"
	"github.com/genvmoroz/lale/service/pkg/sentence/yourdictionary"
)

type Dependency struct {
	service core.Service
}

func NewDependency(ctx context.Context, cfg options.Config) (*Dependency, error) {
	yourDictionarySentenceScraper, err := yourdictionary.NewSentenceScraper(cfg.YourDictionarySentence)
	if err != nil {
		return nil, fmt.Errorf("create yourdictionary sentence scraper: %w", err)
	}

	hippoSentenceScraper, err := hippo.NewSentenceScraper(cfg.HippoSentence)
	if err != nil {
		return nil, fmt.Errorf("create hippo sentence scraper: %w", err)
	}

	openaiScraper, err := openai.NewSentenceScraper(cfg.OpenAISentence)
	if err != nil {
		return nil, fmt.Errorf("create openai sentence scraper: %w", err)
	}

	redisRepo := redis.NewRepo(cfg.Redis)
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
	_ = hippoSentenceScraper
	_ = yourDictionarySentenceScraper

	scrapers := []core.SentenceScraper{
		openaiScraper,
	}

	googleTextToSpeechClient, err := google.NewTextToSpeechClient(ctx, cfg.Google)
	if err != nil {
		return nil, fmt.Errorf("new google text-to-speech client: %w", err)
	}

	textToSpeechRepo := speech.NewRepo(googleTextToSpeechClient)

	service, err := core.NewService(
		cardRepo,
		userSessionRepo,
		scrapers,
		algo.NewAnki(time.Now),
		dictionaryRepo,
		textToSpeechRepo,
		core.DefaultValidator,
	)
	if err != nil {
		return nil, fmt.Errorf("create core service: %w", err)
	}

	// temporary unused clients
	_ = redisRepo

	return &Dependency{service: service}, nil
}

func (d *Dependency) BuildService() core.Service {
	return d.service
}
