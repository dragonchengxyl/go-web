package audiowork

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, work *Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*Work, error)
	List(ctx context.Context, filter ListFilter) ([]*Work, int64, error)
	Like(ctx context.Context, userID, workID uuid.UUID) error
	Unlike(ctx context.Context, userID, workID uuid.UUID) error
	HasLiked(ctx context.Context, userID, workID uuid.UUID) (bool, error)
	IncrementLikeCount(ctx context.Context, workID uuid.UUID) error
	DecrementLikeCount(ctx context.Context, workID uuid.UUID) error
	IncrementCommentCount(ctx context.Context, workID uuid.UUID) error
	DecrementCommentCount(ctx context.Context, workID uuid.UUID) error
}
