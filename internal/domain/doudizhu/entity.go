package doudizhu

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Suit string

const (
	SuitSpade   Suit = "spade"
	SuitHeart   Suit = "heart"
	SuitClub    Suit = "club"
	SuitDiamond Suit = "diamond"
	SuitJoker   Suit = "joker"
)

type Rank int

const (
	Rank3          Rank = 3
	Rank4          Rank = 4
	Rank5          Rank = 5
	Rank6          Rank = 6
	Rank7          Rank = 7
	Rank8          Rank = 8
	Rank9          Rank = 9
	Rank10         Rank = 10
	RankJ          Rank = 11
	RankQ          Rank = 12
	RankK          Rank = 13
	RankA          Rank = 14
	Rank2          Rank = 15
	RankJokerSmall Rank = 16
	RankJokerBig   Rank = 17
)

type Seat int

const (
	Seat0 Seat = iota
	Seat1
	Seat2
)

type MatchMode string

const (
	MatchModePVP    MatchMode = "pvp"
	MatchModeDemoAI MatchMode = "demo_ai"
)

type RoundPhase string

const (
	RoundPhaseWaiting    RoundPhase = "waiting"
	RoundPhaseBidding    RoundPhase = "bidding"
	RoundPhasePlaying    RoundPhase = "playing"
	RoundPhaseSettlement RoundPhase = "settlement"
	RoundPhaseRedeal     RoundPhase = "redeal"
)

type PlayerRole string

const (
	PlayerRoleFarmer   PlayerRole = "farmer"
	PlayerRoleLandlord PlayerRole = "landlord"
)

type ComboType string

const (
	ComboSingle             ComboType = "single"
	ComboPair               ComboType = "pair"
	ComboTriple             ComboType = "triple"
	ComboTripleWithSingle   ComboType = "triple_with_single"
	ComboTripleWithPair     ComboType = "triple_with_pair"
	ComboStraight           ComboType = "straight"
	ComboStraightPairs      ComboType = "straight_pairs"
	ComboAirplane           ComboType = "airplane"
	ComboAirplaneWithSingle ComboType = "airplane_with_single"
	ComboAirplaneWithPair   ComboType = "airplane_with_pair"
	ComboFourWithTwoSingle  ComboType = "four_with_two_single"
	ComboFourWithTwoPair    ComboType = "four_with_two_pair"
	ComboBomb               ComboType = "bomb"
	ComboRocket             ComboType = "rocket"
)

type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

type Combo struct {
	Type           ComboType `json:"type"`
	MainRank       Rank      `json:"main_rank"`
	SequenceLength int       `json:"sequence_length"`
	TotalCards     int       `json:"total_cards"`
}

