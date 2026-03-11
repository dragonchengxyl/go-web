package group

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// GroupPrivacy controls who can see and join the group.
type GroupPrivacy string

const (
	GroupPrivacyPublic  GroupPrivacy = "public"
	GroupPrivacyPrivate GroupPrivacy = "private"
)

// GroupRole is a member's role within the group.
type GroupRole string

const (
	GroupRoleOwner     GroupRole = "owner"
	GroupRoleModerator GroupRole = "moderator"
	GroupRoleMember    GroupRole = "member"
)

// Group is the core group entity.
type Group struct {
	ID          uuid.UUID    `json:"id"`
	OwnerID     uuid.UUID    `json:"owner_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	AvatarKey   *string      `json:"avatar_key,omitempty"`
	Tags        []string     `json:"tags"`
	Privacy     GroupPrivacy `json:"privacy"`
	MemberCount int          `json:"member_count"`
	PostCount   int          `json:"post_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// GroupMember records a user's membership in a group.
type GroupMember struct {
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     GroupRole `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// ListFilter holds filtering options for listing groups.
type ListFilter struct {
	MemberUserID *uuid.UUID
	Privacy      *GroupPrivacy
	Search       string
	Page         int
	PageSize     int
}

var ErrNotFound = errors.New("group not found")
var ErrNotMember = errors.New("not a member of this group")
var ErrAlreadyMember = errors.New("already a member")
