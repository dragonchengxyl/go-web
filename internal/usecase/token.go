package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
)

// RefreshToken refreshes access token using refresh token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken, device, ip string) (*TokenOutput, error) {
	// Verify refresh token
	claims, err := crypto.VerifyToken(refreshToken, s.jwtConfig.Secret)
	if err != nil {
		return nil, apperr.New(apperr.CodeTokenExpired, "刷新令牌无效")
	}

	// Check if token is blacklisted
	blacklisted, err := s.tokenStore.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查令牌黑名单失败", err)
	}
	if blacklisted {
		return nil, apperr.New(apperr.CodeUnauthorized, "令牌已被撤销")
	}

	// Verify refresh token exists in Redis
	_, err = s.tokenStore.GetRefreshToken(ctx, claims.JTI)
	if err != nil {
		return nil, apperr.New(apperr.CodeUnauthorized, "刷新令牌不存在")
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "解析用户ID失败", err)
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Delete old refresh token
	if err := s.tokenStore.DeleteRefreshToken(ctx, claims.JTI); err != nil {
		// Log but don't fail
	}

	// Generate new tokens
	tokens, err := s.generateTokens(ctx, u, device, ip)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Logout logs out a user by blacklisting the access token
func (s *UserService) Logout(ctx context.Context, accessToken string) error {
	// Verify token
	claims, err := crypto.VerifyToken(accessToken, s.jwtConfig.Secret)
	if err != nil {
		// Token invalid, consider already logged out
		return nil
	}

	// Blacklist access token
	if err := s.tokenStore.BlacklistToken(ctx, claims.JTI, s.jwtConfig.AccessExpiry); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "撤销令牌失败", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}
	return u, nil
}
