package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/achievement"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/pkg/apperr"
)

// AchievementService handles achievement and points business logic
type AchievementService struct {
	repo        achievement.Repository
	leaderboard *redis.Leaderboard
}

// NewAchievementService creates a new AchievementService
func NewAchievementService(repo achievement.Repository, leaderboard *redis.Leaderboard) *AchievementService {
	return &AchievementService{repo: repo, leaderboard: leaderboard}
}

// ListAchievements returns all public achievements
func (s *AchievementService) ListAchievements(ctx context.Context) ([]*achievement.Achievement, error) {
	list, err := s.repo.ListAchievements(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询成就列表失败", err)
	}
	return list, nil
}

// GetUserAchievements returns achievements unlocked by the user
func (s *AchievementService) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]*achievement.UserAchievement, error) {
	list, err := s.repo.ListUserAchievements(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户成就失败", err)
	}
	return list, nil
}

// UnlockBySlug grants the named achievement to a user (admin or system call)
func (s *AchievementService) UnlockBySlug(ctx context.Context, userID uuid.UUID, slug string) (*achievement.UserAchievement, error) {
	a, err := s.repo.GetAchievementBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, achievement.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询成就失败", err)
	}

	// Check if already unlocked
	has, err := s.repo.HasAchievement(ctx, userID, a.ID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查成就状态失败", err)
	}
	if has {
		return nil, apperr.New(apperr.CodeInvalidParam, "该成就已解锁")
	}

	now := time.Now()
	ua := &achievement.UserAchievement{
		UserID:        userID,
		AchievementID: a.ID,
		Achievement:   a,
		ObtainedAt:    now,
	}

	if err := s.repo.UnlockAchievement(ctx, ua); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "解锁成就失败", err)
	}

	// Award points for achievement
	if a.Points > 0 {
		if err := s.awardPoints(ctx, userID, a.Points, achievement.SourceAchievement, slug); err != nil {
			// Non-fatal: log in production
			_ = err
		}
	}

	return ua, nil
}

// TryUnlockByDownload checks and unlocks download-related achievements
func (s *AchievementService) TryUnlockByDownload(ctx context.Context, userID uuid.UUID) {
	// first_download
	_ = s.tryUnlock(ctx, userID, "first_download")
}

// TryUnlockByComment checks and unlocks comment-related achievements
func (s *AchievementService) TryUnlockByComment(ctx context.Context, userID uuid.UUID, commentCount int64) {
	if commentCount >= 100 {
		_ = s.tryUnlock(ctx, userID, "chatterbox")
	}
}

// TryUnlockByPurchase checks and unlocks purchase-related achievements
func (s *AchievementService) TryUnlockByPurchase(ctx context.Context, userID uuid.UUID) {
	_ = s.tryUnlock(ctx, userID, "first_purchase")
}

func (s *AchievementService) tryUnlock(ctx context.Context, userID uuid.UUID, slug string) error {
	a, err := s.repo.GetAchievementBySlug(ctx, slug)
	if err != nil {
		return err
	}
	has, err := s.repo.HasAchievement(ctx, userID, a.ID)
	if err != nil || has {
		return err
	}
	ua := &achievement.UserAchievement{
		UserID:        userID,
		AchievementID: a.ID,
		ObtainedAt:    time.Now(),
	}
	if err := s.repo.UnlockAchievement(ctx, ua); err != nil {
		return err
	}
	if a.Points > 0 {
		_ = s.awardPoints(ctx, userID, a.Points, achievement.SourceAchievement, slug)
	}
	return nil
}

// --- Points ---

// GetUserPoints returns a user's point balance and history
func (s *AchievementService) GetUserPoints(ctx context.Context, userID uuid.UUID) (*achievement.UserPoints, error) {
	up, err := s.repo.GetUserPoints(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询积分失败", err)
	}
	return up, nil
}

// GetPointTransactions returns recent point history
func (s *AchievementService) GetPointTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]*achievement.PointTransaction, error) {
	txs, err := s.repo.ListPointTransactions(ctx, userID, limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询积分记录失败", err)
	}
	return txs, nil
}

// AwardPoints grants points to a user (called from other services or admin)
func (s *AchievementService) AwardPoints(ctx context.Context, userID uuid.UUID, amount int, source, refID string) error {
	return s.awardPoints(ctx, userID, amount, source, refID)
}

func (s *AchievementService) awardPoints(ctx context.Context, userID uuid.UUID, amount int, source, refID string) error {
	tx := &achievement.PointTransaction{
		UserID:    userID,
		Amount:    amount,
		Source:    source,
		RefID:     refID,
		CreatedAt: time.Now(),
	}
	if err := s.repo.AddPoints(ctx, tx); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "增加积分失败", err)
	}

	// Sync to Redis leaderboard
	up, err := s.repo.GetUserPoints(ctx, userID)
	if err == nil {
		_ = s.leaderboard.UpdateTotalPoints(ctx, userID, float64(up.TotalEarned))
		if amount > 0 {
			_ = s.leaderboard.AddWeeklyPoints(ctx, userID, float64(amount))
		}
	}
	return nil
}

// --- Leaderboard ---

// GetTopLeaderboard returns the top N users by total points
func (s *AchievementService) GetTopLeaderboard(ctx context.Context, limit int) ([]achievement.LeaderboardEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	entries, err := s.leaderboard.GetTopPoints(ctx, int64(limit))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询排行榜失败", err)
	}
	return entries, nil
}

// GetWeeklyLeaderboard returns the top N users by weekly points
func (s *AchievementService) GetWeeklyLeaderboard(ctx context.Context, limit int) ([]achievement.LeaderboardEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	entries, err := s.leaderboard.GetTopWeekly(ctx, int64(limit))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询周榜失败", err)
	}
	return entries, nil
}
