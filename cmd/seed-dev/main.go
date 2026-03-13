package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/studio/platform/configs"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	"github.com/studio/platform/internal/pkg/crypto"
)

type userSeed struct {
	Username string
	Email    string
	Role     string
	Bio      string
	Location string
	Website  string
	Furry    string
	Species  string
}

type groupSeed struct {
	ID          string
	OwnerEmail  string
	Name        string
	Description string
	Tags        []string
}

type eventSeed struct {
	ID             string
	OrganizerEmail string
	Title          string
	Description    string
	Location       string
	IsOnline       bool
	StartOffsetH   int
	EndOffsetH     int
	MaxCapacity    int
	Tags           []string
}

type postSeed struct {
	ID          string
	AuthorEmail string
	Title       string
	Content     string
	Tags        []string
	Labels      map[string]bool
}

func main() {
	configFile := flag.String("config", "configs/config.local.yaml", "path to config file")
	flag.Parse()

	cfg, err := configs.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := postgresinfra.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now().UTC()
	passwordHash, err := crypto.HashPassword("Passw0rd123")
	if err != nil {
		log.Fatalf("failed to hash seed password: %v", err)
	}

	users := []userSeed{
		{
			Username: "frostfang",
			Email:    "frostfang@furry.local",
			Role:     "admin",
			Bio:      "站内管理员，也会分享兽设与社区动态。",
			Location: "上海",
			Website:  "https://furry.local/admin",
			Furry:    "霜牙",
			Species:  "雪豹",
		},
		{
			Username: "silvertail",
			Email:    "silvertail@furry.local",
			Role:     "creator",
			Bio:      "主做插画和兽设委托，偏暖色系毛绒风。",
			Location: "杭州",
			Website:  "https://furry.local/creator/silvertail",
			Furry:    "银尾",
			Species:  "狐狸",
		},
		{
			Username: "mosspaws",
			Email:    "mosspaws@furry.local",
			Role:     "member",
			Bio:      "喜欢线下聚会、摄影和活动组织。",
			Location: "南京",
			Website:  "",
			Furry:    "苔爪",
			Species:  "狼犬",
		},
		{
			Username: "pixellynx",
			Email:    "pixellynx@furry.local",
			Role:     "member",
			Bio:      "主玩像素风创作和 AI 辅助灵感整理。",
			Location: "深圳",
			Website:  "",
			Furry:    "像素猞猁",
			Species:  "猞猁",
		},
	}

	userIDs := make(map[string]uuid.UUID, len(users))
	for _, item := range users {
		id, err := upsertUser(ctx, tx, item, passwordHash, now)
		if err != nil {
			log.Fatalf("failed to upsert user %s: %v", item.Email, err)
		}
		userIDs[item.Email] = id
	}

	follows := [][2]string{
		{"mosspaws@furry.local", "silvertail@furry.local"},
		{"pixellynx@furry.local", "silvertail@furry.local"},
		{"pixellynx@furry.local", "frostfang@furry.local"},
	}
	for _, item := range follows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_follows (follower_id, followee_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (follower_id, followee_id) DO NOTHING
		`, userIDs[item[0]], userIDs[item[1]], now); err != nil {
			log.Fatalf("failed to insert follow: %v", err)
		}
	}

	groups := []groupSeed{
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222a101",
			OwnerEmail:  "silvertail@furry.local",
			Name:        "兽设灵感交流局",
			Description: "分享角色设定、配色灵感和参考图，欢迎晒自己的 OC。",
			Tags:        []string{"兽设", "创作", "灵感"},
		},
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222a102",
			OwnerEmail:  "mosspaws@furry.local",
			Name:        "线下同好活动筹备所",
			Description: "发布聚会、桌游、摄影和漫展约拍相关活动信息。",
			Tags:        []string{"线下", "活动", "摄影"},
		},
	}
	for _, item := range groups {
		if err := upsertGroup(ctx, tx, item, userIDs[item.OwnerEmail], now); err != nil {
			log.Fatalf("failed to upsert group %s: %v", item.Name, err)
		}
	}

	groupMembers := []struct {
		GroupID string
		Email   string
		Role    string
	}{
		{"ae8bcb6d-575f-4e55-b553-5f2e6222a101", "silvertail@furry.local", "owner"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222a101", "pixellynx@furry.local", "member"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222a102", "mosspaws@furry.local", "owner"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222a102", "frostfang@furry.local", "member"},
	}
	for _, item := range groupMembers {
		groupID := uuid.MustParse(item.GroupID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO group_members (group_id, user_id, role, joined_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (group_id, user_id) DO UPDATE SET role = EXCLUDED.role
		`, groupID, userIDs[item.Email], item.Role, now); err != nil {
			log.Fatalf("failed to insert group member: %v", err)
		}
	}
	for _, item := range groups {
		groupID := uuid.MustParse(item.ID)
		if _, err := tx.Exec(ctx, `
			UPDATE groups
			SET member_count = (SELECT COUNT(*) FROM group_members WHERE group_id = $1),
			    updated_at = $2
			WHERE id = $1
		`, groupID, now); err != nil {
			log.Fatalf("failed to update group member count: %v", err)
		}
	}

	events := []eventSeed{
		{
			ID:             "ae8bcb6d-575f-4e55-b553-5f2e6222b201",
			OrganizerEmail: "mosspaws@furry.local",
			Title:          "周末毛绒摄影散步",
			Description:    "一起带上毛绒装备和相机，在公园散步拍照，结束后可自由聚餐。",
			Location:       "南京玄武湖",
			IsOnline:       false,
			StartOffsetH:   36,
			EndOffsetH:     40,
			MaxCapacity:    20,
			Tags:           []string{"摄影", "线下", "聚会"},
		},
		{
			ID:             "ae8bcb6d-575f-4e55-b553-5f2e6222b202",
			OrganizerEmail: "silvertail@furry.local",
			Title:          "AI 辅助兽设灵感夜聊",
			Description:    "线上语音分享如何把 AI 用在设定整理、配色尝试和世界观脑暴里。",
			Location:       "Discord",
			IsOnline:       true,
			StartOffsetH:   72,
			EndOffsetH:     74,
			MaxCapacity:    50,
			Tags:           []string{"AI", "兽设", "线上"},
		},
	}
	for _, item := range events {
		if err := upsertEvent(ctx, tx, item, userIDs[item.OrganizerEmail], now); err != nil {
			log.Fatalf("failed to upsert event %s: %v", item.Title, err)
		}
	}

	eventAttendees := []struct {
		EventID string
		Email   string
		Status  string
	}{
		{"ae8bcb6d-575f-4e55-b553-5f2e6222b201", "mosspaws@furry.local", "attending"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222b201", "pixellynx@furry.local", "attending"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222b202", "silvertail@furry.local", "attending"},
		{"ae8bcb6d-575f-4e55-b553-5f2e6222b202", "frostfang@furry.local", "maybe"},
	}
	for _, item := range eventAttendees {
		eventID := uuid.MustParse(item.EventID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_attendees (event_id, user_id, status, joined_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (event_id, user_id) DO UPDATE SET status = EXCLUDED.status
		`, eventID, userIDs[item.Email], item.Status, now); err != nil {
			log.Fatalf("failed to insert event attendee: %v", err)
		}
	}
	for _, item := range events {
		eventID := uuid.MustParse(item.ID)
		if _, err := tx.Exec(ctx, `
			UPDATE events
			SET attendee_count = (
				SELECT COUNT(*) FROM event_attendees WHERE event_id = $1 AND status = 'attending'
			),
			    updated_at = $2
			WHERE id = $1
		`, eventID, now); err != nil {
			log.Fatalf("failed to update event attendee count: %v", err)
		}
	}

	posts := []postSeed{
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222c301",
			AuthorEmail: "silvertail@furry.local",
			Title:       "新委托样图：暖色系狐狸设定",
			Content:     "这次整理了一版偏暖橙和奶油白的狐狸兽设，想测试下尾巴的体积和围巾配色，欢迎给意见。",
			Tags:        []string{"兽设", "绘画", "创作"},
			Labels:      map[string]bool{"is_ai_generated": false},
		},
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222c302",
			AuthorEmail: "mosspaws@furry.local",
			Title:       "周末摄影散步招募中",
			Content:     "想在南京约一次轻松的毛绒摄影散步，路线已经踩点好了，欢迎带朋友一起。",
			Tags:        []string{"摄影", "活动", "线下"},
			Labels:      map[string]bool{"is_ai_generated": false},
		},
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222c303",
			AuthorEmail: "pixellynx@furry.local",
			Title:       "用 AI 整理世界观设定的小技巧",
			Content:     "我最近会先让 AI 帮我拆角色关系和设定空缺，再回到手工细化，效率确实高了很多。",
			Tags:        []string{"AI", "设定", "创作"},
			Labels:      map[string]bool{"is_ai_generated": true},
		},
		{
			ID:          "ae8bcb6d-575f-4e55-b553-5f2e6222c304",
			AuthorEmail: "frostfang@furry.local",
			Title:       "站内新助手已上线测试",
			Content:     "右下角的霜牙已经可以帮大家找帖子、圈子和活动了，如果有奇怪回答欢迎继续反馈。",
			Tags:        []string{"公告", "AI", "社区"},
			Labels:      map[string]bool{"is_ai_generated": false},
		},
	}
	for _, item := range posts {
		if err := upsertPost(ctx, tx, item, userIDs[item.AuthorEmail], now); err != nil {
			log.Fatalf("failed to upsert post %s: %v", item.Title, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit seed transaction: %v", err)
	}

	fmt.Println("seeded development data successfully")
	fmt.Println("sample accounts:")
	for _, item := range users {
		fmt.Printf("- %s / Passw0rd123 (%s)\n", item.Email, item.Role)
	}
}

func upsertUser(ctx context.Context, tx pgx.Tx, seed userSeed, passwordHash string, now time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	var website *string
	if seed.Website != "" {
		website = &seed.Website
	}
	var bio *string
	if seed.Bio != "" {
		bio = &seed.Bio
	}
	var location *string
	if seed.Location != "" {
		location = &seed.Location
	}
	var furry *string
	if seed.Furry != "" {
		furry = &seed.Furry
	}
	var species *string
	if seed.Species != "" {
		species = &seed.Species
	}

	err := tx.QueryRow(ctx, `
		INSERT INTO users (
			id, username, email, password_hash, bio, location, website, furry_name, species,
			role, status, email_verified_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11, $12, $12)
		ON CONFLICT (email) DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			bio = EXCLUDED.bio,
			location = EXCLUDED.location,
			website = EXCLUDED.website,
			furry_name = EXCLUDED.furry_name,
			species = EXCLUDED.species,
			role = EXCLUDED.role,
			status = 'active',
			email_verified_at = EXCLUDED.email_verified_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`,
		uuid.New(),
		seed.Username,
		seed.Email,
		passwordHash,
		bio,
		location,
		website,
		furry,
		species,
		seed.Role,
		now,
		now,
	).Scan(&id)
	return id, err
}

func upsertGroup(ctx context.Context, tx pgx.Tx, seed groupSeed, ownerID uuid.UUID, now time.Time) error {
	tagsJSON, _ := json.Marshal(seed.Tags)
	_, err := tx.Exec(ctx, `
		INSERT INTO groups (id, owner_id, name, description, tags, privacy, member_count, post_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'public', 0, 0, $6, $6)
		ON CONFLICT (id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			tags = EXCLUDED.tags,
			privacy = EXCLUDED.privacy,
			updated_at = EXCLUDED.updated_at
	`, uuid.MustParse(seed.ID), ownerID, seed.Name, seed.Description, tagsJSON, now)
	return err
}

func upsertEvent(ctx context.Context, tx pgx.Tx, seed eventSeed, organizerID uuid.UUID, now time.Time) error {
	tagsJSON, _ := json.Marshal(seed.Tags)
	startTime := now.Add(time.Duration(seed.StartOffsetH) * time.Hour)
	endTime := now.Add(time.Duration(seed.EndOffsetH) * time.Hour)
	_, err := tx.Exec(ctx, `
		INSERT INTO events (
			id, organizer_id, title, description, location, is_online,
			start_time, end_time, max_capacity, tags, status, attendee_count, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'published', 0, $11, $11)
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
	`,
		uuid.MustParse(seed.ID),
		organizerID,
		seed.Title,
		seed.Description,
		seed.Location,
		seed.IsOnline,
		startTime,
		endTime,
		seed.MaxCapacity,
		tagsJSON,
		now,
	)
	return err
}

func upsertPost(ctx context.Context, tx pgx.Tx, seed postSeed, authorID uuid.UUID, now time.Time) error {
	mediaJSON, _ := json.Marshal([]string{})
	labelsJSON, _ := json.Marshal(seed.Labels)
	_, err := tx.Exec(ctx, `
		INSERT INTO posts (
			id, author_id, title, content, media_urls, tags, visibility, moderation_status,
			content_labels, like_count, comment_count, is_pinned, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'public', 'approved', $7, $8, $9, false, $10, $10)
		ON CONFLICT (id) DO UPDATE SET
			author_id = EXCLUDED.author_id,
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			media_urls = EXCLUDED.media_urls,
			tags = EXCLUDED.tags,
			visibility = EXCLUDED.visibility,
			moderation_status = EXCLUDED.moderation_status,
			content_labels = EXCLUDED.content_labels,
			like_count = EXCLUDED.like_count,
			comment_count = EXCLUDED.comment_count,
			updated_at = EXCLUDED.updated_at
	`,
		uuid.MustParse(seed.ID),
		authorID,
		seed.Title,
		seed.Content,
		mediaJSON,
		seed.Tags,
		labelsJSON,
		3,
		1,
		now,
	)
	return err
}
