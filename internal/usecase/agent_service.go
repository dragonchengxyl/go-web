package usecase

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	agentdomain "github.com/studio/platform/internal/domain/agent"
	"github.com/studio/platform/internal/pkg/apperr"
)

type CreateAgentRunInput struct {
	Title           string         `json:"title"`
	Goal            string         `json:"goal"`
	Scenario        string         `json:"scenario"`
	ContextSnapshot map[string]any `json:"context_snapshot,omitempty"`
}

type ListAgentRunsOutput struct {
	Runs  []*agentdomain.Run `json:"runs"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type AgentService struct {
	repo agentdomain.Repository
}

func NewAgentService(repo agentdomain.Repository) *AgentService {
	return &AgentService{repo: repo}
}

func (s *AgentService) Enabled() bool {
	return s != nil && s.repo != nil
}

func (s *AgentService) CreateRun(ctx context.Context, userID uuid.UUID, input CreateAgentRunInput) (*agentdomain.RunDetail, error) {
	if !s.Enabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	if userID == uuid.Nil {
		return nil, apperr.ErrUnauthorized
	}

	goal := strings.TrimSpace(input.Goal)
	if goal == "" {
		return nil, apperr.BadRequest("请输入这次希望 Agent 完成的目标")
	}

	scenario := strings.TrimSpace(input.Scenario)
	if scenario == "" {
		scenario = agentdomain.ScenarioPostAgent
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = truncateAgentText(goal, 40)
	}
	now := time.Now()
	run := &agentdomain.Run{
		ID:              uuid.New(),
		UserID:          userID,
		Title:           truncateAgentText(title, 80),
		Goal:            truncateAgentText(goal, 1200),
		Scenario:        truncateAgentText(scenario, 64),
		Status:          agentdomain.RunStatusQueued,
		ContextSnapshot: cloneAgentMap(input.ContextSnapshot),
		LatestSummary:   "任务已创建，等待执行器接入。",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建 Agent 任务失败", err)
	}

	step := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       run.ID,
		StepIndex:   1,
		Kind:        "system",
		Title:       "任务已创建",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     "当前版本先完成运行骨架和时间线持久化，后续会接入真正的规划与工具执行。",
		CreatedAt:   now,
		StartedAt:   &now,
		CompletedAt: &now,
	}
	if err := s.repo.CreateStep(ctx, step); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "记录 Agent 初始步骤失败", err)
	}

	return s.GetRunDetail(ctx, userID, run.ID)
}

func (s *AgentService) ListRuns(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ListAgentRunsOutput, error) {
	if !s.Enabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	if userID == uuid.Nil {
		return nil, apperr.ErrUnauthorized
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	items, total, err := s.repo.ListRuns(ctx, userID, page, pageSize)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 任务列表失败", err)
	}
	return &ListAgentRunsOutput{
		Runs:  items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (s *AgentService) GetRunDetail(ctx context.Context, userID, runID uuid.UUID) (*agentdomain.RunDetail, error) {
	if !s.Enabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	if userID == uuid.Nil {
		return nil, apperr.ErrUnauthorized
	}

	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		if err == agentdomain.ErrRunNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 任务失败", err)
	}
	if run.UserID != userID {
		return nil, apperr.ErrForbidden
	}

	steps, err := s.repo.ListSteps(ctx, runID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 步骤失败", err)
	}
	toolCalls, err := s.repo.ListToolCalls(ctx, runID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 工具调用失败", err)
	}
	approvals, err := s.repo.ListApprovals(ctx, runID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 审批记录失败", err)
	}
	artifacts, err := s.repo.ListArtifacts(ctx, runID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "读取 Agent 产物失败", err)
	}

	return &agentdomain.RunDetail{
		Run:       run,
		Steps:     steps,
		ToolCalls: toolCalls,
		Approvals: approvals,
		Artifacts: artifacts,
	}, nil
}

func (s *AgentService) CancelRun(ctx context.Context, userID, runID uuid.UUID) (*agentdomain.RunDetail, error) {
	if !s.Enabled() {
		return nil, apperr.Wrap(apperr.CodeInternalError, "Agent 服务未初始化", nil)
	}
	if userID == uuid.Nil {
		return nil, apperr.ErrUnauthorized
	}

	detail, err := s.GetRunDetail(ctx, userID, runID)
	if err != nil {
		return nil, err
	}

	switch detail.Run.Status {
	case agentdomain.RunStatusCompleted, agentdomain.RunStatusFailed, agentdomain.RunStatusCancelled:
		return detail, nil
	}

	now := time.Now()
	if err := s.repo.UpdateRunStatus(ctx, runID, agentdomain.RunStatusCancelled, "任务已取消。", "", &now); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "取消 Agent 任务失败", err)
	}
	step := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       runID,
		StepIndex:   len(detail.Steps) + 1,
		Kind:        "system",
		Title:       "任务已取消",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     "用户主动取消了本次 Agent 任务。",
		CreatedAt:   now,
		StartedAt:   &now,
		CompletedAt: &now,
	}
	if err := s.repo.CreateStep(ctx, step); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "记录取消步骤失败", err)
	}

	return s.GetRunDetail(ctx, userID, runID)
}

func truncateAgentText(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func cloneAgentMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}
