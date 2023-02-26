package repository

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Host    string
	Port    uint
	Timeout time.Duration
}

func connect(ctx context.Context, cfg ClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), opts...)
}
