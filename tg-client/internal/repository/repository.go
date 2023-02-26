package repository

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"

	"github.com/genvmoroz/lale/service/api"
)

type LaleRepo struct {
	client api.LaleServiceClient
}

func NewLaleRepo(ctx context.Context, cfg ClientConfig) (*LaleRepo, error) {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
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
