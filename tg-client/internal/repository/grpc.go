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

func defaultDeadlineUnaryInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func connectToGRPCService(cfg ClientConfig) (*grpc.ClientConn, error) {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if cfg.Timeout > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(defaultDeadlineUnaryInterceptor(cfg.Timeout)))
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc: dial error: %w", err)
	}

	return conn, nil
}
