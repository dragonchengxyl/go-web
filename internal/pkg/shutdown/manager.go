package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Manager manages graceful shutdown of the application
type Manager struct {
	logger    *zap.Logger
	timeout   time.Duration
	callbacks []func(context.Context) error
	mu        sync.Mutex
}

// NewManager creates a new shutdown manager
func NewManager(logger *zap.Logger, timeout time.Duration) *Manager {
	return &Manager{
		logger:    logger,
		timeout:   timeout,
		callbacks: make([]func(context.Context) error, 0),
	}
}

// Register registers a shutdown callback
func (m *Manager) Register(callback func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

// Wait waits for shutdown signal and executes callbacks
func (m *Manager) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-quit
	m.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	m.executeCallbacks(ctx)
}

// executeCallbacks executes all registered callbacks
func (m *Manager) executeCallbacks(ctx context.Context) {
	m.mu.Lock()
	callbacks := make([]func(context.Context) error, len(m.callbacks))
	copy(callbacks, m.callbacks)
	m.mu.Unlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(callbacks))

	for i, callback := range callbacks {
		wg.Add(1)
		go func(index int, cb func(context.Context) error) {
			defer wg.Done()
			m.logger.Info("Executing shutdown callback", zap.Int("index", index))

			if err := cb(ctx); err != nil {
				m.logger.Error("Shutdown callback failed",
					zap.Int("index", index),
					zap.Error(err))
				errors <- err
			} else {
				m.logger.Info("Shutdown callback completed", zap.Int("index", index))
			}
		}(i, callback)
	}

	// Wait for all callbacks to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("All shutdown callbacks completed successfully")
	case <-ctx.Done():
		m.logger.Warn("Shutdown timeout exceeded, forcing shutdown")
	}

	close(errors)

	// Log any errors
	for err := range errors {
		m.logger.Error("Shutdown error", zap.Error(err))
	}
}
