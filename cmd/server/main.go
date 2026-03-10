package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/moderation"
	"github.com/studio/platform/internal/infra/oss"
	paymentalipay "github.com/studio/platform/internal/infra/payment/alipay"
	"github.com/studio/platform/internal/infra/payment/mock"
	paymentwechat "github.com/studio/platform/internal/infra/payment/wechat"
	paymentpkg "github.com/studio/platform/internal/infra/payment"
	"github.com/studio/platform/internal/infra/postgres"
	"github.com/studio/platform/internal/infra/redis"
	transporthttp "github.com/studio/platform/internal/transport/http"
	"github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "configs/config.local.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := configs.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	ctx := context.Background()

	// Initialize PostgreSQL
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("Connected to PostgreSQL")

	// Run database migrations
	migrator := postgres.NewMigrator(logger)
	if err := migrator.RunMigrations(cfg.Database.DSN, "migrations"); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()
	logger.Info("Connected to Redis")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(pool)
	albumRepo := postgres.NewAlbumRepository(pool)
	trackRepo := postgres.NewTrackRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	achievementRepo := postgres.NewAchievementRepository(pool)
	postRepo := postgres.NewPostRepository(pool)
	followRepo := postgres.NewFollowRepository(pool)
	chatRepo := postgres.NewChatRepository(pool)
	notificationRepo := postgres.NewNotificationRepository(pool)

	// Initialize token store
	tokenStore := redis.NewTokenStore(redisClient)
	leaderboard := redis.NewLeaderboard(redisClient)

	// Initialize storage service (R2 preferred, fallback to Aliyun)
	var storageService oss.StorageService
	if cfg.OSS.Provider == "r2" {
		storageService = oss.NewCloudflareR2(cfg.OSS)
		logger.Info("Using Cloudflare R2 storage")
	} else {
		storageService = oss.NewAliyunOSS(cfg.OSS)
		logger.Info("Using Aliyun OSS storage")
	}

	// Initialize payment gateways (fall back to mock if credentials not configured)
	mockGW := mock.New()
	var alipayGW, wechatGW paymentpkg.PaymentGateway = mockGW, mockGW
	if cfg.Payment.Alipay.AppID != "" {
		if gw, err := paymentalipay.New(cfg.Payment.Alipay); err == nil {
			alipayGW = gw
		} else {
			logger.Warn("alipay gateway init failed, using mock", zap.Error(err))
		}
	}
	if cfg.Payment.Wechat.AppID != "" {
		wechatGW = paymentwechat.New(cfg.Payment.Wechat)
	}

	// Initialize services
	userService := usecase.NewUserService(userRepo, tokenStore, cfg.JWT)
	musicService := usecase.NewMusicService(albumRepo, trackRepo, storageService)
	commentService := usecase.NewCommentService(commentRepo)
	orderService := usecase.NewOrderService(orderRepo, nil, nil)
	paymentService := usecase.NewPaymentService(orderRepo, alipayGW, wechatGW)
	searchService := usecase.NewSearchService(pool)
	statsService := usecase.NewStatsService(pool)
	achievementService := usecase.NewAchievementService(achievementRepo, leaderboard)
	postService := usecase.NewPostService(postRepo)
	// Enable content moderation if Aliyun Green is configured
	if cfg.Moderation.AccessKeyID != "" {
		moderator := moderation.NewAliyunGreen(
			cfg.Moderation.AccessKeyID,
			cfg.Moderation.AccessKeySecret,
			cfg.Moderation.Endpoint,
		)
		postService = usecase.NewPostService(postRepo,
			usecase.WithModerator(moderator, logger),
			usecase.WithAllowedHosts(cfg.OSS.AllowedHosts),
		)
		logger.Info("Aliyun Green content moderation enabled")
	} else if len(cfg.OSS.AllowedHosts) > 0 {
		postService = usecase.NewPostService(postRepo, usecase.WithAllowedHosts(cfg.OSS.AllowedHosts))
	}
	followService := usecase.NewFollowService(followRepo)
	chatService := usecase.NewChatService(chatRepo)
	tipService := usecase.NewTipService(orderRepo)
	ossService := usecase.NewOSSService(storageService)

	// Initialize WebSocket hub (distributed mode via Redis Pub/Sub)
	hub := ws.NewDistributedHub(redisClient, logger)
	hubCtx, hubCancel := context.WithCancel(context.Background())
	defer hubCancel()
	go hub.Run(hubCtx)

	notificationService := usecase.NewNotificationService(notificationRepo, hub)

	reportRepo := postgres.NewReportRepository(pool)
	blockRepo := postgres.NewBlockRepository(pool)

	// Initialize HTTP router
	router := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:             cfg,
		Logger:             logger,
		Pool:               pool,
		RedisClient:        redisClient,
		UserService:        userService,
		MusicService:       musicService,
		CommentService:     commentService,
		OrderService:       orderService,
		PaymentService:     paymentService,
		SearchService:      searchService,
		StatsService:       statsService,
		AchievementService: achievementService,
		PostService:        postService,
		FollowService:      followService,
		ChatService:        chatService,
		TipService:          tipService,
		NotificationService: notificationService,
		OSSService:          ossService,
		Hub:                 hub,
		TokenStore:         tokenStore,
		ReportRepo:         reportRepo,
		BlockRepo:          blockRepo,
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting server", zap.String("addr", addr))

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Background scheduled tasks
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	// Cancel expired orders every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				n, err := orderService.CancelExpiredOrders(bgCtx)
				if err != nil {
					logger.Error("cancel expired orders failed", zap.Error(err))
				} else if n > 0 {
					logger.Info("cancelled expired orders", zap.Int("count", n))
				}
			case <-bgCtx.Done():
				return
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

func initLogger(mode string) (*zap.Logger, error) {
	if mode == "release" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
