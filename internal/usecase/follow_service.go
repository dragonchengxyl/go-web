package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/follow"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/pkg/apperr"
)

// FollowService handles follow-related business logic
type FollowService struct {
	followRepo follow.Repository
	publisher  *streams.Publisher // may be nil
}

func NewFollowService(followRepo follow.Repository, opts ...func(*FollowService)) *FollowService {
	s := &FollowService{followRepo: followRepo}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithFollowPublisher injects an event publisher into FollowService.
func WithFollowPublisher(p *streams.Publisher) func(*FollowService) {
	return func(s *FollowService) { s.publisher = p }
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
	if err := s.followRepo.Follow(ctx, f); err != nil {
		if errors.Is(err, follow.ErrAlreadyFollowing) {
			return apperr.BadRequest("已关注该用户")
		}
		return err
	}
	if s.publisher != nil {
		go func() {
			_ = s.publisher.Publish(context.Background(), streams.EventUserFollowed, streams.UserFollowedPayload{
				FollowerID: followerID.String(),
				FolloweeID: followeeID.String(),
			})
		}()
	}
	return nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if err := s.followRepo.Unfollow(ctx, followerID, followeeID); err != nil {
		if errors.Is(err, follow.ErrNotFollowing) {
			return apperr.BadRequest("未关注该用户")
		}
		return err
	}
	return nil
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
