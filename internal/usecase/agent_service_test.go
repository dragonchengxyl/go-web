package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	agentdomain "github.com/studio/platform/internal/domain/agent"
)

type inMemoryAgentRepo struct {
	runs      map[uuid.UUID]*agentdomain.Run
	steps     map[uuid.UUID][]*agentdomain.Step
	toolCalls map[uuid.UUID][]*agentdomain.ToolCall
	approvals map[uuid.UUID][]*agentdomain.Approval
	artifacts map[uuid.UUID][]*agentdomain.Artifact
}

func newInMemoryAgentRepo() *inMemoryAgentRepo {
	return &inMemoryAgentRepo{
		runs:      make(map[uuid.UUID]*agentdomain.Run),
		steps:     make(map[uuid.UUID][]*agentdomain.Step),
		toolCalls: make(map[uuid.UUID][]*agentdomain.ToolCall),
		approvals: make(map[uuid.UUID][]*agentdomain.Approval),
		artifacts: make(map[uuid.UUID][]*agentdomain.Artifact),
	}
}

func (r *inMemoryAgentRepo) CreateRun(_ context.Context, item *agentdomain.Run) error {
	clone := *item
	r.runs[item.ID] = &clone
	return nil
}

func (r *inMemoryAgentRepo) GetRunByID(_ context.Context, id uuid.UUID) (*agentdomain.Run, error) {
	item, ok := r.runs[id]
	if !ok {
		return nil, agentdomain.ErrRunNotFound
	}
	clone := *item
	return &clone, nil
}

func (r *inMemoryAgentRepo) ListRuns(_ context.Context, userID uuid.UUID, _, _ int) ([]*agentdomain.Run, int64, error) {
	items := make([]*agentdomain.Run, 0)
	for _, item := range r.runs {
		if item.UserID != userID {
			continue
		}
		clone := *item
		items = append(items, &clone)
	}
	return items, int64(len(items)), nil
}

func (r *inMemoryAgentRepo) TryStartRun(_ context.Context, id uuid.UUID, startedAt time.Time) (bool, error) {
	item, claimed, err := r.ClaimRunForProcessing(context.Background(), id, startedAt)
	if err != nil {
		return false, err
	}
	return claimed && item != nil, nil
}

func (r *inMemoryAgentRepo) ClaimRunForProcessing(_ context.Context, id uuid.UUID, now time.Time) (*agentdomain.Run, bool, error) {
	item, ok := r.runs[id]
	if !ok {
		return nil, false, agentdomain.ErrRunNotFound
	}
	if item.Status != agentdomain.RunStatusQueued {
		return nil, false, nil
	}
	if item.NextRetryAt != nil && item.NextRetryAt.After(now) {
		return nil, false, nil
	}
	item.Status = agentdomain.RunStatusRunning
	item.AttemptCount++
	item.StartedAt = &now
	item.CompletedAt = nil
	item.NextRetryAt = nil
	item.LastError = ""
	item.UpdatedAt = now
	clone := *item
	return &clone, true, nil
}

func (r *inMemoryAgentRepo) ListRunnableRunIDs(_ context.Context, readyBefore time.Time, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 20
	}
	ids := make([]uuid.UUID, 0, limit)
	for id, item := range r.runs {
		if item.Status != agentdomain.RunStatusQueued {
			continue
		}
		if item.NextRetryAt != nil && item.NextRetryAt.After(readyBefore) {
			continue
		}
		ids = append(ids, id)
		if len(ids) >= limit {
			break
		}
	}
	return ids, nil
}

func (r *inMemoryAgentRepo) UpdateRun(_ context.Context, item *agentdomain.Run) error {
	_, ok := r.runs[item.ID]
	if !ok {
		return agentdomain.ErrRunNotFound
	}
	clone := *item
	r.runs[item.ID] = &clone
	return nil
}

func (r *inMemoryAgentRepo) UpdateRunStatus(_ context.Context, id uuid.UUID, status agentdomain.RunStatus, summary, lastError string, completedAt *time.Time) error {
	item, ok := r.runs[id]
	if !ok {
		return agentdomain.ErrRunNotFound
	}
	item.Status = status
	item.LatestSummary = summary
	item.LastError = lastError
	item.UpdatedAt = time.Now()
	item.CompletedAt = completedAt
	if lastError != "" {
		now := time.Now()
		item.LastErrorAt = &now
	}
	return nil
}

