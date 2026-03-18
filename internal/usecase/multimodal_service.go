package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
	post "github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/observability/assistantmetrics"
	"github.com/studio/platform/internal/pkg/resilience"
)

type visionAnalyzer interface {
	Configured() bool
	AnalyzeImages(ctx context.Context, prompt string, imageURLs []string) (string, error)
}

type MediaAnalysisResult struct {
	Items        []*assistantdomain.MediaAnalysis `json:"items"`
	Fallback     bool                             `json:"fallback"`
	Provider     string                           `json:"provider"`
	CircuitState string                           `json:"circuit_state"`
}

type MultimodalService struct {
	cfg          configs.AssistantConfig
	repo         assistantdomain.Repository
	vision       visionAnalyzer
	postService  *PostService
	allowedHosts []string
	frontendURL  string
	analysisTTL  time.Duration
	circuit      *resilience.CircuitBreaker
}

func NewMultimodalService(
	cfg configs.AssistantConfig,
	repo assistantdomain.Repository,
	vision visionAnalyzer,
	postService *PostService,
	allowedHosts []string,
	frontendURL string,
) *MultimodalService {
	if cfg.VisionTimeoutSec <= 0 {
		cfg.VisionTimeoutSec = 45
	}
	if cfg.VisionRetryMax <= 0 {
		cfg.VisionRetryMax = 2
	}
	if cfg.CircuitFailures <= 0 {
		cfg.CircuitFailures = 3
	}
	if cfg.CircuitOpenSec <= 0 {
		cfg.CircuitOpenSec = 60
	}
	if cfg.MediaCacheTTLSec <= 0 {
		cfg.MediaCacheTTLSec = 86400
	}
	return &MultimodalService{
		cfg:          cfg,
		repo:         repo,
		vision:       vision,
		postService:  postService,
		allowedHosts: allowedHosts,
		frontendURL:  frontendURL,
		analysisTTL:  time.Duration(cfg.MediaCacheTTLSec) * time.Second,
		circuit:      resilience.NewCircuitBreaker(cfg.CircuitFailures, time.Duration(cfg.CircuitOpenSec)*time.Second),
	}
}

func (s *MultimodalService) AnalyzeMedia(ctx context.Context, mediaURLs []string, purpose string) (*MediaAnalysisResult, error) {
	start := time.Now()
	cacheHit := true
	fallbackUsed := false
	retries := 0
	var lastErr error
	defer func() {
		state := ""
		if s.circuit != nil {
			state = string(s.circuit.Snapshot().State)
			assistantmetrics.RecordCircuitState("vision", state)
		}
		assistantmetrics.RecordMultimodal(firstNonEmpty(purpose, "generic"), time.Since(start), cacheHit, fallbackUsed, retries, lastErr)
	}()

	items := make([]*assistantdomain.MediaAnalysis, 0, len(mediaURLs))
	for _, rawURL := range mediaURLs {
		normalizedURL, err := s.validateMediaURL(rawURL)
		if err != nil {
			return nil, err
		}
		item, itemCacheHit, itemFallback, itemRetries, err := s.analyzeSingle(ctx, normalizedURL, purpose)
		if err != nil {
			lastErr = err
			return nil, err
		}
		if !itemCacheHit {
			cacheHit = false
		}
		if itemFallback {
			fallbackUsed = true
		}
		retries += itemRetries
		items = append(items, item)
	}

	return &MediaAnalysisResult{
		Items:        items,
		Fallback:     fallbackUsed,
		Provider:     s.providerLabel(fallbackUsed),
		CircuitState: string(s.circuit.Snapshot().State),
	}, nil
}

