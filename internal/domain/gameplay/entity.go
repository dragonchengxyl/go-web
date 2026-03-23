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

type Match struct {
	ID          uuid.UUID     `json:"id"`
	RoomID      uuid.UUID     `json:"room_id"`
	RoomCode    string        `json:"room_code"`
	RoomTitle   string        `json:"room_title"`
	GameSlug    string        `json:"game_slug"`
	StartedAt   time.Time     `json:"started_at"`
	FinishedAt  time.Time     `json:"finished_at"`
	DurationSec int           `json:"duration_sec"`
	CreatedAt   time.Time     `json:"created_at"`
	Results     []MatchResult `json:"results,omitempty"`
}

type MatchResult struct {
	ID          uuid.UUID  `json:"id"`
	MatchID     uuid.UUID  `json:"match_id"`
	RoomID      uuid.UUID  `json:"room_id"`
	RoomCode    string     `json:"room_code"`
	RoomTitle   string     `json:"room_title"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	PlayerName  string     `json:"player_name"`
	DisplayName string     `json:"display_name"`
	Score       int        `json:"score"`
	Rank        int        `json:"rank"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  time.Time  `json:"finished_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type LeaderboardEntry struct {
	Rank        int64      `json:"rank"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	PlayerName  string     `json:"player_name"`
	DisplayName string     `json:"display_name"`
	BestScore   int        `json:"best_score"`
	Matches     int        `json:"matches"`
	LastPlayed  time.Time  `json:"last_played"`
}

type MatchSummary struct {
	MatchID     uuid.UUID     `json:"match_id"`
	RoomID      uuid.UUID     `json:"room_id"`
	RoomCode    string        `json:"room_code"`
	RoomTitle   string        `json:"room_title"`
	GameSlug    string        `json:"game_slug"`
	FinishedAt  time.Time     `json:"finished_at"`
	DurationSec int           `json:"duration_sec"`
	WinnerName  string        `json:"winner_name"`
	WinnerScore int           `json:"winner_score"`
	PlayerCount int           `json:"player_count"`
	TopResults  []MatchResult `json:"top_results"`
}
