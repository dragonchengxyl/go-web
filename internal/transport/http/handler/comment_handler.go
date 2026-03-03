package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/comment"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentService *usecase.CommentService
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(commentService *usecase.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment creates a new comment
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input usecase.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	comment, err := h.commentService.CreateComment(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, comment)
}

// UpdateComment updates a comment
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的评论ID"))
		return
	}

	var input usecase.UpdateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), userID, commentID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, comment)
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的评论ID"))
		return
	}

	if err := h.commentService.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "评论已删除"})
}

// ListComments retrieves comments
func (h *CommentHandler) ListComments(c *gin.Context) {
	commentableType := comment.CommentableType(c.Query("commentable_type"))
	if commentableType == "" {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "评论对象类型不能为空"))
		return
	}

	commentableID, err := uuid.Parse(c.Query("commentable_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的评论对象ID"))
		return
	}

	var input usecase.ListCommentsInput
	input.CommentableType = commentableType
	input.CommentableID = commentableID
	input.Page = 1
	input.PageSize = 20

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			input.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			input.PageSize = ps
		}
	}

	if parentID := c.Query("parent_id"); parentID != "" {
		if pid, err := uuid.Parse(parentID); err == nil {
			input.ParentID = &pid
		}
	}

	output, err := h.commentService.ListComments(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, output)
}

// LikeComment likes a comment
func (h *CommentHandler) LikeComment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的评论ID"))
		return
	}

	if err := h.commentService.LikeComment(c.Request.Context(), userID, commentID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "点赞成功"})
}

// UnlikeComment unlikes a comment
func (h *CommentHandler) UnlikeComment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的评论ID"))
		return
	}

	if err := h.commentService.UnlikeComment(c.Request.Context(), userID, commentID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "取消点赞成功"})
}
