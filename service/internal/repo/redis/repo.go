package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type (
	Config struct {
		Host               string `envconfig:"APP_REDIS_HOST" required:"true"`
		Port               int    `envconfig:"APP_REDIS_PORT" required:"true"`
		Pass               string `envconfig:"APP_REDIS_PASS" required:"true"`
		DB                 int    `envconfig:"APP_REDIS_DB" required:"true"`
		UseTLS             bool   `envconfig:"APP_REDIS_USE_TLS" required:"true"`
		InsecureSkipVerify bool   `envconfig:"APP_REDIS_INSECURE_SKIP_VERIFY" required:"true"`
	}

	Repo struct {
		client *redis.Client
	}
)

func NewRepo(config Config) *Repo {
	cfg := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:       config.DB,
		Password: config.Pass,
	}

	if config.UseTLS {
		cfg.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		}
	}

	return &Repo{client: redis.NewClient(&cfg)}
}

func (c *Repo) Get(ctx context.Context, key string, val interface{}) (bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	return true, json.Unmarshal(data, val)
}

func (c *Repo) SAdd(ctx context.Context, key, val string) error {
	return c.client.SAdd(ctx, key, val).Err()
}

func (c *Repo) All(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *Repo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err = c.client.Set(ctx, key, p, expiration).Err(); err != nil {
		return err
	}

	return nil
}
