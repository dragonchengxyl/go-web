// Code generated manually. DO NOT EDIT.

package notificationv1

import (
	"context"

	commonv1 "github.com/studio/platform/proto/common/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NotificationService_Send_FullMethodName                    = "/notification.v1.NotificationService/Send"
	NotificationService_List_FullMethodName                    = "/notification.v1.NotificationService/List"
	NotificationService_MarkRead_FullMethodName                = "/notification.v1.NotificationService/MarkRead"
	NotificationService_CountUnread_FullMethodName             = "/notification.v1.NotificationService/CountUnread"
	NotificationService_StreamUserNotifications_FullMethodName = "/notification.v1.NotificationService/StreamUserNotifications"
)

// NotificationServiceClient is the client API.
type NotificationServiceClient interface {
	Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*commonv1.Empty, error)
	CountUnread(ctx context.Context, in *UserRequest, opts ...grpc.CallOption) (*CountResponse, error)
	StreamUserNotifications(ctx context.Context, in *UserRequest, opts ...grpc.CallOption) (NotificationService_StreamUserNotificationsClient, error)
}

// NotificationService_StreamUserNotificationsClient is the client streaming interface.
type NotificationService_StreamUserNotificationsClient interface {
	Recv() (*NotificationProto, error)
	grpc.ClientStream
}

type notificationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc}
}

func (c *notificationServiceClient) Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error) {
	out := new(SendResponse)
	err := c.cc.Invoke(ctx, NotificationService_Send_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, NotificationService_List_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*commonv1.Empty, error) {
	out := new(commonv1.Empty)
	err := c.cc.Invoke(ctx, NotificationService_MarkRead_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) CountUnread(ctx context.Context, in *UserRequest, opts ...grpc.CallOption) (*CountResponse, error) {
	out := new(CountResponse)
	err := c.cc.Invoke(ctx, NotificationService_CountUnread_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) StreamUserNotifications(ctx context.Context, in *UserRequest, opts ...grpc.CallOption) (NotificationService_StreamUserNotificationsClient, error) {
	stream, err := c.cc.NewStream(ctx, &NotificationService_ServiceDesc.Streams[0], NotificationService_StreamUserNotifications_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &notificationServiceStreamUserNotificationsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type notificationServiceStreamUserNotificationsClient struct {
	grpc.ClientStream
}

func (x *notificationServiceStreamUserNotificationsClient) Recv() (*NotificationProto, error) {
	m := new(NotificationProto)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NotificationServiceServer is the server API.
type NotificationServiceServer interface {
	Send(context.Context, *SendRequest) (*SendResponse, error)
	List(context.Context, *ListRequest) (*ListResponse, error)
	MarkRead(context.Context, *MarkReadRequest) (*commonv1.Empty, error)
	CountUnread(context.Context, *UserRequest) (*CountResponse, error)
	StreamUserNotifications(*UserRequest, NotificationService_StreamUserNotificationsServer) error
	mustEmbedUnimplementedNotificationServiceServer()
}

// UnimplementedNotificationServiceServer must be embedded.
type UnimplementedNotificationServiceServer struct{}

func (UnimplementedNotificationServiceServer) Send(context.Context, *SendRequest) (*SendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedNotificationServiceServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedNotificationServiceServer) MarkRead(context.Context, *MarkReadRequest) (*commonv1.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkRead not implemented")
}
func (UnimplementedNotificationServiceServer) CountUnread(context.Context, *UserRequest) (*CountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CountUnread not implemented")
}
func (UnimplementedNotificationServiceServer) StreamUserNotifications(*UserRequest, NotificationService_StreamUserNotificationsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamUserNotifications not implemented")
}
func (UnimplementedNotificationServiceServer) mustEmbedUnimplementedNotificationServiceServer() {}

// NotificationService_StreamUserNotificationsServer is the server streaming interface.
type NotificationService_StreamUserNotificationsServer interface {
	Send(*NotificationProto) error
	grpc.ServerStream
}

type notificationServiceStreamUserNotificationsServer struct {
	grpc.ServerStream
}

func (x *notificationServiceStreamUserNotificationsServer) Send(m *NotificationProto) error {
	return x.ServerStream.SendMsg(m)
}

func RegisterNotificationServiceServer(s grpc.ServiceRegistrar, srv NotificationServiceServer) {
	s.RegisterService(&NotificationService_ServiceDesc, srv)
}

func _NotificationService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: NotificationService_Send_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).Send(ctx, req.(*SendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: NotificationService_List_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_MarkRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarkReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).MarkRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: NotificationService_MarkRead_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).MarkRead(ctx, req.(*MarkReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_CountUnread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).CountUnread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: NotificationService_CountUnread_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).CountUnread(ctx, req.(*UserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_StreamUserNotifications_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UserRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NotificationServiceServer).StreamUserNotifications(m, &notificationServiceStreamUserNotificationsServer{stream})
}

var NotificationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "notification.v1.NotificationService",
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Send", Handler: _NotificationService_Send_Handler},
		{MethodName: "List", Handler: _NotificationService_List_Handler},
		{MethodName: "MarkRead", Handler: _NotificationService_MarkRead_Handler},
		{MethodName: "CountUnread", Handler: _NotificationService_CountUnread_Handler},
	},
	Streams: []grpc.StreamDesc{
		{StreamName: "StreamUserNotifications", Handler: _NotificationService_StreamUserNotifications_Handler, ServerStreams: true},
	},
	Metadata: "proto/notification/v1/notification.proto",
}
