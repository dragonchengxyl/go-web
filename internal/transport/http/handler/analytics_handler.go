package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/analytics"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AnalyticsHandler handles analytics requests
type AnalyticsHandler struct {
	analyticsService *usecase.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService *usecase.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// TrackEvent tracks an analytics event
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var req struct {
		EventType  analytics.EventType `json:"event_type" binding:"required"`
		SessionID  string              `json:"session_id" binding:"required"`
		Properties map[string]any      `json:"properties"`
		Referrer   string              `json:"referrer"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uuid.UUID); ok {
			userID = &id
		}
	}

	event := &analytics.Event{
		EventType:  req.EventType,
		UserID:     userID,
		SessionID:  req.SessionID,
		Properties: req.Properties,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Referrer:   req.Referrer,
	}

	if err := h.analyticsService.TrackEvent(c.Request.Context(), event); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// GetDashboardMetrics retrieves dashboard metrics
func (h *AnalyticsHandler) GetDashboardMetrics(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		response.Error(c, err)
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		response.Error(c, err)
		return
	}

	metrics, err := h.analyticsService.GetDashboardMetrics(c.Request.Context(), start, end)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, metrics)
}

// GetConversionFunnel retrieves conversion funnel data
func (h *AnalyticsHandler) GetConversionFunnel(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		response.Error(c, err)
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		response.Error(c, err)
		return
	}

	funnel, err := h.analyticsService.GetConversionFunnel(c.Request.Context(), start, end)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, funnel)
}

// GetUserRetention retrieves user retention data
func (h *AnalyticsHandler) GetUserRetention(c *gin.Context) {
	cohortDate := c.DefaultQuery("cohort_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	days := c.DefaultQuery("days", "30")

	cohort, err := time.Parse("2006-01-02", cohortDate)
	if err != nil {
		response.Error(c, err)
		return
	}

	var daysInt int
	if _, err := fmt.Sscanf(days, "%d", &daysInt); err != nil {
		daysInt = 30
	}

	retention, err := h.analyticsService.GetUserRetention(c.Request.Context(), cohort, daysInt)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, retention)
}

// RefreshMetrics refreshes the daily metrics materialized view
func (h *AnalyticsHandler) RefreshMetrics(c *gin.Context) {
	if err := h.analyticsService.RefreshDailyMetrics(c.Request.Context()); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Metrics refreshed successfully"})
}