func (s *MultimodalService) ExplainPostModeration(ctx context.Context, postID uuid.UUID) (*AdminAIToolResult, error) {
	if s.postService == nil {
		return nil, fmt.Errorf("post service is not configured")
	}
	item, err := s.postService.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	result, err := s.AnalyzeMedia(ctx, item.MediaURLs, "moderation")
	if err != nil {
		return nil, err
	}
	bullets := []string{
		"帖子作者：@" + firstNonEmpty(item.AuthorUsername, item.AuthorID.String()),
		"文本摘要：" + truncateText(item.Content, 72),
		"图片数量：" + fmt.Sprintf("%d", len(item.MediaURLs)),
	}
	for _, media := range result.Items {
		if media == nil {
			continue
		}
		bullets = append(bullets, "图片说明："+firstNonEmpty(media.ModerationSummary, media.ImageSummary, media.AltText))
		if media.RiskLevel != "" {
			bullets = append(bullets, "风险等级："+media.RiskLevel)
		}
	}

	draft := buildModerationExplanationDraft(item, result.Items)
	return &AdminAIToolResult{
		RunID:   uuid.NewString(),
		Tool:    "moderation_explanation",
		Title:   "审核解释摘要",
		Summary: "基于帖子文本和已上传图片生成解释性摘要，帮助审核员快速判断风险点。",
		Sections: []AdminAIToolSection{
			{Title: "审核要点", Bullets: uniqueTrimmed(bullets, 8)},
		},
		Drafts: []AdminAIToolDraft{
			{Label: "审核说明草稿", Content: draft},
		},
		Fallback:    result.Fallback,
		Provider:    result.Provider,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MultimodalService) analyzeSingle(ctx context.Context, mediaURL, purpose string) (*assistantdomain.MediaAnalysis, bool, bool, int, error) {
	if s.repo != nil {
		if cached, err := s.repo.GetMediaAnalysis(ctx, mediaURL); err == nil && cached != nil {
			return cached, true, cached.Fallback, 0, nil
		}
	}

	fallback := buildFallbackMediaAnalysis(mediaURL, purpose, s.analysisTTL)
	if s.vision == nil || !s.vision.Configured() || s.circuit == nil || !s.circuit.Allow() {
		if s.repo != nil {
			_ = s.repo.UpsertMediaAnalysis(ctx, fallback)
		}
		return fallback, false, true, 0, nil
	}

	prompt := buildVisionPrompt(purpose)
	maxRetries := max(s.cfg.VisionRetryMax, 1)
	for attempt := 0; attempt < maxRetries; attempt++ {
		visionCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.VisionTimeoutSec)*time.Second)
		raw, err := s.vision.AnalyzeImages(visionCtx, prompt, []string{mediaURL})
		cancel()
		if err == nil {
			parsed, parseErr := parseMediaAnalysisResponse(raw, mediaURL, purpose, s.analysisTTL)
			if parseErr == nil {
				parsed.Provider = s.providerLabel(false)
				parsed.Model = firstNonEmpty(s.cfg.VisionModel, s.cfg.Model)
				s.circuit.RecordSuccess()
				if s.repo != nil {
					_ = s.repo.UpsertMediaAnalysis(ctx, parsed)
				}
				return parsed, false, parsed.Fallback, attempt, nil
			}
			err = parseErr
		}
		s.circuit.RecordFailure(err)
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
		}
	}

	fallback.Fallback = true
	if s.repo != nil {
		_ = s.repo.UpsertMediaAnalysis(ctx, fallback)
	}
	return fallback, false, true, maxRetries - 1, nil
}

func (s *MultimodalService) validateMediaURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("媒体地址不能为空")
	}
	if strings.HasPrefix(raw, "/uploads/") {
		if strings.TrimSpace(s.frontendURL) == "" {
			return raw, nil
		}
		return strings.TrimRight(s.frontendURL, "/") + raw, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("媒体地址无效: %w", err)
	}
	host := strings.ToLower(u.Host)
	for _, allowed := range s.allowedHosts {
		if strings.EqualFold(strings.TrimSpace(allowed), host) {
			return raw, nil
		}
	}
	return "", fmt.Errorf("媒体地址不在允许的域名白名单内")
}

func (s *MultimodalService) providerLabel(fallback bool) string {
	if fallback || strings.TrimSpace(s.cfg.VisionAPIKey) == "" {
		return "multimodal-fallback"
	}
	return firstNonEmpty(s.cfg.VisionModel, "vision")
}

