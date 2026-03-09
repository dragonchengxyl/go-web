package post

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for post data access
type Repository interface {
	Create(ctx context.Context, p *Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	Update(ctx context.Context, p *Post) error
	Delete(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, filter ListFilter) ([]*Post, int64, error)
	ListFeed(ctx context.Context, followeeIDs []uuid.UUID, filter ListFilter) ([]*Post, int64, error)

	LikePost(ctx context.Context, like *PostLike) error
	UnlikePost(ctx context.Context, userID, postID uuid.UUID) error
	HasLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	IncrementLikeCount(ctx context.Context, postID uuid.UUID) error
	DecrementLikeCount(ctx context.Context, postID uuid.UUID) error
	IncrementCommentCount(ctx context.Context, postID uuid.UUID) error
	DecrementCommentCount(ctx context.Context, postID uuid.UUID) error
}
