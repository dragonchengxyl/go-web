package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	httpSlowRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_slow_requests_total",
		Help: "Total number of HTTP requests slower than the service threshold.",
	}, []string{"method", "path", "status"})

	httpRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of in-flight HTTP requests.",
	}, []string{"method", "path"})
)

// PrometheusMetrics returns a Gin middleware that records Prometheus HTTP metrics.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
			if path == "" {
				path = "unknown"
			}
		}

		httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Inc()
		defer httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		if time.Duration(duration*float64(time.Second)) > slowRequestThreshold {
			httpSlowRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		}
	}
}
