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

type ReleaseRepository struct {
	pool *pgxpool.Pool
}

func NewReleaseRepository(pool *pgxpool.Pool) *ReleaseRepository {
	return &ReleaseRepository{pool: pool}
}

const createReleaseSQL = `
	INSERT INTO game_releases (id, branch_id, version, title, changelog, oss_key, manifest_key,
	                           file_size, checksum_sha256, min_os_version, is_published,
	                           published_at, created_by, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

func (r *ReleaseRepository) Create(ctx context.Context, rel *game.Release) error {
	_, err := r.pool.Exec(ctx, createReleaseSQL,
		rel.ID, rel.BranchID, rel.Version, rel.Title, rel.Changelog, rel.OSSKey, rel.ManifestKey,
		rel.FileSize, rel.ChecksumSHA256, rel.MinOSVersion, rel.IsPublished,
		rel.PublishedAt, rel.CreatedBy, rel.CreatedAt, rel.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}
	return nil
}

const getReleaseByIDSQL = `
	SELECT id, branch_id, version, title, changelog, oss_key, manifest_key,
	       file_size, checksum_sha256, min_os_version, is_published,
	       published_at, created_by, created_at, updated_at
	FROM game_releases WHERE id = $1
`

func (r *ReleaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*game.Release, error) {
	var rel game.Release
	err := r.pool.QueryRow(ctx, getReleaseByIDSQL, id).Scan(
		&rel.ID, &rel.BranchID, &rel.Version, &rel.Title, &rel.Changelog, &rel.OSSKey, &rel.ManifestKey,
		&rel.FileSize, &rel.ChecksumSHA256, &rel.MinOSVersion, &rel.IsPublished,
		&rel.PublishedAt, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get release: %w", err)
	}
	return &rel, nil
}

const getReleaseByVersionSQL = `
	SELECT id, branch_id, version, title, changelog, oss_key, manifest_key,
	       file_size, checksum_sha256, min_os_version, is_published,
	       published_at, created_by, created_at, updated_at
	FROM game_releases WHERE branch_id = $1 AND version = $2
`

func (r *ReleaseRepository) GetByVersion(ctx context.Context, branchID uuid.UUID, version string) (*game.Release, error) {
	var rel game.Release
	err := r.pool.QueryRow(ctx, getReleaseByVersionSQL, branchID, version).Scan(
		&rel.ID, &rel.BranchID, &rel.Version, &rel.Title, &rel.Changelog, &rel.OSSKey, &rel.ManifestKey,
		&rel.FileSize, &rel.ChecksumSHA256, &rel.MinOSVersion, &rel.IsPublished,
		&rel.PublishedAt, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get release by version: %w", err)
	}
	return &rel, nil
}

func (r *ReleaseRepository) ListByBranchID(ctx context.Context, branchID uuid.UUID, publishedOnly bool) ([]*game.Release, error) {
	query := `
		SELECT id, branch_id, version, title, changelog, oss_key, manifest_key,
		       file_size, checksum_sha256, min_os_version, is_published,
		       published_at, created_by, created_at, updated_at
		FROM game_releases WHERE branch_id = $1
	`
	if publishedOnly {
		query += " AND is_published = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}
	defer rows.Close()

	releases := make([]*game.Release, 0)
	for rows.Next() {
		var rel game.Release
		err := rows.Scan(
			&rel.ID, &rel.BranchID, &rel.Version, &rel.Title, &rel.Changelog, &rel.OSSKey, &rel.ManifestKey,
			&rel.FileSize, &rel.ChecksumSHA256, &rel.MinOSVersion, &rel.IsPublished,
			&rel.PublishedAt, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan release: %w", err)
		}
		releases = append(releases, &rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return releases, nil
}

const updateReleaseSQL = `
	UPDATE game_releases
	SET version = $2, title = $3, changelog = $4, oss_key = $5, manifest_key = $6,
	    file_size = $7, checksum_sha256 = $8, min_os_version = $9, is_published = $10,
	    published_at = $11, updated_at = $12
	WHERE id = $1
`

func (r *ReleaseRepository) Update(ctx context.Context, rel *game.Release) error {
	_, err := r.pool.Exec(ctx, updateReleaseSQL,
		rel.ID, rel.Version, rel.Title, rel.Changelog, rel.OSSKey, rel.ManifestKey,
		rel.FileSize, rel.ChecksumSHA256, rel.MinOSVersion, rel.IsPublished,
		rel.PublishedAt, rel.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update release: %w", err)
	}
	return nil
}

const deleteReleaseSQL = `DELETE FROM game_releases WHERE id = $1`

func (r *ReleaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteReleaseSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete release: %w", err)
	}

	if result.RowsAffected() == 0 {
		return game.ErrNotFound
	}

	return nil
}

const getLatestReleaseSQL = `
	SELECT id, branch_id, version, title, changelog, oss_key, manifest_key,
	       file_size, checksum_sha256, min_os_version, is_published,
	       published_at, created_by, created_at, updated_at
	FROM game_releases
	WHERE branch_id = $1 AND is_published = true
	ORDER BY created_at DESC
	LIMIT 1
`

func (r *ReleaseRepository) GetLatestRelease(ctx context.Context, branchID uuid.UUID) (*game.Release, error) {
	var rel game.Release
	err := r.pool.QueryRow(ctx, getLatestReleaseSQL, branchID).Scan(
		&rel.ID, &rel.BranchID, &rel.Version, &rel.Title, &rel.Changelog, &rel.OSSKey, &rel.ManifestKey,
		&rel.FileSize, &rel.ChecksumSHA256, &rel.MinOSVersion, &rel.IsPublished,
		&rel.PublishedAt, &rel.CreatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}
	return &rel, nil
}
