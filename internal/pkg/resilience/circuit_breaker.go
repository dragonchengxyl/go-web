package resilience

import (
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half_open"
)

type Snapshot struct {
	State               State      `json:"state"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	LastError           string     `json:"last_error,omitempty"`
	OpenedAt            *time.Time `json:"opened_at,omitempty"`
	RetryAfter          *time.Time `json:"retry_after,omitempty"`
}

type CircuitBreaker struct {
	mu         sync.Mutex
	threshold  int
	openFor    time.Duration
	state      State
	failures   int
	lastError  string
	openedAt   *time.Time
	retryAfter *time.Time
}

func NewCircuitBreaker(threshold int, openFor time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if openFor <= 0 {
		openFor = time.Minute
	}
	return &CircuitBreaker{
		threshold: threshold,
		openFor:   openFor,
		state:     StateClosed,
	}
}

func (c *CircuitBreaker) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateOpen && c.retryAfter != nil && time.Now().After(*c.retryAfter) {
		c.state = StateHalfOpen
		return true
	}
	return c.state != StateOpen
}

func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = StateClosed
	c.failures = 0
	c.lastError = ""
	c.openedAt = nil
	c.retryAfter = nil
}

func (c *CircuitBreaker) RecordFailure(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures++
	if err != nil {
		c.lastError = err.Error()
	}
	if c.failures >= c.threshold {
		now := time.Now()
		retryAfter := now.Add(c.openFor)
		c.state = StateOpen
		c.openedAt = &now
		c.retryAfter = &retryAfter
	}
}

func (c *CircuitBreaker) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	var openedAt *time.Time
	if c.openedAt != nil {
		t := *c.openedAt
		openedAt = &t
	}
	var retryAfter *time.Time
	if c.retryAfter != nil {
		t := *c.retryAfter
		retryAfter = &t
	}

	return Snapshot{
		State:               c.state,
		ConsecutiveFailures: c.failures,
		LastError:           c.lastError,
		OpenedAt:            openedAt,
		RetryAfter:          retryAfter,
	}
}
