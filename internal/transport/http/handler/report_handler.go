package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/report"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
)

type ReportHandler struct {
	repo report.Repository
}

func NewReportHandler(repo report.Repository) *ReportHandler {
	return &ReportHandler{repo: repo}
}

// CreateReport POST /api/v1/reports
func (h *ReportHandler) CreateReport(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		TargetType  string `json:"target_type" binding:"required"`
		TargetID    string `json:"target_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的目标ID"))
		return
	}

	targetType := report.TargetType(req.TargetType)
	if targetType != report.TargetTypePost && targetType != report.TargetTypeComment && targetType != report.TargetTypeUser {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的举报类型"))
		return
	}

	rep := &report.Report{
		ID:          uuid.New(),
		ReporterID:  uid,
		TargetType:  targetType,
		TargetID:    targetID,
		Reason:      req.Reason,
		Description: req.Description,
		Status:      report.StatusPending,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), rep); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "举报已提交，感谢你的反馈"})
}
