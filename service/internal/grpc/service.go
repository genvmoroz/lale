package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/genvmoroz/lale/service/api"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Server struct {
	port int
	srv  *grpc.Server
}

const network = "tcp"

func NewServer(port int, resolver api.LaleServiceServer) (*Server, error) {
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	reg := prometheus.DefaultRegisterer

	if err := reg.Register(srvMetrics); err != nil {
		return nil, fmt.Errorf("registering grpc metrics: %w", err)
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
		),
	)

	api.RegisterLaleServiceServer(srv, resolver)

	srvMetrics.InitializeMetrics(srv)

	return &Server{
		port: port,
		srv:  srv,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(s.port))
	lis, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("listen address [%s]: %w", addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		s.close()
	case srvErr := <-errCh:
		if srvErr != nil {
			return fmt.Errorf("serve grpc: %w", srvErr)
		}
	}

	return nil
}

func (s *Server) close() {
	logrus.Debug("stop grpc service")
	s.srv.Stop()
}
