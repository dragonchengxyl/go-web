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

func (r *inMemoryAgentRepo) ListApprovals(_ context.Context, runID uuid.UUID) ([]*agentdomain.Approval, error) {
	items := r.approvals[runID]
	out := make([]*agentdomain.Approval, 0, len(items))
	for _, item := range items {
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
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

func TestAgentServiceCreateRunGeneratesPostArtifacts(t *testing.T) {
	repo := newInMemoryAgentRepo()
	svc := NewAgentService(repo, nil)
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
	if detail.Run.Status != agentdomain.RunStatusCompleted {
		t.Fatalf("run status = %q, want completed", detail.Run.Status)
	}
	if len(detail.Steps) < 3 {
		t.Fatalf("expected multiple steps, got %d", len(detail.Steps))
	}
	if len(detail.ToolCalls) < 2 {
		t.Fatalf("expected tool calls, got %d", len(detail.ToolCalls))
	}
	if len(detail.Artifacts) == 0 {
		t.Fatalf("expected artifacts to be generated")
	}
	foundTitleOptions := false
	for _, item := range detail.Artifacts {
		if item.Kind == "title_options" {
			foundTitleOptions = true
			break
		}
	}
	if !foundTitleOptions {
		t.Fatalf("expected title_options artifact, got %+v", detail.Artifacts)
	}
}
