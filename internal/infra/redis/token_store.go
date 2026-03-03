package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore manages refresh tokens and blacklist in Redis
type TokenStore struct {
	client *redis.Client
}

// NewTokenStore creates a new TokenStore
func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

// RefreshTokenData represents refresh token metadata
type RefreshTokenData struct {
	UserID    string    `json:"user_id"`
	Device    string    `json:"device"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

// SaveRefreshToken saves a refresh token
func (s *TokenStore) SaveRefreshToken(ctx context.Context, jti, userID, device, ip string, expiry time.Duration) error {
	data := RefreshTokenData{
		UserID:    userID,
		Device:    device,
		IP:        ip,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	key := fmt.Sprintf("auth:refresh:%s", jti)
	if err := s.client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves refresh token data
func (s *TokenStore) GetRefreshToken(ctx context.Context, jti string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("auth:refresh:%s", jti)
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &data, nil
}

// DeleteRefreshToken deletes a refresh token
func (s *TokenStore) DeleteRefreshToken(ctx context.Context, jti string) error {
	key := fmt.Sprintf("auth:refresh:%s", jti)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// BlacklistToken adds a token to the blacklist
func (s *TokenStore) BlacklistToken(ctx context.Context, jti string, expiry time.Duration) error {
	key := fmt.Sprintf("auth:blacklist:%s", jti)
	if err := s.client.Set(ctx, key, "1", expiry).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *TokenStore) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("auth:blacklist:%s", jti)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return exists > 0, nil
}
