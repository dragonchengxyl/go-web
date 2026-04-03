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
	"github.com/studio/platform/internal/infra/kafkaevent"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/observability/httpserver"
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

	postRepo := postgresinfra.NewPostRepository(pool)
	groupRepo := postgresinfra.NewGroupRepository(pool)
	agentRepo := postgresinfra.NewAgentRepository(pool)

	postService := usecase.NewPostService(
		postRepo,
		usecase.WithAllowedHosts(cfg.OSS.AllowedHosts),
		usecase.WithGroupRepository(groupRepo),
	)
	agentService := usecase.NewAgentService(agentRepo, postService)
	worker := usecase.NewAgentWorker(agentService, logger, 5*time.Second, 20)

	httpPort := cfg.Observability.AgentWorkerHTTPPort
	if httpPort == 0 {
		httpPort = 18056
	}
	obsServer := httpserver.New("agent-worker", httpPort, logger, map[string]httpserver.CheckFunc{
		"database": pool.Ping,
		"redis": func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		},
		"kafka": func(ctx context.Context) error {
			return kafkaevent.CheckConnectivity(ctx, cfg.Kafka)
		},
	})
	obsServer.Start()
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = obsServer.Shutdown(shutCtx)
	}()

	worker.Start(ctx)
	logger.Info("Agent worker started",
		zap.Duration("poll_interval", 5*time.Second),
		zap.Int("retry_batch_size", 20),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	logger.Info("Agent worker stopped")
}
