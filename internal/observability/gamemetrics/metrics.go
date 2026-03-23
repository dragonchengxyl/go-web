package gamemetrics

import (
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	roomEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hex_blitz_room_events_total",
		Help: "Total room lifecycle events for Hex Blitz.",
	}, []string{"event"})
	scoreReportsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hex_blitz_score_reports_total",
		Help: "Total score report attempts for Hex Blitz.",
	}, []string{"result", "reason"})
	matchesFinishedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hex_blitz_matches_finished_total",
		Help: "Total finished Hex Blitz matches.",
	})
	matchPlayersHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "hex_blitz_match_players",
		Help:    "Player count per finished Hex Blitz match.",
		Buckets: []float64{1, 2, 3, 4},
	})
	activeRoomsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hex_blitz_active_rooms",
		Help: "Current number of active Hex Blitz rooms.",
	})
	activePlayersGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hex_blitz_active_players",
		Help: "Current number of players in active Hex Blitz rooms.",
	})
	activeConnectionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hex_blitz_active_ws_connections",
		Help: "Current number of active Hex Blitz WebSocket connections.",
	})
	roomsByStatusGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hex_blitz_rooms_by_status",
		Help: "Current number of Hex Blitz rooms by status.",
	}, []string{"status"})
)

type Snapshot struct {
	RoomEventsTotal      uint64            `json:"room_events_total"`
	ScoreReportsTotal    uint64            `json:"score_reports_total"`
	RejectedScoreReports uint64            `json:"rejected_score_reports"`
	ActiveRooms          int               `json:"active_rooms"`
	ActivePlayers        int               `json:"active_players"`
	ActiveConnections    int               `json:"active_connections"`
	MatchesFinishedTotal uint64            `json:"matches_finished_total"`
	RoomsByStatus        map[string]int    `json:"rooms_by_status"`
	RoomEventsByType     map[string]uint64 `json:"room_events_by_type"`
	ScoreReportReasons   map[string]uint64 `json:"score_report_reasons"`
}

type runtimeState struct {
	roomEventsTotal      atomic.Uint64
	scoreReportsTotal    atomic.Uint64
	rejectedScoreReports atomic.Uint64
	matchesFinishedTotal atomic.Uint64
	activeRooms          atomic.Int64
	activePlayers        atomic.Int64
	activeConnections    atomic.Int64
	mu                   sync.Mutex
	roomsByStatus        map[string]int
	roomEventsByType     map[string]uint64
	scoreReportReasons   map[string]uint64
}

var state = runtimeState{
	roomsByStatus:      make(map[string]int),
	roomEventsByType:   make(map[string]uint64),
	scoreReportReasons: make(map[string]uint64),
}

func RecordRoomEvent(event string) {
	roomEventsTotal.WithLabelValues(event).Inc()
	state.roomEventsTotal.Add(1)
	state.mu.Lock()
	state.roomEventsByType[event]++
	state.mu.Unlock()
}

func RecordScoreReport(accepted bool, reason string) {
	result := "accepted"
	if !accepted {
		result = "rejected"
		state.rejectedScoreReports.Add(1)
	}
	if reason == "" {
		reason = "none"
	}
	scoreReportsTotal.WithLabelValues(result, reason).Inc()
	state.scoreReportsTotal.Add(1)
	state.mu.Lock()
	if !accepted {
		state.scoreReportReasons[reason]++
	}
	state.mu.Unlock()
}

func RecordMatchFinished(playerCount int) {
	matchesFinishedTotal.Inc()
	matchPlayersHistogram.Observe(float64(playerCount))
	state.matchesFinishedTotal.Add(1)
}

func SetActiveConnections(count int) {
	activeConnectionsGauge.Set(float64(count))
	state.activeConnections.Store(int64(count))
}

func UpdateRooms(activeRooms, activePlayers int, roomsByStatus map[string]int) {
	activeRoomsGauge.Set(float64(activeRooms))
	activePlayersGauge.Set(float64(activePlayers))
	state.activeRooms.Store(int64(activeRooms))
	state.activePlayers.Store(int64(activePlayers))

	statuses := []string{"waiting", "countdown", "running", "finished"}
	for _, status := range statuses {
		roomsByStatusGauge.WithLabelValues(status).Set(float64(roomsByStatus[status]))
	}

	state.mu.Lock()
	state.roomsByStatus = make(map[string]int, len(roomsByStatus))
	for status, count := range roomsByStatus {
		state.roomsByStatus[status] = count
	}
	state.mu.Unlock()
}

func GetSnapshot() Snapshot {
	state.mu.Lock()
	defer state.mu.Unlock()

	roomEventsByType := make(map[string]uint64, len(state.roomEventsByType))
	for key, value := range state.roomEventsByType {
		roomEventsByType[key] = value
	}
	scoreReportReasons := make(map[string]uint64, len(state.scoreReportReasons))
	for key, value := range state.scoreReportReasons {
		scoreReportReasons[key] = value
	}
	roomsByStatus := make(map[string]int, len(state.roomsByStatus))
	for key, value := range state.roomsByStatus {
		roomsByStatus[key] = value
	}

	return Snapshot{
		RoomEventsTotal:      state.roomEventsTotal.Load(),
		ScoreReportsTotal:    state.scoreReportsTotal.Load(),
		RejectedScoreReports: state.rejectedScoreReports.Load(),
		ActiveRooms:          int(state.activeRooms.Load()),
		ActivePlayers:        int(state.activePlayers.Load()),
		ActiveConnections:    int(state.activeConnections.Load()),
		MatchesFinishedTotal: state.matchesFinishedTotal.Load(),
		RoomsByStatus:        roomsByStatus,
		RoomEventsByType:     roomEventsByType,
		ScoreReportReasons:   scoreReportReasons,
	}
}
