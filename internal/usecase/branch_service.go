package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/game"
	"github.com/studio/platform/internal/pkg/apperr"
)

// BranchService handles game branch business logic
type BranchService struct {
	branchRepo game.BranchRepository
	gameRepo   game.Repository
}

// NewBranchService creates a new BranchService
func NewBranchService(branchRepo game.BranchRepository, gameRepo game.Repository) *BranchService {
	return &BranchService{
		branchRepo: branchRepo,
		gameRepo:   gameRepo,
	}
}

// CreateBranchInput represents input for creating a branch
type CreateBranchInput struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	IsDefault   bool    `json:"is_default"`
}

// CreateBranch creates a new game branch
func (s *BranchService) CreateBranch(ctx context.Context, gameID uuid.UUID, input CreateBranchInput) (*game.Branch, error) {
	// Check if game exists
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}

	// Create branch
	branch := &game.Branch{
		ID:          uuid.New(),
		GameID:      gameID,
		Name:        input.Name,
		Description: input.Description,
		IsDefault:   input.IsDefault,
		CreatedAt:   time.Now(),
	}

	if err := s.branchRepo.Create(ctx, branch); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建分支失败", err)
	}

	return branch, nil
}

// GetBranch retrieves a branch by ID
func (s *BranchService) GetBranch(ctx context.Context, branchID uuid.UUID) (*game.Branch, error) {
	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询分支失败", err)
	}
	return branch, nil
}

// ListBranches retrieves branches for a game
func (s *BranchService) ListBranches(ctx context.Context, gameID uuid.UUID) ([]*game.Branch, error) {
	branches, err := s.branchRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询分支列表失败", err)
	}
	return branches, nil
}

// UpdateBranchInput represents input for updating a branch
type UpdateBranchInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

// UpdateBranch updates a branch
func (s *BranchService) UpdateBranch(ctx context.Context, branchID uuid.UUID, input UpdateBranchInput) (*game.Branch, error) {
	// Get branch
	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询分支失败", err)
	}

	// Update fields
	if input.Name != nil {
		branch.Name = *input.Name
	}
	if input.Description != nil {
		branch.Description = input.Description
	}
	if input.IsDefault != nil {
		branch.IsDefault = *input.IsDefault
	}

	// Save
	if err := s.branchRepo.Update(ctx, branch); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新分支失败", err)
	}

	return branch, nil
}

// DeleteBranch deletes a branch
func (s *BranchService) DeleteBranch(ctx context.Context, branchID uuid.UUID) error {
	if err := s.branchRepo.Delete(ctx, branchID); err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "删除分支失败", err)
	}
	return nil
}
