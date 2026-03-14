package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/follow"
)

// FollowRepository implements follow.Repository using PostgreSQL
type FollowRepository struct {
	pool *pgxpool.Pool
}

func NewFollowRepository(pool *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{pool: pool}
}

func (r *FollowRepository) Follow(ctx context.Context, f *follow.UserFollow) error {
	result, err := r.pool.Exec(ctx,
		`INSERT INTO user_follows (follower_id, followee_id, created_at) VALUES ($1, $2, $3)
		 ON CONFLICT (follower_id, followee_id) DO NOTHING`,
		f.FollowerID, f.FolloweeID, f.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}
	if result.RowsAffected() == 0 {
		return follow.ErrAlreadyFollowing
	}
	return nil
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM user_follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, followeeID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return follow.ErrNotFollowing
	}
	return nil
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND followee_id = $2)`,
		followerID, followeeID,
	).Scan(&exists)
	return exists, err
}

func (r *FollowRepository) ListFollowers(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*follow.UserFollow, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_follows WHERE followee_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT follower_id, followee_id, created_at FROM user_follows WHERE followee_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanFollows(rows, total)
}

func (r *FollowRepository) ListFollowing(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*follow.UserFollow, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_follows WHERE follower_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT follower_id, followee_id, created_at FROM user_follows WHERE follower_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanFollows(rows, total)
}

func scanFollows(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}, total int64) ([]*follow.UserFollow, int64, error) {
	follows := make([]*follow.UserFollow, 0)
	for rows.Next() {
		var f follow.UserFollow
		if err := rows.Scan(&f.FollowerID, &f.FolloweeID, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		follows = append(follows, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return follows, total, nil
}

func (r *FollowRepository) GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT followee_id FROM user_follows WHERE follower_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *FollowRepository) GetStats(ctx context.Context, userID uuid.UUID) (*follow.FollowStats, error) {
	stats := &follow.FollowStats{UserID: userID}
	err := r.pool.QueryRow(ctx,
		`SELECT
		  (SELECT COUNT(*) FROM user_follows WHERE followee_id = $1) AS follower_count,
		  (SELECT COUNT(*) FROM user_follows WHERE follower_id = $1) AS following_count`,
		userID,
	).Scan(&stats.FollowerCount, &stats.FollowingCount)
	return stats, err
}
