package usecase

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type assistantCopilotBundle struct {
	Insights     []AssistantInsight
	ContextText  string
	FallbackText string
}

// buildCopilotBundle computes safe read-only copilot outputs for selected pages.
func (s *AssistantService) buildCopilotBundle(ctx context.Context, query string, pageContext *AssistantPageContext) assistantCopilotBundle {
	if pageContext == nil {
		return assistantCopilotBundle{}
	}

	var insights []AssistantInsight
	switch pageContext.Kind {
	case "post_create":
		insights = s.buildPostCreateInsights(ctx, query, pageContext)
	case "group_detail":
		insights = s.buildGroupDetailInsights(query, pageContext)
	case "event_detail":
		insights = s.buildEventDetailInsights(query, pageContext)
	default:
		return assistantCopilotBundle{}
	}

	if len(insights) == 0 {
		return assistantCopilotBundle{}
	}

	return assistantCopilotBundle{
		Insights:     insights,
		ContextText:  buildCopilotPromptContext(insights),
		FallbackText: buildCopilotFallbackText(insights),
	}
}

func (s *AssistantService) buildPostCreateInsights(ctx context.Context, query string, pageContext *AssistantPageContext) []AssistantInsight {
	fields := pageContext.Fields
	draftTitle := strings.TrimSpace(fields["draft_title"])
	draftContent := strings.TrimSpace(fields["draft_content"])
	draftTags := splitCSVLike(fields["draft_tags"])
	groupName := strings.TrimSpace(fields["group_name"])
	currentVisibility := strings.TrimSpace(fields["visibility"])
	aiGenerated := strings.TrimSpace(fields["ai_generated"]) != ""

	if draftTitle == "" && draftContent == "" {
		return []AssistantInsight{
			{
				Kind:    "draft_polish",
				Title:   "发帖 Copilot 预备建议",
				Summary: "先写出 1 到 2 句你最想表达的核心内容，我再帮你补标题、标签和可见性建议。",
				Bullets: []string{
					"先交代你想分享的对象、作品或场景",
					"再补充 1 到 2 个具体细节，方便我做更准的润色和标签推荐",
				},
			},
		}
	}

	titleOptions := generateTitleOptions(draftTitle, draftContent, groupName)
	tagSuggestions := s.suggestTagsForDraft(ctx, draftTitle, draftContent, draftTags, groupName)
	visibilitySummary, visibilityBullets := suggestVisibility(draftContent, currentVisibility, groupName, aiGenerated)

	insights := []AssistantInsight{
		{
			Kind:    "draft_polish",
			Title:   "正文润色建议",
			Summary: postPolishSummary(draftContent, groupName),
			Bullets: postPolishBullets(draftContent, groupName, aiGenerated),
		},
		{
			Kind:    "title_options",
			Title:   "标题备选",
			Summary: "下面这些标题都尽量保留你当前草稿的主题，不会把语气改得太生硬。",
			Bullets: titleOptions,
		},
		{
			Kind:    "tag_suggestions",
			Title:   "推荐标签",
			Summary: "优先给出和草稿内容、圈子场景更贴近的标签。",
			Bullets: tagSuggestions,
		},
		{
			Kind:    "visibility_suggestion",
			Title:   "可见性建议",
			Summary: visibilitySummary,
			Bullets: visibilityBullets,
		},
	}

	if strings.Contains(query, "标题") {
		slices.SortStableFunc(insights, func(a, b AssistantInsight) int {
			return copilotPriority(a.Kind, "title_options") - copilotPriority(b.Kind, "title_options")
		})
	}
	if strings.Contains(query, "标签") {
		slices.SortStableFunc(insights, func(a, b AssistantInsight) int {
			return copilotPriority(a.Kind, "tag_suggestions") - copilotPriority(b.Kind, "tag_suggestions")
		})
	}
	return insights
}

