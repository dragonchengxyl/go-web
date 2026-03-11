// Code generated manually. DO NOT EDIT.
// Run `make proto-gen` to regenerate from proto/stats/v1/stats.proto

package statsv1

import (
	"context"

	commonv1 "github.com/studio/platform/proto/common/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	StatsService_GetDashboardStats_FullMethodName = "/stats.v1.StatsService/GetDashboardStats"
	StatsService_GetRevenueChart_FullMethodName   = "/stats.v1.StatsService/GetRevenueChart"
	StatsService_StreamUserGrowth_FullMethodName  = "/stats.v1.StatsService/StreamUserGrowth"
)

// StatsServiceClient is the client API for StatsService.
type StatsServiceClient interface {
	GetDashboardStats(ctx context.Context, in *commonv1.Empty, opts ...grpc.CallOption) (*DashboardStatsProto, error)
	GetRevenueChart(ctx context.Context, in *ChartRequest, opts ...grpc.CallOption) (*ChartResponse, error)
	StreamUserGrowth(ctx context.Context, in *ChartRequest, opts ...grpc.CallOption) (StatsService_StreamUserGrowthClient, error)
}

// StatsService_StreamUserGrowthClient is the client streaming interface.
type StatsService_StreamUserGrowthClient interface {
	Recv() (*ChartPointProto, error)
	grpc.ClientStream
}

type statsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStatsServiceClient(cc grpc.ClientConnInterface) StatsServiceClient {
	return &statsServiceClient{cc}
}

func (c *statsServiceClient) GetDashboardStats(ctx context.Context, in *commonv1.Empty, opts ...grpc.CallOption) (*DashboardStatsProto, error) {
	out := new(DashboardStatsProto)
	err := c.cc.Invoke(ctx, StatsService_GetDashboardStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *statsServiceClient) GetRevenueChart(ctx context.Context, in *ChartRequest, opts ...grpc.CallOption) (*ChartResponse, error) {
	out := new(ChartResponse)
	err := c.cc.Invoke(ctx, StatsService_GetRevenueChart_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *statsServiceClient) StreamUserGrowth(ctx context.Context, in *ChartRequest, opts ...grpc.CallOption) (StatsService_StreamUserGrowthClient, error) {
	stream, err := c.cc.NewStream(ctx, &StatsService_ServiceDesc.Streams[0], StatsService_StreamUserGrowth_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &statsServiceStreamUserGrowthClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type statsServiceStreamUserGrowthClient struct {
	grpc.ClientStream
}

func (x *statsServiceStreamUserGrowthClient) Recv() (*ChartPointProto, error) {
	m := new(ChartPointProto)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StatsServiceServer is the server API for StatsService.
type StatsServiceServer interface {
	GetDashboardStats(context.Context, *commonv1.Empty) (*DashboardStatsProto, error)
	GetRevenueChart(context.Context, *ChartRequest) (*ChartResponse, error)
	StreamUserGrowth(*ChartRequest, StatsService_StreamUserGrowthServer) error
	mustEmbedUnimplementedStatsServiceServer()
}

// UnimplementedStatsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedStatsServiceServer struct{}

func (UnimplementedStatsServiceServer) GetDashboardStats(context.Context, *commonv1.Empty) (*DashboardStatsProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDashboardStats not implemented")
}
func (UnimplementedStatsServiceServer) GetRevenueChart(context.Context, *ChartRequest) (*ChartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRevenueChart not implemented")
}
func (UnimplementedStatsServiceServer) StreamUserGrowth(*ChartRequest, StatsService_StreamUserGrowthServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamUserGrowth not implemented")
}
func (UnimplementedStatsServiceServer) mustEmbedUnimplementedStatsServiceServer() {}

// UnsafeStatsServiceServer may be embedded to opt out of forward compatibility for this service.
type UnsafeStatsServiceServer interface {
	mustEmbedUnimplementedStatsServiceServer()
}

// StatsService_StreamUserGrowthServer is the server streaming interface.
type StatsService_StreamUserGrowthServer interface {
	Send(*ChartPointProto) error
	grpc.ServerStream
}

type statsServiceStreamUserGrowthServer struct {
	grpc.ServerStream
}

func (x *statsServiceStreamUserGrowthServer) Send(m *ChartPointProto) error {
	return x.ServerStream.SendMsg(m)
}

func RegisterStatsServiceServer(s grpc.ServiceRegistrar, srv StatsServiceServer) {
	s.RegisterService(&StatsService_ServiceDesc, srv)
}

func _StatsService_GetDashboardStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(commonv1.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatsServiceServer).GetDashboardStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StatsService_GetDashboardStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatsServiceServer).GetDashboardStats(ctx, req.(*commonv1.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _StatsService_GetRevenueChart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatsServiceServer).GetRevenueChart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StatsService_GetRevenueChart_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatsServiceServer).GetRevenueChart(ctx, req.(*ChartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StatsService_StreamUserGrowth_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ChartRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StatsServiceServer).StreamUserGrowth(m, &statsServiceStreamUserGrowthServer{stream})
}

var StatsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "stats.v1.StatsService",
	HandlerType: (*StatsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDashboardStats",
			Handler:    _StatsService_GetDashboardStats_Handler,
		},
		{
			MethodName: "GetRevenueChart",
			Handler:    _StatsService_GetRevenueChart_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamUserGrowth",
			Handler:       _StatsService_StreamUserGrowth_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/stats/v1/stats.proto",
}
