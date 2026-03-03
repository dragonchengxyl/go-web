package comment

import (
	"time"

	"github.com/google/uuid"
)

// CommentableType represents the type of entity being commented on
type CommentableType string

const (
	CommentableTypeGame  CommentableType = "game"
	CommentableTypeTrack CommentableType = "track"
	CommentableTypePost  CommentableType = "post"
)

// Comment represents a comment entity
type Comment struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	CommentableType CommentableType `json:"commentable_type"`
	CommentableID   uuid.UUID       `json:"commentable_id"`
	ParentID        *uuid.UUID      `json:"parent_id,omitempty"`
	Content         string          `json:"content"`
	IsEdited        bool            `json:"is_edited"`
	IsDeleted       bool            `json:"is_deleted"`
	LikeCount       int             `json:"like_count"`
	ReplyCount      int             `json:"reply_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CommentLike represents a comment like
type CommentLike struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CommentID uuid.UUID `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}
