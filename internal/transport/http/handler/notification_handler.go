package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService *usecase.NotificationService
}

func NewNotificationHandler(notificationService *usecase.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// ListNotifications GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	notifications, total, err := h.notificationService.ListNotifications(c.Request.Context(), usecase.ListNotificationsInput{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"notifications": notifications, "total": total, "page": page, "size": len(notifications)})
}

// MarkRead POST /api/v1/notifications/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&req)

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		if id, err := uuid.Parse(idStr); err == nil {
			ids = append(ids, id)
		}
	}

	if err := h.notificationService.MarkRead(c.Request.Context(), usecase.MarkReadInput{
		UserID: userID,
		IDs:    ids,
	}); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已标记为已读"})
}

// CountUnread GET /api/v1/notifications/unread-count
func (h *NotificationHandler) CountUnread(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	count, err := h.notificationService.CountUnread(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}
