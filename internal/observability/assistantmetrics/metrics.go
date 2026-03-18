package assistantmetrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	retrievalDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "assistant_retrieval_duration_seconds",
		Help:    "Assistant retrieval latency in seconds.",
		Buckets: prometheus.DefBuckets,
	})
	firstTokenLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "assistant_first_token_latency_seconds",
		Help:    "Assistant first token latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"provider", "fallback"})
	fallbackTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "assistant_fallback_total",
		Help: "Total number of assistant fallback responses.",
	}, []string{"reason"})
	feedbackTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "assistant_feedback_total",
		Help: "Total number of assistant feedback submissions.",
	}, []string{"value"})
	multimodalDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "assistant_multimodal_duration_seconds",
		Help:    "Assistant multimodal analysis latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"purpose", "cache_hit", "fallback"})
	multimodalRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "assistant_multimodal_requests_total",
		Help: "Total number of assistant multimodal analysis requests.",
	}, []string{"purpose"})
)

// Snapshot is an in-memory runtime view used by the admin console.
type Snapshot struct {
	RetrievalsTotal         uint64            `json:"retrievals_total"`
	LastRetrievalDurationMs float64           `json:"last_retrieval_duration_ms"`
	LastRetrievedDocuments  int               `json:"last_retrieved_documents"`
	FirstTokenObservedTotal uint64            `json:"first_token_observed_total"`
	LastFirstTokenLatencyMs float64           `json:"last_first_token_latency_ms"`
	FallbackTotal           uint64            `json:"fallback_total"`
	FallbackByReason        map[string]uint64 `json:"fallback_by_reason"`
	FeedbackTotal           uint64            `json:"feedback_total"`
	FeedbackByValue         map[string]uint64 `json:"feedback_by_value"`
	LastIndexSyncDurationMs float64           `json:"last_index_sync_duration_ms"`
	LastIndexedDocuments    int               `json:"last_indexed_documents"`
	LastIndexSyncedAt       *time.Time        `json:"last_index_synced_at,omitempty"`
	LastIndexError          string            `json:"last_index_error,omitempty"`
	MultimodalRequestsTotal uint64            `json:"multimodal_requests_total"`
	MultimodalCacheHits     uint64            `json:"multimodal_cache_hits"`
	MultimodalRetryTotal    uint64            `json:"multimodal_retry_total"`
	MultimodalFallbackTotal uint64            `json:"multimodal_fallback_total"`
	LastMultimodalLatencyMs float64           `json:"last_multimodal_latency_ms"`
	LastMultimodalError     string            `json:"last_multimodal_error,omitempty"`
	ChatCircuitState        string            `json:"chat_circuit_state,omitempty"`
	VisionCircuitState      string            `json:"vision_circuit_state,omitempty"`
}

type runtimeState struct {
	retrievalsTotal         atomic.Uint64
	firstTokenObservedTotal atomic.Uint64
	fallbackTotal           atomic.Uint64
	feedbackTotal           atomic.Uint64
	lastRetrievalNs         atomic.Int64
	lastRetrievedDocuments  atomic.Int64
	lastFirstTokenNs        atomic.Int64
	lastIndexSyncNs         atomic.Int64
	lastIndexedDocuments    atomic.Int64
	multimodalRequests      atomic.Uint64
	multimodalCacheHits     atomic.Uint64
	multimodalRetryTotal    atomic.Uint64
	multimodalFallbackTotal atomic.Uint64
	lastMultimodalNs        atomic.Int64
	mu                      sync.Mutex
	fallbackByReason        map[string]uint64
	feedbackByValue         map[string]uint64
	lastIndexSyncedAt       *time.Time
	lastIndexError          string
	lastMultimodalError     string
	chatCircuitState        string
	visionCircuitState      string
}

var state = runtimeState{
	fallbackByReason: make(map[string]uint64),
	feedbackByValue:  make(map[string]uint64),
}

// ObserveRetrieval records retrieval latency and result count.
func ObserveRetrieval(duration time.Duration, resultCount int) {
	retrievalDuration.Observe(duration.Seconds())
	state.retrievalsTotal.Add(1)
	state.lastRetrievalNs.Store(duration.Nanoseconds())
	state.lastRetrievedDocuments.Store(int64(resultCount))
}

// ObserveFirstToken records the first token latency.
func ObserveFirstToken(provider string, fallback bool, duration time.Duration) {
	fallbackLabel := "false"
	if fallback {
		fallbackLabel = "true"
	}
	firstTokenLatency.WithLabelValues(provider, fallbackLabel).Observe(duration.Seconds())
	state.firstTokenObservedTotal.Add(1)
	state.lastFirstTokenNs.Store(duration.Nanoseconds())
}

