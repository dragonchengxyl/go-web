package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/studio/platform/internal/domain/permission"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
)

// Login authenticates a user and returns tokens
func (s *UserService) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	// Get user by email
	u, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.New(apperr.CodeUnauthorized, "邮箱或密码错误")
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Verify password
	valid, err := crypto.VerifyPassword(input.Password, u.PasswordHash)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "密码验证失败", err)
	}
	if !valid {
		return nil, apperr.New(apperr.CodeUnauthorized, "邮箱或密码错误")
	}

	// Check account status
	if u.Status == user.StatusSuspended {
		return nil, apperr.New(apperr.CodeForbidden, "账号已被暂停")
	}
	if u.Status == user.StatusBanned {
		return nil, apperr.New(apperr.CodeForbidden, "账号已被封禁")
	}

	// Generate tokens
	tokens, err := s.generateTokens(ctx, u, input.Device, input.IP)
	if err != nil {
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, u.ID, input.IP); err != nil {
		// Log error but don't fail login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	return &AuthOutput{
		User:   u,
		Tokens: tokens,
	}, nil
}

// generateTokens generates access and refresh tokens
func (s *UserService) generateTokens(ctx context.Context, u *user.User, device, ip string) (*TokenOutput, error) {
	userID := u.ID.String()
	permissions := permission.GetPermissionStrings(u.Role)

	// Generate access token
	accessToken, _, err := crypto.GenerateToken(
		userID,
		string(u.Role),
		permissions,
		s.jwtConfig.Secret,
		s.jwtConfig.AccessExpiry,
	)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "生成访问令牌失败", err)
	}

	// Generate refresh token
	refreshToken, refreshJTI, err := crypto.GenerateToken(
		userID,
		string(u.Role),
		permissions,
		s.jwtConfig.Secret,
		s.jwtConfig.RefreshExpiry,
	)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "生成刷新令牌失败", err)
	}

	// Save refresh token to Redis
	if err := s.tokenStore.SaveRefreshToken(ctx, refreshJTI, userID, device, ip, s.jwtConfig.RefreshExpiry); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "保存刷新令牌失败", err)
	}

	return &TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessExpiry.Seconds()),
	}, nil
}
