package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/follow"
	"github.com/studio/platform/internal/pkg/apperr"
)

// FollowService handles follow-related business logic
type FollowService struct {
	followRepo follow.Repository
}

func NewFollowService(followRepo follow.Repository) *FollowService {
	return &FollowService{followRepo: followRepo}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return apperr.BadRequest("不能关注自己")
	}

	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if isFollowing {
		return apperr.BadRequest("已关注该用户")
	}

	f := &follow.UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt:  time.Now(),
	}
	return s.followRepo.Follow(ctx, f)
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return s.followRepo.Unfollow(ctx, followerID, followeeID)
}

func (s *FollowService) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	return s.followRepo.IsFollowing(ctx, followerID, followeeID)
}

func (s *FollowService) ListFollowers(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*follow.UserFollow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.followRepo.ListFollowers(ctx, userID, page, pageSize)
}

func (s *FollowService) ListFollowing(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*follow.UserFollow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.followRepo.ListFollowing(ctx, userID, page, pageSize)
}

func (s *FollowService) GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.followRepo.GetFollowingIDs(ctx, userID)
}

func (s *FollowService) GetStats(ctx context.Context, userID uuid.UUID) (*follow.FollowStats, error) {
	return s.followRepo.GetStats(ctx, userID)
}
