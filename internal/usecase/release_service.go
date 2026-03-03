package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/game"
	"github.com/studio/platform/internal/pkg/apperr"
)

// ReleaseService handles game release business logic
type ReleaseService struct {
	releaseRepo game.ReleaseRepository
	branchRepo  game.BranchRepository
}

// NewReleaseService creates a new ReleaseService
func NewReleaseService(releaseRepo game.ReleaseRepository, branchRepo game.BranchRepository) *ReleaseService {
	return &ReleaseService{
		releaseRepo: releaseRepo,
		branchRepo:  branchRepo,
	}
}

// CreateReleaseInput represents input for creating a release
type CreateReleaseInput struct {
	Version        string  `json:"version" binding:"required"`
	Title          *string `json:"title,omitempty"`
	Changelog      *string `json:"changelog,omitempty"`
	OSSKey         *string `json:"oss_key,omitempty"`
	ManifestKey    *string `json:"manifest_key,omitempty"`
	FileSize       *int64  `json:"file_size,omitempty"`
	ChecksumSHA256 *string `json:"checksum_sha256,omitempty"`
	MinOSVersion   *string `json:"min_os_version,omitempty"`
}

// CreateRelease creates a new game release
func (s *ReleaseService) CreateRelease(ctx context.Context, branchID, creatorID uuid.UUID, input CreateReleaseInput) (*game.Release, error) {
	// Check if branch exists
	_, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询分支失败", err)
	}

	// Check if version already exists
	existing, err := s.releaseRepo.GetByVersion(ctx, branchID, input.Version)
	if err != nil && !errors.Is(err, game.ErrNotFound) {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查版本失败", err)
	}
	if existing != nil {
		return nil, apperr.New(apperr.CodeInvalidParam, "版本号已存在")
	}

	// Create release
	release := &game.Release{
		ID:             uuid.New(),
		BranchID:       branchID,
		Version:        input.Version,
		Title:          input.Title,
		Changelog:      input.Changelog,
		OSSKey:         input.OSSKey,
		ManifestKey:    input.ManifestKey,
		FileSize:       input.FileSize,
		ChecksumSHA256: input.ChecksumSHA256,
		MinOSVersion:   input.MinOSVersion,
		IsPublished:    false,
		CreatedBy:      creatorID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.releaseRepo.Create(ctx, release); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建版本失败", err)
	}

	return release, nil
}

// GetRelease retrieves a release by ID
func (s *ReleaseService) GetRelease(ctx context.Context, releaseID uuid.UUID) (*game.Release, error) {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询版本失败", err)
	}
	return release, nil
}

// ListReleases retrieves releases for a branch
func (s *ReleaseService) ListReleases(ctx context.Context, branchID uuid.UUID, publishedOnly bool) ([]*game.Release, error) {
	releases, err := s.releaseRepo.ListByBranchID(ctx, branchID, publishedOnly)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询版本列表失败", err)
	}
	return releases, nil
}

// UpdateReleaseInput represents input for updating a release
type UpdateReleaseInput struct {
	Title          *string `json:"title,omitempty"`
	Changelog      *string `json:"changelog,omitempty"`
	OSSKey         *string `json:"oss_key,omitempty"`
	ManifestKey    *string `json:"manifest_key,omitempty"`
	FileSize       *int64  `json:"file_size,omitempty"`
	ChecksumSHA256 *string `json:"checksum_sha256,omitempty"`
	MinOSVersion   *string `json:"min_os_version,omitempty"`
}

// UpdateRelease updates a release
func (s *ReleaseService) UpdateRelease(ctx context.Context, releaseID uuid.UUID, input UpdateReleaseInput) (*game.Release, error) {
	// Get release
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询版本失败", err)
	}

	// Update fields
	if input.Title != nil {
		release.Title = input.Title
	}
	if input.Changelog != nil {
		release.Changelog = input.Changelog
	}
	if input.OSSKey != nil {
		release.OSSKey = input.OSSKey
	}
	if input.ManifestKey != nil {
		release.ManifestKey = input.ManifestKey
	}
	if input.FileSize != nil {
		release.FileSize = input.FileSize
	}
	if input.ChecksumSHA256 != nil {
		release.ChecksumSHA256 = input.ChecksumSHA256
	}
	if input.MinOSVersion != nil {
		release.MinOSVersion = input.MinOSVersion
	}

	release.UpdatedAt = time.Now()

	// Save
	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新版本失败", err)
	}

	return release, nil
}

// PublishRelease publishes a release
func (s *ReleaseService) PublishRelease(ctx context.Context, releaseID uuid.UUID) (*game.Release, error) {
	// Get release
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询版本失败", err)
	}

	if release.IsPublished {
		return nil, apperr.New(apperr.CodeInvalidParam, "版本已发布")
	}

	// Publish
	now := time.Now()
	release.IsPublished = true
	release.PublishedAt = &now
	release.UpdatedAt = now

	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "发布版本失败", err)
	}

	return release, nil
}

// UnpublishRelease unpublishes a release
func (s *ReleaseService) UnpublishRelease(ctx context.Context, releaseID uuid.UUID) (*game.Release, error) {
	// Get release
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询版本失败", err)
	}

	if !release.IsPublished {
		return nil, apperr.New(apperr.CodeInvalidParam, "版本未发布")
	}

	// Unpublish
	release.IsPublished = false
	release.PublishedAt = nil
	release.UpdatedAt = time.Now()

	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "取消发布失败", err)
	}

	return release, nil
}

// DeleteRelease deletes a release
func (s *ReleaseService) DeleteRelease(ctx context.Context, releaseID uuid.UUID) error {
	if err := s.releaseRepo.Delete(ctx, releaseID); err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "删除版本失败", err)
	}
	return nil
}

// GetLatestRelease retrieves the latest published release for a branch
func (s *ReleaseService) GetLatestRelease(ctx context.Context, branchID uuid.UUID) (*game.Release, error) {
	release, err := s.releaseRepo.GetLatestRelease(ctx, branchID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询最新版本失败", err)
	}
	return release, nil
}
