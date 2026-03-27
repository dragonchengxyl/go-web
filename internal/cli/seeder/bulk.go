package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/domain/report"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	"github.com/studio/platform/internal/pkg/crypto"
)

type BulkSeedOptions struct {
	Profile   string
	Namespace string
	Seed      int64
}

type bulkPlan struct {
	Users                    int
	Follows                  int
	Groups                   int
	GroupAnnouncements       int
	Events                   int
	Posts                    int
	Comments                 int
	PostLikes                int
	CommentLikes             int
	Bookmarks                int
	Conversations            int
	Messages                 int
	Notifications            int
	Reports                  int
	Orders                   int
	Albums                   int
	TracksPerAlbum           int
	AudioJobs                int
	AudioWorks               int
	AudioWorkLikes           int
	AssistantConversations   int
	AssistantMessages        int
	AssistantFeedback        int
	AssistantKnowledgeDocs   int
	AuditLogs                int
	AnalyticsEvents          int
	HexBlitzMatches          int
	HexBlitzMoveEventsPerHit int
	DoudizhuMatches          int
	DoudizhuActionsPerMatch  int
	UserAchievements         int
	PointTransactions        int
}

type bulkUser struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      string
	Status    string
	Bio       string
	Location  string
	Website   string
	FurryName string
	Species   string
	CreatedAt time.Time
}

type bulkGroup struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
	Announcement string
	Rules       string
	Tags        []string
	Privacy     string
	CreatedAt   time.Time
}

type bulkEvent struct {
	ID          uuid.UUID
	OrganizerID uuid.UUID
	Title       string
	Description string
	Location    string
	IsOnline    bool
	StartTime   time.Time
	EndTime     time.Time
	MaxCapacity int
	Tags        []string
	Status      string
	CreatedAt   time.Time
}

type bulkPost struct {
	ID               uuid.UUID
	AuthorID         uuid.UUID
	GroupID          *uuid.UUID
	Title            string
	Content          string
	Tags             []string
	Visibility       string
	ModerationStatus string
	ContentLabels    map[string]bool
	CreatedAt        time.Time
}

type bulkComment struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	PostID      uuid.UUID
	ParentID    *uuid.UUID
	Content     string
	CreatedAt   time.Time
}

type bulkConversation struct {
	ID        uuid.UUID
	Type      string
	Name      *string
	Members   []uuid.UUID
	CreatedAt time.Time
}

type bulkMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
	CreatedAt      time.Time
}

type bulkNotification struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ActorID    *uuid.UUID
	Type       string
	TargetID   *uuid.UUID
	TargetType *string
	IsRead     bool
	CreatedAt  time.Time
}

type bulkReport struct {
	ID          uuid.UUID
	ReporterID  uuid.UUID
	TargetType  report.TargetType
	TargetID    uuid.UUID
	Reason      string
	Description string
	Status      report.Status
	ReviewedBy  *uuid.UUID
	ReviewedAt  *time.Time
	ActionTaken *report.Action
	CreatedAt   time.Time
}

type bulkOrder struct {
	ID             uuid.UUID
	OrderNo        string
	UserID         uuid.UUID
	Status         order.OrderStatus
	TotalCents     int
	Currency       string
	PaymentMethod  order.PaymentMethod
	PaidAt         *time.Time
	IdempotencyKey string
	Metadata       map[string]any
	CreatedAt      time.Time
	ExpiresAt      *time.Time
}

type bulkAlbum struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Subtitle    string
	Description string
	Artist      string
	Composer    string
	ReleaseDate time.Time
	AlbumType   string
	CreatedAt   time.Time
}

type bulkTrack struct {
	ID          uuid.UUID
	AlbumID     uuid.UUID
	TrackNumber int
	DiscNumber  int
	Title       string
	Artist      string
	DurationSec int
	PlayCount   int64
	CreatedAt   time.Time
}

type bulkAudioJob struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Title             string
	TaskType          string
	Status            string
	SourceAudioURL    *string
	ReferenceAudioURL *string
	Prompt            *string
	Params            map[string]any
	Result            map[string]any
	ErrorMessage      *string
	AttemptCount      int
	MaxAttempts       int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	StartedAt         *time.Time
	FinishedAt        *time.Time
	LastErrorAt       *time.Time
}

type bulkAudioWork struct {
	ID               uuid.UUID
	AuthorID         uuid.UUID
	SourceJobID      uuid.UUID
	Title            string
	Description      string
	CoverImageURL    *string
	AudioURL         string
	DurationSec      float64
	Visibility       string
	ModerationStatus string
	ModerationNote   *string
	Tags             []string
	Metadata         map[string]any
	PublishedAt      time.Time
}

type bulkAssistantConversation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	CreatedAt time.Time
}

type bulkAssistantMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	Role           string
	Content        string
	Cards          []map[string]any
	Insights       []map[string]any
	CreatedAt      time.Time
}

type bulkAssistantFeedback struct {
	ID             uuid.UUID
	ResponseID     uuid.UUID
	ConversationID *uuid.UUID
	UserID         *uuid.UUID
	Value          string
	QueryText      string
	ReplyExcerpt   string
	Provider       string
	Intent         string
	Fallback       bool
	PagePath       string
	SourceCounts   map[string]int
	Cards          []map[string]any
	CreatedAt      time.Time
}

type bulkKnowledgeDocument struct {
	ID              uuid.UUID
	SourceType      string
	SourceKey       string
	ChunkIndex      int
	Title           string
	Summary         string
	Content         string
	Href            string
	Meta            string
	SourceLabel     string
	Tags            []string
	SearchText      string
	Embedding       []float64
	IndexedAt       time.Time
	SourceUpdatedAt time.Time
}

type bulkAuditLog struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	Username     string
	Action       audit.Action
	Resource     audit.Resource
	ResourceID   *uuid.UUID
	IPAddress    string
	UserAgent    string
	BeforeData   *string
	AfterData    *string
	ErrorMessage *string
	CreatedAt    time.Time
}

type bulkAnalyticsEvent struct {
	ID         uuid.UUID
	EventType  string
	UserID     *uuid.UUID
	SessionID  string
	Properties map[string]any
	IPAddress  string
	UserAgent  string
	Referrer   *string
	CreatedAt  time.Time
}

type bulkHexMatch struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	RoomCode   string
	RoomTitle  string
	Seed       int64
	StartedAt  time.Time
	FinishedAt time.Time
}

type bulkHexResult struct {
	ID         uuid.UUID
	MatchID    uuid.UUID
	RoomID     uuid.UUID
	RoomCode   string
	RoomTitle  string
	UserID     *uuid.UUID
	PlayerName string
	SessionID  string
	Score      int
	Rank       int
	CreatedAt  time.Time
}

type bulkHexMoveEvent struct {
	ID           uuid.UUID
	MatchID      uuid.UUID
	SessionID    string
	UserID       *uuid.UUID
	PlayerName   string
	TileID       string
	MoveIndex    int
	ClearedCount int
	GainedScore  int
	ScoreAfter   int
	ComboAfter   int
	OccurredAt   time.Time
}

type bulkDoudizhuMatch struct {
	ID           uuid.UUID
	RoomID       uuid.UUID
	RoomCode     string
	RoomTitle    string
	MatchMode    string
	StartedAt    time.Time
	FinishedAt   time.Time
	LandlordSeat int16
	WinnerSide   string
	Multiplier   int
	BombCount    int
	Spring       bool
	AntiSpring   bool
}

type bulkDoudizhuPlayer struct {
	ID         uuid.UUID
	MatchID    uuid.UUID
	SessionID  string
	UserID     *uuid.UUID
	IsBot      bool
	BotLevel   *string
	Seat       int16
	PlayerName string
	Role       string
	BidScore   int
	CardsLeft  int
	IsWinner   bool
	ScoreDelta int
	CreatedAt  time.Time
}

type bulkDoudizhuAction struct {
	ID                  uuid.UUID
	MatchID             uuid.UUID
	TurnNo              int
	ActionIndex         int
	SessionID           string
	UserID              *uuid.UUID
	PlayerName          string
	Seat                int16
	ActionType          string
	CardsJSON           any
	ComboType           *string
	ComboMainRank       *int
	ComboSequenceLength *int
	ComboTotalCards     *int
	MultiplierAfter     int
	OccurredAt          time.Time
}

type bulkBatcher struct {
	ctx   context.Context
	tx    pgx.Tx
	batch pgx.Batch
	count int
}

func (b *bulkBatcher) Queue(query string, args ...any) error {
	b.batch.Queue(query, args...)
	b.count++
	if b.count >= 500 {
		return b.Flush()
	}
	return nil
}

func (b *bulkBatcher) Flush() error {
	if b.count == 0 {
		return nil
	}
	res := b.tx.SendBatch(b.ctx, &b.batch)
	err := res.Close()
	b.batch = pgx.Batch{}
	b.count = 0
	return err
}

