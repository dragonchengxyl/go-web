package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/audiojob"
	"github.com/studio/platform/internal/domain/audiowork"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/usecase"
)

func TestAudioWorkServicePublishFromSucceededJob(t *testing.T) {
	t.Parallel()

	jobRepo := newFakeAudioJobRepo()
	workRepo := newFakeAudioWorkRepo()
	userID := uuid.New()
	jobID := uuid.New()

	jobRepo.items[jobID] = &audiojob.Job{
		ID:        jobID,
		UserID:    userID,
		Title:     "夜色人声 demo",
		TaskType:  audiojob.TaskTypeVoiceEnhance,
		Status:    audiojob.StatusSucceeded,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Result: map[string]any{
			"output_audio_url": "/uploads/processed-audio/demo.wav",
			"summary":          "处理完成",
			"style_tags":       []any{"ambient", "night"},
			"output_analysis": map[string]any{
				"duration_sec":     12.5,
				"waveform_preview": []any{0.1, 0.3, 0.7},
			},
		},
	}

	svc := usecase.NewAudioWorkService(workRepo, jobRepo)
	work, err := svc.PublishFromJob(context.Background(), usecase.PublishAudioWorkInput{
		UserID:     userID,
		JobID:      jobID,
		Visibility: audiowork.VisibilityPublic,
	})
	require.NoError(t, err)
	assert.Equal(t, "夜色人声 demo", work.Title)
	assert.Equal(t, "/uploads/processed-audio/demo.wav", work.AudioURL)
	assert.Equal(t, 12.5, work.DurationSec)
	assert.ElementsMatch(t, []string{"ambient", "night"}, work.Tags)
	assert.Len(t, work.WaveformPreview, 3)
}

type fakeAudioWorkRepo struct {
	items map[uuid.UUID]*audiowork.Work
}

func newFakeAudioWorkRepo() *fakeAudioWorkRepo {
	return &fakeAudioWorkRepo{items: make(map[uuid.UUID]*audiowork.Work)}
}

func (r *fakeAudioWorkRepo) Create(_ context.Context, work *audiowork.Work) error {
	for _, existing := range r.items {
		if existing.SourceJobID == work.SourceJobID {
			return audiowork.ErrAlreadyPublished
		}
	}
	copyWork := *work
	r.items[work.ID] = &copyWork
	return nil
}

func (r *fakeAudioWorkRepo) GetByID(_ context.Context, id uuid.UUID) (*audiowork.Work, error) {
	work, ok := r.items[id]
	if !ok {
		return nil, audiowork.ErrNotFound
	}
	copyWork := *work
	return &copyWork, nil
}

func (r *fakeAudioWorkRepo) List(_ context.Context, filter audiowork.ListFilter) ([]*audiowork.Work, int64, error) {
	items := make([]*audiowork.Work, 0)
	for _, work := range r.items {
		if filter.AuthorID != nil && work.AuthorID != *filter.AuthorID {
			continue
		}
		if filter.Visibility != nil && work.Visibility != *filter.Visibility {
			continue
		}
		copyWork := *work
		items = append(items, &copyWork)
	}
	return items, int64(len(items)), nil
}

func (r *fakeAudioWorkRepo) Update(_ context.Context, work *audiowork.Work) error {
	if _, ok := r.items[work.ID]; !ok {
		return audiowork.ErrNotFound
	}
	copyWork := *work
	r.items[work.ID] = &copyWork
	return nil
}

func (r *fakeAudioWorkRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.items[id]; !ok {
		return audiowork.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *fakeAudioWorkRepo) UpdateModerationStatus(_ context.Context, id uuid.UUID, status post.ModerationStatus, note *string) error {
	work, ok := r.items[id]
	if !ok {
		return audiowork.ErrNotFound
	}
	work.ModerationStatus = status
	work.ModerationNote = note
	return nil
}

func (r *fakeAudioWorkRepo) Like(_ context.Context, userID, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	if work.IsLikedByMe {
		return audiowork.ErrAlreadyLiked
	}
	work.IsLikedByMe = true
	return nil
}

func (r *fakeAudioWorkRepo) Unlike(_ context.Context, userID, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	work.IsLikedByMe = false
	return nil
}

func (r *fakeAudioWorkRepo) HasLiked(_ context.Context, userID, workID uuid.UUID) (bool, error) {
	work, ok := r.items[workID]
	if !ok {
		return false, audiowork.ErrNotFound
	}
	return work.IsLikedByMe, nil
}

func (r *fakeAudioWorkRepo) IncrementLikeCount(_ context.Context, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	work.LikeCount++
	return nil
}

func (r *fakeAudioWorkRepo) DecrementLikeCount(_ context.Context, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	if work.LikeCount > 0 {
		work.LikeCount--
	}
	return nil
}

func (r *fakeAudioWorkRepo) IncrementCommentCount(_ context.Context, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	work.CommentCount++
	return nil
}

func (r *fakeAudioWorkRepo) DecrementCommentCount(_ context.Context, workID uuid.UUID) error {
	work, ok := r.items[workID]
	if !ok {
		return audiowork.ErrNotFound
	}
	if work.CommentCount > 0 {
		work.CommentCount--
	}
	return nil
}
