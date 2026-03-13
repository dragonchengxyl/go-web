package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/bookmark"
)

type BookmarkRepository struct {
	pool *pgxpool.Pool
}

func NewBookmarkRepository(pool *pgxpool.Pool) *BookmarkRepository {
	return &BookmarkRepository{pool: pool}
}

func (r *BookmarkRepository) Create(ctx context.Context, item *bookmark.Bookmark) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_bookmarks (user_id, target_type, target_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, target_type, target_id) DO NOTHING
	`, item.UserID, string(item.TargetType), item.TargetID, item.CreatedAt)
	return err
}

func (r *BookmarkRepository) Delete(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_bookmarks WHERE user_id = $1 AND target_type = $2 AND target_id = $3
	`, userID, string(targetType), targetID)
	return err
}

func (r *BookmarkRepository) DeleteBatch(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetIDs []uuid.UUID) error {
	if len(targetIDs) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_bookmarks WHERE user_id = $1 AND target_type = $2 AND target_id = ANY($3)
	`, userID, string(targetType), targetIDs)
	return err
}

func (r *BookmarkRepository) Exists(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_bookmarks WHERE user_id = $1 AND target_type = $2 AND target_id = $3
		)
	`, userID, string(targetType), targetID).Scan(&exists)
	return exists, err
}

func (r *BookmarkRepository) List(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, page, pageSize int, sort string) ([]*bookmark.Bookmark, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_bookmarks WHERE user_id = $1 AND target_type = $2
	`, userID, string(targetType)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookmarks: %w", err)
	}

	orderBy := "created_at DESC"
	if sort == "oldest" {
		orderBy = "created_at ASC"
	}

	rows, err := r.pool.Query(ctx, `
		SELECT user_id, target_type, target_id, created_at
		FROM user_bookmarks
		WHERE user_id = $1 AND target_type = $2
		ORDER BY `+orderBy+`
		LIMIT $3 OFFSET $4
	`, userID, string(targetType), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	items := make([]*bookmark.Bookmark, 0, pageSize)
	for rows.Next() {
		var item bookmark.Bookmark
		var targetTypeStr string
		if err := rows.Scan(&item.UserID, &targetTypeStr, &item.TargetID, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan bookmark: %w", err)
		}
		item.TargetType = bookmark.TargetType(targetTypeStr)
		items = append(items, &item)
	}
	return items, total, rows.Err()
}
