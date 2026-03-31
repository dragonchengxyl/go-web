package eventbus

import (
	"fmt"

	redisClient "github.com/redis/go-redis/v9"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/kafkaevent"
	"github.com/studio/platform/internal/infra/streams"
	"go.uber.org/zap"
)

func NewPublisher(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger) (Publisher, error) {
	if cfg != nil && cfg.Kafka.Enabled {
		_ = logger
		return &NoopPublisher{}, nil
	}
	if redis == nil {
		return nil, fmt.Errorf("eventbus: redis client is required when kafka is disabled")
	}
	return streams.NewPublisher(redis), nil
}

func NewTransportPublisher(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger) (Publisher, error) {
	if cfg != nil && cfg.Kafka.Enabled {
		return kafkaevent.NewPublisher(cfg.Kafka, logger)
	}
	if redis == nil {
		return nil, fmt.Errorf("eventbus: redis client is required when kafka is disabled")
	}
	return streams.NewPublisher(redis), nil
}

func NewConsumer(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger, consumerID string) (Consumer, error) {
	if cfg != nil && cfg.Kafka.Enabled {
		return kafkaevent.NewConsumer(cfg.Kafka, logger, consumerID)
	}
	if redis == nil {
		return nil, fmt.Errorf("eventbus: redis client is required when kafka is disabled")
	}
	return streams.NewConsumer(redis, logger, consumerID), nil
}
