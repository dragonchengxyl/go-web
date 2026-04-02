package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	agentdomain "github.com/studio/platform/internal/domain/agent"
)

type agentToolDefinition struct {
	Name        string
	AccessLevel agentdomain.ToolAccessLevel
	Execute     func(context.Context, *agentToolState) (*agentToolExecutionResult, error)
}

type agentToolExecutionResult struct {
	OutputData map[string]any
	Summary    string
}

type agentToolRegistry struct {
	tools map[string]agentToolDefinition
}

type agentToolState struct {
	Run            *agentdomain.Run
	PageContext    *AssistantPageContext
	ContextSummary string
	Insights       []AssistantInsight
	DraftPatch     map[string]any
}

func newAgentToolRegistry(s *AgentService) *agentToolRegistry {
	registry := &agentToolRegistry{
		tools: make(map[string]agentToolDefinition),
	}

	registry.register(agentToolDefinition{
		Name:        "get_post_create_context",
		AccessLevel: agentdomain.ToolAccessReadOnly,
		Execute: func(_ context.Context, state *agentToolState) (*agentToolExecutionResult, error) {
			if state == nil || state.Run == nil {
				return nil, fmt.Errorf("agent tool state is not ready")
			}
			pageContext, summary := buildAgentPostPageContext(state.Run.ContextSnapshot)
			state.PageContext = pageContext
			state.ContextSummary = summary
			return &agentToolExecutionResult{
				Summary: "已读取发帖页上下文。",
				OutputData: map[string]any{
					"context_summary": summary,
					"page_kind":       pageContext.Kind,
					"path":            pageContext.Path,
				},
			}, nil
		},
	})

	registry.register(agentToolDefinition{
		Name:        "generate_post_copilot_artifacts",
		AccessLevel: agentdomain.ToolAccessReadOnly,
		Execute: func(ctx context.Context, state *agentToolState) (*agentToolExecutionResult, error) {
			if state == nil || state.Run == nil {
				return nil, fmt.Errorf("agent tool state is not ready")
			}
			if state.PageContext == nil {
				pageContext, summary := buildAgentPostPageContext(state.Run.ContextSnapshot)
				state.PageContext = pageContext
				state.ContextSummary = summary
			}
			assistantHelper := &AssistantService{postService: s.postService}
			insights := assistantHelper.buildPostCreateInsights(ctx, state.Run.Goal, state.PageContext)
			state.Insights = insights
			return &agentToolExecutionResult{
				Summary: "已生成发帖建议草稿。",
				OutputData: map[string]any{
					"insight_count": len(insights),
					"insight_kinds": collectInsightKinds(insights),
				},
			}, nil
		},
	})

	registry.register(agentToolDefinition{
		Name:        "build_post_draft_patch",
		AccessLevel: agentdomain.ToolAccessReadOnly,
		Execute: func(_ context.Context, state *agentToolState) (*agentToolExecutionResult, error) {
			if state == nil || state.Run == nil {
				return nil, fmt.Errorf("agent tool state is not ready")
			}
			if state.PageContext == nil {
				pageContext, summary := buildAgentPostPageContext(state.Run.ContextSnapshot)
				state.PageContext = pageContext
				state.ContextSummary = summary
			}
			patch := buildPostDraftPatch(state.PageContext, state.Insights)
			state.DraftPatch = patch
			return &agentToolExecutionResult{
				Summary: "已生成可回填的草稿补丁。",
				OutputData: map[string]any{
					"patch_keys": sortedMapKeys(patch),
					"has_patch":  len(patch) > 0,
				},
			}, nil
		},
	})

	return registry
}

func (r *agentToolRegistry) register(def agentToolDefinition) {
	if r == nil {
		return
	}
	r.tools[def.Name] = def
}

func (r *agentToolRegistry) get(name string) (agentToolDefinition, bool) {
	if r == nil {
		return agentToolDefinition{}, false
	}
	def, ok := r.tools[name]
	return def, ok
}

func collectInsightKinds(items []AssistantInsight) []string {
	if len(items) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(items))
	for _, item := range items {
		if item.Kind == "" {
			continue
		}
		kinds = append(kinds, item.Kind)
	}
	return kinds
}

func sortedMapKeys(items map[string]any) []string {
	if len(items) == 0 {
		return nil
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *AgentService) executeTool(
	ctx context.Context,
	state *agentToolState,
	runID uuid.UUID,
	stepID *uuid.UUID,
	toolName string,
) (*agentToolExecutionResult, error) {
	if s == nil || s.tools == nil {
		return nil, fmt.Errorf("agent tool registry is not initialized")
	}
	def, ok := s.tools.get(toolName)
	if !ok {
		return nil, fmt.Errorf("unknown agent tool: %s", toolName)
	}

	startedAt := time.Now()
	result, err := def.Execute(ctx, state)
	if err != nil {
		failedAt := time.Now()
		_ = s.repo.CreateToolCall(ctx, &agentdomain.ToolCall{
			ID:          uuid.New(),
			RunID:       runID,
			StepID:      stepID,
			ToolName:    def.Name,
			AccessLevel: def.AccessLevel,
			Status:      agentdomain.StepStatusFailed,
			ErrorText:   err.Error(),
			CreatedAt:   startedAt,
			StartedAt:   &startedAt,
			CompletedAt: &failedAt,
		})
		s.publishRunEvent(runID, "tool_failed", fmt.Sprintf("工具 %s 执行失败。", def.Name))
		return nil, err
	}

	completedAt := time.Now()
	call := &agentdomain.ToolCall{
		ID:          uuid.New(),
		RunID:       runID,
		StepID:      stepID,
		ToolName:    def.Name,
		AccessLevel: def.AccessLevel,
		Status:      agentdomain.StepStatusCompleted,
		OutputData:  cloneAgentMap(result.OutputData),
		CreatedAt:   startedAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}
	if err := s.repo.CreateToolCall(ctx, call); err != nil {
		return nil, err
	}
	s.publishRunEvent(runID, "tool_completed", result.Summary)
	return result, nil
}
