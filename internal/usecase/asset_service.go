package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/asset"
	"github.com/studio/platform/internal/domain/game"
	"github.com/studio/platform/internal/pkg/apperr"
)

// AssetService handles user asset business logic
type AssetService struct {
	assetRepo   asset.Repository
	gameRepo    game.Repository
	branchRepo  game.BranchRepository
	releaseRepo game.ReleaseRepository
}

// NewAssetService creates a new AssetService
func NewAssetService(assetRepo asset.Repository, gameRepo game.Repository, branchRepo game.BranchRepository, releaseRepo game.ReleaseRepository) *AssetService {
	return &AssetService{
		assetRepo:   assetRepo,
		gameRepo:    gameRepo,
		branchRepo:  branchRepo,
		releaseRepo: releaseRepo,
	}
}

// GrantAssetInput represents input for granting an asset
type GrantAssetInput struct {
	GameID    uuid.UUID         `json:"game_id" binding:"required"`
	AssetType asset.AssetType   `json:"asset_type" binding:"required"`
	AssetID   uuid.UUID         `json:"asset_id" binding:"required"`
	Source    asset.AssetSource `json:"source" binding:"required"`
}

// GrantAsset grants an asset to a user
func (s *AssetService) GrantAsset(ctx context.Context, userID uuid.UUID, input GrantAssetInput) (*asset.UserGameAsset, error) {
	// Check if game exists
	_, err := s.gameRepo.GetByID(ctx, input.GameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}

	// Check if user already owns the asset
	hasAsset, err := s.assetRepo.HasAsset(ctx, userID, input.AssetType, input.AssetID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查资产所有权失败", err)
	}
	if hasAsset {
		return nil, apperr.New(apperr.CodeInvalidParam, "用户已拥有该资产")
	}

	// Grant asset
	now := time.Now()
	userAsset := &asset.UserGameAsset{
		ID:         uuid.New(),
		UserID:     userID,
		GameID:     input.GameID,
		AssetType:  input.AssetType,
		AssetID:    input.AssetID,
		ObtainedAt: now,
		Source:     input.Source,
		CreatedAt:  now,
	}

	if err := s.assetRepo.GrantAsset(ctx, userAsset); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "授予资产失败", err)
	}

	return userAsset, nil
}

// GetUserAssets retrieves all assets for a user
func (s *AssetService) GetUserAssets(ctx context.Context, userID uuid.UUID) ([]*asset.UserGameAsset, error) {
	assets, err := s.assetRepo.GetUserAssets(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户资产失败", err)
	}
	return assets, nil
}

// GetUserGameAssets retrieves assets for a specific game
func (s *AssetService) GetUserGameAssets(ctx context.Context, userID, gameID uuid.UUID) ([]*asset.UserGameAsset, error) {
	assets, err := s.assetRepo.GetUserGameAssets(ctx, userID, gameID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户游戏资产失败", err)
	}
	return assets, nil
}

// HasAsset checks if user owns a specific asset
func (s *AssetService) HasAsset(ctx context.Context, userID uuid.UUID, assetType asset.AssetType, assetID uuid.UUID) (bool, error) {
	hasAsset, err := s.assetRepo.HasAsset(ctx, userID, assetType, assetID)
	if err != nil {
		return false, apperr.Wrap(apperr.CodeInternalError, "检查资产所有权失败", err)
	}
	return hasAsset, nil
}

// RevokeAsset revokes an asset from a user
func (s *AssetService) RevokeAsset(ctx context.Context, userID uuid.UUID, assetType asset.AssetType, assetID uuid.UUID) error {
	if err := s.assetRepo.RevokeAsset(ctx, userID, assetType, assetID); err != nil {
		if errors.Is(err, asset.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "撤销资产失败", err)
	}
	return nil
}

// CheckDownloadPermission checks if user can download a release
func (s *AssetService) CheckDownloadPermission(ctx context.Context, userID, releaseID uuid.UUID) error {
	// Get release
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询版本失败", err)
	}

	// Check if release is published
	if !release.IsPublished {
		return apperr.New(apperr.CodeForbidden, "版本未发布")
	}

	// Get branch to find game ID
	branch, err := s.branchRepo.GetByID(ctx, release.BranchID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "查询分支失败", err)
	}

	// Check if user owns the game
	hasGame, err := s.assetRepo.HasAsset(ctx, userID, asset.AssetTypeBaseGame, branch.GameID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "检查游戏所有权失败", err)
	}
	if !hasGame {
		return apperr.New(apperr.CodeForbidden, "用户未拥有该游戏")
	}

	// Check download limit (3 times per day)
	dayAgo := time.Now().Add(-24 * time.Hour).Unix()
	count, err := s.assetRepo.GetDownloadCount(ctx, userID, releaseID, dayAgo)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "检查下载次数失败", err)
	}
	if count >= 3 {
		return apperr.New(apperr.CodeForbidden, "今日下载次数已达上限")
	}

	return nil
}

// LogDownload logs a download attempt
func (s *AssetService) LogDownload(ctx context.Context, userID, releaseID uuid.UUID, clientIP, userAgent string) error {
	now := time.Now()
	log := &asset.DownloadLog{
		ID:           uuid.New(),
		UserID:       userID,
		ReleaseID:    releaseID,
		ClientIP:     clientIP,
		UserAgent:    &userAgent,
		DownloadedAt: now,
		CreatedAt:    now,
	}

	if err := s.assetRepo.LogDownload(ctx, log); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "记录下载日志失败", err)
	}

	return nil
}
