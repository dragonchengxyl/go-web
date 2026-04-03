package usecase

import (
	"context"
	"fmt"
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

type DecideAgentApprovalInput struct {
	ApprovalID string `json:"approval_id"`
	Decision   string `json:"decision"`
}

type DecideAgentApprovalOutput struct {
	Detail       *agentdomain.RunDetail `json:"detail"`
	ApplyPayload map[string]any         `json:"apply_payload,omitempty"`
}

type AgentService struct {
	repo        agentdomain.Repository
	postService *PostService
	async       bool
	hub         *agentRunHub
	tools       *agentToolRegistry
	maxAttempts int
}

func NewAgentService(repo agentdomain.Repository, postService *PostService) *AgentService {
	return newAgentService(repo, postService, true)
}

func newAgentService(repo agentdomain.Repository, postService *PostService, async bool) *AgentService {
	service := &AgentService{
		repo:        repo,
		postService: postService,
		async:       async,
		hub:         newAgentRunHub(),
		maxAttempts: 3,
	}
	service.tools = newAgentToolRegistry(service)
	return service
}

func (s *AgentService) Enabled() bool {
	return s != nil && s.repo != nil
}

func (s *AgentService) SubscribeRun(runID uuid.UUID) (chan AgentRunEvent, func()) {
	if s == nil || s.hub == nil {
		ch := make(chan AgentRunEvent)
		close(ch)
		return ch, func() {}
	}
	ch := s.hub.Subscribe(runID)
	return ch, func() {
		s.hub.Unsubscribe(runID, ch)
	}
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
		LatestSummary:   "任务已创建，等待执行器处理。",
		AttemptCount:    0,
		MaxAttempts:     s.maxAttempts,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建 Agent 任务失败", err)
	}

	queuedStep := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       run.ID,
		StepIndex:   1,
		Kind:        "system",
		Title:       "任务已创建",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     "任务已经入队，等待执行。",
		CreatedAt:   now,
		StartedAt:   &now,
		CompletedAt: &now,
	}
	if err := s.repo.CreateStep(ctx, queuedStep); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "记录 Agent 初始步骤失败", err)
	}
	if s.async {
		s.publishRunEvent(run.ID, "queued", queuedStep.Summary)
		return s.GetRunDetail(ctx, userID, run.ID)
	}

	s.publishRunEvent(run.ID, "queued", queuedStep.Summary)
	s.executeRunSafe(context.Background(), *run)
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
	s.publishRunEvent(runID, "cancelled", step.Summary)

	return s.GetRunDetail(ctx, userID, runID)
}

func (s *AgentService) DecideApproval(ctx context.Context, userID, runID uuid.UUID, input DecideAgentApprovalInput) (*DecideAgentApprovalOutput, error) {
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

	approvalID, err := uuid.Parse(strings.TrimSpace(input.ApprovalID))
	if err != nil {
		return nil, apperr.BadRequest("无效的审批ID")
	}
	decision := strings.TrimSpace(strings.ToLower(input.Decision))
	if decision != "approved" && decision != "rejected" {
		return nil, apperr.BadRequest("无效的审批决定")
	}

	var target *agentdomain.Approval
	for _, item := range detail.Approvals {
		if item != nil && item.ID == approvalID {
			target = item
			break
		}
	}
	if target == nil {
		return nil, apperr.ErrNotFound
	}
	if target.Status != agentdomain.ApprovalStatusPending {
		return &DecideAgentApprovalOutput{
			Detail:       detail,
			ApplyPayload: cloneAgentMap(target.Payload),
		}, nil
	}

	now := time.Now()
	nextStatus := agentdomain.ApprovalStatusApproved
	runSummary := "已经批准把 Agent 产物回填到发帖页草稿。"
	stepTitle := "已批准回填草稿"
	stepSummary := "你已经批准把当前这轮 Agent 生成的草稿建议带回发布页。"
	var applyPayload map[string]any
	if decision == "rejected" {
		nextStatus = agentdomain.ApprovalStatusRejected
		runSummary = "你保留了这轮分析结果，但没有把建议回填到草稿。"
		stepTitle = "已拒绝回填草稿"
		stepSummary = "这次运行仍然保留结构化建议，但不会自动带回发布页。"
	} else {
		applyPayload = cloneAgentMap(target.Payload)
	}

	if err := s.repo.UpdateApprovalStatus(ctx, target.ID, nextStatus, &userID, &now); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新审批状态失败", err)
	}
	if err := s.repo.UpdateRunStatus(ctx, runID, agentdomain.RunStatusCompleted, runSummary, "", &now); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新 Agent 任务状态失败", err)
	}
	step := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       runID,
		StepIndex:   len(detail.Steps) + 1,
		Kind:        "approval",
		Title:       stepTitle,
		Status:      agentdomain.StepStatusCompleted,
		Summary:     stepSummary,
		OutputData:  map[string]any{"decision": decision},
		CreatedAt:   now,
		StartedAt:   &now,
		CompletedAt: &now,
	}
	if err := s.repo.CreateStep(ctx, step); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "记录审批步骤失败", err)
	}
	s.publishRunEvent(runID, "approval_decided", step.Summary)

	nextDetail, err := s.GetRunDetail(ctx, userID, runID)
	if err != nil {
		return nil, err
	}
	return &DecideAgentApprovalOutput{
		Detail:       nextDetail,
		ApplyPayload: applyPayload,
	}, nil
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

