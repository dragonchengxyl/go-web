package http

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/transport/http/handler"
	"github.com/studio/platform/internal/transport/http/middleware"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
	redisClient "github.com/redis/go-redis/v9"
)

// RouterConfig holds router dependencies
type RouterConfig struct {
	Config         *configs.Config
	Logger         *zap.Logger
	RedisClient    *redisClient.Client
	UserService    *usecase.UserService
	GameService    *usecase.GameService
	BranchService  *usecase.BranchService
	ReleaseService *usecase.ReleaseService
	AssetService   *usecase.AssetService
	MusicService   *usecase.MusicService
	CommentService *usecase.CommentService
	TokenStore     *redis.TokenStore
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
		// Initialize handlers
		authHandler := handler.NewAuthHandler(cfg.UserService)
		userHandler := handler.NewUserHandler(cfg.UserService)
		gameHandler := handler.NewGameHandler(cfg.GameService)
		branchHandler := handler.NewBranchHandler(cfg.BranchService)
		releaseHandler := handler.NewReleaseHandler(cfg.ReleaseService)
		assetHandler := handler.NewAssetHandler(cfg.AssetService)
		musicHandler := handler.NewMusicHandler(cfg.MusicService)
		commentHandler := handler.NewCommentHandler(cfg.CommentService)
		authMiddleware := middleware.NewAuth(cfg.Config.JWT, cfg.TokenStore)

		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			protected.POST("/auth/logout", authHandler.Logout)

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetProfile)
				users.PUT("/me", userHandler.UpdateProfile)
				users.GET("/:id", userHandler.GetUserByID)

				// User assets
				users.GET("/me/assets", assetHandler.GetMyAssets)
				users.GET("/me/games/:game_id/assets", assetHandler.GetMyGameAssets)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				// User management
				admin.GET("/users", userHandler.ListUsers)
				admin.PUT("/users/:id/role", userHandler.UpdateUserRole)
				admin.PUT("/users/:id/status", userHandler.UpdateUserStatus)
				admin.DELETE("/users/:id", userHandler.DeleteUser)

				// Asset management
				admin.POST("/users/:user_id/assets", assetHandler.GrantAsset)
				admin.GET("/users/:user_id/assets", assetHandler.GetUserAssets)
				admin.DELETE("/users/:user_id/assets", assetHandler.RevokeAsset)
			}
		}

		// Game routes (public)
		games := v1.Group("/games")
		{
			games.GET("", gameHandler.ListGames)
			games.GET("/:id", gameHandler.GetGameByID)
			games.GET("/slug/:slug", gameHandler.GetGameBySlug)

			// Branch routes (public read)
			games.GET("/:game_id/branches", branchHandler.ListBranches)

			// Protected game routes (Admin only)
			gamesProtected := games.Group("")
			gamesProtected.Use(authMiddleware.Authenticate())
			gamesProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				gamesProtected.POST("", gameHandler.CreateGame)
				gamesProtected.PUT("/:id", gameHandler.UpdateGame)
				gamesProtected.DELETE("/:id", gameHandler.DeleteGame)

				// Branch management (Admin only)
				gamesProtected.POST("/:game_id/branches", branchHandler.CreateBranch)
			}
		}

		// Branch routes
		branches := v1.Group("/branches")
		{
			branches.GET("/:branch_id", branchHandler.GetBranch)
			branches.GET("/:branch_id/releases", releaseHandler.ListReleases)
			branches.GET("/:branch_id/releases/latest", releaseHandler.GetLatestRelease)

			// Protected branch routes (Admin only)
			branchesProtected := branches.Group("")
			branchesProtected.Use(authMiddleware.Authenticate())
			branchesProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				branchesProtected.PUT("/:branch_id", branchHandler.UpdateBranch)
				branchesProtected.DELETE("/:branch_id", branchHandler.DeleteBranch)

				// Release management (Admin only)
				branchesProtected.POST("/:branch_id/releases", releaseHandler.CreateRelease)
			}
		}

		// Release routes
		releases := v1.Group("/releases")
		{
			releases.GET("/:release_id", releaseHandler.GetRelease)

			// Protected release routes (Admin only)
			releasesProtected := releases.Group("")
			releasesProtected.Use(authMiddleware.Authenticate())
			releasesProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				releasesProtected.PUT("/:release_id", releaseHandler.UpdateRelease)
				releasesProtected.POST("/:release_id/publish", releaseHandler.PublishRelease)
				releasesProtected.POST("/:release_id/unpublish", releaseHandler.UnpublishRelease)
				releasesProtected.DELETE("/:release_id", releaseHandler.DeleteRelease)
			}

			// Download route (authenticated users)
			releasesAuth := releases.Group("")
			releasesAuth.Use(authMiddleware.Authenticate())
			{
				releasesAuth.POST("/:release_id/download", assetHandler.RequestDownload)
			}
		}

		// Music routes (public)
		albums := v1.Group("/albums")
		{
			albums.GET("", musicHandler.ListAlbums)
			albums.GET("/:album_id", musicHandler.GetAlbum)
			albums.GET("/slug/:slug", musicHandler.GetAlbumBySlug)
			albums.GET("/:album_id/tracks", musicHandler.GetAlbumTracks)

			// Protected album routes (Admin only)
			albumsProtected := albums.Group("")
			albumsProtected.Use(authMiddleware.Authenticate())
			albumsProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				albumsProtected.POST("", musicHandler.CreateAlbum)
			}
		}

		// Track routes
		tracks := v1.Group("/tracks")
		{
			tracks.GET("/:track_id", musicHandler.GetTrack)
			tracks.POST("/:track_id/stream", musicHandler.StreamTrack)
		}

		// Comment routes
		comments := v1.Group("/comments")
		{
			comments.GET("", commentHandler.ListComments)

			// Protected comment routes (authenticated users)
			commentsProtected := comments.Group("")
			commentsProtected.Use(authMiddleware.Authenticate())
			{
				commentsProtected.POST("", commentHandler.CreateComment)
				commentsProtected.PUT("/:comment_id", commentHandler.UpdateComment)
				commentsProtected.DELETE("/:comment_id", commentHandler.DeleteComment)
				commentsProtected.POST("/:comment_id/like", commentHandler.LikeComment)
				commentsProtected.DELETE("/:comment_id/like", commentHandler.UnlikeComment)
			}
		}
	}

	return r
}
