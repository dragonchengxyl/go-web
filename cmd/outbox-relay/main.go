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
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/infra/kafkaevent"
	"github.com/studio/platform/internal/infra/outbox"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/observability/httpserver"
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

	pool, err := postgresinfra.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	redisClient, err := redisinfra.NewClient(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	publisher, err := eventbus.NewTransportPublisher(cfg, redisClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize event transport publisher", zap.Error(err))
	}
	defer func() { _ = publisher.Close() }()

	relay := outbox.NewRelay(pool, publisher, logger, 100, time.Second, 10)
	go relay.Start(ctx)

	httpPort := cfg.Observability.OutboxHTTPPort
	if httpPort == 0 {
		httpPort = 18055
	}
	obsServer := httpserver.New("outbox-relay", httpPort, logger, map[string]httpserver.CheckFunc{
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

	logger.Info("Outbox relay started",
		zap.Bool("kafka_enabled", cfg.Kafka.Enabled),
		zap.String("kafka_brokers", cfg.Kafka.Brokers),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	logger.Info("Outbox relay stopped")
}
