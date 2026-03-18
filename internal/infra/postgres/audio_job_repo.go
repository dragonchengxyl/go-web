package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/audiojob"
)

type AudioJobRepository struct {
	pool *pgxpool.Pool
}

func NewAudioJobRepository(pool *pgxpool.Pool) *AudioJobRepository {
	return &AudioJobRepository{pool: pool}
}

const createAudioJobSQL = `
	INSERT INTO audio_jobs (
		id, user_id, title, task_type, status, source_audio_url, reference_audio_url,
		prompt, params, result, error_message, created_at, updated_at, started_at, finished_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13, $14, $15
	)
`

func (r *AudioJobRepository) Create(ctx context.Context, job *audiojob.Job) error {
	paramsJSON, err := marshalJSONMap(job.Params, true)
	if err != nil {
		return fmt.Errorf("marshal audio job params: %w", err)
	}
	resultJSON, err := marshalJSONMap(job.Result, false)
	if err != nil {
		return fmt.Errorf("marshal audio job result: %w", err)
	}

	_, err = r.pool.Exec(ctx, createAudioJobSQL,
		job.ID,
		job.UserID,
		job.Title,
		job.TaskType,
		job.Status,
		job.SourceAudioURL,
		job.ReferenceAudioURL,
		job.Prompt,
		paramsJSON,
		resultJSON,
		job.ErrorMessage,
		job.CreatedAt,
		job.UpdatedAt,
		job.StartedAt,
		job.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("create audio job: %w", err)
	}
	return nil
}

const getAudioJobByIDSQL = `
	SELECT id, user_id, title, task_type, status, source_audio_url, reference_audio_url,
	       prompt, params, result, error_message, created_at, updated_at, started_at, finished_at
	FROM audio_jobs
	WHERE id = $1
`

func (r *AudioJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*audiojob.Job, error) {
	job, err := scanAudioJob(r.pool.QueryRow(ctx, getAudioJobByIDSQL, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, audiojob.ErrNotFound
		}
		return nil, fmt.Errorf("get audio job: %w", err)
	}
	return job, nil
}

func (r *AudioJobRepository) ListByUser(ctx context.Context, filter audiojob.ListFilter) ([]*audiojob.Job, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	var (
		args       = []any{filter.UserID}
		conditions = []string{"user_id = $1"}
	)

	if filter.Status != nil && *filter.Status != "" {
		args = append(args, *filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if filter.TaskType != nil && *filter.TaskType != "" {
		args = append(args, *filter.TaskType)
		conditions = append(conditions, fmt.Sprintf("task_type = $%d", len(args)))
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	query := fmt.Sprintf(`
		SELECT id, user_id, title, task_type, status, source_audio_url, reference_audio_url,
		       prompt, params, result, error_message, created_at, updated_at, started_at, finished_at,
		       COUNT(*) OVER() AS total_count
		FROM audio_jobs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), limitPos, offsetPos)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audio jobs: %w", err)
	}
	defer rows.Close()

	items := make([]*audiojob.Job, 0)
	var total int64
	for rows.Next() {
		job, count, err := scanAudioJobWithTotal(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan audio job: %w", err)
		}
		total = count
		items = append(items, job)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audio jobs: %w", err)
	}
	return items, total, nil
}

const updateAudioJobSQL = `
	UPDATE audio_jobs
	SET title = $2,
	    task_type = $3,
	    status = $4,
	    source_audio_url = $5,
	    reference_audio_url = $6,
	    prompt = $7,
	    params = $8,
	    result = $9,
	    error_message = $10,
	    updated_at = $11,
	    started_at = $12,
	    finished_at = $13
	WHERE id = $1
`

func (r *AudioJobRepository) Update(ctx context.Context, job *audiojob.Job) error {
	paramsJSON, err := marshalJSONMap(job.Params, true)
	if err != nil {
		return fmt.Errorf("marshal audio job params: %w", err)
	}
	resultJSON, err := marshalJSONMap(job.Result, false)
	if err != nil {
		return fmt.Errorf("marshal audio job result: %w", err)
	}

	result, err := r.pool.Exec(ctx, updateAudioJobSQL,
		job.ID,
		job.Title,
		job.TaskType,
		job.Status,
		job.SourceAudioURL,
		job.ReferenceAudioURL,
		job.Prompt,
		paramsJSON,
		resultJSON,
		job.ErrorMessage,
		job.UpdatedAt,
		job.StartedAt,
		job.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("update audio job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return audiojob.ErrNotFound
	}
	return nil
}

type audioJobScanner interface {
	Scan(dest ...any) error
}

func scanAudioJob(scanner audioJobScanner) (*audiojob.Job, error) {
	var (
		job        audiojob.Job
		paramsJSON []byte
		resultJSON []byte
	)

	err := scanner.Scan(
		&job.ID,
		&job.UserID,
		&job.Title,
		&job.TaskType,
		&job.Status,
		&job.SourceAudioURL,
		&job.ReferenceAudioURL,
		&job.Prompt,
		&paramsJSON,
		&resultJSON,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.StartedAt,
		&job.FinishedAt,
	)
	if err != nil {
		return nil, err
	}

	job.Params, err = unmarshalJSONMap(paramsJSON)
	if err != nil {
		return nil, err
	}
	job.Result, err = unmarshalJSONMap(resultJSON)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func scanAudioJobWithTotal(rows pgx.Rows) (*audiojob.Job, int64, error) {
	var (
		job        audiojob.Job
		paramsJSON []byte
		resultJSON []byte
		total      int64
	)

	err := rows.Scan(
		&job.ID,
		&job.UserID,
		&job.Title,
		&job.TaskType,
		&job.Status,
		&job.SourceAudioURL,
		&job.ReferenceAudioURL,
		&job.Prompt,
		&paramsJSON,
		&resultJSON,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&total,
	)
	if err != nil {
		return nil, 0, err
	}

	job.Params, err = unmarshalJSONMap(paramsJSON)
	if err != nil {
		return nil, 0, err
	}
	job.Result, err = unmarshalJSONMap(resultJSON)
	if err != nil {
		return nil, 0, err
	}
	return &job, total, nil
}

func marshalJSONMap(payload map[string]any, emptyObject bool) ([]byte, error) {
	if payload == nil {
		if emptyObject {
			return []byte("{}"), nil
		}
		return nil, nil
	}
	return json.Marshal(payload)
}

func unmarshalJSONMap(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal json map: %w", err)
	}
	return payload, nil
}
