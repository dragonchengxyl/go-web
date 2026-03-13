package usecase

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/llm"
	"github.com/studio/platform/internal/pkg/apperr"
)

// AssistantChatMessage is the user/assistant message payload exchanged with the frontend.
type AssistantChatMessage struct {
	ID        string          `json:"id,omitempty"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Cards     []AssistantCard `json:"cards,omitempty"`
	CreatedAt time.Time       `json:"created_at,omitempty"`
}

// AssistantCard is a structured recommendation displayed next to the assistant reply.
type AssistantCard = assistantdomain.Card

// AssistantMeta is sent before the streamed answer so the UI can render recommendations early.
type AssistantMeta struct {
	Query          string          `json:"query"`
	Provider       string          `json:"provider"`
	Fallback       bool            `json:"fallback"`
	ConversationID string          `json:"conversation_id,omitempty"`
	Cards          []AssistantCard `json:"cards"`
}

// AssistantService powers the site AI helper.
type AssistantService struct {
	cfg          configs.AssistantConfig
	llmClient    *llm.OpenAICompatibleClient
	historyRepo  assistantdomain.Repository
	postService  *PostService
	groupService *GroupService
	eventService *EventService
	userService  *UserService
}

// NewAssistantService creates a lightweight assistant service.
func NewAssistantService(
	cfg configs.AssistantConfig,
	llmClient *llm.OpenAICompatibleClient,
	historyRepo assistantdomain.Repository,
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

	return &AssistantService{
		cfg:          cfg,
		llmClient:    llmClient,
		historyRepo:  historyRepo,
		postService:  postService,
		groupService: groupService,
		eventService: eventService,
		userService:  userService,
	}
}

// HistoryEnabled reports whether server-side assistant persistence is available.
func (s *AssistantService) HistoryEnabled() bool {
	return s != nil && s.historyRepo != nil
}

// StreamReply streams a response for the provided conversation history.
func (s *AssistantService) StreamReply(
	ctx context.Context,
	messages []AssistantChatMessage,
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

	meta, contextText, fallbackAnswer := s.buildPromptContext(ctx, latestUser, settings)
	if onMeta != nil {
		if err := onMeta(meta); err != nil {
			return err
		}
	}

	if s.llmClient == nil || !s.llmClient.Configured() {
		return streamText(fallbackAnswer, onToken)
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
	err = s.llmClient.StreamChat(ctx, llmMessages, func(token string) error {
		streamedAny = true
		return onToken(token)
	})
	if err == nil {
		return nil
	}
	if streamedAny {
		return err
	}

	meta.Fallback = true
	if onMeta != nil {
		if err := onMeta(meta); err != nil {
			return err
		}
	}
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
			Role:    string(msg.Role),
			Content: msg.Content,
			Cards:   msg.Cards,
		})
	}
	return conv, messages, nil
}

// SaveAssistantReply persists the assistant answer for a conversation.
func (s *AssistantService) SaveAssistantReply(ctx context.Context, conversationID uuid.UUID, content string, cards []AssistantCard) error {
	if !s.HistoryEnabled() || conversationID == uuid.Nil || strings.TrimSpace(content) == "" {
		return nil
	}
	if err := s.historyRepo.CreateMessage(ctx, &assistantdomain.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           assistantdomain.RoleAssistant,
		Content:        content,
		Cards:          cards,
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

func (s *AssistantService) buildPromptContext(ctx context.Context, query string, settings *assistantdomain.Settings) (AssistantMeta, string, string) {
	cards := s.collectCards(ctx, query, settings)
	meta := AssistantMeta{
		Query:    query,
		Provider: s.cfg.Provider,
		Fallback: s.llmClient == nil || !s.llmClient.Configured(),
		Cards:    cards,
	}

	var contextParts []string
	contextParts = append(contextParts, fmt.Sprintf("当前日期: %s", time.Now().Format("2006-01-02")))
	contextParts = append(contextParts, siteOverviewContext())
	if len(cards) > 0 {
		var itemLines []string
		for i, card := range cards {
			line := fmt.Sprintf("%d. [%s] %s - %s (链接: %s)", i+1, card.Kind, card.Title, card.Summary, card.Href)
			if card.Meta != "" {
				line += " | " + card.Meta
			}
			itemLines = append(itemLines, line)
		}
		contextParts = append(contextParts, "可引用的站内信息:\n"+strings.Join(itemLines, "\n"))
	}

	return meta, strings.Join(contextParts, "\n\n"), buildFallbackAnswer(settings.PersonaName, query, cards)
}

func (s *AssistantService) collectCards(ctx context.Context, query string, settings *assistantdomain.Settings) []AssistantCard {
	maxItems := settings.MaxContextItems
	if maxItems <= 0 {
		maxItems = 6
	}

	seen := make(map[string]struct{}, maxItems)
	cards := make([]AssistantCard, 0, maxItems)
	appendUnique := func(items ...AssistantCard) {
		for _, item := range items {
			if len(cards) >= maxItems {
				return
			}
			if item.Href == "" {
				continue
			}
			if _, ok := seen[item.Href]; ok {
				continue
			}
			seen[item.Href] = struct{}{}
			cards = append(cards, item)
		}
	}

	if settings.IncludePages {
		appendUnique(recommendPageCards(query)...)
	}
	if settings.IncludePosts {
		appendUnique(s.collectPostCards(ctx, query)...)
	}
	if settings.IncludeUsers {
		appendUnique(s.collectUserCards(ctx, query)...)
	}
	if settings.IncludeTags {
		appendUnique(s.collectTagCards(ctx, query)...)
	}
	if settings.IncludeGroups {
		appendUnique(s.collectGroupCards(ctx, query)...)
	}
	if settings.IncludeEvents {
		appendUnique(s.collectEventCards(ctx, query)...)
	}

	return cards
}

func (s *AssistantService) collectUserCards(ctx context.Context, query string) []AssistantCard {
	if s.userService == nil {
		return nil
	}

	users, err := s.userService.SearchUsers(ctx, query, 2)
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
		})
	}
	return cards
}

func (s *AssistantService) collectTagCards(ctx context.Context, query string) []AssistantCard {
	if s.postService == nil {
		return nil
	}

	tags, err := s.postService.GetHotTags(ctx, 12)
	if err != nil || len(tags) == 0 {
		return nil
	}

	query = strings.TrimSpace(strings.ToLower(query))
	cards := make([]AssistantCard, 0, 2)
	for _, tag := range tags {
		if len(cards) >= 2 {
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
		})
	}

	if len(cards) == 0 {
		for _, tag := range tags[:min(2, len(tags))] {
			cards = append(cards, AssistantCard{
				Kind:    "tag",
				Title:   "#" + tag,
				Summary: "查看这个标签下的相关动态。",
				Href:    "/tags/" + url.PathEscape(tag),
				Meta:    "/tags/" + tag,
			})
		}
	}
	return cards
}

func (s *AssistantService) collectPostCards(ctx context.Context, query string) []AssistantCard {
	if s.postService == nil {
		return nil
	}

	posts, err := s.postService.SearchPosts(ctx, query, 2)
	if err != nil || len(posts) == 0 {
		posts, _, _ = s.postService.ListExplore(ctx, 1, 2, "")
	}

	cards := make([]AssistantCard, 0, len(posts))
	for _, post := range posts {
		title := strings.TrimSpace(post.Title)
		if title == "" {
			title = truncateText(post.Content, 18)
		}
		summary := truncateText(post.Content, 56)
		meta := fmt.Sprintf("@%s · %d 赞 · %d 评论", post.AuthorUsername, post.LikeCount, post.CommentCount)
		cards = append(cards, AssistantCard{
			Kind:    "post",
			Title:   title,
			Summary: summary,
			Href:    "/posts/" + post.ID.String(),
			Meta:    meta,
		})
	}
	return cards
}

func (s *AssistantService) collectGroupCards(ctx context.Context, query string) []AssistantCard {
	if s.groupService == nil {
		return nil
	}

	privacy := group.GroupPrivacyPublic
	groups, _, err := s.groupService.ListGroups(ctx, ListGroupsInput{
		Privacy:  &privacy,
		Search:   strings.TrimSpace(query),
		Page:     1,
		PageSize: 2,
	})
	if err != nil || len(groups) == 0 {
		groups, _, _ = s.groupService.ListGroups(ctx, ListGroupsInput{
			Privacy:  &privacy,
			Page:     1,
			PageSize: 2,
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
		})
	}
	return cards
}

func (s *AssistantService) collectEventCards(ctx context.Context, query string) []AssistantCard {
	if s.eventService == nil {
		return nil
	}

	status := event.EventStatusPublished
	events, _, err := s.eventService.ListEvents(ctx, ListEventsInput{
		Status:   &status,
		Page:     1,
		PageSize: 6,
	})
	if err != nil || len(events) == 0 {
		return nil
	}

	filtered := filterEvents(events, query)
	if len(filtered) == 0 {
		filtered = events
	}
	if len(filtered) > 2 {
		filtered = filtered[:2]
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

func buildFallbackAnswer(personaName, query string, cards []AssistantCard) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s 在这。你刚才问的是“%s”。\n\n", personaName, query)

	if len(cards) == 0 {
		b.WriteString("我先给你一个站内导航建议：如果你是第一次来，建议先看“发现页 /explore”、再逛“圈子 /groups”和“活动 /events”，想发内容就去“/posts/create”。")
		return b.String()
	}

	b.WriteString("我先根据站内信息帮你整理了几个值得直接点开的入口：\n")
	for i, card := range cards {
		fmt.Fprintf(&b, "%d. %s：%s", i+1, card.Title, card.Summary)
		if card.Meta != "" {
			fmt.Fprintf(&b, "（%s）", card.Meta)
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
			Role:    role,
			Content: truncateText(content, 1200),
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

func recommendPageCards(query string) []AssistantCard {
	pages := []AssistantCard{
		{
			Kind:    "page",
			Title:   "发现页",
			Summary: "看热门动态、标签和创作者，适合第一次来先逛。",
			Href:    "/explore",
			Meta:    "/explore",
		},
		{
			Kind:    "page",
			Title:   "关注动态",
			Summary: "查看你关注对象的最新内容和互动。",
			Href:    "/feed",
			Meta:    "/feed",
		},
		{
			Kind:    "page",
			Title:   "发布动态",
			Summary: "支持图文发布、图片上传、AI 内容标记和可见性设置。",
			Href:    "/posts/create",
			Meta:    "/posts/create",
		},
		{
			Kind:    "page",
			Title:   "圈子广场",
			Summary: "按兴趣找同好、加入圈子、看成员和帖子数。",
			Href:    "/groups",
			Meta:    "/groups",
		},
		{
			Kind:    "page",
			Title:   "活动广场",
			Summary: "查看近期线上线下活动，支持报名参加。",
			Href:    "/events",
			Meta:    "/events",
		},
		{
			Kind:    "page",
			Title:   "创作者面板",
			Summary: "查看帖子、粉丝、互动和打赏数据。",
			Href:    "/creator",
			Meta:    "/creator",
		},
	}

	type keywordRule struct {
		keywords []string
		indexes  []int
	}
	rules := []keywordRule{
		{keywords: []string{"发帖", "发布", "创作", "作品", "动态"}, indexes: []int{2, 5}},
		{keywords: []string{"圈子", "社群", "同好", "群组"}, indexes: []int{3, 0}},
		{keywords: []string{"活动", "聚会", "线下", "线上"}, indexes: []int{4, 0}},
		{keywords: []string{"第一次", "新手", "怎么逛", "先看"}, indexes: []int{0, 1}},
		{keywords: []string{"数据", "创作者", "收益", "打赏"}, indexes: []int{5, 2}},
	}

	query = strings.ToLower(strings.TrimSpace(query))
	selected := make([]AssistantCard, 0, 2)
	seen := map[int]struct{}{}
	appendPage := func(idx int) {
		if len(selected) >= 2 {
			return
		}
		if _, ok := seen[idx]; ok {
			return
		}
		seen[idx] = struct{}{}
		selected = append(selected, pages[idx])
	}

	for _, rule := range rules {
		matched := false
		for _, keyword := range rule.keywords {
			if strings.Contains(query, keyword) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		for _, idx := range rule.indexes {
			appendPage(idx)
		}
	}

	if len(selected) == 0 {
		appendPage(0)
		appendPage(3)
	}
	return selected
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
