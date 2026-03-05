package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/studio/platform/internal/domain/achievement"
)

const (
	leaderboardPointsKey = "leaderboard:points:total"
	leaderboardWeekKey   = "leaderboard:points:week"
)

// Leaderboard manages Redis sorted-set leaderboards
type Leaderboard struct {
	client *goredis.Client
}

// NewLeaderboard creates a new Leaderboard
func NewLeaderboard(client *goredis.Client) *Leaderboard {
	return &Leaderboard{client: client}
}

// AddScore upserts a user's score in the given leaderboard key
func (lb *Leaderboard) AddScore(ctx context.Context, key string, userID uuid.UUID, delta float64) error {
	_, err := lb.client.ZIncrBy(ctx, key, delta, userID.String()).Result()
	return err
}

// UpdateTotalPoints syncs a user's total point balance to the leaderboard
func (lb *Leaderboard) UpdateTotalPoints(ctx context.Context, userID uuid.UUID, totalPoints float64) error {
	_, err := lb.client.ZAdd(ctx, leaderboardPointsKey, goredis.Z{
		Score:  totalPoints,
		Member: userID.String(),
	}).Result()
	return err
}

// GetTopPoints returns the top N users by total points
func (lb *Leaderboard) GetTopPoints(ctx context.Context, limit int64) ([]achievement.LeaderboardEntry, error) {
	return lb.getTop(ctx, leaderboardPointsKey, limit)
}

// GetTopWeekly returns the top N users by weekly points
func (lb *Leaderboard) GetTopWeekly(ctx context.Context, limit int64) ([]achievement.LeaderboardEntry, error) {
	return lb.getTop(ctx, leaderboardWeekKey, limit)
}

// AddWeeklyPoints increments a user's weekly points score
func (lb *Leaderboard) AddWeeklyPoints(ctx context.Context, userID uuid.UUID, delta float64) error {
	_, err := lb.client.ZIncrBy(ctx, leaderboardWeekKey, delta, userID.String()).Result()
	return err
}

// GetUserRank returns the rank (0-based) and score of a user
func (lb *Leaderboard) GetUserRank(ctx context.Context, key string, userID uuid.UUID) (int64, float64, error) {
	rank, err := lb.client.ZRevRank(ctx, key, userID.String()).Result()
	if err != nil {
		return -1, 0, err
	}
	score, err := lb.client.ZScore(ctx, key, userID.String()).Result()
	if err != nil {
		return -1, 0, err
	}
	return rank, score, nil
}

func (lb *Leaderboard) getTop(ctx context.Context, key string, limit int64) ([]achievement.LeaderboardEntry, error) {
	results, err := lb.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("leaderboard get top: %w", err)
	}

	entries := make([]achievement.LeaderboardEntry, 0, len(results))
	for i, z := range results {
		memberStr, ok := z.Member.(string)
		if !ok {
			continue
		}
		uid, err := uuid.Parse(memberStr)
		if err != nil {
			continue
		}
		entries = append(entries, achievement.LeaderboardEntry{
			Rank:   int64(i + 1),
			UserID: uid,
			Score:  z.Score,
		})
	}
	return entries, nil
}
