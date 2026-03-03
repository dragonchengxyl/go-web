package user

import (
	"time"

	"github.com/google/uuid"
)

// Role represents user role
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleModerator  Role = "moderator"
	RoleCreator    Role = "creator"
	RolePremium    Role = "premium"
	RolePlayer     Role = "player"
	RoleGuest      Role = "guest"
)

// Status represents user status
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
	StatusBanned    Status = "banned"
)

// User represents a user entity
type User struct {
	ID                uuid.UUID  `json:"id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"`
	AvatarKey         *string    `json:"avatar_key,omitempty"`
	Bio               *string    `json:"bio,omitempty"`
	Location          *string    `json:"location,omitempty"`
	Role              Role       `json:"role"`
	Status            Status     `json:"status"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP       *string    `json:"last_login_ip,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
