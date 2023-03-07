package repository

import (
	"context"
	"fmt"

	"github.com/genvmoroz/lale/service/api"
)

type LaleRepo struct {
	client api.LaleServiceClient
}

func NewLaleRepo(ctx context.Context, cfg ClientConfig) (*LaleRepo, error) {
	conn, err := connectToGRPCService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GRPC service: %w", err)
	}

	return &LaleRepo{
		client: api.NewLaleServiceClient(conn),
	}, nil
}

func (r *LaleRepo) InspectCard(ctx context.Context, req *api.InspectCardRequest) (*api.InspectCardResponse, error) {
	return r.client.InspectCard(ctx, req)
}

func (r *LaleRepo) CreateCard(ctx context.Context, req *api.CreateCardRequest) (*api.CreateCardResponse, error) {
	return r.client.CreateCard(ctx, req)
}

func (r *LaleRepo) GetAllCards(ctx context.Context, req *api.GetCardsRequest) (*api.GetCardsResponse, error) {
	return r.client.GetAllCards(ctx, req)
}

func (r *LaleRepo) UpdateCardPerformance(ctx context.Context, req *api.UpdateCardPerformanceRequest) (*api.UpdateCardPerformanceResponse, error) {
	return r.client.UpdateCardPerformance(ctx, req)
}

func (r *LaleRepo) GetCardsToReview(ctx context.Context, req *api.GetCardsForReviewRequest) (*api.GetCardsResponse, error) {
	return r.client.GetCardsToReview(ctx, req)
}

func (r *LaleRepo) DeleteCard(ctx context.Context, req *api.DeleteCardRequest) (*api.DeleteCardResponse, error) {
	return r.client.DeleteCard(ctx, req)
}
