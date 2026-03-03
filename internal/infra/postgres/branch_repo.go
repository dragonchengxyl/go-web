package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/game"
)

type BranchRepository struct {
	pool *pgxpool.Pool
}

func NewBranchRepository(pool *pgxpool.Pool) *BranchRepository {
	return &BranchRepository{pool: pool}
}

const createBranchSQL = `
	INSERT INTO game_branches (id, game_id, name, description, is_default, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
`

func (r *BranchRepository) Create(ctx context.Context, b *game.Branch) error {
	_, err := r.pool.Exec(ctx, createBranchSQL,
		b.ID, b.GameID, b.Name, b.Description, b.IsDefault, b.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

const getBranchByIDSQL = `
	SELECT id, game_id, name, description, is_default, created_at
	FROM game_branches WHERE id = $1
`

func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*game.Branch, error) {
	var b game.Branch
	err := r.pool.QueryRow(ctx, getBranchByIDSQL, id).Scan(
		&b.ID, &b.GameID, &b.Name, &b.Description, &b.IsDefault, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return &b, nil
}

const listBranchesByGameIDSQL = `
	SELECT id, game_id, name, description, is_default, created_at
	FROM game_branches WHERE game_id = $1
	ORDER BY is_default DESC, created_at ASC
`

func (r *BranchRepository) ListByGameID(ctx context.Context, gameID uuid.UUID) ([]*game.Branch, error) {
	rows, err := r.pool.Query(ctx, listBranchesByGameIDSQL, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	defer rows.Close()

	branches := make([]*game.Branch, 0)
	for rows.Next() {
		var b game.Branch
		err := rows.Scan(&b.ID, &b.GameID, &b.Name, &b.Description, &b.IsDefault, &b.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return branches, nil
}

const updateBranchSQL = `
	UPDATE game_branches
	SET name = $2, description = $3, is_default = $4
	WHERE id = $1
`

func (r *BranchRepository) Update(ctx context.Context, b *game.Branch) error {
	_, err := r.pool.Exec(ctx, updateBranchSQL,
		b.ID, b.Name, b.Description, b.IsDefault,
	)
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}
	return nil
}

const deleteBranchSQL = `DELETE FROM game_branches WHERE id = $1`

func (r *BranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteBranchSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	if result.RowsAffected() == 0 {
		return game.ErrNotFound
	}

	return nil
}

const getDefaultBranchSQL = `
	SELECT id, game_id, name, description, is_default, created_at
	FROM game_branches WHERE game_id = $1 AND is_default = true
	LIMIT 1
`

func (r *BranchRepository) GetDefaultBranch(ctx context.Context, gameID uuid.UUID) (*game.Branch, error) {
	var b game.Branch
	err := r.pool.QueryRow(ctx, getDefaultBranchSQL, gameID).Scan(
		&b.ID, &b.GameID, &b.Name, &b.Description, &b.IsDefault, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get default branch: %w", err)
	}
	return &b, nil
}