func SeedBulk(ctx context.Context, cfg *configs.Config, opts BulkSeedOptions, out io.Writer) error {
	profile := strings.ToLower(strings.TrimSpace(opts.Profile))
	if profile == "" {
		profile = "medium"
	}
	namespace := sanitizeNamespace(opts.Namespace)
	if namespace == "" {
		namespace = "bulk"
	}

	plan, ok := bulkProfiles()[profile]
	if !ok {
		return fmt.Errorf("unknown bulk seed profile %q", profile)
	}

	seed := opts.Seed
	if seed == 0 {
		seed = namespaceSeed(namespace)
	}

	pool, err := postgresinfra.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	beforePretty, _ := databaseSizePretty(ctx, pool)
	rng := rand.New(rand.NewSource(seed))
	now := time.Now().UTC()
	passwordHash, err := crypto.HashPassword(DemoPassword)
	if err != nil {
		return fmt.Errorf("hash bulk seed password: %w", err)
	}

	fmt.Fprintf(out, "bulk seed start: profile=%s namespace=%s seed=%d db_size_before=%s\n", profile, namespace, seed, beforePretty)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	b := &bulkBatcher{ctx: ctx, tx: tx}

	users := buildBulkUsers(namespace, plan, rng, now)
	if err := insertBulkUsers(b, users, passwordHash); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	groupData, groupMembers, groupAnnouncements := buildBulkGroups(namespace, plan, users, rng, now)
	if err := insertBulkGroups(b, groupData); err != nil {
		return fmt.Errorf("seed groups: %w", err)
	}
	if err := insertBulkGroupMembers(b, groupMembers); err != nil {
		return fmt.Errorf("seed group members: %w", err)
	}
	if err := insertBulkGroupAnnouncements(b, groupAnnouncements); err != nil {
		return fmt.Errorf("seed group announcements: %w", err)
	}

	events, eventAttendees := buildBulkEvents(namespace, plan, users, rng, now)
	if err := insertBulkEvents(b, events); err != nil {
		return fmt.Errorf("seed events: %w", err)
	}
	if err := insertBulkEventAttendees(b, eventAttendees); err != nil {
		return fmt.Errorf("seed event attendees: %w", err)
	}

	posts, featuredGroupPosts := buildBulkPosts(namespace, plan, users, groupData, rng, now)
	if err := insertBulkPosts(b, posts); err != nil {
		return fmt.Errorf("seed posts: %w", err)
	}
	if err := assignBulkFeaturedPosts(b, featuredGroupPosts); err != nil {
		return fmt.Errorf("seed featured posts: %w", err)
	}

	postLikes := buildBulkPostLikes(namespace, plan, users, posts, rng, now)
	if err := insertBulkPostLikes(b, postLikes); err != nil {
		return fmt.Errorf("seed post likes: %w", err)
	}

	comments := buildBulkComments(namespace, plan, users, posts, rng, now)
	if err := insertBulkComments(b, comments); err != nil {
		return fmt.Errorf("seed comments: %w", err)
	}

	commentLikes := buildBulkCommentLikes(namespace, plan, users, comments, rng, now)
	if err := insertBulkCommentLikes(b, commentLikes); err != nil {
		return fmt.Errorf("seed comment likes: %w", err)
	}

	conversations, messages := buildBulkChat(namespace, plan, users, rng, now)
	if err := insertBulkConversations(b, conversations); err != nil {
		return fmt.Errorf("seed conversations: %w", err)
	}
	if err := insertBulkConversationMembers(b, conversations); err != nil {
		return fmt.Errorf("seed conversation members: %w", err)
	}
	if err := insertBulkMessages(b, messages); err != nil {
		return fmt.Errorf("seed messages: %w", err)
	}

	audioJobs := buildBulkAudioJobs(namespace, plan, users, rng, now)
	if err := insertBulkAudioJobs(b, audioJobs); err != nil {
		return fmt.Errorf("seed audio jobs: %w", err)
	}

	audioWorks := buildBulkAudioWorks(namespace, plan, users, audioJobs, rng, now)
	if err := insertBulkAudioWorks(b, audioWorks); err != nil {
		return fmt.Errorf("seed audio works: %w", err)
	}

	audioWorkLikes := buildBulkAudioWorkLikes(namespace, plan, users, audioWorks, rng, now)
	if err := insertBulkAudioWorkLikes(b, audioWorkLikes); err != nil {
		return fmt.Errorf("seed audio work likes: %w", err)
	}

	bookmarks := buildBulkBookmarks(namespace, plan, users, posts, groupData, events, audioWorks, rng, now)
	if err := insertBulkBookmarks(b, bookmarks); err != nil {
		return fmt.Errorf("seed bookmarks: %w", err)
	}

	follows := buildBulkFollows(namespace, plan, users, rng, now)
	if err := insertBulkFollows(b, follows); err != nil {
		return fmt.Errorf("seed follows: %w", err)
	}

	notifications := buildBulkNotifications(namespace, plan, users, posts, comments, audioWorks, rng, now)
	if err := insertBulkNotifications(b, notifications); err != nil {
		return fmt.Errorf("seed notifications: %w", err)
	}

	reports := buildBulkReports(namespace, plan, users, posts, comments, audioWorks, rng, now)
	if err := insertBulkReports(b, reports); err != nil {
		return fmt.Errorf("seed reports: %w", err)
	}

	orders := buildBulkOrders(namespace, plan, users, rng, now)
	if err := insertBulkOrders(b, orders); err != nil {
		return fmt.Errorf("seed orders: %w", err)
	}

	albums, tracks := buildBulkMusic(namespace, plan, rng, now)
	if err := insertBulkAlbums(b, albums); err != nil {
		return fmt.Errorf("seed albums: %w", err)
	}
	if err := insertBulkTracks(b, tracks); err != nil {
		return fmt.Errorf("seed tracks: %w", err)
	}

	assistantConversations, assistantMessages := buildBulkAssistantConversations(namespace, plan, users, rng, now)
	if err := insertBulkAssistantConversations(b, assistantConversations); err != nil {
		return fmt.Errorf("seed assistant conversations: %w", err)
	}
	if err := insertBulkAssistantMessages(b, assistantMessages); err != nil {
		return fmt.Errorf("seed assistant messages: %w", err)
	}

	assistantFeedback := buildBulkAssistantFeedback(namespace, plan, users, assistantConversations, rng, now)
	if err := insertBulkAssistantFeedback(b, assistantFeedback); err != nil {
		return fmt.Errorf("seed assistant feedback: %w", err)
	}

	knowledgeDocs := buildBulkKnowledgeDocs(namespace, plan, posts, groupData, events, rng, now)
	if err := insertBulkKnowledgeDocs(b, knowledgeDocs); err != nil {
		return fmt.Errorf("seed assistant knowledge documents: %w", err)
	}

	auditLogs := buildBulkAuditLogs(namespace, plan, users, posts, orders, reports, rng, now)
	if err := insertBulkAuditLogs(b, auditLogs); err != nil {
		return fmt.Errorf("seed audit logs: %w", err)
	}

	analyticsEvents := buildBulkAnalyticsEvents(namespace, plan, users, rng, now)
	if err := insertBulkAnalyticsEvents(b, analyticsEvents); err != nil {
		return fmt.Errorf("seed analytics events: %w", err)
	}

	hexMatches, hexResults, hexMoves := buildBulkHexBlitz(namespace, plan, users, rng, now)
	if err := insertBulkHexBlitzMatches(b, hexMatches, hexResults, hexMoves); err != nil {
		return fmt.Errorf("seed hex blitz matches: %w", err)
	}

	doudizhuMatches, doudizhuPlayers, doudizhuActions := buildBulkDoudizhu(namespace, plan, users, rng, now)
	if err := insertBulkDoudizhuMatches(b, doudizhuMatches, doudizhuPlayers, doudizhuActions); err != nil {
		return fmt.Errorf("seed doudizhu matches: %w", err)
	}

	pointTransactions := buildBulkPointTransactions(namespace, plan, users, rng, now)
	if err := resetNamespacePointTransactions(ctx, tx, namespace); err != nil {
		return fmt.Errorf("reset point transactions namespace: %w", err)
	}
	if err := insertBulkPointTransactions(b, pointTransactions); err != nil {
		return fmt.Errorf("seed point transactions: %w", err)
	}

	userAchievements := buildBulkUserAchievements(namespace, plan, users, rng, now)
	if err := insertBulkUserAchievements(b, userAchievements); err != nil {
		return fmt.Errorf("seed user achievements: %w", err)
	}

	if err := upsertBulkSponsorSettings(b, users[0].ID, namespace, now); err != nil {
		return fmt.Errorf("seed sponsor settings: %w", err)
	}
	if err := upsertBulkAssistantSettings(b, users[0].ID, now); err != nil {
		return fmt.Errorf("seed assistant settings: %w", err)
	}

	if err := b.Flush(); err != nil {
		return fmt.Errorf("flush batched statements: %w", err)
	}

	if err := reconcileBulkCounters(ctx, tx, users); err != nil {
		return fmt.Errorf("reconcile counters: %w", err)
	}
	if err := refreshDailyMetrics(ctx, tx); err != nil {
		return fmt.Errorf("refresh daily metrics: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit bulk seed transaction: %w", err)
	}

	afterPretty, _ := databaseSizePretty(ctx, pool)
	fmt.Fprintf(out, "bulk seeded successfully: users=%d posts=%d comments=%d messages=%d orders=%d audio_works=%d reports=%d db_size_after=%s\n",
		len(users),
		len(posts),
		len(comments),
		len(messages),
		len(orders),
		len(audioWorks),
		len(reports),
		afterPretty,
	)
	fmt.Fprintf(out, "shared password for seeded users: %s\n", DemoPassword)
	fmt.Fprintf(out, "namespace tag: [bulk:%s]\n", namespace)
	return nil
}

func bulkProfiles() map[string]bulkPlan {
	return map[string]bulkPlan{
		"small": {
			Users:                    120,
			Follows:                  360,
			Groups:                   18,
			GroupAnnouncements:       18,
			Events:                   36,
			Posts:                    260,
			Comments:                 760,
			PostLikes:                900,
			CommentLikes:             320,
			Bookmarks:                240,
			Conversations:            42,
			Messages:                 480,
			Notifications:            520,
			Reports:                  120,
			Orders:                   120,
			Albums:                   8,
			TracksPerAlbum:           8,
			AudioJobs:                60,
			AudioWorks:               36,
			AudioWorkLikes:           120,
			AssistantConversations:   28,
			AssistantMessages:        240,
			AssistantFeedback:        60,
			AssistantKnowledgeDocs:   140,
			AuditLogs:                220,
			AnalyticsEvents:          500,
			HexBlitzMatches:          24,
			HexBlitzMoveEventsPerHit: 18,
			DoudizhuMatches:          24,
			DoudizhuActionsPerMatch:  28,
			UserAchievements:         180,
			PointTransactions:        220,
		},
		"medium": {
			Users:                    800,
			Follows:                  3200,
			Groups:                   90,
			GroupAnnouncements:       90,
			Events:                   220,
			Posts:                    2600,
			Comments:                 8400,
			PostLikes:                9500,
			CommentLikes:             4200,
			Bookmarks:                3200,
			Conversations:            320,
			Messages:                 5400,
			Notifications:            4200,
			Reports:                  900,
			Orders:                   1000,
			Albums:                   24,
			TracksPerAlbum:           12,
			AudioJobs:                320,
			AudioWorks:               220,
			AudioWorkLikes:           900,
			AssistantConversations:   180,
			AssistantMessages:        2160,
			AssistantFeedback:        420,
			AssistantKnowledgeDocs:   1200,
			AuditLogs:                2400,
			AnalyticsEvents:          3000,
			HexBlitzMatches:          120,
			HexBlitzMoveEventsPerHit: 22,
			DoudizhuMatches:          120,
			DoudizhuActionsPerMatch:  34,
			UserAchievements:         2800,
			PointTransactions:        3200,
		},
		"large": {
			Users:                    3000,
			Follows:                  16000,
			Groups:                   320,
			GroupAnnouncements:       320,
			Events:                   700,
			Posts:                    12000,
			Comments:                 42000,
			PostLikes:                52000,
			CommentLikes:             18000,
			Bookmarks:                18000,
			Conversations:            1400,
			Messages:                 26000,
			Notifications:            22000,
			Reports:                  4200,
			Orders:                   4000,
			Albums:                   60,
			TracksPerAlbum:           14,
			AudioJobs:                1200,
			AudioWorks:               900,
			AudioWorkLikes:           4200,
			AssistantConversations:   720,
			AssistantMessages:        9600,
			AssistantFeedback:        1800,
			AssistantKnowledgeDocs:   4200,
			AuditLogs:                9600,
			AnalyticsEvents:          12000,
			HexBlitzMatches:          360,
			HexBlitzMoveEventsPerHit: 26,
			DoudizhuMatches:          360,
			DoudizhuActionsPerMatch:  42,
			UserAchievements:         12000,
			PointTransactions:        14000,
		},
	}
}

func sanitizeNamespace(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		default:
			return -1
		}
	}, value)
	return strings.Trim(value, "-")
}

func namespaceSeed(namespace string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(namespace))
	return int64(h.Sum64())
}

func bulkUUID(namespace, key string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("bulk:"+namespace+":"+key))
}

func databaseSizePretty(ctx context.Context, tx interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) (string, error) {
	var pretty string
	err := tx.QueryRow(ctx, `SELECT pg_size_pretty(pg_database_size(current_database()))`).Scan(&pretty)
	return pretty, err
}

func buildBulkUsers(namespace string, plan bulkPlan, rng *rand.Rand, now time.Time) []bulkUser {
	users := make([]bulkUser, 0, plan.Users)
	roleCycle := []string{"member", "member", "member", "creator", "member", "supporter", "moderator", "member"}
	statusCycle := []string{"active", "active", "active", "active", "active", "inactive", "suspended", "banned"}

	for i := 0; i < plan.Users; i++ {
		role := roleCycle[i%len(roleCycle)]
		switch i {
		case 0:
			role = "super_admin"
		case 1, 2:
			role = "admin"
		case 3, 4, 5:
			role = "moderator"
		}
		status := statusCycle[i%len(statusCycle)]
		if role == "super_admin" || role == "admin" || role == "moderator" {
			status = "active"
		}
		username := bulkUsername(namespace, i)
		users = append(users, bulkUser{
			ID:        bulkUUID(namespace, fmt.Sprintf("user-%d", i)),
			Username:  username,
			Email:     fmt.Sprintf("%s@seed.local", username),
			Role:      role,
			Status:    status,
			Bio:       sentence(rng, 14),
			Location:  randomCity(rng),
			Website:   fmt.Sprintf("https://seed.local/%s", username),
			FurryName: randomFurryName(rng),
			Species:   randomSpecies(rng),
			CreatedAt: randomPastTime(rng, now, 180),
		})
	}
	return users
}

func buildBulkGroups(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkGroup, [][3]any, []struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
}) {
	groups := make([]bulkGroup, 0, plan.Groups)
	members := make([][3]any, 0, plan.Groups*8)
	announcements := make([]struct {
		ID        uuid.UUID
		GroupID   uuid.UUID
		AuthorID  uuid.UUID
		Content   string
		CreatedAt time.Time
	}, 0, plan.GroupAnnouncements)

	for i := 0; i < plan.Groups; i++ {
		owner := users[i%len(users)]
		createdAt := randomPastTime(rng, now, 120)
		privacy := "public"
		if i%4 == 0 {
			privacy = "private"
		}
		group := bulkGroup{
			ID:           bulkUUID(namespace, fmt.Sprintf("group-%d", i)),
			OwnerID:      owner.ID,
			Name:         fmt.Sprintf("%s交流局 %03d", randomGroupTopic(rng), i+1),
			Description:  sentence(rng, 22),
			Announcement: sentence(rng, 18),
			Rules:        "请保持友善交流，禁止刷屏和人身攻击，涉及 NSFW 请明确标注。",
			Tags:         randomTags(rng, 3),
			Privacy:      privacy,
			CreatedAt:    createdAt,
		}
		groups = append(groups, group)
		members = append(members, [3]any{group.ID, owner.ID, "owner"})

		memberCount := 4 + rng.Intn(10)
		for j := 0; j < memberCount; j++ {
			user := users[(i*11+j*7+3)%len(users)]
			if user.ID == owner.ID {
				continue
			}
			role := "member"
			if j == 0 && i%5 == 0 {
				role = "moderator"
			}
			members = append(members, [3]any{group.ID, user.ID, role})
		}

		if i < plan.GroupAnnouncements {
			announcements = append(announcements, struct {
				ID        uuid.UUID
				GroupID   uuid.UUID
				AuthorID  uuid.UUID
				Content   string
				CreatedAt time.Time
			}{
				ID:        bulkUUID(namespace, fmt.Sprintf("group-announcement-%d", i)),
				GroupID:   group.ID,
				AuthorID:  owner.ID,
				Content:   "欢迎加入本圈子，近期会集中整理精选内容、公告和活动信息。",
				CreatedAt: createdAt.Add(2 * time.Hour),
			})
		}
	}
	return groups, members, announcements
}

