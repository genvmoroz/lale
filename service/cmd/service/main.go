package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/genvmoroz/lale/service/internal/dependency"
	"github.com/genvmoroz/lale/service/internal/grpc"
	"github.com/genvmoroz/lale/service/internal/infrastructure"
	"github.com/genvmoroz/lale/service/internal/options"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		logrus.Errorf("service error: %s", err.Error())
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	cfg, err := options.FromEnv()
	if err != nil {
		return fmt.Errorf("read env config: %w", err)
	}

	logrus.SetLevel(cfg.LogLevel)

	logrus.Info("build deps")
	deps, err := dependency.NewDependency(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build deps: %w", err)
	}

	logrus.Info("build service")
	coreService := deps.BuildService()

	logrus.Info("build gRPC service")
	resolver, err := grpc.NewResolver(coreService, grpc.DefaultTransformer())
	if err != nil {
		return fmt.Errorf("create gRPC service: %w", err)
	}

	grpcServer, err := grpc.NewServer(cfg.GRPCPort, resolver)
	if err != nil {
		return fmt.Errorf("create gRPC service: %w", err)
	}

	infoServer, err := infrastructure.NewServer(cfg.Infra, logrus.StandardLogger())
	if err != nil {
		return fmt.Errorf("create info server: %w", err)
	}

	errGroup.Go(func() error {
		return grpcServer.Run(ctx)
	})

	errGroup.Go(func() error {
		return infoServer.Run(ctx)
	})

	logrus.Info("service started")
	defer logrus.Info("service stopped")

	return errGroup.Wait()
}
