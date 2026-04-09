package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

type AgentWorker struct {
	service           *AgentService
	logger            *zap.Logger
	pollInterval      time.Duration
	retryBatchSize    int
	workerID          string
	leaseTTL          time.Duration
	heartbeatInterval time.Duration
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
		service:           service,
		logger:            logger,
		pollInterval:      pollInterval,
		retryBatchSize:    retryBatchSize,
		workerID:          "agent-worker-1",
		leaseTTL:          30 * time.Second,
		heartbeatInterval: 10 * time.Second,
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
				if err := w.service.ProcessRun(ctx, id, w.workerID, w.leaseTTL, w.heartbeatInterval); err != nil {
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

func (s *AgentService) ProcessRun(ctx context.Context, runID uuid.UUID, workerID string, leaseTTL, heartbeatInterval time.Duration) error {
	if !s.Enabled() {
		return apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	run, claimed, err := s.repo.ClaimRunForProcessing(ctx, runID, workerID, time.Now(), leaseTTL)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "领取 Agent 任务失败", err)
	}
	if !claimed || run == nil {
		return nil
	}

	leaseCtx, stopHeartbeat := s.startRunLeaseHeartbeat(ctx, run.ID, workerID, leaseTTL, heartbeatInterval)
	defer stopHeartbeat()

	if err := s.executeRun(leaseCtx, run); err != nil {
		s.handleRunExecutionFailure(leaseCtx, run, err)
		return err
	}
	return nil
}

func (s *AgentService) startRunLeaseHeartbeat(ctx context.Context, runID uuid.UUID, workerID string, leaseTTL, heartbeatInterval time.Duration) (context.Context, context.CancelFunc) {
	if leaseTTL <= 0 {
		leaseTTL = 30 * time.Second
	}
	if heartbeatInterval <= 0 {
		heartbeatInterval = leaseTTL / 3
	}
	heartbeatCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				_ = s.repo.RenewRunLease(heartbeatCtx, runID, workerID, now, now.Add(leaseTTL))
			case <-heartbeatCtx.Done():
				return
			}
		}
	}()
	return heartbeatCtx, cancel
}