func buildBulkEvents(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkEvent, [][3]any) {
	events := make([]bulkEvent, 0, plan.Events)
	attendees := make([][3]any, 0, plan.Events*12)
	statuses := []string{"published", "published", "completed", "draft", "cancelled"}

	for i := 0; i < plan.Events; i++ {
		status := statuses[i%len(statuses)]
		start := randomEventStart(rng, now, status)
		end := start.Add(time.Duration(2+rng.Intn(5)) * time.Hour)
		organizer := users[(i*3+7)%len(users)]
		event := bulkEvent{
			ID:          bulkUUID(namespace, fmt.Sprintf("event-%d", i)),
			OrganizerID: organizer.ID,
			Title:       fmt.Sprintf("%s %03d", randomEventTopic(rng), i+1),
			Description: sentence(rng, 24),
			Location:    randomLocation(rng),
			IsOnline:    i%3 == 0,
			StartTime:   start,
			EndTime:     end,
			MaxCapacity: 20 + rng.Intn(80),
			Tags:        randomTags(rng, 3),
			Status:      status,
			CreatedAt:   start.Add(-24 * time.Hour),
		}
		events = append(events, event)

		attendeeCount := 5 + rng.Intn(18)
		for j := 0; j < attendeeCount; j++ {
			user := users[(i*17+j*9+5)%len(users)]
			status := "attending"
			if j%7 == 0 {
				status = "maybe"
			}
			attendees = append(attendees, [3]any{event.ID, user.ID, status})
		}
	}

	return events, attendees
}

func buildBulkPosts(namespace string, plan bulkPlan, users []bulkUser, groups []bulkGroup, rng *rand.Rand, now time.Time) ([]bulkPost, map[uuid.UUID]uuid.UUID) {
	posts := make([]bulkPost, 0, plan.Posts)
	featuredByGroup := make(map[uuid.UUID]uuid.UUID)
	visibilities := []string{"public", "public", "public", "followers_only", "private"}
	moderationStatuses := []string{"approved", "approved", "approved", "approved", "pending", "blocked"}

	for i := 0; i < plan.Posts; i++ {
		author := users[(i*5+11)%len(users)]
		createdAt := randomPastTime(rng, now, 150)
		var groupID *uuid.UUID
		if len(groups) > 0 && i%3 == 0 {
			group := groups[i%len(groups)]
			groupID = &group.ID
			if _, ok := featuredByGroup[group.ID]; !ok {
				postID := bulkUUID(namespace, fmt.Sprintf("post-%d", i))
				featuredByGroup[group.ID] = postID
			}
		}
		postID := bulkUUID(namespace, fmt.Sprintf("post-%d", i))
		posts = append(posts, bulkPost{
			ID:               postID,
			AuthorID:         author.ID,
			GroupID:          groupID,
			Title:            fmt.Sprintf("%s %03d", randomPostTitle(rng), i+1),
			Content:          paragraph(rng, 3),
			Tags:             randomTags(rng, 4),
			Visibility:       visibilities[i%len(visibilities)],
			ModerationStatus: moderationStatuses[i%len(moderationStatuses)],
			ContentLabels: map[string]bool{
				"is_ai_generated": i%7 == 0,
			},
			CreatedAt: createdAt,
		})
	}
	return posts, featuredByGroup
}

func buildBulkComments(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, rng *rand.Rand, now time.Time) []bulkComment {
	comments := make([]bulkComment, 0, plan.Comments)
	topLevelIDs := make([]uuid.UUID, 0, plan.Comments/3)

	for i := 0; i < plan.Comments; i++ {
		post := posts[(i*7+13)%len(posts)]
		user := users[(i*9+5)%len(users)]
		createdAt := randomPastTime(rng, now, 120)
		var parentID *uuid.UUID
		if len(topLevelIDs) > 0 && i%4 == 0 {
			ref := topLevelIDs[(i*3)%len(topLevelIDs)]
			parentID = &ref
		}
		commentID := bulkUUID(namespace, fmt.Sprintf("comment-%d", i))
		comments = append(comments, bulkComment{
			ID:        commentID,
			UserID:    user.ID,
			PostID:    post.ID,
			ParentID:  parentID,
			Content:   sentence(rng, 18),
			CreatedAt: createdAt,
		})
		if parentID == nil {
			topLevelIDs = append(topLevelIDs, commentID)
		}
	}
	return comments
}

func buildBulkChat(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkConversation, []bulkMessage) {
	conversations := make([]bulkConversation, 0, plan.Conversations)
	messages := make([]bulkMessage, 0, plan.Messages)

	for i := 0; i < plan.Conversations; i++ {
		createdAt := randomPastTime(rng, now, 90)
		conv := bulkConversation{
			ID:        bulkUUID(namespace, fmt.Sprintf("conversation-%d", i)),
			Type:      "direct",
			CreatedAt: createdAt,
		}
		if i%5 == 0 {
			name := fmt.Sprintf("临时讨论组 %03d", i+1)
			conv.Type = "group"
			conv.Name = &name
			memberCount := 3 + rng.Intn(3)
			for j := 0; j < memberCount; j++ {
				conv.Members = append(conv.Members, users[(i*7+j*5+1)%len(users)].ID)
			}
		} else {
			conv.Members = []uuid.UUID{
				users[(i*5+1)%len(users)].ID,
				users[(i*5+3)%len(users)].ID,
			}
		}
		conversations = append(conversations, conv)
	}

	for i := 0; i < plan.Messages; i++ {
		conv := conversations[i%len(conversations)]
		sender := conv.Members[i%len(conv.Members)]
		messages = append(messages, bulkMessage{
			ID:             bulkUUID(namespace, fmt.Sprintf("message-%d", i)),
			ConversationID: conv.ID,
			SenderID:       sender,
			Content:        sentence(rng, 12),
			CreatedAt:      randomPastTime(rng, now, 90),
		})
	}

	return conversations, messages
}

func buildBulkAudioJobs(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) []bulkAudioJob {
	jobs := make([]bulkAudioJob, 0, plan.AudioJobs)
	taskTypes := []string{"ai_music", "voice_convert", "voice_enhance", "audio_master"}
	statuses := []string{"queued", "running", "succeeded", "failed", "dead_lettered"}

	for i := 0; i < plan.AudioJobs; i++ {
		user := users[(i*7+3)%len(users)]
		status := statuses[i%len(statuses)]
		taskType := taskTypes[i%len(taskTypes)]
		createdAt := randomPastTime(rng, now, 90)
		var startedAt *time.Time
		var finishedAt *time.Time
		var errorMessage *string
		var lastErrorAt *time.Time
		if status != "queued" {
			start := createdAt.Add(time.Duration(rng.Intn(90)) * time.Minute)
			startedAt = &start
		}
		if status == "succeeded" || status == "failed" || status == "dead_lettered" {
			base := createdAt
			if startedAt != nil {
				base = *startedAt
			}
			finish := base.Add(time.Duration(10+rng.Intn(240)) * time.Minute)
			finishedAt = &finish
		}
		if status == "failed" || status == "dead_lettered" {
			msg := "模拟音频处理失败，请检查上游素材或参数。"
			errorMessage = &msg
			lastErrorAt = finishedAt
		}
		sourceURL := fmt.Sprintf("/uploads/audio/%s/job-%04d.wav", namespace, i)
		refURL := fmt.Sprintf("/uploads/audio/%s/reference-%04d.wav", namespace, i)
		prompt := fmt.Sprintf("生成一段%s风格的虚拟演示音频。", randomMood(rng))
		result := map[string]any{}
		if status == "succeeded" {
			result["output_audio_url"] = fmt.Sprintf("/uploads/processed-audio/%s/output-%04d.wav", namespace, i)
			result["waveform_preview"] = []float64{0.12, 0.24, 0.18, 0.36, 0.22}
		}
		jobs = append(jobs, bulkAudioJob{
			ID:                bulkUUID(namespace, fmt.Sprintf("audio-job-%d", i)),
			UserID:            user.ID,
			Title:             fmt.Sprintf("%s 音频任务 %03d", randomAudioTitle(rng), i+1),
			TaskType:          taskType,
			Status:            status,
			SourceAudioURL:    &sourceURL,
			ReferenceAudioURL: &refURL,
			Prompt:            &prompt,
			Params: map[string]any{
				"tempo":        80 + rng.Intn(80),
				"demo_profile": true,
			},
			Result:       result,
			ErrorMessage: errorMessage,
			AttemptCount: 1 + rng.Intn(3),
			MaxAttempts:  3,
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt,
			StartedAt:    startedAt,
			FinishedAt:   finishedAt,
			LastErrorAt:  lastErrorAt,
		})
	}
	return jobs
}

func buildBulkAudioWorks(namespace string, plan bulkPlan, users []bulkUser, jobs []bulkAudioJob, rng *rand.Rand, now time.Time) []bulkAudioWork {
	works := make([]bulkAudioWork, 0, plan.AudioWorks)
	successJobs := make([]bulkAudioJob, 0, len(jobs))
	for _, job := range jobs {
		if job.Status == "succeeded" {
			successJobs = append(successJobs, job)
		}
	}
	if len(successJobs) == 0 {
		return works
	}
	for i := 0; i < plan.AudioWorks; i++ {
		job := successJobs[i%len(successJobs)]
		user := users[(i*11+9)%len(users)]
		coverURL := fmt.Sprintf("/uploads/images/%s/audio-cover-%03d.webp", namespace, i)
		moderationStatus := "approved"
		var moderationNote *string
		if i%9 == 0 {
			moderationStatus = "pending"
		}
		if i%17 == 0 {
			moderationStatus = "blocked"
			note := "自动生成的模拟审核备注。"
			moderationNote = &note
		}
		works = append(works, bulkAudioWork{
			ID:               bulkUUID(namespace, fmt.Sprintf("audio-work-%d", i)),
			AuthorID:         user.ID,
			SourceJobID:      job.ID,
			Title:            fmt.Sprintf("%s %03d", randomAudioTitle(rng), i+1),
			Description:      sentence(rng, 20),
			CoverImageURL:    &coverURL,
			AudioURL:         fmt.Sprintf("/uploads/processed-audio/%s/public-%03d.wav", namespace, i),
			DurationSec:      float64(90 + rng.Intn(220)),
			Visibility:       map[bool]string{true: "private", false: "public"}[i%6 == 0],
			ModerationStatus: moderationStatus,
			ModerationNote:   moderationNote,
			Tags:             randomTags(rng, 3),
			Metadata: map[string]any{
				"genre":      randomMood(rng),
				"source_job": job.ID.String(),
			},
			PublishedAt: randomPastTime(rng, now, 60),
		})
	}
	return works
}

func buildBulkNotifications(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, comments []bulkComment, works []bulkAudioWork, rng *rand.Rand, now time.Time) []bulkNotification {
	items := make([]bulkNotification, 0, plan.Notifications)
	types := []string{"post_liked", "comment_replied", "followed", "report_reviewed", "event_reminder", "audio_work_liked"}
	for i := 0; i < plan.Notifications; i++ {
		user := users[(i*7+1)%len(users)]
		actor := users[(i*7+5)%len(users)].ID
		nType := types[i%len(types)]
		var targetID *uuid.UUID
		var targetType *string
		switch nType {
		case "post_liked":
			id := posts[i%len(posts)].ID
			targetID = &id
			tt := "post"
			targetType = &tt
		case "comment_replied":
			id := comments[i%len(comments)].ID
			targetID = &id
			tt := "comment"
			targetType = &tt
		case "audio_work_liked":
			if len(works) > 0 {
				id := works[i%len(works)].ID
				targetID = &id
				tt := "audio_work"
				targetType = &tt
			}
		}
		items = append(items, bulkNotification{
			ID:         bulkUUID(namespace, fmt.Sprintf("notification-%d", i)),
			UserID:     user.ID,
			ActorID:    &actor,
			Type:       nType,
			TargetID:   targetID,
			TargetType: targetType,
			IsRead:     i%4 == 0,
			CreatedAt:  randomPastTime(rng, now, 45),
		})
	}
	return items
}

