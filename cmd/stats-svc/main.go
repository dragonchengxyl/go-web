package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/postgres"
	transportgrpc "github.com/studio/platform/internal/transport/grpc"
	"github.com/studio/platform/internal/usecase"
	statsv1 "github.com/studio/platform/proto/stats/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	statsService := usecase.NewStatsService(pool)
	statsServer := transportgrpc.NewStatsServer(statsService)

	port := cfg.GRPC.StatsPort
	if port == 0 {
		port = 50051
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	statsv1.RegisterStatsServiceServer(s, statsServer)

	logger.Info("Stats service starting", zap.Int("port", port))

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down stats service...")
	s.GracefulStop()
	logger.Info("Stats service stopped")
}