func (r *inMemoryAgentRepo) CreateStep(_ context.Context, item *agentdomain.Step) error {
	clone := *item
	r.steps[item.RunID] = append(r.steps[item.RunID], &clone)
	return nil
}

func (r *inMemoryAgentRepo) ListSteps(_ context.Context, runID uuid.UUID) ([]*agentdomain.Step, error) {
	items := r.steps[runID]
	out := make([]*agentdomain.Step, 0, len(items))
	for _, item := range items {
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *inMemoryAgentRepo) CreateToolCall(_ context.Context, item *agentdomain.ToolCall) error {
	clone := *item
	r.toolCalls[item.RunID] = append(r.toolCalls[item.RunID], &clone)
	return nil
}

func (r *inMemoryAgentRepo) ListToolCalls(_ context.Context, runID uuid.UUID) ([]*agentdomain.ToolCall, error) {
	items := r.toolCalls[runID]
	out := make([]*agentdomain.ToolCall, 0, len(items))
	for _, item := range items {
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *inMemoryAgentRepo) CreateApproval(_ context.Context, item *agentdomain.Approval) error {
	clone := *item
	r.approvals[item.RunID] = append(r.approvals[item.RunID], &clone)
	return nil
}

func (r *inMemoryAgentRepo) GetApprovalByID(_ context.Context, id uuid.UUID) (*agentdomain.Approval, error) {
	for _, items := range r.approvals {
		for _, item := range items {
			if item.ID != id {
				continue
			}
			clone := *item
			return &clone, nil
		}
	}
	return nil, agentdomain.ErrRunNotFound
}

func (r *inMemoryAgentRepo) ListApprovals(_ context.Context, runID uuid.UUID) ([]*agentdomain.Approval, error) {
	items := r.approvals[runID]
	out := make([]*agentdomain.Approval, 0, len(items))
	for _, item := range items {
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *inMemoryAgentRepo) UpdateApprovalStatus(_ context.Context, id uuid.UUID, status agentdomain.ApprovalStatus, approvedBy *uuid.UUID, approvedAt *time.Time) error {
	for _, items := range r.approvals {
		for _, item := range items {
			if item.ID != id {
				continue
			}
			item.Status = status
			item.ApprovedBy = approvedBy
			item.ApprovedAt = approvedAt
			return nil
		}
	}
	return agentdomain.ErrRunNotFound
}

func (r *inMemoryAgentRepo) CreateArtifact(_ context.Context, item *agentdomain.Artifact) error {
	clone := *item
	r.artifacts[item.RunID] = append(r.artifacts[item.RunID], &clone)
	return nil
}

func (r *inMemoryAgentRepo) ListArtifacts(_ context.Context, runID uuid.UUID) ([]*agentdomain.Artifact, error) {
	items := r.artifacts[runID]
	out := make([]*agentdomain.Artifact, 0, len(items))
	for _, item := range items {
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func TestAgentServiceCreateRunQueuesWhenAsync(t *testing.T) {
	repo := newInMemoryAgentRepo()
	svc := newAgentService(repo, nil, true)
	userID := uuid.New()

	detail, err := svc.CreateRun(context.Background(), userID, CreateAgentRunInput{
		Scenario: agentdomain.ScenarioPostAgent,
		Goal:     "请帮我整理这条动态草稿",
	})
	if err != nil {
		t.Fatalf("CreateRun error: %v", err)
	}
	if detail.Run.Status != agentdomain.RunStatusQueued {
		t.Fatalf("run status = %q, want queued", detail.Run.Status)
	}
	if detail.Run.MaxAttempts != 3 {
		t.Fatalf("max attempts = %d, want 3", detail.Run.MaxAttempts)
	}
	if detail.Run.AttemptCount != 0 {
		t.Fatalf("attempt count = %d, want 0", detail.Run.AttemptCount)
	}
}

func TestAgentWorkerProcessRunGeneratesPostArtifacts(t *testing.T) {
	repo := newInMemoryAgentRepo()
	svc := newAgentService(repo, nil, true)
	worker := NewAgentWorker(svc, nil, time.Millisecond, 10)
	_ = worker
	userID := uuid.New()

	detail, err := svc.CreateRun(context.Background(), userID, CreateAgentRunInput{
		Scenario: agentdomain.ScenarioPostAgent,
		Goal:     "请帮我整理这条动态草稿",
		ContextSnapshot: map[string]any{
			"draft_title":   "春日兽设草稿",
			"draft_content": "最近把角色设定调整了一版，想分享新的配色和情绪氛围。",
			"group_name":    "原创设定研究所",
			"visibility":    "公开",
		},
	})
	if err != nil {
		t.Fatalf("CreateRun error: %v", err)
	}
	if detail.Run.Status != agentdomain.RunStatusQueued {
		t.Fatalf("run status = %q, want queued", detail.Run.Status)
	}

	if err := svc.ProcessRun(context.Background(), detail.Run.ID); err != nil {
		t.Fatalf("ProcessRun error: %v", err)
	}

	detail, err = svc.GetRunDetail(context.Background(), userID, detail.Run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail error: %v", err)
	}
	if detail.Run.Status != agentdomain.RunStatusWaitingApproval {
		t.Fatalf("run status = %q, want waiting_approval", detail.Run.Status)
	}
	if detail.Run.AttemptCount != 1 {
		t.Fatalf("attempt count = %d, want 1", detail.Run.AttemptCount)
	}
	if len(detail.ToolCalls) < 2 {
		t.Fatalf("expected tool calls, got %d", len(detail.ToolCalls))
	}
}

func TestAgentServiceCreateRunGeneratesPostArtifacts(t *testing.T) {
	repo := newInMemoryAgentRepo()
	svc := newAgentService(repo, nil, false)
	userID := uuid.New()

	detail, err := svc.CreateRun(context.Background(), userID, CreateAgentRunInput{
		Scenario: agentdomain.ScenarioPostAgent,
		Goal:     "请帮我整理这条动态草稿",
		ContextSnapshot: map[string]any{
			"draft_title":   "春日兽设草稿",
			"draft_content": "最近把角色设定调整了一版，想分享新的配色和情绪氛围。",
			"group_name":    "原创设定研究所",
			"visibility":    "公开",
		},
	})
	if err != nil {
		t.Fatalf("CreateRun error: %v", err)
	}
	if detail.Run.Status != agentdomain.RunStatusWaitingApproval {
		t.Fatalf("run status = %q, want waiting_approval", detail.Run.Status)
	}
	if len(detail.Steps) < 3 {
		t.Fatalf("expected multiple steps, got %d", len(detail.Steps))
	}
	if len(detail.ToolCalls) < 2 {
		t.Fatalf("expected tool calls, got %d", len(detail.ToolCalls))
	}
	toolNames := make(map[string]struct{}, len(detail.ToolCalls))
	for _, item := range detail.ToolCalls {
		toolNames[item.ToolName] = struct{}{}
	}
	for _, expected := range []string{
		"get_post_create_context",
		"generate_post_copilot_artifacts",
		"build_post_draft_patch",
	} {
		if _, ok := toolNames[expected]; !ok {
			t.Fatalf("expected tool %q to be executed, got %+v", expected, detail.ToolCalls)
		}
	}
	if len(detail.Artifacts) == 0 {
		t.Fatalf("expected artifacts to be generated")
	}
	if len(detail.Approvals) != 1 {
		t.Fatalf("expected one approval, got %d", len(detail.Approvals))
	}
	foundTitleOptions := false
	foundDraftPatch := false
	for _, item := range detail.Artifacts {
		if item.Kind == "title_options" {
			foundTitleOptions = true
		}
		if item.Kind == "draft_patch" {
			foundDraftPatch = true
		}
	}
	if !foundTitleOptions {
		t.Fatalf("expected title_options artifact, got %+v", detail.Artifacts)
	}
	if !foundDraftPatch {
		t.Fatalf("expected draft_patch artifact, got %+v", detail.Artifacts)
	}

	result, err := svc.DecideApproval(context.Background(), userID, detail.Run.ID, DecideAgentApprovalInput{
		ApprovalID: detail.Approvals[0].ID.String(),
		Decision:   "approved",
	})
	if err != nil {
		t.Fatalf("DecideApproval error: %v", err)
	}
	if result.Detail.Run.Status != agentdomain.RunStatusCompleted {
		t.Fatalf("approved run status = %q, want completed", result.Detail.Run.Status)
	}
	if result.ApplyPayload == nil {
		t.Fatalf("expected apply payload on approval")
	}
	if got := result.ApplyPayload["visibility"]; got == nil {
		t.Fatalf("expected visibility in apply payload")
	}
}
