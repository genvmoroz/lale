// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: api/lale-service.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	LaleService_InspectCard_FullMethodName           = "/api.LaleService/InspectCard"
	LaleService_PromptCard_FullMethodName            = "/api.LaleService/PromptCard"
	LaleService_CreateCard_FullMethodName            = "/api.LaleService/CreateCard"
	LaleService_GetAllCards_FullMethodName           = "/api.LaleService/GetAllCards"
	LaleService_UpdateCardPerformance_FullMethodName = "/api.LaleService/UpdateCardPerformance"
	LaleService_GetCardsToReview_FullMethodName      = "/api.LaleService/GetCardsToReview"
	LaleService_GetSentences_FullMethodName          = "/api.LaleService/GetSentences"
	LaleService_DeleteCard_FullMethodName            = "/api.LaleService/DeleteCard"
)

// LaleServiceClient is the client API for LaleService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LaleServiceClient interface {
	InspectCard(ctx context.Context, in *InspectCardRequest, opts ...grpc.CallOption) (*InspectCardResponse, error)
	PromptCard(ctx context.Context, in *PromptCardRequest, opts ...grpc.CallOption) (*PromptCardResponse, error)
	CreateCard(ctx context.Context, in *CreateCardRequest, opts ...grpc.CallOption) (*CreateCardResponse, error)
	GetAllCards(ctx context.Context, in *GetCardsRequest, opts ...grpc.CallOption) (*GetCardsResponse, error)
	UpdateCardPerformance(ctx context.Context, in *UpdateCardPerformanceRequest, opts ...grpc.CallOption) (*UpdateCardPerformanceResponse, error)
	GetCardsToReview(ctx context.Context, in *GetCardsForReviewRequest, opts ...grpc.CallOption) (*GetCardsResponse, error)
	GetSentences(ctx context.Context, in *GetSentencesRequest, opts ...grpc.CallOption) (*GetSentencesResponse, error)
	DeleteCard(ctx context.Context, in *DeleteCardRequest, opts ...grpc.CallOption) (*DeleteCardResponse, error)
}

type laleServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLaleServiceClient(cc grpc.ClientConnInterface) LaleServiceClient {
	return &laleServiceClient{cc}
}

