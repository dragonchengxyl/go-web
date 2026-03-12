package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
)

// Register creates a user and returns an authenticated session.
func (s *UserService) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	// Validate email format
	if err := crypto.ValidateEmail(input.Email); err != nil {
		return nil, apperr.New(apperr.CodeInvalidParam, err.Error())
	}

	// Validate username format
	if err := crypto.ValidateUsername(input.Username); err != nil {
		return nil, apperr.New(apperr.CodeInvalidParam, err.Error())
	}

	// Validate password strength
	if err := crypto.ValidatePassword(input.Password, crypto.DefaultPasswordStrength()); err != nil {
		return nil, apperr.New(apperr.CodeInvalidParam, err.Error())
	}

	// Check if password is too common
	if crypto.IsCommonPassword(input.Password) {
		return nil, apperr.New(apperr.CodeInvalidParam, "密码过于简单，请使用更复杂的密码")
	}

	// Check if email exists
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查邮箱失败", err)
	}
	if exists {
		return nil, apperr.ErrEmailExists
	}

	// Check if username exists
	exists, err = s.userRepo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查用户名失败", err)
	}
	if exists {
		return nil, apperr.ErrUsernameExists
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "密码加密失败", err)
	}

	// Create user
	now := time.Now()
	u := &user.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
		Role:         user.RoleMember,
		Status:       user.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建用户失败", err)
	}

	tokens, err := s.generateTokens(ctx, u, input.Device, input.IP)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		User:   u,
		Tokens: tokens,
	}, nil
}
