package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// EventHandler handles event HTTP endpoints.
type EventHandler struct {
	eventSvc *usecase.EventService
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(eventSvc *usecase.EventService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc}
}

// ListEvents handles GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	page, pageSize := getPageParams(c)
	events, total, err := h.eventSvc.ListEvents(c.Request.Context(), usecase.ListEventsInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"events": events, "total": total, "page": page, "page_size": pageSize})
}

// GetEvent handles GET /api/v1/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的活动ID"))
		return
	}
	e, err := h.eventSvc.GetEvent(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, e)
}

type createEventRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Location    string   `json:"location"`
	IsOnline    bool     `json:"is_online"`
	StartTime   string   `json:"start_time" binding:"required"`
	EndTime     string   `json:"end_time" binding:"required"`
	MaxCapacity int      `json:"max_capacity"`
	Tags        []string `json:"tags"`
}

// CreateEvent handles POST /api/v1/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest(err.Error()))
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.Error(c, apperr.BadRequest("开始时间格式无效，请使用 RFC3339"))
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.Error(c, apperr.BadRequest("结束时间格式无效，请使用 RFC3339"))
		return
	}

	e, err := h.eventSvc.CreateEvent(c.Request.Context(), usecase.CreateEventInput{
		OrganizerID: userID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		IsOnline:    req.IsOnline,
		StartTime:   startTime,
		EndTime:     endTime,
		MaxCapacity: req.MaxCapacity,
		Tags:        req.Tags,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, e)
}

type updateEventRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Location    string            `json:"location"`
	IsOnline    bool              `json:"is_online"`
	StartTime   string            `json:"start_time"`
	EndTime     string            `json:"end_time"`
	MaxCapacity int               `json:"max_capacity"`
	Tags        []string          `json:"tags"`
	Status      event.EventStatus `json:"status"`
}

// UpdateEvent handles PUT /api/v1/events/:id
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的活动ID"))
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest(err.Error()))
		return
	}

	input := usecase.UpdateEventInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		IsOnline:    req.IsOnline,
		MaxCapacity: req.MaxCapacity,
		Tags:        req.Tags,
		Status:      req.Status,
	}
	if req.StartTime != "" {
		if t, parseErr := time.Parse(time.RFC3339, req.StartTime); parseErr == nil {
			input.StartTime = t
		}
	}
	if req.EndTime != "" {
		if t, parseErr := time.Parse(time.RFC3339, req.EndTime); parseErr == nil {
			input.EndTime = t
		}
	}

	e, err := h.eventSvc.UpdateEvent(c.Request.Context(), userID, id, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, e)
}

// CancelEvent handles DELETE /api/v1/events/:id
func (h *EventHandler) CancelEvent(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的活动ID"))
		return
	}
	if err := h.eventSvc.CancelEvent(c.Request.Context(), userID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

type attendRequest struct {
	Status event.AttendeeStatus `json:"status"`
}

// AttendEvent handles POST /api/v1/events/:id/attend
func (h *EventHandler) AttendEvent(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的活动ID"))
		return
	}

	var req attendRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		req.Status = event.AttendeeStatusAttending
	}

	if err := h.eventSvc.AttendEvent(c.Request.Context(), userID, id, req.Status); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListAttendees handles GET /api/v1/events/:id/attendees
func (h *EventHandler) ListAttendees(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的活动ID"))
		return
	}
	page, pageSize := getPageParams(c)
	attendees, total, err := h.eventSvc.ListAttendees(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"attendees": attendees, "total": total})
}

// MyEvents handles GET /api/v1/users/me/events
func (h *EventHandler) MyEvents(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	page, pageSize := getPageParams(c)
	events, total, err := h.eventSvc.MyEvents(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"events": events, "total": total})
}

// MyAttending handles GET /api/v1/users/me/attending
func (h *EventHandler) MyAttending(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	page, pageSize := getPageParams(c)
	events, total, err := h.eventSvc.MyAttending(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"events": events, "total": total})
}
