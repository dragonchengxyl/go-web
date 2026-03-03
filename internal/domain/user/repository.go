package user

import (
	"context"

	"github.com/google/uuid"
)

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
}
