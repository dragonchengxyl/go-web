package group

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence operations for groups.
type Repository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*Group, error)
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]*Group, int64, error)

	// Membership
	AddMember(ctx context.Context, m *GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*GroupMember, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role GroupRole) error
	ListMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*GroupMember, int64, error)
	CreateAnnouncement(ctx context.Context, item *GroupAnnouncement) error
	ListAnnouncements(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*GroupAnnouncement, int64, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*Group, int64, error)
	ListByRole(ctx context.Context, userID uuid.UUID, role GroupRole, page, pageSize int) ([]*Group, int64, error)
	ListRecentActiveMembers(ctx context.Context, groupID uuid.UUID, limit int) ([]*GroupMember, error)

	// Counts
	IncrementMemberCount(ctx context.Context, groupID uuid.UUID) error
	DecrementMemberCount(ctx context.Context, groupID uuid.UUID) error
	IncrementPostCount(ctx context.Context, groupID uuid.UUID) error
	DecrementPostCount(ctx context.Context, groupID uuid.UUID) error
}
