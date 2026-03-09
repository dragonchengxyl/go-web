package http

import (
	redisClient "github.com/redis/go-redis/v9"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/transport/http/handler"
	"github.com/studio/platform/internal/transport/http/middleware"
	"github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

// RouterConfig holds router dependencies
type RouterConfig struct {
	Config              *configs.Config
	Logger              *zap.Logger
	Pool                *pgxpool.Pool
	RedisClient         *redisClient.Client
	UserService         *usecase.UserService
	MusicService        *usecase.MusicService
	CommentService      *usecase.CommentService
	OrderService        *usecase.OrderService
	PaymentService      *usecase.PaymentService
	SearchService       *usecase.SearchService
	StatsService        *usecase.StatsService
	AchievementService  *usecase.AchievementService
	PostService         *usecase.PostService
	FollowService       *usecase.FollowService
	ChatService         *usecase.ChatService
	TipService          *usecase.TipService
	NotificationService *usecase.NotificationService
	Hub                 *ws.Hub
	TokenStore          *redis.TokenStore
}

// NewRouter creates a new HTTP router
func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(cfg.Config.Server.Mode)

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.Config.Server.AllowOrigins))
	r.Use(middleware.PrometheusMetrics())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	rateLimiter := middleware.NewRateLimiter(cfg.RedisClient, cfg.Config.RateLimit)
	r.Use(rateLimiter.Limit())

	healthHandler := handler.NewHealthHandler(cfg.Pool, cfg.RedisClient)
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// Static file serving for local uploads
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		// Initialize handlers
		authHandler := handler.NewAuthHandler(cfg.UserService)
		userHandler := handler.NewUserHandler(cfg.UserService)
		musicHandler := handler.NewMusicHandler(cfg.MusicService)
		commentHandler := handler.NewCommentHandler(cfg.CommentService, cfg.PostService, cfg.NotificationService)
		searchHandler := handler.NewSearchHandler(nil, cfg.MusicService, cfg.SearchService)
		postHandler := handler.NewPostHandler(cfg.PostService, cfg.FollowService, cfg.NotificationService)
		followHandler := handler.NewFollowHandler(cfg.FollowService, cfg.NotificationService)
		chatHandler := handler.NewChatHandler(cfg.ChatService, cfg.Hub, cfg.Logger)
		tipHandler := handler.NewTipHandler(cfg.TipService, cfg.PaymentService)
		achievementHandler := handler.NewAchievementHandler(cfg.AchievementService)
		notificationHandler := handler.NewNotificationHandler(cfg.NotificationService)

		authMiddleware := middleware.NewAuth(cfg.Config.JWT, cfg.TokenStore)

		// ── Public routes ──────────────────────────────────────────────

		// Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Search
		search := v1.Group("/search")
		{
			search.GET("", searchHandler.SearchAll)
			search.GET("/albums", searchHandler.SearchAlbums)
			search.GET("/popular", searchHandler.GetPopularSearches)
		}

		// Explore (public)
		v1.GET("/explore", postHandler.GetExplore)

		// Music (public read)
		albums := v1.Group("/albums")
		{
			albums.GET("", musicHandler.ListAlbums)
			albums.GET("/:album_id", musicHandler.GetAlbum)
			albums.GET("/slug/:slug", musicHandler.GetAlbumBySlug)
			albums.GET("/:album_id/tracks", musicHandler.GetAlbumTracks)
		}

		tracks := v1.Group("/tracks")
		{
			tracks.GET("/:track_id", musicHandler.GetTrack)
			tracks.POST("/:track_id/stream", musicHandler.StreamTrack)
		}

		// Comments (public list)
		comments := v1.Group("/comments")
		{
			comments.GET("", commentHandler.ListComments)
		}

		// Achievements & leaderboard (public)
		v1.GET("/achievements", achievementHandler.ListAchievements)
		v1.GET("/leaderboard", achievementHandler.GetLeaderboard)
		v1.GET("/leaderboard/weekly", achievementHandler.GetWeeklyLeaderboard)

		// Posts (public read)
		posts := v1.Group("/posts")
		{
			posts.GET("/:id", postHandler.GetPost)
		}

		// Users public profile
		v1.GET("/users/:id/achievements", achievementHandler.GetUserAchievements)
		v1.GET("/users/:id/followers", followHandler.ListFollowers)
		v1.GET("/users/:id/following", followHandler.ListFollowing)
		v1.GET("/users/:id/follow-stats", followHandler.GetFollowStats)
		v1.GET("/users/:id/posts", postHandler.ListUserPosts)
		v1.GET("/users/:id/tips/received", tipHandler.ListReceivedTips)
		v1.GET("/users/:id", userHandler.GetUserByID)

		// ── Authenticated routes ────────────────────────────────────────
		protected := v1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			protected.POST("/auth/logout", authHandler.Logout)

			// Profile
			protected.GET("/users/me", userHandler.GetProfile)
			protected.PUT("/users/me", userHandler.UpdateProfile)

			// My achievements & points
			protected.GET("/users/me/achievements", achievementHandler.GetMyAchievements)
			protected.GET("/users/me/points", achievementHandler.GetMyPoints)
			protected.GET("/users/me/points/history", achievementHandler.GetMyPointHistory)

			// Posts (authenticated write)
			protected.POST("/posts", postHandler.CreatePost)
			protected.PUT("/posts/:id", postHandler.UpdatePost)
			protected.DELETE("/posts/:id", postHandler.DeletePost)
			protected.POST("/posts/:id/like", postHandler.LikePost)
			protected.DELETE("/posts/:id/like", postHandler.UnlikePost)

			// Feed
			protected.GET("/feed", postHandler.GetFeed)

			// Follow
			protected.POST("/users/:id/follow", followHandler.Follow)
			protected.DELETE("/users/:id/follow", followHandler.Unfollow)

			// Comments (authenticated write)
			commentsAuth := protected.Group("/comments")
			{
				commentsAuth.POST("", commentHandler.CreateComment)
				commentsAuth.PUT("/:comment_id", commentHandler.UpdateComment)
				commentsAuth.DELETE("/:comment_id", commentHandler.DeleteComment)
				commentsAuth.POST("/:comment_id/like", commentHandler.LikeComment)
				commentsAuth.DELETE("/:comment_id/like", commentHandler.UnlikeComment)
			}

			// Conversations / Chat
			protected.GET("/conversations", chatHandler.ListConversations)
			protected.POST("/conversations", chatHandler.CreateConversation)
			protected.GET("/conversations/:id/messages", chatHandler.ListMessages)
			protected.POST("/conversations/:id/messages", chatHandler.SendMessage)
			protected.PUT("/conversations/:id/read", chatHandler.MarkRead)

			// Tips
			protected.POST("/tips", tipHandler.CreateTip)

			// Notifications
			protected.GET("/notifications", notificationHandler.ListNotifications)
			protected.POST("/notifications/read", notificationHandler.MarkRead)
			protected.GET("/notifications/unread-count", notificationHandler.CountUnread)

			// Upload
			uploadHandler := handler.NewUploadHandler("./uploads", 10*1024*1024)
			protected.POST("/upload/image", uploadHandler.UploadImage)

			// Music (admin write)
			albumsProtected := protected.Group("/albums")
			albumsProtected.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				albumsProtected.POST("", musicHandler.CreateAlbum)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.RequireRole(user.RoleAdmin))
			{
				adminHandler := handler.NewAdminHandler(cfg.StatsService, cfg.UserService, nil, cfg.CommentService)

				admin.GET("/stats/dashboard", adminHandler.GetDashboardStats)
				admin.GET("/stats/revenue", adminHandler.GetRevenueChart)
				admin.GET("/stats/user-growth", adminHandler.GetUserGrowthChart)
				admin.GET("/stats/popular-games", adminHandler.GetPopularGames)

				admin.GET("/users", adminHandler.ListUsers)
				admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
				admin.POST("/users/:id/ban", adminHandler.BanUser)
				admin.POST("/users/:id/unban", adminHandler.UnbanUser)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)

				admin.GET("/comments", adminHandler.ListComments)
				admin.DELETE("/comments/:id", adminHandler.DeleteComment)

				// Admin post moderation
				admin.DELETE("/posts/:id", postHandler.DeletePost)

				// Admin achievements
				admin.POST("/users/:id/achievements", achievementHandler.AdminUnlockAchievement)
				admin.POST("/users/:id/points", achievementHandler.AdminAwardPoints)
			}

			// Payment callbacks (no auth — called by gateways)
			if cfg.PaymentService != nil {
				paymentHandler := handler.NewPaymentHandler(cfg.PaymentService)
				v1.POST("/payment/callback/alipay", paymentHandler.AlipayCallback)
				v1.POST("/payment/callback/wechat", paymentHandler.WechatCallback)
				protected.POST("/orders/:id/pay/alipay", paymentHandler.PayWithAlipay)
				protected.POST("/orders/:id/pay/wechat", paymentHandler.PayWithWechat)
			}
		}
	}

	// WebSocket endpoint (auth via query param token)
	chatHandler := handler.NewChatHandler(cfg.ChatService, cfg.Hub, cfg.Logger)
	r.GET("/ws/chat", middleware.NewAuth(cfg.Config.JWT, cfg.TokenStore).Authenticate(), chatHandler.ServeWS)

	return r
}
