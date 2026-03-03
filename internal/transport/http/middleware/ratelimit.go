package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	client *redis.Client
	config configs.RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client, config configs.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		client: client,
		config: config,
	}
}

// Limit returns a rate limiting middleware
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine identifier (user_id or IP)
		identifier := c.ClientIP()
		limit := rl.config.Unauthenticated

		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%s", userID)
			limit = rl.config.Authenticated

			// Check if admin
			if role, exists := c.Get("role"); exists && role == "admin" {
				limit = rl.config.Admin
			}
		}

		// Check rate limit
		allowed, remaining, err := rl.checkLimit(c.Request.Context(), identifier, limit)
		if err != nil {
			// Fail open: allow request if rate limiter fails
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			c.Header("Retry-After", "60")
			response.Error(c, apperr.ErrRateLimited)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkLimit checks if request is within rate limit
func (rl *RateLimiter) checkLimit(ctx context.Context, identifier string, limit int) (bool, int, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now().Unix()
	window := int64(60) // 1 minute window

	// Use Lua script for atomic operation
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local current = redis.call('GET', key)
		if current == false then
			redis.call('SET', key, 1, 'EX', window)
			return {1, limit - 1}
		end

		current = tonumber(current)
		if current < limit then
			redis.call('INCR', key)
			return {1, limit - current - 1}
		end

		return {0, 0}
	`

	result, err := rl.client.Eval(ctx, script, []string{key}, limit, window, now).Result()
	if err != nil {
		return false, 0, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	remaining := int(values[1].(int64))

	return allowed, remaining, nil
}
