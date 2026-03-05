package audit

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for audit log storage
type Repository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *Log) error

	// List retrieves audit logs with filters
	List(ctx context.Context, filter ListFilter) ([]*Log, int64, error)

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Log, error)

	// GetByUserID retrieves audit logs for a specific user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*Log, error)

	// GetByResource retrieves audit logs for a specific resource
	GetByResource(ctx context.Context, resource Resource, resourceID uuid.UUID, limit int) ([]*Log, error)
}

// ListFilter represents filters for listing audit logs
type ListFilter struct {
	UserID     *uuid.UUID
	Action     *Action
	Resource   *Resource
	ResourceID *uuid.UUID
	StartTime  *string
	EndTime    *string
	Page       int
	PageSize   int
}
