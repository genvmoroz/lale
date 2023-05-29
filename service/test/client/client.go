package client

import (
	"context"
	"fmt"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Config struct {
		Host    string        `envconfig:"APP_CLIENT_HOST" required:"true"`
		Port    int           `envconfig:"APP_CLIENT_PORT" required:"true"`
		Timeout time.Duration `envconfig:"APP_CLIENT_TIMEOUT" default:"10s"`
	}

	Client struct {
		conn api.LaleServiceClient
	}
)

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	target := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc: dial error: %w", err)
	}

	return &Client{conn: api.NewLaleServiceClient(conn)}, nil
}

func (c *Client) MustDo() api.LaleServiceClient {
	if c == nil || c.conn == nil {
		panic("nil grpc client")
	}

	return c.conn
}
