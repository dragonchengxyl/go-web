package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type AgentHandler struct {
	service      *usecase.AgentService
	auditService *usecase.AuditService
}

func NewAgentHandler(service *usecase.AgentService, auditService *usecase.AuditService) *AgentHandler {
	return &AgentHandler{service: service, auditService: auditService}
}

func (h *AgentHandler) CreateRun(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req usecase.CreateAgentRunInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	detail, err := h.service.CreateRun(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.logAudit(c, userID, audit.ActionCreate, detail.Run.ID, gin.H{
		"scenario": detail.Run.Scenario,
		"title":    detail.Run.Title,
		"status":   detail.Run.Status,
	})
	response.Success(c, detail)
}

func (h *AgentHandler) ListRuns(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	page, pageSize := getPageParams(c)
	data, err := h.service.ListRuns(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AgentHandler) GetRun(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	detail, err := h.service.GetRunDetail(c.Request.Context(), userID, runID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *AgentHandler) CancelRun(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	detail, err := h.service.CancelRun(c.Request.Context(), userID, runID)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.logAudit(c, userID, audit.ActionUpdate, runID, gin.H{
		"status": detail.Run.Status,
	})
	response.Success(c, detail)
}

func (h *AgentHandler) logAudit(c *gin.Context, userID uuid.UUID, action audit.Action, runID uuid.UUID, afterData any) {
	if h.auditService == nil {
		return
	}
	username := strings.TrimSpace(c.GetString("username"))
	if username == "" {
		username = "user"
	}
	_ = h.auditService.Log(c.Request.Context(), usecase.LogInput{
		UserID:     &userID,
		Username:   username,
		Action:     action,
		Resource:   audit.ResourceAssistant,
		ResourceID: &runID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		AfterData:  afterData,
	})
}
