package user

import (
	"context"

	"github.com/google/uuid"
)

// ListFilter represents filters for listing users
type ListFilter struct {
	Page     int
	PageSize int
	Role     *Role
	Status   *Status
	Search   string // Search by username or email
}

// Repository defines the interface for user data access
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update updates a user
	Update(ctx context.Context, user *User) error

	// UpdateLastLogin updates last login information
	UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error

	// ExistsByEmail checks if email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if username exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// List retrieves users with pagination and filters
	List(ctx context.Context, filter ListFilter) ([]*User, int64, error)

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error
}
