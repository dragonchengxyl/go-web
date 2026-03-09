package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// CreatorHandler handles creator dashboard endpoints
type CreatorHandler struct {
	postService   *usecase.PostService
	followService *usecase.FollowService
	tipService    *usecase.TipService
}

func NewCreatorHandler(postService *usecase.PostService, followService *usecase.FollowService, tipService *usecase.TipService) *CreatorHandler {
	return &CreatorHandler{
		postService:   postService,
		followService: followService,
		tipService:    tipService,
	}
}

// GetStats GET /api/v1/creator/stats
func (h *CreatorHandler) GetStats(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	ctx := c.Request.Context()

	// Post stats: total posts + aggregate likes/comments
	posts, total, err := h.postService.ListUserPosts(ctx, usecase.ListUserPostsInput{
		AuthorID: uid,
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	var totalLikes, totalComments int64
	for _, p := range posts {
		totalLikes += int64(p.LikeCount)
		totalComments += int64(p.CommentCount)
	}

	// Follow stats
	followStats, err := h.followService.GetStats(ctx, uid)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Tip stats
	tipTotal, tipCount, _ := h.tipService.GetMyTipStats(ctx, uid)

	response.Success(c, gin.H{
		"post_count":     total,
		"total_likes":    totalLikes,
		"total_comments": totalComments,
		"follower_count": followStats.FollowerCount,
		"following_count": followStats.FollowingCount,
		"tip_total_cents": tipTotal,
		"tip_count":      tipCount,
	})
}
