package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/response"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns service health status
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"service": "studio-platform",
	})
}

// Ready returns service readiness status
func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: Check database and Redis connectivity
	response.Success(c, gin.H{
		"status": "ready",
	})
}
