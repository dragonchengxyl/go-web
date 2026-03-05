package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Pool           *pgxpool.Pool
	RedisClient    *redisClient.Client
	UserService    *usecase.UserService
	GameService    *usecase.GameService
	BranchService  *usecase.BranchService
	ReleaseService *usecase.ReleaseService
	AssetService   *usecase.AssetService
	MusicService   *usecase.MusicService
	CommentService *usecase.CommentService
	ProductService *usecase.ProductService
	OrderService   *usecase.OrderService
	CouponService  *usecase.CouponService
	StatsService   *usecase.StatsService
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
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.Config.Server.AllowOrigins))

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RedisClient, cfg.Config.RateLimit)
	r.Use(rateLimiter.Limit())

	// Health check endpoints (no auth required)
	healthHandler := handler.NewHealthHandler(cfg.Pool, cfg.RedisClient)
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
		productHandler := handler.NewProductHandler(cfg.ProductService)
		orderHandler := handler.NewOrderHandler(cfg.OrderService)
		couponHandler := handler.NewCouponHandler(cfg.CouponService)
		searchHandler := handler.NewSearchHandler(cfg.GameService, cfg.MusicService)
		authMiddleware := middleware.NewAuth(cfg.Config.JWT, cfg.TokenStore)

		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Search routes (public)
		search := v1.Group("/search")
		{
			search.GET("", searchHandler.SearchAll)
			search.GET("/games", searchHandler.SearchGames)
			search.GET("/albums", searchHandler.SearchAlbums)
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
				adminHandler := handler.NewAdminHandler(cfg.StatsService, cfg.UserService, cfg.GameService, cfg.CommentService)

			// Dashboard stats
			admin.GET("/stats/dashboard", adminHandler.GetDashboardStats)
			admin.GET("/stats/revenue", adminHandler.GetRevenueChart)
			admin.GET("/stats/user-growth", adminHandler.GetUserGrowthChart)
			admin.GET("/stats/popular-games", adminHandler.GetPopularGames)

			// User management
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
			admin.POST("/users/:id/ban", adminHandler.BanUser)
			admin.POST("/users/:id/unban", adminHandler.UnbanUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)

			// Comment moderation
			admin.GET("/comments", adminHandler.ListComments)
			admin.DELETE("/comments/:id", adminHandler.DeleteComment)
			}
		}

// Game routes (public)
games := v1.Group("/games")
{
	games.GET("", gameHandler.ListGames)
	games.GET("/slug/:slug", gameHandler.GetGameBySlug)
	games.GET("/:id", gameHandler.GetGameByID)

	// Protected game routes (Admin only)
	gamesProtected := games.Group("")
	gamesProtected.Use(authMiddleware.Authenticate())
	gamesProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
	{
		gamesProtected.POST("", gameHandler.CreateGame)
		gamesProtected.PUT("/:id", gameHandler.UpdateGame)
		gamesProtected.DELETE("/:id", gameHandler.DeleteGame)
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

// Game-Branch relationship (use /games/:id/branches)
v1.GET("/games/:id/branches", branchHandler.ListBranches)
v1.POST("/games/:id/branches", authMiddleware.Authenticate(), authMiddleware.RequireRole(user.RoleAdmin), branchHandler.CreateBranch)

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

		// Product routes (public read)
		products := v1.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)

			// Protected product routes (Admin only)
			productsProtected := products.Group("")
			productsProtected.Use(authMiddleware.Authenticate())
			productsProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				productsProtected.POST("", productHandler.CreateProduct)
				productsProtected.PUT("/:id", productHandler.UpdateProduct)
				productsProtected.DELETE("/:id", productHandler.DeleteProduct)
				productsProtected.POST("/discounts", productHandler.CreateDiscount)
			}
		}

		// Order routes (authenticated users)
		orders := v1.Group("/orders")
		orders.Use(authMiddleware.Authenticate())
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListMyOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("/:id/pay", orderHandler.PayOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// Coupon routes
		coupons := v1.Group("/coupons")
		{
			// Public coupon validation (authenticated)
			couponsAuth := coupons.Group("")
			couponsAuth.Use(authMiddleware.Authenticate())
			{
				couponsAuth.GET("/validate", couponHandler.ValidateCoupon)
				couponsAuth.POST("/redeem", couponHandler.RedeemCode)
			}

			// Admin coupon management
			couponsAdmin := coupons.Group("")
			couponsAdmin.Use(authMiddleware.Authenticate())
			couponsAdmin.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				couponsAdmin.POST("", couponHandler.CreateCoupon)
				couponsAdmin.POST("/redeem-codes", couponHandler.CreateRedeemCode)
				couponsAdmin.POST("/redeem-codes/batch", couponHandler.BatchCreateRedeemCodes)
			}
		}
	}

	return r
}
