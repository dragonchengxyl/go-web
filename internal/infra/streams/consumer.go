package streams

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// HandlerFunc processes a single stream event.
type HandlerFunc func(ctx context.Context, event StreamEvent) error

// Consumer reads from a Redis stream consumer group.
type Consumer struct {
	client     *redis.Client
	logger     *zap.Logger
	consumerID string
}

// NewConsumer creates a new Consumer.
func NewConsumer(client *redis.Client, logger *zap.Logger, consumerID string) *Consumer {
	return &Consumer{
		client:     client,
		logger:     logger,
		consumerID: consumerID,
	}
}

// EnsureGroup creates the consumer group if it does not exist.
func (c *Consumer) EnsureGroup(ctx context.Context, group string) error {
	err := c.client.XGroupCreateMkStream(ctx, StreamKey, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("streams: create group %q: %w", group, err)
	}
	return nil
}

// Start begins consuming messages from the given group.
// It blocks until ctx is cancelled. Each message is ACKed after handler returns nil.
func (c *Consumer) Start(ctx context.Context, group string, handler HandlerFunc) error {
	if err := c.EnsureGroup(ctx, group); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: c.consumerID,
			Streams:  []string{StreamKey, ">"},
			Count:    10,
			Block:    2 * time.Second,
			NoAck:    false,
		}).Result()

		if err != nil {
			if err == redis.Nil || err.Error() == "redis: nil" {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error("xreadgroup error", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				if err := c.processMessage(ctx, group, msg, handler); err != nil {
					c.logger.Error("event handler error",
						zap.String("id", msg.ID),
						zap.Error(err),
					)
				}
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, group string, msg redis.XMessage, handler HandlerFunc) error {
	raw, ok := msg.Values["data"].(string)
	if !ok {
		// ACK to discard malformed messages
		_ = c.client.XAck(ctx, StreamKey, group, msg.ID).Err()
		return fmt.Errorf("streams: message %q has no data field", msg.ID)
	}

	var ev StreamEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		_ = c.client.XAck(ctx, StreamKey, group, msg.ID).Err()
		return fmt.Errorf("streams: unmarshal event: %w", err)
	}

	if err := handler(ctx, ev); err != nil {
		return err
	}

	return c.client.XAck(ctx, StreamKey, group, msg.ID).Err()
}
