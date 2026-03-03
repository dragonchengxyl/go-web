package music

import (
	"time"

	"github.com/google/uuid"
)

// AlbumType represents the type of album
type AlbumType string

const (
	AlbumTypeOST   AlbumType = "ost"
	AlbumTypeDrama AlbumType = "drama"
	AlbumTypeVocal AlbumType = "vocal"
	AlbumTypeBGM   AlbumType = "bgm"
)

// Album represents a music album
type Album struct {
	ID          uuid.UUID  `json:"id"`
	GameID      *uuid.UUID `json:"game_id,omitempty"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Subtitle    *string    `json:"subtitle,omitempty"`
	Description *string    `json:"description,omitempty"`
	CoverKey    *string    `json:"cover_key,omitempty"`
	Artist      *string    `json:"artist,omitempty"`
	Composer    *string    `json:"composer,omitempty"`
	Arranger    *string    `json:"arranger,omitempty"`
	Lyricist    *string    `json:"lyricist,omitempty"`
	TotalTracks *int       `json:"total_tracks,omitempty"`
	DurationSec *int       `json:"duration_sec,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	AlbumType   AlbumType  `json:"album_type"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Track represents a music track
type Track struct {
	ID             uuid.UUID `json:"id"`
	AlbumID        uuid.UUID `json:"album_id"`
	TrackNumber    int       `json:"track_number"`
	DiscNumber     int       `json:"disc_number"`
	Title          string    `json:"title"`
	Artist         *string   `json:"artist,omitempty"`
	DurationSec    *int      `json:"duration_sec,omitempty"`
	StreamKey      *string   `json:"stream_key,omitempty"`
	StreamSize     *int64    `json:"stream_size,omitempty"`
	HiFiKey        *string   `json:"hifi_key,omitempty"`
	HiFiFormat     *string   `json:"hifi_format,omitempty"`
	HiFiBitDepth   *int      `json:"hifi_bitdepth,omitempty"`
	HiFiSampleRate *int      `json:"hifi_samplerate,omitempty"`
	HiFiSize       *int64    `json:"hifi_size,omitempty"`
	LRCKey         *string   `json:"lrc_key,omitempty"`
	PlayCount      int64     `json:"play_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
