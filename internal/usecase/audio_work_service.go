package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audiojob"
	"github.com/studio/platform/internal/domain/audiowork"
	"github.com/studio/platform/internal/domain/bookmark"
	"github.com/studio/platform/internal/domain/comment"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/pkg/apperr"
)

type AudioWorkService struct {
	workRepo     audiowork.Repository
	audioJobRepo audiojob.Repository
	allowedHosts []string
	bookmarkRepo bookmark.Repository
	commentRepo  comment.Repository
}

type AudioWorkServiceOption func(*AudioWorkService)

func NewAudioWorkService(workRepo audiowork.Repository, audioJobRepo audiojob.Repository, opts ...AudioWorkServiceOption) *AudioWorkService {
	svc := &AudioWorkService{
		workRepo:     workRepo,
		audioJobRepo: audioJobRepo,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithAudioWorkAllowedHosts(hosts []string) AudioWorkServiceOption {
	return func(s *AudioWorkService) {
		s.allowedHosts = append([]string(nil), hosts...)
	}
}

func WithAudioWorkCleanup(bookmarkRepo bookmark.Repository, commentRepo comment.Repository) AudioWorkServiceOption {
	return func(s *AudioWorkService) {
		s.bookmarkRepo = bookmarkRepo
		s.commentRepo = commentRepo
	}
}

type PublishAudioWorkInput struct {
	UserID        uuid.UUID
	JobID         uuid.UUID
	Title         string
	Description   string
	CoverImageURL string
	Visibility    audiowork.Visibility
	Tags          []string
}

type ListAudioWorksInput struct {
	Search   string
	Tag      string
	Sort     string
	Page     int
	PageSize int
}

type UpdateAudioWorkInput struct {
	UserID        uuid.UUID
	WorkID        uuid.UUID
	Title         string
	Description   string
	CoverImageURL string
	Visibility    audiowork.Visibility
	Tags          []string
}

func (s *AudioWorkService) PublishFromJob(ctx context.Context, input PublishAudioWorkInput) (*audiowork.Work, error) {
	job, err := s.audioJobRepo.GetByID(ctx, input.JobID)
	if err != nil {
		if errors.Is(err, audiojob.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音频任务失败", err)
	}
	if job.UserID != input.UserID {
		return nil, apperr.ErrForbidden
	}
	if job.Status != audiojob.StatusSucceeded {
		return nil, apperr.BadRequest("只有成功任务可以发布为作品")
	}

	audioURL := readStringFromMap(job.Result, "output_audio_url")
	if audioURL == "" {
		return nil, apperr.BadRequest("当前任务暂无可发布的音频输出")
	}
	if err := s.validateAudioURL(audioURL); err != nil {
		return nil, err
	}
	if err := s.validateCoverURL(input.CoverImageURL); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = job.Title
	}
	if title == "" {
		title = "未命名音频作品"
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = readStringFromMap(job.Result, "summary")
	}

	now := time.Now()
	visibility := input.Visibility
	if visibility == "" {
		visibility = audiowork.VisibilityPublic
	}
	if visibility != audiowork.VisibilityPublic && visibility != audiowork.VisibilityPrivate {
		return nil, apperr.BadRequest("无效的作品可见性")
	}

	work := &audiowork.Work{
		ID:               uuid.New(),
		AuthorID:         input.UserID,
		SourceJobID:      input.JobID,
		Title:            title,
		AudioURL:         audioURL,
		DurationSec:      readDuration(job.Result),
		Visibility:       visibility,
		ModerationStatus: post.ModerationPending,
		Tags:             mergeWorkTags(input.Tags, job.Result),
		WaveformPreview:  readWaveform(job.Result),
		Metadata:         cloneMap(job.Result),
		PublishedAt:      now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if description != "" {
		work.Description = &description
	}
	if cover := strings.TrimSpace(input.CoverImageURL); cover != "" {
		work.CoverImageURL = &cover
	}

	if err := s.workRepo.Create(ctx, work); err != nil {
		if errors.Is(err, audiowork.ErrAlreadyPublished) {
			return nil, apperr.BadRequest("该任务已经发布过作品")
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "发布音频作品失败", err)
	}
	return work, nil
}

func (s *AudioWorkService) ListPublicWorks(ctx context.Context, input ListAudioWorksInput) ([]*audiowork.Work, int64, error) {
	visibility := audiowork.VisibilityPublic
	status := post.ModerationApproved
	items, total, err := s.workRepo.List(ctx, audiowork.ListFilter{
		Visibility:       &visibility,
		ModerationStatus: &status,
		Search:           input.Search,
		Tag:              input.Tag,
		Sort:             input.Sort,
		Page:             input.Page,
		PageSize:         input.PageSize,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询音频作品失败", err)
	}
	return items, total, nil
}

func (s *AudioWorkService) ListMyWorks(ctx context.Context, userID uuid.UUID, input ListAudioWorksInput) ([]*audiowork.Work, int64, error) {
	items, total, err := s.workRepo.List(ctx, audiowork.ListFilter{
		AuthorID: userIDPtr(userID),
		Search:   input.Search,
		Tag:      input.Tag,
		Sort:     input.Sort,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询我的音频作品失败", err)
	}
	return items, total, nil
}

func (s *AudioWorkService) ListUserPublicWorks(ctx context.Context, userID uuid.UUID, input ListAudioWorksInput) ([]*audiowork.Work, int64, error) {
	visibility := audiowork.VisibilityPublic
	status := post.ModerationApproved
	items, total, err := s.workRepo.List(ctx, audiowork.ListFilter{
		AuthorID:         &userID,
		Visibility:       &visibility,
		ModerationStatus: &status,
		Search:           input.Search,
		Tag:              input.Tag,
		Sort:             input.Sort,
		Page:             input.Page,
		PageSize:         input.PageSize,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询用户音频作品失败", err)
	}
	return items, total, nil
}

func (s *AudioWorkService) ListRelatedWorks(ctx context.Context, workID uuid.UUID, viewerID *uuid.UUID, limit int) ([]*audiowork.Work, error) {
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}

	work, err := s.GetWorkForViewer(ctx, workID, viewerID)
	if err != nil {
		return nil, err
	}

	related := make([]*audiowork.Work, 0, limit)
	seen := map[uuid.UUID]struct{}{
		work.ID: {},
	}

	appendUnique := func(items []*audiowork.Work) {
		for _, item := range items {
			if item == nil {
				continue
			}
			if _, exists := seen[item.ID]; exists {
				continue
			}
			seen[item.ID] = struct{}{}
			related = append(related, item)
			if len(related) >= limit {
				return
			}
		}
	}

	authorItems, _, err := s.ListUserPublicWorks(ctx, work.AuthorID, ListAudioWorksInput{
		Page:     1,
		PageSize: limit + 1,
		Sort:     "popular",
	})
	if err != nil {
		return nil, err
	}
	appendUnique(authorItems)

	if len(related) < limit {
		seenTags := make(map[string]struct{})
		for _, tag := range work.Tags {
			normalized := strings.TrimSpace(tag)
			if normalized == "" {
				continue
			}
			if _, exists := seenTags[normalized]; exists {
				continue
			}
			seenTags[normalized] = struct{}{}

			items, _, tagErr := s.ListPublicWorks(ctx, ListAudioWorksInput{
				Tag:      normalized,
				Sort:     "recommended",
				Page:     1,
				PageSize: limit,
			})
			if tagErr != nil {
				return nil, tagErr
			}
			appendUnique(items)
			if len(related) >= limit {
				break
			}
		}
	}

	if len(related) < limit {
		popularItems, _, err := s.ListPublicWorks(ctx, ListAudioWorksInput{
			Sort:     "popular",
			Page:     1,
			PageSize: limit,
		})
		if err != nil {
			return nil, err
		}
		appendUnique(popularItems)
	}

	if len(related) > limit {
		related = related[:limit]
	}
	return related, nil
}

func (s *AudioWorkService) GetWorkForViewer(ctx context.Context, workID uuid.UUID, viewerID *uuid.UUID) (*audiowork.Work, error) {
	work, err := s.workRepo.GetByID(ctx, workID)
	if err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音频作品失败", err)
	}
	if viewerID != nil && work.AuthorID == *viewerID {
		return work, nil
	}
	if work.Visibility != audiowork.VisibilityPublic || work.ModerationStatus != post.ModerationApproved {
		return nil, apperr.ErrNotFound
	}
	return work, nil
}

func (s *AudioWorkService) GetByID(ctx context.Context, workID uuid.UUID) (*audiowork.Work, error) {
	work, err := s.workRepo.GetByID(ctx, workID)
	if err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音频作品失败", err)
	}
	return work, nil
}

func (s *AudioWorkService) UpdateWork(ctx context.Context, input UpdateAudioWorkInput) (*audiowork.Work, error) {
	work, err := s.GetByID(ctx, input.WorkID)
	if err != nil {
		return nil, err
	}
	if work.AuthorID != input.UserID {
		return nil, apperr.ErrForbidden
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, apperr.BadRequest("作品标题不能为空")
	}
	if err := s.validateCoverURL(input.CoverImageURL); err != nil {
		return nil, err
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = work.Visibility
	}
	if visibility != audiowork.VisibilityPublic && visibility != audiowork.VisibilityPrivate {
		return nil, apperr.BadRequest("无效的作品可见性")
	}

	work.Title = title
	work.Visibility = visibility
	work.Tags = normalizeTags(input.Tags)
	work.ModerationStatus = post.ModerationPending
	work.ModerationNote = nil
	work.UpdatedAt = time.Now()

	description := strings.TrimSpace(input.Description)
	if description == "" {
		work.Description = nil
	} else {
		work.Description = &description
	}
	if cover := strings.TrimSpace(input.CoverImageURL); cover == "" {
		work.CoverImageURL = nil
	} else {
		work.CoverImageURL = &cover
	}

	if err := s.workRepo.Update(ctx, work); err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新音频作品失败", err)
	}
	return work, nil
}

func (s *AudioWorkService) DeleteWork(ctx context.Context, userID, workID uuid.UUID) error {
	work, err := s.GetByID(ctx, workID)
	if err != nil {
		return err
	}
	if work.AuthorID != userID {
		return apperr.ErrForbidden
	}

	if s.commentRepo != nil {
		if err := s.commentRepo.DeleteByTarget(ctx, comment.CommentableTypeAudioWork, workID); err != nil {
			return apperr.Wrap(apperr.CodeInternalError, "清理音频作品评论失败", err)
		}
	}
	if s.bookmarkRepo != nil {
		if err := s.bookmarkRepo.DeleteForTarget(ctx, bookmark.TargetAudioWork, workID); err != nil {
			return apperr.Wrap(apperr.CodeInternalError, "清理音频作品收藏失败", err)
		}
	}
	if err := s.workRepo.Delete(ctx, workID); err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "删除音频作品失败", err)
	}
	return nil
}

