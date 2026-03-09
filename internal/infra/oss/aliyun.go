package oss

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // Aliyun OSS requires SHA1 for signature
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/studio/platform/configs"
)

// AliyunOSS implements StorageService using Aliyun Object Storage Service.
// It generates pre-signed URLs without requiring the Aliyun SDK.
type AliyunOSS struct {
	cfg configs.OSSConfig
}

// NewAliyunOSS creates a new AliyunOSS storage service.
func NewAliyunOSS(cfg configs.OSSConfig) *AliyunOSS {
	return &AliyunOSS{cfg: cfg}
}

// GeneratePresignedURL generates a temporary pre-signed GET URL for the given OSS object key.
func (a *AliyunOSS) GeneratePresignedURL(_ context.Context, objectKey string, expires time.Duration) (string, error) {
	expiry := time.Now().Add(expires).Unix()

	// StringToSign for OSS GET request
	stringToSign := fmt.Sprintf("GET\n\n\n%d\n/%s/%s", expiry, a.cfg.Bucket, objectKey)

	// HMAC-SHA1 signature
	mac := hmac.New(sha1.New, []byte(a.cfg.AccessKeySecret)) //nolint:gosec
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Build pre-signed URL: https://{bucket}.{endpoint}/{object}?...
	rawURL := fmt.Sprintf("https://%s.%s/%s", a.cfg.Bucket, a.cfg.Endpoint, objectKey)
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid oss url: %w", err)
	}

	q := u.Query()
	q.Set("OSSAccessKeyId", a.cfg.AccessKeyID)
	q.Set("Expires", fmt.Sprintf("%d", expiry))
	q.Set("Signature", signature)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
