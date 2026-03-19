package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audiojob"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

type AudioJobService struct {
	repo         audiojob.Repository
	publisher    *streams.Publisher
	logger       *zap.Logger
	allowedHosts []string
	processor    audioJobProcessor
	maxAttempts  int
	retryBackoff time.Duration
}

type AudioJobServiceOption func(*AudioJobService)

type audioJobProcessor interface {
	Process(ctx context.Context, job *audiojob.Job) (map[string]any, error)
}

func NewAudioJobService(repo audiojob.Repository, opts ...AudioJobServiceOption) *AudioJobService {
	svc := &AudioJobService{
		repo:         repo,
		maxAttempts:  3,
		retryBackoff: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithAudioJobPublisher(p *streams.Publisher) AudioJobServiceOption {
	return func(s *AudioJobService) {
		s.publisher = p
	}
}

func WithAudioJobLogger(logger *zap.Logger) AudioJobServiceOption {
	return func(s *AudioJobService) {
		s.logger = logger
	}
}

func WithAudioJobAllowedHosts(hosts []string) AudioJobServiceOption {
	return func(s *AudioJobService) {
		s.allowedHosts = append([]string(nil), hosts...)
	}
}

func WithAudioJobProcessor(processor audioJobProcessor) AudioJobServiceOption {
	return func(s *AudioJobService) {
		s.processor = processor
	}
}

func WithAudioJobRetryPolicy(maxAttempts int, retryBackoff time.Duration) AudioJobServiceOption {
	return func(s *AudioJobService) {
		if maxAttempts > 0 {
			s.maxAttempts = maxAttempts
		}
		if retryBackoff > 0 {
			s.retryBackoff = retryBackoff
		}
	}
}

type CreateAudioJobInput struct {
	UserID            uuid.UUID
	Title             string
	TaskType          audiojob.TaskType
	SourceAudioURL    string
	ReferenceAudioURL string
	Prompt            string
	Params            map[string]any
}

type ListAudioJobsInput struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
	Status   *audiojob.Status
	TaskType *audiojob.TaskType
}

func (s *AudioJobService) CreateJob(ctx context.Context, input CreateAudioJobInput) (*audiojob.Job, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}

	sourceURL, err := s.normalizeAudioURL(input.SourceAudioURL)
	if err != nil {
		return nil, err
	}
	refURL, err := s.normalizeAudioURL(input.ReferenceAudioURL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	job := &audiojob.Job{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Title:       strings.TrimSpace(input.Title),
		TaskType:    input.TaskType,
		Status:      audiojob.StatusQueued,
		Params:      cloneMap(input.Params),
		MaxAttempts: s.maxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if sourceURL != "" {
		job.SourceAudioURL = &sourceURL
	}
	if refURL != "" {
		job.ReferenceAudioURL = &refURL
	}
	if prompt := strings.TrimSpace(input.Prompt); prompt != "" {
		job.Prompt = &prompt
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建音频任务失败", err)
	}

	s.enqueueJob(ctx, job.ID)
	return job, nil
}

func (s *AudioJobService) ListJobs(ctx context.Context, input ListAudioJobsInput) ([]*audiojob.Job, int64, error) {
	items, total, err := s.repo.ListByUser(ctx, audiojob.ListFilter{
		UserID:   input.UserID,
		Page:     input.Page,
		PageSize: input.PageSize,
		Status:   input.Status,
		TaskType: input.TaskType,
	})
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询音频任务失败", err)
	}
	return items, total, nil
}

func (s *AudioJobService) GetJob(ctx context.Context, userID, jobID uuid.UUID) (*audiojob.Job, error) {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, audiojob.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询音频任务失败", err)
	}
	if job.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return job, nil
}

