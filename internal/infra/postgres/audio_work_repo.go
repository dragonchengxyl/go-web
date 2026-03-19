package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/audiowork"
	"github.com/studio/platform/internal/domain/post"
)

type AudioWorkRepository struct {
	pool *pgxpool.Pool
}

func NewAudioWorkRepository(pool *pgxpool.Pool) *AudioWorkRepository {
	return &AudioWorkRepository{pool: pool}
}

const createAudioWorkSQL = `
	INSERT INTO audio_works (
		id, author_id, source_job_id, title, description, cover_image_url, audio_url,
		duration_sec, visibility, moderation_status, moderation_note, like_count, comment_count, tags, waveform_preview, metadata, published_at, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
	)
`

func (r *AudioWorkRepository) Create(ctx context.Context, work *audiowork.Work) error {
	tagsJSON, err := json.Marshal(work.Tags)
	if err != nil {
		return fmt.Errorf("marshal audio work tags: %w", err)
	}
	waveformJSON, err := json.Marshal(work.WaveformPreview)
	if err != nil {
		return fmt.Errorf("marshal waveform preview: %w", err)
	}
	metadataJSON, err := marshalJSONMap(work.Metadata, true)
	if err != nil {
		return fmt.Errorf("marshal audio work metadata: %w", err)
	}

	_, err = r.pool.Exec(ctx, createAudioWorkSQL,
		work.ID,
		work.AuthorID,
		work.SourceJobID,
		work.Title,
		work.Description,
		work.CoverImageURL,
		work.AudioURL,
		work.DurationSec,
		work.Visibility,
		work.ModerationStatus,
		work.ModerationNote,
		work.LikeCount,
		work.CommentCount,
		tagsJSON,
		waveformJSON,
		metadataJSON,
		work.PublishedAt,
		work.CreatedAt,
		work.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return audiowork.ErrAlreadyPublished
		}
		return fmt.Errorf("create audio work: %w", err)
	}
	return nil
}

const getAudioWorkByIDSQL = `
	SELECT w.id, w.author_id, w.source_job_id, w.title, w.description, w.cover_image_url, w.audio_url,
	       w.duration_sec, w.visibility, w.moderation_status, w.moderation_note, w.like_count, w.comment_count, w.tags, w.waveform_preview, w.metadata,
	       w.published_at, w.created_at, w.updated_at, COALESCE(u.username, '')
	FROM audio_works w
	LEFT JOIN users u ON u.id = w.author_id
	WHERE w.id = $1
`

func (r *AudioWorkRepository) GetByID(ctx context.Context, id uuid.UUID) (*audiowork.Work, error) {
	work, err := scanAudioWork(r.pool.QueryRow(ctx, getAudioWorkByIDSQL, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, audiowork.ErrNotFound
		}
		return nil, fmt.Errorf("get audio work: %w", err)
	}
	return work, nil
}

func (r *AudioWorkRepository) List(ctx context.Context, filter audiowork.ListFilter) ([]*audiowork.Work, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var (
		args       []any
		conditions []string
	)

	if filter.AuthorID != nil {
		args = append(args, *filter.AuthorID)
		conditions = append(conditions, fmt.Sprintf("w.author_id = $%d", len(args)))
	}
	if filter.Visibility != nil && *filter.Visibility != "" {
		args = append(args, *filter.Visibility)
		conditions = append(conditions, fmt.Sprintf("w.visibility = $%d", len(args)))
	}
	if filter.ModerationStatus != nil && *filter.ModerationStatus != "" {
		args = append(args, *filter.ModerationStatus)
		conditions = append(conditions, fmt.Sprintf("w.moderation_status = $%d", len(args)))
	}
	if len(conditions) == 0 {
		conditions = append(conditions, "TRUE")
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	query := fmt.Sprintf(`
		SELECT w.id, w.author_id, w.source_job_id, w.title, w.description, w.cover_image_url, w.audio_url,
		       w.duration_sec, w.visibility, w.moderation_status, w.moderation_note, w.like_count, w.comment_count, w.tags, w.waveform_preview, w.metadata,
		       w.published_at, w.created_at, w.updated_at, COALESCE(u.username, ''),
		       COUNT(*) OVER() AS total_count
		FROM audio_works w
		LEFT JOIN users u ON u.id = w.author_id
		WHERE %s
		ORDER BY w.published_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), limitPos, offsetPos)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audio works: %w", err)
	}
	defer rows.Close()

	items := make([]*audiowork.Work, 0)
	var total int64
	for rows.Next() {
		work, count, err := scanAudioWorkWithTotal(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan audio work: %w", err)
		}
		total = count
		items = append(items, work)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audio works: %w", err)
	}
	return items, total, nil
}

const updateAudioWorkSQL = `
	UPDATE audio_works
	SET title = $2,
	    description = $3,
	    cover_image_url = $4,
	    visibility = $5,
	    tags = $6,
	    moderation_status = $7,
	    moderation_note = $8,
	    updated_at = $9
	WHERE id = $1
`

func (r *AudioWorkRepository) Update(ctx context.Context, work *audiowork.Work) error {
	tagsJSON, err := json.Marshal(work.Tags)
	if err != nil {
		return fmt.Errorf("marshal audio work tags: %w", err)
	}
	result, err := r.pool.Exec(ctx, updateAudioWorkSQL,
		work.ID,
		work.Title,
		work.Description,
		work.CoverImageURL,
		work.Visibility,
		tagsJSON,
		work.ModerationStatus,
		work.ModerationNote,
		work.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update audio work: %w", err)
	}
	if result.RowsAffected() == 0 {
		return audiowork.ErrNotFound
	}
	return nil
}

func (r *AudioWorkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM audio_works WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete audio work: %w", err)
	}
	if result.RowsAffected() == 0 {
		return audiowork.ErrNotFound
	}
	return nil
}

func (r *AudioWorkRepository) UpdateModerationStatus(ctx context.Context, id uuid.UUID, status post.ModerationStatus, note *string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE audio_works
		SET moderation_status = $2, moderation_note = $3, updated_at = NOW()
		WHERE id = $1
	`, id, status, note)
	if err != nil {
		return fmt.Errorf("update audio work moderation status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return audiowork.ErrNotFound
	}
	return nil
}

type audioWorkScanner interface {
	Scan(dest ...any) error
}

func scanAudioWork(scanner audioWorkScanner) (*audiowork.Work, error) {
	var (
		work         audiowork.Work
		tagsJSON     []byte
		waveformJSON []byte
		metadataJSON []byte
	)

	err := scanner.Scan(
		&work.ID,
		&work.AuthorID,
		&work.SourceJobID,
		&work.Title,
		&work.Description,
		&work.CoverImageURL,
		&work.AudioURL,
		&work.DurationSec,
		&work.Visibility,
		&work.ModerationStatus,
		&work.ModerationNote,
		&work.LikeCount,
		&work.CommentCount,
		&tagsJSON,
		&waveformJSON,
		&metadataJSON,
		&work.PublishedAt,
		&work.CreatedAt,
		&work.UpdatedAt,
		&work.AuthorUsername,
	)
	if err != nil {
		return nil, err
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &work.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal audio work tags: %w", err)
		}
	}
	if len(waveformJSON) > 0 {
		if err := json.Unmarshal(waveformJSON, &work.WaveformPreview); err != nil {
			return nil, fmt.Errorf("unmarshal waveform preview: %w", err)
		}
	}
	metadata, err := unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, err
	}
	work.Metadata = metadata
	return &work, nil
}

