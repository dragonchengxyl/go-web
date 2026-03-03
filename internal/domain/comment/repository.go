package comment

import (
	"context"

	"github.com/google/uuid"
)

// ListFilter represents filters for listing comments
type ListFilter struct {
	CommentableType CommentableType
	CommentableID   uuid.UUID
	ParentID        *uuid.UUID
	UserID          *uuid.UUID
	Page            int
	PageSize        int
}

// Repository defines the interface for comment data access
type Repository interface {
	// Create creates a new comment
	Create(ctx context.Context, comment *Comment) error

	// GetByID retrieves a comment by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)

	// Update updates a comment
	Update(ctx context.Context, comment *Comment) error

	// Delete deletes a comment (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves comments with pagination and filters
	List(ctx context.Context, filter ListFilter) ([]*Comment, int64, error)

	// IncrementReplyCount increments the reply count for a comment
	IncrementReplyCount(ctx context.Context, commentID uuid.UUID) error

	// DecrementReplyCount decrements the reply count for a comment
	DecrementReplyCount(ctx context.Context, commentID uuid.UUID) error

	// LikeComment likes a comment
	LikeComment(ctx context.Context, like *CommentLike) error

	// UnlikeComment unlikes a comment
	UnlikeComment(ctx context.Context, userID, commentID uuid.UUID) error

	// HasLiked checks if user has liked a comment
	HasLiked(ctx context.Context, userID, commentID uuid.UUID) (bool, error)

	// IncrementLikeCount increments the like count for a comment
	IncrementLikeCount(ctx context.Context, commentID uuid.UUID) error

	// DecrementLikeCount decrements the like count for a comment
	DecrementLikeCount(ctx context.Context, commentID uuid.UUID) error
}
