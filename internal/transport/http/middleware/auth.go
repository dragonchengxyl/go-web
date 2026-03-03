package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/permission"
	"github.com/studio/platform/internal/domain/user"
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

// RequirePermission checks if user has required permission
func (a *Auth) RequirePermission(perm permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context
		roleStr, exists := c.Get("role")
		if !exists {
			response.Error(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		role := user.Role(roleStr.(string))

		// Check permission
		if !permission.HasPermission(role, perm) {
			response.Error(c, apperr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has required role or higher
func (a *Auth) RequireRole(requiredRole user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context
		roleStr, exists := c.Get("role")
		if !exists {
			response.Error(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		role := user.Role(roleStr.(string))

		// Check role hierarchy
		if !hasRoleOrHigher(role, requiredRole) {
			response.Error(c, apperr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasRoleOrHigher checks if user role is equal or higher than required role
func hasRoleOrHigher(userRole, requiredRole user.Role) bool {
	roleHierarchy := map[user.Role]int{
		user.RoleSuperAdmin: 7,
		user.RoleAdmin:      6,
		user.RoleModerator:  5,
		user.RoleCreator:    4,
		user.RolePremium:    3,
		user.RolePlayer:     2,
		user.RoleGuest:      1,
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil
	}
	return userID
}

