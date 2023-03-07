package repository

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Host    string
	Port    uint
	Timeout time.Duration
}

func connectToGRPCService(ctx context.Context, cfg ClientConfig) (*grpc.ClientConn, error) {
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

	return conn, nil
}