func buildBulkReports(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, comments []bulkComment, works []bulkAudioWork, rng *rand.Rand, now time.Time) []bulkReport {
	items := make([]bulkReport, 0, plan.Reports)
	adminID := users[1].ID
	reasons := []string{"垃圾信息", "攻击性内容", "疑似广告", "内容不当", "侵权风险", "违规图片"}
	seen := make(map[string]struct{})
	for i := 0; i < plan.Reports; i++ {
		reporter := users[(i*5+3)%len(users)]
		var targetType report.TargetType
		var targetID uuid.UUID
		switch i % 4 {
		case 0:
			targetType = report.TargetTypePost
			targetID = posts[i%len(posts)].ID
		case 1:
			targetType = report.TargetTypeComment
			targetID = comments[i%len(comments)].ID
		case 2:
			targetType = report.TargetTypeUser
			targetID = users[(i*7+9)%len(users)].ID
		default:
			if len(works) > 0 {
				targetType = report.TargetTypeAudioWork
				targetID = works[i%len(works)].ID
			} else {
				targetType = report.TargetTypePost
				targetID = posts[i%len(posts)].ID
			}
		}
		key := reporter.ID.String() + ":" + string(targetType) + ":" + targetID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		status := report.StatusPending
		var reviewedBy *uuid.UUID
		var reviewedAt *time.Time
		var actionTaken *report.Action
		if i%3 == 0 {
			status = report.StatusReviewed
			reviewedBy = &adminID
			at := randomPastTime(rng, now, 15)
			reviewedAt = &at
			action := report.ActionNone
			switch targetType {
			case report.TargetTypePost:
				action = report.ActionBlockPost
			case report.TargetTypeComment:
				action = report.ActionDeleteComment
			case report.TargetTypeUser:
				action = report.ActionBanUser
			case report.TargetTypeAudioWork:
				action = report.ActionBlockAudioWork
			}
			actionTaken = &action
		} else if i%5 == 0 {
			status = report.StatusDismissed
			reviewedBy = &adminID
			at := randomPastTime(rng, now, 15)
			reviewedAt = &at
		}
		items = append(items, bulkReport{
			ID:          bulkUUID(namespace, fmt.Sprintf("report-%d", i)),
			ReporterID:  reporter.ID,
			TargetType:  targetType,
			TargetID:    targetID,
			Reason:      reasons[i%len(reasons)],
			Description: sentence(rng, 16),
			Status:      status,
			ReviewedBy:  reviewedBy,
			ReviewedAt:  reviewedAt,
			ActionTaken: actionTaken,
			CreatedAt:   randomPastTime(rng, now, 45),
		})
	}
	return items
}

func buildBulkOrders(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) []bulkOrder {
	items := make([]bulkOrder, 0, plan.Orders)
	statuses := []order.OrderStatus{
		order.OrderStatusPendingPayment,
		order.OrderStatusPaid,
		order.OrderStatusFulfilled,
		order.OrderStatusCancelled,
		order.OrderStatusFailed,
		order.OrderStatusRefunded,
	}
	methods := []order.PaymentMethod{order.PaymentMethodAlipay, order.PaymentMethodWechat}
	for i := 0; i < plan.Orders; i++ {
		user := users[(i*5+1)%len(users)]
		recipient := users[(i*7+11)%len(users)]
		status := statuses[i%len(statuses)]
		createdAt := randomPastTime(rng, now, 120)
		expiresAt := createdAt.Add(6 * time.Hour)
		var paidAt *time.Time
		if status == order.OrderStatusPaid || status == order.OrderStatusFulfilled || status == order.OrderStatusRefunded {
			pt := createdAt.Add(time.Duration(20+rng.Intn(360)) * time.Minute)
			paidAt = &pt
		}
		items = append(items, bulkOrder{
			ID:             bulkUUID(namespace, fmt.Sprintf("order-%d", i)),
			OrderNo:        fmt.Sprintf("TIP-%s-%05d", strings.ToUpper(namespace[:min(4, len(namespace))]), i+1),
			UserID:         user.ID,
			Status:         status,
			TotalCents:     500 + rng.Intn(24500),
			Currency:       "CNY",
			PaymentMethod:  methods[i%len(methods)],
			PaidAt:         paidAt,
			IdempotencyKey: fmt.Sprintf("seed-%s-order-%05d", namespace, i+1),
			Metadata: map[string]any{
				"type":       "tip",
				"to_user_id": recipient.ID.String(),
			},
			CreatedAt: createdAt,
			ExpiresAt: &expiresAt,
		})
	}
	return items
}

func buildBulkMusic(namespace string, plan bulkPlan, rng *rand.Rand, now time.Time) ([]bulkAlbum, []bulkTrack) {
	albums := make([]bulkAlbum, 0, plan.Albums)
	tracks := make([]bulkTrack, 0, plan.Albums*plan.TracksPerAlbum)
	for i := 0; i < plan.Albums; i++ {
		albumID := bulkUUID(namespace, fmt.Sprintf("album-%d", i))
		releaseAt := randomPastTime(rng, now, 500)
		album := bulkAlbum{
			ID:          albumID,
			Slug:        fmt.Sprintf("%s-ost-%03d", namespace, i+1),
			Title:       fmt.Sprintf("%s 原声集 %03d", randomAlbumTitle(rng), i+1),
			Subtitle:    randomMood(rng) + " / 虚拟样本",
			Description: paragraph(rng, 2),
			Artist:      "Furry Studio Collective",
			Composer:    randomComposer(rng),
			ReleaseDate: releaseAt,
			AlbumType:   "ost",
			CreatedAt:   releaseAt,
		}
		albums = append(albums, album)
		for trackNo := 1; trackNo <= plan.TracksPerAlbum; trackNo++ {
			tracks = append(tracks, bulkTrack{
				ID:          bulkUUID(namespace, fmt.Sprintf("album-%d-track-%d", i, trackNo)),
				AlbumID:     albumID,
				TrackNumber: trackNo,
				DiscNumber:  1,
				Title:       fmt.Sprintf("%s %02d", randomTrackTitle(rng), trackNo),
				Artist:      album.Artist,
				DurationSec: 90 + rng.Intn(240),
				PlayCount:   int64(50 + rng.Intn(5000)),
				CreatedAt:   releaseAt,
			})
		}
	}
	return albums, tracks
}

func buildBulkAssistantConversations(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkAssistantConversation, []bulkAssistantMessage) {
	conversations := make([]bulkAssistantConversation, 0, plan.AssistantConversations)
	messages := make([]bulkAssistantMessage, 0, plan.AssistantMessages)

	for i := 0; i < plan.AssistantConversations; i++ {
		createdAt := randomPastTime(rng, now, 60)
		conversations = append(conversations, bulkAssistantConversation{
			ID:        bulkUUID(namespace, fmt.Sprintf("assistant-conv-%d", i)),
			UserID:    users[(i*3+1)%len(users)].ID,
			Title:     fmt.Sprintf("灵感整理会话 %03d", i+1),
			CreatedAt: createdAt,
		})
	}

	for i := 0; i < plan.AssistantMessages; i++ {
		conv := conversations[i%len(conversations)]
		role := "assistant"
		if i%2 == 0 {
			role = "user"
		}
		content := sentence(rng, 16)
		if role == "assistant" {
			content = paragraph(rng, 2)
		}
		messages = append(messages, bulkAssistantMessage{
			ID:             bulkUUID(namespace, fmt.Sprintf("assistant-msg-%d", i)),
			ConversationID: conv.ID,
			Role:           role,
			Content:        content,
			Cards: []map[string]any{
				{"kind": "summary", "title": randomCardTitle(rng)},
			},
			Insights: []map[string]any{
				{"intent": randomIntent(rng), "confidence": 0.72},
			},
			CreatedAt: randomPastTime(rng, now, 60),
		})
	}

	return conversations, messages
}

func buildBulkAssistantFeedback(namespace string, plan bulkPlan, users []bulkUser, conversations []bulkAssistantConversation, rng *rand.Rand, now time.Time) []bulkAssistantFeedback {
	items := make([]bulkAssistantFeedback, 0, plan.AssistantFeedback)
	values := []string{"helpful", "unhelpful"}
	for i := 0; i < plan.AssistantFeedback; i++ {
		conv := conversations[i%len(conversations)]
		userID := conv.UserID
		items = append(items, bulkAssistantFeedback{
			ID:             bulkUUID(namespace, fmt.Sprintf("assistant-feedback-%d", i)),
			ResponseID:     bulkUUID(namespace, fmt.Sprintf("assistant-feedback-response-%d", i)),
			ConversationID: &conv.ID,
			UserID:         &userID,
			Value:          values[i%len(values)],
			QueryText:      sentence(rng, 10),
			ReplyExcerpt:   sentence(rng, 16),
			Provider:       "deepseek",
			Intent:         randomIntent(rng),
			Fallback:       i%6 == 0,
			PagePath:       fmt.Sprintf("/posts/%s", bulkUUID(namespace, fmt.Sprintf("feedback-page-%d", i)).String()),
			SourceCounts:   map[string]int{"post": 2 + rng.Intn(4), "group": 1 + rng.Intn(2)},
			Cards:          []map[string]any{{"kind": "hint", "label": "示例反馈卡片"}},
			CreatedAt:      randomPastTime(rng, now, 45),
		})
	}
	return items
}

func buildBulkKnowledgeDocs(namespace string, plan bulkPlan, posts []bulkPost, groups []bulkGroup, events []bulkEvent, rng *rand.Rand, now time.Time) []bulkKnowledgeDocument {
	items := make([]bulkKnowledgeDocument, 0, plan.AssistantKnowledgeDocs)
	sources := []string{"page", "post", "group", "event"}
	for i := 0; i < plan.AssistantKnowledgeDocs; i++ {
		sourceType := sources[i%len(sources)]
		sourceKey := fmt.Sprintf("%s-doc-%04d", namespace, i)
		href := fmt.Sprintf("/explore/%s/%04d", sourceType, i)
		title := fmt.Sprintf("%s 知识切片 %04d", strings.ToUpper(sourceType), i+1)
		if sourceType == "post" && len(posts) > 0 {
			sourceKey = posts[i%len(posts)].ID.String()
			href = fmt.Sprintf("/posts/%s", sourceKey)
			title = posts[i%len(posts)].Title
		} else if sourceType == "group" && len(groups) > 0 {
			sourceKey = groups[i%len(groups)].ID.String()
			href = fmt.Sprintf("/groups/%s", sourceKey)
			title = groups[i%len(groups)].Name
		} else if sourceType == "event" && len(events) > 0 {
			sourceKey = events[i%len(events)].ID.String()
			href = fmt.Sprintf("/events/%s", sourceKey)
			title = events[i%len(events)].Title
		}
		content := paragraph(rng, 3)
		items = append(items, bulkKnowledgeDocument{
			ID:              bulkUUID(namespace, fmt.Sprintf("knowledge-doc-%d", i)),
			SourceType:      sourceType,
			SourceKey:       sourceKey,
			ChunkIndex:      i / len(sources),
			Title:           title,
			Summary:         sentence(rng, 12),
			Content:         content,
			Href:            href,
			Meta:            randomTags(rng, 2)[0],
			SourceLabel:     sourceType,
			Tags:            randomTags(rng, 3),
			SearchText:      content,
			Embedding:       []float64{0.12, 0.27, 0.41, 0.55, 0.63, 0.74},
			IndexedAt:       randomPastTime(rng, now, 20),
			SourceUpdatedAt: randomPastTime(rng, now, 20),
		})
	}
	return items
}

func buildBulkAuditLogs(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, orders []bulkOrder, reports []bulkReport, rng *rand.Rand, now time.Time) []bulkAuditLog {
	items := make([]bulkAuditLog, 0, plan.AuditLogs)
	actions := []audit.Action{audit.ActionCreate, audit.ActionUpdate, audit.ActionDelete, audit.ActionView, audit.ActionExport}
	resources := []audit.Resource{audit.ResourceUser, audit.ResourcePost, audit.ResourceOrder, audit.ResourceReport, audit.ResourceAssistant, audit.ResourceSystem}

	for i := 0; i < plan.AuditLogs; i++ {
		operator := users[i%min(8, len(users))]
		resource := resources[i%len(resources)]
		action := actions[i%len(actions)]
		var resourceID *uuid.UUID
		switch resource {
		case audit.ResourcePost:
			id := posts[i%len(posts)].ID
			resourceID = &id
		case audit.ResourceOrder:
			id := orders[i%len(orders)].ID
			resourceID = &id
		case audit.ResourceReport:
			id := reports[i%len(reports)].ID
			resourceID = &id
		default:
			id := operator.ID
			resourceID = &id
		}
		after := stringPtr(fmt.Sprintf(`{"namespace":"%s","index":%d}`, namespace, i))
		var before *string
		if action == audit.ActionUpdate {
			before = stringPtr(`{"status":"before"}`)
		}
		var errorMessage *string
		if i%19 == 0 {
			errorMessage = stringPtr("模拟后台操作失败，用于审计页展示。")
		}
		userID := operator.ID
		items = append(items, bulkAuditLog{
			ID:           bulkUUID(namespace, fmt.Sprintf("audit-log-%d", i)),
			UserID:       &userID,
			Username:     operator.Username,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			IPAddress:    fmt.Sprintf("10.0.%d.%d", (i%32)+1, (i%200)+10),
			UserAgent:    "bulk-seeder/1.0",
			BeforeData:   before,
			AfterData:    after,
			ErrorMessage: errorMessage,
			CreatedAt:    randomPastTime(rng, now, 45),
		})
	}
	return items
}

