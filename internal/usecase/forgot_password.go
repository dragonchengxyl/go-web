package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

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

	resetURL := strings.TrimRight(s.frontendURL, "/") + "/reset-password?token=" + token
	if s.emailSender != nil && s.frontendURL != "" {
		if err := s.emailSender.SendPasswordReset(u.Email, u.Username, resetURL); err != nil {
			fmt.Printf("[ForgotPassword] failed to send reset email to %s: %v\n", email, err)
			fmt.Printf("[ForgotPassword] fallback reset url for %s: %s\n", email, resetURL)
		}
	} else {
		// Local/dev fallback when SMTP or frontend URL is not configured.
		fmt.Printf("[ForgotPassword] reset url for %s: %s\n", email, resetURL)
	}

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
	u.UpdatedAt = time.Now()

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

func (s *UserService) sendVerificationEmail(ctx context.Context, u *user.User) error {
	if u.EmailVerifiedAt != nil {
		return nil
	}

	token, err := generateSecureToken()
	if err != nil {
		return err
	}
	if err := s.tokenStore.SaveVerifyEmailToken(ctx, token, u.ID.String()); err != nil {
		return err
	}

	verifyURL := strings.TrimRight(s.frontendURL, "/") + "/verify-email?token=" + token
	if s.emailSender != nil && s.frontendURL != "" {
		if err := s.emailSender.SendEmailVerification(u.Email, u.Username, verifyURL); err != nil {
			fmt.Printf("[VerifyEmail] failed to send verification email to %s: %v\n", u.Email, err)
			fmt.Printf("[VerifyEmail] fallback verification url for %s: %s\n", u.Email, verifyURL)
		}
	} else {
		fmt.Printf("[VerifyEmail] verification url for %s: %s\n", u.Email, verifyURL)
	}

	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.tokenStore.GetVerifyEmailToken(ctx, token)
	if err != nil {
		return apperr.New(apperr.CodeUnauthorized, "验证链接无效或已过期")
	}

	u, err := s.userRepo.GetByID(ctx, uuid.MustParse(userID))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	now := time.Now()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
	if err := s.userRepo.Update(ctx, u); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新邮箱验证状态失败", err)
	}

	_ = s.tokenStore.DeleteVerifyEmailToken(ctx, token)
	return nil
}

func (s *UserService) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}
	if u.EmailVerifiedAt != nil {
		return nil
	}
	return s.sendVerificationEmail(ctx, u)
}
