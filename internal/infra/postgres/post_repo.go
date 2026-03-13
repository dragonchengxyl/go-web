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
	"github.com/studio/platform/internal/domain/post"
)

// PostRepository implements post.Repository using PostgreSQL
type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

const createPostSQL = `
	INSERT INTO posts (id, author_id, group_id, title, content, media_urls, tags, visibility,
	                   moderation_status, content_labels, like_count, comment_count, is_pinned, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	labelsJSON, err := json.Marshal(p.ContentLabels)
	if err != nil {
		return fmt.Errorf("marshal content_labels: %w", err)
	}
	_, err = r.pool.Exec(ctx, createPostSQL,
		p.ID, p.AuthorID, p.GroupID, p.Title, p.Content, p.MediaURLs, p.Tags, string(p.Visibility),
		string(p.ModerationStatus), labelsJSON, p.LikeCount, p.CommentCount, p.IsPinned, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

const getPostByIDSQL = `
	SELECT p.id, p.author_id, p.title, p.content, p.media_urls, p.tags, p.visibility,
	       p.moderation_status, p.content_labels, p.like_count, p.comment_count, p.is_pinned,
	       p.created_at, p.updated_at, p.deleted_at,
	       u.username, u.avatar_key, p.group_id, g.name
	FROM posts p
	JOIN users u ON u.id = p.author_id
	LEFT JOIN groups g ON g.id = p.group_id
	WHERE p.id = $1 AND p.deleted_at IS NULL
`

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	var p post.Post
	var visStr, modStr string
	var labelsJSON []byte
	err := r.pool.QueryRow(ctx, getPostByIDSQL, id).Scan(
		&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.MediaURLs, &p.Tags, &visStr,
		&modStr, &labelsJSON, &p.LikeCount, &p.CommentCount, &p.IsPinned,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		&p.AuthorUsername, &p.AuthorAvatarKey, &p.GroupID, &p.GroupName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	p.Visibility = post.Visibility(visStr)
	p.ModerationStatus = post.ModerationStatus(modStr)
	if len(labelsJSON) > 0 {
		_ = json.Unmarshal(labelsJSON, &p.ContentLabels)
	}
	return &p, nil
}

const updatePostSQL = `
	UPDATE posts
	SET title = $2, content = $3, media_urls = $4, tags = $5, visibility = $6,
	    is_pinned = $7, updated_at = $8
	WHERE id = $1 AND deleted_at IS NULL
