package asset

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for user asset data access
type Repository interface {
	// GrantAsset grants an asset to a user
	GrantAsset(ctx context.Context, asset *UserGameAsset) error

	// GetUserAssets retrieves all assets for a user
	GetUserAssets(ctx context.Context, userID uuid.UUID) ([]*UserGameAsset, error)

	// GetUserGameAssets retrieves assets for a specific game
	GetUserGameAssets(ctx context.Context, userID, gameID uuid.UUID) ([]*UserGameAsset, error)

	// HasAsset checks if user owns a specific asset
	HasAsset(ctx context.Context, userID uuid.UUID, assetType AssetType, assetID uuid.UUID) (bool, error)

	// RevokeAsset revokes an asset from a user
	RevokeAsset(ctx context.Context, userID uuid.UUID, assetType AssetType, assetID uuid.UUID) error

	// LogDownload logs a download attempt
	LogDownload(ctx context.Context, log *DownloadLog) error

	// GetDownloadCount gets download count for a user and release within a time period
	GetDownloadCount(ctx context.Context, userID, releaseID uuid.UUID, since int64) (int64, error)
}