func (s *AgentService) publishRunEvent(runID uuid.UUID, eventType, summary string) {
	if s == nil || s.hub == nil {
		return
	}
	s.hub.Publish(runID, AgentRunEvent{
		RunID:   runID.String(),
		Type:    eventType,
		Summary: summary,
	})
}

func (s *AgentService) executeRunAsync(run agentdomain.Run) {
	s.executeRunSafe(context.Background(), run)
}

func (s *AgentService) executeRunSafe(ctx context.Context, run agentdomain.Run) {
	startedAt := time.Now()
	started, err := s.repo.TryStartRun(ctx, run.ID, startedAt)
	if err != nil {
		s.recordRunFailure(ctx, run.ID, err)
		return
	}
	if !started {
		return
	}

	currentRun, err := s.repo.GetRunByID(ctx, run.ID)
	if err != nil {
		s.recordRunFailure(ctx, run.ID, err)
		return
	}
	if err := s.executeRun(ctx, currentRun); err != nil {
		s.recordRunFailure(ctx, run.ID, err)
	}
}

func (s *AgentService) executeRun(ctx context.Context, run *agentdomain.Run) error {
	if run == nil {
		return fmt.Errorf("agent run is nil")
	}

	switch run.Scenario {
	case "", agentdomain.ScenarioPostAgent:
		return s.executePostAgent(ctx, run)
	default:
		return fmt.Errorf("unsupported agent scenario: %s", run.Scenario)
	}
}

