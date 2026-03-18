package usecase

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/embedding"
	"github.com/studio/platform/internal/infra/llm"
	"github.com/studio/platform/internal/observability/assistantmetrics"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/cache"
	"github.com/studio/platform/internal/pkg/resilience"
)

// AssistantChatMessage is the user/assistant message payload exchanged with the frontend.
type AssistantChatMessage struct {
	ID        string             `json:"id,omitempty"`
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	Cards     []AssistantCard    `json:"cards,omitempty"`
	Insights  []AssistantInsight `json:"insights,omitempty"`
	CreatedAt time.Time          `json:"created_at,omitempty"`
}

// AssistantPageContext carries non-persisted page-level hints for the current request.
type AssistantPageContext struct {
	Path        string            `json:"path,omitempty"`
	Kind        string            `json:"kind"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary,omitempty"`
	PromptHints []string          `json:"prompt_hints,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
}

// AssistantCard is a structured recommendation displayed next to the assistant reply.
type AssistantCard = assistantdomain.Card

// AssistantInsight is a structured copilot output displayed with the reply.
type AssistantInsight = assistantdomain.Insight

// AssistantMeta is sent before the streamed answer so the UI can render recommendations early.
type AssistantMeta struct {
	Query          string             `json:"query"`
	Provider       string             `json:"provider"`
	Fallback       bool               `json:"fallback"`
	Intent         string             `json:"intent,omitempty"`
	IntentLabel    string             `json:"intent_label,omitempty"`
	SourceCounts   map[string]int     `json:"source_counts,omitempty"`
	ConversationID string             `json:"conversation_id,omitempty"`
	ResponseID     string             `json:"response_id,omitempty"`
	Cards          []AssistantCard    `json:"cards"`
	Insights       []AssistantInsight `json:"insights,omitempty"`
}

type assistantIntent string

const (
	assistantIntentGeneral    assistantIntent = "general"
	assistantIntentOnboarding assistantIntent = "onboarding"
	assistantIntentGroups     assistantIntent = "groups"
	assistantIntentEvents     assistantIntent = "events"
	assistantIntentPosting    assistantIntent = "posting"
	assistantIntentUsers      assistantIntent = "users"
	assistantIntentContent    assistantIntent = "content"
)

// AssistantService powers the site AI helper.
type AssistantService struct {
	cfg                 configs.AssistantConfig
	llmClient           *llm.OpenAICompatibleClient
	embedder            embedding.Embedder
	historyRepo         assistantdomain.Repository
	bookmarkSvc         *BookmarkService
	postService         *PostService
	groupService        *GroupService
	eventService        *EventService
	userService         *UserService
	syncGroup           cache.Group
	lastKnowledgeSyncNS atomic.Int64
	llmCircuit          *resilience.CircuitBreaker
}

// NewAssistantService creates a lightweight assistant service.
func NewAssistantService(
	cfg configs.AssistantConfig,
	llmClient *llm.OpenAICompatibleClient,
	embedder embedding.Embedder,
	historyRepo assistantdomain.Repository,
	bookmarkSvc *BookmarkService,
	postService *PostService,
	groupService *GroupService,
	eventService *EventService,
	userService *UserService,
) *AssistantService {
	if cfg.MaxContextItems <= 0 {
		cfg.MaxContextItems = 6
	}
	if cfg.PersonaName == "" {
		cfg.PersonaName = "霜牙"
	}
	if cfg.Provider == "" {
		cfg.Provider = "deepseek"
	}
	if cfg.RetrievalLimit <= 0 {
		cfg.RetrievalLimit = 8
	}
	if cfg.VectorScanLimit <= 0 {
		cfg.VectorScanLimit = 1200
	}
	if cfg.SyncIntervalSec <= 0 {
		cfg.SyncIntervalSec = 600
	}
	if cfg.LLMRetryMax <= 0 {
		cfg.LLMRetryMax = 2
	}
	if cfg.CircuitFailures <= 0 {
		cfg.CircuitFailures = 3
	}
	if cfg.CircuitOpenSec <= 0 {
		cfg.CircuitOpenSec = 60
	}

	return &AssistantService{
		cfg:          cfg,
		llmClient:    llmClient,
		embedder:     embedder,
		historyRepo:  historyRepo,
		bookmarkSvc:  bookmarkSvc,
		postService:  postService,
		groupService: groupService,
		eventService: eventService,
		userService:  userService,
		llmCircuit:   resilience.NewCircuitBreaker(cfg.CircuitFailures, time.Duration(cfg.CircuitOpenSec)*time.Second),
	}
}

// HistoryEnabled reports whether server-side assistant persistence is available.
func (s *AssistantService) HistoryEnabled() bool {
	return s != nil && s.historyRepo != nil
}

// StreamReply streams a response for the provided conversation history.
func (s *AssistantService) StreamReply(
	ctx context.Context,
	userID uuid.UUID,
	messages []AssistantChatMessage,
	pageContext *AssistantPageContext,
	onMeta func(AssistantMeta) error,
	onToken func(string) error,
) error {
	settings, err := s.resolveRuntimeSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return apperr.New(apperr.CodeForbidden, "AI 助手当前已关闭")
	}

	normalized := sanitizeAssistantMessages(messages)
	latestUser := latestUserMessage(normalized)
	if latestUser == "" {
		return apperr.BadRequest("请输入你想咨询的问题")
	}

	meta, contextText, fallbackAnswer := s.buildPromptContext(ctx, userID, latestUser, pageContext, settings)
	if onMeta != nil {
		if err := onMeta(meta); err != nil {
			return err
		}
	}

	if s.llmClient == nil || !s.llmClient.Configured() {
		assistantmetrics.RecordFallback("provider_unconfigured")
		return streamText(fallbackAnswer, onToken)
	}
	if s.llmCircuit != nil {
		assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
		if !s.llmCircuit.Allow() {
			assistantmetrics.RecordFallback("chat_circuit_open")
			return streamText(fallbackAnswer+"\n\n（当前模型链路暂时熔断，我先给你返回站内检索结果。）", onToken)
		}
	}

	llmMessages := make([]llm.ChatMessage, 0, len(normalized)+1)
	llmMessages = append(llmMessages, llm.ChatMessage{
		Role:    "system",
		Content: s.buildSystemPrompt(contextText, settings),
	})
	for _, msg := range normalized {
		llmMessages = append(llmMessages, llm.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	streamedAny := false
	maxAttempts := max(s.cfg.LLMRetryMax, 1)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = s.llmClient.StreamChat(ctx, llmMessages, func(token string) error {
			streamedAny = true
			return onToken(token)
		})
		if err == nil {
			if s.llmCircuit != nil {
				s.llmCircuit.RecordSuccess()
				assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
			}
			return nil
		}
		if streamedAny {
			if s.llmCircuit != nil {
				s.llmCircuit.RecordFailure(err)
				assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
			}
			return err
		}
		if s.llmCircuit != nil {
			s.llmCircuit.RecordFailure(err)
			assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
		}
		if attempt < maxAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
	}

	meta.Fallback = true
	if onMeta != nil {
		if err := onMeta(meta); err != nil {
			return err
		}
	}
	assistantmetrics.RecordFallback("provider_error")
	return streamText(fallbackAnswer+"\n\n（模型暂时未响应，我先把站内检索到的信息整理给你。）", onToken)
}

// GetSettings returns the effective assistant settings.
func (s *AssistantService) GetSettings(ctx context.Context) (*assistantdomain.Settings, error) {
	return s.resolveRuntimeSettings(ctx)
}

// UpdateSettings persists assistant settings.
func (s *AssistantService) UpdateSettings(ctx context.Context, updatedBy uuid.UUID, input assistantdomain.Settings) (*assistantdomain.Settings, error) {
	if !s.HistoryEnabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "AI 设置存储未启用", nil)
	}

	settings := s.defaultSettings()
	settings.Enabled = input.Enabled
	if name := strings.TrimSpace(input.PersonaName); name != "" {
		settings.PersonaName = truncateText(name, 32)
	}
	settings.SystemPrompt = strings.TrimSpace(input.SystemPrompt)
	if input.MaxContextItems > 0 {
		settings.MaxContextItems = input.MaxContextItems
	}
	if settings.MaxContextItems < 2 {
		settings.MaxContextItems = 2
	}
	if settings.MaxContextItems > 12 {
		settings.MaxContextItems = 12
	}
	settings.IncludePages = input.IncludePages
	settings.IncludePosts = input.IncludePosts
	settings.IncludeUsers = input.IncludeUsers
	settings.IncludeTags = input.IncludeTags
	settings.IncludeGroups = input.IncludeGroups
	settings.IncludeEvents = input.IncludeEvents
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = &updatedBy

	if !settings.IncludePages && !settings.IncludePosts && !settings.IncludeUsers && !settings.IncludeTags && !settings.IncludeGroups && !settings.IncludeEvents {
		return nil, apperr.BadRequest("至少保留一种检索来源")
	}

	if err := s.historyRepo.UpsertSettings(ctx, settings); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "保存 AI 设置失败", err)
	}
	return settings, nil
}

// PrepareConversation resolves or creates a persisted conversation and appends the latest user message.
func (s *AssistantService) PrepareConversation(
	ctx context.Context,
	userID uuid.UUID,
	conversationID *uuid.UUID,
	latestUserContent string,
) (*assistantdomain.Conversation, []AssistantChatMessage, error) {
	if !s.HistoryEnabled() {
		return nil, nil, apperr.Wrap(apperr.CodeInternalError, "AI 会话存储未启用", nil)
	}
	if userID == uuid.Nil {
		return nil, nil, apperr.ErrUnauthorized
	}

	latestUserContent = strings.TrimSpace(latestUserContent)
	if latestUserContent == "" {
		return nil, nil, apperr.BadRequest("请输入你想咨询的问题")
	}

	conv, err := s.resolveConversation(ctx, userID, conversationID, latestUserContent)
	if err != nil {
		return nil, nil, err
	}

	if err := s.historyRepo.CreateMessage(ctx, &assistantdomain.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           assistantdomain.RoleUser,
		Content:        latestUserContent,
		CreatedAt:      time.Now(),
	}); err != nil {
		return nil, nil, apperr.Wrap(apperr.CodeInternalError, "保存 AI 提问失败", err)
	}

	recent, err := s.historyRepo.ListRecentMessages(ctx, conv.ID, 12)
	if err != nil {
		return nil, nil, apperr.Wrap(apperr.CodeInternalError, "读取 AI 会话上下文失败", err)
	}

	messages := make([]AssistantChatMessage, 0, len(recent))
	for _, msg := range recent {
		messages = append(messages, AssistantChatMessage{
			Role:     string(msg.Role),
			Content:  msg.Content,
			Cards:    msg.Cards,
			Insights: msg.Insights,
		})
	}
	return conv, messages, nil
}

// SaveAssistantReply persists the assistant answer for a conversation.
func (s *AssistantService) SaveAssistantReply(ctx context.Context, responseID, conversationID uuid.UUID, content string, cards []AssistantCard, insights []AssistantInsight) error {
	if !s.HistoryEnabled() || conversationID == uuid.Nil || strings.TrimSpace(content) == "" {
		return nil
	}
	if responseID == uuid.Nil {
		responseID = uuid.New()
	}
	if err := s.historyRepo.CreateMessage(ctx, &assistantdomain.Message{
		ID:             responseID,
		ConversationID: conversationID,
		Role:           assistantdomain.RoleAssistant,
		Content:        content,
		Cards:          cards,
		Insights:       insights,
		CreatedAt:      time.Now(),
	}); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "保存 AI 回复失败", err)
	}
	return nil
}

// ListConversations returns persisted assistant conversations for the given user.
func (s *AssistantService) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*assistantdomain.Conversation, int64, error) {
	if !s.HistoryEnabled() {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "AI 会话存储未启用", nil)
	}
	if userID == uuid.Nil {
		return nil, 0, apperr.ErrUnauthorized
	}

	items, total, err := s.historyRepo.ListConversations(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取 AI 会话列表失败", err)
	}
	return items, total, nil
}

// GetConversation loads a persisted assistant conversation and validates ownership.
func (s *AssistantService) GetConversation(ctx context.Context, userID, conversationID uuid.UUID, page, pageSize int) (*assistantdomain.Conversation, []*assistantdomain.Message, int64, error) {
	if !s.HistoryEnabled() {
		return nil, nil, 0, apperr.Wrap(apperr.CodeInternalError, "AI 会话存储未启用", nil)
	}

	conv, err := s.historyRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		if err == assistantdomain.ErrConversationNotFound {
			return nil, nil, 0, apperr.ErrNotFound
		}
		return nil, nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取 AI 会话失败", err)
	}
	if conv.UserID != userID {
		return nil, nil, 0, apperr.ErrForbidden
	}

	items, total, err := s.historyRepo.ListMessages(ctx, conversationID, page, pageSize)
	if err != nil {
		return nil, nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取 AI 会话消息失败", err)
	}
	return conv, items, total, nil
}

func (s *AssistantService) buildPromptContext(ctx context.Context, userID uuid.UUID, query string, pageContext *AssistantPageContext, settings *assistantdomain.Settings) (AssistantMeta, string, string) {
	intent := detectAssistantIntent(query, pageContext)
	cards := s.collectCards(ctx, query, intent, settings)
	copilot := s.buildCopilotBundle(ctx, query, pageContext)
	meta := AssistantMeta{
		Query:        query,
		Provider:     s.cfg.Provider,
		Fallback:     s.llmClient == nil || !s.llmClient.Configured(),
		Intent:       string(intent),
		IntentLabel:  assistantIntentDisplayLabel(intent),
		SourceCounts: assistantSourceCounts(cards),
		Cards:        cards,
		Insights:     copilot.Insights,
	}

	var contextParts []string
	contextParts = append(contextParts, fmt.Sprintf("当前日期: %s", time.Now().Format("2006-01-02")))
	contextParts = append(contextParts, fmt.Sprintf("用户当前问题类型: %s", assistantIntentPromptLabel(intent)))
	contextParts = append(contextParts, siteOverviewContext())
	if pageContextText := buildAssistantPageContext(pageContext); pageContextText != "" {
		contextParts = append(contextParts, pageContextText)
	}
	if copilot.ContextText != "" {
		contextParts = append(contextParts, copilot.ContextText)
	}
	if bookmarkContext := s.buildBookmarkContext(ctx, userID); bookmarkContext != "" {
		contextParts = append(contextParts, bookmarkContext)
	}
	if len(cards) > 0 {
		var itemLines []string
		for _, card := range cards {
			ref := card.Ref
			if ref == "" {
				ref = "R?"
			}
			line := fmt.Sprintf("[%s] [%s] %s - %s (链接: %s)", ref, card.Kind, card.Title, card.Summary, card.Href)
			if card.Meta != "" {
				line += " | " + card.Meta
			}
			if card.Reason != "" {
				line += " | 推荐理由: " + card.Reason
			}
			if card.Source != "" {
				line += " | 来源: " + card.Source
			}
			itemLines = append(itemLines, line)
		}
		contextParts = append(contextParts, "可引用的站内信息:\n"+strings.Join(itemLines, "\n"))
	}

	return meta, strings.Join(contextParts, "\n\n"), buildFallbackAnswer(settings.PersonaName, query, cards, copilot.FallbackText)
}

func (s *AssistantService) collectCards(ctx context.Context, query string, intent assistantIntent, settings *assistantdomain.Settings) []AssistantCard {
	maxItems := settings.MaxContextItems
	if maxItems <= 0 {
		maxItems = 6
	}

	seen := make(map[string]struct{}, maxItems)
	cards := make([]AssistantCard, 0, maxItems*2)
	type scoredCard struct {
		card  AssistantCard
		score int
	}
	scored := make([]scoredCard, 0, maxItems*2)

	appendUnique := func(items ...AssistantCard) {
		for _, item := range items {
			if item.Href == "" {
				continue
			}
			if _, ok := seen[item.Href]; ok {
				continue
			}
			seen[item.Href] = struct{}{}
			scored = append(scored, scoredCard{
				card:  item,
				score: scoreAssistantCard(item, query, intent),
			})
		}
	}

	if knowledgeCards, err := s.collectKnowledgeCards(ctx, query, intent, settings); err == nil {
		appendUnique(knowledgeCards...)
	}
	if settings.IncludeUsers {
		appendUnique(s.collectUserCards(ctx, query, assistantCandidateLimit(intent, "user", maxItems))...)
	}
	if settings.IncludeTags {
		appendUnique(s.collectTagCards(ctx, query, assistantCandidateLimit(intent, "tag", maxItems))...)
	}
	if len(scored) == 0 {
		if settings.IncludePages {
			appendUnique(recommendPageCards()...)
		}
		if settings.IncludePosts {
			appendUnique(s.collectPostCards(ctx, query, assistantCandidateLimit(intent, "post", maxItems))...)
		}
		if settings.IncludeGroups {
			appendUnique(s.collectGroupCards(ctx, query, assistantCandidateLimit(intent, "group", maxItems))...)
		}
		if settings.IncludeEvents {
			appendUnique(s.collectEventCards(ctx, query, assistantCandidateLimit(intent, "event", maxItems))...)
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].card.Title < scored[j].card.Title
		}
		return scored[i].score > scored[j].score
	})

	selectedByKind := make(map[string]int, len(scored))
	for _, item := range scored {
		if len(cards) >= maxItems {
			break
		}
		kindLimit := assistantFinalKindLimit(intent, item.card.Kind, maxItems)
		if kindLimit > 0 && selectedByKind[item.card.Kind] >= kindLimit {
			continue
		}
		cards = append(cards, item.card)
		selectedByKind[item.card.Kind]++
	}

	for i := range cards {
		cards[i].Ref = fmt.Sprintf("R%d", i+1)
	}

	return cards
}

func (s *AssistantService) collectUserCards(ctx context.Context, query string, limit int) []AssistantCard {
	if s.userService == nil {
		return nil
	}
	if limit <= 0 {
		limit = 2
	}

	users, err := s.userService.SearchUsers(ctx, query, limit)
	if err != nil || len(users) == 0 {
		return nil
	}

	cards := make([]AssistantCard, 0, len(users))
	for _, item := range users {
		if item.Status != user.StatusActive {
			continue
		}

		displayName := item.Username
		if item.FurryName != nil && strings.TrimSpace(*item.FurryName) != "" {
			displayName = *item.FurryName
		}

		var summaryParts []string
		if item.Species != nil && strings.TrimSpace(*item.Species) != "" {
			summaryParts = append(summaryParts, "物种："+strings.TrimSpace(*item.Species))
		}
		if item.Bio != nil && strings.TrimSpace(*item.Bio) != "" {
			summaryParts = append(summaryParts, truncateText(strings.TrimSpace(*item.Bio), 36))
		}
		summary := strings.Join(summaryParts, " · ")
		if summary == "" {
			summary = "查看这个用户的主页、动态和关注关系。"
		}

		meta := "@" + item.Username
		if item.Role == user.RoleCreator {
			meta += " · 创作者"
		}

		cards = append(cards, AssistantCard{
			Kind:    "user",
			Title:   displayName,
			Summary: summary,
			Href:    "/users/" + item.ID.String(),
			Meta:    meta,
			Reason:  userRecommendationReason(item, query),
			Source:  "用户主页",
		})
	}
	return cards
}

func (s *AssistantService) collectTagCards(ctx context.Context, query string, limit int) []AssistantCard {
	if s.postService == nil {
		return nil
	}
	if limit <= 0 {
		limit = 2
	}

	tags, err := s.postService.GetHotTags(ctx, max(12, limit*4))
	if err != nil || len(tags) == 0 {
		return nil
	}

	query = strings.TrimSpace(strings.ToLower(query))
	cards := make([]AssistantCard, 0, limit)
	for _, tag := range tags {
		if len(cards) >= limit {
			break
		}
		if query != "" && !strings.Contains(strings.ToLower(tag), query) && !strings.Contains(query, strings.ToLower(tag)) {
			continue
		}
		cards = append(cards, AssistantCard{
			Kind:    "tag",
			Title:   "#" + tag,
			Summary: "查看这个标签下的相关动态。",
			Href:    "/tags/" + url.PathEscape(tag),
			Meta:    "/tags/" + tag,
			Reason:  "热门标签，适合快速扩展相关内容",
			Source:  "标签聚合页",
		})
	}

	if len(cards) == 0 {
		for _, tag := range tags[:min(limit, len(tags))] {
			cards = append(cards, AssistantCard{
				Kind:    "tag",
				Title:   "#" + tag,
				Summary: "查看这个标签下的相关动态。",
				Href:    "/tags/" + url.PathEscape(tag),
				Meta:    "/tags/" + tag,
				Reason:  "站内热门标签",
				Source:  "标签聚合页",
			})
		}
	}
	return cards
}

func (s *AssistantService) collectPostCards(ctx context.Context, query string, limit int) []AssistantCard {
	if s.postService == nil {
		return nil
	}
	if limit <= 0 {
		limit = 2
	}

	posts, err := s.postService.SearchPosts(ctx, query, limit)
	if err != nil || len(posts) == 0 {
		posts, _, _ = s.postService.ListExplore(ctx, 1, limit, "")
	}

	cards := make([]AssistantCard, 0, len(posts))
	for _, post := range posts {
		title := strings.TrimSpace(post.Title)
		if title == "" {
			title = truncateText(post.Content, 18)
		}
		summary := truncateText(post.Content, 56)
		meta := fmt.Sprintf("@%s · %d 赞 · %d 评论", post.AuthorUsername, post.LikeCount, post.CommentCount)
		reason := "公开且已审核通过的动态"
		if query != "" && strings.Contains(strings.ToLower(post.Content+" "+post.Title), strings.ToLower(query)) {
			reason = "内容与你的问题关键词相关"
		} else if post.LikeCount+post.CommentCount > 0 {
			reason = "这条动态当前互动表现更高"
		}
		cards = append(cards, AssistantCard{
			Kind:    "post",
			Title:   title,
			Summary: summary,
			Href:    "/posts/" + post.ID.String(),
			Meta:    meta,
			Reason:  reason,
			Source:  "帖子详情页",
		})
	}
	return cards
}

func (s *AssistantService) collectGroupCards(ctx context.Context, query string, limit int) []AssistantCard {
	if s.groupService == nil {
		return nil
	}
	if limit <= 0 {
		limit = 2
	}

	privacy := group.GroupPrivacyPublic
	groups, _, err := s.groupService.ListGroups(ctx, ListGroupsInput{
		Privacy:  &privacy,
		Search:   strings.TrimSpace(query),
		Page:     1,
		PageSize: limit,
	})
	if err != nil || len(groups) == 0 {
		groups, _, _ = s.groupService.ListGroups(ctx, ListGroupsInput{
			Privacy:  &privacy,
			Page:     1,
			PageSize: limit,
		})
	}

	cards := make([]AssistantCard, 0, len(groups))
	for _, item := range groups {
		cards = append(cards, AssistantCard{
			Kind:    "group",
			Title:   item.Name,
			Summary: truncateText(item.Description, 52),
			Href:    "/groups/" + item.ID.String(),
			Meta:    fmt.Sprintf("%d 成员 · %d 帖子", item.MemberCount, item.PostCount),
			Reason:  groupRecommendationReason(item, query),
			Source:  "圈子详情页",
		})
	}
	return cards
}

func (s *AssistantService) collectEventCards(ctx context.Context, query string, limit int) []AssistantCard {
	if s.eventService == nil {
		return nil
	}
	if limit <= 0 {
		limit = 2
	}

	status := event.EventStatusPublished
	events, _, err := s.eventService.ListEvents(ctx, ListEventsInput{
		Status:   &status,
		Page:     1,
		PageSize: max(limit*2, 6),
	})
	if err != nil || len(events) == 0 {
		return nil
	}

	filtered := filterEvents(events, query)
	if len(filtered) == 0 {
		filtered = events
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	cards := make([]AssistantCard, 0, len(filtered))
	for _, item := range filtered {
		location := item.Location
		if item.IsOnline {
			location = "线上活动"
		}
		cards = append(cards, AssistantCard{
			Kind:    "event",
			Title:   item.Title,
			Summary: truncateText(item.Description, 52),
			Href:    "/events/" + item.ID.String(),
			Meta:    fmt.Sprintf("%s · %s", item.StartTime.Format("01-02 15:04"), location),
			Reason:  eventRecommendationReason(item, query),
			Source:  "活动详情页",
		})
	}
	return cards
}

func (s *AssistantService) buildSystemPrompt(contextText string, settings *assistantdomain.Settings) string {
	persona := settings.PersonaName
	base := strings.TrimSpace(fmt.Sprintf(`
你是 %s，一位帅气、可靠、语气自然的 Furry 社区 AI 导览助手。

你的职责：
1. 用简体中文回答，优先帮助用户理解这个网站有什么、去哪里、值得看什么。
2. 只根据给定的站内上下文和通用产品常识回答，不要编造不存在的页面、功能、活动或数据。
3. 如果上下文里已经有推荐内容，优先围绕这些内容给出建议。
4. 语气友好、干练，不要油腻，不要过度卖萌，不要把自己说成真人。
5. 回答尽量简洁，通常 2 到 5 段即可；必要时优先用短 Markdown 列表。
6. 如果用户的问题超出站内信息范围，要明确说明你主要负责本网站导览与推荐。
7. 当你推荐具体内容时，尽量说明推荐理由，并写清楚用户该从哪个入口进入。
8. 只要使用了给定来源中的具体内容或做出具体推荐，就在对应句末附上来源引用，例如 [R1]、[R2]。
9. 不要伪造引用编号；只能使用给定上下文里出现的引用编号。
10. 如果给定上下文里存在“用户最近收藏偏好”，优先结合这些偏好做个性化推荐。
11. 如果上下文里存在“当前页面 Copilot 工具结果”，优先把这些结果转成直接可执行的建议，例如标题备选、标签建议、规则摘要、准备清单。
12. 这些 Copilot 工具只提供只读分析和草稿建议；不要承诺替用户自动发帖、自动加入圈子或自动报名。

以下是你可用的站内信息：
%s
`, persona, contextText))

	if strings.TrimSpace(settings.SystemPrompt) == "" {
		return base
	}
	return base + "\n\n额外规则：\n" + strings.TrimSpace(settings.SystemPrompt)
}

func siteOverviewContext() string {
	return strings.TrimSpace(`
网站定位：一个面向 Furry 同好的社区平台。
主要能力：
- 发布图文动态，支持图片上传、可见性设置、AI 内容标记。
- 浏览关注流、发现页、标签页和搜索页。
- 加入兴趣圈子，查看或参加活动。
- 私信聊天、查看通知、举报与屏蔽。
- 创作者可以查看数据面板和赞助页。
关键页面：
- /feed 关注动态
- /explore 发现页
- /search 搜索
- /groups 圈子
- /events 活动
- /posts/create 发布动态
- /creator 创作者面板
- /notifications 通知中心
- /reports 我的举报
`)
}

func buildFallbackAnswer(personaName, query string, cards []AssistantCard, copilotText string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s 在这。你刚才问的是“%s”。\n\n", personaName, query)

	if strings.TrimSpace(copilotText) != "" {
		b.WriteString(copilotText)
		if len(cards) > 0 {
			b.WriteString("\n\n另外，我再补几个可以直接点开的站内入口：\n")
		}
	}

	if len(cards) == 0 {
		if strings.TrimSpace(copilotText) == "" {
			b.WriteString("我先给你一个站内导航建议：如果你是第一次来，建议先看“发现页 /explore”、再逛“圈子 /groups”和“活动 /events”，想发内容就去“/posts/create”。")
		}
		return b.String()
	}

	if strings.TrimSpace(copilotText) == "" {
		b.WriteString("我先根据站内信息帮你整理了几个值得直接点开的入口：\n")
	}
	for i, card := range cards {
		ref := card.Ref
		if ref == "" {
			ref = fmt.Sprintf("R%d", i+1)
		}
		fmt.Fprintf(&b, "%d. [%s] %s：%s", i+1, ref, card.Title, card.Summary)
		if card.Meta != "" {
			fmt.Fprintf(&b, "（%s）", card.Meta)
		}
		if card.Reason != "" {
			fmt.Fprintf(&b, "\n   推荐理由：%s", card.Reason)
		}
		if card.Source != "" {
			fmt.Fprintf(&b, "\n   来源：%s", card.Source)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n如果你愿意，我还可以继续按你的偏好细化，比如“偏创作向”“偏社交向”“偏线下活动”。")
	return b.String()
}

func sanitizeAssistantMessages(messages []AssistantChatMessage) []AssistantChatMessage {
	out := make([]AssistantChatMessage, 0, len(messages))
	for _, msg := range messages {
		role := strings.TrimSpace(msg.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		out = append(out, AssistantChatMessage{
			Role:     role,
			Content:  truncateText(content, 1200),
			Insights: msg.Insights,
		})
	}
	if len(out) > 12 {
		out = out[len(out)-12:]
	}
	return out
}

func latestUserMessage(messages []AssistantChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

func streamText(text string, onToken func(string) error) error {
	for _, chunk := range chunkText(text, 18) {
		if err := onToken(chunk); err != nil {
			return err
		}
	}
	return nil
}

func chunkText(text string, size int) []string {
	if size <= 0 || utf8.RuneCountInString(text) <= size {
		return []string{text}
	}
	runes := []rune(text)
	chunks := make([]string, 0, len(runes)/size+1)
	for start := 0; start < len(runes); start += size {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}

func truncateText(text string, limit int) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if limit <= 0 || utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func filterEvents(items []*event.Event, query string) []*event.Event {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return items
	}

	filtered := make([]*event.Event, 0, len(items))
	for _, item := range items {
		var haystack []string
		haystack = append(haystack, item.Title, item.Description, item.Location)
		haystack = append(haystack, item.Tags...)
		if strings.Contains(strings.ToLower(strings.Join(haystack, " ")), query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func recommendPageCards() []AssistantCard {
	return []AssistantCard{
		{
			Kind:    "page",
			Title:   "发现页",
			Summary: "看热门动态、标签和创作者，适合第一次来先逛。",
			Href:    "/explore",
			Meta:    "/explore",
			Reason:  "适合第一次来快速熟悉社区内容结构",
			Source:  "站内固定导航",
		},
		{
			Kind:    "page",
			Title:   "关注动态",
			Summary: "查看你关注对象的最新内容和互动。",
			Href:    "/feed",
			Meta:    "/feed",
			Reason:  "如果你已经关注了一些人，这里最能体现个性化内容",
			Source:  "站内固定导航",
		},
		{
			Kind:    "page",
			Title:   "发布动态",
			Summary: "支持图文发布、图片上传、AI 内容标记和可见性设置。",
			Href:    "/posts/create",
			Meta:    "/posts/create",
			Reason:  "适合想马上开始发内容的用户",
			Source:  "站内固定导航",
		},
		{
			Kind:    "page",
			Title:   "圈子广场",
			Summary: "按兴趣找同好、加入圈子、看成员和帖子数。",
			Href:    "/groups",
			Meta:    "/groups",
			Reason:  "适合按兴趣主题找社区和同好",
			Source:  "站内固定导航",
		},
		{
			Kind:    "page",
			Title:   "活动广场",
			Summary: "查看近期线上线下活动，支持报名参加。",
			Href:    "/events",
			Meta:    "/events",
			Reason:  "适合找近期活动或线下聚会信息",
			Source:  "站内固定导航",
		},
		{
			Kind:    "page",
			Title:   "创作者面板",
			Summary: "查看帖子、粉丝、互动和打赏数据。",
			Href:    "/creator",
			Meta:    "/creator",
			Reason:  "适合已经在创作或打算持续运营内容的用户",
			Source:  "站内固定导航",
		},
	}
}

func (s *AssistantService) defaultSettings() *assistantdomain.Settings {
	persona := strings.TrimSpace(s.cfg.PersonaName)
	if persona == "" {
		persona = "霜牙"
	}
	maxItems := s.cfg.MaxContextItems
	if maxItems <= 0 {
		maxItems = 6
	}

	return &assistantdomain.Settings{
		Enabled:         true,
		PersonaName:     persona,
		SystemPrompt:    "",
		MaxContextItems: maxItems,
		IncludePages:    true,
		IncludePosts:    true,
		IncludeUsers:    true,
		IncludeTags:     true,
		IncludeGroups:   true,
		IncludeEvents:   true,
	}
}

func (s *AssistantService) resolveRuntimeSettings(ctx context.Context) (*assistantdomain.Settings, error) {
	settings := s.defaultSettings()
	if !s.HistoryEnabled() {
		return settings, nil
	}

	stored, err := s.historyRepo.GetSettings(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 AI 设置失败", err)
	}
	if stored == nil {
		return settings, nil
	}

	if strings.TrimSpace(stored.PersonaName) == "" {
		stored.PersonaName = settings.PersonaName
	}
	if stored.MaxContextItems <= 0 {
		stored.MaxContextItems = settings.MaxContextItems
	}
	return stored, nil
}

func (s *AssistantService) resolveConversation(
	ctx context.Context,
	userID uuid.UUID,
	conversationID *uuid.UUID,
	latestUserContent string,
) (*assistantdomain.Conversation, error) {
	if conversationID != nil && *conversationID != uuid.Nil {
		conv, err := s.historyRepo.GetConversationByID(ctx, *conversationID)
		if err != nil {
			if err == assistantdomain.ErrConversationNotFound {
				return nil, apperr.ErrNotFound
			}
			return nil, apperr.Wrap(apperr.CodeInternalError, "读取 AI 会话失败", err)
		}
		if conv.UserID != userID {
			return nil, apperr.ErrForbidden
		}
		return conv, nil
	}

	now := time.Now()
	conv := &assistantdomain.Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     truncateText(latestUserContent, 28),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.historyRepo.CreateConversation(ctx, conv); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建 AI 会话失败", err)
	}
	return conv, nil
}

func userRecommendationReason(u *user.User, query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query != "" {
		if strings.Contains(strings.ToLower(u.Username), query) {
			return "用户名与你的问题关键词匹配"
		}
		if u.FurryName != nil && strings.Contains(strings.ToLower(strings.TrimSpace(*u.FurryName)), query) {
			return "兽名与你的问题关键词匹配"
		}
		if u.Species != nil && strings.Contains(strings.ToLower(strings.TrimSpace(*u.Species)), query) {
			return "物种信息与你的问题关键词相关"
		}
	}
	if u.Role == user.RoleCreator {
		return "这是创作者账号，适合继续查看其内容和动态"
	}
	return "这个用户的主页信息和你的问题更相关"
}

func (s *AssistantService) buildBookmarkContext(ctx context.Context, userID uuid.UUID) string {
	if userID == uuid.Nil || s.bookmarkSvc == nil {
		return ""
	}

	var lines []string
	if posts, _, err := s.bookmarkSvc.ListPosts(ctx, userID, 1, 2, "latest"); err == nil {
		for _, item := range posts {
			lines = append(lines, fmt.Sprintf("- 帖子：%s", truncateText(firstNonEmpty(item.Title, item.Content), 32)))
		}
	}
	if groups, _, err := s.bookmarkSvc.ListGroups(ctx, userID, 1, 2, "latest"); err == nil {
		for _, item := range groups {
			lines = append(lines, fmt.Sprintf("- 圈子：%s", item.Name))
		}
	}
	if events, _, err := s.bookmarkSvc.ListEvents(ctx, userID, 1, 2, "latest"); err == nil {
		for _, item := range events {
			lines = append(lines, fmt.Sprintf("- 活动：%s", item.Title))
		}
	}

	if len(lines) == 0 {
		return ""
	}
	return "用户最近收藏偏好：\n" + strings.Join(lines, "\n")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func groupRecommendationReason(g *group.Group, query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query != "" {
		haystack := strings.ToLower(g.Name + " " + g.Description + " " + strings.Join(g.Tags, " "))
		if strings.Contains(haystack, query) {
			return "圈子名称或简介与你的问题关键词相关"
		}
	}
	if g.MemberCount > 0 {
		return fmt.Sprintf("这个圈子已经有 %d 位成员，适合继续深入了解", g.MemberCount)
	}
	return "这是一个公开圈子，适合继续查看详情"
}

func eventRecommendationReason(e *event.Event, query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query != "" {
		haystack := strings.ToLower(e.Title + " " + e.Description + " " + e.Location + " " + strings.Join(e.Tags, " "))
		if strings.Contains(haystack, query) {
			return "活动主题与你的问题关键词相关"
		}
	}
	if e.IsOnline {
		return "这是线上活动，参与门槛更低"
	}
	return "这是近期可参加的公开活动"
}

func scoreAssistantCard(card AssistantCard, query string, intent assistantIntent) int {
	score := kindBaseScore(card.Kind)
	score += assistantIntentKindBoost(intent, card.Kind)
	score += pageIntentBoost(intent, card)

	fullQuery := strings.TrimSpace(strings.ToLower(query))
	if fullQuery == "" {
		return score
	}

	title := strings.ToLower(card.Title)
	summary := strings.ToLower(card.Summary)
	meta := strings.ToLower(card.Meta)
	reason := strings.ToLower(card.Reason)

	if strings.Contains(title, fullQuery) {
		score += 140
	}
	if strings.Contains(summary, fullQuery) {
		score += 90
	}
	if strings.Contains(meta, fullQuery) {
		score += 55
	}
	if strings.Contains(reason, fullQuery) {
		score += 40
	}

	for _, token := range queryTokens(fullQuery) {
		switch {
		case strings.Contains(title, token):
			score += 20
		case strings.Contains(summary, token):
			score += 14
		case strings.Contains(meta, token):
			score += 8
		case strings.Contains(reason, token):
			score += 6
		}
	}

	score += kindKeywordBoost(card.Kind, fullQuery)
	return score
}

func kindBaseScore(kind string) int {
	switch kind {
	case "post":
		return 55
	case "user":
		return 48
	case "group":
		return 45
	case "event":
		return 45
	case "tag":
		return 35
	case "page":
		return 25
	default:
		return 10
	}
}

func kindKeywordBoost(kind, query string) int {
	type rule struct {
		kind     string
		keywords []string
		boost    int
	}

	rules := []rule{
		{kind: "event", keywords: []string{"活动", "聚会", "线下", "线上", "报名"}, boost: 50},
		{kind: "group", keywords: []string{"圈子", "社群", "同好", "群组"}, boost: 50},
		{kind: "post", keywords: []string{"帖子", "动态", "内容", "发帖", "作品"}, boost: 45},
		{kind: "user", keywords: []string{"用户", "创作者", "关注谁", "作者"}, boost: 45},
		{kind: "tag", keywords: []string{"标签", "话题"}, boost: 40},
		{kind: "page", keywords: []string{"怎么用", "入口", "先逛", "第一次来"}, boost: 40},
	}

	boost := 0
	for _, rule := range rules {
		if rule.kind != kind {
			continue
		}
		for _, keyword := range rule.keywords {
			if strings.Contains(query, keyword) {
				boost += rule.boost
				break
			}
		}
	}
	return boost
}

func detectAssistantIntent(query string, pageContext *AssistantPageContext) assistantIntent {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		if pageContext != nil && pageContext.Kind == "post_create" {
			return assistantIntentPosting
		}
		return assistantIntentGeneral
	}

	type rule struct {
		intent   assistantIntent
		score    int
		keywords []string
	}

	rules := []rule{
		{intent: assistantIntentOnboarding, score: 4, keywords: []string{"第一次", "新手", "刚来", "先逛", "先看", "怎么玩", "怎么逛", "从哪里开始", "上手"}},
		{intent: assistantIntentGroups, score: 5, keywords: []string{"圈子", "社群", "群组", "同好群", "加群", "社区"}},
		{intent: assistantIntentEvents, score: 5, keywords: []string{"活动", "聚会", "报名", "线下", "线上活动", "漫展"}},
		{intent: assistantIntentPosting, score: 5, keywords: []string{"发帖", "发布", "发动态", "创作", "写帖子", "发作品", "上传图片", "标题", "标签怎么选"}},
		{intent: assistantIntentUsers, score: 5, keywords: []string{"用户", "创作者", "关注谁", "关注什么人", "作者", "找人", "推荐关注"}},
		{intent: assistantIntentContent, score: 4, keywords: []string{"帖子", "动态", "内容", "热门", "有什么可看", "推荐内容", "标签", "话题"}},
	}

	scores := map[assistantIntent]int{
		assistantIntentGeneral: 1,
	}
	for _, rule := range rules {
		for _, keyword := range rule.keywords {
			if strings.Contains(query, keyword) {
				scores[rule.intent] += rule.score
			}
		}
	}

	if strings.Contains(query, "推荐") {
		scores[assistantIntentContent]++
	}
	if strings.Contains(query, "哪里") || strings.Contains(query, "入口") {
		scores[assistantIntentOnboarding]++
	}
	if pageContext != nil {
		switch pageContext.Kind {
		case "post_create":
			scores[assistantIntentPosting] += 4
			if strings.Contains(query, "润色") || strings.Contains(query, "改写") || strings.Contains(query, "标题") || strings.Contains(query, "标签") || strings.Contains(query, "公开") {
				scores[assistantIntentPosting] += 2
			}
		case "group_detail":
			scores[assistantIntentGroups] += 4
			if strings.Contains(query, "加入") || strings.Contains(query, "适合") || strings.Contains(query, "圈子") || strings.Contains(query, "群规") || strings.Contains(query, "发什么") {
				scores[assistantIntentGroups] += 2
			}
		case "event_detail":
			scores[assistantIntentEvents] += 4
			if strings.Contains(query, "参加") || strings.Contains(query, "报名") || strings.Contains(query, "适合") || strings.Contains(query, "准备") || strings.Contains(query, "注意") {
				scores[assistantIntentEvents] += 2
			}
		}
	}

	order := []assistantIntent{
		assistantIntentOnboarding,
		assistantIntentGroups,
		assistantIntentEvents,
		assistantIntentPosting,
		assistantIntentUsers,
		assistantIntentContent,
		assistantIntentGeneral,
	}

	best := assistantIntentGeneral
	bestScore := 0
	for _, intent := range order {
		if scores[intent] > bestScore {
			best = intent
			bestScore = scores[intent]
		}
	}
	return best
}

func assistantIntentPromptLabel(intent assistantIntent) string {
	switch intent {
	case assistantIntentOnboarding:
		return "新用户导览，优先推荐上手入口和适合先逛的内容"
	case assistantIntentGroups:
		return "圈子探索，优先推荐相关圈子和对应入口"
	case assistantIntentEvents:
		return "活动探索，优先推荐近期相关活动和报名入口"
	case assistantIntentPosting:
		return "发帖或创作帮助，优先推荐发布入口、相关帖子和标签"
	case assistantIntentUsers:
		return "找人或关注建议，优先推荐用户主页和相关内容"
	case assistantIntentContent:
		return "内容发现，优先推荐帖子、标签和可继续扩展浏览的入口"
	default:
		return "综合导览，优先给出最值得访问的站内入口和内容"
	}
}

func assistantIntentDisplayLabel(intent assistantIntent) string {
	switch intent {
	case assistantIntentOnboarding:
		return "新手导览"
	case assistantIntentGroups:
		return "圈子探索"
	case assistantIntentEvents:
		return "活动探索"
	case assistantIntentPosting:
		return "发帖帮助"
	case assistantIntentUsers:
		return "找人推荐"
	case assistantIntentContent:
		return "内容发现"
	default:
		return "综合导览"
	}
}

func assistantCandidateLimit(intent assistantIntent, kind string, maxItems int) int {
	base := max(3, maxItems)
	switch intent {
	case assistantIntentOnboarding:
		switch kind {
		case "page":
			return max(base, 6)
		case "group", "event":
			return max(base, 4)
		default:
			return 2
		}
	case assistantIntentGroups:
		switch kind {
		case "group":
			return max(base, 5)
		case "user", "page":
			return 3
		default:
			return 2
		}
	case assistantIntentEvents:
		switch kind {
		case "event":
			return max(base, 5)
		case "group", "page":
			return 3
		default:
			return 2
		}
	case assistantIntentPosting:
		switch kind {
		case "page", "post", "tag":
			return max(base, 4)
		case "user":
			return 3
		default:
			return 2
		}
	case assistantIntentUsers:
		switch kind {
		case "user":
			return max(base, 5)
		case "post", "group":
			return 3
		default:
			return 2
		}
	case assistantIntentContent:
		switch kind {
		case "post":
			return max(base, 5)
		case "tag", "user":
			return 3
		default:
			return 2
		}
	default:
		switch kind {
		case "page", "post", "group", "event":
			return 3
		default:
			return 2
		}
	}
}

func assistantFinalKindLimit(intent assistantIntent, kind string, maxItems int) int {
	switch intent {
	case assistantIntentOnboarding:
		switch kind {
		case "page":
			return min(3, maxItems)
		case "group", "event":
			return 2
		default:
			return 1
		}
	case assistantIntentGroups:
		switch kind {
		case "group":
			return min(3, maxItems)
		case "page", "user":
			return 2
		default:
			return 1
		}
	case assistantIntentEvents:
		switch kind {
		case "event":
			return min(3, maxItems)
		case "page", "group":
			return 2
		default:
			return 1
		}
	case assistantIntentPosting:
		switch kind {
		case "page", "post":
			return min(3, maxItems)
		case "tag":
			return 2
		default:
			return 1
		}
	case assistantIntentUsers:
		switch kind {
		case "user":
			return min(3, maxItems)
		case "post", "group":
			return 2
		default:
			return 1
		}
	case assistantIntentContent:
		switch kind {
		case "post":
			return min(3, maxItems)
		case "tag", "user":
			return 2
		default:
			return 1
		}
	default:
		switch kind {
		case "post", "group", "event", "user":
			return 2
		case "page", "tag":
			return 1
		default:
			return 1
		}
	}
}

func assistantIntentKindBoost(intent assistantIntent, kind string) int {
	switch intent {
	case assistantIntentOnboarding:
		switch kind {
		case "page":
			return 90
		case "group", "event":
			return 60
		case "tag":
			return 25
		default:
			return 0
		}
	case assistantIntentGroups:
		switch kind {
		case "group":
			return 100
		case "user":
			return 35
		case "page":
			return 25
		default:
			return 0
		}
	case assistantIntentEvents:
		switch kind {
		case "event":
			return 100
		case "group":
			return 30
		case "page":
			return 25
		default:
			return 0
		}
	case assistantIntentPosting:
		switch kind {
		case "page":
			return 80
		case "post":
			return 70
		case "tag":
			return 55
		case "user":
			return 20
		default:
			return 0
		}
	case assistantIntentUsers:
		switch kind {
		case "user":
			return 95
		case "post":
			return 45
		case "group":
			return 25
		default:
			return 0
		}
	case assistantIntentContent:
		switch kind {
		case "post":
			return 95
		case "tag":
			return 55
		case "user":
			return 35
		default:
			return 0
		}
	default:
		return 0
	}
}

func pageIntentBoost(intent assistantIntent, card AssistantCard) int {
	if card.Kind != "page" {
		return 0
	}
	switch intent {
	case assistantIntentOnboarding:
		switch card.Href {
		case "/explore", "/groups", "/events":
			return 60
		case "/feed":
			return 35
		default:
			return 0
		}
	case assistantIntentPosting:
		switch card.Href {
		case "/posts/create":
			return 80
		case "/creator":
			return 35
		default:
			return 0
		}
	case assistantIntentGroups:
		if card.Href == "/groups" {
			return 70
		}
	case assistantIntentEvents:
		if card.Href == "/events" {
			return 70
		}
	}
	return 0
}

func assistantSourceCounts(cards []AssistantCard) map[string]int {
	if len(cards) == 0 {
		return nil
	}
	counts := make(map[string]int, len(cards))
	for _, card := range cards {
		if card.Kind == "" {
			continue
		}
		counts[card.Kind]++
	}
	return counts
}

func buildAssistantPageContext(pageContext *AssistantPageContext) string {
	if pageContext == nil {
		return ""
	}

	var lines []string
	if title := strings.TrimSpace(pageContext.Title); title != "" {
		lines = append(lines, fmt.Sprintf("- 当前页面：%s", title))
	}
	if kind := strings.TrimSpace(pageContext.Kind); kind != "" {
		lines = append(lines, fmt.Sprintf("- 页面场景：%s", kind))
	}
	if path := strings.TrimSpace(pageContext.Path); path != "" {
		lines = append(lines, fmt.Sprintf("- 页面路径：%s", path))
	}
	if summary := strings.TrimSpace(pageContext.Summary); summary != "" {
		lines = append(lines, fmt.Sprintf("- 页面说明：%s", summary))
	}
	if len(pageContext.PromptHints) > 0 {
		lines = append(lines, fmt.Sprintf("- 当前页面适合的问题：%s", strings.Join(pageContext.PromptHints, "；")))
	}
	if len(pageContext.Fields) > 0 {
		keys := make([]string, 0, len(pageContext.Fields))
		for key := range pageContext.Fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := strings.TrimSpace(pageContext.Fields[key])
			if value == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("- %s：%s", assistantPageFieldLabel(key), value))
		}
	}

	if len(lines) == 0 {
		return ""
	}
	return "当前页面上下文：\n" + strings.Join(lines, "\n")
}

func assistantPageFieldLabel(key string) string {
	switch key {
	case "draft_title":
		return "当前草稿标题"
	case "draft_content":
		return "当前草稿内容"
	case "draft_tags":
		return "当前草稿标签"
	case "group_name":
		return "目标圈子"
	case "group_privacy":
		return "圈子可见性"
	case "member_count":
		return "圈子成员数"
	case "post_count":
		return "圈子帖子数"
	case "group_tags":
		return "圈子标签"
	case "group_description":
		return "圈子简介"
	case "group_rules":
		return "圈子规则"
	case "group_announcement":
		return "圈子公告"
	case "featured_post":
		return "圈子精选内容"
	case "highlights_count":
		return "圈子精选数量"
	case "is_member":
		return "当前用户是否已加入"
	case "visibility":
		return "当前可见性"
	case "ai_generated":
		return "是否勾选 AI 生成标记"
	case "image_tags":
		return "图片建议标签"
	case "image_alt_notes":
		return "图片摘要"
	case "event_status":
		return "活动状态"
	case "event_time":
		return "活动时间"
	case "event_location":
		return "活动地点"
	case "event_attendees":
		return "活动参与人数"
	case "event_tags":
		return "活动标签"
	case "event_description":
		return "活动简介"
	case "has_attended":
		return "当前用户是否已报名"
	default:
		return key
	}
}

func queryTokens(query string) []string {
	parts := strings.FieldsFunc(query, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', ',', '，', '.', '。', '/', '|', '-', '_', '、', ':', '：', '(', ')', '（', '）':
			return true
		default:
			return false
		}
	})

	tokens := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if utf8.RuneCountInString(part) < 2 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		tokens = append(tokens, part)
	}
	return tokens
}
