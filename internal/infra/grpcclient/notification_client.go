package grpcclient

import (
	"context"
	"time"

	"github.com/google/uuid"
	notificationv1 "github.com/studio/platform/proto/notification/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// NotificationGRPCClient wraps the gRPC notification client.
type NotificationGRPCClient struct {
	client notificationv1.NotificationServiceClient
	conn   *grpc.ClientConn
}

// NewNotificationGRPCClient dials the notification service.
func NewNotificationGRPCClient(addr string) (*NotificationGRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, err
	}
	return &NotificationGRPCClient{
		client: notificationv1.NewNotificationServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *NotificationGRPCClient) Close() error {
	return c.conn.Close()
}

// Send sends a notification via gRPC and returns the created notification ID.
func (c *NotificationGRPCClient) Send(ctx context.Context, userID uuid.UUID, nType, actorID, targetType, targetID, message string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.Send(ctx, &notificationv1.SendRequest{
		UserId:     userID.String(),
		Type:       nType,
		ActorId:    actorID,
		TargetType: targetType,
		TargetId:   targetID,
		Message:    message,
	})
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

// ListNotifications retrieves paginated notifications for a user.
func (c *NotificationGRPCClient) ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) (*notificationv1.ListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.List(ctx, &notificationv1.ListRequest{
		UserId:   userID.String(),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
}

// MarkRead marks the given notification IDs as read for a user.
func (c *NotificationGRPCClient) MarkRead(ctx context.Context, userID uuid.UUID, ids []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.MarkRead(ctx, &notificationv1.MarkReadRequest{
		UserId: userID.String(),
		Ids:    ids,
	})
	return err
}

// CountUnread returns the number of unread notifications for a user.
func (c *NotificationGRPCClient) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CountUnread(ctx, &notificationv1.UserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
