package follow

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for follow data access
type Repository interface {
	Follow(ctx context.Context, f *UserFollow) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)

	ListFollowers(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*UserFollow, int64, error)
	ListFollowing(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*UserFollow, int64, error)

	GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*FollowStats, error)
}
