package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/infra/embedding"
	"github.com/studio/platform/internal/observability/assistantmetrics"
)

const (
	assistantKnowledgeChunkSize    = 320
	assistantKnowledgeChunkOverlap = 48
	assistantIndexPageSize         = 100
)

// AssistantOverview is returned by the admin API for retrieval/index visibility.
type AssistantOverview struct {
	Overview *assistantdomain.Overview `json:"overview"`
	Metrics  assistantmetrics.Snapshot `json:"metrics"`
}

// AssistantFeedbackInput is submitted by the UI after a response is rendered.
type AssistantFeedbackInput struct {
	ResponseID     string          `json:"response_id"`
	ConversationID string          `json:"conversation_id,omitempty"`
	Value          string          `json:"value"`
	Query          string          `json:"query"`
	ReplyExcerpt   string          `json:"reply_excerpt"`
	Provider       string          `json:"provider,omitempty"`
	Intent         string          `json:"intent,omitempty"`
	Fallback       bool            `json:"fallback"`
	PagePath       string          `json:"page_path,omitempty"`
	SourceCounts   map[string]int  `json:"source_counts,omitempty"`
	Cards          []AssistantCard `json:"cards,omitempty"`
}

type assistantRetrievalCandidate struct {
	doc          *assistantdomain.KnowledgeDocument
	keywordScore float64
	vectorScore  float64
	finalScore   float64
	reason       string
}

// KnowledgeSyncInterval returns the assistant knowledge refresh cadence.
func (s *AssistantService) KnowledgeSyncInterval() time.Duration {
	if s == nil || s.cfg.SyncIntervalSec <= 0 {
		return 10 * time.Minute
	}
	return time.Duration(s.cfg.SyncIntervalSec) * time.Second
}

// SyncKnowledgeIndex rebuilds the phase-1 retrieval corpus.
func (s *AssistantService) SyncKnowledgeIndex(ctx context.Context) error {
	if s == nil || !s.HistoryEnabled() {
		return nil
	}

	start := time.Now()
	totalIndexed := 0
	var syncErr error
	defer func() {
		assistantmetrics.RecordIndexSync(time.Since(start), totalIndexed, syncErr)
	}()

	indexedAt := time.Now()
	type sourceBatch struct {
		sourceType string
		builder    func(context.Context, time.Time) ([]*assistantdomain.KnowledgeDocument, error)
	}

	batches := []sourceBatch{
		{
			sourceType: assistantdomain.KnowledgeSourcePage,
			builder: func(_ context.Context, now time.Time) ([]*assistantdomain.KnowledgeDocument, error) {
				return s.buildPageKnowledgeDocuments(now)
			},
		},
		{
			sourceType: assistantdomain.KnowledgeSourcePost,
			builder:    s.buildPostKnowledgeDocuments,
		},
		{
			sourceType: assistantdomain.KnowledgeSourceGroup,
			builder:    s.buildGroupKnowledgeDocuments,
		},
		{
			sourceType: assistantdomain.KnowledgeSourceEvent,
			builder:    s.buildEventKnowledgeDocuments,
		},
	}

	for _, batch := range batches {
		docs, err := batch.builder(ctx, indexedAt)
		if err != nil {
			syncErr = err
			return err
		}
		if err := s.historyRepo.ReplaceKnowledgeDocuments(ctx, batch.sourceType, docs); err != nil {
			syncErr = err
			return err
		}
		totalIndexed += len(docs)
	}

	s.lastKnowledgeSyncNS.Store(indexedAt.UnixNano())
	return nil
}

