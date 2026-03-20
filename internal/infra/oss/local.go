package oss

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/studio/platform/configs"
)

// LocalStorage implements StorageService for locally served media under /uploads.
type LocalStorage struct {
	cfg configs.OSSConfig
}

// NewLocalStorage creates a local media storage implementation.
func NewLocalStorage(cfg configs.OSSConfig) *LocalStorage {
	return &LocalStorage{cfg: cfg}
}

// GeneratePresignedURL returns a locally served media URL.
func (l *LocalStorage) GeneratePresignedURL(_ context.Context, objectKey string, _ time.Duration) (string, error) {
	trimmed := strings.TrimSpace(objectKey)
	if trimmed == "" {
		return "", fmt.Errorf("local media key is empty")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}
	if strings.HasPrefix(trimmed, "/") {
		return trimmed, nil
	}
	if strings.HasPrefix(trimmed, "uploads/") {
		return "/" + trimmed, nil
	}
	return "/uploads/" + strings.TrimLeft(trimmed, "/"), nil
}

// GenerateUploadPolicy is intentionally unsupported in local mode.
// The frontend should fall back to the authenticated local upload endpoints.
func (l *LocalStorage) GenerateUploadPolicy(_ context.Context, _ string, _ time.Duration) (*UploadPolicy, error) {
	return nil, fmt.Errorf("local storage does not support direct upload policy; use the local upload API")
}
