package gameplay

import (
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomStatusWaiting   RoomStatus = "waiting"
	RoomStatusCountdown RoomStatus = "countdown"
	RoomStatusRunning   RoomStatus = "running"
	RoomStatusFinished  RoomStatus = "finished"
)

type RoomPlayer struct {
	SessionID string     `json:"session_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Name      string     `json:"name"`
	Ready     bool       `json:"ready"`
	Connected bool       `json:"connected"`
	IsHost    bool       `json:"is_host"`
	Score     int        `json:"score"`
	JoinedAt  time.Time  `json:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Room struct {
	ID               uuid.UUID    `json:"id"`
	Code             string       `json:"code"`
	GameSlug         string       `json:"game_slug"`
	Title            string       `json:"title"`
	Status           RoomStatus   `json:"status"`
	HostSessionID    string       `json:"host_session_id"`
	CountdownSec     int          `json:"countdown_sec"`
	RoundDurationSec int          `json:"round_duration_sec"`
	PlayerCount      int          `json:"player_count"`
	ReadyCount       int          `json:"ready_count"`
	CountdownStarted *time.Time   `json:"countdown_started_at,omitempty"`
	StartedAt        *time.Time   `json:"started_at,omitempty"`
	EndsAt           *time.Time   `json:"ends_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	Players          []RoomPlayer `json:"players"`
}