// RecordFallback increments the fallback counters.
func RecordFallback(reason string) {
	fallbackTotal.WithLabelValues(reason).Inc()
	state.fallbackTotal.Add(1)
	state.mu.Lock()
	state.fallbackByReason[reason]++
	state.mu.Unlock()
}

// RecordFeedback increments the feedback counters.
func RecordFeedback(value string) {
	feedbackTotal.WithLabelValues(value).Inc()
	state.feedbackTotal.Add(1)
	state.mu.Lock()
	state.feedbackByValue[value]++
	state.mu.Unlock()
}

// RecordIndexSync stores the latest index sync status.
func RecordIndexSync(duration time.Duration, indexedDocuments int, err error) {
	state.lastIndexSyncNs.Store(duration.Nanoseconds())
	state.lastIndexedDocuments.Store(int64(indexedDocuments))
	now := time.Now()

	state.mu.Lock()
	state.lastIndexSyncedAt = &now
	if err != nil {
		state.lastIndexError = err.Error()
	} else {
		state.lastIndexError = ""
	}
	state.mu.Unlock()
}

// RecordMultimodal records multimodal latency and resilience metadata.
func RecordMultimodal(purpose string, duration time.Duration, cacheHit, fallback bool, retries int, err error) {
	cacheLabel := "false"
	if cacheHit {
		cacheLabel = "true"
		state.multimodalCacheHits.Add(1)
	}
	fallbackLabel := "false"
	if fallback {
		fallbackLabel = "true"
		state.multimodalFallbackTotal.Add(1)
	}
	if retries > 0 {
		state.multimodalRetryTotal.Add(uint64(retries))
	}
	multimodalRequestsTotal.WithLabelValues(purpose).Inc()
	multimodalDuration.WithLabelValues(purpose, cacheLabel, fallbackLabel).Observe(duration.Seconds())
	state.multimodalRequests.Add(1)
	state.lastMultimodalNs.Store(duration.Nanoseconds())

	state.mu.Lock()
	if err != nil {
		state.lastMultimodalError = err.Error()
	} else {
		state.lastMultimodalError = ""
	}
	state.mu.Unlock()
}

// RecordCircuitState captures latest circuit breaker states for admin visibility.
func RecordCircuitState(kind, circuitState string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	switch kind {
	case "chat":
		state.chatCircuitState = circuitState
	case "vision":
		state.visionCircuitState = circuitState
	}
}

// GetSnapshot returns the current runtime snapshot.
func GetSnapshot() Snapshot {
	state.mu.Lock()
	defer state.mu.Unlock()

	fallbackByReason := make(map[string]uint64, len(state.fallbackByReason))
	for k, v := range state.fallbackByReason {
		fallbackByReason[k] = v
	}
	feedbackByValue := make(map[string]uint64, len(state.feedbackByValue))
	for k, v := range state.feedbackByValue {
		feedbackByValue[k] = v
	}

	var lastSyncedAt *time.Time
	if state.lastIndexSyncedAt != nil {
		t := *state.lastIndexSyncedAt
		lastSyncedAt = &t
	}

	return Snapshot{
		RetrievalsTotal:         state.retrievalsTotal.Load(),
		LastRetrievalDurationMs: nsToMs(state.lastRetrievalNs.Load()),
		LastRetrievedDocuments:  int(state.lastRetrievedDocuments.Load()),
		FirstTokenObservedTotal: state.firstTokenObservedTotal.Load(),
		LastFirstTokenLatencyMs: nsToMs(state.lastFirstTokenNs.Load()),
		FallbackTotal:           state.fallbackTotal.Load(),
		FallbackByReason:        fallbackByReason,
		FeedbackTotal:           state.feedbackTotal.Load(),
		FeedbackByValue:         feedbackByValue,
		LastIndexSyncDurationMs: nsToMs(state.lastIndexSyncNs.Load()),
		LastIndexedDocuments:    int(state.lastIndexedDocuments.Load()),
		LastIndexSyncedAt:       lastSyncedAt,
		LastIndexError:          state.lastIndexError,
		MultimodalRequestsTotal: state.multimodalRequests.Load(),
		MultimodalCacheHits:     state.multimodalCacheHits.Load(),
		MultimodalRetryTotal:    state.multimodalRetryTotal.Load(),
		MultimodalFallbackTotal: state.multimodalFallbackTotal.Load(),
		LastMultimodalLatencyMs: nsToMs(state.lastMultimodalNs.Load()),
		LastMultimodalError:     state.lastMultimodalError,
		ChatCircuitState:        state.chatCircuitState,
		VisionCircuitState:      state.visionCircuitState,
	}
}

func nsToMs(ns int64) float64 {
	if ns <= 0 {
		return 0
	}
	return float64(ns) / float64(time.Millisecond)
}
