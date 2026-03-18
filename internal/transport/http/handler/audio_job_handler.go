package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audiojob"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type AudioJobHandler struct {
	service *usecase.AudioJobService
}

func NewAudioJobHandler(service *usecase.AudioJobService) *AudioJobHandler {
	return &AudioJobHandler{service: service}
}

// CreateJob handles POST /api/v1/audio/jobs
func (h *AudioJobHandler) CreateJob(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		Title             string         `json:"title" binding:"required"`
		TaskType          string         `json:"task_type" binding:"required"`
		SourceAudioURL    string         `json:"source_audio_url"`
		ReferenceAudioURL string         `json:"reference_audio_url"`
		Prompt            string         `json:"prompt"`
		Params            map[string]any `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	job, err := h.service.CreateJob(c.Request.Context(), usecase.CreateAudioJobInput{
		UserID:            userID,
		Title:             req.Title,
		TaskType:          audiojob.TaskType(req.TaskType),
		SourceAudioURL:    req.SourceAudioURL,
		ReferenceAudioURL: req.ReferenceAudioURL,
		Prompt:            req.Prompt,
		Params:            req.Params,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, job)
}

// ListJobs handles GET /api/v1/audio/jobs
func (h *AudioJobHandler) ListJobs(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	page, pageSize := getPageParams(c)
	var status *audiojob.Status
	if raw := c.Query("status"); raw != "" {
		value := audiojob.Status(raw)
		status = &value
	}
	var taskType *audiojob.TaskType
	if raw := c.Query("task_type"); raw != "" {
		value := audiojob.TaskType(raw)
		taskType = &value
	}

	items, total, err := h.service.ListJobs(c.Request.Context(), usecase.ListAudioJobsInput{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		TaskType: taskType,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithPagination(c, items, int(total), page, pageSize)
}

// GetJob handles GET /api/v1/audio/jobs/:id
func (h *AudioJobHandler) GetJob(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的任务ID"))
		return
	}

	job, err := h.service.GetJob(c.Request.Context(), userID, jobID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, job)
}

// RetryJob handles POST /api/v1/audio/jobs/:id/retry
func (h *AudioJobHandler) RetryJob(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的任务ID"))
		return
	}

	job, err := h.service.RetryJob(c.Request.Context(), userID, jobID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, job)
}
