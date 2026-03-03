package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studio/platform/configs"
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
	defer logger.Sync()

	ctx := context.Background()

	// Initialize PostgreSQL
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("Connected to PostgreSQL")

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

	// Initialize token store
	tokenStore := redis.NewTokenStore(redisClient)

	// Initialize services
	userService := usecase.NewUserService(userRepo, tokenStore, cfg.JWT)
	gameService := usecase.NewGameService(gameRepo)
	branchService := usecase.NewBranchService(branchRepo, gameRepo)
	releaseService := usecase.NewReleaseService(releaseRepo, branchRepo)
	assetService := usecase.NewAssetService(assetRepo, gameRepo, branchRepo, releaseRepo)
	musicService := usecase.NewMusicService(albumRepo, trackRepo)
	commentService := usecase.NewCommentService(commentRepo)
	productService := usecase.NewProductService(productRepo)
	orderService := usecase.NewOrderService(orderRepo, productRepo, couponRepo)
	couponService := usecase.NewCouponService(couponRepo, productRepo)

	// Initialize HTTP router
	router := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:         cfg,
		Logger:         logger,
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
		CouponService:  couponService,
		TokenStore:     tokenStore,
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting server", zap.String("addr", addr))

	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx // TODO: implement graceful shutdown

	logger.Info("Server stopped")
}

func initLogger(mode string) (*zap.Logger, error) {
	if mode == "release" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