func buildVisionPrompt(purpose string) string {
	mode := "为发帖辅助生成图片理解结果"
	if purpose == "moderation" {
		mode = "为内容审核生成解释性摘要"
	}
	return strings.TrimSpace(`
你是一个社区图片理解助手，需要根据图片生成严格的 JSON。

任务：` + mode + `

输出要求：
1. 只能返回 JSON，不要加 markdown 代码块。
2. JSON 结构固定为：
{
  "alt_text": "...",
  "tags": ["..."],
  "image_summary": "...",
  "moderation_summary": "...",
  "risk_level": "low|medium|high",
  "safety_notes": ["..."]
}
3. alt_text 用中文，尽量客观描述画面，不要编造身份或剧情。
4. tags 给 3 到 5 个中文短标签。
5. moderation_summary 重点解释画面里哪些元素可能影响审核判断。
6. 如果画面无明显风险，risk_level 填 low，safety_notes 可以为空数组。
`)
}

func parseMediaAnalysisResponse(raw, mediaURL, purpose string, ttl time.Duration) (*assistantdomain.MediaAnalysis, error) {
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(raw, "```"), "```json"))
	var payload struct {
		AltText           string   `json:"alt_text"`
		Tags              []string `json:"tags"`
		ImageSummary      string   `json:"image_summary"`
		ModerationSummary string   `json:"moderation_summary"`
		RiskLevel         string   `json:"risk_level"`
		SafetyNotes       []string `json:"safety_notes"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	now := time.Now()
	if payload.AltText == "" {
		payload.AltText = truncateText(payload.ImageSummary, 60)
	}
	return &assistantdomain.MediaAnalysis{
		ID:                uuid.New(),
		MediaURL:          mediaURL,
		AltText:           truncateText(payload.AltText, 120),
		Tags:              uniqueTrimmed(payload.Tags, 5),
		ImageSummary:      truncateText(payload.ImageSummary, 180),
		ModerationSummary: truncateText(payload.ModerationSummary, 180),
		RiskLevel:         normalizeRiskLevel(payload.RiskLevel),
		SafetyNotes:       uniqueTrimmed(payload.SafetyNotes, 4),
		Fallback:          false,
		CachedAt:          now,
		ExpiresAt:         now.Add(ttl),
	}, nil
}

func buildFallbackMediaAnalysis(mediaURL, purpose string, ttl time.Duration) *assistantdomain.MediaAnalysis {
	now := time.Now()
	filename := strings.Trim(strings.TrimSpace(path.Base(mediaURL)), "/")
	name := strings.TrimSuffix(filename, path.Ext(filename))
	if name == "" {
		name = "已上传图片"
	}
	tags := []string{"图片", "待人工确认"}
	if purpose == "moderation" {
		tags = []string{"图片", "审核辅助", "待人工确认"}
	}
	return &assistantdomain.MediaAnalysis{
		ID:                uuid.New(),
		MediaURL:          mediaURL,
		AltText:           "图片内容待视觉模型补充，当前仅记录为已上传图片：" + truncateText(name, 24),
		Tags:              tags,
		ImageSummary:      "当前走的是本地降级路径，没有拿到稳定的视觉模型结果。",
		ModerationSummary: "无法自动给出高置信度审核解释，建议人工查看原图后判断。",
		RiskLevel:         "medium",
		SafetyNotes: []string{
			"当前结果来自降级路径，不应单独作为最终审核依据",
			"建议结合帖子正文和原图一起人工判断",
		},
		Provider:  "multimodal-fallback",
		Model:     "fallback",
		Fallback:  true,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}
}

func normalizeRiskLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "medium"
	}
}

func buildModerationExplanationDraft(post *post.Post, analyses []*assistantdomain.MediaAnalysis) string {
	var parts []string
	parts = append(parts, "帖子文本摘要："+truncateText(post.Content, 72))
	for idx, item := range analyses {
		if item == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("图片%d：%s。风险等级：%s。", idx+1, firstNonEmpty(item.ModerationSummary, item.ImageSummary, item.AltText), firstNonEmpty(item.RiskLevel, "medium")))
	}
	return "这条内容的审核解释如下：\n" + strings.Join(parts, "\n")
}
