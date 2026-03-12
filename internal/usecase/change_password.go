package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
)

// ChangePasswordInput holds the change-password parameters
type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

// ChangePassword verifies the old password and updates it to a new one
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	valid, err := crypto.VerifyPassword(input.OldPassword, u.PasswordHash)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "密码验证失败", err)
	}
	if !valid {
		return apperr.New(apperr.CodeUnauthorized, "当前密码错误")
	}

	newHash, err := crypto.HashPassword(input.NewPassword)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "密码加密失败", err)
	}

	u.PasswordHash = newHash
	u.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, u); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新密码失败", err)
	}

	return nil
}
