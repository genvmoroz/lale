package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/genvmoroz/lale-service/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	port int
	srv  *grpc.Server
}

const network = "tcp"

func NewServer(port int, resolver api.LaleServiceServer) *Server {
	srv := grpc.NewServer()
	api.RegisterLaleServiceServer(srv, resolver)

	return &Server{
		port: port,
		srv:  srv,
	}
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
