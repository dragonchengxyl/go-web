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
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/infra/kafkaevent"
	"github.com/studio/platform/internal/infra/moderation"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/observability/httpserver"
	transportgrpc "github.com/studio/platform/internal/transport/grpc"
	moderationv1 "github.com/studio/platform/proto/moderation/v1"
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

	postRepo := postgresinfra.NewPostRepository(pool)

	if cfg.Moderation.AccessKeyID == "" {
		logger.Fatal("moderation credentials not configured")
	}
	moderator := moderation.NewAliyunGreen(
		cfg.Moderation.AccessKeyID,
		cfg.Moderation.AccessKeySecret,
		cfg.Moderation.Endpoint,
	)

	port := cfg.GRPC.ModerationPort
	if port == 0 {
		port = 50053
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	moderationv1.RegisterModerationServiceServer(s, transportgrpc.NewModerationServer(moderator))

	go func() {
		logger.Info("Moderation gRPC service starting", zap.Int("port", port))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	httpPort := cfg.Observability.ModerationHTTPPort
	if httpPort == 0 {
		httpPort = 18053
	}
	obsServer := httpserver.New("moderation-svc", httpPort, logger, map[string]httpserver.CheckFunc{
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

	// Start configured event bus consumer/publisher for post.created / post.moderated events.
	consumer, err := eventbus.NewConsumer(cfg, redisClient, logger, "moderation-svc-1")
	if err != nil {
		logger.Fatal("Failed to initialize event consumer", zap.Error(err))
	}
	defer func() { _ = consumer.Close() }()

	publisher, err := eventbus.NewPublisher(cfg, redisClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize event publisher", zap.Error(err))
	}
	defer func() { _ = publisher.Close() }()

	go func() {
		logger.Info("Starting moderation stream consumer")
		_ = consumer.Start(ctx, eventbus.GroupModeration, func(ctx context.Context, ev eventbus.Event) error {
			if ev.Type != eventbus.EventPostCreated {
				return nil
			}

			var payload eventbus.PostCreatedPayload
			if err := json.Unmarshal(ev.Payload, &payload); err != nil {
				return fmt.Errorf("moderation-svc: unmarshal post.created: %w", err)
			}

			// Review text content.
			decision, _, reviewErr := moderator.ReviewText(ctx, payload.Content)
			if reviewErr != nil {
				logger.Error("text review failed", zap.Error(reviewErr), zap.String("post_id", payload.PostID))
				decision = "pass" // fail open
			}

			// Review first image if text passed.
			if string(decision) == "pass" && len(payload.MediaURLs) > 0 {
				imgDecision, _, imgErr := moderator.ReviewImage(ctx, payload.MediaURLs[0])
				if imgErr != nil {
					logger.Error("image review failed", zap.Error(imgErr), zap.String("post_id", payload.PostID))
				} else if string(imgDecision) == "block" {
					decision = imgDecision
				}
			}

			modStatus := post.ModerationApproved
			pubStatus := "approved"
			if string(decision) == "block" {
				modStatus = post.ModerationBlocked
				pubStatus = "blocked"
			}

			postID, err := uuid.Parse(payload.PostID)
			if err != nil {
				return fmt.Errorf("moderation-svc: invalid post_id %q: %w", payload.PostID, err)
			}

			if err := postRepo.UpdateModerationStatus(ctx, postID, modStatus); err != nil {
				logger.Error("failed to update moderation status", zap.Error(err), zap.String("post_id", payload.PostID))
			}

			return publisher.Publish(ctx, eventbus.EventPostModerated, eventbus.PostModeratedPayload{
				PostID:   payload.PostID,
				AuthorID: payload.AuthorID,
				Status:   pubStatus,
			})
		})
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	logger.Info("Shutting down moderation service...")
	s.GracefulStop()
	logger.Info("Moderation service stopped")
}
