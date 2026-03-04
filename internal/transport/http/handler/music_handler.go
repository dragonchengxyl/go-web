package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/music"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// MusicHandler handles music-related HTTP requests
type MusicHandler struct {
	musicService *usecase.MusicService
}

// NewMusicHandler creates a new MusicHandler
func NewMusicHandler(musicService *usecase.MusicService) *MusicHandler {
	return &MusicHandler{
		musicService: musicService,
	}
}

// CreateAlbum creates a new album (Admin only)
func (h *MusicHandler) CreateAlbum(c *gin.Context) {
	var input usecase.CreateAlbumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	album, err := h.musicService.CreateAlbum(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, album)
}

// GetAlbum retrieves an album by ID
func (h *MusicHandler) GetAlbum(c *gin.Context) {
	albumID, err := uuid.Parse(c.Param("album_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的专辑ID"))
		return
	}

	album, err := h.musicService.GetAlbum(c.Request.Context(), albumID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, album)
}

// GetAlbumBySlug retrieves an album by slug
func (h *MusicHandler) GetAlbumBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "专辑标识不能为空"))
		return
	}

	album, err := h.musicService.GetAlbumBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, album)
}

// ListAlbums retrieves albums with pagination and filters
func (h *MusicHandler) ListAlbums(c *gin.Context) {
	var input usecase.ListAlbumsInput
	input.Page = 1
	input.PageSize = 20

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			input.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			input.PageSize = ps
		}
	}

	if gameID := c.Query("game_id"); gameID != "" {
		if gid, err := uuid.Parse(gameID); err == nil {
			input.GameID = &gid
		}
	}

	if albumType := c.Query("album_type"); albumType != "" {
		at := music.AlbumType(albumType)
		input.AlbumType = &at
	}

	input.Search = c.Query("search")

	output, err := h.musicService.ListAlbums(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, output)
}

// GetAlbumTracks retrieves tracks for an album
func (h *MusicHandler) GetAlbumTracks(c *gin.Context) {
	albumID, err := uuid.Parse(c.Param("album_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的专辑ID"))
		return
	}

	tracks, err := h.musicService.GetAlbumTracks(c.Request.Context(), albumID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, tracks)
}

// GetTrack retrieves a track by ID
func (h *MusicHandler) GetTrack(c *gin.Context) {
	trackID, err := uuid.Parse(c.Param("track_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的音轨ID"))
		return
	}

	track, err := h.musicService.GetTrack(c.Request.Context(), trackID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, track)
}

// StreamTrack streams a track (increments play count)
func (h *MusicHandler) StreamTrack(c *gin.Context) {
	trackID, err := uuid.Parse(c.Param("track_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的音轨ID"))
		return
	}

	// Increment play count
	if err := h.musicService.IncrementPlayCount(c.Request.Context(), trackID); err != nil {
		_ = c.Error(err)
	}

	// TODO: Generate streaming URL from OSS
	response.Success(c, gin.H{
		"stream_url": "https://example.com/stream/placeholder",
		"message":    "OSS integration pending",
	})
}
