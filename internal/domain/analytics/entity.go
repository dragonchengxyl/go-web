package analytics

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of analytics event
type EventType string

const (
	EventPageView         EventType = "page_view"
	EventGameView         EventType = "game_view"
	EventGameDownload     EventType = "game_download"
	EventUserRegister     EventType = "user_register"
	EventUserLogin        EventType = "user_login"
	EventPurchaseInitiate EventType = "purchase_initiate"
	EventPurchaseComplete EventType = "purchase_complete"
	EventMusicPlay        EventType = "music_play"
	EventCommentCreate    EventType = "comment_create"
	EventSearch           EventType = "search"
)

// Event represents an analytics event
type Event struct {
	ID         uuid.UUID         `json:"id"`
	EventType  EventType         `json:"event_type"`
	UserID     *uuid.UUID        `json:"user_id,omitempty"`
	SessionID  string            `json:"session_id"`
	Properties map[string]any    `json:"properties"`
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	Referrer   string            `json:"referrer,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// Repository defines the interface for analytics event storage
type Repository interface {
	// Track records an analytics event
	Track(event *Event) error

	// GetEvents retrieves events with filters
	GetEvents(filter EventFilter) ([]*Event, error)

	// GetEventCount counts events matching the filter
	GetEventCount(filter EventFilter) (int64, error)
}

// EventFilter represents filters for querying events
type EventFilter struct {
	EventType *EventType
	UserID    *uuid.UUID
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}
