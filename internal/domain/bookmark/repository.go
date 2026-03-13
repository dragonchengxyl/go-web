package bookmark

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence operations for bookmarks.
type Repository interface {
	Create(ctx context.Context, item *Bookmark) error
	Delete(ctx context.Context, userID uuid.UUID, targetType TargetType, targetID uuid.UUID) error
	DeleteBatch(ctx context.Context, userID uuid.UUID, targetType TargetType, targetIDs []uuid.UUID) error
	Exists(ctx context.Context, userID uuid.UUID, targetType TargetType, targetID uuid.UUID) (bool, error)
	List(ctx context.Context, userID uuid.UUID, targetType TargetType, page, pageSize int, sort string) ([]*Bookmark, int64, error)
}
