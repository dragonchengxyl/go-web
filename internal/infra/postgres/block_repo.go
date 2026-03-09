package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/block"
)

type BlockRepository struct {
	pool *pgxpool.Pool
}

func NewBlockRepository(pool *pgxpool.Pool) *BlockRepository {
	return &BlockRepository{pool: pool}
}

func (r *BlockRepository) Block(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_blocks (blocker_id, blocked_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`,
		blockerID, blockedID,
	)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}
	return nil
}

func (r *BlockRepository) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`,
		blockerID, blockedID,
	)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return block.ErrNotBlocked
	}
	return nil
}

func (r *BlockRepository) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2)`,
		blockerID, blockedID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check block: %w", err)
	}
	return exists, nil
}

func (r *BlockRepository) ListBlockedIDs(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT blocked_id FROM user_blocks WHERE blocker_id = $1`,
		blockerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan blocked id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
