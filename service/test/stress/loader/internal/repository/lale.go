package repository

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	LaleRepoConfig struct {
		Host    string
		Port    uint32
		Timeout time.Duration
	}

	LaleRepo struct {
		Client api.LaleServiceClient
	}
)

func NewLaleRepo(cfg LaleRepoConfig) (*LaleRepo, error) {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc: dial error: %w", err)
	}

	return &LaleRepo{
		Client: api.NewLaleServiceClient(conn),
	}, nil
}