func (s *AudioWorkService) LikeWork(ctx context.Context, userID, workID uuid.UUID) error {
	if _, err := s.GetByID(ctx, workID); err != nil {
		return err
	}
	liked, err := s.workRepo.HasLiked(ctx, userID, workID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "检查点赞状态失败", err)
	}
	if liked {
		return apperr.BadRequest("已点赞")
	}
	if err := s.workRepo.Like(ctx, userID, workID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "点赞音频作品失败", err)
	}
	if err := s.workRepo.IncrementLikeCount(ctx, workID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新作品点赞数失败", err)
	}
	return nil
}

func (s *AudioWorkService) UnlikeWork(ctx context.Context, userID, workID uuid.UUID) error {
	if _, err := s.GetByID(ctx, workID); err != nil {
		return err
	}
	liked, err := s.workRepo.HasLiked(ctx, userID, workID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "检查点赞状态失败", err)
	}
	if !liked {
		return apperr.BadRequest("未点赞")
	}
	if err := s.workRepo.Unlike(ctx, userID, workID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "取消点赞音频作品失败", err)
	}
	if err := s.workRepo.DecrementLikeCount(ctx, workID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新作品点赞数失败", err)
	}
	return nil
}

func (s *AudioWorkService) HasLiked(ctx context.Context, userID, workID uuid.UUID) (bool, error) {
	liked, err := s.workRepo.HasLiked(ctx, userID, workID)
	if err != nil {
		return false, apperr.Wrap(apperr.CodeInternalError, "查询作品点赞状态失败", err)
	}
	return liked, nil
}

