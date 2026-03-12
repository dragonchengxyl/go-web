package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/notification"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/infra/streams"
	transportgrpc "github.com/studio/platform/internal/transport/grpc"
	"github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	notificationv1 "github.com/studio/platform/proto/notification/v1"
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

	notificationRepo := postgresinfra.NewNotificationRepository(pool)

	hub := ws.NewDistributedHub(redisClient, logger)
	hubCtx, hubCancel := context.WithCancel(ctx)
	defer hubCancel()
	go hub.Run(hubCtx)

	notificationService := usecase.NewNotificationService(notificationRepo, hub)

	port := cfg.GRPC.NotificationPort
	if port == 0 {
		port = 50052
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	notificationv1.RegisterNotificationServiceServer(s, transportgrpc.NewNotificationServer(notificationService))

	go func() {
		logger.Info("Notification gRPC service starting", zap.Int("port", port))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Consume notification-triggering events from the Redis stream.
	consumer := streams.NewConsumer(redisClient, logger, "notification-svc-1")

	go func() {
		logger.Info("Starting notification stream consumer")
		_ = consumer.Start(ctx, streams.GroupNotification, func(ctx context.Context, ev streams.StreamEvent) error {
			switch ev.Type {
			case streams.EventTipSent:
				var p streams.TipSentPayload
				if err := json.Unmarshal(ev.Payload, &p); err != nil {
					return fmt.Errorf("notification-svc: unmarshal tip.sent: %w", err)
				}
				receiverID, err := uuid.Parse(p.ReceiverID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid receiver_id: %w", err)
				}
				senderID, err := uuid.Parse(p.SenderID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid sender_id: %w", err)
				}
				return notificationService.Notify(ctx, &notification.Notification{
					UserID:    receiverID,
					Type:      notification.TypeTip,
					ActorID:   &senderID,
					CreatedAt: time.Now(),
				})

			case streams.EventPostModerated:
				var p streams.PostModeratedPayload
				if err := json.Unmarshal(ev.Payload, &p); err != nil {
					return fmt.Errorf("notification-svc: unmarshal post.moderated: %w", err)
				}
				authorID, err := uuid.Parse(p.AuthorID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid author_id: %w", err)
				}
				postID, err := uuid.Parse(p.PostID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid post_id: %w", err)
				}

				targetType := "post_blocked"
				if p.Status == "approved" {
					targetType = "post_approved"
				}

				return notificationService.Notify(ctx, &notification.Notification{
					UserID:     authorID,
					Type:       notification.TypeSystem,
					TargetID:   &postID,
					TargetType: targetType,
					CreatedAt:  time.Now(),
				})
			}
			return nil
		})
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	logger.Info("Shutting down notification service...")
	s.GracefulStop()
	logger.Info("Notification service stopped")
}