type RoomPlayer struct {
	SessionID string     `json:"session_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Seat      Seat       `json:"seat"`
	Name      string     `json:"name"`
	Ready     bool       `json:"ready"`
	Connected bool       `json:"connected"`
	IsHost    bool       `json:"is_host"`
	IsBot     bool       `json:"is_bot"`
	BotLevel  string     `json:"bot_level,omitempty"`
	AutoPlay  bool       `json:"auto_play"`
	CardCount int        `json:"card_count"`
	Role      PlayerRole `json:"role"`
	JoinedAt  time.Time  `json:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Room struct {
	ID            uuid.UUID      `json:"id"`
	Code          string         `json:"code"`
	Title         string         `json:"title"`
	MatchMode     MatchMode      `json:"match_mode"`
	Status        RoundPhase     `json:"status"`
	HostSessionID string         `json:"host_session_id"`
	PlayerCount   int            `json:"player_count"`
	ReadyCount    int            `json:"ready_count"`
	CurrentBidder *Seat          `json:"current_bidder,omitempty"`
	HighestBid    int            `json:"highest_bid"`
	HighestBidder *Seat          `json:"highest_bidder,omitempty"`
	Landlord      *Seat          `json:"landlord,omitempty"`
	CurrentTurn   *Seat          `json:"current_turn,omitempty"`
	LastPlay      *Combo         `json:"last_play,omitempty"`
	LastPlaySeat  *Seat          `json:"last_play_seat,omitempty"`
	WinningSide   *PlayerRole    `json:"winning_side,omitempty"`
	TurnExpiresAt *time.Time     `json:"turn_expires_at,omitempty"`
	BottomCards   []Card         `json:"bottom_cards,omitempty"`
	Players       []RoomPlayer   `json:"players"`
	BidHistory    []BidRecord    `json:"bid_history,omitempty"`
	RecentActions []ActionRecord `json:"recent_actions,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type BidRecord struct {
	Seat  Seat      `json:"seat"`
	Score int       `json:"score"`
	At    time.Time `json:"at"`
}

type ActionRecord struct {
	Seat        Seat        `json:"seat"`
	ActionType  string      `json:"action_type"`
	Cards       []Card      `json:"cards,omitempty"`
	Combo       *Combo      `json:"combo,omitempty"`
	At          time.Time   `json:"at"`
	Message     string      `json:"message,omitempty"`
	ActorName   string      `json:"actor_name,omitempty"`
	WinningSide *PlayerRole `json:"winning_side,omitempty"`
}

type PrivateState struct {
	SessionID     string     `json:"session_id"`
	RoomID        uuid.UUID  `json:"room_id"`
	Status        RoundPhase `json:"status"`
	Hand          []Card     `json:"hand"`
	CanPass       bool       `json:"can_pass"`
	Role          PlayerRole `json:"role"`
	BottomCards   []Card     `json:"bottom_cards,omitempty"`
	LastPlay      *Combo     `json:"last_play,omitempty"`
	LastPlaySeat  *Seat      `json:"last_play_seat,omitempty"`
	TurnExpiresAt *time.Time `json:"turn_expires_at,omitempty"`
}

type ActionResult struct {
	ActionType  string      `json:"action_type"`
	Seat        Seat        `json:"seat"`
	ActorName   string      `json:"actor_name"`
	Combo       *Combo      `json:"combo,omitempty"`
	Cards       []Card      `json:"cards,omitempty"`
	HandCount   int         `json:"hand_count"`
	NextTurn    *Seat       `json:"next_turn,omitempty"`
	Status      RoundPhase  `json:"status"`
	WinningSide *PlayerRole `json:"winning_side,omitempty"`
	Message     string      `json:"message,omitempty"`
}

type Match struct {
	ID           uuid.UUID           `json:"id"`
	RoomID       uuid.UUID           `json:"room_id"`
	RoomCode     string              `json:"room_code"`
	RoomTitle    string              `json:"room_title"`
	MatchMode    MatchMode           `json:"match_mode"`
	StartedAt    time.Time           `json:"started_at"`
	FinishedAt   time.Time           `json:"finished_at"`
	LandlordSeat Seat                `json:"landlord_seat"`
	WinnerSide   PlayerRole          `json:"winner_side"`
	Multiplier   int                 `json:"multiplier"`
	BombCount    int                 `json:"bomb_count"`
	Spring       bool                `json:"spring"`
	AntiSpring   bool                `json:"anti_spring"`
	CreatedAt    time.Time           `json:"created_at"`
	Results      []MatchPlayerResult `json:"results,omitempty"`
}

type MatchPlayerResult struct {
	ID          uuid.UUID  `json:"id"`
	MatchID     uuid.UUID  `json:"match_id"`
	SessionID   string     `json:"session_id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	IsBot       bool       `json:"is_bot"`
	BotLevel    string     `json:"bot_level,omitempty"`
	Seat        Seat       `json:"seat"`
	PlayerName  string     `json:"player_name"`
	DisplayName string     `json:"display_name"`
	Role        PlayerRole `json:"role"`
	BidScore    int        `json:"bid_score"`
	CardsLeft   int        `json:"cards_left"`
	IsWinner    bool       `json:"is_winner"`
	ScoreDelta  int        `json:"score_delta"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ActionEvent struct {
	ID              uuid.UUID  `json:"id"`
	MatchID         uuid.UUID  `json:"match_id"`
	TurnNo          int        `json:"turn_no"`
	ActionIndex     int        `json:"action_index"`
	SessionID       string     `json:"session_id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	PlayerName      string     `json:"player_name"`
	DisplayName     string     `json:"display_name"`
	Seat            Seat       `json:"seat"`
	ActionType      string     `json:"action_type"`
	Cards           []Card     `json:"cards,omitempty"`
	Combo           *Combo     `json:"combo,omitempty"`
	MultiplierAfter int        `json:"multiplier_after"`
	OccurredAt      time.Time  `json:"occurred_at"`
}

type MatchSummary struct {
	MatchID      uuid.UUID           `json:"match_id"`
	RoomID       uuid.UUID           `json:"room_id"`
	RoomCode     string              `json:"room_code"`
	RoomTitle    string              `json:"room_title"`
	MatchMode    MatchMode           `json:"match_mode"`
	FinishedAt   time.Time           `json:"finished_at"`
	WinnerSide   PlayerRole          `json:"winner_side"`
	LandlordSeat Seat                `json:"landlord_seat"`
	Multiplier   int                 `json:"multiplier"`
	PlayerCount  int                 `json:"player_count"`
	TopResults   []MatchPlayerResult `json:"top_results"`
}

type MatchReplay struct {
	Match   *Match              `json:"match"`
	Results []MatchPlayerResult `json:"results"`
	Events  []ActionEvent       `json:"events"`
}

type LeaderboardEntry struct {
	Rank        int64      `json:"rank"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	PlayerName  string     `json:"player_name"`
	DisplayName string     `json:"display_name"`
	Matches     int        `json:"matches"`
	Wins        int        `json:"wins"`
	TotalScore  int        `json:"total_score"`
	LastPlayed  time.Time  `json:"last_played"`
}

var (
	ErrInvalidSeat          = errors.New("invalid doudizhu seat")
	ErrInvalidBidScore      = errors.New("invalid doudizhu bid score")
	ErrRoundNotBidding      = errors.New("doudizhu round is not in bidding phase")
	ErrNotCurrentBidder     = errors.New("player is not the current bidder")
	ErrInvalidCombo         = errors.New("invalid doudizhu combo")
	ErrCannotBeatTargetPlay = errors.New("combo does not beat target play")
)

func (s Seat) Valid() bool {
	return s >= Seat0 && s <= Seat2
}

func (s Seat) Next() Seat {
	return Seat((int(s) + 1) % 3)
}

func (r Rank) String() string {
	switch r {
	case RankJ:
		return "J"
	case RankQ:
		return "Q"
	case RankK:
		return "K"
	case RankA:
		return "A"
	case Rank2:
		return "2"
	case RankJokerSmall:
		return "SJ"
	case RankJokerBig:
		return "BJ"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

func (c Card) String() string {
	var prefix string
	switch c.Suit {
	case SuitSpade:
		prefix = "S"
	case SuitHeart:
		prefix = "H"
	case SuitClub:
		prefix = "C"
	case SuitDiamond:
		prefix = "D"
	default:
		prefix = "J"
	}
	return prefix + c.Rank.String()
}
