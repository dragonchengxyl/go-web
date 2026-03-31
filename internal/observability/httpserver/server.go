package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type CheckFunc func(ctx context.Context) error

type Server struct {
	service   string
	startedAt time.Time
	checks    map[string]CheckFunc
	logger    *zap.Logger
	http      *http.Server
}

func New(service string, port int, logger *zap.Logger, checks map[string]CheckFunc) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	if checks == nil {
		checks = map[string]CheckFunc{}
	}

	s := &Server{
		service:   service,
		startedAt: time.Now(),
		checks:    checks,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)

	s.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) Start() {
	go func() {
		s.logger.Info("observability http server starting",
			zap.String("service", s.service),
			zap.String("addr", s.http.Addr),
		)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("observability http server failed",
				zap.String("service", s.service),
				zap.Error(err),
			)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

type checkResult struct {
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	ResponseMS int64  `json:"response_ms"`
}

type serviceStatus struct {
	Service    string                 `json:"service"`
	Status     string                 `json:"status"`
	Uptime     string                 `json:"uptime"`
	Timestamp  int64                  `json:"timestamp"`
	Goroutines int                    `json:"goroutines"`
	Checks     map[string]checkResult `json:"checks,omitempty"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, serviceStatus{
		Service:    s.service,
		Status:     "ok",
		Uptime:     time.Since(s.startedAt).String(),
		Timestamp:  time.Now().Unix(),
		Goroutines: runtime.NumGoroutine(),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	results := make(map[string]checkResult, len(s.checks))
	healthy := true

	for name, check := range s.checks {
		start := time.Now()
		err := check(ctx)
		result := checkResult{
			Status:     "healthy",
			ResponseMS: time.Since(start).Milliseconds(),
		}
		if err != nil {
			result.Status = "unhealthy"
			result.Error = err.Error()
			healthy = false
		}
		results[name] = result
	}

	statusCode := http.StatusOK
	status := "ready"
	if !healthy {
		statusCode = http.StatusServiceUnavailable
		status = "not_ready"
	}

	writeJSON(w, statusCode, serviceStatus{
		Service:    s.service,
		Status:     status,
		Uptime:     time.Since(s.startedAt).String(),
		Timestamp:  time.Now().Unix(),
		Goroutines: runtime.NumGoroutine(),
		Checks:     results,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
