package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService         *usecase.PostService
	followService       *usecase.FollowService
	notificationService *usecase.NotificationService
}

func NewPostHandler(postService *usecase.PostService, followService *usecase.FollowService, notificationService *usecase.NotificationService) *PostHandler {
	return &PostHandler{
		postService:         postService,
		followService:       followService,
		notificationService: notificationService,
	}
}

// CreatePost POST /api/v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		Title           string   `json:"title"`
		Content         string   `json:"content" binding:"required"`
		MediaURLs       []string `json:"media_urls"`
		Tags            []string `json:"tags"`
		Visibility      string   `json:"visibility"`
		IsAIGenerated   bool     `json:"is_ai_generated"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	vis := post.Visibility(req.Visibility)
	if vis == "" {
		vis = post.VisibilityPublic
	}

	labels := map[string]bool{"is_ai_generated": req.IsAIGenerated}

	p, err := h.postService.CreatePost(c.Request.Context(), usecase.CreatePostInput{
		AuthorID:      userID,
		Title:         req.Title,
		Content:       req.Content,
		MediaURLs:     req.MediaURLs,
		Tags:          req.Tags,
		ContentLabels: labels,
		Visibility:    vis,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, p)
}

// GetPost GET /api/v1/posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的帖子ID"))
		return
	}
	p, err := h.postService.GetPost(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, p)
}

// UpdatePost PUT /api/v1/posts/:id
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的帖子ID"))
		return
	}

	var req struct {
		Title      string   `json:"title"`
		Content    string   `json:"content" binding:"required"`
		MediaURLs  []string `json:"media_urls"`
		Tags       []string `json:"tags"`
		Visibility string   `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	p, err := h.postService.UpdatePost(c.Request.Context(), userID, postID, usecase.UpdatePostInput{
		Title:      req.Title,
		Content:    req.Content,
		MediaURLs:  req.MediaURLs,
		Tags:       req.Tags,
		Visibility: post.Visibility(req.Visibility),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, p)
}

// DeletePost DELETE /api/v1/posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的帖子ID"))
		return
	}

	roleVal, _ := c.Get("role")
	role := user.Role(roleVal.(string))
	isAdmin := role == user.RoleAdmin || role == user.RoleSuperAdmin || role == user.RoleModerator

	if err := h.postService.DeletePost(c.Request.Context(), userID, postID, isAdmin); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "帖子已删除"})
}

// LikePost POST /api/v1/posts/:id/like
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的帖子ID"))
		return
	}
	if err := h.postService.LikePost(c.Request.Context(), userID, postID); err != nil {
		response.Error(c, err)
		return
	}

	// Notify post author async (fire-and-forget)
	if h.notificationService != nil {
		actorID := userID
		targetID := postID
		notifSvc := h.notificationService
		postSvc := h.postService
		go func() {
			defer func() { _ = recover() }()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			p, err := postSvc.GetPost(ctx, targetID)
			if err != nil || p.AuthorID == actorID {
				return
			}
			_ = notifSvc.Notify(ctx, &notification.Notification{
				UserID:     p.AuthorID,
				ActorID:    &actorID,
				Type:       notification.TypeLike,
				TargetID:   &targetID,
				TargetType: "post",
			})
		}()
	}

	response.Success(c, gin.H{"message": "点赞成功"})
}

// UnlikePost DELETE /api/v1/posts/:id/like
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的帖子ID"))
		return
	}
	if err := h.postService.UnlikePost(c.Request.Context(), userID, postID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "取消点赞成功"})
}

// GetFeed GET /api/v1/feed
func (h *PostHandler) GetFeed(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)

	followingIDs, err := h.followService.GetFollowingIDs(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	posts, total, err := h.postService.ListFeed(c.Request.Context(), followingIDs, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"posts": posts, "total": total, "page": page, "size": len(posts)})
}

// GetExplore GET /api/v1/explore
func (h *PostHandler) GetExplore(c *gin.Context) {
	page, pageSize := getPageParams(c)
	tag := c.Query("tag")
	posts, total, err := h.postService.ListExplore(c.Request.Context(), page, pageSize, tag)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"posts": posts, "total": total, "page": page, "size": len(posts)})
}

// GetHotTags GET /api/v1/explore/tags
func (h *PostHandler) GetHotTags(c *gin.Context) {
	tags, err := h.postService.GetHotTags(c.Request.Context(), 20)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, tags)
}

// ListUserPosts GET /api/v1/users/:username/posts
func (h *PostHandler) ListUserPosts(c *gin.Context) {
	page, pageSize := getPageParams(c)
	authorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	vis := post.VisibilityPublic
	posts, total, err := h.postService.ListUserPosts(c.Request.Context(), usecase.ListUserPostsInput{
		AuthorID:   authorID,
		Visibility: &vis,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"posts": posts, "total": total, "page": page, "size": len(posts)})
}
