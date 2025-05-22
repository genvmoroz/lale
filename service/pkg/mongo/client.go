package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type (
	Config struct {
		Protocol    string
		Host        string
		Port        *int
		Params      map[string]string
		MaxPoolSize uint64

		Creds Creds
	}

	Creds struct {
		User string
		Pass string
	}
)

func NewGracefulClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	client, err := mongo.Connect(ctx,
		options.Client().
			ApplyURI(constructURI(cfg)).
			SetMaxPoolSize(cfg.MaxPoolSize),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Disconnect(shutdownCtx); err != nil {
			logrus.Errorf("graceful disconnect: %s", err)
		}
	}()

	return client, nil
}

func constructURI(cfg Config) string {
	uri := strings.Builder{}

	uri.WriteString(fmt.Sprintf("%s://", cfg.Protocol))
	uri.WriteString(fmt.Sprintf("%s:%s", cfg.Creds.User, cfg.Creds.Pass))
	uri.WriteString(fmt.Sprintf("@%s", cfg.Host))

	if cfg.Port != nil {
		uri.WriteString(fmt.Sprintf(":%d", *cfg.Port))
	}

	uri.WriteString("/")

	if len(cfg.Params) != 0 {
		uri.WriteString("?")

		firstParam := true
		for k, v := range cfg.Params {
			if !firstParam {
				uri.WriteString("&")
			}
			uri.WriteString(fmt.Sprintf("%s=%s", k, v))
			firstParam = false
		}
	}

	return uri.String()
}

func (cfg Config) validate() error {
	errMessages := make([]string, 0)

	if strings.TrimSpace(cfg.Protocol) == "" {
		errMessages = append(errMessages, "protocol is required")
	}
	if strings.TrimSpace(cfg.Host) == "" {
		errMessages = append(errMessages, "host is required")
	}
	if strings.TrimSpace(cfg.Creds.User) == "" {
		errMessages = append(errMessages, "user is required")
	}
	if strings.TrimSpace(cfg.Creds.Pass) == "" {
		errMessages = append(errMessages, "pass is required")
	}

	if len(errMessages) != 0 {
		return errors.New(strings.Join(errMessages, ", "))
	}

	return nil
}
