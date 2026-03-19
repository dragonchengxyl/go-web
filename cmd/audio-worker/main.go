package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/studio/platform/configs"
	audioinfra "github.com/studio/platform/internal/infra/audio"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "configs/config.local.yaml", "path to config file")
	flag.Parse()

	cfg, err := configs.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := redisinfra.NewClient(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	pool, err := postgresinfra.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	audioJobRepo := postgresinfra.NewAudioJobRepository(pool)
	audioProcessor := audioinfra.NewLocalProcessor("./uploads", "/uploads")

	retryBackoff := time.Duration(max(cfg.Audio.RetryBackoffSec, 10)) * time.Second
	pollInterval := time.Duration(max(cfg.Audio.RetryPollSec, 5)) * time.Second
	retryBatchSize := max(cfg.Audio.RetryBatchSize, 20)
	maxAttempts := max(cfg.Audio.MaxAttempts, 3)

	audioJobService := usecase.NewAudioJobService(
		audioJobRepo,
		usecase.WithAudioJobLogger(logger),
		usecase.WithAudioJobAllowedHosts(cfg.OSS.AllowedHosts),
		usecase.WithAudioJobProcessor(audioProcessor),
		usecase.WithAudioJobRetryPolicy(maxAttempts, retryBackoff),
	)

	consumer := streams.NewConsumer(redisClient, logger, "audio-worker-1")
	worker := usecase.NewAudioWorker(consumer, audioJobService, logger, pollInterval, retryBatchSize)
	worker.Start(ctx)
	logger.Info("Audio worker started",
		zap.Int("max_attempts", maxAttempts),
		zap.Duration("retry_backoff", retryBackoff),
		zap.Duration("retry_poll_interval", pollInterval),
		zap.Int("retry_batch_size", retryBatchSize),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	logger.Info("Audio worker stopped")
}

func max(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
