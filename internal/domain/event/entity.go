package event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// EventStatus represents the lifecycle of an event.
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

// AttendeeStatus represents a user's attendance status.
type AttendeeStatus string

const (
	AttendeeStatusAttending AttendeeStatus = "attending"
	AttendeeStatusMaybe     AttendeeStatus = "maybe"
	AttendeeStatusNotGoing  AttendeeStatus = "not_going"
)

// Event is the core event entity.
type Event struct {
	ID            uuid.UUID   `json:"id"`
	OrganizerID   uuid.UUID   `json:"organizer_id"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Location      string      `json:"location"`
	IsOnline      bool        `json:"is_online"`
	StartTime     time.Time   `json:"start_time"`
	EndTime       time.Time   `json:"end_time"`
	MaxCapacity   int         `json:"max_capacity"`
	Tags          []string    `json:"tags"`
	Status        EventStatus `json:"status"`
	AttendeeCount int         `json:"attendee_count"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// EventAttendee records a user's attendance at an event.
type EventAttendee struct {
	EventID  uuid.UUID      `json:"event_id"`
	UserID   uuid.UUID      `json:"user_id"`
	Status   AttendeeStatus `json:"status"`
	JoinedAt time.Time      `json:"joined_at"`
}

// ListFilter holds filtering options for listing events.
type ListFilter struct {
	OrganizerID *uuid.UUID
	Status      *EventStatus
	Tags        []string
	FromTime    *time.Time
	ToTime      *time.Time
	Page        int
	PageSize    int
}

var ErrNotFound = errors.New("event not found")
