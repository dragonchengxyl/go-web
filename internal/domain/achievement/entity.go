package achievement

import (
	"time"

	"github.com/google/uuid"
)

// Rarity represents achievement rarity level
type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

// Achievement represents an achievement definition
type Achievement struct {
	ID             int        `json:"id"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	IconKey        string     `json:"icon_key"`
	Rarity         Rarity     `json:"rarity"`
	Points         int        `json:"points"`
	ConditionType  string     `json:"condition_type"`
	ConditionValue []byte     `json:"condition_value,omitempty"` // raw JSONB
	IsSecret       bool       `json:"is_secret"`
	CreatedAt      time.Time  `json:"created_at"`
}

// UserAchievement represents an achievement unlocked by a user
type UserAchievement struct {
	ID            int64       `json:"id"`
	UserID        uuid.UUID   `json:"user_id"`
	AchievementID int         `json:"achievement_id"`
	Achievement   *Achievement `json:"achievement,omitempty"`
	ObtainedAt    time.Time   `json:"obtained_at"`
}

// PointTransaction records a point earn/spend event
type PointTransaction struct {
	ID        int64     `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Amount    int       `json:"amount"`
	Source    string    `json:"source"`
	RefID     string    `json:"ref_id,omitempty"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// UserPoints holds a user's point balance
type UserPoints struct {
	UserID      uuid.UUID `json:"user_id"`
	Balance     int       `json:"balance"`
	TotalEarned int       `json:"total_earned"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LeaderboardEntry is a single rank entry
type LeaderboardEntry struct {
	Rank     int64     `json:"rank"`
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Score    float64   `json:"score"`
}

// PointSource constants
const (
	SourceRegister     = "register"
	SourceDailyCheckin = "daily_checkin"
	SourceComment      = "comment"
	SourceCommentLike  = "comment_like"
	SourcePurchase     = "purchase"
	SourceAchievement  = "achievement"
	SourceAdmin        = "admin"
)
