package achievement

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines data access for achievements
type Repository interface {
	// Achievement definitions
	ListAchievements(ctx context.Context) ([]*Achievement, error)
	GetAchievementByID(ctx context.Context, id int) (*Achievement, error)
	GetAchievementBySlug(ctx context.Context, slug string) (*Achievement, error)

	// User achievements
	UnlockAchievement(ctx context.Context, ua *UserAchievement) error
	ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]*UserAchievement, error)
	HasAchievement(ctx context.Context, userID uuid.UUID, achievementID int) (bool, error)

	// Points
	GetUserPoints(ctx context.Context, userID uuid.UUID) (*UserPoints, error)
	AddPoints(ctx context.Context, tx *PointTransaction) error
	ListPointTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]*PointTransaction, error)
}
