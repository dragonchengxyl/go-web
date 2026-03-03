package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
)

// UpdateProfileInput represents profile update input
type UpdateProfileInput struct {
	Username  *string `json:"username,omitempty"`
	AvatarKey *string `json:"avatar_key,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Website   *string `json:"website,omitempty"`
	Location  *string `json:"location,omitempty"`
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*user.User, error) {
	// Get current user
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Update fields if provided
	if input.Username != nil && *input.Username != u.Username {
		// Check if username is taken
		exists, err := s.userRepo.ExistsByUsername(ctx, *input.Username)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeInternalError, "检查用户名失败", err)
		}
		if exists {
			return nil, apperr.New(apperr.CodeUsernameExists, "用户名已被使用")
		}
		u.Username = *input.Username
	}

	if input.AvatarKey != nil {
		u.AvatarKey = input.AvatarKey
	}

	if input.Bio != nil {
		u.Bio = input.Bio
	}

	if input.Website != nil {
		u.Website = input.Website
	}

	if input.Location != nil {
		u.Location = input.Location
	}

	// Update timestamp
	u.UpdatedAt = time.Now()

	// Save to database
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新用户信息失败", err)
	}

	return u, nil
}

// GetProfile retrieves user profile
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}
	return u, nil
}

