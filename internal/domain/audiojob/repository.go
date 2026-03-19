package audiojob

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	ListByUser(ctx context.Context, filter ListFilter) ([]*Job, int64, error)
	Update(ctx context.Context, job *Job) error
	ClaimForProcessing(ctx context.Context, id uuid.UUID, now time.Time) (*Job, bool, error)
	ListDueRetryIDs(ctx context.Context, readyBefore time.Time, limit int) ([]uuid.UUID, error)
}