`

func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	_, err := r.pool.Exec(ctx, updatePostSQL,
		p.ID, p.Title, p.Content, p.MediaURLs, p.Tags, string(p.Visibility),
		p.IsPinned, p.UpdatedAt,
	)
	return err
}

const deletePostSQL = `UPDATE posts SET deleted_at = $2 WHERE id = $1`

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, deletePostSQL, id, time.Now())
	return err
}

func (r *PostRepository) List(ctx context.Context, filter post.ListFilter) ([]*post.Post, int64, error) {
	args := []any{}
	idx := 1
	where := "WHERE p.deleted_at IS NULL"

	if filter.AuthorID != nil {
		where += fmt.Sprintf(" AND p.author_id = $%d", idx)
		args = append(args, *filter.AuthorID)
		idx++
	}
	if filter.GroupID != nil {
		where += fmt.Sprintf(" AND p.group_id = $%d", idx)
		args = append(args, *filter.GroupID)
		idx++
	}
	if filter.Visibility != nil {
		where += fmt.Sprintf(" AND p.visibility = $%d", idx)
		args = append(args, string(*filter.Visibility))
		idx++
	}
	if filter.ModerationStatus != nil {
		where += fmt.Sprintf(" AND p.moderation_status = $%d", idx)
		args = append(args, string(*filter.ModerationStatus))
		idx++
	}
	if len(filter.Tags) > 0 {
		where += fmt.Sprintf(" AND p.tags && $%d", idx)
		args = append(args, filter.Tags)
		idx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.content ILIKE $%d)", idx, idx)
		args = append(args, "%"+filter.Search+"%")
		idx++
	}

	countSQL := "SELECT COUNT(*) FROM posts p " + where
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	orderBy := "p.is_pinned DESC, p.created_at DESC"
	if filter.SortByScore {
		// Engagement score: (like_count + comment_count*3) / time_decay
		// EXTRACT(EPOCH ...) gives seconds; add 1 to avoid division by zero.
		orderBy = "p.is_pinned DESC, (p.like_count + p.comment_count * 3)::float / GREATEST(EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600 + 1, 1) DESC, p.created_at DESC"
	}

	listSQL := `SELECT p.id, p.author_id, p.title, p.content, p.media_urls, p.tags, p.visibility,
	                   p.moderation_status, p.content_labels, p.like_count, p.comment_count, p.is_pinned,
	                   p.created_at, p.updated_at, p.deleted_at,
	                   u.username, u.avatar_key, p.group_id, g.name
	            FROM posts p JOIN users u ON u.id = p.author_id
	            LEFT JOIN groups g ON g.id = p.group_id ` + where + " ORDER BY " + orderBy

	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		listSQL += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	posts, _, err := scanPosts(rows)
	return posts, total, err
}

func (r *PostRepository) ListFeed(ctx context.Context, followeeIDs []uuid.UUID, filter post.ListFilter) ([]*post.Post, int64, error) {
	if len(followeeIDs) == 0 {
		return []*post.Post{}, 0, nil
	}

	args := []any{followeeIDs}
	idx := 2
	where := "WHERE p.deleted_at IS NULL AND p.author_id = ANY($1) AND p.visibility = 'public' AND p.moderation_status = 'approved'"

	if filter.Cursor != nil {
		where += fmt.Sprintf(" AND p.created_at < $%d", idx)
		args = append(args, *filter.Cursor)
		idx++
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM posts p " + where
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count feed: %w", err)
	}

	listSQL := `SELECT p.id, p.author_id, p.title, p.content, p.media_urls, p.tags, p.visibility,
	                   p.moderation_status, p.content_labels, p.like_count, p.comment_count, p.is_pinned,
	                   p.created_at, p.updated_at, p.deleted_at,
	                   u.username, u.avatar_key, p.group_id, g.name
	            FROM posts p JOIN users u ON u.id = p.author_id
	            LEFT JOIN groups g ON g.id = p.group_id ` + where + " ORDER BY p.created_at DESC"
	if filter.PageSize > 0 {
		listSQL += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, filter.PageSize)
	}

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feed: %w", err)
	}
	defer rows.Close()

	posts, _, err := scanPosts(rows)
	return posts, total, err
}

func scanPosts(rows pgx.Rows) ([]*post.Post, int64, error) {
	posts := make([]*post.Post, 0)
	for rows.Next() {
		var p post.Post
		var visStr, modStr string
		var labelsJSON []byte
		err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.MediaURLs, &p.Tags, &visStr,
			&modStr, &labelsJSON, &p.LikeCount, &p.CommentCount, &p.IsPinned,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
			&p.AuthorUsername, &p.AuthorAvatarKey, &p.GroupID, &p.GroupName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		p.Visibility = post.Visibility(visStr)
		p.ModerationStatus = post.ModerationStatus(modStr)
		if len(labelsJSON) > 0 {
			_ = json.Unmarshal(labelsJSON, &p.ContentLabels)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	return posts, int64(len(posts)), nil
}

func (r *PostRepository) GetHotTags(ctx context.Context, limit int) ([]string, error) {
	const sql = `
		SELECT tag, COUNT(*) AS cnt
		FROM posts, unnest(tags) AS tag
		WHERE deleted_at IS NULL AND visibility = 'public'
		GROUP BY tag
		ORDER BY cnt DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get hot tags: %w", err)
	}
	defer rows.Close()

	tags := make([]string, 0, limit)
	for rows.Next() {
		var tag string
		var cnt int64
		if err := rows.Scan(&tag, &cnt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *PostRepository) GetGroupHotTags(ctx context.Context, groupID uuid.UUID, limit int) ([]string, error) {
	const sql = `
		SELECT tag, COUNT(*) AS cnt
		FROM posts p, unnest(p.tags) AS tag
		WHERE p.deleted_at IS NULL
		  AND p.group_id = $1
		  AND p.visibility = 'public'
		  AND p.moderation_status = 'approved'
		GROUP BY tag
		ORDER BY cnt DESC, tag ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, sql, groupID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get group hot tags: %w", err)
	}
	defer rows.Close()

	tags := make([]string, 0, limit)
	for rows.Next() {
		var tag string
		var cnt int64
		if err := rows.Scan(&tag, &cnt); err != nil {
			return nil, fmt.Errorf("failed to scan group tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *PostRepository) LikePost(ctx context.Context, like *post.PostLike) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO post_likes (post_id, user_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		like.PostID, like.UserID, like.CreatedAt,
	)
	return err
}

func (r *PostRepository) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	return err
}

func (r *PostRepository) HasLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	return exists, err
}

func (r *PostRepository) IncrementLikeCount(ctx context.Context, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET like_count = like_count + 1 WHERE id = $1`, postID)
	return err
}

func (r *PostRepository) DecrementLikeCount(ctx context.Context, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, postID)
	return err
}

func (r *PostRepository) IncrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET comment_count = comment_count + 1 WHERE id = $1`, postID)
	return err
}

func (r *PostRepository) DecrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, postID)
	return err
}

func (r *PostRepository) UpdateModerationStatus(ctx context.Context, id uuid.UUID, status post.ModerationStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE posts SET moderation_status = $2 WHERE id = $1`,
		id, string(status),
	)
	return err
}
