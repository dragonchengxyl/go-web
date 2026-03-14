package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger logs HTTP requests with stable structured fields.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		requestID := c.GetString("request_id")
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", duration),
			zap.Int("response_bytes", c.Writer.Size()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
		}
		if userID := c.GetString("user_id"); userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}
		if role := c.GetString("role"); role != "" {
			fields = append(fields, zap.String("role", role))
		}

		c.Header("X-Response-Time", duration.String())

		switch {
		case c.Writer.Status() >= 500:
			logger.Error("http_request", fields...)
		case c.Writer.Status() >= 400 || duration > slowRequestThreshold:
			logger.Warn("http_request", fields...)
		default:
			logger.Info("http_request", fields...)
		}
	}
}
