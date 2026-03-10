package oss

import (
	"context"
	"time"
)

// UploadPolicy contains the credentials needed for frontend direct upload to OSS.
type UploadPolicy struct {
	// Host is the OSS bucket endpoint for the POST request.
	Host string `json:"host"`
	// OSSAccessKeyID is the Aliyun AccessKeyID.
	OSSAccessKeyID string `json:"OSSAccessKeyId"`
	// Policy is the Base64-encoded upload policy JSON.
	Policy string `json:"policy"`
	// Signature is the HMAC-SHA1 signature of the policy.
	Signature string `json:"signature"`
	// Expire is the Unix timestamp when the policy expires.
	Expire int64 `json:"expire"`
	// Dir is the key prefix that uploaded objects must start with.
	Dir string `json:"dir"`
}

// StorageService defines the interface for cloud object storage operations.
type StorageService interface {
	// GeneratePresignedURL generates a temporary pre-signed download/stream URL for the given object key.
	// expires specifies how long the URL remains valid.
	GeneratePresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)

	// GenerateUploadPolicy returns credentials for frontend direct upload.
	// dir is the key prefix (e.g. "posts/uid123/"); expires controls validity window.
	GenerateUploadPolicy(ctx context.Context, dir string, expires time.Duration) (*UploadPolicy, error)
}

