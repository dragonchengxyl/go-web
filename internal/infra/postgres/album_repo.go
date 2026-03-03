package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/music"
)

type AlbumRepository struct {
	pool *pgxpool.Pool
}

func NewAlbumRepository(pool *pgxpool.Pool) *AlbumRepository {
	return &AlbumRepository{pool: pool}
}

const createAlbumSQL = `
	INSERT INTO albums (id, game_id, slug, title, subtitle, description, cover_key,
	                    artist, composer, arranger, lyricist, total_tracks, duration_sec,
	                    release_date, album_type, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`

func (r *AlbumRepository) Create(ctx context.Context, a *music.Album) error {
	_, err := r.pool.Exec(ctx, createAlbumSQL,
		a.ID, a.GameID, a.Slug, a.Title, a.Subtitle, a.Description, a.CoverKey,
		a.Artist, a.Composer, a.Arranger, a.Lyricist, a.TotalTracks, a.DurationSec,
		a.ReleaseDate, a.AlbumType, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create album: %w", err)
	}
	return nil
}

const getAlbumByIDSQL = `
	SELECT id, game_id, slug, title, subtitle, description, cover_key,
	       artist, composer, arranger, lyricist, total_tracks, duration_sec,
	       release_date, album_type, created_at, updated_at
	FROM albums WHERE id = $1
`

func (r *AlbumRepository) GetByID(ctx context.Context, id uuid.UUID) (*music.Album, error) {
	var a music.Album
	err := r.pool.QueryRow(ctx, getAlbumByIDSQL, id).Scan(
		&a.ID, &a.GameID, &a.Slug, &a.Title, &a.Subtitle, &a.Description, &a.CoverKey,
		&a.Artist, &a.Composer, &a.Arranger, &a.Lyricist, &a.TotalTracks, &a.DurationSec,
		&a.ReleaseDate, &a.AlbumType, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, music.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get album: %w", err)
	}
	return &a, nil
}

const getAlbumBySlugSQL = `
	SELECT id, game_id, slug, title, subtitle, description, cover_key,
	       artist, composer, arranger, lyricist, total_tracks, duration_sec,
	       release_date, album_type, created_at, updated_at
	FROM albums WHERE slug = $1
`

func (r *AlbumRepository) GetBySlug(ctx context.Context, slug string) (*music.Album, error) {
	var a music.Album
	err := r.pool.QueryRow(ctx, getAlbumBySlugSQL, slug).Scan(
		&a.ID, &a.GameID, &a.Slug, &a.Title, &a.Subtitle, &a.Description, &a.CoverKey,
		&a.Artist, &a.Composer, &a.Arranger, &a.Lyricist, &a.TotalTracks, &a.DurationSec,
		&a.ReleaseDate, &a.AlbumType, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, music.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get album by slug: %w", err)
	}
	return &a, nil
}

const updateAlbumSQL = `
	UPDATE albums
	SET game_id = $2, slug = $3, title = $4, subtitle = $5, description = $6, cover_key = $7,
	    artist = $8, composer = $9, arranger = $10, lyricist = $11, total_tracks = $12,
	    duration_sec = $13, release_date = $14, album_type = $15, updated_at = $16
	WHERE id = $1
`

func (r *AlbumRepository) Update(ctx context.Context, a *music.Album) error {
	_, err := r.pool.Exec(ctx, updateAlbumSQL,
		a.ID, a.GameID, a.Slug, a.Title, a.Subtitle, a.Description, a.CoverKey,
		a.Artist, a.Composer, a.Arranger, a.Lyricist, a.TotalTracks, a.DurationSec,
		a.ReleaseDate, a.AlbumType, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}
	return nil
}

const deleteAlbumSQL = `DELETE FROM albums WHERE id = $1`

func (r *AlbumRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteAlbumSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	if result.RowsAffected() == 0 {
		return music.ErrNotFound
	}

	return nil
}

func (r *AlbumRepository) List(ctx context.Context, filter music.AlbumListFilter) ([]*music.Album, int64, error) {
	query := `
		SELECT id, game_id, slug, title, subtitle, description, cover_key,
		       artist, composer, arranger, lyricist, total_tracks, duration_sec,
		       release_date, album_type, created_at, updated_at
		FROM albums
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM albums WHERE 1=1`
	args := []any{}
	argIndex := 1

	if filter.GameID != nil {
		query += fmt.Sprintf(" AND game_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND game_id = $%d", argIndex)
		args = append(args, *filter.GameID)
		argIndex++
	}

	if filter.AlbumType != nil {
		query += fmt.Sprintf(" AND album_type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND album_type = $%d", argIndex)
		args = append(args, *filter.AlbumType)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (title ILIKE $%d OR artist ILIKE $%d)", argIndex, argIndex)
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR artist ILIKE $%d)", argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count albums: %w", err)
	}

	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list albums: %w", err)
	}
	defer rows.Close()

	albums := make([]*music.Album, 0)
	for rows.Next() {
		var a music.Album
		err := rows.Scan(
			&a.ID, &a.GameID, &a.Slug, &a.Title, &a.Subtitle, &a.Description, &a.CoverKey,
			&a.Artist, &a.Composer, &a.Arranger, &a.Lyricist, &a.TotalTracks, &a.DurationSec,
			&a.ReleaseDate, &a.AlbumType, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan album: %w", err)
		}
		albums = append(albums, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return albums, total, nil
}

const albumExistsBySlugSQL = `SELECT EXISTS(SELECT 1 FROM albums WHERE slug = $1)`

func (r *AlbumRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, albumExistsBySlugSQL, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}
	return exists, nil
}