func (s *AudioWorkService) AdminListWorks(ctx context.Context, status string, page, pageSize int) ([]*audiowork.Work, int64, error) {
	filter := audiowork.ListFilter{Page: page, PageSize: pageSize}
	if status != "" {
		ms := post.ModerationStatus(status)
		filter.ModerationStatus = &ms
	}
	return s.workRepo.List(ctx, filter)
}

func (s *AudioWorkService) AdminUpdateModerationStatus(ctx context.Context, workID uuid.UUID, status post.ModerationStatus, note string) error {
	var moderationNote *string
	if trimmed := strings.TrimSpace(note); trimmed != "" {
		moderationNote = &trimmed
	}
	if err := s.workRepo.UpdateModerationStatus(ctx, workID, status, moderationNote); err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "更新音频作品审核状态失败", err)
	}
	return nil
}

func (s *AudioWorkService) validateAudioURL(raw string) error {
	return s.validateMediaURL(raw, "/uploads/processed-audio/", "音频输出地址无效")
}

func (s *AudioWorkService) validateCoverURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return s.validateMediaURL(raw, "/uploads/images/", "封面地址无效")
}

func (s *AudioWorkService) validateMediaURL(raw, localPrefix, invalidMessage string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return apperr.BadRequest(invalidMessage)
	}
	if strings.HasPrefix(raw, localPrefix) {
		return nil
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return apperr.BadRequest(invalidMessage)
	}
	if len(s.allowedHosts) == 0 {
		return nil
	}
	for _, host := range s.allowedHosts {
		if strings.EqualFold(strings.TrimSpace(host), u.Host) {
			return nil
		}
	}
	return apperr.BadRequest("媒体地址不在允许的域名白名单内")
}