func scanAudioWorkWithTotal(rows pgx.Rows) (*audiowork.Work, int64, error) {
	var (
		work         audiowork.Work
		tagsJSON     []byte
		waveformJSON []byte
		metadataJSON []byte
		total        int64
	)

	err := rows.Scan(
		&work.ID,
		&work.AuthorID,
		&work.SourceJobID,
		&work.Title,
		&work.Description,
		&work.CoverImageURL,
		&work.AudioURL,
		&work.DurationSec,
		&work.Visibility,
		&work.ModerationStatus,
		&work.ModerationNote,
		&work.LikeCount,
		&work.CommentCount,
		&tagsJSON,
		&waveformJSON,
		&metadataJSON,
		&work.PublishedAt,
		&work.CreatedAt,
		&work.UpdatedAt,
		&work.AuthorUsername,
		&total,
	)
	if err != nil {
		return nil, 0, err
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &work.Tags); err != nil {
			return nil, 0, fmt.Errorf("unmarshal audio work tags: %w", err)
		}
	}
	if len(waveformJSON) > 0 {
		if err := json.Unmarshal(waveformJSON, &work.WaveformPreview); err != nil {
			return nil, 0, fmt.Errorf("unmarshal waveform preview: %w", err)
		}
	}
	metadata, err := unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, 0, err
	}
	work.Metadata = metadata
	return &work, total, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *AudioWorkRepository) Like(ctx context.Context, userID, workID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		INSERT INTO audio_work_likes (user_id, work_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, work_id) DO NOTHING
	`, userID, workID)
	if err != nil {
		return fmt.Errorf("like audio work: %w", err)
	}
	if result.RowsAffected() == 0 {
		return audiowork.ErrAlreadyLiked
	}
	return nil
}

func (r *AudioWorkRepository) Unlike(ctx context.Context, userID, workID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM audio_work_likes WHERE user_id = $1 AND work_id = $2`, userID, workID)
	if err != nil {
		return fmt.Errorf("unlike audio work: %w", err)
	}
	return nil
}

func (r *AudioWorkRepository) HasLiked(ctx context.Context, userID, workID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM audio_work_likes WHERE user_id = $1 AND work_id = $2)
	`, userID, workID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check audio work like state: %w", err)
	}
	return exists, nil
}

func (r *AudioWorkRepository) IncrementLikeCount(ctx context.Context, workID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE audio_works SET like_count = like_count + 1 WHERE id = $1`, workID)
	if err != nil {
		return fmt.Errorf("increment audio work like count: %w", err)
	}
	return nil
}

func (r *AudioWorkRepository) DecrementLikeCount(ctx context.Context, workID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE audio_works SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, workID)
	if err != nil {
		return fmt.Errorf("decrement audio work like count: %w", err)
	}
	return nil
}

func (r *AudioWorkRepository) IncrementCommentCount(ctx context.Context, workID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE audio_works SET comment_count = comment_count + 1 WHERE id = $1`, workID)
	if err != nil {
		return fmt.Errorf("increment audio work comment count: %w", err)
	}
	return nil
}

func (r *AudioWorkRepository) DecrementCommentCount(ctx context.Context, workID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE audio_works SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, workID)
	if err != nil {
		return fmt.Errorf("decrement audio work comment count: %w", err)
	}
	return nil
}
