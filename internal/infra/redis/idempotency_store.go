package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var releaseIdempotencyLockScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

// IdempotencyStore provides short-lived Redis locks for request de-bounce.
type IdempotencyStore struct {
	client *goredis.Client
}

func NewIdempotencyStore(client *goredis.Client) *IdempotencyStore {
	return &IdempotencyStore{client: client}
}

func (s *IdempotencyStore) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	token := uuid.NewString()
	ok, err := s.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, fmt.Errorf("acquire idempotency lock: %w", err)
	}
	return token, ok, nil
}

func (s *IdempotencyStore) Unlock(ctx context.Context, key, token string) error {
	if token == "" {
		return nil
	}
	if err := releaseIdempotencyLockScript.Run(ctx, s.client, []string{key}, token).Err(); err != nil && err != goredis.Nil {
		return fmt.Errorf("release idempotency lock: %w", err)
	}
	return nil
}
