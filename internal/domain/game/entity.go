package game

import (
	"time"

	"github.com/google/uuid"
)

// Status represents game status
type Status string

const (
	StatusActive     Status = "active"
	StatusArchived   Status = "archived"
	StatusComingSoon Status = "coming_soon"
)

// Game represents a game entity
type Game struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Subtitle    *string    `json:"subtitle,omitempty"`
	Description *string    `json:"description,omitempty"`
	CoverKey    *string    `json:"cover_key,omitempty"`
	BannerKey   *string    `json:"banner_key,omitempty"`
	TrailerURL  *string    `json:"trailer_url,omitempty"`
	Genre       []string   `json:"genre"`
	Tags        []string   `json:"tags"`
	Engine      string     `json:"engine"`
	Status      Status     `json:"status"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	DeveloperID uuid.UUID  `json:"developer_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Screenshot represents a game screenshot
type Screenshot struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	OSSKey    string    `json:"oss_key"`
	Caption   *string   `json:"caption,omitempty"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// Branch represents a game release branch
type Branch struct {
	ID          uuid.UUID `json:"id"`
	GameID      uuid.UUID `json:"game_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

// Release represents a game release version
type Release struct {
	ID             uuid.UUID  `json:"id"`
	BranchID       uuid.UUID  `json:"branch_id"`
	Version        string     `json:"version"`
	Title          *string    `json:"title,omitempty"`
	Changelog      *string    `json:"changelog,omitempty"`
	OSSKey         *string    `json:"oss_key,omitempty"`
	ManifestKey    *string    `json:"manifest_key,omitempty"`
	FileSize       *int64     `json:"file_size,omitempty"`
	ChecksumSHA256 *string    `json:"checksum_sha256,omitempty"`
	MinOSVersion   *string    `json:"min_os_version,omitempty"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DLCType represents DLC type
type DLCType string

const (
	DLCTypeStory    DLCType = "story"
	DLCTypeCostume  DLCType = "costume"
	DLCTypeArtbook  DLCType = "artbook"
	DLCTypeOST      DLCType = "ost"
	DLCTypeBundle   DLCType = "bundle"
)

// DLC represents downloadable content
type DLC struct {
	ID             uuid.UUID  `json:"id"`
	GameID         uuid.UUID  `json:"game_id"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	CoverKey       *string    `json:"cover_key,omitempty"`
	DLCType        DLCType    `json:"dlc_type"`
	PriceCents     int        `json:"price_cents"`
	Currency       string     `json:"currency"`
	IsFree         bool       `json:"is_free"`
	ReleaseDate    *time.Time `json:"release_date,omitempty"`
	OSSKey         *string    `json:"oss_key,omitempty"`
	FileSize       *int64     `json:"file_size,omitempty"`
	ChecksumSHA256 *string    `json:"checksum_sha256,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
