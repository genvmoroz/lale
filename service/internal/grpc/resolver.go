package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/core"
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
		r.transformer.ToCoreUpdateCardPerformanceRequest,
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

func (r *Resolver) DeleteCard(ctx context.Context, req *api.DeleteCardRequest) (*api.DeleteCardResponse, error) {
	return genericResolver[
		api.DeleteCardRequest,
		core.DeleteCardRequest,
		api.DeleteCardResponse,
		core.DeleteCardResponse,
	](
		ctx,
		req,
		r.transformer.ToCoreDeleteCardRequest,
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
	toCoreReq func(*APIRequest) CoreRequest,
	serviceCall func(context.Context, CoreRequest) (CoreResponse, error),
	toAPIResp func(CoreResponse) *APIResponse) (*APIResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("nullable request (%T)", req))
	}

	coreReq := toCoreReq(req)

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
