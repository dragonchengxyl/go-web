package music

import (
	"context"

	"github.com/google/uuid"
)

// AlbumListFilter represents filters for listing albums
type AlbumListFilter struct {
	Page      int
	PageSize  int
	GameID    *uuid.UUID
	AlbumType *AlbumType
	Search    string
}

// AlbumRepository defines the interface for album data access
type AlbumRepository interface {
	// Create creates a new album
	Create(ctx context.Context, album *Album) error

	// GetByID retrieves an album by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Album, error)

	// GetBySlug retrieves an album by slug
	GetBySlug(ctx context.Context, slug string) (*Album, error)

	// Update updates an album
	Update(ctx context.Context, album *Album) error

	// Delete deletes an album
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves albums with pagination and filters
	List(ctx context.Context, filter AlbumListFilter) ([]*Album, int64, error)

	// ExistsBySlug checks if slug exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}

// TrackRepository defines the interface for track data access
type TrackRepository interface {
	// Create creates a new track
	Create(ctx context.Context, track *Track) error

	// GetByID retrieves a track by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Track, error)

	// Update updates a track
	Update(ctx context.Context, track *Track) error

	// Delete deletes a track
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByAlbumID retrieves tracks for an album
	ListByAlbumID(ctx context.Context, albumID uuid.UUID) ([]*Track, error)

	// IncrementPlayCount increments the play count for a track
	IncrementPlayCount(ctx context.Context, trackID uuid.UUID) error
}
