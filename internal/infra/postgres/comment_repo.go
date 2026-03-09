package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/comment"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

const createCommentSQL = `
	INSERT INTO comments (id, user_id, commentable_type, commentable_id, parent_id, content,
	                      is_edited, is_deleted, like_count, reply_count, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`

func (r *CommentRepository) Create(ctx context.Context, c *comment.Comment) error {
	_, err := r.pool.Exec(ctx, createCommentSQL,
		c.ID, c.UserID, c.CommentableType, c.CommentableID, c.ParentID, c.Content,
		c.IsEdited, c.IsDeleted, c.LikeCount, c.ReplyCount, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

const getCommentByIDSQL = `
	SELECT id, user_id, commentable_type, commentable_id, parent_id, content,
	       is_edited, is_deleted, like_count, reply_count, created_at, updated_at
	FROM comments WHERE id = $1
`

func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	var c comment.Comment
	err := r.pool.QueryRow(ctx, getCommentByIDSQL, id).Scan(
		&c.ID, &c.UserID, &c.CommentableType, &c.CommentableID, &c.ParentID, &c.Content,
		&c.IsEdited, &c.IsDeleted, &c.LikeCount, &c.ReplyCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, comment.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	return &c, nil
}

const updateCommentSQL = `
	UPDATE comments
	SET content = $2, is_edited = $3, is_deleted = $4, updated_at = $5
	WHERE id = $1
`

func (r *CommentRepository) Update(ctx context.Context, c *comment.Comment) error {
	_, err := r.pool.Exec(ctx, updateCommentSQL,
		c.ID, c.Content, c.IsEdited, c.IsDeleted, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

const deleteCommentSQL = `
	UPDATE comments SET is_deleted = true, updated_at = NOW() WHERE id = $1
`

func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, deleteCommentSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) List(ctx context.Context, filter comment.ListFilter) ([]*comment.Comment, int64, error) {
	query := `
		SELECT c.id, c.user_id, c.commentable_type, c.commentable_id, c.parent_id, c.content,
		       c.is_edited, c.is_deleted, c.like_count, c.reply_count, c.created_at, c.updated_at,
		       u.username, u.avatar_key
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.commentable_type = $1 AND c.commentable_id = $2 AND c.is_deleted = false
	`
	countQuery := `
		SELECT COUNT(*) FROM comments
		WHERE commentable_type = $1 AND commentable_id = $2 AND is_deleted = false
	`
	args := []any{filter.CommentableType, filter.CommentableID}
	argIndex := 3

	if filter.ParentID != nil {
		query += fmt.Sprintf(" AND c.parent_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND parent_id = $%d", argIndex)
		args = append(args, *filter.ParentID)
		argIndex++
	} else {
		query += " AND c.parent_id IS NULL"
		countQuery += " AND parent_id IS NULL"
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND c.user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	query += " ORDER BY c.created_at ASC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	comments := make([]*comment.Comment, 0)
	for rows.Next() {
		var c comment.Comment
		err := rows.Scan(
			&c.ID, &c.UserID, &c.CommentableType, &c.CommentableID, &c.ParentID, &c.Content,
			&c.IsEdited, &c.IsDeleted, &c.LikeCount, &c.ReplyCount, &c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorAvatarKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return comments, total, nil
}

const incrementReplyCountSQL = `UPDATE comments SET reply_count = reply_count + 1 WHERE id = $1`

func (r *CommentRepository) IncrementReplyCount(ctx context.Context, commentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, incrementReplyCountSQL, commentID)
	return err
}

const decrementReplyCountSQL = `UPDATE comments SET reply_count = reply_count - 1 WHERE id = $1`

func (r *CommentRepository) DecrementReplyCount(ctx context.Context, commentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, decrementReplyCountSQL, commentID)
	return err
}

const likeCommentSQL = `
	INSERT INTO comment_likes (id, user_id, comment_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id, comment_id) DO NOTHING
`

func (r *CommentRepository) LikeComment(ctx context.Context, like *comment.CommentLike) error {
	_, err := r.pool.Exec(ctx, likeCommentSQL, like.ID, like.UserID, like.CommentID, like.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to like comment: %w", err)
	}
	return nil
}

const unlikeCommentSQL = `DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2`

func (r *CommentRepository) UnlikeComment(ctx context.Context, userID, commentID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, unlikeCommentSQL, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to unlike comment: %w", err)
	}
	if result.RowsAffected() == 0 {
		return comment.ErrNotLiked
	}
	return nil
}

const hasLikedSQL = `SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = $1 AND comment_id = $2)`

func (r *CommentRepository) HasLiked(ctx context.Context, userID, commentID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, hasLikedSQL, userID, commentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check like status: %w", err)
	}
	return exists, nil
}

const incrementLikeCountSQL = `UPDATE comments SET like_count = like_count + 1 WHERE id = $1`

func (r *CommentRepository) IncrementLikeCount(ctx context.Context, commentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, incrementLikeCountSQL, commentID)
	return err
}

const decrementLikeCountSQL = `UPDATE comments SET like_count = like_count - 1 WHERE id = $1`

func (r *CommentRepository) DecrementLikeCount(ctx context.Context, commentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, decrementLikeCountSQL, commentID)
	return err
}
