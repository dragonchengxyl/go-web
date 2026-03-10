package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/infra/oss"
)

// OSSService generates OSS upload credentials for frontend direct upload.
type OSSService struct {
	storage oss.StorageService
}

func NewOSSService(storage oss.StorageService) *OSSService {
	return &OSSService{storage: storage}
}

// UploadTokenInput contains parameters for generating an upload token.
type UploadTokenInput struct {
	UserID  uuid.UUID
	Purpose string // e.g. "posts", "avatar"
}

// GenerateUploadToken returns a signed upload policy valid for 5 minutes.
// The key prefix is {purpose}/{userID}/YYYY-MM-DD/ to ensure isolation.
func (s *OSSService) GenerateUploadToken(ctx context.Context, input UploadTokenInput) (*oss.UploadPolicy, error) {
	if input.Purpose == "" {
		input.Purpose = "uploads"
	}
	dir := fmt.Sprintf("%s/%s/%s/", input.Purpose, input.UserID.String(), time.Now().UTC().Format("2006-01-02"))
	policy, err := s.storage.GenerateUploadPolicy(ctx, dir, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("OSSService.GenerateUploadToken: %w", err)
	}
	return policy, nil
}