func readStringFromMap(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func readDuration(payload map[string]any) float64 {
	meta, ok := payload["output_analysis"].(map[string]any)
	if !ok {
		return 0
	}
	switch value := meta["duration_sec"].(type) {
	case float64:
		return value
	case int:
		return float64(value)
	default:
		return 0
	}
}

func readWaveform(payload map[string]any) []float64 {
	meta, ok := payload["output_analysis"].(map[string]any)
	if !ok {
		return nil
	}
	values, ok := meta["waveform_preview"].([]any)
	if !ok {
		return nil
	}
	result := make([]float64, 0, len(values))
	for _, item := range values {
		switch value := item.(type) {
		case float64:
			result = append(result, value)
		case int:
			result = append(result, float64(value))
		}
	}
	return result
}

func mergeWorkTags(tags []string, payload map[string]any) []string {
	seen := map[string]struct{}{}
	merged := make([]string, 0, len(tags)+4)
	appendTag := func(tag string) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return
		}
		if _, ok := seen[tag]; ok {
			return
		}
		seen[tag] = struct{}{}
		merged = append(merged, tag)
	}

	for _, tag := range tags {
		appendTag(tag)
	}

	if values, ok := payload["style_tags"].([]any); ok {
		for _, raw := range values {
			if text, ok := raw.(string); ok {
				appendTag(text)
			}
		}
	}
	return merged
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func userIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