// GetOverview returns indexed corpus, feedback stats and runtime assistant metrics.
func (s *AssistantService) GetOverview(ctx context.Context) (*AssistantOverview, error) {
	if !s.HistoryEnabled() {
		return &AssistantOverview{
			Overview: &assistantdomain.Overview{
				DocumentsBySource: make(map[string]int64),
				RetrievalLimit:    s.cfg.RetrievalLimit,
				VectorScanLimit:   s.cfg.VectorScanLimit,
				SyncIntervalSec:   s.cfg.SyncIntervalSec,
			},
			Metrics: assistantmetrics.GetSnapshot(),
		}, nil
	}

	overview, err := s.historyRepo.GetKnowledgeOverview(ctx)
	if err != nil {
		return nil, err
	}
	if overview.DocumentsBySource == nil {
		overview.DocumentsBySource = make(map[string]int64)
	}
	overview.EmbeddingConfigured = strings.TrimSpace(s.cfg.EmbeddingAPIKey) != ""
	overview.EmbeddingModel = strings.TrimSpace(s.cfg.EmbeddingModel)
	overview.VisionConfigured = strings.TrimSpace(s.cfg.VisionAPIKey) != ""
	overview.VisionModel = strings.TrimSpace(s.cfg.VisionModel)
	overview.RetrievalLimit = s.cfg.RetrievalLimit
	overview.VectorScanLimit = s.cfg.VectorScanLimit
	overview.SyncIntervalSec = s.cfg.SyncIntervalSec

	return &AssistantOverview{
		Overview: overview,
		Metrics:  assistantmetrics.GetSnapshot(),
	}, nil
}

// SubmitFeedback stores a helpful / unhelpful vote for a single assistant response.
func (s *AssistantService) SubmitFeedback(ctx context.Context, userID *uuid.UUID, input AssistantFeedbackInput) error {
	if !s.HistoryEnabled() {
		return nil
	}

	responseID, err := uuid.Parse(strings.TrimSpace(input.ResponseID))
	if err != nil {
		return fmt.Errorf("invalid response id: %w", err)
	}

	var conversationID *uuid.UUID
	if trimmed := strings.TrimSpace(input.ConversationID); trimmed != "" {
		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			return fmt.Errorf("invalid conversation id: %w", err)
		}
		conversationID = &parsed
	}

	value := assistantdomain.FeedbackValue(strings.TrimSpace(input.Value))
	switch value {
	case assistantdomain.FeedbackHelpful, assistantdomain.FeedbackUnhelpful:
	default:
		return fmt.Errorf("invalid feedback value")
	}

	item := &assistantdomain.Feedback{
		ID:             uuid.New(),
		ResponseID:     responseID,
		ConversationID: conversationID,
		UserID:         userID,
		Value:          value,
		QueryText:      truncateText(strings.TrimSpace(input.Query), 500),
		ReplyExcerpt:   truncateText(strings.TrimSpace(input.ReplyExcerpt), 1000),
		Provider:       truncateText(strings.TrimSpace(input.Provider), 64),
		Intent:         truncateText(strings.TrimSpace(input.Intent), 64),
		Fallback:       input.Fallback,
		PagePath:       truncateText(strings.TrimSpace(input.PagePath), 255),
		SourceCounts:   input.SourceCounts,
		Cards:          input.Cards,
		CreatedAt:      time.Now(),
	}
	if err := s.historyRepo.UpsertFeedback(ctx, item); err != nil {
		return err
	}
	assistantmetrics.RecordFeedback(string(value))
	return nil
}

