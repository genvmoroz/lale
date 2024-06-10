package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/genvmoroz/lale-service/api"
	"github.com/liamg/clinch/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultAddr     = "localhost:12022"
	defaultUserID   = "gennadiymoroz"
	defaultLanguage = "en"
)

func askForLaleServiceAddr() (string, int, error) {
	addr := prompt.EnterInput(fmt.Sprintf("Enter Lale gRPC service addr (def. %s): ", defaultAddr))
	if strings.TrimSpace(addr) == "" {
		addr = defaultAddr
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("split host port: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parse port: %w", err)
	}

	return host, port, nil
}

func askForUserIDAndLanguage() (string, string) {
	userID := prompt.EnterInput(fmt.Sprintf("Enter UserID (def. %s): ", defaultUserID))
	if strings.TrimSpace(userID) == "" {
		userID = defaultUserID
	}
	language := prompt.EnterInput(fmt.Sprintf("Enter Language (def. %s): ", defaultLanguage))
	if strings.TrimSpace(language) == "" {
		language = defaultLanguage
	}

	return userID, language
}

func connectToLaleService(ctx context.Context, host string, port int, timeout time.Duration) (api.LaleServiceClient, error) {
	conn, err := connectToGRPCService(ctx, host, port, timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to Lale GRPC service: %w", err)
	}

	return api.NewLaleServiceClient(conn), nil
}

func connectToGRPCService(ctx context.Context, host string, port int, timeout time.Duration) (*grpc.ClientConn, error) {
	target := net.JoinHostPort(host, strconv.Itoa(port))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc: dial error: %w", err)
	}

	return conn, nil
}
