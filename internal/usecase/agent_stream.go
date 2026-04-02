package usecase

import (
	"sync"

	"github.com/google/uuid"
)

type AgentRunEvent struct {
	RunID   string `json:"run_id"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`
}

type agentRunHub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[chan AgentRunEvent]struct{}
}

func newAgentRunHub() *agentRunHub {
	return &agentRunHub{
		subscribers: make(map[uuid.UUID]map[chan AgentRunEvent]struct{}),
	}
}

func (h *agentRunHub) Subscribe(runID uuid.UUID) chan AgentRunEvent {
	ch := make(chan AgentRunEvent, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subscribers[runID] == nil {
		h.subscribers[runID] = make(map[chan AgentRunEvent]struct{})
	}
	h.subscribers[runID][ch] = struct{}{}
	return ch
}

func (h *agentRunHub) Unsubscribe(runID uuid.UUID, ch chan AgentRunEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	items := h.subscribers[runID]
	if items == nil {
		close(ch)
		return
	}
	delete(items, ch)
	if len(items) == 0 {
		delete(h.subscribers, runID)
	}
	close(ch)
}

func (h *agentRunHub) Publish(runID uuid.UUID, event AgentRunEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	items := h.subscribers[runID]
	for ch := range items {
		select {
		case ch <- event:
		default:
		}
	}
}
