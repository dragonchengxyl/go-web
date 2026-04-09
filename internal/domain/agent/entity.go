package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	RunStatusQueued          RunStatus = "queued"
	RunStatusRunning         RunStatus = "running"
	RunStatusWaitingApproval RunStatus = "waiting_approval"
	RunStatusCompleted       RunStatus = "completed"
	RunStatusFailed          RunStatus = "failed"
	RunStatusCancelled       RunStatus = "cancelled"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type ToolAccessLevel string

const (
	ToolAccessReadOnly             ToolAccessLevel = "read_only"
	ToolAccessWriteRequiresApprove ToolAccessLevel = "write_requires_approval"
	ToolAccessAdminOnly            ToolAccessLevel = "admin_only"
)

const (
	ScenarioPostAgent  = "post_agent"
	ScenarioGroupAgent = "group_agent"
)

type Run struct {
	ID              uuid.UUID      `json:"id"`
	UserID          uuid.UUID      `json:"user_id"`
	Title           string         `json:"title"`
	Goal            string         `json:"goal"`
	Scenario        string         `json:"scenario"`
	Status          RunStatus      `json:"status"`
	ContextSnapshot map[string]any `json:"context_snapshot,omitempty"`
	LatestSummary   string         `json:"latest_summary,omitempty"`
	LastError       string         `json:"last_error,omitempty"`
	AttemptCount    int            `json:"attempt_count"`
	MaxAttempts     int            `json:"max_attempts"`
	LeaseOwner      string         `json:"lease_owner,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	NextRetryAt     *time.Time     `json:"next_retry_at,omitempty"`
	LastErrorAt     *time.Time     `json:"last_error_at,omitempty"`
	LeaseExpiresAt  *time.Time     `json:"lease_expires_at,omitempty"`
	HeartbeatAt     *time.Time     `json:"heartbeat_at,omitempty"`
}

type Step struct {
	ID          uuid.UUID      `json:"id"`
	RunID       uuid.UUID      `json:"run_id"`
	StepIndex   int            `json:"step_index"`
	Kind        string         `json:"kind"`
	Title       string         `json:"title"`
	Status      StepStatus     `json:"status"`
	Summary     string         `json:"summary,omitempty"`
	InputData   map[string]any `json:"input_data,omitempty"`
	OutputData  map[string]any `json:"output_data,omitempty"`
	ErrorText   string         `json:"error_text,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

type ToolCall struct {
	ID          uuid.UUID       `json:"id"`
	RunID       uuid.UUID       `json:"run_id"`
	StepID      *uuid.UUID      `json:"step_id,omitempty"`
	ToolName    string          `json:"tool_name"`
	AccessLevel ToolAccessLevel `json:"access_level"`
	Status      StepStatus      `json:"status"`
	InputData   map[string]any  `json:"input_data,omitempty"`
	OutputData  map[string]any  `json:"output_data,omitempty"`
	ErrorText   string          `json:"error_text,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type Approval struct {
	ID         uuid.UUID      `json:"id"`
	RunID      uuid.UUID      `json:"run_id"`
	StepID     *uuid.UUID     `json:"step_id,omitempty"`
	ActionType string         `json:"action_type"`
	Title      string         `json:"title"`
	Status     ApprovalStatus `json:"status"`
	Payload    map[string]any `json:"payload,omitempty"`
	ApprovedBy *uuid.UUID     `json:"approved_by,omitempty"`
	ApprovedAt *time.Time     `json:"approved_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Artifact struct {
	ID             uuid.UUID      `json:"id"`
	RunID          uuid.UUID      `json:"run_id"`
	StepID         *uuid.UUID     `json:"step_id,omitempty"`
	Kind           string         `json:"kind"`
	Title          string         `json:"title"`
	Content        string         `json:"content,omitempty"`
	StructuredData map[string]any `json:"structured_data,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type RunDetail struct {
	Run       *Run        `json:"run"`
	Steps     []*Step     `json:"steps,omitempty"`
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
	Approvals []*Approval `json:"approvals,omitempty"`
	Artifacts []*Artifact `json:"artifacts,omitempty"`
}

var (
	ErrRunNotFound  = errors.New("agent run not found")
	ErrRunForbidden = errors.New("agent run access denied")
)
