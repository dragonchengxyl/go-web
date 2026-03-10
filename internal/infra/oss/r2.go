package oss

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/studio/platform/configs"
)
// CloudflareR2 implements StorageService using Cloudflare R2 (S3-compatible).
type CloudflareR2 struct {
	cfg configs.OSSConfig
}

// NewCloudflareR2 creates a new CloudflareR2 storage service.
func NewCloudflareR2(cfg configs.OSSConfig) *CloudflareR2 {
	return &CloudflareR2{cfg: cfg}
}

// GeneratePresignedURL generates a temporary pre-signed GET URL for the given R2 object key.
// Uses AWS Signature Version 4 (S3-compatible).
func (r *CloudflareR2) GeneratePresignedURL(_ context.Context, objectKey string, expires time.Duration) (string, error) {
	now := time.Now().UTC()
	expirySecs := int64(expires.Seconds())

	// R2 endpoint: https://<account-id>.r2.cloudflarestorage.com/<bucket>
	endpoint := r.cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}
	rawURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(endpoint, "/"), r.cfg.Bucket, objectKey)

	region := r.cfg.Region
	if region == "" {
		region = "auto"
	}

	service := "s3"
	dateShort := now.Format("20060102")
	dateISO := now.Format("20060102T150405Z")
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateShort, region, service)
	credential := fmt.Sprintf("%s/%s", r.cfg.AccessKeyID, credentialScope)

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid r2 url: %w", err)
	}

	q := u.Query()
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", credential)
	q.Set("X-Amz-Date", dateISO)
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", expirySecs))
	q.Set("X-Amz-SignedHeaders", "host")
	u.RawQuery = q.Encode()

	// Canonical request
	host := u.Host
	canonicalURI := "/" + r.cfg.Bucket + "/" + objectKey
	canonicalQueryString := u.RawQuery
	canonicalHeaders := "host:" + host + "\n"
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := strings.Join([]string{
		"GET",
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	// String to sign
	hashCanonical := sha256Hex([]byte(canonicalRequest))
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		dateISO,
		credentialScope,
		hashCanonical,
	}, "\n")

	// Signing key
	signingKey := hmacSHA256(
		hmacSHA256(
			hmacSHA256(
				hmacSHA256(
					[]byte("AWS4"+r.cfg.AccessKeySecret),
					[]byte(dateShort),
				),
				[]byte(region),
			),
			[]byte(service),
		),
		[]byte("aws4_request"),
	)

	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	q.Set("X-Amz-Signature", signature)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateUploadPolicy is not supported for Cloudflare R2 in this implementation.
// Use GeneratePresignedURL for uploads instead.
func (r *CloudflareR2) GenerateUploadPolicy(_ context.Context, _ string, _ time.Duration) (*UploadPolicy, error) {
	return nil, fmt.Errorf("GenerateUploadPolicy is not supported for Cloudflare R2; use GeneratePresignedURL")
}
