package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/follow"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/domain/report"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/llm"
	"github.com/studio/platform/internal/observability/assistantmetrics"
	"github.com/studio/platform/internal/pkg/resilience"
)

// AdminAIToolSection is a structured section in a tool result.
type AdminAIToolSection struct {
	Title   string   `json:"title"`
	Bullets []string `json:"bullets,omitempty"`
}

// AdminAIToolDraft is a copy-ready generated draft.
type AdminAIToolDraft struct {
	Label   string `json:"label"`
	Content string `json:"content"`
}

// AdminAIToolResult is the unified output shape for internal admin AI tools.
type AdminAIToolResult struct {
	RunID       string               `json:"run_id"`
	Tool        string               `json:"tool"`
	Title       string               `json:"title"`
	Summary     string               `json:"summary"`
	Sections    []AdminAIToolSection `json:"sections,omitempty"`
	Drafts      []AdminAIToolDraft   `json:"drafts,omitempty"`
	Fallback    bool                 `json:"fallback"`
	Provider    string               `json:"provider"`
	GeneratedAt time.Time            `json:"generated_at"`
}

// AdminAIToolService powers internal admin-facing AI tools.
type AdminAIToolService struct {
	cfg            configs.AssistantConfig
	llmClient      *llm.OpenAICompatibleClient
	statsService   StatsProvider
	reportRepo     report.Repository
	postService    *PostService
	commentService *CommentService
	userService    *UserService
	followService  *FollowService
	tipService     *TipService
	groupService   *GroupService
	eventService   *EventService
	llmCircuit     *resilience.CircuitBreaker
}

