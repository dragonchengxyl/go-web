package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
)

// GenerateReportSummary handles POST /api/v1/admin/assistant/tools/report-summary.
func (h *AdminHandler) GenerateReportSummary(c *gin.Context) {
	if h.aiToolService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		ReportID string `json:"report_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	reportID, err := uuid.Parse(req.ReportID)
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	result, err := h.aiToolService.GenerateReportSummary(c.Request.Context(), reportID)
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "生成举报摘要失败", err))
		return
	}
	h.logAudit(c, audit.ActionCreate, audit.ResourceAssistant, nil, gin.H{
		"tool":      "report_summary",
		"report_id": reportID,
		"run_id":    result.RunID,
		"fallback":  result.Fallback,
	})
	response.Success(c, result)
}

// GenerateWeeklyReport handles POST /api/v1/admin/assistant/tools/weekly-report.
func (h *AdminHandler) GenerateWeeklyReport(c *gin.Context) {
	if h.aiToolService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		Days int `json:"days"`
	}
	_ = c.ShouldBindJSON(&req)

	result, err := h.aiToolService.GenerateWeeklyReport(c.Request.Context(), req.Days)
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "生成周报失败", err))
		return
	}
	h.logAudit(c, audit.ActionCreate, audit.ResourceAssistant, nil, gin.H{
		"tool":     "weekly_report",
		"days":     req.Days,
		"run_id":   result.RunID,
		"fallback": result.Fallback,
	})
	response.Success(c, result)
}

// GenerateCreatorRecommendation handles POST /api/v1/admin/assistant/tools/creator-recommendation.
func (h *AdminHandler) GenerateCreatorRecommendation(c *gin.Context) {
	if h.aiToolService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	result, err := h.aiToolService.GenerateCreatorRecommendation(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "生成创作者推荐理由失败", err))
		return
	}
	h.logAudit(c, audit.ActionCreate, audit.ResourceAssistant, &userID, gin.H{
		"tool":     "creator_recommendation",
		"user_id":  userID,
		"run_id":   result.RunID,
		"fallback": result.Fallback,
	})
	response.Success(c, result)
}

// GenerateEventCopy handles POST /api/v1/admin/assistant/tools/event-copy.
func (h *AdminHandler) GenerateEventCopy(c *gin.Context) {
	if h.aiToolService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		EventID string `json:"event_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	result, err := h.aiToolService.GenerateEventCopy(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "生成活动文案失败", err))
		return
	}
	h.logAudit(c, audit.ActionCreate, audit.ResourceAssistant, &eventID, gin.H{
		"tool":     "event_copy",
		"event_id": eventID,
		"run_id":   result.RunID,
		"fallback": result.Fallback,
	})
	response.Success(c, result)
}
