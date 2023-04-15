package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/genvmoroz/lale/service/internal/dependency"
	"github.com/genvmoroz/lale/service/internal/grpc"
	"github.com/genvmoroz/lale/service/internal/options"
)

func main() {
	if err := run(); err != nil {
		logrus.Errorf("service error: %s", err.Error())
	}
}

func run() error {
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

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
	resolver, err := grpc.NewResolver(coreService, grpc.DefaultTransformer)
	if err != nil {
		return fmt.Errorf("create gRPC service: %w", err)
	}

	grpcService := grpc.NewService(cfg.GRPCPort, resolver)

	errGroup.Go(func() error {
		return grpcService.Run(ctx)
	})

	logrus.Info("service started")
	defer logrus.Info("service stopped")

	return errGroup.Wait()
}
