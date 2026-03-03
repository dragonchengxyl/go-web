package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/studio/platform/configs"
)

// NewClient creates a new Redis client
func NewClient(ctx context.Context, cfg configs.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