func (s *AgentService) executePostAgent(ctx context.Context, run *agentdomain.Run) error {
	if stopped, err := s.runStopped(ctx, run.ID); err != nil {
		return err
	} else if stopped {
		return nil
	}

	if err := s.repo.UpdateRunStatus(ctx, run.ID, agentdomain.RunStatusRunning, "正在分析当前草稿和页面上下文。", "", nil); err != nil {
		return err
	}
	s.publishRunEvent(run.ID, "running", "正在分析当前草稿和页面上下文。")

	state := &agentToolState{Run: run}
	pageContext, contextSummary := buildAgentPostPageContext(run.ContextSnapshot)
	state.PageContext = pageContext
	state.ContextSummary = contextSummary
	planStepIndex, err := s.nextRunStepIndex(ctx, run.ID)
	if err != nil {
		return err
	}
	planTime := time.Now()
	planStep := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       run.ID,
		StepIndex:   planStepIndex,
		Kind:        "plan",
		Title:       "拆解发帖任务",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     "先读取草稿与页面上下文，再生成标题、正文润色、标签和可见性建议。",
		InputData:   cloneAgentMap(run.ContextSnapshot),
		OutputData:  map[string]any{"scenario": run.Scenario, "context_summary": contextSummary},
		CreatedAt:   planTime,
		StartedAt:   &planTime,
		CompletedAt: &planTime,
	}
	if err := s.repo.CreateStep(ctx, planStep); err != nil {
		return err
	}
	s.publishRunEvent(run.ID, "step_completed", planStep.Summary)
	if stopped, err := s.runStopped(ctx, run.ID); err != nil {
		return err
	} else if stopped {
		return nil
	}

	if _, err := s.executeTool(ctx, state, run.ID, &planStep.ID, "get_post_create_context"); err != nil {
		return err
	}
	if _, err := s.executeTool(ctx, state, run.ID, &planStep.ID, "generate_post_copilot_artifacts"); err != nil {
		return err
	}
	if stopped, err := s.runStopped(ctx, run.ID); err != nil {
		return err
	} else if stopped {
		return nil
	}

	artifactStepIndex, err := s.nextRunStepIndex(ctx, run.ID)
	if err != nil {
		return err
	}
	artifactStepTime := time.Now()
	artifactStep := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       run.ID,
		StepIndex:   artifactStepIndex,
		Kind:        "artifact_generation",
		Title:       "生成发帖建议产物",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     firstNonEmpty(buildCopilotFallbackText(state.Insights), "已经根据当前草稿生成建议。"),
		OutputData:  map[string]any{"insight_count": len(state.Insights)},
		CreatedAt:   artifactStepTime,
		StartedAt:   &artifactStepTime,
		CompletedAt: &artifactStepTime,
	}
	if err := s.repo.CreateStep(ctx, artifactStep); err != nil {
		return err
	}
	s.publishRunEvent(run.ID, "step_completed", artifactStep.Summary)

	for _, insight := range state.Insights {
		if insight.Title == "" && insight.Summary == "" && len(insight.Bullets) == 0 {
			continue
		}
		artifact := &agentdomain.Artifact{
			ID:     uuid.New(),
			RunID:  run.ID,
			StepID: &artifactStep.ID,
			Kind:   insight.Kind,
			Title:  insight.Title,
			Content: strings.TrimSpace(strings.Join([]string{
				insight.Summary,
				formatInsightBullets(insight.Bullets),
			}, "\n")),
			StructuredData: map[string]any{
				"summary": insight.Summary,
				"bullets": insight.Bullets,
			},
			CreatedAt: artifactStepTime,
		}
		if err := s.repo.CreateArtifact(ctx, artifact); err != nil {
			return err
		}
	}

	if _, err := s.executeTool(ctx, state, run.ID, &artifactStep.ID, "build_post_draft_patch"); err != nil {
		return err
	}
	draftPatch := state.DraftPatch
	if len(draftPatch) > 0 {
		artifact := &agentdomain.Artifact{
			ID:             uuid.New(),
			RunID:          run.ID,
			StepID:         &artifactStep.ID,
			Kind:           "draft_patch",
			Title:          "可回填草稿",
			Content:        "这份产物可在批准后直接回填到发布页草稿。",
			StructuredData: draftPatch,
			CreatedAt:      artifactStepTime,
		}
		if err := s.repo.CreateArtifact(ctx, artifact); err != nil {
			return err
		}
	}
	s.publishRunEvent(run.ID, "artifact_ready", "结构化产物已经生成。")

	approvalTime := time.Now()
	approval := &agentdomain.Approval{
		ID:         uuid.New(),
		RunID:      run.ID,
		StepID:     &artifactStep.ID,
		ActionType: "apply_post_draft",
		Title:      "把建议回填到发帖页草稿",
		Status:     agentdomain.ApprovalStatusPending,
		Payload:    draftPatch,
		CreatedAt:  approvalTime,
	}
	if err := s.repo.CreateApproval(ctx, approval); err != nil {
		return err
	}
	if stopped, err := s.runStopped(ctx, run.ID); err != nil {
		return err
	} else if stopped {
		return nil
	}

	runSummary := "建议已生成，等待你批准是否把结果回填到发帖页。"
	if err := s.repo.UpdateRunStatus(ctx, run.ID, agentdomain.RunStatusWaitingApproval, runSummary, "", nil); err != nil {
		return err
	}
	s.publishRunEvent(run.ID, "waiting_approval", runSummary)
	finalStepIndex, err := s.nextRunStepIndex(ctx, run.ID)
	if err != nil {
		return err
	}
	finalStep := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       run.ID,
		StepIndex:   finalStepIndex,
		Kind:        "approval",
		Title:       "等待批准回填草稿",
		Status:      agentdomain.StepStatusCompleted,
		Summary:     runSummary,
		OutputData:  map[string]any{"approval_action": "apply_post_draft"},
		CreatedAt:   approvalTime,
		StartedAt:   &approvalTime,
		CompletedAt: &approvalTime,
	}
	if err := s.repo.CreateStep(ctx, finalStep); err != nil {
		return err
	}
	s.publishRunEvent(run.ID, "step_completed", finalStep.Summary)
	return nil
}

