package game

import (
	"context"

	"github.com/google/uuid"
)

// ListFilter represents filters for listing games
type ListFilter struct {
	Page     int
	PageSize int
	Status   *Status
	Genre    string
	Tag      string
	Search   string // Search by title or slug
}

// Repository defines the interface for game data access
type Repository interface {
	// Create creates a new game
	Create(ctx context.Context, game *Game) error

	// GetByID retrieves a game by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Game, error)

	// GetBySlug retrieves a game by slug
	GetBySlug(ctx context.Context, slug string) (*Game, error)

	// Update updates a game
	Update(ctx context.Context, game *Game) error

	// Delete deletes a game by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves games with pagination and filters
	List(ctx context.Context, filter ListFilter) ([]*Game, int64, error)

	// ExistsBySlug checks if slug exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}

// BranchRepository defines the interface for game branch data access
type BranchRepository interface {
	// Create creates a new branch
	Create(ctx context.Context, branch *Branch) error

	// GetByID retrieves a branch by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Branch, error)

	// ListByGameID retrieves branches for a game
	ListByGameID(ctx context.Context, gameID uuid.UUID) ([]*Branch, error)

	// Update updates a branch
	Update(ctx context.Context, branch *Branch) error

	// Delete deletes a branch
	Delete(ctx context.Context, id uuid.UUID) error

	// GetDefaultBranch retrieves the default branch for a game
	GetDefaultBranch(ctx context.Context, gameID uuid.UUID) (*Branch, error)
}

// ReleaseRepository defines the interface for game release data access
type ReleaseRepository interface {
	// Create creates a new release
	Create(ctx context.Context, release *Release) error

	// GetByID retrieves a release by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Release, error)

	// GetByVersion retrieves a release by branch and version
	GetByVersion(ctx context.Context, branchID uuid.UUID, version string) (*Release, error)

	// ListByBranchID retrieves releases for a branch
	ListByBranchID(ctx context.Context, branchID uuid.UUID, publishedOnly bool) ([]*Release, error)

	// Update updates a release
	Update(ctx context.Context, release *Release) error

	// Delete deletes a release
	Delete(ctx context.Context, id uuid.UUID) error

	// GetLatestRelease retrieves the latest published release for a branch
	GetLatestRelease(ctx context.Context, branchID uuid.UUID) (*Release, error)
}