func buildBulkAnalyticsEvents(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) []bulkAnalyticsEvent {
	items := make([]bulkAnalyticsEvent, 0, plan.AnalyticsEvents)
	eventTypes := []string{"user_login", "user_register", "post_view", "game_view", "game_download", "purchase_complete", "audio_play"}
	for i := 0; i < plan.AnalyticsEvents; i++ {
		user := users[(i*5+7)%len(users)]
		var userID *uuid.UUID
		if i%9 != 0 {
			userID = &user.ID
		}
		ref := "https://seed.local/feed"
		items = append(items, bulkAnalyticsEvent{
			ID:        bulkUUID(namespace, fmt.Sprintf("analytics-event-%d", i)),
			EventType: eventTypes[i%len(eventTypes)],
			UserID:    userID,
			SessionID: fmt.Sprintf("%s-session-%05d", namespace, i),
			Properties: map[string]any{
				"namespace": namespace,
				"screen":    []string{"home", "post", "games", "audio"}[i%4],
			},
			IPAddress: fmt.Sprintf("192.168.%d.%d", (i%12)+1, (i%180)+20),
			UserAgent: "bulk-seeder/browser",
			Referrer:  &ref,
			CreatedAt: randomPastTime(rng, now, 60),
		})
	}
	return items
}

func buildBulkHexBlitz(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkHexMatch, []bulkHexResult, []bulkHexMoveEvent) {
	matches := make([]bulkHexMatch, 0, plan.HexBlitzMatches)
	results := make([]bulkHexResult, 0, plan.HexBlitzMatches*4)
	moves := make([]bulkHexMoveEvent, 0, plan.HexBlitzMatches*plan.HexBlitzMoveEventsPerHit)

	for i := 0; i < plan.HexBlitzMatches; i++ {
		startedAt := randomPastTime(rng, now, 60)
		duration := 75 + rng.Intn(60)
		match := bulkHexMatch{
			ID:         bulkUUID(namespace, fmt.Sprintf("hex-match-%d", i)),
			RoomID:     bulkUUID(namespace, fmt.Sprintf("hex-room-%d", i)),
			RoomCode:   fmt.Sprintf("HX%04d", i+1),
			RoomTitle:  fmt.Sprintf("Hex Blitz 训练房 %03d", i+1),
			Seed:       int64(100000 + i),
			StartedAt:  startedAt,
			FinishedAt: startedAt.Add(time.Duration(duration) * time.Second),
		}
		matches = append(matches, match)
		scoreBase := 2200 + rng.Intn(1200)
		for seat := 0; seat < 4; seat++ {
			user := users[(i*7+seat*3)%len(users)]
			resultID := bulkUUID(namespace, fmt.Sprintf("hex-result-%d-%d", i, seat))
			sessionID := fmt.Sprintf("%s-hx-%03d-%d", namespace, i, seat)
			score := scoreBase - seat*180 + rng.Intn(90)
			results = append(results, bulkHexResult{
				ID:         resultID,
				MatchID:    match.ID,
				RoomID:     match.RoomID,
				RoomCode:   match.RoomCode,
				RoomTitle:  match.RoomTitle,
				UserID:     &user.ID,
				PlayerName: user.Username,
				SessionID:  sessionID,
				Score:      score,
				Rank:       seat + 1,
				CreatedAt:  match.FinishedAt,
			})
			scoreAfter := 0
			for move := 0; move < plan.HexBlitzMoveEventsPerHit/4; move++ {
				gained := 20 + rng.Intn(180)
				scoreAfter += gained
				moves = append(moves, bulkHexMoveEvent{
					ID:           bulkUUID(namespace, fmt.Sprintf("hex-move-%d-%d-%d", i, seat, move)),
					MatchID:      match.ID,
					SessionID:    sessionID,
					UserID:       &user.ID,
					PlayerName:   user.Username,
					TileID:       fmt.Sprintf("tile-%d-%d", seat, move),
					MoveIndex:    move + 1,
					ClearedCount: 2 + rng.Intn(5),
					GainedScore:  gained,
					ScoreAfter:   scoreAfter,
					ComboAfter:   1 + rng.Intn(5),
					OccurredAt:   match.StartedAt.Add(time.Duration(move*6+seat) * time.Second),
				})
			}
		}
	}

	return matches, results, moves
}

func buildBulkDoudizhu(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) ([]bulkDoudizhuMatch, []bulkDoudizhuPlayer, []bulkDoudizhuAction) {
	matches := make([]bulkDoudizhuMatch, 0, plan.DoudizhuMatches)
	players := make([]bulkDoudizhuPlayer, 0, plan.DoudizhuMatches*3)
	actions := make([]bulkDoudizhuAction, 0, plan.DoudizhuMatches*plan.DoudizhuActionsPerMatch)

	for i := 0; i < plan.DoudizhuMatches; i++ {
		startedAt := randomPastTime(rng, now, 60)
		matchMode := "pvp"
		if i%3 == 0 {
			matchMode = "demo_ai"
		}
		landlordSeat := int16(i % 3)
		winnerSide := "landlord"
		if i%2 == 0 {
			winnerSide = "farmer"
		}
		match := bulkDoudizhuMatch{
			ID:           bulkUUID(namespace, fmt.Sprintf("ddz-match-%d", i)),
			RoomID:       bulkUUID(namespace, fmt.Sprintf("ddz-room-%d", i)),
			RoomCode:     fmt.Sprintf("DZ%04d", i+1),
			RoomTitle:    fmt.Sprintf("斗地主牌桌 %03d", i+1),
			MatchMode:    matchMode,
			StartedAt:    startedAt,
			FinishedAt:   startedAt.Add(time.Duration(6+rng.Intn(6)) * time.Minute),
			LandlordSeat: landlordSeat,
			WinnerSide:   winnerSide,
			Multiplier:   1 + rng.Intn(8),
			BombCount:    rng.Intn(3),
			Spring:       i%11 == 0,
			AntiSpring:   i%13 == 0,
		}
		matches = append(matches, match)

		for seat := 0; seat < 3; seat++ {
			user := users[(i*5+seat*7+9)%len(users)]
			isBot := matchMode == "demo_ai" && seat > 0
			role := "farmer"
			if int16(seat) == landlordSeat {
				role = "landlord"
			}
			isWinner := (role == winnerSide)
			scoreDelta := 1 + rng.Intn(6)
			if !isWinner {
				scoreDelta = -scoreDelta
			}
			sessionID := fmt.Sprintf("%s-ddz-%03d-%d", namespace, i, seat)
			playerName := user.Username
			var userID *uuid.UUID = &user.ID
			var botLevel *string
			if isBot {
				level := "demo"
				botLevel = &level
				playerName = fmt.Sprintf("AI-%d", seat)
				userID = nil
			}
			players = append(players, bulkDoudizhuPlayer{
				ID:         bulkUUID(namespace, fmt.Sprintf("ddz-player-%d-%d", i, seat)),
				MatchID:    match.ID,
				SessionID:  sessionID,
				UserID:     userID,
				IsBot:      isBot,
				BotLevel:   botLevel,
				Seat:       int16(seat),
				PlayerName: playerName,
				Role:       role,
				BidScore:   1 + rng.Intn(3),
				CardsLeft:  rng.Intn(8),
				IsWinner:   isWinner,
				ScoreDelta: scoreDelta,
				CreatedAt:  match.FinishedAt,
			})
		}

		for actionIndex := 0; actionIndex < plan.DoudizhuActionsPerMatch; actionIndex++ {
			seat := int16(actionIndex % 3)
			sessionID := fmt.Sprintf("%s-ddz-%03d-%d", namespace, i, seat)
			playerName := fmt.Sprintf("玩家-%d", seat+1)
			var userID *uuid.UUID
			if matchMode != "demo_ai" || seat == 0 {
				user := users[(i*5+int(seat)*7+9)%len(users)]
				playerName = user.Username
				userID = &user.ID
			}
			actionType := []string{"bid", "play_cards", "pass_turn"}[actionIndex%3]
			cards := []map[string]any{
				{"suit": "spade", "rank": 10 + actionIndex%5},
			}
			var comboType *string
			var comboMainRank *int
			var comboSeqLen *int
			var comboTotalCards *int
			if actionType == "play_cards" {
				ct := []string{"single", "pair", "straight"}[actionIndex%3]
				comboType = &ct
				mainRank := 10 + actionIndex%4
				seqLen := 1 + actionIndex%5
				totalCards := 1 + actionIndex%5
				comboMainRank = &mainRank
				comboSeqLen = &seqLen
				comboTotalCards = &totalCards
			}
			actions = append(actions, bulkDoudizhuAction{
				ID:                  bulkUUID(namespace, fmt.Sprintf("ddz-action-%d-%d", i, actionIndex)),
				MatchID:             match.ID,
				TurnNo:              1 + actionIndex/3,
				ActionIndex:         actionIndex + 1,
				SessionID:           sessionID,
				UserID:              userID,
				PlayerName:          playerName,
				Seat:                seat,
				ActionType:          actionType,
				CardsJSON:           cards,
				ComboType:           comboType,
				ComboMainRank:       comboMainRank,
				ComboSequenceLength: comboSeqLen,
				ComboTotalCards:     comboTotalCards,
				MultiplierAfter:     1 + actionIndex/8,
				OccurredAt:          match.StartedAt.Add(time.Duration(actionIndex*9) * time.Second),
			})
		}
	}

	return matches, players, actions
}

func buildBulkPointTransactions(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) []struct {
	UserID    uuid.UUID
	Amount    int
	Source    string
	RefID     *string
	Note      string
	CreatedAt time.Time
} {
	items := make([]struct {
		UserID    uuid.UUID
		Amount    int
		Source    string
		RefID     *string
		Note      string
		CreatedAt time.Time
	}, 0, plan.PointTransactions)
	sources := []string{"register", "comment", "achievement", "admin", "daily_checkin"}
	for i := 0; i < plan.PointTransactions; i++ {
		amount := 5 + rng.Intn(60)
		if i%11 == 0 {
			amount = -(1 + rng.Intn(10))
		}
		ref := fmt.Sprintf("%s-point-%05d", namespace, i)
		items = append(items, struct {
			UserID    uuid.UUID
			Amount    int
			Source    string
			RefID     *string
			Note      string
			CreatedAt time.Time
		}{
			UserID:    users[(i*7+3)%len(users)].ID,
			Amount:    amount,
			Source:    sources[i%len(sources)],
			RefID:     &ref,
			Note:      fmt.Sprintf("[bulk:%s] 虚拟积分流水 %05d", namespace, i),
			CreatedAt: randomPastTime(rng, now, 90),
		})
	}
	return items
}

func buildBulkUserAchievements(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) [][3]any {
	items := make([][3]any, 0, plan.UserAchievements)
	achievementIDs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	seen := make(map[string]struct{})
	for i := 0; i < plan.UserAchievements; i++ {
		user := users[(i*7+1)%len(users)]
		achievementID := achievementIDs[i%len(achievementIDs)]
		key := fmt.Sprintf("%s:%d", user.ID, achievementID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, [3]any{user.ID, achievementID, randomPastTime(rng, now, 160)})
	}
	return items
}

func buildBulkPostLikes(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, rng *rand.Rand, now time.Time) [][3]any {
	return buildUniqueUserTargetPairs(plan.PostLikes, users, len(posts), func(idx int) uuid.UUID {
		return posts[idx].ID
	}, rng, now)
}

func buildBulkCommentLikes(namespace string, plan bulkPlan, users []bulkUser, comments []bulkComment, rng *rand.Rand, now time.Time) [][3]any {
	return buildUniqueUserTargetPairs(plan.CommentLikes, users, len(comments), func(idx int) uuid.UUID {
		return comments[idx].ID
	}, rng, now)
}

func buildBulkAudioWorkLikes(namespace string, plan bulkPlan, users []bulkUser, works []bulkAudioWork, rng *rand.Rand, now time.Time) [][3]any {
	return buildUniqueUserTargetPairs(plan.AudioWorkLikes, users, len(works), func(idx int) uuid.UUID {
		return works[idx].ID
	}, rng, now)
}

func buildUniqueUserTargetPairs(count int, users []bulkUser, targetCount int, targetResolver func(idx int) uuid.UUID, rng *rand.Rand, now time.Time) [][3]any {
	items := make([][3]any, 0, count)
	if targetCount == 0 {
		return items
	}
	seen := make(map[string]struct{})
	for len(items) < count {
		user := users[rng.Intn(len(users))]
		targetIdx := rng.Intn(targetCount)
		targetID := targetResolver(targetIdx)
		key := user.ID.String() + ":" + targetID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, [3]any{user.ID, targetID, randomPastTime(rng, now, 90)})
	}
	return items
}

