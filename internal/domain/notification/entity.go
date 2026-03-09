package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Type represents the kind of notification
type Type string

const (
	TypeLike    Type = "like"
	TypeComment Type = "comment"
	TypeFollow  Type = "follow"
	TypeTip     Type = "tip"
	TypeSystem  Type = "system"
)

// Notification represents a user notification
type Notification struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	ActorID    *uuid.UUID `json:"actor_id,omitempty"`
	Type       Type       `json:"type"`
	TargetID   *uuid.UUID `json:"target_id,omitempty"`
	TargetType string     `json:"target_type,omitempty"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`

	// Joined fields
	ActorUsername  string  `json:"actor_username,omitempty"`
	ActorAvatarKey *string `json:"actor_avatar_key,omitempty"`
}

// Repository is the persistence interface for notifications
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Notification, int64, error)
	MarkRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
}

// Notifier allows other services to create notifications
type Notifier interface {
	Notify(ctx context.Context, n *Notification) error
}

var ErrNotFound = errors.New("notification not found")
