package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/asset"
)

type AssetRepository struct {
	pool *pgxpool.Pool
}

func NewAssetRepository(pool *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{pool: pool}
}

const grantAssetSQL = `
	INSERT INTO user_game_assets (id, user_id, game_id, asset_type, asset_id, obtained_at, source, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (user_id, asset_type, asset_id) DO NOTHING
`

func (r *AssetRepository) GrantAsset(ctx context.Context, a *asset.UserGameAsset) error {
	_, err := r.pool.Exec(ctx, grantAssetSQL,
		a.ID, a.UserID, a.GameID, a.AssetType, a.AssetID, a.ObtainedAt, a.Source, a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to grant asset: %w", err)
	}
	return nil
}

const getUserAssetsSQL = `
	SELECT id, user_id, game_id, asset_type, asset_id, obtained_at, source, created_at
	FROM user_game_assets
	WHERE user_id = $1
	ORDER BY obtained_at DESC
`

func (r *AssetRepository) GetUserAssets(ctx context.Context, userID uuid.UUID) ([]*asset.UserGameAsset, error) {
	rows, err := r.pool.Query(ctx, getUserAssetsSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assets: %w", err)
	}
	defer rows.Close()

	assets := make([]*asset.UserGameAsset, 0)
	for rows.Next() {
		var a asset.UserGameAsset
		err := rows.Scan(&a.ID, &a.UserID, &a.GameID, &a.AssetType, &a.AssetID, &a.ObtainedAt, &a.Source, &a.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return assets, nil
}

const getUserGameAssetsSQL = `
	SELECT id, user_id, game_id, asset_type, asset_id, obtained_at, source, created_at
	FROM user_game_assets
	WHERE user_id = $1 AND game_id = $2
	ORDER BY obtained_at DESC
`

func (r *AssetRepository) GetUserGameAssets(ctx context.Context, userID, gameID uuid.UUID) ([]*asset.UserGameAsset, error) {
	rows, err := r.pool.Query(ctx, getUserGameAssetsSQL, userID, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user game assets: %w", err)
	}
	defer rows.Close()

	assets := make([]*asset.UserGameAsset, 0)
	for rows.Next() {
		var a asset.UserGameAsset
		err := rows.Scan(&a.ID, &a.UserID, &a.GameID, &a.AssetType, &a.AssetID, &a.ObtainedAt, &a.Source, &a.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return assets, nil
}

const hasAssetSQL = `
	SELECT EXISTS(
		SELECT 1 FROM user_game_assets
		WHERE user_id = $1 AND asset_type = $2 AND asset_id = $3
	)
`

func (r *AssetRepository) HasAsset(ctx context.Context, userID uuid.UUID, assetType asset.AssetType, assetID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, hasAssetSQL, userID, assetType, assetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check asset ownership: %w", err)
	}
	return exists, nil
}

const revokeAssetSQL = `
	DELETE FROM user_game_assets
	WHERE user_id = $1 AND asset_type = $2 AND asset_id = $3
`

func (r *AssetRepository) RevokeAsset(ctx context.Context, userID uuid.UUID, assetType asset.AssetType, assetID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, revokeAssetSQL, userID, assetType, assetID)
	if err != nil {
		return fmt.Errorf("failed to revoke asset: %w", err)
	}

	if result.RowsAffected() == 0 {
		return asset.ErrNotFound
	}

	return nil
}

const logDownloadSQL = `
	INSERT INTO download_logs (id, user_id, release_id, client_ip, user_agent, downloaded_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`

func (r *AssetRepository) LogDownload(ctx context.Context, log *asset.DownloadLog) error {
	_, err := r.pool.Exec(ctx, logDownloadSQL,
		log.ID, log.UserID, log.ReleaseID, log.ClientIP, log.UserAgent, log.DownloadedAt, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to log download: %w", err)
	}
	return nil
}

const getDownloadCountSQL = `
	SELECT COUNT(*)
	FROM download_logs
	WHERE user_id = $1 AND release_id = $2 AND EXTRACT(EPOCH FROM downloaded_at) >= $3
`

func (r *AssetRepository) GetDownloadCount(ctx context.Context, userID, releaseID uuid.UUID, since int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, getDownloadCountSQL, userID, releaseID, since).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get download count: %w", err)
	}
	return count, nil
}
