package audiowork

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type Work struct {
	ID              uuid.UUID      `json:"id"`
	AuthorID        uuid.UUID      `json:"author_id"`
	SourceJobID     uuid.UUID      `json:"source_job_id"`
	Title           string         `json:"title"`
	Description     *string        `json:"description,omitempty"`
	CoverImageURL   *string        `json:"cover_image_url,omitempty"`
	AudioURL        string         `json:"audio_url"`
	DurationSec     float64        `json:"duration_sec"`
	Visibility      Visibility     `json:"visibility"`
	Tags            []string       `json:"tags,omitempty"`
	WaveformPreview []float64      `json:"waveform_preview,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	PublishedAt     time.Time      `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	AuthorUsername string `json:"author_username,omitempty"`
}

type ListFilter struct {
	AuthorID   *uuid.UUID
	Visibility *Visibility
	Page       int
	PageSize   int
}

var (
	ErrNotFound         = errors.New("audio work not found")
	ErrForbidden        = errors.New("not authorized to access this audio work")
	ErrAlreadyPublished = errors.New("audio job already published")
)
