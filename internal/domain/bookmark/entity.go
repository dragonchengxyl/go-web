package bookmark

import (
	"time"

	"github.com/google/uuid"
)

// TargetType identifies the bookmarked resource type.
type TargetType string

const (
	TargetPost  TargetType = "post"
	TargetGroup TargetType = "group"
	TargetEvent TargetType = "event"
)

// Bookmark is a generic saved item reference.
type Bookmark struct {
	UserID     uuid.UUID  `json:"user_id"`
	TargetType TargetType `json:"target_type"`
	TargetID   uuid.UUID  `json:"target_id"`
	CreatedAt  time.Time  `json:"created_at"`
}
