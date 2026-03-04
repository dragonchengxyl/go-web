package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/music"
	"github.com/studio/platform/internal/pkg/apperr"
)

// MusicService handles music-related business logic
type MusicService struct {
	albumRepo music.AlbumRepository
	trackRepo music.TrackRepository
}

// NewMusicService creates a new MusicService
func NewMusicService(albumRepo music.AlbumRepository, trackRepo music.TrackRepository) *MusicService {
	return &MusicService{
		albumRepo: albumRepo,
		trackRepo: trackRepo,
	}
}

// CreateAlbumInput represents input for creating an album
type CreateAlbumInput struct {
	GameID      *uuid.UUID      `json:"game_id,omitempty"`
	Slug        string          `json:"slug" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Subtitle    *string         `json:"subtitle,omitempty"`
	Description *string         `json:"description,omitempty"`
	CoverKey    *string         `json:"cover_key,omitempty"`
	Artist      *string         `json:"artist,omitempty"`
	Composer    *string         `json:"composer,omitempty"`
	Arranger    *string         `json:"arranger,omitempty"`
	Lyricist    *string         `json:"lyricist,omitempty"`
	ReleaseDate *time.Time      `json:"release_date,omitempty"`
	AlbumType   music.AlbumType `json:"album_type"`
}

// CreateAlbum creates a new album
func (s *MusicService) CreateAlbum(ctx context.Context, input CreateAlbumInput) (*music.Album, error) {
	exists, err := s.albumRepo.ExistsBySlug(ctx, input.Slug)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查专辑标识失败", err)
	}
	if exists {
		return nil, apperr.New(apperr.CodeInvalidParam, "专辑标识已存在")
	}

	now := time.Now()
	album := &music.Album{
		ID:          uuid.New(),
		GameID:      input.GameID,
		Slug:        input.Slug,
		Title:       input.Title,
		Subtitle:    input.Subtitle,
		Description: input.Description,
		CoverKey:    input.CoverKey,
		Artist:      input.Artist,
		Composer:    input.Composer,
		Arranger:    input.Arranger,
		Lyricist:    input.Lyricist,
		ReleaseDate: input.ReleaseDate,
		AlbumType:   input.AlbumType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.albumRepo.Create(ctx, album); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建专辑失败", err)
	}

	return album, nil
}

// GetAlbum retrieves an album by ID
func (s *MusicService) GetAlbum(ctx context.Context, albumID uuid.UUID) (*music.Album, error) {
	album, err := s.albumRepo.GetByID(ctx, albumID)
	if err != nil {
		if errors.Is(err, music.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询专辑失败", err)
	}
	return album, nil
}

// GetAlbumBySlug retrieves an album by slug
func (s *MusicService) GetAlbumBySlug(ctx context.Context, slug string) (*music.Album, error) {
	album, err := s.albumRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, music.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询专辑失败", err)
	}
	return album, nil
}

// ListAlbumsInput represents input for listing albums
type ListAlbumsInput struct {
	Page      int
	PageSize  int
	GameID    *uuid.UUID
	AlbumType *music.AlbumType
	Search    string
}

// ListAlbumsOutput represents output for listing albums
type ListAlbumsOutput struct {
	Albums []*music.Album `json:"albums"`
	Total  int64          `json:"total"`
	Page   int            `json:"page"`
	Size   int            `json:"size"`
}

// ListAlbums retrieves albums with pagination and filters
func (s *MusicService) ListAlbums(ctx context.Context, input ListAlbumsInput) (*ListAlbumsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := music.AlbumListFilter{
		Page:      input.Page,
		PageSize:  input.PageSize,
		GameID:    input.GameID,
		AlbumType: input.AlbumType,
		Search:    input.Search,
	}

	albums, total, err := s.albumRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询专辑列表失败", err)
	}

	return &ListAlbumsOutput{
		Albums: albums,
		Total:  total,
		Page:   input.Page,
		Size:   len(albums),
	}, nil
}

// GetAlbumTracks retrieves tracks for an album
func (s *MusicService) GetAlbumTracks(ctx context.Context, albumID uuid.UUID) ([]*music.Track, error) {
	tracks, err := s.trackRepo.ListByAlbumID(ctx, albumID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音轨列表失败", err)
	}
	return tracks, nil
}

// GetTrack retrieves a track by ID
func (s *MusicService) GetTrack(ctx context.Context, trackID uuid.UUID) (*music.Track, error) {
	track, err := s.trackRepo.GetByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, music.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音轨失败", err)
	}
	return track, nil
}

// IncrementPlayCount increments the play count for a track
func (s *MusicService) IncrementPlayCount(ctx context.Context, trackID uuid.UUID) error {
	if err := s.trackRepo.IncrementPlayCount(ctx, trackID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新播放次数失败", err)
	}
	return nil
}

// SearchAlbums searches albums by query string
func (s *MusicService) SearchAlbums(ctx context.Context, query string, limit int) ([]*music.Album, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	input := ListAlbumsInput{
		Page:     1,
		PageSize: limit,
		Search:   query,
	}

	output, err := s.ListAlbums(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Albums, nil
}
