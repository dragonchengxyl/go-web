package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PerformanceMonitor creates a performance monitoring middleware
type PerformanceMonitor struct {
	logger *zap.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *zap.Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger: logger,
	}
}

// Monitor returns a middleware that monitors request performance
func (pm *PerformanceMonitor) Monitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Log performance metrics
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		// Add user info if available
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Log slow requests (> 1 second)
		if duration > time.Second {
			pm.logger.Warn("Slow request detected", fields...)
		} else if statusCode >= 500 {
			pm.logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			pm.logger.Warn("Client error", fields...)
		} else {
			pm.logger.Info("Request completed", fields...)
		}

		// Set response headers for monitoring
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Request-ID", c.GetString("request_id"))
	}
}

// Metrics holds performance metrics
type Metrics struct {
	TotalRequests   int64
	TotalErrors     int64
	TotalDuration   time.Duration
	SlowRequests    int64
	AverageDuration time.Duration
}

// MetricsCollector collects performance metrics
type MetricsCollector struct {
	metrics map[string]*Metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metrics),
	}
}

// Collect returns a middleware that collects metrics
func (mc *MetricsCollector) Collect() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Update metrics
		if _, exists := mc.metrics[path]; !exists {
			mc.metrics[path] = &Metrics{}
		}

		m := mc.metrics[path]
		m.TotalRequests++
		m.TotalDuration += duration

		if statusCode >= 400 {
			m.TotalErrors++
		}

		if duration > time.Second {
			m.SlowRequests++
		}

		m.AverageDuration = time.Duration(int64(m.TotalDuration) / m.TotalRequests)
	}
}

// GetMetrics returns collected metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metrics {
	return mc.metrics
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.metrics = make(map[string]*Metrics)
}