func buildBulkBookmarks(namespace string, plan bulkPlan, users []bulkUser, posts []bulkPost, groups []bulkGroup, events []bulkEvent, works []bulkAudioWork, rng *rand.Rand, now time.Time) [][4]any {
	items := make([][4]any, 0, plan.Bookmarks)
	seen := make(map[string]struct{})
	targetTypes := []string{"post", "group", "event", "audio_work"}
	for len(items) < plan.Bookmarks {
		user := users[rng.Intn(len(users))]
		targetType := targetTypes[rng.Intn(len(targetTypes))]
		var targetID uuid.UUID
		switch targetType {
		case "post":
			targetID = posts[rng.Intn(len(posts))].ID
		case "group":
			targetID = groups[rng.Intn(len(groups))].ID
		case "event":
			targetID = events[rng.Intn(len(events))].ID
		default:
			if len(works) == 0 {
				continue
			}
			targetID = works[rng.Intn(len(works))].ID
		}
		key := user.ID.String() + ":" + targetType + ":" + targetID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, [4]any{user.ID, targetType, targetID, randomPastTime(rng, now, 90)})
	}
	return items
}

func buildBulkFollows(namespace string, plan bulkPlan, users []bulkUser, rng *rand.Rand, now time.Time) [][3]any {
	items := make([][3]any, 0, plan.Follows)
	seen := make(map[string]struct{})
	for len(items) < plan.Follows {
		follower := users[rng.Intn(len(users))]
		followee := users[rng.Intn(len(users))]
		if follower.ID == followee.ID {
			continue
		}
		key := follower.ID.String() + ":" + followee.ID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, [3]any{follower.ID, followee.ID, randomPastTime(rng, now, 120)})
	}
	return items
}