func (s *AssistantService) buildGroupDetailInsights(query string, pageContext *AssistantPageContext) []AssistantInsight {
	fields := pageContext.Fields
	groupName := strings.TrimSpace(strings.TrimPrefix(pageContext.Title, "圈子详情："))
	groupTags := splitListLike(fields["group_tags"])
	memberCount := strings.TrimSpace(fields["member_count"])
	postCount := strings.TrimSpace(fields["post_count"])
	description := strings.TrimSpace(fields["group_description"])
	rules := strings.TrimSpace(fields["group_rules"])
	announcement := strings.TrimSpace(fields["group_announcement"])
	featuredPost := strings.TrimSpace(fields["featured_post"])
	isMember := strings.TrimSpace(fields["is_member"]) == "已加入"

	atmosphereBullets := []string{
		fmt.Sprintf("成员规模：%s，帖子量：%s", firstNonEmpty(memberCount, "未知"), firstNonEmpty(postCount, "未知")),
	}
	if len(groupTags) > 0 {
		atmosphereBullets = append(atmosphereBullets, "圈子标签："+strings.Join(groupTags, "、"))
	}
	if featuredPost != "" {
		atmosphereBullets = append(atmosphereBullets, "精选内容偏向："+featuredPost)
	}
	if description != "" {
		atmosphereBullets = append(atmosphereBullets, "简介重点："+truncateText(description, 72))
	}

	ruleBullets := extractListBullets(rules, 4)
	if len(ruleBullets) == 0 {
		ruleBullets = extractListBullets(announcement, 3)
	}
	if len(ruleBullets) == 0 {
		ruleBullets = []string{
			"先看公告和精选内容，确认圈子当前的交流重点",
			"发帖前先对照简介和标签，避免内容跑偏",
		}
	}

	joinSummary, joinBullets := assessGroupJoinFit(groupName, isMember, memberCount, postCount, groupTags, rules, announcement)
	postIdeas := suggestGroupPostIdeas(groupName, groupTags, description, featuredPost, isMember)

	insights := []AssistantInsight{
		{
			Kind:    "group_atmosphere",
			Title:   "圈子氛围速览",
			Summary: summarizeGroupAtmosphere(groupName, description, groupTags, isMember),
			Bullets: atmosphereBullets,
		},
		{
			Kind:    "rules_summary",
			Title:   "规则与公告摘要",
			Summary: "先看这些重点，再决定要不要加入或发什么内容。",
			Bullets: ruleBullets,
		},
		{
			Kind:    "join_suggestion",
			Title:   "是否适合加入",
			Summary: joinSummary,
			Bullets: joinBullets,
		},
		{
			Kind:    "posting_ideas",
			Title:   "适合发什么内容",
			Summary: "下面这些方向更容易贴合这个圈子的当前语境。",
			Bullets: postIdeas,
		},
	}

	if strings.Contains(query, "规则") || strings.Contains(query, "群规") {
		slices.SortStableFunc(insights, func(a, b AssistantInsight) int {
			return copilotPriority(a.Kind, "rules_summary") - copilotPriority(b.Kind, "rules_summary")
		})
	}
	if strings.Contains(query, "加入") || strings.Contains(query, "适合") {
		slices.SortStableFunc(insights, func(a, b AssistantInsight) int {
			return copilotPriority(a.Kind, "join_suggestion") - copilotPriority(b.Kind, "join_suggestion")
		})
	}
	return insights
}

