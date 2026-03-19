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
	"github.com/studio/platform/internal/pkg/apperr"
)

type AudioWorkService struct {
	workRepo     audiowork.Repository
	audioJobRepo audiojob.Repository
	allowedHosts []string
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
	Page     int
	PageSize int
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
		ID:              uuid.New(),
		AuthorID:        input.UserID,
		SourceJobID:     input.JobID,
		Title:           title,
		AudioURL:        audioURL,
		DurationSec:     readDuration(job.Result),
		Visibility:      visibility,
		Tags:            mergeWorkTags(input.Tags, job.Result),
		WaveformPreview: readWaveform(job.Result),
		Metadata:        cloneMap(job.Result),
		PublishedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
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
	items, total, err := s.workRepo.List(ctx, audiowork.ListFilter{
		Visibility: &visibility,
		Page:       input.Page,
		PageSize:   input.PageSize,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询音频作品失败", err)
	}
	return items, total, nil
}

func (s *AudioWorkService) ListMyWorks(ctx context.Context, userID uuid.UUID, input ListAudioWorksInput) ([]*audiowork.Work, int64, error) {
	items, total, err := s.workRepo.List(ctx, audiowork.ListFilter{
		AuthorID: &userID,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询我的音频作品失败", err)
	}
	return items, total, nil
}

func (s *AudioWorkService) GetPublicWork(ctx context.Context, workID uuid.UUID) (*audiowork.Work, error) {
	work, err := s.workRepo.GetByID(ctx, workID)
	if err != nil {
		if errors.Is(err, audiowork.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音频作品失败", err)
	}
	if work.Visibility != audiowork.VisibilityPublic {
		return nil, apperr.ErrNotFound
	}
	return work, nil
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
