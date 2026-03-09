package follow

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserFollow represents a follow relationship between users
type UserFollow struct {
	FollowerID uuid.UUID `json:"follower_id"`
	FolloweeID uuid.UUID `json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// FollowStats holds follow counts for a user
type FollowStats struct {
	UserID         uuid.UUID `json:"user_id"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
}

var (
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrNotFollowing     = errors.New("not following this user")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)
