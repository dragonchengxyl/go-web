package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/transport/ws"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo notification.Repository
	hub  ws.HubInterface
}

func NewNotificationService(repo notification.Repository, hub ws.HubInterface) *NotificationService {
	return &NotificationService{repo: repo, hub: hub}
}

// Notify creates a notification and pushes it to the recipient via WebSocket
func (s *NotificationService) Notify(ctx context.Context, n *notification.Notification) error {
	n.ID = uuid.New()
	n.IsRead = false
	n.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, n); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "创建通知失败", err)
	}

	if s.hub != nil {
		s.hub.SendToUser(n.UserID, ws.WSMessage{
			Type:    ws.MessageTypeNotification,
			Payload: n,
		})
	}
	return nil
}

// ListNotificationsInput represents input for listing notifications
type ListNotificationsInput struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
}

func (s *NotificationService) ListNotifications(ctx context.Context, input ListNotificationsInput) ([]*notification.Notification, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	return s.repo.List(ctx, input.UserID, input.Page, input.PageSize)
}

// MarkReadInput represents input for marking notifications as read
type MarkReadInput struct {
	UserID uuid.UUID
	IDs    []uuid.UUID // empty = mark all
}

func (s *NotificationService) MarkRead(ctx context.Context, input MarkReadInput) error {
	if len(input.IDs) == 0 {
		return s.repo.MarkAllRead(ctx, input.UserID)
	}
	return s.repo.MarkRead(ctx, input.UserID, input.IDs)
}

func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}
