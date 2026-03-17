package sponsor

import (
	"context"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
)

// Repository persists sponsor display settings for runtime updates.
type Repository interface {
	Get(ctx context.Context) (*configs.SponsorConfig, error)
	Upsert(ctx context.Context, cfg configs.SponsorConfig, updatedBy *uuid.UUID) error
}
