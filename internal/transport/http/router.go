package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/block"
	"github.com/studio/platform/internal/domain/report"
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
	Config                *configs.Config
	Logger                *zap.Logger
	Pool                  *pgxpool.Pool
	RedisClient           *redisClient.Client
	UserService           *usecase.UserService
	MusicService          *usecase.MusicService
	CommentService        *usecase.CommentService
	PaymentService        *usecase.PaymentService
	SearchService         *usecase.SearchService
	StatsService          usecase.StatsProvider
	AchievementService    *usecase.AchievementService
	PostService           *usecase.PostService
	FollowService         *usecase.FollowService
	ChatService           *usecase.ChatService
	TipService            *usecase.TipService
	NotificationService   *usecase.NotificationService
	OSSService            *usecase.OSSService
	EventService          *usecase.EventService
	GroupService          *usecase.GroupService
	RecommendationService *usecase.RecommendationService
	Hub                   ws.HubInterface
	TokenStore            *redis.TokenStore
	ReportRepo            report.Repository
	BlockRepo             block.Repository
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
		authHandler := handler.NewAuthHandler(cfg.UserService, cfg.RedisClient)
		userHandler := handler.NewUserHandler(cfg.UserService)
		musicHandler := handler.NewMusicHandler(cfg.MusicService)
		commentHandler := handler.NewCommentHandler(cfg.CommentService, cfg.PostService)
		searchHandler := handler.NewSearchHandler(cfg.MusicService, cfg.SearchService, cfg.PostService, cfg.UserService)
		postHandler := handler.NewPostHandler(cfg.PostService, cfg.FollowService, cfg.UserService)
		followHandler := handler.NewFollowHandler(cfg.FollowService)
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
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
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
		v1.GET("/explore/tags", postHandler.GetHotTags)

		// Sponsor dashboard (public)
		sponsorHandler := handler.NewSponsorHandler(cfg.Config.Sponsor)
		v1.GET("/sponsor", sponsorHandler.GetSponsorInfo)

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

		// Events (public read)
		if cfg.EventService != nil {
			eventHandler := handler.NewEventHandler(cfg.EventService, cfg.UserService)
			events := v1.Group("/events")
			events.GET("", eventHandler.ListEvents)
			events.GET("/:id", eventHandler.GetEvent)
			events.GET("/:id/attendees", eventHandler.ListAttendees)
		}

		// Groups (public read)
		if cfg.GroupService != nil {
			groupHandler := handler.NewGroupHandler(cfg.GroupService, cfg.UserService)
			groups := v1.Group("/groups")
			groups.GET("", groupHandler.ListGroups)
			groups.GET("/:id", groupHandler.GetGroup)
			groups.GET("/:id/members", groupHandler.ListMembers)
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
			protected.PUT("/auth/password", authHandler.ChangePassword)
			protected.POST("/auth/resend-verification", authHandler.ResendVerification)

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

			// Personalised recommendations
			if cfg.RecommendationService != nil {
				recHandler := handler.NewRecommendationHandler(cfg.RecommendationService)
				protected.GET("/posts/recommended", recHandler.GetRecommended)
			}

			// Events (authenticated write)
			if cfg.EventService != nil {
				eventHandler := handler.NewEventHandler(cfg.EventService, cfg.UserService)
				protected.POST("/events", eventHandler.CreateEvent)
				protected.PUT("/events/:id", eventHandler.UpdateEvent)
				protected.DELETE("/events/:id", eventHandler.CancelEvent)
				protected.POST("/events/:id/attend", eventHandler.AttendEvent)
				protected.GET("/users/me/events", eventHandler.MyEvents)
				protected.GET("/users/me/attending", eventHandler.MyAttending)
			}

			// Groups (authenticated write)
			if cfg.GroupService != nil {
				groupHandler := handler.NewGroupHandler(cfg.GroupService, cfg.UserService)
				protected.POST("/groups", groupHandler.CreateGroup)
				protected.PUT("/groups/:id", groupHandler.UpdateGroup)
				protected.POST("/groups/:id/join", groupHandler.JoinGroup)
				protected.DELETE("/groups/:id/leave", groupHandler.LeaveGroup)
				protected.PUT("/groups/:id/members/:uid", groupHandler.UpdateMemberRole)
				protected.DELETE("/groups/:id/members/:uid", groupHandler.KickMember)
				protected.GET("/users/me/groups", groupHandler.MyGroups)
			}

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

			// OSS direct upload policy
			if cfg.OSSService != nil {
				ossHandler := handler.NewOSSHandler(cfg.OSSService)
				protected.POST("/upload/oss-policy", ossHandler.GetUploadToken)
			}

			// Creator dashboard
			creatorHandler := handler.NewCreatorHandler(cfg.PostService, cfg.FollowService, cfg.TipService)
			protected.GET("/creator/stats", creatorHandler.GetStats)

			// Reports
			reportHandler := handler.NewReportHandler(cfg.ReportRepo)
			protected.POST("/reports", reportHandler.CreateReport)

			// Block
			blockHandler := handler.NewBlockHandler(cfg.BlockRepo, cfg.UserService)
			protected.POST("/users/:id/block", blockHandler.Block)
			protected.DELETE("/users/:id/block", blockHandler.Unblock)
			protected.GET("/users/me/blocked", blockHandler.ListBlocked)

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
				adminHandler := handler.NewAdminHandler(cfg.StatsService, cfg.UserService, nil, cfg.CommentService, cfg.PostService, cfg.ReportRepo)

				admin.GET("/stats/dashboard", adminHandler.GetDashboardStats)
				admin.GET("/stats/user-growth", adminHandler.GetUserGrowthChart)

				admin.GET("/users", adminHandler.ListUsers)
				admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
				admin.POST("/users/:id/ban", adminHandler.BanUser)
				admin.POST("/users/:id/unban", adminHandler.UnbanUser)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)

				admin.GET("/comments", adminHandler.ListComments)
				admin.DELETE("/comments/:id", adminHandler.DeleteComment)

				// Admin post moderation
				admin.GET("/posts", adminHandler.ListPosts)
				admin.PUT("/posts/:id/moderation", adminHandler.UpdatePostModeration)
				admin.DELETE("/posts/:id", postHandler.DeletePost)

				// Admin reports
				admin.GET("/reports", adminHandler.ListReports)
				admin.PUT("/reports/:id", adminHandler.UpdateReport)

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
