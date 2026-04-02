package agent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateRun(ctx context.Context, item *Run) error
	GetRunByID(ctx context.Context, id uuid.UUID) (*Run, error)
	ListRuns(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Run, int64, error)
	UpdateRunStatus(ctx context.Context, id uuid.UUID, status RunStatus, summary, lastError string, completedAt *time.Time) error

	CreateStep(ctx context.Context, item *Step) error
	ListSteps(ctx context.Context, runID uuid.UUID) ([]*Step, error)

	CreateToolCall(ctx context.Context, item *ToolCall) error
	ListToolCalls(ctx context.Context, runID uuid.UUID) ([]*ToolCall, error)

	CreateApproval(ctx context.Context, item *Approval) error
	ListApprovals(ctx context.Context, runID uuid.UUID) ([]*Approval, error)

	CreateArtifact(ctx context.Context, item *Artifact) error
	ListArtifacts(ctx context.Context, runID uuid.UUID) ([]*Artifact, error)
}
