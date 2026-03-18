package resilience

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerOpensAndRecovers(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond)
	if !cb.Allow() {
		t.Fatalf("expected closed breaker to allow")
	}

	cb.RecordFailure(errors.New("first"))
	if !cb.Allow() {
		t.Fatalf("breaker should still allow before threshold")
	}

	cb.RecordFailure(errors.New("second"))
	if cb.Allow() {
		t.Fatalf("breaker should be open after threshold")
	}

	time.Sleep(12 * time.Millisecond)
	if !cb.Allow() {
		t.Fatalf("breaker should allow after cooldown")
	}

	cb.RecordSuccess()
	if got := cb.Snapshot().State; got != StateClosed {
		t.Fatalf("state = %s, want closed", got)
	}
}
