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
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/infra/kafkaevent"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/observability/httpserver"
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

	httpPort := cfg.Observability.NotificationHTTPPort
	if httpPort == 0 {
		httpPort = 18052
	}
	obsServer := httpserver.New("notification-svc", httpPort, logger, map[string]httpserver.CheckFunc{
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

	// Consume notification-triggering events from the configured event bus.
	consumer, err := eventbus.NewConsumer(cfg, redisClient, logger, "notification-svc-1")
	if err != nil {
		logger.Fatal("Failed to initialize event consumer", zap.Error(err))
	}
	defer func() { _ = consumer.Close() }()

	go func() {
		logger.Info("Starting notification stream consumer")
		_ = consumer.Start(ctx, eventbus.GroupNotification, func(ctx context.Context, ev eventbus.Event) error {
			eventID, err := uuid.Parse(ev.EventID)
			if err != nil {
				return fmt.Errorf("notification-svc: invalid event_id %q: %w", ev.EventID, err)
			}
			switch ev.Type {
			case eventbus.EventUserFollowed:
				var p eventbus.UserFollowedPayload
				if err := json.Unmarshal(ev.Payload, &p); err != nil {
					return fmt.Errorf("notification-svc: unmarshal user.followed: %w", err)
				}
				followeeID, err := uuid.Parse(p.FolloweeID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid followee_id: %w", err)
				}
				followerID, err := uuid.Parse(p.FollowerID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid follower_id: %w", err)
				}
				return notificationService.NotifyFromEvent(ctx, eventbus.GroupNotification, eventID, &notification.Notification{
					UserID:     followeeID,
					Type:       notification.TypeFollow,
					ActorID:    &followerID,
					TargetID:   &followerID,
					TargetType: "user",
				})

			case eventbus.EventPostLiked:
				var p eventbus.PostLikedPayload
				if err := json.Unmarshal(ev.Payload, &p); err != nil {
					return fmt.Errorf("notification-svc: unmarshal post.liked: %w", err)
				}
				authorID, err := uuid.Parse(p.AuthorID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid author_id: %w", err)
				}
				actorID, err := uuid.Parse(p.ActorID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid actor_id: %w", err)
				}
				postID, err := uuid.Parse(p.PostID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid post_id: %w", err)
				}
				return notificationService.NotifyFromEvent(ctx, eventbus.GroupNotification, eventID, &notification.Notification{
					UserID:     authorID,
					Type:       notification.TypeLike,
					ActorID:    &actorID,
					TargetID:   &postID,
					TargetType: "post",
				})

			case eventbus.EventCommentCreated:
				var p eventbus.CommentCreatedPayload
				if err := json.Unmarshal(ev.Payload, &p); err != nil {
					return fmt.Errorf("notification-svc: unmarshal comment.created: %w", err)
				}
				targetUserID, err := uuid.Parse(p.TargetUserID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid target_user_id: %w", err)
				}
				authorID, err := uuid.Parse(p.AuthorID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid author_id: %w", err)
				}
				postID, err := uuid.Parse(p.PostID)
				if err != nil {
					return fmt.Errorf("notification-svc: invalid post_id: %w", err)
				}
				return notificationService.NotifyFromEvent(ctx, eventbus.GroupNotification, eventID, &notification.Notification{
					UserID:     targetUserID,
					Type:       notification.TypeComment,
					ActorID:    &authorID,
					TargetID:   &postID,
					TargetType: "post",
				})

			case eventbus.EventTipSent:
				var p eventbus.TipSentPayload
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
				return notificationService.NotifyFromEvent(ctx, eventbus.GroupNotification, eventID, &notification.Notification{
					UserID:  receiverID,
					Type:    notification.TypeTip,
					ActorID: &senderID,
				})

			case eventbus.EventPostModerated:
				var p eventbus.PostModeratedPayload
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

				return notificationService.NotifyFromEvent(ctx, eventbus.GroupNotification, eventID, &notification.Notification{
					UserID:     authorID,
					Type:       notification.TypeSystem,
					TargetID:   &postID,
					TargetType: targetType,
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