func (s *AudioJobService) RetryJob(ctx context.Context, userID, jobID uuid.UUID) (*audiojob.Job, error) {
	job, err := s.GetJob(ctx, userID, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != audiojob.StatusFailed && job.Status != audiojob.StatusDeadLettered {
		return nil, apperr.BadRequest("只有失败或死信任务可以重试")
	}

	now := time.Now()
	job.Status = audiojob.StatusQueued
	job.ErrorMessage = nil
	job.Result = nil
	job.AttemptCount = 0
	job.StartedAt = nil
	job.FinishedAt = nil
	job.NextRetryAt = nil
	job.LastErrorAt = nil
	job.DeadLetteredAt = nil
	job.UpdatedAt = now

	if err := s.repo.Update(ctx, job); err != nil {
		if errors.Is(err, audiojob.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "重试音频任务失败", err)
	}

	s.enqueueJob(ctx, job.ID)
	return job, nil
}

func (s *AudioJobService) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	job, claimed, err := s.repo.ClaimForProcessing(ctx, jobID, time.Now())
	if err != nil {
		return fmt.Errorf("claim audio job: %w", err)
	}
	if !claimed {
		return nil
	}

	result, err := s.processJobResult(ctx, job)
	finishedAt := time.Now()
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt

	if err != nil {
		msg := err.Error()
		job.LastErrorAt = &finishedAt
		job.ErrorMessage = &msg
		job.Result = nil
		if job.AttemptCount >= job.MaxAttempts {
			job.Status = audiojob.StatusDeadLettered
			job.DeadLetteredAt = &finishedAt
			job.NextRetryAt = nil
		} else {
			job.Status = audiojob.StatusQueued
			nextRetryAt := finishedAt.Add(s.retryDelay(job.AttemptCount))
			job.NextRetryAt = &nextRetryAt
			job.DeadLetteredAt = nil
		}
	} else {
		job.Status = audiojob.StatusSucceeded
		job.ErrorMessage = nil
		job.LastErrorAt = nil
		job.NextRetryAt = nil
		job.DeadLetteredAt = nil
		job.Result = result
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("finalize audio job: %w", err)
	}
	return nil
}

func (s *AudioJobService) ListDueRetryIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	ids, err := s.repo.ListDueRetryIDs(ctx, time.Now(), limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询待重试音频任务失败", err)
	}
	return ids, nil
}

func (s *AudioJobService) enqueueJob(ctx context.Context, jobID uuid.UUID) {
	if s.publisher != nil {
		err := s.publisher.Publish(ctx, streams.EventAudioJobCreated, streams.AudioJobCreatedPayload{
			JobID: jobID.String(),
		})
		if err == nil {
			return
		}
		if s.logger != nil {
			s.logger.Error("publish audio.job.created failed, falling back to local processing",
				zap.Error(err),
				zap.String("job_id", jobID.String()),
			)
		}
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.ProcessJob(bgCtx, jobID); err != nil && s.logger != nil {
			s.logger.Error("local audio job processing failed", zap.Error(err), zap.String("job_id", jobID.String()))
		}
	}()
}

func (s *AudioJobService) validateCreateInput(input CreateAudioJobInput) error {
	if input.UserID == uuid.Nil {
		return apperr.ErrUnauthorized
	}
	if strings.TrimSpace(input.Title) == "" {
		return apperr.BadRequest("任务标题不能为空")
	}
	if !isSupportedAudioTaskType(input.TaskType) {
		return apperr.BadRequest("不支持的音频任务类型")
	}

	source := strings.TrimSpace(input.SourceAudioURL)
	reference := strings.TrimSpace(input.ReferenceAudioURL)
	prompt := strings.TrimSpace(input.Prompt)

	switch input.TaskType {
	case audiojob.TaskTypeAIMusic:
		if prompt == "" {
			return apperr.BadRequest("AI 作曲任务需要 prompt")
		}
	case audiojob.TaskTypeVoiceConvert:
		if source == "" || reference == "" {
			return apperr.BadRequest("音色转换任务需要源音频和参考音频")
		}
	case audiojob.TaskTypeVoiceEnhance, audiojob.TaskTypeAudioMaster:
		if source == "" {
			return apperr.BadRequest("当前任务需要上传源音频")
		}
	}

	return nil
}

func (s *AudioJobService) normalizeAudioURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if strings.HasPrefix(raw, "/uploads/audio/") {
		return raw, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", apperr.BadRequest("音频地址无效")
	}
	if u.Scheme == "" || u.Host == "" {
		return "", apperr.BadRequest("音频地址必须是完整 URL 或本地上传地址")
	}
	for _, host := range s.allowedHosts {
		if strings.EqualFold(strings.TrimSpace(host), u.Host) {
			return raw, nil
		}
	}
	return "", apperr.BadRequest("音频地址不在允许的域名白名单内")
}