func (s *AssistantService) collectKnowledgeCards(ctx context.Context, query string, intent assistantIntent, settings *assistantdomain.Settings) ([]AssistantCard, error) {
	start := time.Now()
	resultCount := 0
	defer func() {
		assistantmetrics.ObserveRetrieval(time.Since(start), resultCount)
	}()

	if err := s.ensureKnowledgeFresh(ctx); err != nil {
		return nil, err
	}

	sourceTypes := knowledgeSourceTypes(settings)
	if len(sourceTypes) == 0 {
		return nil, nil
	}

	retrievalLimit := max(settings.MaxContextItems*3, s.cfg.RetrievalLimit)
	keywordDocs, err := s.historyRepo.SearchKnowledgeDocuments(ctx, query, sourceTypes, retrievalLimit)
	if err != nil {
		return nil, err
	}

	scanLimit := max(retrievalLimit*20, s.cfg.VectorScanLimit)
	vectorDocs, err := s.historyRepo.ListKnowledgeDocumentsForScan(ctx, sourceTypes, scanLimit)
	if err != nil {
		return nil, err
	}

	queryVec, err := s.embedKnowledgeText(query)
	if err != nil {
		return nil, err
	}

	candidates := mergeKnowledgeCandidates(keywordDocs, vectorDocs, queryVec, intent)
	if len(candidates) == 0 {
		return nil, nil
	}

	maxItems := settings.MaxContextItems
	if maxItems <= 0 {
		maxItems = 6
	}

	cards := make([]AssistantCard, 0, maxItems)
	selectedByKind := make(map[string]int, len(candidates))
	for _, candidate := range candidates {
		if len(cards) >= maxItems {
			break
		}
		kindLimit := assistantFinalKindLimit(intent, candidate.doc.SourceType, maxItems)
		if kindLimit > 0 && selectedByKind[candidate.doc.SourceType] >= kindLimit {
			continue
		}
		cards = append(cards, AssistantCard{
			Kind:    candidate.doc.SourceType,
			Title:   candidate.doc.Title,
			Summary: candidate.doc.Summary,
			Href:    candidate.doc.Href,
			Meta:    candidate.doc.Meta,
			Reason:  candidate.reason,
			Source:  firstNonEmpty(candidate.doc.SourceLabel, "知识索引"),
		})
		selectedByKind[candidate.doc.SourceType]++
	}

	resultCount = len(cards)
	return cards, nil
}

func (s *AssistantService) ensureKnowledgeFresh(ctx context.Context) error {
	if !s.HistoryEnabled() {
		return nil
	}
	lastSyncNS := s.lastKnowledgeSyncNS.Load()
	if lastSyncNS > 0 && time.Since(time.Unix(0, lastSyncNS)) < s.KnowledgeSyncInterval() {
		return nil
	}

	_, err := s.syncGroup.DoWithContext(ctx, "assistant_knowledge_sync", func(ctx context.Context) (any, error) {
		lastSyncNS := s.lastKnowledgeSyncNS.Load()
		if lastSyncNS > 0 && time.Since(time.Unix(0, lastSyncNS)) < s.KnowledgeSyncInterval() {
			return nil, nil
		}
		return nil, s.SyncKnowledgeIndex(ctx)
	})
	return err
}