func (s *AssistantService) buildEventDetailInsights(query string, pageContext *AssistantPageContext) []AssistantInsight {
	fields := pageContext.Fields
	title := strings.TrimSpace(strings.TrimPrefix(pageContext.Title, "活动详情："))
	status := strings.TrimSpace(fields["event_status"])
	timeText := strings.TrimSpace(fields["event_time"])
	location := strings.TrimSpace(fields["event_location"])
	attendees := strings.TrimSpace(fields["event_attendees"])
	description := strings.TrimSpace(fields["event_description"])
	eventTags := splitListLike(fields["event_tags"])
	hasAttended := strings.TrimSpace(fields["has_attended"]) == "已报名"

	summaryBullets := []string{
		"状态：" + firstNonEmpty(status, "未知"),
		"时间：" + firstNonEmpty(timeText, "未提供"),
		"地点：" + firstNonEmpty(location, "未提供"),
		"人数：" + firstNonEmpty(attendees, "未知"),
	}
	if len(eventTags) > 0 {
		summaryBullets = append(summaryBullets, "标签："+strings.Join(eventTags, "、"))
	}
	if description != "" {
		summaryBullets = append(summaryBullets, "简介重点："+truncateText(description, 72))
	}

	fitSummary, fitBullets := assessEventFit(status, location, attendees, eventTags, hasAttended)
	signupSummary, signupBullets := summarizeEventSignup(status, attendees, hasAttended, location)
	checklistSummary, checklistBullets := buildEventChecklist(status, location, hasAttended)

	insights := []AssistantInsight{
		{
			Kind:    "event_summary",
			Title:   "活动关键信息",
			Summary: summarizeEventOverview(title, status, location, hasAttended),
			Bullets: summaryBullets,
		},
		{
			Kind:    "fit_assessment",
			Title:   "是否适合参加",
			Summary: fitSummary,
			Bullets: fitBullets,
		},
		{
			Kind:    "signup_notes",
			Title:   "报名与参与提示",
			Summary: signupSummary,
			Bullets: signupBullets,
		},
		{
			Kind:    "preparation_checklist",
			Title:   "准备清单",
			Summary: checklistSummary,
			Bullets: checklistBullets,
		},
	}

	if strings.Contains(query, "准备") || strings.Contains(query, "清单") || strings.Contains(query, "注意") {
		slices.SortStableFunc(insights, func(a, b AssistantInsight) int {
			return copilotPriority(a.Kind, "preparation_checklist") - copilotPriority(b.Kind, "preparation_checklist")
		})
	}
	return insights
}

