package audiowork

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, work *Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*Work, error)
	List(ctx context.Context, filter ListFilter) ([]*Work, int64, error)
}
