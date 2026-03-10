package report

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TargetType string

const (
	TargetTypePost    TargetType = "post"
	TargetTypeComment TargetType = "comment"
	TargetTypeUser    TargetType = "user"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusReviewed  Status = "reviewed"
	StatusDismissed Status = "dismissed"
)

type Report struct {
	ID               uuid.UUID  `json:"id"`
	ReporterID       uuid.UUID  `json:"reporter_id"`
	ReporterUsername string     `json:"reporter_username,omitempty"`
	TargetType       TargetType `json:"target_type"`
	TargetID         uuid.UUID  `json:"target_id"`
	Reason           string     `json:"reason"`
	Description      string     `json:"description,omitempty"`
	Status           Status     `json:"status"`
	ReviewedBy       *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, r *Report) error
	List(ctx context.Context, status string, page, size int) ([]*Report, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status, reviewedBy uuid.UUID) error
}

var ErrAlreadyReported = errors.New("already reported")
