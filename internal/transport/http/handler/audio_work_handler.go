package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audiowork"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type AudioWorkHandler struct {
	service *usecase.AudioWorkService
}

func NewAudioWorkHandler(service *usecase.AudioWorkService) *AudioWorkHandler {
	return &AudioWorkHandler{service: service}
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

	work, err := h.service.GetPublicWork(c.Request.Context(), workID)
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
