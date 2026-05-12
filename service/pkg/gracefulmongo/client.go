package gracefulmongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/event"
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

	// Option is a functional option for configuring the MongoDB client.
	Option func(*clientOptions)

	clientOptions struct {
		monitor *event.CommandMonitor
	}
)

// WithMonitor sets a monitor for tracking MongoDB operations.
func WithMonitor(monitor *event.CommandMonitor) Option {
	return func(opts *clientOptions) {
		opts.monitor = monitor
	}
}

// disconnectTimeout bounds the graceful Disconnect call on shutdown.
const disconnectTimeout = 10 * time.Second

func NewClient(ctx context.Context, cfg Config, opts ...Option) (*mongo.Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Apply options
	clientOpts := &clientOptions{}
	for _, opt := range opts {
		opt(clientOpts)
	}

	mongoOpts := options.Client().
		ApplyURI(constructURI(cfg)).
		SetMaxPoolSize(cfg.MaxPoolSize)

	// Attach monitor if provided
	if clientOpts.monitor != nil {
		mongoOpts.SetMonitor(clientOpts.monitor)
	}

	client, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	// Graceful shutdown: parent ctx is already done, so a fresh timeout-bounded context is required for Disconnect.
	//nolint:gosec // intentional: parent ctx is canceled here, need a fresh ctx to disconnect cleanly
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), disconnectTimeout)
		defer cancel()

		if disconnectErr := client.Disconnect(shutdownCtx); disconnectErr != nil {
			logrus.Errorf("failed to disconnect mongodb client gracefully: %s", disconnectErr)
		}
	}()

	return client, nil
}

func constructURI(cfg Config) string {
	uri := strings.Builder{}

	fmt.Fprintf(&uri, "%s://", cfg.Protocol)
	fmt.Fprintf(&uri, "%s:%s", cfg.Creds.User, cfg.Creds.Pass)
	fmt.Fprintf(&uri, "@%s", cfg.Host)

	if cfg.Port != nil {
		fmt.Fprintf(&uri, ":%d", *cfg.Port)
	}

	uri.WriteString("/")

	if len(cfg.Params) != 0 {
		uri.WriteString("?")

		firstParam := true
		for k, v := range cfg.Params {
			if !firstParam {
				uri.WriteString("&")
			}
			fmt.Fprintf(&uri, "%s=%s", k, v)
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
