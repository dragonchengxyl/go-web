package oss

import (
	"context"
	"time"
)

// StorageService defines the interface for cloud object storage operations.
type StorageService interface {
	// GeneratePresignedURL generates a temporary pre-signed download/stream URL for the given object key.
	// expires specifies how long the URL remains valid.
	GeneratePresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)
}
