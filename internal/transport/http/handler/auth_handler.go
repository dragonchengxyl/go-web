package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/transport/http/dto"
	"github.com/studio/platform/internal/usecase"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService *usecase.UserService
	rdb         *redis.Client
}

// NewAuthHandler creates a new auth handler. Pass an optional *redis.Client to enable IP rate limiting.
func NewAuthHandler(userService *usecase.UserService, rdb ...*redis.Client) *AuthHandler {
	h := &AuthHandler{userService: userService}
	if len(rdb) > 0 {
		h.rdb = rdb[0]
	}
	return h
}

// checkIPRateLimit enforces a per-IP rate limit for sensitive endpoints using Redis.
// Returns true if the caller is rate-limited and should be rejected.
func (h *AuthHandler) checkIPRateLimit(c *gin.Context, action string, max int, window time.Duration) bool {
	if h.rdb == nil {
		return false
	}
	key := fmt.Sprintf("ratelimit:%s:ip:%s", action, c.ClientIP())
	ctx := context.Background()
	count, err := h.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false // fail open
	}
	if count == 1 {
		h.rdb.Expire(ctx, key, window) //nolint:errcheck
	}
	return count > int64(max)
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	// IP rate limit: max 5 register attempts per IP per hour
	if h.checkIPRateLimit(c, "register", 5, time.Hour) {
		response.Error(c, apperr.ErrRateLimited)
		return
	}

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.ErrValidationFailed.WithDetail(err.Error()))
		return
	}

	input := usecase.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	u, err := h.userService.Register(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, toUserResponse(u))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.ErrValidationFailed.WithDetail(err.Error()))
		return
	}

	input := usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	}

	output, err := h.userService.Login(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, dto.AuthResponse{
		User:         toUserResponse(output.User),
		AccessToken:  output.Tokens.AccessToken,
		RefreshToken: output.Tokens.RefreshToken,
		ExpiresIn:    output.Tokens.ExpiresIn,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.ErrValidationFailed.WithDetail(err.Error()))
		return
	}

	tokens, err := h.userService.RefreshToken(
		c.Request.Context(),
		req.RefreshToken,
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, tokens)
}

// ChangePassword handles password change for authenticated users
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.ErrValidationFailed.WithDetail(err.Error()))
		return
	}
	if err := h.userService.ChangePassword(c.Request.Context(), uid, usecase.ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "密码修改成功"})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	if err := h.userService.Logout(c.Request.Context(), parts[1]); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "登出成功"})
}

// toUserResponse converts user entity to response DTO
func toUserResponse(u *user.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.AvatarKey,
		Bio:      u.Bio,
		Location: u.Location,
		Role:     string(u.Role),
		Status:   string(u.Status),
	}
}