func NewAdminAIToolService(
	cfg configs.AssistantConfig,
	llmClient *llm.OpenAICompatibleClient,
	statsService StatsProvider,
	reportRepo report.Repository,
	postService *PostService,
	commentService *CommentService,
	userService *UserService,
	followService *FollowService,
	tipService *TipService,
	groupService *GroupService,
	eventService *EventService,
) *AdminAIToolService {
	if cfg.Provider == "" {
		cfg.Provider = "deepseek"
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
	return &AdminAIToolService{
		cfg:            cfg,
		llmClient:      llmClient,
		statsService:   statsService,
		reportRepo:     reportRepo,
		postService:    postService,
		commentService: commentService,
		userService:    userService,
		followService:  followService,
		tipService:     tipService,
		groupService:   groupService,
		eventService:   eventService,
		llmCircuit:     resilience.NewCircuitBreaker(cfg.CircuitFailures, time.Duration(cfg.CircuitOpenSec)*time.Second),
	}
}

func (s *AdminAIToolService) GenerateReportSummary(ctx context.Context, reportID uuid.UUID) (*AdminAIToolResult, error) {
	rep, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	targetSummary, targetBullets := s.describeReportTarget(ctx, rep)
	riskLevel, riskBullets, recommendedAction := assessReportRisk(rep, targetSummary)
	sections := []AdminAIToolSection{
		{
			Title: "举报概览",
			Bullets: uniqueTrimmed([]string{
				"举报类型：" + string(rep.TargetType),
				"举报理由：" + firstNonEmpty(rep.Reason, "未填写"),
				"状态：" + string(rep.Status),
				"举报人：" + firstNonEmpty(rep.ReporterUsername, rep.ReporterID.String()),
				"创建时间：" + rep.CreatedAt.Format("2006-01-02 15:04"),
			}, 5),
		},
		{
			Title:   "目标对象摘要",
			Bullets: append([]string{targetSummary}, targetBullets...),
		},
		{
			Title:   "风险判断",
			Bullets: append([]string{"风险等级：" + riskLevel}, riskBullets...),
		},
		{
			Title: "建议动作",
			Bullets: []string{
				"建议优先动作：" + recommendedAction,
				"先确认举报描述与目标内容是否直接对应，再做最终处置",
			},
		},
	}

	reviewerFallback := buildReportReviewerDraft(rep, riskLevel, recommendedAction, targetSummary)
	reporterFallback := buildReportReporterReply(rep, recommendedAction)
	reviewerDraft, reviewerFallbackUsed := s.generateDraft(
		ctx,
		"你是运营后台的审核助手。请根据已给定事实，写一段给审核员看的中文处理建议。要求克制、具体、不要编造事实。",
		buildToolPrompt("举报处理建议", sections),
		reviewerFallback,
	)
	reporterDraft, reporterFallbackUsed := s.generateDraft(
		ctx,
		"你是社区运营助手。请根据已给定事实，写一段发给举报用户的中文通知文案。要求礼貌、清晰、不泄露内部决策细节。",
		buildToolPrompt("举报处理通知", sections),
		reporterFallback,
	)

	return &AdminAIToolResult{
		RunID:    uuid.NewString(),
		Tool:     "report_summary",
		Title:    "举报摘要与处理建议",
		Summary:  fmt.Sprintf("围绕举报 #%s 输出了风险判断、建议动作和两段可直接复用的说明文案。", rep.ID.String()[:8]),
		Sections: sections,
		Drafts: []AdminAIToolDraft{
			{Label: "审核员处理建议", Content: reviewerDraft},
			{Label: "给举报人的回复草稿", Content: reporterDraft},
		},
		Fallback:    reviewerFallbackUsed || reporterFallbackUsed,
		Provider:    s.cfg.Provider,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *AdminAIToolService) GenerateWeeklyReport(ctx context.Context, days int) (*AdminAIToolResult, error) {
	if days <= 0 || days > 30 {
		days = 7
	}

	stats, err := s.statsService.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	growth, _ := s.statsService.GetUserGrowthChart(ctx, days)
	posts, _, _ := s.postService.ListExplore(ctx, 1, 5, "")
	privacy := groupPublicPrivacyPtr()
	groups, _, _ := s.groupService.ListGroups(ctx, ListGroupsInput{
		Privacy:  privacy,
		Page:     1,
		PageSize: 3,
	})
	status := event.EventStatusPublished
	events, _, _ := s.eventService.ListEvents(ctx, ListEventsInput{
		Status:   &status,
		Page:     1,
		PageSize: 3,
	})
	reports, _, _ := s.reportRepo.List(ctx, string(report.StatusPending), 1, 5)

	sections := []AdminAIToolSection{
		{
			Title: "核心指标",
			Bullets: []string{
				fmt.Sprintf("累计用户：%d，本日新增：%d", stats.TotalUsers, stats.NewUsersToday),
				fmt.Sprintf("累计帖子：%d，当前待处理举报：%d", stats.TotalPosts, stats.TotalReports),
				buildGrowthSummary(growth),
			},
		},
		{
			Title:   "内容亮点",
			Bullets: summarizeTopPosts(posts),
		},
		{
			Title:   "圈子与活动观察",
			Bullets: append(summarizeTopGroups(groups), summarizeUpcomingEvents(events)...),
		},
		{
			Title:   "风险与跟进建议",
			Bullets: buildWeeklyRiskBullets(stats, reports, events),
		},
	}

	fallbackBody := buildWeeklyReportFallback(days, sections)
	reportDraft, fallbackUsed := s.generateDraft(
		ctx,
		"你是社区运营周报助手。请根据已给定事实，输出一份中文周报正文，结构清晰，适合直接发给产品或运营同事。",
		buildToolPrompt(fmt.Sprintf("%d天社区周报", days), sections),
		fallbackBody,
	)

	return &AdminAIToolResult{
		RunID:       uuid.NewString(),
		Tool:        "weekly_report",
		Title:       "社区运营周报生成",
		Summary:     fmt.Sprintf("汇总了最近 %d 天的核心指标、内容亮点、圈子活动观察与风险建议。", days),
		Sections:    sections,
		Drafts:      []AdminAIToolDraft{{Label: "周报正文", Content: reportDraft}},
		Fallback:    fallbackUsed,
		Provider:    s.cfg.Provider,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *AdminAIToolService) GenerateCreatorRecommendation(ctx context.Context, userID uuid.UUID) (*AdminAIToolResult, error) {
	creator, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	vis := visibilityPublicPtr()
	posts, totalPosts, _ := s.postService.ListUserPosts(ctx, ListUserPostsInput{
		AuthorID:   userID,
		Visibility: vis,
		Page:       1,
		PageSize:   5,
	})
	followStats, _ := s.followService.GetStats(ctx, userID)
	tipTotal := int64(0)
	tipCount := 0
	if s.tipService != nil {
		tipTotal, tipCount, _ = s.tipService.GetMyTipStats(ctx, userID)
	}

	totalLikes, totalComments, sampleTitles := summarizeCreatorPosts(posts)
	sections := []AdminAIToolSection{
		{
			Title: "创作者画像",
			Bullets: uniqueTrimmed([]string{
				"账号：" + creator.Username,
				"角色：" + string(creator.Role),
				optionalBullet("兽设名", creator.FurryName),
				optionalBullet("物种", creator.Species),
				optionalBullet("简介", creator.Bio),
			}, 5),
		},
		{
			Title: "推荐理由",
			Bullets: uniqueTrimmed([]string{
				fmt.Sprintf("公开内容数：%d，样本帖子互动：%d 赞 / %d 评论", totalPosts, totalLikes, totalComments),
				fmt.Sprintf("关注关系：%d 粉丝 / %d 关注", followerCount(followStats), followingCount(followStats)),
				fmt.Sprintf("打赏数据：%s 元，共 %d 笔", FormatAmount(int(tipTotal)), tipCount),
				optionalListBullet("代表内容", sampleTitles),
			}, 4),
		},
		{
			Title:   "适合的运营场景",
			Bullets: buildCreatorOpsBullets(totalPosts, totalLikes, totalComments, tipCount, sampleTitles),
		},
	}

	shortFallback := buildCreatorShortCopy(creator, sampleTitles, followerCount(followStats), totalLikes, totalComments)
	longFallback := buildCreatorLongCopy(creator, sampleTitles, totalPosts, totalLikes, totalComments, tipCount)
	shortDraft, shortFallbackUsed := s.generateDraft(
		ctx,
		"你是创作者推荐助手。请根据已给定事实，写一段适合后台运营直接复用的中文短推荐文案，80字以内。",
		buildToolPrompt("创作者短推荐", sections),
		shortFallback,
	)
	longDraft, longFallbackUsed := s.generateDraft(
		ctx,
		"你是创作者推荐助手。请根据已给定事实，写一段适合首页、专题位或活动页使用的中文推荐理由，120到180字。",
		buildToolPrompt("创作者长推荐", sections),
		longFallback,
	)

	return &AdminAIToolResult{
		RunID:    uuid.NewString(),
		Tool:     "creator_recommendation",
		Title:    "创作者推荐理由生成",
		Summary:  fmt.Sprintf("为 @%s 生成了推荐理由、适用运营场景和两版可直接复用的推荐文案。", creator.Username),
		Sections: sections,
		Drafts: []AdminAIToolDraft{
			{Label: "短推荐文案", Content: shortDraft},
			{Label: "长推荐文案", Content: longDraft},
		},
		Fallback:    shortFallbackUsed || longFallbackUsed,
		Provider:    s.cfg.Provider,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *AdminAIToolService) GenerateEventCopy(ctx context.Context, eventID uuid.UUID) (*AdminAIToolResult, error) {
	item, err := s.eventService.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	location := firstNonEmpty(item.Location, "未填写地点")
	if item.IsOnline {
		location = "线上活动"
	}

	sections := []AdminAIToolSection{
		{
			Title: "活动亮点",
			Bullets: uniqueTrimmed([]string{
				"标题：" + item.Title,
				"时间：" + item.StartTime.Format("2006-01-02 15:04") + " - " + item.EndTime.Format("15:04"),
				"地点：" + location,
				fmt.Sprintf("人数：%d / %d", item.AttendeeCount, item.MaxCapacity),
				optionalListBullet("标签", item.Tags),
				"简介重点：" + truncateText(item.Description, 80),
			}, 6),
		},
		{
			Title:   "宣传卖点",
			Bullets: buildEventSellingPoints(item, location),
		},
		{
			Title:   "发布前检查",
			Bullets: buildEventPublishChecklist(item, location),
		},
	}

	shortFallback := buildEventShortCopy(item, location)
	longFallback := buildEventLongCopy(item, location)
	shortDraft, shortFallbackUsed := s.generateDraft(
		ctx,
		"你是活动运营文案助手。请根据已给定事实，写一段适合活动卡片、海报角标或列表页使用的中文短文案，80字以内。",
		buildToolPrompt("活动短文案", sections),
		shortFallback,
	)
	longDraft, longFallbackUsed := s.generateDraft(
		ctx,
		"你是活动运营文案助手。请根据已给定事实，写一段适合活动详情页或社区推送使用的中文宣传文案，120到180字。",
		buildToolPrompt("活动长文案", sections),
		longFallback,
	)

	return &AdminAIToolResult{
		RunID:    uuid.NewString(),
		Tool:     "event_copy",
		Title:    "活动推荐文案生成",
		Summary:  fmt.Sprintf("为活动“%s”生成了宣传卖点、发布前检查项和两版活动文案。", item.Title),
		Sections: sections,
		Drafts: []AdminAIToolDraft{
			{Label: "短文案", Content: shortDraft},
			{Label: "长文案", Content: longDraft},
		},
		Fallback:    shortFallbackUsed || longFallbackUsed,
		Provider:    s.cfg.Provider,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *AdminAIToolService) generateDraft(ctx context.Context, systemPrompt, userPrompt, fallback string) (string, bool) {
	if s.llmClient == nil || !s.llmClient.Configured() {
		return fallback, true
	}
	if s.llmCircuit != nil {
		assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
		if !s.llmCircuit.Allow() {
			return fallback, true
		}
	}

	maxAttempts := max(s.cfg.LLMRetryMax, 1)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var reply strings.Builder
		streamed := false
		err := s.llmClient.StreamChat(ctx, []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		}, func(token string) error {
			streamed = true
			reply.WriteString(token)
			return nil
		})
		if err == nil && strings.TrimSpace(reply.String()) != "" {
			if s.llmCircuit != nil {
				s.llmCircuit.RecordSuccess()
				assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
			}
			return strings.TrimSpace(reply.String()), false
		}
		if s.llmCircuit != nil {
			s.llmCircuit.RecordFailure(err)
			assistantmetrics.RecordCircuitState("chat", string(s.llmCircuit.Snapshot().State))
		}
		if streamed {
			break
		}
		if attempt < maxAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
	}
	return fallback, true
}

func (s *AdminAIToolService) describeReportTarget(ctx context.Context, rep *report.Report) (string, []string) {
	switch rep.TargetType {
	case report.TargetTypePost:
		if s.postService != nil {
			if item, err := s.postService.GetPost(ctx, rep.TargetID); err == nil && item != nil {
				return "目标是帖子内容。", uniqueTrimmed([]string{
					"作者：@" + firstNonEmpty(item.AuthorUsername, item.AuthorID.String()),
					optionalInline("标题", firstNonEmpty(item.Title, "")),
					"内容摘要：" + truncateText(item.Content, 64),
					optionalListBullet("标签", item.Tags),
				}, 4)
			}
		}
	case report.TargetTypeComment:
		if s.commentService != nil {
			if item, err := s.commentService.commentRepo.GetByID(ctx, rep.TargetID); err == nil && item != nil {
				return "目标是评论内容。", uniqueTrimmed([]string{
					"评论作者 ID：" + item.UserID.String(),
					"评论摘要：" + truncateText(item.Content, 64),
					fmt.Sprintf("点赞 %d / 回复 %d", item.LikeCount, item.ReplyCount),
				}, 3)
			}
		}
	case report.TargetTypeUser:
		if s.userService != nil {
			if item, err := s.userService.GetUserByID(ctx, rep.TargetID); err == nil && item != nil {
				return "目标是用户账号。", uniqueTrimmed([]string{
					"用户名：@" + item.Username,
					"角色：" + string(item.Role),
					optionalBullet("简介", item.Bio),
					optionalBullet("兽设名", item.FurryName),
				}, 4)
			}
		}
	}
	return "目标对象详情暂未取到，建议审核时回到原始内容核对。", []string{"目标 ID：" + rep.TargetID.String()}
}

func assessReportRisk(rep *report.Report, targetSummary string) (string, []string, string) {
	text := strings.ToLower(rep.Reason + " " + rep.Description + " " + targetSummary)
	highSignals := []string{"骚扰", "辱骂", "仇恨", "色情", "未成年", "泄露", "人身攻击", "威胁", "诈骗", "冒充"}
	mediumSignals := []string{"广告", "刷屏", "引战", "低俗", "不适", "争议", "搬运", "抄袭"}
	action := "人工复核后再决定"
	level := "中"
	bullets := []string{}

	for _, signal := range highSignals {
		if strings.Contains(text, signal) {
			level = "高"
			bullets = append(bullets, "命中高风险信号："+signal)
		}
	}
	if level != "高" {
		for _, signal := range mediumSignals {
			if strings.Contains(text, signal) {
				bullets = append(bullets, "命中中风险信号："+signal)
			}
		}
	}

	switch rep.TargetType {
	case report.TargetTypePost:
		if level == "高" {
			action = "优先考虑临时屏蔽帖子，再补做人工复核"
		} else {
			action = "建议先人工复核帖子语境，确认是否需要 block_post"
		}
	case report.TargetTypeComment:
		if level == "高" {
			action = "优先考虑删除评论，并检查是否需要继续处理账号"
		} else {
			action = "建议先人工复核评论语气和上下文，再决定是否 delete_comment"
		}
	case report.TargetTypeUser:
		if level == "高" {
			action = "优先调查账号历史行为，必要时再考虑 ban_user"
		} else {
			action = "建议先回看该用户近期行为，再决定是否升级处罚"
		}
	}

	if len(bullets) == 0 {
		bullets = append(bullets, "当前举报描述偏概括，建议结合原始内容再判断")
	}
	return level, uniqueTrimmed(bullets, 4), action
}

func buildReportReviewerDraft(rep *report.Report, riskLevel, recommendedAction, targetSummary string) string {
	return strings.TrimSpace(fmt.Sprintf(
		"这条举报目前可归类为%s风险。举报理由是“%s”，目标对象摘要为：%s。建议%s，并在处理前再次核对原始内容与举报描述是否直接对应。",
		riskLevel,
		firstNonEmpty(rep.Reason, "未填写"),
		targetSummary,
		recommendedAction,
	))
}

func buildReportReporterReply(rep *report.Report, recommendedAction string) string {
	return strings.TrimSpace(fmt.Sprintf(
		"你好，已收到你关于%s内容的举报。我们会结合站内规则和原始内容进行复核，并按流程处理。当前建议方向是：%s。感谢你帮助维护社区环境。",
		rep.TargetType,
		recommendedAction,
	))
}

func buildGrowthSummary(points []ChartPoint) string {
	if len(points) == 0 {
		return "近期开启了周报统计，但暂时没有足够的增长点数据。"
	}
	first := points[0]
	last := points[len(points)-1]
	return fmt.Sprintf("最近增长曲线从 %s 的 %.0f 到 %s 的 %.0f。", first.Date, first.Value, last.Date, last.Value)
}

func summarizeTopPosts(posts []*post.Post) []string {
	if len(posts) == 0 {
		return []string{"本周期没有可用的热门帖子样本。"}
	}
	out := make([]string, 0, min(4, len(posts)))
	for _, item := range posts[:min(4, len(posts))] {
		title := truncateText(firstNonEmpty(item.Title, item.Content), 26)
		out = append(out, fmt.Sprintf("%s：%d 赞 / %d 评论", title, item.LikeCount, item.CommentCount))
	}
	return out
}

func summarizeTopGroups(groups []*group.Group) []string {
	if len(groups) == 0 {
		return []string{"本周期没有可用的圈子样本。"}
	}
	out := make([]string, 0, min(3, len(groups)))
	for _, item := range groups[:min(3, len(groups))] {
		out = append(out, fmt.Sprintf("%s：%d 成员 / %d 帖子", item.Name, item.MemberCount, item.PostCount))
	}
	return out
}

func summarizeUpcomingEvents(events []*event.Event) []string {
	if len(events) == 0 {
		return []string{"近期没有公开活动样本。"}
	}
	out := make([]string, 0, min(3, len(events)))
	for _, item := range events[:min(3, len(events))] {
		location := firstNonEmpty(item.Location, "未填写地点")
		if item.IsOnline {
			location = "线上活动"
		}
		out = append(out, fmt.Sprintf("%s：%s · %s", item.Title, item.StartTime.Format("01-02 15:04"), location))
	}
	return out
}

func buildWeeklyRiskBullets(stats *DashboardStats, reports []*report.Report, events []*event.Event) []string {
	bullets := make([]string, 0, 4)
	if stats.TotalReports > 0 {
		bullets = append(bullets, fmt.Sprintf("当前仍有 %d 条待处理举报，建议优先清空高风险项。", stats.TotalReports))
	}
	if len(reports) > 0 {
		bullets = append(bullets, "待处理举报里优先关注："+firstNonEmpty(reports[0].Reason, string(reports[0].TargetType)))
	}
	if len(events) == 0 {
		bullets = append(bullets, "近期公开活动偏少，可以提前准备下一轮活动排期。")
	} else {
		bullets = append(bullets, "可优先围绕近期活动做内容预热和社区联动。")
	}
	if len(bullets) == 0 {
		bullets = append(bullets, "当前风险信号不强，重点保持内容和活动节奏。")
	}
	return uniqueTrimmed(bullets, 4)
}

func buildWeeklyReportFallback(days int, sections []AdminAIToolSection) string {
	var b strings.Builder
	fmt.Fprintf(&b, "以下是最近 %d 天的社区运营周报：\n", days)
	for _, section := range sections {
		fmt.Fprintf(&b, "\n## %s\n", section.Title)
		for _, bullet := range section.Bullets {
			if strings.TrimSpace(bullet) != "" {
				b.WriteString("- ")
				b.WriteString(strings.TrimSpace(bullet))
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func summarizeCreatorPosts(posts []*post.Post) (int64, int64, []string) {
	var totalLikes int64
	var totalComments int64
	titles := make([]string, 0, min(3, len(posts)))
	for _, item := range posts {
		totalLikes += int64(item.LikeCount)
		totalComments += int64(item.CommentCount)
		if len(titles) < 3 {
			titles = append(titles, truncateText(firstNonEmpty(item.Title, item.Content), 24))
		}
	}
	return totalLikes, totalComments, titles
}

func buildCreatorOpsBullets(totalPosts int64, totalLikes, totalComments int64, tipCount int, sampleTitles []string) []string {
	bullets := []string{}
	if totalPosts >= 3 {
		bullets = append(bullets, "适合放进首页或发现页的创作者推荐位")
	}
	if totalLikes+totalComments >= 20 {
		bullets = append(bullets, "适合放进高互动创作者专题或活动联动名单")
	}
	if tipCount > 0 {
		bullets = append(bullets, "适合用于创作者支持或赞助页案例展示")
	}
	if len(sampleTitles) > 0 {
		bullets = append(bullets, "当前内容主题偏向："+strings.Join(sampleTitles, "、"))
	}
	if len(bullets) == 0 {
		bullets = append(bullets, "当前公开信号还偏少，更适合先观察后再做推荐")
	}
	return uniqueTrimmed(bullets, 4)
}

func buildCreatorShortCopy(item *user.User, sampleTitles []string, followers, totalLikes, totalComments int64) string {
	name := firstNonEmpty(ptrString(item.FurryName), item.Username)
	topic := firstNonEmpty(strings.Join(sampleTitles, "、"), "持续创作")
	return fmt.Sprintf("%s近期围绕%s持续输出，当前有%d粉丝、样本内容累计%d赞和%d评论，适合作为创作者推荐位候选。", name, topic, followers, totalLikes, totalComments)
}

func buildCreatorLongCopy(item *user.User, sampleTitles []string, totalPosts, totalLikes, totalComments int64, tipCount int) string {
	name := firstNonEmpty(ptrString(item.FurryName), item.Username)
	topic := firstNonEmpty(strings.Join(sampleTitles, "、"), "稳定创作")
	return fmt.Sprintf("%s（@%s）近期围绕%s保持稳定输出，公开内容数为%d，样本内容累计获得%d个赞和%d条评论。%s这类账号更适合用于创作者推荐、内容专题或社区活动联动。",
		name,
		item.Username,
		topic,
		totalPosts,
		totalLikes,
		totalComments,
		func() string {
			if tipCount > 0 {
				return fmt.Sprintf("同时其已有 %d 笔打赏记录，说明有一定支持度。", tipCount)
			}
			return ""
		}(),
	)
}

func buildEventSellingPoints(item *event.Event, location string) []string {
	bullets := []string{
		"优先突出活动主题、时间和参与门槛，别只写标题",
	}
	if location == "线上活动" {
		bullets = append(bullets, "线上活动适合强调参与轻量、时间明确、报名成本低")
	} else {
		bullets = append(bullets, "线下活动适合强调现场氛围、地点信息和参与体验")
	}
	if len(item.Tags) > 0 {
		bullets = append(bullets, "可围绕这些标签提炼卖点："+strings.Join(item.Tags, "、"))
	}
	bullets = append(bullets, "简介里最值得提前放大的点是："+truncateText(item.Description, 52))
	return uniqueTrimmed(bullets, 4)
}

func buildEventPublishChecklist(item *event.Event, location string) []string {
	bullets := []string{
		"确认标题、时间、地点和报名状态在文案里都写清楚",
	}
	if location == "线上活动" {
		bullets = append(bullets, "线上活动记得补平台入口或报名后的进群/通知方式")
	} else {
		bullets = append(bullets, "线下活动记得补交通、场地或签到相关说明")
	}
	if item.MaxCapacity > 0 {
		bullets = append(bullets, fmt.Sprintf("当前人数 %d / %d，文案里可适度制造稀缺感，但不要虚报名额", item.AttendeeCount, item.MaxCapacity))
	}
	return uniqueTrimmed(bullets, 4)
}

func buildEventShortCopy(item *event.Event, location string) string {
	return fmt.Sprintf("%s｜%s · %s。%s，感兴趣可以尽快报名。",
		item.Title,
		item.StartTime.Format("01-02 15:04"),
		location,
		truncateText(item.Description, 42),
	)
}

func buildEventLongCopy(item *event.Event, location string) string {
	return fmt.Sprintf("%s将于%s在%s开启。%s如果你对%s感兴趣，这会是一次比较适合参与和交流的机会。",
		item.Title,
		item.StartTime.Format("2006-01-02 15:04"),
		location,
		truncateText(item.Description, 72),
		firstNonEmpty(strings.Join(item.Tags, "、"), "相关主题"),
	)
}

func buildToolPrompt(title string, sections []AdminAIToolSection) string {
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, section := range sections {
		if strings.TrimSpace(section.Title) == "" {
			continue
		}
		b.WriteString("## ")
		b.WriteString(section.Title)
		b.WriteString("\n")
		for _, bullet := range section.Bullets {
			if strings.TrimSpace(bullet) != "" {
				b.WriteString("- ")
				b.WriteString(strings.TrimSpace(bullet))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func optionalBullet(label string, value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return ""
	}
	return label + "：" + strings.TrimSpace(*value)
}

func optionalInline(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return label + "：" + strings.TrimSpace(value)
}

func optionalListBullet(label string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	return label + "：" + strings.Join(values, "、")
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func followerCount(value *follow.FollowStats) int64 {
	if value == nil {
		return 0
	}
	return value.FollowerCount
}

func followingCount(value *follow.FollowStats) int64 {
	if value == nil {
		return 0
	}
	return value.FollowingCount
}

func visibilityPublicPtr() *post.Visibility {
	vis := post.VisibilityPublic
	return &vis
}

func groupPublicPrivacyPtr() *group.GroupPrivacy {
	privacy := group.GroupPrivacyPublic
	return &privacy
}
