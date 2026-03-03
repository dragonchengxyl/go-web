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

type GameRepository struct {
	pool *pgxpool.Pool
}

func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{pool: pool}
}

const createGameSQL = `
	INSERT INTO games (id, slug, title, subtitle, description, cover_key, banner_key,
	                   trailer_url, genre, tags, engine, status, release_date, developer_id,
	                   created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
`

func (r *GameRepository) Create(ctx context.Context, g *game.Game) error {
	_, err := r.pool.Exec(ctx, createGameSQL,
		g.ID, g.Slug, g.Title, g.Subtitle, g.Description, g.CoverKey, g.BannerKey,
		g.TrailerURL, g.Genre, g.Tags, g.Engine, g.Status, g.ReleaseDate, g.DeveloperID,
		g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	return nil
}

const getGameByIDSQL = `
	SELECT id, slug, title, subtitle, description, cover_key, banner_key, trailer_url,
	       genre, tags, engine, status, release_date, developer_id, created_at, updated_at
	FROM games WHERE id = $1
`

func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*game.Game, error) {
	var g game.Game
	err := r.pool.QueryRow(ctx, getGameByIDSQL, id).Scan(
		&g.ID, &g.Slug, &g.Title, &g.Subtitle, &g.Description, &g.CoverKey, &g.BannerKey,
		&g.TrailerURL, &g.Genre, &g.Tags, &g.Engine, &g.Status, &g.ReleaseDate,
		&g.DeveloperID, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get game by id: %w", err)
	}
	return &g, nil
}

const getGameBySlugSQL = `
	SELECT id, slug, title, subtitle, description, cover_key, banner_key, trailer_url,
	       genre, tags, engine, status, release_date, developer_id, created_at, updated_at
	FROM games WHERE slug = $1
`

func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*game.Game, error) {
	var g game.Game
	err := r.pool.QueryRow(ctx, getGameBySlugSQL, slug).Scan(
		&g.ID, &g.Slug, &g.Title, &g.Subtitle, &g.Description, &g.CoverKey, &g.BannerKey,
		&g.TrailerURL, &g.Genre, &g.Tags, &g.Engine, &g.Status, &g.ReleaseDate,
		&g.DeveloperID, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get game by slug: %w", err)
	}
	return &g, nil
}

const updateGameSQL = `
	UPDATE games
	SET slug = $2, title = $3, subtitle = $4, description = $5, cover_key = $6,
	    banner_key = $7, trailer_url = $8, genre = $9, tags = $10, engine = $11,
	    status = $12, release_date = $13, developer_id = $14, updated_at = $15
	WHERE id = $1
`

func (r *GameRepository) Update(ctx context.Context, g *game.Game) error {
	_, err := r.pool.Exec(ctx, updateGameSQL,
		g.ID, g.Slug, g.Title, g.Subtitle, g.Description, g.CoverKey, g.BannerKey,
		g.TrailerURL, g.Genre, g.Tags, g.Engine, g.Status, g.ReleaseDate,
		g.DeveloperID, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	return nil
}

const deleteGameSQL = `DELETE FROM games WHERE id = $1`

func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteGameSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	if result.RowsAffected() == 0 {
		return game.ErrNotFound
	}

	return nil
}

func (r *GameRepository) List(ctx context.Context, filter game.ListFilter) ([]*game.Game, int64, error) {
	query := `
		SELECT id, slug, title, subtitle, description, cover_key, banner_key, trailer_url,
		       genre, tags, engine, status, release_date, developer_id, created_at, updated_at
		FROM games
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM games WHERE 1=1`
	args := []any{}
	argIndex := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Genre != "" {
		query += fmt.Sprintf(" AND $%d = ANY(genre)", argIndex)
		countQuery += fmt.Sprintf(" AND $%d = ANY(genre)", argIndex)
		args = append(args, filter.Genre)
		argIndex++
	}

	if filter.Tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argIndex)
		countQuery += fmt.Sprintf(" AND $%d = ANY(tags)", argIndex)
		args = append(args, filter.Tag)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (title ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex)
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count games: %w", err)
	}

	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list games: %w", err)
	}
	defer rows.Close()

	games := make([]*game.Game, 0)
	for rows.Next() {
		var g game.Game
		err := rows.Scan(
			&g.ID, &g.Slug, &g.Title, &g.Subtitle, &g.Description, &g.CoverKey, &g.BannerKey,
			&g.TrailerURL, &g.Genre, &g.Tags, &g.Engine, &g.Status, &g.ReleaseDate,
			&g.DeveloperID, &g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return games, total, nil
}

const existsBySlugSQL = `SELECT EXISTS(SELECT 1 FROM games WHERE slug = $1)`

func (r *GameRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, existsBySlugSQL, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}
	return exists, nil
}

