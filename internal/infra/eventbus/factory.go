package eventbus

import (
	"fmt"

	redisClient "github.com/redis/go-redis/v9"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/kafkaevent"
	"go.uber.org/zap"
)

func NewPublisher(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger) (Publisher, error) {
	_ = redis
	if cfg == nil || !cfg.Kafka.Enabled {
		return nil, fmt.Errorf("eventbus: kafka must be enabled for business event publishing")
	}
	_ = logger
	return &NoopPublisher{}, nil
}

func NewTransportPublisher(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger) (Publisher, error) {
	_ = redis
	if cfg == nil || !cfg.Kafka.Enabled {
		return nil, fmt.Errorf("eventbus: kafka must be enabled for transport publishing")
	}
	return kafkaevent.NewPublisher(cfg.Kafka, logger)
}

func NewConsumer(cfg *configs.Config, redis *redisClient.Client, logger *zap.Logger, consumerID string) (Consumer, error) {
	_ = redis
	if cfg == nil || !cfg.Kafka.Enabled {
		return nil, fmt.Errorf("eventbus: kafka must be enabled for event consumers")
	}
	return kafkaevent.NewConsumer(cfg.Kafka, logger, consumerID)
}