func isSupportedAudioTaskType(taskType audiojob.TaskType) bool {
	switch taskType {
	case audiojob.TaskTypeAIMusic, audiojob.TaskTypeVoiceConvert, audiojob.TaskTypeVoiceEnhance, audiojob.TaskTypeAudioMaster:
		return true
	default:
		return false
	}
}

func buildMockAudioResult(job *audiojob.Job) (map[string]any, error) {
	result := map[string]any{
		"mock":         true,
		"provider":     "mock-audio-pipeline",
		"task_type":    job.TaskType,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	switch job.TaskType {
	case audiojob.TaskTypeAIMusic:
		if job.Prompt == nil || strings.TrimSpace(*job.Prompt) == "" {
			return nil, fmt.Errorf("prompt 不能为空")
		}
		result["summary"] = "已根据提示词生成一版歌曲草案，可继续进入编曲和人声制作。"
		result["arrangement"] = []string{"intro", "verse", "hook", "bridge", "outro"}
		result["style_tags"] = extractStyleTags(*job.Prompt)
	case audiojob.TaskTypeVoiceConvert:
		if job.SourceAudioURL == nil || job.ReferenceAudioURL == nil {
			return nil, fmt.Errorf("缺少音色转换所需的音频输入")
		}
		result["summary"] = "已完成 mock 音色转换，当前结果复用了源音频地址，便于先打通全链路。"
		result["output_audio_url"] = *job.SourceAudioURL
		result["reference_audio_url"] = *job.ReferenceAudioURL
		result["quality_report"] = map[string]any{
			"clarity":       "good",
			"pitch_stable":  true,
			"speaker_match": "medium",
		}
	case audiojob.TaskTypeVoiceEnhance:
		if job.SourceAudioURL == nil {
			return nil, fmt.Errorf("缺少待增强音频")
		}
		result["summary"] = "已完成 mock 降噪和响度均衡，可作为发布前预处理结果。"
		result["output_audio_url"] = *job.SourceAudioURL
		result["metrics"] = map[string]any{
			"noise_reduction_db": 12,
			"target_lufs":        -14,
			"peak_db":            -1.0,
		}
	case audiojob.TaskTypeAudioMaster:
		if job.SourceAudioURL == nil {
			return nil, fmt.Errorf("缺少待母带处理音频")
		}
		result["summary"] = "已完成 mock 母带处理，补齐了响度、峰值和交付说明。"
		result["output_audio_url"] = *job.SourceAudioURL
		result["delivery"] = map[string]any{
			"format":      "wav",
			"sample_rate": 48000,
			"bit_depth":   24,
		}
	default:
		return nil, fmt.Errorf("unsupported task type: %s", job.TaskType)
	}

	if job.Params != nil {
		paramsJSON, _ := json.Marshal(job.Params)
		result["params_echo"] = string(paramsJSON)
	}
	return result, nil
}

func (s *AudioJobService) processJobResult(ctx context.Context, job *audiojob.Job) (map[string]any, error) {
	if s.processor != nil {
		result, err := s.processor.Process(ctx, job)
		if err == nil {
			return result, nil
		}
		if s.logger != nil {
			s.logger.Warn("audio processor failed, falling back to mock result",
				zap.Error(err),
				zap.String("job_id", job.ID.String()),
				zap.String("task_type", string(job.TaskType)),
			)
		}
	}
	return buildMockAudioResult(job)
}

func (s *AudioJobService) retryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return s.retryBackoff
	}
	return time.Duration(attempt) * s.retryBackoff
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	output := make(map[string]any, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func extractStyleTags(prompt string) []string {
	lower := strings.ToLower(prompt)
	tags := make([]string, 0, 4)
	if strings.Contains(lower, "rock") || strings.Contains(lower, "摇滚") {
		tags = append(tags, "rock")
	}
	if strings.Contains(lower, "电子") || strings.Contains(lower, "edm") {
		tags = append(tags, "electronic")
	}
	if strings.Contains(lower, "抒情") || strings.Contains(lower, "ballad") {
		tags = append(tags, "ballad")
	}
	if strings.Contains(lower, "动漫") || strings.Contains(lower, "二次元") {
		tags = append(tags, "anime")
	}
	if len(tags) == 0 {
		tags = append(tags, "demo")
	}
	return tags
}
