package gameplay

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	SaveMatch(ctx context.Context, match *Match, results []MatchResult) error
	ListLeaderboard(ctx context.Context, limit int) ([]*LeaderboardEntry, error)
	ListRecentMatches(ctx context.Context, limit int) ([]*MatchSummary, error)
	ListUserRecentMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*MatchSummary, error)
}
