package usecase_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/audiojob"
	"github.com/studio/platform/internal/usecase"
)

func TestAudioJobServiceProcessesQueuedJobLocally(t *testing.T) {
	t.Parallel()

	repo := newFakeAudioJobRepo()
	svc := usecase.NewAudioJobService(repo)
	userID := uuid.New()

	job, err := svc.CreateJob(context.Background(), usecase.CreateAudioJobInput{
		UserID:         userID,
		Title:          "人声增强测试",
		TaskType:       audiojob.TaskTypeVoiceEnhance,
		SourceAudioURL: "/uploads/audio/demo.wav",
	})
	require.NoError(t, err)
	require.Equal(t, audiojob.StatusQueued, job.Status)

	require.Eventually(t, func() bool {
		stored, err := repo.GetByID(context.Background(), job.ID)
		require.NoError(t, err)
		return stored.Status == audiojob.StatusSucceeded
	}, 2*time.Second, 20*time.Millisecond)

	stored, err := repo.GetByID(context.Background(), job.ID)
	require.NoError(t, err)
	assert.NotNil(t, stored.Result)
	assert.Equal(t, "/uploads/audio/demo.wav", stored.Result["output_audio_url"])
}

func TestAudioJobServiceRejectsIncompleteVoiceConvertJob(t *testing.T) {
	t.Parallel()

	repo := newFakeAudioJobRepo()
	svc := usecase.NewAudioJobService(repo)

	_, err := svc.CreateJob(context.Background(), usecase.CreateAudioJobInput{
		UserID:         uuid.New(),
		Title:          "音色转换测试",
		TaskType:       audiojob.TaskTypeVoiceConvert,
		SourceAudioURL: "/uploads/audio/source.wav",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "参考音频")
}

type fakeAudioJobRepo struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*audiojob.Job
}

func newFakeAudioJobRepo() *fakeAudioJobRepo {
	return &fakeAudioJobRepo{items: make(map[uuid.UUID]*audiojob.Job)}
}

func (r *fakeAudioJobRepo) Create(_ context.Context, job *audiojob.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[job.ID] = cloneAudioJob(job)
	return nil
}

func (r *fakeAudioJobRepo) GetByID(_ context.Context, id uuid.UUID) (*audiojob.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.items[id]
	if !ok {
		return nil, audiojob.ErrNotFound
	}
	return cloneAudioJob(job), nil
}

func (r *fakeAudioJobRepo) ListByUser(_ context.Context, filter audiojob.ListFilter) ([]*audiojob.Job, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]*audiojob.Job, 0)
	for _, item := range r.items {
		if item.UserID != filter.UserID {
			continue
		}
		items = append(items, cloneAudioJob(item))
	}
	return items, int64(len(items)), nil
}

func (r *fakeAudioJobRepo) Update(_ context.Context, job *audiojob.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[job.ID]; !ok {
		return audiojob.ErrNotFound
	}
	r.items[job.ID] = cloneAudioJob(job)
	return nil
}

func cloneAudioJob(job *audiojob.Job) *audiojob.Job {
	if job == nil {
		return nil
	}
	copyJob := *job
	if job.Params != nil {
		copyJob.Params = make(map[string]any, len(job.Params))
		for k, v := range job.Params {
			copyJob.Params[k] = v
		}
	}
	if job.Result != nil {
		copyJob.Result = make(map[string]any, len(job.Result))
		for k, v := range job.Result {
			copyJob.Result[k] = v
		}
	}
	return &copyJob
}
