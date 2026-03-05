package handler

import (
	"context"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/studio/platform/internal/pkg/response"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *pgxpool.Pool
	redis     *redis.Client
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		startTime: time.Now(),
	}
}

// Health returns service health status
func (h *HealthHandler) Health(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response.Success(c, gin.H{
		"status":  "ok",
		"service": "studio-platform",
		"uptime":  time.Since(h.startTime).String(),
		"memory": gin.H{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
	})
}

// Ready returns service readiness status
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]gin.H)
	allHealthy := true

	// Check database connectivity
	if h.db != nil {
		dbStart := time.Now()
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = gin.H{
				"status":       "unhealthy",
				"error":        err.Error(),
				"response_ms":  time.Since(dbStart).Milliseconds(),
			}
			allHealthy = false
		} else {
			stats := h.db.Stat()
			checks["database"] = gin.H{
				"status":           "healthy",
				"response_ms":      time.Since(dbStart).Milliseconds(),
				"max_connections":  stats.MaxConns(),
				"idle_connections": stats.IdleConns(),
				"total_connections": stats.TotalConns(),
			}
		}
	} else {
		checks["database"] = gin.H{"status": "not configured"}
	}

	// Check Redis connectivity
	if h.redis != nil {
		redisStart := time.Now()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = gin.H{
				"status":      "unhealthy",
				"error":       err.Error(),
				"response_ms": time.Since(redisStart).Milliseconds(),
			}
			allHealthy = false
		} else {
			// Get Redis info
			info, _ := h.redis.Info(ctx, "stats").Result()
			checks["redis"] = gin.H{
				"status":      "healthy",
				"response_ms": time.Since(redisStart).Milliseconds(),
				"info":        info,
			}
		}
	} else {
		checks["redis"] = gin.H{"status": "not configured"}
	}

	status := "ready"
	statusCode := 200
	if !allHealthy {
		status = "not ready"
		statusCode = 503
	}

	c.JSON(statusCode, gin.H{
		"status": status,
		"checks": checks,
		"timestamp": time.Now().Unix(),
	})
}

