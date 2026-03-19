package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audiowork"
	"github.com/studio/platform/internal/domain/bookmark"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type AudioWorkHandler struct {
	service         *usecase.AudioWorkService
	bookmarkService *usecase.BookmarkService
}

func NewAudioWorkHandler(service *usecase.AudioWorkService, bookmarkService *usecase.BookmarkService) *AudioWorkHandler {
	return &AudioWorkHandler{service: service, bookmarkService: bookmarkService}
}

// ListPublicWorks handles GET /api/v1/audio/works.
func (h *AudioWorkHandler) ListPublicWorks(c *gin.Context) {
	page, pageSize := getPageParams(c)
	items, total, err := h.service.ListPublicWorks(c.Request.Context(), usecase.ListAudioWorksInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithPagination(c, items, int(total), page, pageSize)
}

// GetPublicWork handles GET /api/v1/audio/works/:id.
func (h *AudioWorkHandler) GetPublicWork(c *gin.Context) {
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}

	var viewerID *uuid.UUID
	if value, ok := getUserID(c); ok {
		viewerID = &value
	}
	work, err := h.service.GetWorkForViewer(c.Request.Context(), workID, viewerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, work)
}

// ListMyWorks handles GET /api/v1/users/me/audio/works.
func (h *AudioWorkHandler) ListMyWorks(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	items, total, err := h.service.ListMyWorks(c.Request.Context(), userID, usecase.ListAudioWorksInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithPagination(c, items, int(total), page, pageSize)
}

// ListUserWorks handles GET /api/v1/users/:id/audio/works.
func (h *AudioWorkHandler) ListUserWorks(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的用户ID"))
		return
	}
	page, pageSize := getPageParams(c)
	items, total, err := h.service.ListUserPublicWorks(c.Request.Context(), userID, usecase.ListAudioWorksInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithPagination(c, items, int(total), page, pageSize)
}

// PublishFromJob handles POST /api/v1/audio/jobs/:id/publish.
func (h *AudioWorkHandler) PublishFromJob(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频任务ID"))
		return
	}

	var req struct {
		Title         string   `json:"title"`
		Description   string   `json:"description"`
		CoverImageURL string   `json:"cover_image_url"`
		Visibility    string   `json:"visibility"`
		Tags          []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	work, err := h.service.PublishFromJob(c.Request.Context(), usecase.PublishAudioWorkInput{
		UserID:        userID,
		JobID:         jobID,
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		Visibility:    audiowork.Visibility(req.Visibility),
		Tags:          req.Tags,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, work)
}

// UpdateWork handles PUT /api/v1/audio/works/:id.
func (h *AudioWorkHandler) UpdateWork(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}

	var req struct {
		Title         string   `json:"title" binding:"required"`
		Description   string   `json:"description"`
		CoverImageURL string   `json:"cover_image_url"`
		Visibility    string   `json:"visibility"`
		Tags          []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	work, err := h.service.UpdateWork(c.Request.Context(), usecase.UpdateAudioWorkInput{
		UserID:        userID,
		WorkID:        workID,
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		Visibility:    audiowork.Visibility(req.Visibility),
		Tags:          req.Tags,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, work)
}

// DeleteWork handles DELETE /api/v1/audio/works/:id.
func (h *AudioWorkHandler) DeleteWork(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}
	if err := h.service.DeleteWork(c.Request.Context(), userID, workID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "音频作品已删除"})
}

// LikeWork handles POST /api/v1/audio/works/:id/like.
func (h *AudioWorkHandler) LikeWork(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}
	if err := h.service.LikeWork(c.Request.Context(), userID, workID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "点赞成功"})
}

// UnlikeWork handles DELETE /api/v1/audio/works/:id/like.
func (h *AudioWorkHandler) UnlikeWork(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}
	if err := h.service.UnlikeWork(c.Request.Context(), userID, workID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "取消点赞成功"})
}

// GetMeState handles GET /api/v1/audio/works/:id/me-state.
func (h *AudioWorkHandler) GetMeState(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的音频作品ID"))
		return
	}

	liked, err := h.service.HasLiked(c.Request.Context(), userID, workID)
	if err != nil {
		response.Error(c, err)
		return
	}

	bookmarked := false
	if h.bookmarkService != nil {
		value, bookmarkErr := h.bookmarkService.Exists(c.Request.Context(), userID, bookmark.TargetAudioWork, workID)
		if bookmarkErr == nil {
			bookmarked = value
		}
	}
	response.Success(c, gin.H{"liked": liked, "bookmarked": bookmarked})
}