func (c *laleServiceClient) InspectCard(ctx context.Context, in *InspectCardRequest, opts ...grpc.CallOption) (*InspectCardResponse, error) {
	out := new(InspectCardResponse)
	err := c.cc.Invoke(ctx, LaleService_InspectCard_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) PromptCard(ctx context.Context, in *PromptCardRequest, opts ...grpc.CallOption) (*PromptCardResponse, error) {
	out := new(PromptCardResponse)
	err := c.cc.Invoke(ctx, LaleService_PromptCard_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) CreateCard(ctx context.Context, in *CreateCardRequest, opts ...grpc.CallOption) (*CreateCardResponse, error) {
	out := new(CreateCardResponse)
	err := c.cc.Invoke(ctx, LaleService_CreateCard_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) GetAllCards(ctx context.Context, in *GetCardsRequest, opts ...grpc.CallOption) (*GetCardsResponse, error) {
	out := new(GetCardsResponse)
	err := c.cc.Invoke(ctx, LaleService_GetAllCards_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) UpdateCardPerformance(ctx context.Context, in *UpdateCardPerformanceRequest, opts ...grpc.CallOption) (*UpdateCardPerformanceResponse, error) {
	out := new(UpdateCardPerformanceResponse)
	err := c.cc.Invoke(ctx, LaleService_UpdateCardPerformance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) GetCardsToReview(ctx context.Context, in *GetCardsForReviewRequest, opts ...grpc.CallOption) (*GetCardsResponse, error) {
	out := new(GetCardsResponse)
	err := c.cc.Invoke(ctx, LaleService_GetCardsToReview_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) GetSentences(ctx context.Context, in *GetSentencesRequest, opts ...grpc.CallOption) (*GetSentencesResponse, error) {
	out := new(GetSentencesResponse)
	err := c.cc.Invoke(ctx, LaleService_GetSentences_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *laleServiceClient) DeleteCard(ctx context.Context, in *DeleteCardRequest, opts ...grpc.CallOption) (*DeleteCardResponse, error) {
	out := new(DeleteCardResponse)
	err := c.cc.Invoke(ctx, LaleService_DeleteCard_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LaleServiceServer is the server API for LaleService service.
// All implementations must embed UnimplementedLaleServiceServer
// for forward compatibility
type LaleServiceServer interface {
	InspectCard(context.Context, *InspectCardRequest) (*InspectCardResponse, error)
	PromptCard(context.Context, *PromptCardRequest) (*PromptCardResponse, error)
	CreateCard(context.Context, *CreateCardRequest) (*CreateCardResponse, error)
	GetAllCards(context.Context, *GetCardsRequest) (*GetCardsResponse, error)
	UpdateCardPerformance(context.Context, *UpdateCardPerformanceRequest) (*UpdateCardPerformanceResponse, error)
	GetCardsToReview(context.Context, *GetCardsForReviewRequest) (*GetCardsResponse, error)
	GetSentences(context.Context, *GetSentencesRequest) (*GetSentencesResponse, error)
	DeleteCard(context.Context, *DeleteCardRequest) (*DeleteCardResponse, error)
	mustEmbedUnimplementedLaleServiceServer()
}

// UnimplementedLaleServiceServer must be embedded to have forward compatible implementations.
type UnimplementedLaleServiceServer struct {
}

func (UnimplementedLaleServiceServer) InspectCard(context.Context, *InspectCardRequest) (*InspectCardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InspectCard not implemented")
}
func (UnimplementedLaleServiceServer) PromptCard(context.Context, *PromptCardRequest) (*PromptCardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PromptCard not implemented")
}
func (UnimplementedLaleServiceServer) CreateCard(context.Context, *CreateCardRequest) (*CreateCardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCard not implemented")
}
func (UnimplementedLaleServiceServer) GetAllCards(context.Context, *GetCardsRequest) (*GetCardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllCards not implemented")
}
func (UnimplementedLaleServiceServer) UpdateCardPerformance(context.Context, *UpdateCardPerformanceRequest) (*UpdateCardPerformanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCardPerformance not implemented")
}
func (UnimplementedLaleServiceServer) GetCardsToReview(context.Context, *GetCardsForReviewRequest) (*GetCardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCardsToReview not implemented")
}
func (UnimplementedLaleServiceServer) GetSentences(context.Context, *GetSentencesRequest) (*GetSentencesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSentences not implemented")
}
func (UnimplementedLaleServiceServer) DeleteCard(context.Context, *DeleteCardRequest) (*DeleteCardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCard not implemented")
}
func (UnimplementedLaleServiceServer) mustEmbedUnimplementedLaleServiceServer() {}

// UnsafeLaleServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LaleServiceServer will
// result in compilation errors.
type UnsafeLaleServiceServer interface {
	mustEmbedUnimplementedLaleServiceServer()
}

func RegisterLaleServiceServer(s grpc.ServiceRegistrar, srv LaleServiceServer) {
	s.RegisterService(&LaleService_ServiceDesc, srv)
}

func _LaleService_InspectCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InspectCardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).InspectCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_InspectCard_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).InspectCard(ctx, req.(*InspectCardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_PromptCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PromptCardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).PromptCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_PromptCard_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).PromptCard(ctx, req.(*PromptCardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_CreateCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).CreateCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_CreateCard_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).CreateCard(ctx, req.(*CreateCardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_GetAllCards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCardsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).GetAllCards(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_GetAllCards_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).GetAllCards(ctx, req.(*GetCardsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_UpdateCardPerformance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCardPerformanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).UpdateCardPerformance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_UpdateCardPerformance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).UpdateCardPerformance(ctx, req.(*UpdateCardPerformanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_GetCardsToReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCardsForReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).GetCardsToReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_GetCardsToReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).GetCardsToReview(ctx, req.(*GetCardsForReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_GetSentences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSentencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).GetSentences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_GetSentences_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).GetSentences(ctx, req.(*GetSentencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LaleService_DeleteCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LaleServiceServer).DeleteCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LaleService_DeleteCard_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LaleServiceServer).DeleteCard(ctx, req.(*DeleteCardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LaleService_ServiceDesc is the grpc.ServiceDesc for LaleService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LaleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.LaleService",
	HandlerType: (*LaleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InspectCard",
			Handler:    _LaleService_InspectCard_Handler,
		},
		{
			MethodName: "PromptCard",
			Handler:    _LaleService_PromptCard_Handler,
		},
		{
			MethodName: "CreateCard",
			Handler:    _LaleService_CreateCard_Handler,
		},
		{
			MethodName: "GetAllCards",
			Handler:    _LaleService_GetAllCards_Handler,
		},
		{
			MethodName: "UpdateCardPerformance",
			Handler:    _LaleService_UpdateCardPerformance_Handler,
		},
		{
			MethodName: "GetCardsToReview",
			Handler:    _LaleService_GetCardsToReview_Handler,
		},
		{
			MethodName: "GetSentences",
			Handler:    _LaleService_GetSentences_Handler,
		},
		{
			MethodName: "DeleteCard",
			Handler:    _LaleService_DeleteCard_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/lale-service.proto",
}
