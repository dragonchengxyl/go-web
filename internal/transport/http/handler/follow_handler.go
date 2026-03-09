package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// FollowHandler handles follow-related HTTP requests
type FollowHandler struct {
	followService       *usecase.FollowService
	notificationService *usecase.NotificationService
}

func NewFollowHandler(followService *usecase.FollowService, notificationService *usecase.NotificationService) *FollowHandler {
	return &FollowHandler{
		followService:       followService,
		notificationService: notificationService,
	}
}

// Follow POST /api/v1/users/:id/follow
func (h *FollowHandler) Follow(c *gin.Context) {
	followerID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	followeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	if err := h.followService.Follow(c.Request.Context(), followerID, followeeID); err != nil {
		response.Error(c, err)
		return
	}

	// Notify followee async (fire-and-forget)
	if h.notificationService != nil {
		actorID := followerID
		targetID := followeeID
		notifSvc := h.notificationService
		go func() {
			defer func() { recover() }()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = notifSvc.Notify(ctx, &notification.Notification{
				UserID:     targetID,
				ActorID:    &actorID,
				Type:       notification.TypeFollow,
				TargetID:   &actorID,
				TargetType: "user",
			})
		}()
	}

	response.Success(c, gin.H{"message": "关注成功"})
}

// Unfollow DELETE /api/v1/users/:id/follow
func (h *FollowHandler) Unfollow(c *gin.Context) {
	followerID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	followeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	if err := h.followService.Unfollow(c.Request.Context(), followerID, followeeID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "取消关注成功"})
}

// ListFollowers GET /api/v1/users/:id/followers
func (h *FollowHandler) ListFollowers(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	page, pageSize := getPageParams(c)
	followers, total, err := h.followService.ListFollowers(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"followers": followers, "total": total, "page": page, "size": len(followers)})
}

// ListFollowing GET /api/v1/users/:id/following
func (h *FollowHandler) ListFollowing(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	page, pageSize := getPageParams(c)
	following, total, err := h.followService.ListFollowing(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"following": following, "total": total, "page": page, "size": len(following)})
}

// GetFollowStats GET /api/v1/users/:id/follow-stats
func (h *FollowHandler) GetFollowStats(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	stats, err := h.followService.GetStats(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, stats)
}
