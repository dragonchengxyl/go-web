package kafkaevent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/infra/events"
	"github.com/studio/platform/internal/infra/eventspec"
	"github.com/studio/platform/internal/observability/eventbusmetrics"
	"go.uber.org/zap"
)

type Consumer struct {
	cfg        configs.KafkaConfig
	logger     *zap.Logger
	consumerID string
	reader     *kafka.Reader
}

func NewConsumer(cfg configs.KafkaConfig, logger *zap.Logger, consumerID string) (*Consumer, error) {
	if len(parseBrokers(cfg.Brokers)) == 0 {
		return nil, fmt.Errorf("kafka consumer: no brokers configured")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Consumer{
		cfg:        cfg,
		logger:     logger,
		consumerID: consumerID,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, group string, handler events.HandlerFunc) error {
	if c.reader != nil {
		_ = c.reader.Close()
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     parseBrokers(c.cfg.Brokers),
		GroupID:     group,
		GroupTopics: eventspec.TopicsForConsumerGroup(c.cfg, group),
		MinBytes:    minBytes(c.cfg.ReadMinBytes),
		MaxBytes:    maxBytes(c.cfg.ReadMaxBytes),
		StartOffset: startOffset(c.cfg.StartOffset),
	})
	c.reader = reader

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("kafka consumer: fetch message: %w", err)
		}

		var ev events.Event
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			c.logger.Error("kafka consumer: invalid event payload", zap.Error(err))
			eventbusmetrics.RecordConsume("kafka", group, msg.Topic, "invalid", "decode_error")
			if commitErr := reader.CommitMessages(ctx, msg); commitErr != nil && ctx.Err() == nil {
				c.logger.Error("kafka consumer: commit malformed message failed", zap.Error(commitErr))
			}
			continue
		}

		if err := handler(ctx, ev); err != nil {
			eventbusmetrics.RecordConsume("kafka", group, msg.Topic, ev.Type, "handler_error")
			c.logger.Error("kafka consumer handler error",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
			)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			eventbusmetrics.RecordConsume("kafka", group, msg.Topic, ev.Type, "commit_error")
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("kafka consumer: commit message: %w", err)
		}
		eventbusmetrics.RecordConsume("kafka", group, msg.Topic, ev.Type, "ok")
	}
}

func (c *Consumer) Close() error {
	if c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

func minBytes(value int) int {
	if value <= 0 {
		return 1024
	}
	return value
}

func maxBytes(value int) int {
	if value <= 0 {
		return 10e6
	}
	return value
}

func startOffset(value string) int64 {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "oldest", "first", "earliest":
		return kafka.FirstOffset
	default:
		return kafka.LastOffset
	}
}
