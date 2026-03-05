package audit

import (
	"time"

	"github.com/google/uuid"
)

// Action represents the type of action performed
type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionLogin  Action = "login"
	ActionLogout Action = "logout"
	ActionView   Action = "view"
	ActionExport Action = "export"
)

// Resource represents the type of resource being acted upon
type Resource string

const (
	ResourceUser        Resource = "user"
	ResourceGame        Resource = "game"
	ResourceRelease     Resource = "release"
	ResourceProduct     Resource = "product"
	ResourceOrder       Resource = "order"
	ResourceComment     Resource = "comment"
	ResourceAchievement Resource = "achievement"
	ResourceCoupon      Resource = "coupon"
)

// Log represents an audit log entry
type Log struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Username     string     `json:"username"`
	Action       Action     `json:"action"`
	Resource     Resource   `json:"resource"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	BeforeData   *string    `json:"before_data,omitempty"`
	AfterData    *string    `json:"after_data,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
