package options

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/genvmoroz/lale/service/test/client"
)

const defaultAppPrefix = "APP"

type Config struct {
	ClientConfig client.Config
}

func FromEnv() (Config, error) {
	config := Config{}

	err := envconfig.Process(defaultAppPrefix, &config)
	if err != nil {
		return config, fmt.Errorf("load env config: %w", err)
	}

	return config, nil
}
