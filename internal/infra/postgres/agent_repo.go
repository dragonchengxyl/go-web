package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	agentdomain "github.com/studio/platform/internal/domain/agent"
)

type AgentRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRepository(pool *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

func (r *AgentRepository) CreateRun(ctx context.Context, item *agentdomain.Run) error {
	contextJSON, err := marshalAgentJSON(item.ContextSnapshot)
	if err != nil {
		return fmt.Errorf("marshal agent run context: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO agent_runs (
			id, user_id, title, goal, scenario, status, context_snapshot,
			latest_summary, last_error, created_at, updated_at, started_at, completed_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13
		)
	`,
		item.ID,
		item.UserID,
		item.Title,
		item.Goal,
		item.Scenario,
		string(item.Status),
		contextJSON,
		item.LatestSummary,
		item.LastError,
		item.CreatedAt,
		item.UpdatedAt,
		item.StartedAt,
		item.CompletedAt,
	)
	return err
}

func (r *AgentRepository) GetRunByID(ctx context.Context, id uuid.UUID) (*agentdomain.Run, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, title, goal, scenario, status, context_snapshot,
		       latest_summary, last_error, created_at, updated_at, started_at, completed_at
		FROM agent_runs
		WHERE id = $1
	`, id)
	item, err := scanAgentRun(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, agentdomain.ErrRunNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *AgentRepository) ListRuns(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*agentdomain.Run, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM agent_runs WHERE user_id = $1
	`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, goal, scenario, status, context_snapshot,
		       latest_summary, last_error, created_at, updated_at, started_at, completed_at
		FROM agent_runs
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*agentdomain.Run, 0, pageSize)
	for rows.Next() {
		item, err := scanAgentRun(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *AgentRepository) UpdateRunStatus(ctx context.Context, id uuid.UUID, status agentdomain.RunStatus, summary, lastError string, completedAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE agent_runs
		SET status = $2,
		    latest_summary = $3,
		    last_error = $4,
		    updated_at = $5,
		    completed_at = $6
		WHERE id = $1
	`, id, string(status), summary, lastError, time.Now(), completedAt)
	return err
}

func (r *AgentRepository) CreateStep(ctx context.Context, item *agentdomain.Step) error {
	inputJSON, err := marshalAgentJSON(item.InputData)
	if err != nil {
		return fmt.Errorf("marshal step input: %w", err)
	}
	outputJSON, err := marshalAgentJSON(item.OutputData)
	if err != nil {
		return fmt.Errorf("marshal step output: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO agent_steps (
			id, run_id, step_index, kind, title, status, summary,
			input_data, output_data, error_text, created_at, started_at, completed_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13
		)
	`,
		item.ID,
		item.RunID,
		item.StepIndex,
		item.Kind,
		item.Title,
		string(item.Status),
		item.Summary,
		inputJSON,
		outputJSON,
		item.ErrorText,
		item.CreatedAt,
		item.StartedAt,
		item.CompletedAt,
	)
	return err
}

func (r *AgentRepository) ListSteps(ctx context.Context, runID uuid.UUID) ([]*agentdomain.Step, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, run_id, step_index, kind, title, status, summary,
		       input_data, output_data, error_text, created_at, started_at, completed_at
		FROM agent_steps
		WHERE run_id = $1
		ORDER BY step_index ASC, created_at ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*agentdomain.Step, 0)
	for rows.Next() {
		item, err := scanAgentStep(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentRepository) CreateToolCall(ctx context.Context, item *agentdomain.ToolCall) error {
	inputJSON, err := marshalAgentJSON(item.InputData)
	if err != nil {
		return fmt.Errorf("marshal tool call input: %w", err)
	}
	outputJSON, err := marshalAgentJSON(item.OutputData)
	if err != nil {
		return fmt.Errorf("marshal tool call output: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO agent_tool_calls (
			id, run_id, step_id, tool_name, access_level, status, input_data,
			output_data, error_text, created_at, started_at, completed_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12
		)
	`,
		item.ID,
		item.RunID,
		item.StepID,
		item.ToolName,
		string(item.AccessLevel),
		string(item.Status),
		inputJSON,
		outputJSON,
		item.ErrorText,
		item.CreatedAt,
		item.StartedAt,
		item.CompletedAt,
	)
	return err
}

func (r *AgentRepository) ListToolCalls(ctx context.Context, runID uuid.UUID) ([]*agentdomain.ToolCall, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, run_id, step_id, tool_name, access_level, status, input_data,
		       output_data, error_text, created_at, started_at, completed_at
		FROM agent_tool_calls
		WHERE run_id = $1
		ORDER BY created_at ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*agentdomain.ToolCall, 0)
	for rows.Next() {
		item, err := scanAgentToolCall(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentRepository) CreateApproval(ctx context.Context, item *agentdomain.Approval) error {
	payloadJSON, err := marshalAgentJSON(item.Payload)
	if err != nil {
		return fmt.Errorf("marshal approval payload: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO agent_approvals (
			id, run_id, step_id, action_type, title, status, payload,
			approved_by, approved_at, created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10
		)
	`,
		item.ID,
		item.RunID,
		item.StepID,
		item.ActionType,
		item.Title,
		string(item.Status),
		payloadJSON,
		item.ApprovedBy,
		item.ApprovedAt,
		item.CreatedAt,
	)
	return err
}

func (r *AgentRepository) GetApprovalByID(ctx context.Context, id uuid.UUID) (*agentdomain.Approval, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, run_id, step_id, action_type, title, status, payload,
		       approved_by, approved_at, created_at
		FROM agent_approvals
		WHERE id = $1
	`, id)
	item, err := scanAgentApproval(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, agentdomain.ErrRunNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *AgentRepository) ListApprovals(ctx context.Context, runID uuid.UUID) ([]*agentdomain.Approval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, run_id, step_id, action_type, title, status, payload,
		       approved_by, approved_at, created_at
		FROM agent_approvals
		WHERE run_id = $1
		ORDER BY created_at ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*agentdomain.Approval, 0)
	for rows.Next() {
		item, err := scanAgentApproval(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentRepository) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, status agentdomain.ApprovalStatus, approvedBy *uuid.UUID, approvedAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE agent_approvals
		SET status = $2,
		    approved_by = $3,
		    approved_at = $4
		WHERE id = $1
	`, id, string(status), approvedBy, approvedAt)
	return err
}

func (r *AgentRepository) CreateArtifact(ctx context.Context, item *agentdomain.Artifact) error {
	structuredJSON, err := marshalAgentJSON(item.StructuredData)
	if err != nil {
		return fmt.Errorf("marshal artifact data: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO agent_artifacts (
			id, run_id, step_id, kind, title, content, structured_data, created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`,
		item.ID,
		item.RunID,
		item.StepID,
		item.Kind,
		item.Title,
		item.Content,
		structuredJSON,
		item.CreatedAt,
	)
	return err
}

func (r *AgentRepository) ListArtifacts(ctx context.Context, runID uuid.UUID) ([]*agentdomain.Artifact, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, run_id, step_id, kind, title, content, structured_data, created_at
		FROM agent_artifacts
		WHERE run_id = $1
		ORDER BY created_at ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*agentdomain.Artifact, 0)
	for rows.Next() {
		item, err := scanAgentArtifact(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanAgentRun(row interface{ Scan(dest ...any) error }) (*agentdomain.Run, error) {
	var item agentdomain.Run
	var status string
	var contextJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Goal,
		&item.Scenario,
		&status,
		&contextJSON,
		&item.LatestSummary,
		&item.LastError,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.StartedAt,
		&item.CompletedAt,
	); err != nil {
		return nil, err
	}
	item.Status = agentdomain.RunStatus(status)
	item.ContextSnapshot = unmarshalAgentJSON(contextJSON)
	return &item, nil
}

func scanAgentStep(row interface{ Scan(dest ...any) error }) (*agentdomain.Step, error) {
	var item agentdomain.Step
	var status string
	var inputJSON []byte
	var outputJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.RunID,
		&item.StepIndex,
		&item.Kind,
		&item.Title,
		&status,
		&item.Summary,
		&inputJSON,
		&outputJSON,
		&item.ErrorText,
		&item.CreatedAt,
		&item.StartedAt,
		&item.CompletedAt,
	); err != nil {
		return nil, err
	}
	item.Status = agentdomain.StepStatus(status)
	item.InputData = unmarshalAgentJSON(inputJSON)
	item.OutputData = unmarshalAgentJSON(outputJSON)
	return &item, nil
}

func scanAgentToolCall(row interface{ Scan(dest ...any) error }) (*agentdomain.ToolCall, error) {
	var item agentdomain.ToolCall
	var accessLevel string
	var status string
	var inputJSON []byte
	var outputJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.RunID,
		&item.StepID,
		&item.ToolName,
		&accessLevel,
		&status,
		&inputJSON,
		&outputJSON,
		&item.ErrorText,
		&item.CreatedAt,
		&item.StartedAt,
		&item.CompletedAt,
	); err != nil {
		return nil, err
	}
	item.AccessLevel = agentdomain.ToolAccessLevel(accessLevel)
	item.Status = agentdomain.StepStatus(status)
	item.InputData = unmarshalAgentJSON(inputJSON)
	item.OutputData = unmarshalAgentJSON(outputJSON)
	return &item, nil
}

func scanAgentApproval(row interface{ Scan(dest ...any) error }) (*agentdomain.Approval, error) {
	var item agentdomain.Approval
	var status string
	var payloadJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.RunID,
		&item.StepID,
		&item.ActionType,
		&item.Title,
		&status,
		&payloadJSON,
		&item.ApprovedBy,
		&item.ApprovedAt,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.Status = agentdomain.ApprovalStatus(status)
	item.Payload = unmarshalAgentJSON(payloadJSON)
	return &item, nil
}

func scanAgentArtifact(row interface{ Scan(dest ...any) error }) (*agentdomain.Artifact, error) {
	var item agentdomain.Artifact
	var structuredJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.RunID,
		&item.StepID,
		&item.Kind,
		&item.Title,
		&item.Content,
		&structuredJSON,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.StructuredData = unmarshalAgentJSON(structuredJSON)
	return &item, nil
}

func marshalAgentJSON(value map[string]any) ([]byte, error) {
	if len(value) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(value)
}

func unmarshalAgentJSON(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	if len(payload) == 0 {
		return nil
	}
	return payload
}
