package infrastructure

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort uint `envconfig:"APP_INFRA_SERVER_PORT" default:"8888"`
}

type Server struct {
	cfg    Config
	echo   *echo.Echo
	logger logrus.FieldLogger
}

func NewServer(cfg Config, logger logrus.FieldLogger) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	if cfg.ServerPort < 1 {
		return nil, fmt.Errorf("server port should be greater than 0")
	}

	if err := tryToListen(cfg.ServerPort); err != nil {
		return nil, fmt.Errorf("can't listen on port %d: %w", cfg.ServerPort, err)
	}

	server := &Server{
		echo:   echo.New(),
		logger: logger,
		cfg:    cfg,
	}

	server.echo.HideBanner = true
	server.echo.HidePort = true

	// register prometheus and pprof handlers
	server.echo.GET("/metrics", echoprometheus.NewHandler())
	pprof.Register(server.echo)

	return server, nil
}

// Run starts the server and listens for incoming requests.
// The server will be stopped when the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func(ch chan error) {
		s.logger.Debug("starting infra http server")
		ch <- s.echo.Start(fmt.Sprintf(":%d", s.cfg.ServerPort))
	}(errChan)

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	const shutdownTimeout = 2 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout) //nolint:contextcheck // false-positive: https://github.com/kkHAIKE/contextcheck/issues/2
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown infra http server: %w", err)
	}
	s.logger.Debug("infra http server stopped")

	return nil
}

func tryToListen(port uint) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("can't listen on port %d: %w", port, err)
	}

	if err = ln.Close(); err != nil {
		return fmt.Errorf("can't close listener: %w", err)
	}

	return nil
}