func (s *AssistantService) buildPageKnowledgeDocuments(now time.Time) ([]*assistantdomain.KnowledgeDocument, error) {
	type pageDoc struct {
		key     string
		title   string
		summary string
		href    string
		meta    string
		content string
		tags    []string
	}

	docsConfig := []pageDoc{
		{
			key:     "home",
			title:   "首页",
			summary: "浏览社区热门内容和平台特色入口。",
			href:    "/",
			meta:    "/",
			content: "首页适合第一次进入社区时快速了解站内热门动态、社区亮点和主要入口。",
			tags:    []string{"首页", "导航", "新手"},
		},
		{
			key:     "explore",
			title:   "发现页",
			summary: "按热度和标签探索热门动态与创作者。",
			href:    "/explore",
			meta:    "/explore",
			content: "发现页适合新用户浏览热门帖子、标签和创作者，快速熟悉社区内容风格。",
			tags:    []string{"发现", "热门", "推荐"},
		},
		{
			key:     "feed",
			title:   "关注动态",
			summary: "查看已关注对象的最新动态和互动。",
			href:    "/feed",
			meta:    "/feed",
			content: "关注动态会展示你已关注用户的最新内容，适合进行个性化浏览。",
			tags:    []string{"关注", "时间线"},
		},
		{
			key:     "search",
			title:   "搜索页",
			summary: "搜索帖子、用户和标签，快速定位站内内容。",
			href:    "/search",
			meta:    "/search",
			content: "搜索页支持按关键词查找帖子和用户，也适合找特定主题。",
			tags:    []string{"搜索", "帖子", "用户"},
		},
		{
			key:     "groups",
			title:   "圈子广场",
			summary: "按兴趣发现圈子，查看成员规模和帖子活跃度。",
			href:    "/groups",
			meta:    "/groups",
			content: "圈子广场适合按兴趣找同好，加入公开圈子并查看圈内规则、公告和内容。",
			tags:    []string{"圈子", "社群", "兴趣"},
		},
		{
			key:     "events",
			title:   "活动广场",
			summary: "查看近期线上线下活动并报名参加。",
			href:    "/events",
			meta:    "/events",
			content: "活动广场适合查看近期线上线下活动，了解活动时间、地点、人数和报名情况。",
			tags:    []string{"活动", "报名", "聚会"},
		},
		{
			key:     "create-post",
			title:   "发布动态",
			summary: "撰写帖子、上传图片、设置标签和可见性。",
			href:    "/posts/create",
			meta:    "/posts/create",
			content: "发布动态页面支持标题、正文、图片上传、标签推荐、AI 内容标记和可见性设置。",
			tags:    []string{"发帖", "创作", "标签"},
		},
		{
			key:     "creator",
			title:   "创作者面板",
			summary: "查看帖子、互动、粉丝和打赏数据。",
			href:    "/creator",
			meta:    "/creator",
			content: "创作者面板适合持续创作的用户查看帖子、粉丝、互动和打赏统计。",
			tags:    []string{"创作者", "数据", "打赏"},
		},
		{
			key:     "notifications",
			title:   "通知中心",
			summary: "查看点赞、评论、关注和打赏通知。",
			href:    "/notifications",
			meta:    "/notifications",
			content: "通知中心会汇总点赞、评论、关注和打赏等提醒，帮助用户及时回看互动。",
			tags:    []string{"通知", "互动"},
		},
		{
			key:     "reports",
			title:   "我的举报",
			summary: "查看自己提交的举报记录和处理状态。",
			href:    "/reports",
			meta:    "/reports",
			content: "我的举报页面展示用户提交过的举报记录，适合追踪处理进展。",
			tags:    []string{"举报", "治理"},
		},
	}

	docs := make([]*assistantdomain.KnowledgeDocument, 0, len(docsConfig))
	for _, item := range docsConfig {
		chunks, err := s.makeKnowledgeDocuments(
			assistantdomain.KnowledgeSourcePage,
			item.key,
			item.title,
			item.summary,
			item.content,
			item.href,
			item.meta,
			"站内固定导航",
			item.tags,
			now,
			now,
		)
		if err != nil {
			return nil, err
		}
		docs = append(docs, chunks...)
	}
	return docs, nil
}

