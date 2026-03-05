package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/usecase"
)

// AuditLogger creates an audit logging middleware
type AuditLogger struct {
	auditService *usecase.AuditService
}

// NewAuditLogger creates a new audit logger middleware
func NewAuditLogger(auditService *usecase.AuditService) *AuditLogger {
	return &AuditLogger{
		auditService: auditService,
	}
}

// Log returns a middleware that logs important operations
func (al *AuditLogger) Log(action audit.Action, resource audit.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture request body for logging
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Only log successful operations (2xx status codes)
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}

		// Get user info
		var userID *uuid.UUID
		username := "anonymous"
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = &id
			}
		}
		if uname, exists := c.Get("username"); exists {
			if name, ok := uname.(string); ok {
				username = name
			}
		}

		// Get resource ID from URL params or context
		var resourceID *uuid.UUID
		if id := c.Param("id"); id != "" {
			if parsed, err := uuid.Parse(id); err == nil {
				resourceID = &parsed
			}
		}
		if rid, exists := c.Get("resource_id"); exists {
			if id, ok := rid.(uuid.UUID); ok {
				resourceID = &id
			}
		}

		// Log asynchronously to avoid blocking the response
		go func() {
			ctx := c.Request.Context()
			_ = al.auditService.Log(ctx, usecase.LogInput{
				UserID:     userID,
				Username:   username,
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				IPAddress:  c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
			})
		}()
	}
}

// LogWithData returns a middleware that logs operations with before/after data
func (al *AuditLogger) LogWithData(action audit.Action, resource audit.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get before data if available
		var beforeData any
		if data, exists := c.Get("before_data"); exists {
			beforeData = data
		}

		// Process request
		c.Next()

		// Only log successful operations
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}

		// Get after data if available
		var afterData any
		if data, exists := c.Get("after_data"); exists {
			afterData = data
		}

		// Get user info
		var userID *uuid.UUID
		username := "anonymous"
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = &id
			}
		}
		if uname, exists := c.Get("username"); exists {
			if name, ok := uname.(string); ok {
				username = name
			}
		}

		// Get resource ID
		var resourceID *uuid.UUID
		if id := c.Param("id"); id != "" {
			if parsed, err := uuid.Parse(id); err == nil {
				resourceID = &parsed
			}
		}
		if rid, exists := c.Get("resource_id"); exists {
			if id, ok := rid.(uuid.UUID); ok {
				resourceID = &id
			}
		}

		// Log asynchronously
		go func() {
			ctx := c.Request.Context()
			_ = al.auditService.Log(ctx, usecase.LogInput{
				UserID:     userID,
				Username:   username,
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				IPAddress:  c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
				BeforeData: beforeData,
				AfterData:  afterData,
			})
		}()
	}
}
