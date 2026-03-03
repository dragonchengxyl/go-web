package http

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/transport/http/handler"
	"github.com/studio/platform/internal/transport/http/middleware"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
	redisClient "github.com/redis/go-redis/v9"
)

// RouterConfig holds router dependencies
type RouterConfig struct {
	Config      *configs.Config
	Logger      *zap.Logger
	RedisClient *redisClient.Client
	UserService *usecase.UserService
	TokenStore  *redis.TokenStore
}

// NewRouter creates a new HTTP router
func NewRouter(cfg RouterConfig) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Config.Server.Mode)

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.CORS())

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RedisClient, cfg.Config.RateLimit)
	r.Use(rateLimiter.Limit())

	// Health check endpoints (no auth required)
	healthHandler := handler.NewHealthHandler()
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		authHandler := handler.NewAuthHandler(cfg.UserService)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		authMiddleware := middleware.NewAuth(cfg.Config.JWT, cfg.TokenStore)
		protected := v1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			protected.POST("/auth/logout", authHandler.Logout)
			// TODO: Add more protected routes
		}
	}

	return r
}
