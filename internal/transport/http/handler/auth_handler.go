package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/transport/http/dto"
	"github.com/studio/platform/internal/usecase"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService *usecase.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *usecase.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
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

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from Authorization header
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