func (s *AssistantService) buildPostKnowledgeDocuments(ctx context.Context, now time.Time) ([]*assistantdomain.KnowledgeDocument, error) {
	if s.postService == nil {
		return nil, nil
	}

	docs := make([]*assistantdomain.KnowledgeDocument, 0, 256)
	for page := 1; ; page++ {
		posts, _, err := s.postService.ListExplore(ctx, page, assistantIndexPageSize, "")
		if err != nil {
			return nil, err
		}
		if len(posts) == 0 {
			break
		}

		for _, item := range posts {
			if item == nil {
				continue
			}
			title := strings.TrimSpace(item.Title)
			if title == "" {
				title = truncateText(item.Content, 28)
			}
			meta := fmt.Sprintf("@%s · %d 赞 · %d 评论", item.AuthorUsername, item.LikeCount, item.CommentCount)
			if item.GroupName != nil && strings.TrimSpace(*item.GroupName) != "" {
				meta += " · 圈子：" + strings.TrimSpace(*item.GroupName)
			}
			content := strings.Join([]string{
				title,
				item.Content,
				"标签：" + strings.Join(item.Tags, "、"),
			}, "\n")
			chunks, err := s.makeKnowledgeDocuments(
				assistantdomain.KnowledgeSourcePost,
				item.ID.String(),
				title,
				truncateText(item.Content, 72),
				content,
				"/posts/"+item.ID.String(),
				meta,
				"帖子详情页",
				item.Tags,
				now,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			docs = append(docs, chunks...)
		}

		if len(posts) < assistantIndexPageSize {
			break
		}
	}
	return docs, nil
}

func (s *AssistantService) buildGroupKnowledgeDocuments(ctx context.Context, now time.Time) ([]*assistantdomain.KnowledgeDocument, error) {
	if s.groupService == nil {
		return nil, nil
	}

	privacy := group.GroupPrivacyPublic
	docs := make([]*assistantdomain.KnowledgeDocument, 0, 128)
	for page := 1; ; page++ {
		items, _, err := s.groupService.ListGroups(ctx, ListGroupsInput{
			Privacy:  &privacy,
			Page:     page,
			PageSize: assistantIndexPageSize,
		})
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			meta := fmt.Sprintf("%d 成员 · %d 帖子", item.MemberCount, item.PostCount)
			content := strings.Join([]string{
				item.Description,
				"规则：" + item.Rules,
				"公告：" + item.Announcement,
				"标签：" + strings.Join(item.Tags, "、"),
			}, "\n")
			chunks, err := s.makeKnowledgeDocuments(
				assistantdomain.KnowledgeSourceGroup,
				item.ID.String(),
				item.Name,
				truncateText(item.Description, 72),
				content,
				"/groups/"+item.ID.String(),
				meta,
				"圈子详情页",
				item.Tags,
				now,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			docs = append(docs, chunks...)
		}

		if len(items) < assistantIndexPageSize {
			break
		}
	}
	return docs, nil
}

func (s *AssistantService) buildEventKnowledgeDocuments(ctx context.Context, now time.Time) ([]*assistantdomain.KnowledgeDocument, error) {
	if s.eventService == nil {
		return nil, nil
	}

	status := event.EventStatusPublished
	docs := make([]*assistantdomain.KnowledgeDocument, 0, 128)
	for page := 1; ; page++ {
		items, _, err := s.eventService.ListEvents(ctx, ListEventsInput{
			Status:   &status,
			Page:     page,
			PageSize: assistantIndexPageSize,
		})
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			if item == nil {
				continue
			}
			location := item.Location
			if item.IsOnline {
				location = "线上活动"
			}
			meta := fmt.Sprintf("%s · %s · %d 人", item.StartTime.Format("01-02 15:04"), location, item.AttendeeCount)
			content := strings.Join([]string{
				item.Description,
				"地点：" + location,
				"标签：" + strings.Join(item.Tags, "、"),
			}, "\n")
			chunks, err := s.makeKnowledgeDocuments(
				assistantdomain.KnowledgeSourceEvent,
				item.ID.String(),
				item.Title,
				truncateText(item.Description, 72),
				content,
				"/events/"+item.ID.String(),
				meta,
				"活动详情页",
				item.Tags,
				now,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			docs = append(docs, chunks...)
		}

		if len(items) < assistantIndexPageSize {
			break
		}
	}
	return docs, nil
}

func (s *AssistantService) makeKnowledgeDocuments(
	sourceType string,
	sourceKey string,
	title string,
	summary string,
	content string,
	href string,
	meta string,
	sourceLabel string,
	tags []string,
	indexedAt time.Time,
	sourceUpdatedAt time.Time,
) ([]*assistantdomain.KnowledgeDocument, error) {
	normalizedSummary := strings.TrimSpace(summary)
	chunks := splitKnowledgeChunks(content, assistantKnowledgeChunkSize, assistantKnowledgeChunkOverlap)
	if len(chunks) == 0 {
		fallback := firstNonEmpty(normalizedSummary, title)
		if fallback != "" {
			chunks = []string{fallback}
		}
	}

	items := make([]*assistantdomain.KnowledgeDocument, 0, len(chunks))
	for idx, chunk := range chunks {
		nextSummary := normalizedSummary
		if nextSummary == "" {
			nextSummary = truncateText(chunk, 96)
		}
		searchText := buildKnowledgeSearchText(title, nextSummary, chunk, meta, tags)
		vec, err := s.embedKnowledgeText(searchText)
		if err != nil {
			return nil, err
		}
		items = append(items, &assistantdomain.KnowledgeDocument{
			ID:              uuid.New(),
			SourceType:      sourceType,
			SourceKey:       sourceKey,
			ChunkIndex:      idx,
			Title:           title,
			Summary:         nextSummary,
			Content:         chunk,
			Href:            href,
			Meta:            meta,
			SourceLabel:     sourceLabel,
			Tags:            tags,
			SearchText:      searchText,
			Embedding:       vec,
			IndexedAt:       indexedAt,
			SourceUpdatedAt: sourceUpdatedAt,
		})
	}
	return items, nil
}

func (s *AssistantService) embedKnowledgeText(text string) ([]float64, error) {
	text = normalizeKnowledgeText(text)
	if text == "" {
		return nil, nil
	}

	if s.embedder != nil {
		if vec, err := s.embedder.Embed(text); err == nil && len(vec) > 0 {
			return normalizeEmbedding(vec), nil
		}
	}

	vec, err := embedding.NewSimpleEmbedder().Embed(text)
	if err != nil {
		return nil, err
	}
	return normalizeEmbedding(vec), nil
}

func mergeKnowledgeCandidates(
	keywordDocs []*assistantdomain.KnowledgeDocument,
	vectorDocs []*assistantdomain.KnowledgeDocument,
	queryVec []float64,
	intent assistantIntent,
) []assistantRetrievalCandidate {
	bySource := make(map[string]*assistantRetrievalCandidate, len(keywordDocs)+len(vectorDocs))
	var maxKeyword float64
	var maxVector float64

	for _, doc := range keywordDocs {
		if doc == nil {
			continue
		}
		key := doc.SourceType + ":" + doc.SourceKey
		item, ok := bySource[key]
		if !ok {
			item = &assistantRetrievalCandidate{doc: doc}
			bySource[key] = item
		}
		if doc.KeywordScore >= item.keywordScore {
			prev := item.keywordScore
			item.keywordScore = doc.KeywordScore
			if item.doc == nil || doc.KeywordScore >= prev {
				item.doc = doc
			}
		}
		maxKeyword = max(maxKeyword, doc.KeywordScore)
	}

	for _, doc := range vectorDocs {
		if doc == nil || len(doc.Embedding) == 0 || len(queryVec) == 0 {
			continue
		}
		score := cosineSimilarity(queryVec, doc.Embedding)
		if score <= 0 {
			continue
		}

		key := doc.SourceType + ":" + doc.SourceKey
		item, ok := bySource[key]
		if !ok {
			item = &assistantRetrievalCandidate{doc: doc}
			bySource[key] = item
		}
		if score >= item.vectorScore {
			prev := item.vectorScore
			item.vectorScore = score
			if item.doc == nil || score >= prev {
				item.doc = doc
			}
		}
		maxVector = max(maxVector, score)
	}

	candidates := make([]assistantRetrievalCandidate, 0, len(bySource))
	for _, item := range bySource {
		if item.doc == nil {
			continue
		}
		keywordNorm := normalizedScore(item.keywordScore, maxKeyword)
		vectorNorm := normalizedScore(item.vectorScore, maxVector)
		if keywordNorm == 0 && vectorNorm == 0 {
			continue
		}

		item.finalScore = keywordNorm*0.58 + vectorNorm*0.42
		item.finalScore += float64(assistantIntentKindBoost(intent, item.doc.SourceType)) * 0.04
		item.finalScore += assistantFreshnessBoost(item.doc.SourceType, item.doc.SourceUpdatedAt)
		item.reason = knowledgeCandidateReason(keywordNorm, vectorNorm)
		candidates = append(candidates, *item)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].finalScore == candidates[j].finalScore {
			if candidates[i].doc.SourceUpdatedAt.Equal(candidates[j].doc.SourceUpdatedAt) {
				return candidates[i].doc.Title < candidates[j].doc.Title
			}
			return candidates[i].doc.SourceUpdatedAt.After(candidates[j].doc.SourceUpdatedAt)
		}
		return candidates[i].finalScore > candidates[j].finalScore
	})
	return candidates
}

