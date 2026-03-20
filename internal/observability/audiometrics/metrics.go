package audiometrics

import (
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	playbackEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "audio_playback_events_total",
		Help: "Total number of audio playback events.",
	}, []string{"event", "source_kind", "authenticated"})
	playbackPositionSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "audio_playback_position_seconds",
		Help:    "Playback position reported by audio playback events.",
		Buckets: []float64{0, 5, 15, 30, 60, 120, 300, 600, 1200, 1800, 3600},
	}, []string{"event"})
)

type Snapshot struct {
	PlaybackEventsTotal uint64            `json:"playback_events_total"`
	EventsByType        map[string]uint64 `json:"events_by_type"`
	LastEvent           string            `json:"last_event,omitempty"`
	LastSourceKind      string            `json:"last_source_kind,omitempty"`
	LastPositionSec     float64           `json:"last_position_sec"`
}

type runtimeState struct {
	playbackEventsTotal atomic.Uint64
	lastPositionMilli   atomic.Int64
	mu                  sync.Mutex
	eventsByType        map[string]uint64
	lastEvent           string
	lastSourceKind      string
}

var state = runtimeState{
	eventsByType: make(map[string]uint64),
}

func RecordPlaybackEvent(event, sourceKind string, authenticated bool, positionSec float64) {
	authLabel := "false"
	if authenticated {
		authLabel = "true"
	}
	if sourceKind == "" {
		sourceKind = "unknown"
	}

	playbackEventsTotal.WithLabelValues(event, sourceKind, authLabel).Inc()
	if positionSec >= 0 {
		playbackPositionSeconds.WithLabelValues(event).Observe(positionSec)
		state.lastPositionMilli.Store(int64(positionSec * 1000))
	}

	state.playbackEventsTotal.Add(1)
	state.mu.Lock()
	state.eventsByType[event]++
	state.lastEvent = event
	state.lastSourceKind = sourceKind
	state.mu.Unlock()
}

func GetSnapshot() Snapshot {
	state.mu.Lock()
	defer state.mu.Unlock()

	eventsByType := make(map[string]uint64, len(state.eventsByType))
	for k, v := range state.eventsByType {
		eventsByType[k] = v
	}

	return Snapshot{
		PlaybackEventsTotal: state.playbackEventsTotal.Load(),
		EventsByType:        eventsByType,
		LastEvent:           state.lastEvent,
		LastSourceKind:      state.lastSourceKind,
		LastPositionSec:     float64(state.lastPositionMilli.Load()) / 1000,
	}
}
