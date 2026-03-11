package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/usecase"
	commonv1 "github.com/studio/platform/proto/common/v1"
	notificationv1 "github.com/studio/platform/proto/notification/v1"
)

// NotificationServer implements notificationv1.NotificationServiceServer.
type NotificationServer struct {
	notificationv1.UnimplementedNotificationServiceServer
	svc *usecase.NotificationService
}

// NewNotificationServer creates a new NotificationServer.
func NewNotificationServer(svc *usecase.NotificationService) *NotificationServer {
	return &NotificationServer{svc: svc}
}

// Send creates and sends a notification.
func (s *NotificationServer) Send(ctx context.Context, req *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	n := &notification.Notification{
		UserID:     userID,
		Type:       notification.Type(req.Type),
		TargetType: req.TargetType,
		CreatedAt:  time.Now(),
	}

	if req.ActorId != "" {
		if actorID, err := uuid.Parse(req.ActorId); err == nil {
			n.ActorID = &actorID
		}
	}
	if req.TargetId != "" {
		if targetID, err := uuid.Parse(req.TargetId); err == nil {
			n.TargetID = &targetID
		}
	}

	if err := s.svc.Notify(ctx, n); err != nil {
		return nil, err
	}
	return &notificationv1.SendResponse{Id: n.ID.String()}, nil
}

// List returns paginated notifications for a user.
func (s *NotificationServer) List(ctx context.Context, req *notificationv1.ListRequest) (*notificationv1.ListResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	notifications, total, err := s.svc.ListNotifications(ctx, usecase.ListNotificationsInput{
		UserID:   userID,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	protos := make([]*notificationv1.NotificationProto, len(notifications))
	for i, n := range notifications {
		p := &notificationv1.NotificationProto{
			Id:         n.ID.String(),
			UserId:     n.UserID.String(),
			Type:       string(n.Type),
			TargetType: n.TargetType,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt.Unix(),
		}
		if n.ActorID != nil {
			p.ActorId = n.ActorID.String()
		}
		if n.TargetID != nil {
			p.TargetId = n.TargetID.String()
		}
		protos[i] = p
	}

	return &notificationv1.ListResponse{
		Notifications: protos,
		Total:         total,
	}, nil
}

// MarkRead marks notifications as read.
func (s *NotificationServer) MarkRead(ctx context.Context, req *notificationv1.MarkReadRequest) (*commonv1.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(req.Ids))
	for _, id := range req.Ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		ids = append(ids, uid)
	}

	if err := s.svc.MarkRead(ctx, usecase.MarkReadInput{
		UserID: userID,
		IDs:    ids,
	}); err != nil {
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

// CountUnread returns the count of unread notifications.
func (s *NotificationServer) CountUnread(ctx context.Context, req *notificationv1.UserRequest) (*notificationv1.CountResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}
	count, err := s.svc.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &notificationv1.CountResponse{Count: count}, nil
}

// StreamUserNotifications streams unread notifications for a user.
func (s *NotificationServer) StreamUserNotifications(req *notificationv1.UserRequest, stream notificationv1.NotificationService_StreamUserNotificationsServer) error {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return err
	}

	notifications, _, err := s.svc.ListNotifications(stream.Context(), usecase.ListNotificationsInput{
		UserID:   userID,
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		return err
	}

	for _, n := range notifications {
		if n.IsRead {
			continue
		}
		p := &notificationv1.NotificationProto{
			Id:         n.ID.String(),
			UserId:     n.UserID.String(),
			Type:       string(n.Type),
			TargetType: n.TargetType,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt.Unix(),
		}
		if n.ActorID != nil {
			p.ActorId = n.ActorID.String()
		}
		if n.TargetID != nil {
			p.TargetId = n.TargetID.String()
		}
		if err := stream.Send(p); err != nil {
			return err
		}
	}
	return nil
}
