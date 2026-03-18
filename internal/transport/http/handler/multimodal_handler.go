package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type MultimodalHandler struct {
	service      *usecase.MultimodalService
	auditService *usecase.AuditService
}

func NewMultimodalHandler(service *usecase.MultimodalService, auditService *usecase.AuditService) *MultimodalHandler {
	return &MultimodalHandler{service: service, auditService: auditService}
}

// AnalyzeMedia handles POST /api/v1/assistant/media/analyze.
func (h *MultimodalHandler) AnalyzeMedia(c *gin.Context) {
	if h.service == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		MediaURLs []string `json:"media_urls" binding:"required"`
		Purpose   string   `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	result, err := h.service.AnalyzeMedia(c.Request.Context(), req.MediaURLs, req.Purpose)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, err.Error()))
		return
	}
	response.Success(c, result)
}

// ExplainModeration handles POST /api/v1/admin/assistant/tools/moderation-explanation.
func (h *MultimodalHandler) ExplainModeration(c *gin.Context) {
	if h.service == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		PostID string `json:"post_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}
	postID, err := uuid.Parse(req.PostID)
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	result, err := h.service.ExplainPostModeration(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "生成审核解释失败", err))
		return
	}

	if h.auditService != nil {
		var userID *uuid.UUID
		var username = "admin"
		if value, ok := getUserID(c); ok {
			userID = &value
			if raw := c.GetString("username"); raw != "" {
				username = raw
			}
		}
		_ = h.auditService.Log(c.Request.Context(), usecase.LogInput{
			UserID:     userID,
			Username:   username,
			Action:     audit.ActionCreate,
			Resource:   audit.ResourceAssistant,
			ResourceID: &postID,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			AfterData: gin.H{
				"tool":     "moderation_explanation",
				"post_id":  postID,
				"run_id":   result.RunID,
				"fallback": result.Fallback,
			},
		})
	}
	response.Success(c, result)
}