func splitKnowledgeChunks(text string, size, overlap int) []string {
	text = normalizeKnowledgeText(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= size || size <= 0 {
		return []string{text}
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 4
	}

	step := max(size-overlap, 1)
	chunks := make([]string, 0, len(runes)/step+1)
	for start := 0; start < len(runes); start += step {
		end := min(start+size, len(runes))
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func buildKnowledgeSearchText(parts ...any) string {
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		switch value := part.(type) {
		case string:
			if trimmed := normalizeKnowledgeText(value); trimmed != "" {
				lines = append(lines, trimmed)
			}
		case []string:
			if len(value) > 0 {
				lines = append(lines, strings.Join(value, " "))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func normalizeKnowledgeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	limit := min(len(a), len(b))
	var dot float64
	var normA float64
	var normB float64
	for i := 0; i < limit; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func normalizeEmbedding(vec []float64) []float64 {
	if len(vec) == 0 {
		return vec
	}
	var norm float64
	for _, value := range vec {
		norm += value * value
	}
	if norm == 0 {
		return vec
	}
	norm = math.Sqrt(norm)
	out := make([]float64, len(vec))
	for i, value := range vec {
		out[i] = value / norm
	}
	return out
}

func normalizedScore(value, maxValue float64) float64 {
	if value <= 0 || maxValue <= 0 {
		return 0
	}
	score := value / maxValue
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func knowledgeCandidateReason(keywordNorm, vectorNorm float64) string {
	switch {
	case keywordNorm >= 0.5 && vectorNorm >= 0.5:
		return "关键词和语义都与问题高度相关"
	case keywordNorm >= 0.5:
		return "关键词匹配度高"
	case vectorNorm >= 0.5:
		return "语义相关度高"
	default:
		return "与当前问题相关"
	}
}

func assistantFreshnessBoost(sourceType string, updatedAt time.Time) float64 {
	if updatedAt.IsZero() {
		return 0
	}
	age := time.Since(updatedAt)
	switch sourceType {
	case assistantdomain.KnowledgeSourceEvent:
		if age <= 14*24*time.Hour {
			return 0.08
		}
		if age <= 45*24*time.Hour {
			return 0.04
		}
	case assistantdomain.KnowledgeSourcePost:
		if age <= 7*24*time.Hour {
			return 0.05
		}
		if age <= 30*24*time.Hour {
			return 0.02
		}
	case assistantdomain.KnowledgeSourceGroup:
		if age <= 14*24*time.Hour {
			return 0.03
		}
	}
	return 0
}

func knowledgeSourceTypes(settings *assistantdomain.Settings) []string {
	if settings == nil {
		return nil
	}
	items := make([]string, 0, 4)
	if settings.IncludePages {
		items = append(items, assistantdomain.KnowledgeSourcePage)
	}
	if settings.IncludePosts {
		items = append(items, assistantdomain.KnowledgeSourcePost)
	}
	if settings.IncludeGroups {
		items = append(items, assistantdomain.KnowledgeSourceGroup)
	}
	if settings.IncludeEvents {
		items = append(items, assistantdomain.KnowledgeSourceEvent)
	}
	return items
}
