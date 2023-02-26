package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/genvmoroz/lale/service/api"
)

type Service struct {
	port int
	srv  *grpc.Server
}

const network = "tcp"

func NewService(port int, resolver api.LaleServiceServer) *Service {
	srv := grpc.NewServer()
	api.RegisterLaleServiceServer(srv, resolver)

	return &Service{
		port: port,
		srv:  srv,
	}
}

func (s *Service) Run(ctx context.Context) error {
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(s.port))
	lis, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("failed to listen address [%s]: %w", addr, err)
	}

	go func() {
		<-ctx.Done()
		s.Close()
	}()

	return s.srv.Serve(lis)
}

func (s *Service) Close() {
	logrus.Debug("stop grpc service")
	s.srv.Stop()
}