func (s *AgentService) recordRunFailure(ctx context.Context, runID uuid.UUID, err error) {
	current, getErr := s.repo.GetRunByID(ctx, runID)
	if getErr == nil && current != nil && current.Status == agentdomain.RunStatusCancelled {
		return
	}

	failAt := time.Now()
	if current != nil {
		s.handleRunExecutionFailure(ctx, current, err)
		return
	}
	_ = s.repo.UpdateRunStatus(ctx, runID, agentdomain.RunStatusFailed, "", err.Error(), &failAt)
	stepIndex, stepErr := s.nextRunStepIndex(ctx, runID)
	if stepErr != nil {
		stepIndex = 9999
	}
	failStep := &agentdomain.Step{
		ID:          uuid.New(),
		RunID:       runID,
		StepIndex:   stepIndex,
		Kind:        "system",
		Title:       "任务执行失败",
		Status:      agentdomain.StepStatusFailed,
		Summary:     "当前这次 Agent 运行没有顺利完成。",
		ErrorText:   err.Error(),
		CreatedAt:   failAt,
		StartedAt:   &failAt,
		CompletedAt: &failAt,
	}
	_ = s.repo.CreateStep(ctx, failStep)
	s.publishRunEvent(runID, "failed", failStep.Summary)
}

func (s *AgentService) handleRunExecutionFailure(ctx context.Context, run *agentdomain.Run, err error) {
	if run == nil {
		return
	}
	failAt := time.Now()
	run.UpdatedAt = failAt
	run.LastError = err.Error()
	run.LastErrorAt = &failAt
	run.LatestSummary = "当前这次 Agent 运行没有顺利完成。"

	if run.AttemptCount >= run.MaxAttempts {
		run.Status = agentdomain.RunStatusFailed
		run.CompletedAt = &failAt
		run.NextRetryAt = nil
	} else {
		run.Status = agentdomain.RunStatusQueued
		run.CompletedAt = nil
		nextRetryAt := failAt.Add(s.retryDelay(run.AttemptCount))
		run.NextRetryAt = &nextRetryAt
		run.LatestSummary = "任务执行失败，稍后会自动重试。"
	}
	_ = s.repo.UpdateRun(ctx, run)

	stepIndex, stepErr := s.nextRunStepIndex(ctx, run.ID)
	if stepErr != nil {
		stepIndex = 9999
	}
	failStep := &agentdomain.Step{
		ID:        uuid.New(),
		RunID:     run.ID,
		StepIndex: stepIndex,
		Kind:      "system",
		Status:    agentdomain.StepStatusFailed,
		ErrorText: err.Error(),
		CreatedAt: failAt,
		StartedAt: &failAt,
	}
	if run.Status == agentdomain.RunStatusQueued {
		failStep.Title = "任务执行失败，等待重试"
		failStep.Summary = "当前这次 Agent 运行失败，系统会稍后自动重试。"
	} else {
		failStep.Title = "任务执行失败"
		failStep.Summary = "当前这次 Agent 运行没有顺利完成。"
		failStep.CompletedAt = &failAt
	}
	_ = s.repo.CreateStep(ctx, failStep)

	if run.Status == agentdomain.RunStatusQueued {
		s.publishRunEvent(run.ID, "retry_scheduled", failStep.Summary)
		return
	}
	s.publishRunEvent(run.ID, "failed", failStep.Summary)
}

func (s *AgentService) retryDelay(attemptCount int) time.Duration {
	if attemptCount < 1 {
		attemptCount = 1
	}
	return time.Duration(attemptCount*attemptCount) * time.Second
}

func (s *AgentService) nextRunStepIndex(ctx context.Context, runID uuid.UUID) (int, error) {
	steps, err := s.repo.ListSteps(ctx, runID)
	if err != nil {
		return 0, err
	}
	return len(steps) + 1, nil
}

func (s *AgentService) runStopped(ctx context.Context, runID uuid.UUID) (bool, error) {
	item, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return false, err
	}
	switch item.Status {
	case agentdomain.RunStatusCancelled, agentdomain.RunStatusCompleted, agentdomain.RunStatusFailed:
		return true, nil
	default:
		return false, nil
	}
}

