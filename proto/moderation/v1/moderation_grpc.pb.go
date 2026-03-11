// Code generated manually. DO NOT EDIT.

package moderationv1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ModerationService_ReviewText_FullMethodName  = "/moderation.v1.ModerationService/ReviewText"
	ModerationService_ReviewImage_FullMethodName = "/moderation.v1.ModerationService/ReviewImage"
)

// ModerationServiceClient is the client API.
type ModerationServiceClient interface {
	ReviewText(ctx context.Context, in *ReviewTextRequest, opts ...grpc.CallOption) (*ReviewResult, error)
	ReviewImage(ctx context.Context, in *ReviewImageRequest, opts ...grpc.CallOption) (*ReviewResult, error)
}

type moderationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModerationServiceClient(cc grpc.ClientConnInterface) ModerationServiceClient {
	return &moderationServiceClient{cc}
}

func (c *moderationServiceClient) ReviewText(ctx context.Context, in *ReviewTextRequest, opts ...grpc.CallOption) (*ReviewResult, error) {
	out := new(ReviewResult)
	err := c.cc.Invoke(ctx, ModerationService_ReviewText_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moderationServiceClient) ReviewImage(ctx context.Context, in *ReviewImageRequest, opts ...grpc.CallOption) (*ReviewResult, error) {
	out := new(ReviewResult)
	err := c.cc.Invoke(ctx, ModerationService_ReviewImage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModerationServiceServer is the server API.
type ModerationServiceServer interface {
	ReviewText(context.Context, *ReviewTextRequest) (*ReviewResult, error)
	ReviewImage(context.Context, *ReviewImageRequest) (*ReviewResult, error)
	mustEmbedUnimplementedModerationServiceServer()
}

// UnimplementedModerationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedModerationServiceServer struct{}

func (UnimplementedModerationServiceServer) ReviewText(context.Context, *ReviewTextRequest) (*ReviewResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReviewText not implemented")
}
func (UnimplementedModerationServiceServer) ReviewImage(context.Context, *ReviewImageRequest) (*ReviewResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReviewImage not implemented")
}
func (UnimplementedModerationServiceServer) mustEmbedUnimplementedModerationServiceServer() {}

func RegisterModerationServiceServer(s grpc.ServiceRegistrar, srv ModerationServiceServer) {
	s.RegisterService(&ModerationService_ServiceDesc, srv)
}

func _ModerationService_ReviewText_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReviewTextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModerationServiceServer).ReviewText(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ModerationService_ReviewText_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModerationServiceServer).ReviewText(ctx, req.(*ReviewTextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModerationService_ReviewImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReviewImageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModerationServiceServer).ReviewImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ModerationService_ReviewImage_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModerationServiceServer).ReviewImage(ctx, req.(*ReviewImageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var ModerationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "moderation.v1.ModerationService",
	HandlerType: (*ModerationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ReviewText", Handler: _ModerationService_ReviewText_Handler},
		{MethodName: "ReviewImage", Handler: _ModerationService_ReviewImage_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/moderation/v1/moderation.proto",
}
