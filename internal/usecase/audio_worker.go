package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/infra/eventbus"
	"go.uber.org/zap"
)

type AudioWorker struct {
	consumer       eventbus.Consumer
	service        *AudioJobService
	logger         *zap.Logger
	pollInterval   time.Duration
	retryBatchSize int
}

func NewAudioWorker(consumer eventbus.Consumer, service *AudioJobService, logger *zap.Logger, pollInterval time.Duration, retryBatchSize int) *AudioWorker {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	if retryBatchSize <= 0 {
		retryBatchSize = 20
	}
	return &AudioWorker{
		consumer:       consumer,
		service:        service,
		logger:         logger,
		pollInterval:   pollInterval,
		retryBatchSize: retryBatchSize,
	}
}

func (w *AudioWorker) Start(ctx context.Context) {
	go w.consumeStream(ctx)
	go w.pollRetries(ctx)
}

func (w *AudioWorker) consumeStream(ctx context.Context) {
	w.logger.Info("Starting audio job stream consumer")
	_ = w.consumer.Start(ctx, eventbus.GroupAudioJobs, func(ctx context.Context, ev eventbus.Event) error {
		if ev.Type != eventbus.EventAudioJobCreated {
			return nil
		}

		var payload eventbus.AudioJobCreatedPayload
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			return fmt.Errorf("audio worker: unmarshal payload: %w", err)
		}
		jobID, err := uuid.Parse(payload.JobID)
		if err != nil {
			return fmt.Errorf("audio worker: invalid job id: %w", err)
		}
		return w.service.ProcessJob(ctx, jobID)
	})
}

func (w *AudioWorker) pollRetries(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ids, err := w.service.ListDueRetryIDs(ctx, w.retryBatchSize)
			if err != nil {
				w.logger.Error("list due retry audio jobs failed", zap.Error(err))
				continue
			}
			for _, id := range ids {
				if err := w.service.ProcessJob(ctx, id); err != nil {
					w.logger.Error("retry audio job failed", zap.Error(err), zap.String("job_id", id.String()))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
