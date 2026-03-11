package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
)

// ForgotPassword generates a password-reset token and stores it in Redis.
// It always returns success to prevent email enumeration.
func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil // silent success — don't reveal whether email exists
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	token, err := generateSecureToken()
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "生成重置令牌失败", err)
	}

	if err := s.tokenStore.SaveResetToken(ctx, token, u.ID.String()); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "保存重置令牌失败", err)
	}

	// Email sending is intentionally omitted here — wire an email service
	// at the handler/infrastructure layer if SMTP is configured.
	// For local dev the token can be retrieved from Redis via: GET auth:reset:<token>
	fmt.Printf("[ForgotPassword] reset token for %s: %s\n", email, token)

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := s.tokenStore.GetResetToken(ctx, token)
	if err != nil {
		return apperr.New(apperr.CodeUnauthorized, "重置链接无效或已过期")
	}

	u, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "密码加密失败", err)
	}
	u.PasswordHash = newHash

	if err := s.userRepo.Update(ctx, u); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新密码失败", err)
	}

	// Consume the token so it can't be reused
	_ = s.tokenStore.DeleteResetToken(ctx, token)

	return nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
