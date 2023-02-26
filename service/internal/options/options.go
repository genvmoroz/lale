package options

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/genvmoroz/lale/service/internal/repo/card"
	"github.com/genvmoroz/lale/service/internal/repo/redis"
	"github.com/genvmoroz/lale/service/internal/repo/session"
	"github.com/genvmoroz/lale/service/pkg/sentence/hippo"
	"github.com/genvmoroz/lale/service/pkg/sentence/yourdictionary"
)

type (
	Config struct {
		GRPCPort int          `envconfig:"APP_GRPC_PORT" required:"true"`
		LogLevel logrus.Level `envconfig:"APP_LOG_LEVEL" required:"true"`

		Redis                  redis.Config
		YourDictionarySentence yourdictionary.Config
		HippoSentence          hippo.Config
		Session                session.Config
		CardRepo               card.Config
		Dictionary             DictionaryConfig
	}

	DictionaryConfig struct {
		Host    string        `envconfig:"APP_DICTIONARY_HOST" required:"true"`
		Retries uint          `envconfig:"APP_DICTIONARY_RETRIES" default:"3"`
		Timeout time.Duration `envconfig:"APP_DICTIONARY_TIMEOUT" default:"5s"`
	}
)

const appPrefix = "APP"

func FromEnv() (Config, error) {
	config := Config{}

	err := envconfig.Process(appPrefix, &config)
	if err != nil {
		return config, fmt.Errorf("failed to load config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
