package audiojob

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeAIMusic      TaskType = "ai_music"
	TaskTypeVoiceConvert TaskType = "voice_convert"
	TaskTypeVoiceEnhance TaskType = "voice_enhance"
	TaskTypeAudioMaster  TaskType = "audio_master"
)

type Status string

const (
	StatusQueued       Status = "queued"
	StatusRunning      Status = "running"
	StatusSucceeded    Status = "succeeded"
	StatusFailed       Status = "failed"
	StatusDeadLettered Status = "dead_lettered"
)

type Job struct {
	ID                uuid.UUID      `json:"id"`
	UserID            uuid.UUID      `json:"user_id"`
	Title             string         `json:"title"`
	TaskType          TaskType       `json:"task_type"`
	Status            Status         `json:"status"`
	SourceAudioURL    *string        `json:"source_audio_url,omitempty"`
	ReferenceAudioURL *string        `json:"reference_audio_url,omitempty"`
	Prompt            *string        `json:"prompt,omitempty"`
	Params            map[string]any `json:"params,omitempty"`
	Result            map[string]any `json:"result,omitempty"`
	ErrorMessage      *string        `json:"error_message,omitempty"`
	AttemptCount      int            `json:"attempt_count"`
	MaxAttempts       int            `json:"max_attempts"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	StartedAt         *time.Time     `json:"started_at,omitempty"`
	FinishedAt        *time.Time     `json:"finished_at,omitempty"`
	NextRetryAt       *time.Time     `json:"next_retry_at,omitempty"`
	LastErrorAt       *time.Time     `json:"last_error_at,omitempty"`
	DeadLetteredAt    *time.Time     `json:"dead_lettered_at,omitempty"`
}

type ListFilter struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
	Status   *Status
	TaskType *TaskType
}

var (
	ErrNotFound          = errors.New("audio job not found")
	ErrForbidden         = errors.New("not authorized to access this audio job")
	ErrInvalidTaskType   = errors.New("invalid audio job task type")
	ErrInvalidTransition = errors.New("invalid audio job state transition")
)
