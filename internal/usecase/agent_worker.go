package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

type AgentWorker struct {
	service        *AgentService
	logger         *zap.Logger
	pollInterval   time.Duration
	retryBatchSize int
}

func NewAgentWorker(service *AgentService, logger *zap.Logger, pollInterval time.Duration, retryBatchSize int) *AgentWorker {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	if retryBatchSize <= 0 {
		retryBatchSize = 20
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AgentWorker{
		service:        service,
		logger:         logger,
		pollInterval:   pollInterval,
		retryBatchSize: retryBatchSize,
	}
}

func (w *AgentWorker) Start(ctx context.Context) {
	go w.pollRuns(ctx)
}

func (w *AgentWorker) pollRuns(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ids, err := w.service.ListRunnableRunIDs(ctx, w.retryBatchSize)
			if err != nil {
				w.logger.Error("list runnable agent runs failed", zap.Error(err))
				continue
			}
			for _, id := range ids {
				if err := w.service.ProcessRun(ctx, id); err != nil {
					w.logger.Error("process agent run failed", zap.Error(err), zap.String("run_id", id.String()))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *AgentService) ListRunnableRunIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if !s.Enabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	ids, err := s.repo.ListRunnableRunIDs(ctx, time.Now(), limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取可执行 Agent 任务失败", err)
	}
	return ids, nil
}

func (s *AgentService) ProcessRun(ctx context.Context, runID uuid.UUID) error {
	if !s.Enabled() {
		return apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	run, claimed, err := s.repo.ClaimRunForProcessing(ctx, runID, time.Now())
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "领取 Agent 任务失败", err)
	}
	if !claimed || run == nil {
		return nil
	}
	if err := s.executeRun(ctx, run); err != nil {
		s.handleRunExecutionFailure(ctx, run, err)
		return err
	}
	return nil
}
