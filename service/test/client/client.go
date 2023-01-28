package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/genvmoroz/lale-service/api"
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

	return &Client{
		conn: api.NewLaleServiceClient(conn),
	}, nil
}

func (c *Client) InspectCard(ctx context.Context, in *api.InspectCardRequest) (*api.InspectCardResponse, error) {
	return c.conn.InspectCard(ctx, in)
}

func (c *Client) CreateCard(ctx context.Context, in *api.CreateCardRequest) (*api.CreateCardResponse, error) {
	return c.conn.CreateCard(ctx, in)
}

func (c *Client) GetAllCards(ctx context.Context, in *api.GetCardsRequest) (*api.GetCardsResponse, error) {
	return c.conn.GetAllCards(ctx, in)
}

func (c *Client) UpdateCardPerformance(ctx context.Context, in *api.UpdateCardPerformanceRequest) (*api.UpdateCardPerformanceResponse, error) {
	return c.conn.UpdateCardPerformance(ctx, in)

}

func (c *Client) GetCardsToReview(ctx context.Context, in *api.GetCardsForReviewRequest) (*api.GetCardsResponse, error) {
	return c.conn.GetCardsToReview(ctx, in)

}

func (c *Client) DeleteCard(ctx context.Context, in *api.DeleteCardRequest) (*api.DeleteCardResponse, error) {
	return c.conn.DeleteCard(ctx, in)
}
