package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/crypto"
	"github.com/studio/platform/internal/pkg/response"
)

// Auth authenticates requests using JWT
type Auth struct {
	jwtConfig  configs.JWTConfig
	tokenStore *redis.TokenStore
}

// NewAuth creates a new auth middleware
func NewAuth(jwtConfig configs.JWTConfig, tokenStore *redis.TokenStore) *Auth {
	return &Auth{
		jwtConfig:  jwtConfig,
		tokenStore: tokenStore,
	}
}

// Authenticate validates JWT token
func (a *Auth) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify token
		claims, err := crypto.VerifyToken(tokenString, a.jwtConfig.Secret)
		if err != nil {
			response.Error(c, apperr.ErrTokenExpired)
			c.Abort()
			return
		}

		// Check if token is blacklisted
		blacklisted, err := a.tokenStore.IsTokenBlacklisted(c.Request.Context(), claims.JTI)
		if err != nil {
			response.Error(c, apperr.ErrInternalError)
			c.Abort()
			return
		}
		if blacklisted {
			response.Error(c, apperr.New(apperr.CodeUnauthorized, "令牌已被撤销"))
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}
