package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/comment"
	"github.com/studio/platform/internal/pkg/apperr"
)

// CommentService handles comment-related business logic
type CommentService struct {
	commentRepo comment.Repository
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo comment.Repository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
	}
}

// CreateCommentInput represents input for creating a comment
type CreateCommentInput struct {
	CommentableType comment.CommentableType `json:"commentable_type" binding:"required"`
	CommentableID   uuid.UUID               `json:"commentable_id" binding:"required"`
	ParentID        *uuid.UUID              `json:"parent_id,omitempty"`
	Content         string                  `json:"content" binding:"required"`
}

// CreateComment creates a new comment
func (s *CommentService) CreateComment(ctx context.Context, userID uuid.UUID, input CreateCommentInput) (*comment.Comment, error) {
	now := time.Now()
	c := &comment.Comment{
		ID:              uuid.New(),
		UserID:          userID,
		CommentableType: input.CommentableType,
		CommentableID:   input.CommentableID,
		ParentID:        input.ParentID,
		Content:         input.Content,
		IsEdited:        false,
		IsDeleted:       false,
		LikeCount:       0,
		ReplyCount:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建评论失败", err)
	}

	// Increment parent reply count if this is a reply
	if input.ParentID != nil {
		if err := s.commentRepo.IncrementReplyCount(ctx, *input.ParentID); err != nil {
			// Log error but don't fail
		}
	}

	return c, nil
}

// UpdateCommentInput represents input for updating a comment
type UpdateCommentInput struct {
	Content string `json:"content" binding:"required"`
}

// UpdateComment updates a comment
func (s *CommentService) UpdateComment(ctx context.Context, userID, commentID uuid.UUID, input UpdateCommentInput) (*comment.Comment, error) {
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, comment.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询评论失败", err)
	}

	if c.UserID != userID {
		return nil, apperr.New(apperr.CodeForbidden, "无权修改此评论")
	}

	c.Content = input.Content
	c.IsEdited = true
	c.UpdatedAt = time.Now()

	if err := s.commentRepo.Update(ctx, c); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新评论失败", err)
	}

	return c, nil
}

// DeleteComment deletes a comment
func (s *CommentService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, comment.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询评论失败", err)
	}

	if c.UserID != userID {
		return apperr.New(apperr.CodeForbidden, "无权删除此评论")
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "删除评论失败", err)
	}

	// Decrement parent reply count if this is a reply
	if c.ParentID != nil {
		if err := s.commentRepo.DecrementReplyCount(ctx, *c.ParentID); err != nil {
			// Log error but don't fail
		}
	}

	return nil
}

// ListCommentsInput represents input for listing comments
type ListCommentsInput struct {
	CommentableType comment.CommentableType
	CommentableID   uuid.UUID
	ParentID        *uuid.UUID
	Page            int
	PageSize        int
}

// ListCommentsOutput represents output for listing comments
type ListCommentsOutput struct {
	Comments []*comment.Comment `json:"comments"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Size     int                `json:"size"`
}

// ListComments retrieves comments with pagination
func (s *CommentService) ListComments(ctx context.Context, input ListCommentsInput) (*ListCommentsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := comment.ListFilter{
		CommentableType: input.CommentableType,
		CommentableID:   input.CommentableID,
		ParentID:        input.ParentID,
		Page:            input.Page,
		PageSize:        input.PageSize,
	}

	comments, total, err := s.commentRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询评论列表失败", err)
	}

	return &ListCommentsOutput{
		Comments: comments,
		Total:    total,
		Page:     input.Page,
		Size:     len(comments),
	}, nil
}

// LikeComment likes a comment
func (s *CommentService) LikeComment(ctx context.Context, userID, commentID uuid.UUID) error {
	// Check if already liked
	hasLiked, err := s.commentRepo.HasLiked(ctx, userID, commentID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "检查点赞状态失败", err)
	}
	if hasLiked {
		return apperr.New(apperr.CodeInvalidParam, "已点赞")
	}

	// Create like
	like := &comment.CommentLike{
		ID:        uuid.New(),
		UserID:    userID,
		CommentID: commentID,
		CreatedAt: time.Now(),
	}

	if err := s.commentRepo.LikeComment(ctx, like); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "点赞失败", err)
	}

	// Increment like count
	if err := s.commentRepo.IncrementLikeCount(ctx, commentID); err != nil {
		// Log error but don't fail
	}

	return nil
}

// UnlikeComment unlikes a comment
func (s *CommentService) UnlikeComment(ctx context.Context, userID, commentID uuid.UUID) error {
	if err := s.commentRepo.UnlikeComment(ctx, userID, commentID); err != nil {
		if errors.Is(err, comment.ErrNotLiked) {
			return apperr.New(apperr.CodeInvalidParam, "未点赞")
		}
		return apperr.Wrap(apperr.CodeInternalError, "取消点赞失败", err)
	}

	// Decrement like count
	if err := s.commentRepo.DecrementLikeCount(ctx, commentID); err != nil {
		// Log error but don't fail
	}

	return nil
}

// AdminListComments lists all comments for admin moderation
func (s *CommentService) AdminListComments(ctx context.Context, page, pageSize int) (*ListCommentsOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := comment.ListFilter{
		Page:     page,
		PageSize: pageSize,
	}

	comments, total, err := s.commentRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询评论失败", err)
	}

	return &ListCommentsOutput{
		Comments: comments,
		Total:    total,
		Page:     page,
		Size:     len(comments),
	}, nil
}

// AdminDeleteComment deletes any comment without ownership check (admin only)
func (s *CommentService) AdminDeleteComment(ctx context.Context, commentID uuid.UUID) error {
	_, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, comment.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询评论失败", err)
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "删除评论失败", err)
	}
	return nil
}
