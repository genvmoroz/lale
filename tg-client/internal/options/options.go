package options

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		LogLevel              logrus.Level `envconfig:"APP_LOG_LEVEL" required:"true"`
		TelegramToken         string       `envconfig:"APP_TELEGRAM_TOKEN" required:"true"`
		TelegramUpdateTimeout uint         `envconfig:"APP_TELEGRAM_UPDATE_TIMEOUT" default:"60"`

		LaleServiceConfig LaleServiceConfig
	}

	LaleServiceConfig struct {
		Host    string        `envconfig:"APP_LALE_SERVICE_HOST" required:"true"`
		Port    uint          `envconfig:"APP_LALE_SERVICE_PORT" required:"true"`
		Timeout time.Duration `envconfig:"APP_LALE_SERVICE_TIMEOUT" default:"5s"`
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
