package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/oss"
	paymentalipay "github.com/studio/platform/internal/infra/payment/alipay"
	"github.com/studio/platform/internal/infra/payment/mock"
	paymentwechat "github.com/studio/platform/internal/infra/payment/wechat"
	paymentpkg "github.com/studio/platform/internal/infra/payment"
	"github.com/studio/platform/internal/infra/postgres"
	"github.com/studio/platform/internal/infra/redis"
	transporthttp "github.com/studio/platform/internal/transport/http"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := configs.Load()
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
	gameRepo := postgres.NewGameRepository(pool)
	branchRepo := postgres.NewBranchRepository(pool)
	releaseRepo := postgres.NewReleaseRepository(pool)
	assetRepo := postgres.NewAssetRepository(pool)
	albumRepo := postgres.NewAlbumRepository(pool)
	trackRepo := postgres.NewTrackRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)
	productRepo := postgres.NewProductRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	couponRepo := postgres.NewCouponRepository(pool)
	achievementRepo := postgres.NewAchievementRepository(pool)

	// Initialize token store
	tokenStore := redis.NewTokenStore(redisClient)
	leaderboard := redis.NewLeaderboard(redisClient)

	// Initialize storage service
	storageService := oss.NewAliyunOSS(cfg.OSS)

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
	gameService := usecase.NewGameService(gameRepo)
	branchService := usecase.NewBranchService(branchRepo, gameRepo)
	releaseService := usecase.NewReleaseService(releaseRepo, branchRepo)
	assetService := usecase.NewAssetService(assetRepo, gameRepo, branchRepo, releaseRepo, storageService)
	musicService := usecase.NewMusicService(albumRepo, trackRepo, storageService)
	commentService := usecase.NewCommentService(commentRepo)
	productService := usecase.NewProductService(productRepo)
	orderService := usecase.NewOrderService(orderRepo, productRepo, couponRepo)
	couponService := usecase.NewCouponService(couponRepo, productRepo)
	paymentService := usecase.NewPaymentService(orderRepo, alipayGW, wechatGW)
	searchService := usecase.NewSearchService(pool)
	statsService := usecase.NewStatsService(pool)
	achievementService := usecase.NewAchievementService(achievementRepo, leaderboard)

	// Initialize HTTP router
	router := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:         cfg,
		Logger:         logger,
		Pool:           pool,
		RedisClient:    redisClient,
		UserService:    userService,
		GameService:    gameService,
		BranchService:  branchService,
		ReleaseService: releaseService,
		AssetService:   assetService,
		MusicService:   musicService,
		CommentService: commentService,
		ProductService: productService,
		OrderService:   orderService,
		CouponService:       couponService,
		PaymentService:      paymentService,
		SearchService:       searchService,
		StatsService:        statsService,
		AchievementService:  achievementService,
		TokenStore:          tokenStore,
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
