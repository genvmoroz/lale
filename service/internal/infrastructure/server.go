package infrastructure // todo: rewrite using standard library

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort uint `envconfig:"APP_INFRA_SERVER_PORT" default:"8888"`
}

type Server struct {
	cfg Config
	srv *http.Server
}

func NewServer(ctx context.Context, cfg Config, logger logrus.FieldLogger) (*Server, error) {
	if cfg.ServerPort < 1 {
		return nil, fmt.Errorf("server port must be greater than 0")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		return nil, fmt.Errorf("can't listen on port %d: %w", cfg.ServerPort, err)
	}

	if err = ln.Close(); err != nil {
		return nil, fmt.Errorf("can't close listener: %w", err)
	}

	server := &Server{
		cfg: cfg,
	}

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		ReadHeaderTimeout: time.Second * 10, //nolint:mnd // 10 seconds is a reasonable timeout
	}
	m := http.NewServeMux()
	// Create HTTP handler for Prometheus metrics.
	m.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics e.g. to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	httpSrv.Handler = m
	logger.Info("starting HTTP server", "addr", httpSrv.Addr)

	server.srv = httpSrv

	return server, nil
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	defer close(errChan)

	go func(ch chan error) {
		ch <- s.srv.ListenAndServe()
	}(errChan)

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	const shutdownTimeout = 2 * time.Second

	timeout, cancel := context.WithTimeout(context.Background(), shutdownTimeout) //nolint:contextcheck,lll // false-positive: https://github.com/kkHAIKE/contextcheck/issues/2
	defer cancel()

	if err := s.srv.Shutdown(timeout); err != nil {
		return fmt.Errorf("shutdown infra http server: %w", err)
	}

	return nil
}

func (s *Server) Close() error {
	if s != nil && s.srv != nil {
		if err := s.srv.Close(); err != nil {
			return fmt.Errorf("close infra http server: %w", err)
		}
	}
	return nil
}