func buildAgentPostPageContext(snapshot map[string]any) (*AssistantPageContext, string) {
	fields := make(map[string]string)
	if len(snapshot) > 0 {
		for key, value := range snapshot {
			trimmed := strings.TrimSpace(fmt.Sprint(value))
			if trimmed == "" || trimmed == "<nil>" {
				continue
			}
			fields[key] = trimmed
		}
	}

	title := "发布动态"
	if groupName := strings.TrimSpace(fields["group_name"]); groupName != "" {
		title = "发布到圈子：" + groupName
	}

	pageContext := &AssistantPageContext{
		Path:    firstNonEmpty(fields["source_path"], "/posts/create"),
		Kind:    "post_create",
		Title:   title,
		Summary: "当前任务来自发帖页，适合输出标题、正文润色、标签和可见性建议。",
		Fields:  fields,
	}

	var parts []string
	if value := strings.TrimSpace(fields["draft_title"]); value != "" {
		parts = append(parts, "标题已填写")
	}
	if value := strings.TrimSpace(fields["draft_content"]); value != "" {
		parts = append(parts, fmt.Sprintf("正文约 %d 字", utf8.RuneCountInString(value)))
	}
	if value := strings.TrimSpace(fields["draft_tags"]); value != "" {
		parts = append(parts, "已有标签草稿")
	}
	if value := strings.TrimSpace(fields["group_name"]); value != "" {
		parts = append(parts, "目标圈子："+value)
	}
	if value := strings.TrimSpace(fields["visibility"]); value != "" {
		parts = append(parts, "当前可见性："+value)
	}
	if len(parts) == 0 {
		parts = append(parts, "未提供足够草稿信息，先按通用发帖场景处理")
	}

	return pageContext, strings.Join(parts, "；")
}

func formatInsightBullets(items []string) string {
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		lines = append(lines, "- "+trimmed)
	}
	return strings.Join(lines, "\n")
}

func buildPostDraftPatch(pageContext *AssistantPageContext, insights []AssistantInsight) map[string]any {
	if pageContext == nil {
		return nil
	}

	fields := pageContext.Fields
	if fields == nil {
		fields = map[string]string{}
	}

	patch := make(map[string]any)
	title := strings.TrimSpace(fields["draft_title"])
	content := strings.TrimSpace(fields["draft_content"])
	groupName := strings.TrimSpace(fields["group_name"])
	visibilityValue := normalizeDraftVisibility(fields["visibility"])

	for _, insight := range insights {
		switch insight.Kind {
		case "title_options":
			if title == "" && len(insight.Bullets) > 0 {
				title = strings.TrimSpace(insight.Bullets[0])
			}
		case "tag_suggestions":
			if len(insight.Bullets) > 0 {
				tags := make([]string, 0, len(insight.Bullets))
				for _, item := range insight.Bullets {
					trimmed := strings.TrimSpace(strings.TrimPrefix(item, "#"))
					if trimmed == "" {
						continue
					}
					tags = append(tags, trimmed)
				}
				if len(tags) > 0 {
					patch["tags"] = strings.Join(tags, ", ")
				}
			}
		case "visibility_suggestion":
			if normalized := inferVisibilityFromInsight(insight); normalized != "" {
				visibilityValue = normalized
			}
		}
	}

	if title != "" {
		patch["title"] = title
	}
	if content != "" {
		patch["content"] = buildSuggestedDraftContent(content, groupName)
	}
	if visibilityValue != "" {
		patch["visibility"] = visibilityValue
	}
	if path := strings.TrimSpace(pageContext.Path); path != "" {
		patch["source_path"] = path
	}
	if groupName != "" {
		patch["group_name"] = groupName
	}
	return patch
}

func buildSuggestedDraftContent(content, groupName string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	if !strings.HasSuffix(content, "。") && !strings.HasSuffix(content, "！") && !strings.HasSuffix(content, "？") {
		content += "。"
	}
	if groupName != "" && !strings.Contains(content, groupName) {
		content += "\n\n想把这次调整带到 " + groupName + " 这个语境里，也想听听大家对这版方向的看法。"
	}
	return content
}

func normalizeDraftVisibility(value string) string {
	switch strings.TrimSpace(value) {
	case "仅关注者可见", "followers_only":
		return "followers_only"
	case "私密", "private":
		return "private"
	case "公开", "public":
		return "public"
	default:
		return "public"
	}
}

func inferVisibilityFromInsight(insight AssistantInsight) string {
	text := strings.TrimSpace(insight.Summary + " " + strings.Join(insight.Bullets, " "))
	switch {
	case strings.Contains(text, "仅关注者可见"):
		return "followers_only"
	case strings.Contains(text, "私密"):
		return "private"
	case strings.Contains(text, "公开"):
		return "public"
	default:
		return ""
	}
}
