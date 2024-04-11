package repository

import (
	"context"
	"fmt"

	"github.com/genvmoroz/lale/service/api"
)

type LaleRepo struct {
	Client api.LaleServiceClient
}

func NewLaleRepo(ctx context.Context, cfg ClientConfig) (*LaleRepo, error) {
	conn, err := connectToGRPCService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to GRPC service: %w", err)
	}

	return &LaleRepo{
		Client: api.NewLaleServiceClient(conn),
	}, nil
}