func buildCopilotPromptContext(insights []AssistantInsight) string {
	if len(insights) == 0 {
		return ""
	}
	lines := make([]string, 0, len(insights)*4)
	lines = append(lines, "当前页面 Copilot 工具结果：")
	for _, insight := range insights {
		if strings.TrimSpace(insight.Title) == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- [%s] %s", insight.Kind, insight.Title))
		if summary := strings.TrimSpace(insight.Summary); summary != "" {
			lines = append(lines, "  结论："+summary)
		}
		for _, bullet := range insight.Bullets {
			if trimmed := strings.TrimSpace(bullet); trimmed != "" {
				lines = append(lines, "  - "+trimmed)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func buildCopilotFallbackText(insights []AssistantInsight) string {
	if len(insights) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("我先把这个页面的 Copilot 结论整理给你：\n")
	for _, insight := range insights {
		if strings.TrimSpace(insight.Title) == "" {
			continue
		}
		fmt.Fprintf(&b, "\n**%s**\n", insight.Title)
		if strings.TrimSpace(insight.Summary) != "" {
			b.WriteString(insight.Summary)
			b.WriteString("\n")
		}
		for _, bullet := range insight.Bullets {
			if trimmed := strings.TrimSpace(bullet); trimmed != "" {
				b.WriteString("- ")
				b.WriteString(trimmed)
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func generateTitleOptions(existingTitle, draftContent, groupName string) []string {
	baseTopic := firstNonEmpty(existingTitle, extractLeadSentence(draftContent))
	baseTopic = strings.Trim(baseTopic, "：:。！？!?,， ")
	if baseTopic == "" {
		baseTopic = "这次想分享的内容"
	}
	baseTopic = truncateText(baseTopic, 20)

	options := []string{
		baseTopic,
		baseTopic + "的一点记录",
	}
	if groupName != "" {
		options = append(options, fmt.Sprintf("%s｜%s", groupName, baseTopic))
	} else {
		options = append(options, baseTopic+"｜最近想说的话")
	}
	return uniqueTrimmed(options, 3)
}

func (s *AssistantService) suggestTagsForDraft(ctx context.Context, title, content string, currentTags []string, groupName string) []string {
	candidates := make([]string, 0, 6)
	candidates = append(candidates, currentTags...)

	if s.postService != nil {
		if hotTags, err := s.postService.GetHotTags(ctx, 12); err == nil {
			haystack := strings.ToLower(title + " " + content + " " + groupName)
			for _, tag := range hotTags {
				lowerTag := strings.ToLower(tag)
				if haystack != "" && (strings.Contains(haystack, lowerTag) || strings.Contains(lowerTag, haystack)) {
					candidates = append(candidates, tag)
				}
			}
		}
	}

	if phrase := extractLeadSentence(content); phrase != "" {
		candidates = append(candidates, truncateText(phrase, 10))
	}
	if groupName != "" {
		candidates = append(candidates, groupName)
	}

	return uniqueTrimmed(prefixTagCandidates(candidates), 4)
}

func postPolishSummary(draftContent, groupName string) string {
	if groupName != "" {
		return "这条动态更适合先点明你和圈子主题的关系，再补 1 到 2 个具体细节，最后收束成一句可互动的结尾。"
	}
	if utf8Len(draftContent) < 60 {
		return "当前草稿偏短，适合补一个更明确的开场句和一个能引发互动的收尾句。"
	}
	return "当前草稿信息已经够用，重点是把段落拆清楚，并把最想表达的重点提前。"
}

func postPolishBullets(draftContent, groupName string, aiGenerated bool) []string {
	bullets := []string{
		"开头先说明你想分享的对象、场景或主题，别让读者猜",
		"中间保留 1 到 2 个最具体的细节，避免一句话塞太多信息",
		"结尾可以补一句感受、问题或邀请互动的话",
	}
	if groupName != "" {
		bullets = append(bullets, "如果是发到圈子里，可以点一下它和“"+groupName+"”的关联")
	}
	if aiGenerated {
		bullets = append(bullets, "既然已经勾选 AI 生成标记，正文里最好再补一两句你的个人判断或补充说明")
	}
	if lead := extractLeadSentence(draftContent); lead != "" {
		bullets = append(bullets, "当前草稿最适合放到开头的核心句是："+truncateText(lead, 28))
	}
	return uniqueTrimmed(bullets, 4)
}

func suggestVisibility(draftContent, currentVisibility, groupName string, aiGenerated bool) (string, []string) {
	normalized := strings.ToLower(draftContent)
	privateSignals := []string{"约稿", "私稿", "测试", "仅自己", "联系方式", "微信", "qq", "还没定稿", "未公开"}
	for _, signal := range privateSignals {
		if strings.Contains(normalized, signal) {
			return "这条内容更像阶段性草稿或带私密信息，建议至少设成“仅关注者可见”，更敏感时可以直接设为“私密”。", []string{
				"如果只是想先收集熟人反馈，用“仅关注者可见”比较稳妥",
				"如果还没准备公开，直接设成“私密”更合适",
			}
		}
	}

	summary := "当前内容没有明显的隐私风险，默认“公开”是可接受的。"
	bullets := []string{
		"如果你希望陌生人也能看到并互动，保持“公开”即可",
		"如果只是想先给熟人或老关注者看，可以改成“仅关注者可见”",
	}
	if groupName != "" {
		bullets = append(bullets, "既然已经带有明确圈子场景，公开也不会太偏题，但最好正文里点出和圈子的关系")
	}
	if currentVisibility != "" {
		bullets = append(bullets, "你当前选择的是："+currentVisibility)
	}
	if aiGenerated {
		bullets = append(bullets, "如果内容含 AI 参与，公开时建议保留 AI 标记，避免误解")
	}
	return summary, uniqueTrimmed(bullets, 4)
}

func summarizeGroupAtmosphere(groupName, description string, tags []string, isMember bool) string {
	base := "这个圈子更像一个围绕共同兴趣持续交流的空间。"
	if len(tags) > 0 {
		base = "这个圈子的主题比较明确，主要围绕“" + strings.Join(tags[:min(2, len(tags))], "、") + "”展开。"
	}
	if description != "" {
		base += " 简介里强调的是：" + truncateText(description, 44)
	}
	if isMember {
		base += " 你已经加入，接下来更适合考虑怎么参与和发什么。"
	}
	return base
}

func assessGroupJoinFit(groupName string, isMember bool, memberCount, postCount string, tags []string, rules, announcement string) (string, []string) {
	if isMember {
		return "你已经加入这个圈子，当前更值得做的是先看规则和精选，再决定要不要发帖或互动。", []string{
			"先看公告和精选内容，判断最近大家在聊什么",
			"如果想发帖，尽量把内容往圈子标签和规则上靠",
		}
	}

	bullets := []string{
		"先确认圈子规则是否和你的内容方向相符",
		"再看最近精选或热门内容，判断氛围偏创作、交流还是资讯",
	}
	if memberCount != "" && postCount != "" {
		bullets = append(bullets, "成员数和帖子量能帮助你判断这个圈子是否活跃："+memberCount+" 成员 / "+postCount+" 帖子")
	}
	if len(tags) > 0 {
		bullets = append(bullets, "如果你对这些标签有持续兴趣，匹配度会更高："+strings.Join(tags, "、"))
	}

	if rules == "" && announcement == "" {
		return fmt.Sprintf("从公开信息看，%s 适合先观察一下内容氛围；如果标签方向和你一致，再加入会更稳。", firstNonEmpty(groupName, "这个圈子")), uniqueTrimmed(bullets, 4)
	}
	return fmt.Sprintf("%s的公开信息已经足够判断方向。如果它的规则、公告和标签都和你一致，就值得加入。", firstNonEmpty(groupName, "这个圈子")), uniqueTrimmed(bullets, 4)
}

func suggestGroupPostIdeas(groupName string, tags []string, description, featuredPost string, isMember bool) []string {
	ideas := make([]string, 0, 4)
	if len(tags) > 0 {
		ideas = append(ideas, "围绕“"+strings.Join(tags[:min(2, len(tags))], "、")+"”发一条带细节和个人观点的内容")
	}
	if featuredPost != "" {
		ideas = append(ideas, "参考当前精选内容的切入方式，但换成你自己的经历或作品")
	}
	if description != "" {
		ideas = append(ideas, "根据圈子简介里的主题，发一条更贴近圈子语境的分享")
	}
	if isMember {
		ideas = append(ideas, "先发一条轻量内容试水，比如进度、灵感或最近的相关观察")
	} else {
		ideas = append(ideas, "加入前先收藏或观察几条内容，确认你的题材不会明显跑偏")
	}
	return uniqueTrimmed(ideas, 4)
}

func summarizeEventOverview(title, status, location string, hasAttended bool) string {
	base := fmt.Sprintf("这是“%s”的详情页，当前状态是%s。", firstNonEmpty(title, "该活动"), firstNonEmpty(status, "未知"))
	if location != "" {
		base += " 形式上偏向" + location + "。"
	}
	if hasAttended {
		base += " 你已经报名，重点转成行前准备。"
	}
	return base
}

func assessEventFit(status, location, attendees string, tags []string, hasAttended bool) (string, []string) {
	if hasAttended {
		return "你已经报名了，这一步不用再纠结适不适合，重点是确认时间地点和准备事项。", []string{
			"优先核对活动时间和地点",
			"再对照活动标签和简介做行前准备",
		}
	}

	bullets := []string{}
	if location == "线上活动" {
		bullets = append(bullets, "线上活动对新手通常更友好，参与门槛更低")
	} else if location != "" {
		bullets = append(bullets, "线下活动更看重时间、路程和现场社交舒适度")
	}
	if attendees != "" {
		bullets = append(bullets, "当前人数信息："+attendees)
	}
	if len(tags) > 0 {
		bullets = append(bullets, "活动主题偏向："+strings.Join(tags, "、"))
	}

	if status != "报名中" {
		return "这个活动当前不在标准报名状态，先确认是否还能参加，再考虑准备事项。", uniqueTrimmed(bullets, 4)
	}
	return "如果活动主题和你的兴趣对得上，而且时间地点也能接受，这个活动就值得考虑参加。", uniqueTrimmed(bullets, 4)
}

func summarizeEventSignup(status, attendees string, hasAttended bool, location string) (string, []string) {
	if hasAttended {
		return "你已经在参与名单里了，重点是别错过时间和现场信息。", []string{
			"出发前再核对一次时间、地点和交通安排",
			"如果是线上活动，提前确认设备、网络和平台入口",
		}
	}

	bullets := []string{
		"先看活动是否还在“报名中”",
		"再确认时间和地点是否能配合你的安排",
	}
	if attendees != "" {
		bullets = append(bullets, "参与人数信息："+attendees)
	}
	if location == "线上活动" {
		bullets = append(bullets, "线上报名后通常要特别留意活动平台或群入口")
	}
	return "报名前先确认状态、时间和地点，别只看标题就直接冲。", uniqueTrimmed(bullets, 4)
}

func buildEventChecklist(status, location string, hasAttended bool) (string, []string) {
	bullets := []string{}
	if location == "线上活动" {
		bullets = append(bullets,
			"提前确认网络、耳机麦克风和活动平台入口",
			"把开始时间加到提醒里，避免错过开场",
		)
	} else {
		bullets = append(bullets,
			"提前确认路线、到场时间和现场联络方式",
			"准备好水、充电设备和必要证件或票据",
		)
	}
	if status == "报名中" && !hasAttended {
		bullets = append(bullets, "如果还没报名，先确认名额和报名方式")
	}
	bullets = append(bullets, "再读一遍活动简介，确认有没有 dress code、携带要求或注意事项")
	return "这份清单只覆盖安全准备项，适合先快速检查。", uniqueTrimmed(bullets, 4)
}

func extractListBullets(text string, limit int) []string {
	if limit <= 0 {
		limit = 4
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := regexp.MustCompile(`[；;。！？\n]+`).Split(text, -1)
	items := make([]string, 0, limit)
	for _, part := range parts {
		part = strings.TrimSpace(strings.Trim(part, "-•0123456789. "))
		if part == "" {
			continue
		}
		items = append(items, truncateText(part, 36))
		if len(items) >= limit {
			break
		}
	}
	return uniqueTrimmed(items, limit)
}

func extractLeadSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	parts := regexp.MustCompile(`[。！？\n]+`).Split(text, -1)
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			return trimmed
		}
	}
	return text
}

func splitCSVLike(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	items := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '，' || r == '、' || r == ';' || r == '；'
	})
	return uniqueTrimmed(items, 8)
}

func splitListLike(text string) []string {
	return splitCSVLike(text)
}

func prefixTagCandidates(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(strings.TrimPrefix(item, "#"))
		if trimmed == "" {
			continue
		}
		out = append(out, "#"+trimmed)
	}
	return out
}

func uniqueTrimmed(items []string, limit int) []string {
	if limit <= 0 {
		limit = len(items)
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, min(limit, len(items)))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func utf8Len(text string) int {
	return len([]rune(text))
}

func copilotPriority(kind, preferred string) int {
	if kind == preferred {
		return 0
	}
	return 1
}
