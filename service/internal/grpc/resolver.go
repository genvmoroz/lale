package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Resolver struct {
	service core.Service
	api.LaleServiceServer

	transformer Transformer
}

func NewResolver(service core.Service, transformer Transformer) (*Resolver, error) {
	if service == nil {
		return nil, errors.New("service is required")
	}

	return &Resolver{
		service:     service,
		transformer: transformer,
	}, nil
}

func (r *Resolver) InspectCard(ctx context.Context, req *api.InspectCardRequest) (*api.InspectCardResponse, error) {
	return genericResolver[
		api.InspectCardRequest,
		core.InspectCardRequest,
		api.InspectCardResponse,
		core.InspectCardResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreInspectCardRequest,
		r.service.InspectCard,
		r.transformer.ToAPIInspectCardResponse,
	)
}

func (r *Resolver) PromptCard(ctx context.Context, req *api.PromptCardRequest) (*api.PromptCardResponse, error) {
	return genericResolver[
		api.PromptCardRequest,
		core.PromptCardRequest,
		api.PromptCardResponse,
		core.PromptCardResponse,
	](
		ctx,
		req,
		r.transformer.ToCorePromptCardRequest,
		r.service.PromptCard,
		r.transformer.ToAPIPromptCardResponse,
	)
}

func (r *Resolver) CreateCard(ctx context.Context, req *api.CreateCardRequest) (*api.CreateCardResponse, error) {
	return genericResolver[
		api.CreateCardRequest,
		core.CreateCardRequest,
		api.CreateCardResponse,
		core.CreateCardResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreCreateCardRequest,
		r.service.CreateCard,
		r.transformer.ToAPICreateCardResponse,
	)
}

func (r *Resolver) GetAllCards(ctx context.Context, req *api.GetCardsRequest) (*api.GetCardsResponse, error) {
	return genericResolver[
		api.GetCardsRequest,
		core.GetCardsRequest,
		api.GetCardsResponse,
		core.GetCardsResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreGetCardsRequest,
		r.service.GetAllCards,
		r.transformer.ToAPIGetCardsResponse,
	)

}

func (r *Resolver) UpdateCardPerformance(ctx context.Context, req *api.UpdateCardPerformanceRequest) (*api.UpdateCardPerformanceResponse, error) {
	return genericResolver[
		api.UpdateCardPerformanceRequest,
		core.UpdateCardPerformanceRequest,
		api.UpdateCardPerformanceResponse,
		core.UpdateCardPerformanceResponse,
	](
		ctx,
		req,
		func(req *api.UpdateCardPerformanceRequest) (core.UpdateCardPerformanceRequest, error) {
			return r.transformer.ToCoreUpdateCardPerformanceRequest(req), nil
		},
		r.service.UpdateCardPerformance,
		r.transformer.ToAPIUpdateCardPerformanceResponse,
	)
}

func (r *Resolver) GetCardsToReview(ctx context.Context, req *api.GetCardsForReviewRequest) (*api.GetCardsResponse, error) {
	return genericResolver[
		api.GetCardsForReviewRequest,
		core.GetCardsForReviewRequest,
		api.GetCardsResponse,
		core.GetCardsResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreGetCardsForReviewRequest,
		r.service.GetCardsToReview,
		r.transformer.ToAPIGetCardsResponse,
	)
}

func (r *Resolver) GetSentences(ctx context.Context, req *api.GetSentencesRequest) (*api.GetSentencesResponse, error) {
	return genericResolver[
		api.GetSentencesRequest,
		core.GetSentencesRequest,
		api.GetSentencesResponse,
		core.GetSentencesResponse,
	](
		ctx,
		req,
		func(req *api.GetSentencesRequest) (core.GetSentencesRequest, error) {
			return r.transformer.ToCoreGetSentencesRequest(req), nil
		},
		r.service.GetSentences,
		r.transformer.ToAPIGetSentencesResponse,
	)
}

func (r *Resolver) GenerateStory(ctx context.Context, req *api.GenerateStoryRequest) (*api.GenerateStoryResponse, error) {
	return genericResolver[
		api.GenerateStoryRequest,
		core.GenerateStoryRequest,
		api.GenerateStoryResponse,
		core.GenerateStoryResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreGenerateStoryRequest,
		r.service.GenerateStory,
		r.transformer.ToAPIGenerateStoryResponse,
	)
}

func (r *Resolver) DeleteCard(ctx context.Context, req *api.DeleteCardRequest) (*api.DeleteCardResponse, error) {
	return genericResolver[
		api.DeleteCardRequest,
		core.DeleteCardRequest,
		api.DeleteCardResponse,
		core.DeleteCardResponse,
	](
		ctx,
		req,
		func(req *api.DeleteCardRequest) (core.DeleteCardRequest, error) {
			return r.transformer.ToCoreDeleteCardRequest(req), nil
		},
		r.service.DeleteCard,
		r.transformer.ToAPIDeleteCardResponse,
	)
}

func genericResolver[
	APIRequest any,
	CoreRequest any,
	APIResponse any,
	CoreResponse any,
](
	ctx context.Context,
	req *APIRequest,
	toCoreReq func(*APIRequest) (CoreRequest, error),
	serviceCall func(context.Context, CoreRequest) (CoreResponse, error),
	toAPIResp func(CoreResponse) *APIResponse) (*APIResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("nullable request (%T)", req))
	}

	coreReq, err := toCoreReq(req)
	if err != nil {
		return nil, fmt.Errorf("failed to transform request: %w", err)
	}

	coreResp, err := serviceCall(ctx, coreReq)
	if err != nil {
		return nil, resolveCoreError(err)
	}

	return toAPIResp(coreResp), nil
}

func resolveCoreError(err error) error {
	switch {
	case errors.As(err, &core.RequestValidationError{}):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &core.CardNotFoundError{}):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &core.CardAlreadyExistsError{}):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