func insertBulkUsers(b *bulkBatcher, users []bulkUser, passwordHash string) error {
	const sql = `
		INSERT INTO users (
			id, username, email, password_hash, avatar_key, bio, location, website,
			furry_name, species, role, status, force_password_reset, email_verified_at,
			last_login_at, last_login_ip, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT (email) DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			avatar_key = EXCLUDED.avatar_key,
			bio = EXCLUDED.bio,
			location = EXCLUDED.location,
			website = EXCLUDED.website,
			furry_name = EXCLUDED.furry_name,
			species = EXCLUDED.species,
			role = EXCLUDED.role,
			status = EXCLUDED.status,
			force_password_reset = EXCLUDED.force_password_reset,
			email_verified_at = EXCLUDED.email_verified_at,
			last_login_at = EXCLUDED.last_login_at,
			last_login_ip = EXCLUDED.last_login_ip,
			updated_at = EXCLUDED.updated_at
	`
	for _, user := range users {
		avatar := fmt.Sprintf("avatars/%s.webp", user.Username)
		lastLoginAt := user.CreatedAt.Add(48 * time.Hour)
		lastLoginIP := fmt.Sprintf("172.20.%d.%d", user.ID[0]%255, user.ID[1]%255)
		if err := b.Queue(sql,
			user.ID,
			user.Username,
			user.Email,
			passwordHash,
			avatar,
			user.Bio,
			user.Location,
			user.Website,
			user.FurryName,
			user.Species,
			user.Role,
			user.Status,
			false,
			user.CreatedAt,
			lastLoginAt,
			lastLoginIP,
			user.CreatedAt,
			user.CreatedAt,
		); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkGroups(b *bulkBatcher, groups []bulkGroup) error {
	const sql = `
		INSERT INTO groups (
			id, owner_id, name, description, announcement, rules, tags, privacy,
			member_count, post_count, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,0,0,$9,$9)
		ON CONFLICT (id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			announcement = EXCLUDED.announcement,
			rules = EXCLUDED.rules,
			tags = EXCLUDED.tags,
			privacy = EXCLUDED.privacy,
			updated_at = EXCLUDED.updated_at
	`
	for _, group := range groups {
		if err := b.Queue(sql, group.ID, group.OwnerID, group.Name, group.Description, group.Announcement, group.Rules, mustJSON(group.Tags), group.Privacy, group.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkGroupMembers(b *bulkBatcher, members [][3]any) error {
	const sql = `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (group_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`
	for _, member := range members {
		if err := b.Queue(sql, member[0], member[1], member[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkGroupAnnouncements(b *bulkBatcher, items []struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
}) error {
	const sql = `
		INSERT INTO group_announcements (id, group_id, author_id, content, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, created_at = EXCLUDED.created_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.GroupID, item.AuthorID, item.Content, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkEvents(b *bulkBatcher, events []bulkEvent) error {
	const sql = `
		INSERT INTO events (
			id, organizer_id, title, description, location, is_online, start_time,
			end_time, max_capacity, tags, status, attendee_count, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,$12,$12)
		ON CONFLICT (id) DO UPDATE SET
			organizer_id = EXCLUDED.organizer_id,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			location = EXCLUDED.location,
			is_online = EXCLUDED.is_online,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			max_capacity = EXCLUDED.max_capacity,
			tags = EXCLUDED.tags,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	for _, event := range events {
		if err := b.Queue(sql, event.ID, event.OrganizerID, event.Title, event.Description, event.Location, event.IsOnline, event.StartTime, event.EndTime, event.MaxCapacity, mustJSON(event.Tags), event.Status, event.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkEventAttendees(b *bulkBatcher, attendees [][3]any) error {
	const sql = `
		INSERT INTO event_attendees (event_id, user_id, status, joined_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (event_id, user_id) DO UPDATE SET status = EXCLUDED.status
	`
	for _, attendee := range attendees {
		if err := b.Queue(sql, attendee[0], attendee[1], attendee[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkPosts(b *bulkBatcher, posts []bulkPost) error {
	const sql = `
		INSERT INTO posts (
			id, author_id, group_id, title, content, media_urls, tags, visibility,
			moderation_status, content_labels, like_count, comment_count, is_pinned, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,0,0,$11,$12,$12)
		ON CONFLICT (id) DO UPDATE SET
			author_id = EXCLUDED.author_id,
			group_id = EXCLUDED.group_id,
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			media_urls = EXCLUDED.media_urls,
			tags = EXCLUDED.tags,
			visibility = EXCLUDED.visibility,
			moderation_status = EXCLUDED.moderation_status,
			content_labels = EXCLUDED.content_labels,
			is_pinned = EXCLUDED.is_pinned,
			updated_at = EXCLUDED.updated_at
	`
	for i, post := range posts {
		media := []string{}
		if i%5 == 0 {
			media = append(media, fmt.Sprintf("/uploads/images/post-%s-%03d.webp", post.AuthorID.String()[:8], i))
		}
		if err := b.Queue(sql, post.ID, post.AuthorID, post.GroupID, post.Title, post.Content, mustJSON(media), post.Tags, post.Visibility, post.ModerationStatus, mustJSON(post.ContentLabels), i%9 == 0, post.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func assignBulkFeaturedPosts(b *bulkBatcher, featured map[uuid.UUID]uuid.UUID) error {
	const sql = `
		UPDATE groups
		SET featured_post_id = $2, updated_at = NOW()
		WHERE id = $1
	`
	for groupID, postID := range featured {
		if err := b.Queue(sql, groupID, postID); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkPostLikes(b *bulkBatcher, items [][3]any) error {
	const sql = `
		INSERT INTO post_likes (post_id, user_id, created_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`
	for _, item := range items {
		if err := b.Queue(sql, item[1], item[0], item[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkComments(b *bulkBatcher, comments []bulkComment) error {
	const sql = `
		INSERT INTO comments (
			id, user_id, commentable_type, commentable_id, parent_id, content,
			is_edited, is_deleted, like_count, reply_count, created_at, updated_at
		)
		VALUES ($1,$2,'post',$3,$4,$5,FALSE,FALSE,0,0,$6,$6)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			parent_id = EXCLUDED.parent_id,
			updated_at = EXCLUDED.updated_at
	`
	for _, item := range comments {
		if err := b.Queue(sql, item.ID, item.UserID, item.PostID, item.ParentID, item.Content, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkCommentLikes(b *bulkBatcher, items [][3]any) error {
	const sql = `
		INSERT INTO comment_likes (id, user_id, comment_id, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, comment_id) DO NOTHING
	`
	for idx, item := range items {
		id := bulkUUID("comment-like", fmt.Sprintf("%v-%d", item[1], idx))
		if err := b.Queue(sql, id, item[0], item[1], item[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkConversations(b *bulkBatcher, items []bulkConversation) error {
	const sql = `
		INSERT INTO conversations (id, type, name, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (id) DO UPDATE SET type = EXCLUDED.type, name = EXCLUDED.name, updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.Type, item.Name, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkConversationMembers(b *bulkBatcher, conversations []bulkConversation) error {
	const sql = `
		INSERT INTO conversation_members (conversation_id, user_id, joined_at, last_read_at)
		VALUES ($1,$2,$3,$3)
		ON CONFLICT (conversation_id, user_id) DO NOTHING
	`
	for _, conv := range conversations {
		for _, memberID := range conv.Members {
			if err := b.Queue(sql, conv.ID, memberID, conv.CreatedAt); err != nil {
				return err
			}
		}
	}
	return b.Flush()
}

func insertBulkMessages(b *bulkBatcher, messages []bulkMessage) error {
	const sql = `
		INSERT INTO messages (id, conversation_id, sender_id, content, media_url, is_read, created_at)
		VALUES ($1,$2,$3,$4,NULL,FALSE,$5)
		ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, created_at = EXCLUDED.created_at
	`
	for _, item := range messages {
		if err := b.Queue(sql, item.ID, item.ConversationID, item.SenderID, item.Content, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkNotifications(b *bulkBatcher, items []bulkNotification) error {
	const sql = `
		INSERT INTO notifications (id, user_id, actor_id, type, target_id, target_type, is_read, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			actor_id = EXCLUDED.actor_id,
			type = EXCLUDED.type,
			target_id = EXCLUDED.target_id,
			target_type = EXCLUDED.target_type,
			is_read = EXCLUDED.is_read,
			created_at = EXCLUDED.created_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.UserID, item.ActorID, item.Type, item.TargetID, item.TargetType, item.IsRead, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkReports(b *bulkBatcher, items []bulkReport) error {
	const sql = `
		INSERT INTO reports (
			id, reporter_id, target_type, target_id, reason, description, status,
			reviewed_by, reviewed_at, action_taken, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			reason = EXCLUDED.reason,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			reviewed_by = EXCLUDED.reviewed_by,
			reviewed_at = EXCLUDED.reviewed_at,
			action_taken = EXCLUDED.action_taken
	`
	for _, item := range items {
		var actionValue *string
		if item.ActionTaken != nil {
			value := string(*item.ActionTaken)
			actionValue = &value
		}
		if err := b.Queue(sql, item.ID, item.ReporterID, item.TargetType, item.TargetID, item.Reason, item.Description, item.Status, item.ReviewedBy, item.ReviewedAt, actionValue, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkOrders(b *bulkBatcher, items []bulkOrder) error {
	const sql = `
		INSERT INTO orders (
			id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code,
			payment_method, paid_at, idempotency_key, metadata, created_at, expires_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,0,NULL,$7,$8,$9,$10,$11,$12,$11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			total_cents = EXCLUDED.total_cents,
			payment_method = EXCLUDED.payment_method,
			paid_at = EXCLUDED.paid_at,
			metadata = EXCLUDED.metadata,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.OrderNo, item.UserID, item.Status, item.TotalCents, item.Currency, item.PaymentMethod, item.PaidAt, item.IdempotencyKey, mustJSON(item.Metadata), item.CreatedAt, item.ExpiresAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAlbums(b *bulkBatcher, items []bulkAlbum) error {
	const sql = `
		INSERT INTO albums (
			id, game_id, slug, title, subtitle, description, cover_key, artist, composer,
			arranger, lyricist, total_tracks, duration_sec, release_date, album_type, created_at, updated_at
		)
		VALUES ($1,NULL,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$15)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			subtitle = EXCLUDED.subtitle,
			description = EXCLUDED.description,
			cover_key = EXCLUDED.cover_key,
			artist = EXCLUDED.artist,
			composer = EXCLUDED.composer,
			arranger = EXCLUDED.arranger,
			lyricist = EXCLUDED.lyricist,
			total_tracks = EXCLUDED.total_tracks,
			duration_sec = EXCLUDED.duration_sec,
			release_date = EXCLUDED.release_date,
			album_type = EXCLUDED.album_type,
			updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		durationSec := 0
		if err := b.Queue(sql,
			item.ID,
			item.Slug,
			item.Title,
			item.Subtitle,
			item.Description,
			fmt.Sprintf("covers/%s.webp", item.Slug),
			item.Artist,
			item.Composer,
			item.Composer,
			"",
			0,
			durationSec,
			item.ReleaseDate,
			item.AlbumType,
			item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkTracks(b *bulkBatcher, items []bulkTrack) error {
	const sql = `
		INSERT INTO tracks (
			id, album_id, track_number, disc_number, title, artist, duration_sec, stream_key, stream_size,
			hifi_key, hifi_format, hifi_bitdepth, hifi_samplerate, hifi_size, lrc_key, play_count, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$17)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			artist = EXCLUDED.artist,
			duration_sec = EXCLUDED.duration_sec,
			stream_key = EXCLUDED.stream_key,
			play_count = EXCLUDED.play_count,
			updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		streamKey := fmt.Sprintf("ost/%s/%s.mp3", item.AlbumID.String()[:8], item.ID.String()[:8])
		hifiKey := fmt.Sprintf("ost/%s/%s.flac", item.AlbumID.String()[:8], item.ID.String()[:8])
		lrcKey := fmt.Sprintf("ost/%s/%s.lrc", item.AlbumID.String()[:8], item.ID.String()[:8])
		if err := b.Queue(sql, item.ID, item.AlbumID, item.TrackNumber, item.DiscNumber, item.Title, item.Artist, item.DurationSec, streamKey, 4_000_000+item.PlayCount, hifiKey, "flac", 24, 48000, 18_000_000+item.PlayCount, lrcKey, item.PlayCount, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAudioJobs(b *bulkBatcher, items []bulkAudioJob) error {
	const sql = `
		INSERT INTO audio_jobs (
			id, user_id, title, task_type, status, source_audio_url, reference_audio_url,
			prompt, params, result, error_message, attempt_count, max_attempts,
			created_at, updated_at, started_at, finished_at, next_retry_at, last_error_at, dead_lettered_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,NULL,$18,NULL)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			result = EXCLUDED.result,
			error_message = EXCLUDED.error_message,
			attempt_count = EXCLUDED.attempt_count,
			updated_at = EXCLUDED.updated_at,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at,
			last_error_at = EXCLUDED.last_error_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.UserID, item.Title, item.TaskType, item.Status, item.SourceAudioURL, item.ReferenceAudioURL, item.Prompt, mustJSON(item.Params), mustJSON(item.Result), item.ErrorMessage, item.AttemptCount, item.MaxAttempts, item.CreatedAt, item.UpdatedAt, item.StartedAt, item.FinishedAt, item.LastErrorAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAudioWorks(b *bulkBatcher, items []bulkAudioWork) error {
	const sql = `
		INSERT INTO audio_works (
			id, author_id, source_job_id, title, description, cover_image_url, audio_url,
			duration_sec, visibility, moderation_status, moderation_note, like_count, comment_count,
			tags, waveform_preview, metadata, published_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,0,$12,$13,$14,$15,$15,$15)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			cover_image_url = EXCLUDED.cover_image_url,
			audio_url = EXCLUDED.audio_url,
			duration_sec = EXCLUDED.duration_sec,
			visibility = EXCLUDED.visibility,
			moderation_status = EXCLUDED.moderation_status,
			moderation_note = EXCLUDED.moderation_note,
			tags = EXCLUDED.tags,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		waveform := []float64{0.05, 0.13, 0.27, 0.33, 0.24, 0.11}
		if err := b.Queue(sql, item.ID, item.AuthorID, item.SourceJobID, item.Title, item.Description, item.CoverImageURL, item.AudioURL, item.DurationSec, item.Visibility, item.ModerationStatus, item.ModerationNote, mustJSON(item.Tags), mustJSON(waveform), mustJSON(item.Metadata), item.PublishedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAudioWorkLikes(b *bulkBatcher, items [][3]any) error {
	const sql = `
		INSERT INTO audio_work_likes (user_id, work_id, created_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id, work_id) DO NOTHING
	`
	for _, item := range items {
		if err := b.Queue(sql, item[0], item[1], item[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkBookmarks(b *bulkBatcher, items [][4]any) error {
	const sql = `
		INSERT INTO user_bookmarks (user_id, target_type, target_id, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, target_type, target_id) DO NOTHING
	`
	for _, item := range items {
		if err := b.Queue(sql, item[0], item[1], item[2], item[3]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkFollows(b *bulkBatcher, items [][3]any) error {
	const sql = `
		INSERT INTO user_follows (follower_id, followee_id, created_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`
	for _, item := range items {
		if err := b.Queue(sql, item[0], item[1], item[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAssistantConversations(b *bulkBatcher, items []bulkAssistantConversation) error {
	const sql = `
		INSERT INTO assistant_conversations (id, user_id, title, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, updated_at = EXCLUDED.updated_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.UserID, item.Title, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAssistantMessages(b *bulkBatcher, items []bulkAssistantMessage) error {
	const sql = `
		INSERT INTO assistant_messages (id, conversation_id, role, content, cards, insights, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, cards = EXCLUDED.cards, insights = EXCLUDED.insights
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.ConversationID, item.Role, item.Content, mustJSON(item.Cards), mustJSON(item.Insights), item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAssistantFeedback(b *bulkBatcher, items []bulkAssistantFeedback) error {
	const sql = `
		INSERT INTO assistant_feedback (
			id, response_id, conversation_id, user_id, value, query_text, reply_excerpt, provider,
			intent, fallback, page_path, source_counts, cards, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (response_id) DO UPDATE SET
			value = EXCLUDED.value,
			reply_excerpt = EXCLUDED.reply_excerpt,
			intent = EXCLUDED.intent,
			fallback = EXCLUDED.fallback,
			source_counts = EXCLUDED.source_counts,
			cards = EXCLUDED.cards
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.ResponseID, item.ConversationID, item.UserID, item.Value, item.QueryText, item.ReplyExcerpt, item.Provider, item.Intent, item.Fallback, item.PagePath, mustJSON(item.SourceCounts), mustJSON(item.Cards), item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkKnowledgeDocs(b *bulkBatcher, items []bulkKnowledgeDocument) error {
	const sql = `
		INSERT INTO assistant_knowledge_documents (
			id, source_type, source_key, chunk_index, title, summary, content, href, meta,
			source_label, tags, search_text, search_vector, embedding, indexed_at, source_updated_at, is_active
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,
			$10,$11,$12,to_tsvector('simple', $12),$13,$14,$15,TRUE
		)
		ON CONFLICT (source_type, source_key, chunk_index) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			href = EXCLUDED.href,
			meta = EXCLUDED.meta,
			source_label = EXCLUDED.source_label,
			tags = EXCLUDED.tags,
			search_text = EXCLUDED.search_text,
			search_vector = EXCLUDED.search_vector,
			embedding = EXCLUDED.embedding,
			indexed_at = EXCLUDED.indexed_at,
			source_updated_at = EXCLUDED.source_updated_at,
			is_active = EXCLUDED.is_active
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.SourceType, item.SourceKey, item.ChunkIndex, item.Title, item.Summary, item.Content, item.Href, item.Meta, item.SourceLabel, mustJSON(item.Tags), item.SearchText, mustJSON(item.Embedding), item.IndexedAt, item.SourceUpdatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAuditLogs(b *bulkBatcher, items []bulkAuditLog) error {
	const sql = `
		INSERT INTO audit_logs (
			id, user_id, username, action, resource, resource_id, ip_address, user_agent,
			before_data, after_data, error_message, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			after_data = EXCLUDED.after_data,
			error_message = EXCLUDED.error_message,
			created_at = EXCLUDED.created_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.UserID, item.Username, item.Action, item.Resource, item.ResourceID, item.IPAddress, item.UserAgent, item.BeforeData, item.AfterData, item.ErrorMessage, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkAnalyticsEvents(b *bulkBatcher, items []bulkAnalyticsEvent) error {
	const sql = `
		INSERT INTO analytics_events (id, event_type, user_id, session_id, properties, ip_address, user_agent, referrer, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			properties = EXCLUDED.properties,
			created_at = EXCLUDED.created_at
	`
	for _, item := range items {
		if err := b.Queue(sql, item.ID, item.EventType, item.UserID, item.SessionID, mustJSON(item.Properties), item.IPAddress, item.UserAgent, item.Referrer, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkHexBlitzMatches(b *bulkBatcher, matches []bulkHexMatch, results []bulkHexResult, moves []bulkHexMoveEvent) error {
	const matchSQL = `
		INSERT INTO hex_blitz_matches (id, room_id, room_code, room_title, game_slug, seed, started_at, finished_at, duration_sec, created_at)
		VALUES ($1,$2,$3,$4,'hex-blitz',$5,$6,$7,$8,$7)
		ON CONFLICT (id) DO UPDATE SET finished_at = EXCLUDED.finished_at, duration_sec = EXCLUDED.duration_sec, seed = EXCLUDED.seed
	`
	for _, item := range matches {
		duration := int(item.FinishedAt.Sub(item.StartedAt).Seconds())
		if err := b.Queue(matchSQL, item.ID, item.RoomID, item.RoomCode, item.RoomTitle, item.Seed, item.StartedAt, item.FinishedAt, duration); err != nil {
			return err
		}
	}
	if err := b.Flush(); err != nil {
		return err
	}

	const resultSQL = `
		INSERT INTO hex_blitz_match_results (
			id, match_id, room_id, room_code, room_title, user_id, player_name, score, rank, session_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET score = EXCLUDED.score, rank = EXCLUDED.rank
	`
	for _, item := range results {
		if err := b.Queue(resultSQL, item.ID, item.MatchID, item.RoomID, item.RoomCode, item.RoomTitle, item.UserID, item.PlayerName, item.Score, item.Rank, item.SessionID, item.CreatedAt); err != nil {
			return err
		}
	}
	if err := b.Flush(); err != nil {
		return err
	}

	const moveSQL = `
		INSERT INTO hex_blitz_move_events (
			id, match_id, session_id, user_id, player_name, tile_id, move_index,
			cleared_count, gained_score, score_after, combo_after, occurred_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET score_after = EXCLUDED.score_after, combo_after = EXCLUDED.combo_after
	`
	for _, item := range moves {
		if err := b.Queue(moveSQL, item.ID, item.MatchID, item.SessionID, item.UserID, item.PlayerName, item.TileID, item.MoveIndex, item.ClearedCount, item.GainedScore, item.ScoreAfter, item.ComboAfter, item.OccurredAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkDoudizhuMatches(b *bulkBatcher, matches []bulkDoudizhuMatch, players []bulkDoudizhuPlayer, actions []bulkDoudizhuAction) error {
	const matchSQL = `
		INSERT INTO doudizhu_matches (
			id, room_id, room_code, room_title, match_mode, started_at, finished_at,
			landlord_seat, winner_side, multiplier, bomb_count, spring, anti_spring, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$7)
		ON CONFLICT (id) DO UPDATE SET
			finished_at = EXCLUDED.finished_at,
			winner_side = EXCLUDED.winner_side,
			multiplier = EXCLUDED.multiplier,
			bomb_count = EXCLUDED.bomb_count,
			spring = EXCLUDED.spring,
			anti_spring = EXCLUDED.anti_spring
	`
	for _, item := range matches {
		if err := b.Queue(matchSQL, item.ID, item.RoomID, item.RoomCode, item.RoomTitle, item.MatchMode, item.StartedAt, item.FinishedAt, item.LandlordSeat, item.WinnerSide, item.Multiplier, item.BombCount, item.Spring, item.AntiSpring); err != nil {
			return err
		}
	}
	if err := b.Flush(); err != nil {
		return err
	}

	const playerSQL = `
		INSERT INTO doudizhu_match_players (
			id, match_id, session_id, user_id, is_bot, bot_level, seat, player_name,
			role, bid_score, cards_left, is_winner, score_delta, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (id) DO UPDATE SET
			bid_score = EXCLUDED.bid_score,
			cards_left = EXCLUDED.cards_left,
			is_winner = EXCLUDED.is_winner,
			score_delta = EXCLUDED.score_delta
	`
	for _, item := range players {
		if err := b.Queue(playerSQL, item.ID, item.MatchID, item.SessionID, item.UserID, item.IsBot, item.BotLevel, item.Seat, item.PlayerName, item.Role, item.BidScore, item.CardsLeft, item.IsWinner, item.ScoreDelta, item.CreatedAt); err != nil {
			return err
		}
	}
	if err := b.Flush(); err != nil {
		return err
	}

	const actionSQL = `
		INSERT INTO doudizhu_action_events (
			id, match_id, turn_no, action_index, session_id, user_id, player_name, seat,
			action_type, cards_json, combo_type, combo_main_rank, combo_sequence_length,
			combo_total_cards, multiplier_after, occurred_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET multiplier_after = EXCLUDED.multiplier_after
	`
	for _, item := range actions {
		if err := b.Queue(actionSQL, item.ID, item.MatchID, item.TurnNo, item.ActionIndex, item.SessionID, item.UserID, item.PlayerName, item.Seat, item.ActionType, mustJSON(item.CardsJSON), item.ComboType, item.ComboMainRank, item.ComboSequenceLength, item.ComboTotalCards, item.MultiplierAfter, item.OccurredAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func insertBulkUserAchievements(b *bulkBatcher, items [][3]any) error {
	const sql = `
		INSERT INTO user_achievements (user_id, achievement_id, obtained_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id, achievement_id) DO NOTHING
	`
	for _, item := range items {
		if err := b.Queue(sql, item[0], item[1], item[2]); err != nil {
			return err
		}
	}
	return b.Flush()
}

func resetNamespacePointTransactions(ctx context.Context, tx pgx.Tx, namespace string) error {
	_, err := tx.Exec(ctx, `DELETE FROM point_transactions WHERE note LIKE $1`, "[bulk:"+namespace+"]%")
	return err
}

func insertBulkPointTransactions(b *bulkBatcher, items []struct {
	UserID    uuid.UUID
	Amount    int
	Source    string
	RefID     *string
	Note      string
	CreatedAt time.Time
}) error {
	const sql = `
		INSERT INTO point_transactions (user_id, amount, source, ref_id, note, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`
	for _, item := range items {
		if err := b.Queue(sql, item.UserID, item.Amount, item.Source, item.RefID, item.Note, item.CreatedAt); err != nil {
			return err
		}
	}
	return b.Flush()
}

func upsertBulkSponsorSettings(b *bulkBatcher, updatedBy uuid.UUID, namespace string, now time.Time) error {
	const sql = `
		INSERT INTO sponsor_settings (
			id, monthly_goal, current_raised, alipay_qr_url, wechat_qr_url, message, updated_at, updated_by
		)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			monthly_goal = EXCLUDED.monthly_goal,
			current_raised = EXCLUDED.current_raised,
			alipay_qr_url = EXCLUDED.alipay_qr_url,
			wechat_qr_url = EXCLUDED.wechat_qr_url,
			message = EXCLUDED.message,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`
	return b.Queue(
		sql,
		12000.0,
		5400.0,
		fmt.Sprintf("https://seed.local/%s/alipay.png", namespace),
		fmt.Sprintf("https://seed.local/%s/wechat.png", namespace),
		fmt.Sprintf("当前赞助展示为虚拟数据样本，命名空间：%s。", namespace),
		now,
		updatedBy,
	)
}

func upsertBulkAssistantSettings(b *bulkBatcher, updatedBy uuid.UUID, now time.Time) error {
	const sql = `
		INSERT INTO assistant_settings (
			id, enabled, persona_name, system_prompt, max_context_items, include_pages,
			include_posts, include_users, include_tags, include_groups, include_events, updated_at, updated_by
		)
		VALUES (1,TRUE,'霜牙','你是站内测试环境的虚拟助手。',8,TRUE,TRUE,TRUE,TRUE,TRUE,TRUE,$1,$2)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			persona_name = EXCLUDED.persona_name,
			system_prompt = EXCLUDED.system_prompt,
			max_context_items = EXCLUDED.max_context_items,
			include_pages = EXCLUDED.include_pages,
			include_posts = EXCLUDED.include_posts,
			include_users = EXCLUDED.include_users,
			include_tags = EXCLUDED.include_tags,
			include_groups = EXCLUDED.include_groups,
			include_events = EXCLUDED.include_events,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`
	return b.Queue(sql, now, updatedBy)
}

func reconcileBulkCounters(ctx context.Context, tx pgx.Tx, users []bulkUser) error {
	queries := []string{
		`UPDATE groups g SET member_count = COALESCE(sub.cnt, 0), updated_at = NOW()
		  FROM (SELECT group_id, COUNT(*) AS cnt FROM group_members GROUP BY group_id) sub
		  WHERE g.id = sub.group_id`,
		`UPDATE groups SET member_count = 0 WHERE id NOT IN (SELECT DISTINCT group_id FROM group_members)`,
		`UPDATE groups g SET post_count = COALESCE(sub.cnt, 0), updated_at = NOW()
		  FROM (SELECT group_id, COUNT(*) AS cnt FROM posts WHERE deleted_at IS NULL AND group_id IS NOT NULL GROUP BY group_id) sub
		  WHERE g.id = sub.group_id`,
		`UPDATE groups SET post_count = 0 WHERE id NOT IN (SELECT DISTINCT group_id FROM posts WHERE group_id IS NOT NULL AND deleted_at IS NULL)`,
		`UPDATE events e SET attendee_count = COALESCE(sub.cnt, 0), updated_at = NOW()
		  FROM (SELECT event_id, COUNT(*) AS cnt FROM event_attendees WHERE status = 'attending' GROUP BY event_id) sub
		  WHERE e.id = sub.event_id`,
		`UPDATE events SET attendee_count = 0 WHERE id NOT IN (SELECT DISTINCT event_id FROM event_attendees WHERE status = 'attending')`,
		`UPDATE posts p SET like_count = COALESCE(sub.cnt, 0)
		  FROM (SELECT post_id, COUNT(*) AS cnt FROM post_likes GROUP BY post_id) sub
		  WHERE p.id = sub.post_id`,
		`UPDATE posts SET like_count = 0 WHERE id NOT IN (SELECT DISTINCT post_id FROM post_likes)`,
		`UPDATE posts p SET comment_count = COALESCE(sub.cnt, 0)
		  FROM (SELECT commentable_id, COUNT(*) AS cnt FROM comments WHERE is_deleted = FALSE AND commentable_type = 'post' GROUP BY commentable_id) sub
		  WHERE p.id = sub.commentable_id`,
		`UPDATE posts SET comment_count = 0 WHERE id NOT IN (SELECT DISTINCT commentable_id FROM comments WHERE commentable_type = 'post' AND is_deleted = FALSE)`,
		`UPDATE comments c SET like_count = COALESCE(sub.cnt, 0)
		  FROM (SELECT comment_id, COUNT(*) AS cnt FROM comment_likes GROUP BY comment_id) sub
		  WHERE c.id = sub.comment_id`,
		`UPDATE comments SET like_count = 0 WHERE id NOT IN (SELECT DISTINCT comment_id FROM comment_likes)`,
		`UPDATE comments c SET reply_count = COALESCE(sub.cnt, 0)
		  FROM (SELECT parent_id, COUNT(*) AS cnt FROM comments WHERE parent_id IS NOT NULL AND is_deleted = FALSE GROUP BY parent_id) sub
		  WHERE c.id = sub.parent_id`,
		`UPDATE comments SET reply_count = 0 WHERE id NOT IN (SELECT DISTINCT parent_id FROM comments WHERE parent_id IS NOT NULL AND is_deleted = FALSE)`,
		`UPDATE audio_works w SET like_count = COALESCE(sub.cnt, 0)
		  FROM (SELECT work_id, COUNT(*) AS cnt FROM audio_work_likes GROUP BY work_id) sub
		  WHERE w.id = sub.work_id`,
		`UPDATE audio_works SET like_count = 0 WHERE id NOT IN (SELECT DISTINCT work_id FROM audio_work_likes)`,
		`UPDATE conversations c SET updated_at = COALESCE(sub.max_created_at, c.updated_at)
		  FROM (SELECT conversation_id, MAX(created_at) AS max_created_at FROM messages GROUP BY conversation_id) sub
		  WHERE c.id = sub.conversation_id`,
	}
	for _, query := range queries {
		if _, err := tx.Exec(ctx, query); err != nil {
			return err
		}
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_points (user_id, balance, total_earned, updated_at)
		SELECT
			user_id,
			COALESCE(SUM(amount), 0) AS balance,
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) AS total_earned,
			NOW()
		FROM point_transactions
		WHERE user_id = ANY($1)
		GROUP BY user_id
		ON CONFLICT (user_id) DO UPDATE SET
			balance = EXCLUDED.balance,
			total_earned = EXCLUDED.total_earned,
			updated_at = EXCLUDED.updated_at
	`, userIDs); err != nil {
		return err
	}

	return nil
}

func refreshDailyMetrics(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `REFRESH MATERIALIZED VIEW daily_metrics`)
	return err
}

func randomPastTime(rng *rand.Rand, now time.Time, days int) time.Time {
	if days <= 0 {
		days = 30
	}
	hours := rng.Intn(days*24 + 1)
	minutes := rng.Intn(60)
	return now.Add(-time.Duration(hours)*time.Hour - time.Duration(minutes)*time.Minute)
}

func randomEventStart(rng *rand.Rand, now time.Time, status string) time.Time {
	switch status {
	case "completed":
		return now.Add(-time.Duration(24+rng.Intn(360)) * time.Hour)
	case "draft":
		return now.Add(time.Duration(24+rng.Intn(240)) * time.Hour)
	case "cancelled":
		return now.Add(-time.Duration(12+rng.Intn(200)) * time.Hour)
	default:
		return now.Add(time.Duration(12+rng.Intn(240)) * time.Hour)
	}
}

func bulkUsername(namespace string, index int) string {
	prefix := namespace
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	return fmt.Sprintf("%s%04d", prefix, index+1)
}

func randomSpecies(rng *rand.Rand) string {
	items := []string{"狐狸", "狼", "雪豹", "猞猁", "犬科", "猫科", "鹿", "龙", "兔", "熊"}
	return items[rng.Intn(len(items))]
}

func randomFurryName(rng *rand.Rand) string {
	prefix := []string{"霜", "星", "月", "岚", "木", "灰", "影", "森", "砂", "曜"}
	suffix := []string{"尾", "爪", "牙", "耳", "羽", "角", "蹄", "鳞", "鬃", "翼"}
	return prefix[rng.Intn(len(prefix))] + suffix[rng.Intn(len(suffix))]
}

func randomCity(rng *rand.Rand) string {
	items := []string{"上海", "杭州", "南京", "深圳", "成都", "武汉", "重庆", "广州", "苏州", "北京"}
	return items[rng.Intn(len(items))]
}

func randomLocation(rng *rand.Rand) string {
	items := []string{"Discord", "线上直播间", "上海徐汇", "杭州拱墅", "南京玄武湖", "深圳南山", "成都天府广场"}
	return items[rng.Intn(len(items))]
}

func randomGroupTopic(rng *rand.Rand) string {
	items := []string{"兽设", "摄影", "像素创作", "活动筹备", "毛绒日常", "世界观脑暴", "配色试验", "夜聊茶会"}
	return items[rng.Intn(len(items))]
}

func randomEventTopic(rng *rand.Rand) string {
	items := []string{"周末毛绒聚会", "线上灵感夜聊", "摄影散步", "设定分享会", "音频试玩沙龙", "轻桌游联谊"}
	return items[rng.Intn(len(items))]
}

func randomPostTitle(rng *rand.Rand) string {
	items := []string{"新稿公开", "活动招募", "灵感记录", "设定整理", "近况分享", "测试帖", "幕后日志", "配色草案"}
	return items[rng.Intn(len(items))]
}

func randomAudioTitle(rng *rand.Rand) string {
	items := []string{"夜航回声", "薄雾尾音", "林地漫游", "霓虹步点", "暖风试音", "城市回响"}
	return items[rng.Intn(len(items))]
}

func randomAlbumTitle(rng *rand.Rand) string {
	items := []string{"森林夜行", "霓虹巡礼", "海湾碎片", "山城回声", "月光跑图", "站内精选"}
	return items[rng.Intn(len(items))]
}

func randomTrackTitle(rng *rand.Rand) string {
	items := []string{"序章", "慢步", "回望", "飞行轨迹", "夜色采样", "灯下", "风的尾迹", "收束"}
	return items[rng.Intn(len(items))]
}

func randomComposer(rng *rand.Rand) string {
	items := []string{"霜牙", "银尾", "像素猞猁", "苔爪", "晨光", "晚星"}
	return items[rng.Intn(len(items))]
}

func randomIntent(rng *rand.Rand) string {
	items := []string{"discover", "community", "creation", "recommendation", "support", "event"}
	return items[rng.Intn(len(items))]
}

func randomMood(rng *rand.Rand) string {
	items := []string{"暖色", "安静", "雨夜", "低饱和", "轻电子", "木吉他", "像素感", "治愈"}
	return items[rng.Intn(len(items))]
}

func randomCardTitle(rng *rand.Rand) string {
	items := []string{"建议入口", "关联圈子", "相似帖子", "推荐活动", "玩法摘要"}
	return items[rng.Intn(len(items))]
}

func randomTags(rng *rand.Rand, n int) []string {
	candidates := []string{"创作", "兽设", "AI", "线下", "活动", "摄影", "音频", "社区", "像素", "毛绒", "讨论", "灵感"}
	if n <= 0 {
		n = 1
	}
	if n > len(candidates) {
		n = len(candidates)
	}
	perm := rng.Perm(len(candidates))
	result := make([]string, 0, n)
	for _, idx := range perm[:n] {
		result = append(result, candidates[idx])
	}
	return result
}

func sentence(rng *rand.Rand, words int) string {
	fragments := []string{
		"这是一条用于丰富数据库的虚拟内容样本",
		"主要用来覆盖后台列表、筛选、搜索和统计视图",
		"不会触发真实业务含义",
		"但会尽量模拟社区平台里常见的表达方式",
		"方便你在联调时看到更接近真实运营环境的数据密度",
		"也适合作为截图、演示和性能压测时的基础内容",
	}
	if words < 4 {
		words = 4
	}
	parts := make([]string, 0, words/2+1)
	for len(parts) < words/2 {
		parts = append(parts, fragments[rng.Intn(len(fragments))])
	}
	return strings.Join(parts, "，") + "。"
}

func paragraph(rng *rand.Rand, sentences int) string {
	if sentences < 1 {
		sentences = 1
	}
	parts := make([]string, 0, sentences)
	for i := 0; i < sentences; i++ {
		parts = append(parts, sentence(rng, 12+rng.Intn(6)))
	}
	return strings.Join(parts, "\n\n")
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func stringPtr(value string) *string {
	return &value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
