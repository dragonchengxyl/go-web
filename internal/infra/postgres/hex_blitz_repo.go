package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/gameplay"
)

type HexBlitzRepository struct {
	pool *pgxpool.Pool
}

func NewHexBlitzRepository(pool *pgxpool.Pool) *HexBlitzRepository {
	return &HexBlitzRepository{pool: pool}
}

func (r *HexBlitzRepository) SaveMatch(ctx context.Context, match *gameplay.Match, results []gameplay.MatchResult) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin hex blitz match tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO hex_blitz_matches (
			id, room_id, room_code, room_title, game_slug, started_at, finished_at, duration_sec, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, match.ID, match.RoomID, match.RoomCode, match.RoomTitle, match.GameSlug, match.StartedAt, match.FinishedAt, match.DurationSec, match.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert hex blitz match: %w", err)
	}

	for _, result := range results {
		_, err = tx.Exec(ctx, `
			INSERT INTO hex_blitz_match_results (
				id, match_id, room_id, room_code, room_title, user_id, player_name, score, rank, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (id) DO NOTHING
		`, result.ID, result.MatchID, result.RoomID, result.RoomCode, result.RoomTitle, result.UserID, result.PlayerName, result.Score, result.Rank, result.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert hex blitz match result: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit hex blitz match tx: %w", err)
	}
	return nil
}

func (r *HexBlitzRepository) ListLeaderboard(ctx context.Context, limit int) ([]*gameplay.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx, `
		WITH player_stats AS (
			SELECT
				CASE
					WHEN r.user_id IS NOT NULL THEN r.user_id::text
					ELSE 'guest:' || LOWER(r.player_name)
				END AS player_key,
				r.user_id,
				r.player_name,
				COALESCE(u.username, r.player_name) AS display_name,
				r.score,
				r.created_at,
				COUNT(*) OVER (
					PARTITION BY CASE
						WHEN r.user_id IS NOT NULL THEN r.user_id::text
						ELSE 'guest:' || LOWER(r.player_name)
					END
				) AS matches_played,
				MAX(r.created_at) OVER (
					PARTITION BY CASE
						WHEN r.user_id IS NOT NULL THEN r.user_id::text
						ELSE 'guest:' || LOWER(r.player_name)
					END
				) AS last_played,
				ROW_NUMBER() OVER (
					PARTITION BY CASE
						WHEN r.user_id IS NOT NULL THEN r.user_id::text
						ELSE 'guest:' || LOWER(r.player_name)
					END
					ORDER BY r.score DESC, r.created_at DESC
				) AS best_rank
			FROM hex_blitz_match_results r
			LEFT JOIN users u ON u.id = r.user_id
		)
		SELECT
			ROW_NUMBER() OVER (ORDER BY score DESC, last_played DESC) AS leaderboard_rank,
			user_id,
			player_name,
			display_name,
			score,
			matches_played,
			last_played
		FROM player_stats
		WHERE best_rank = 1
		ORDER BY score DESC, last_played DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list hex blitz leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]*gameplay.LeaderboardEntry, 0, limit)
	for rows.Next() {
		entry := &gameplay.LeaderboardEntry{}
		if err := rows.Scan(
			&entry.Rank,
			&entry.UserID,
			&entry.PlayerName,
			&entry.DisplayName,
			&entry.BestScore,
			&entry.Matches,
			&entry.LastPlayed,
		); err != nil {
			return nil, fmt.Errorf("scan hex blitz leaderboard: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *HexBlitzRepository) ListRecentMatches(ctx context.Context, limit int) ([]*gameplay.MatchSummary, error) {
	return r.listRecentMatches(ctx, nil, limit)
}

func (r *HexBlitzRepository) ListUserRecentMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*gameplay.MatchSummary, error) {
	return r.listRecentMatches(ctx, &userID, limit)
}

func (r *HexBlitzRepository) listRecentMatches(ctx context.Context, userID *uuid.UUID, limit int) ([]*gameplay.MatchSummary, error) {
	if limit <= 0 || limit > 50 {
		limit = 8
	}

	query := `
		SELECT DISTINCT m.id, m.room_id, m.room_code, m.room_title, m.game_slug, m.finished_at, m.duration_sec
		FROM hex_blitz_matches m
	`
	args := []any{}
	if userID != nil {
		query += `
			JOIN hex_blitz_match_results r ON r.match_id = m.id
			WHERE r.user_id = $1
		`
		args = append(args, *userID)
	}
	query += `
		ORDER BY m.finished_at DESC
		LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list hex blitz recent matches: %w", err)
	}
	defer rows.Close()

	matchIDs := make([]uuid.UUID, 0, limit)
	items := make([]*gameplay.MatchSummary, 0, limit)
	indexByMatchID := make(map[uuid.UUID]*gameplay.MatchSummary, limit)
	for rows.Next() {
		item := &gameplay.MatchSummary{}
		if err := rows.Scan(
			&item.MatchID,
			&item.RoomID,
			&item.RoomCode,
			&item.RoomTitle,
			&item.GameSlug,
			&item.FinishedAt,
			&item.DurationSec,
		); err != nil {
			return nil, fmt.Errorf("scan hex blitz recent match: %w", err)
		}
		matchIDs = append(matchIDs, item.MatchID)
		items = append(items, item)
		indexByMatchID[item.MatchID] = item
	}
	if len(matchIDs) == 0 {
		return items, nil
	}

	resultRows, err := r.pool.Query(ctx, `
		SELECT
			r.id,
			r.match_id,
			r.room_id,
			r.room_code,
			r.room_title,
			r.user_id,
			r.player_name,
			COALESCE(u.username, r.player_name) AS display_name,
			r.score,
			r.rank,
			m.started_at,
			m.finished_at,
			r.created_at
		FROM hex_blitz_match_results r
		JOIN hex_blitz_matches m ON m.id = r.match_id
		LEFT JOIN users u ON u.id = r.user_id
		WHERE r.match_id = ANY($1)
		ORDER BY r.match_id, r.rank ASC, r.score DESC, r.created_at ASC
	`, matchIDs)
	if err != nil {
		return nil, fmt.Errorf("list hex blitz recent match results: %w", err)
	}
	defer resultRows.Close()

	for resultRows.Next() {
		result := gameplay.MatchResult{}
		if err := resultRows.Scan(
			&result.ID,
			&result.MatchID,
			&result.RoomID,
			&result.RoomCode,
			&result.RoomTitle,
			&result.UserID,
			&result.PlayerName,
			&result.DisplayName,
			&result.Score,
			&result.Rank,
			&result.StartedAt,
			&result.FinishedAt,
			&result.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan hex blitz recent match result: %w", err)
		}
		summary := indexByMatchID[result.MatchID]
		if summary == nil {
			continue
		}
		summary.PlayerCount++
		if summary.WinnerName == "" || result.Rank == 1 {
			summary.WinnerName = result.DisplayName
			summary.WinnerScore = result.Score
		}
		if len(summary.TopResults) < 3 {
			summary.TopResults = append(summary.TopResults, result)
		}
	}

	return items, nil
}

var _ gameplay.Repository = (*HexBlitzRepository)(nil)
