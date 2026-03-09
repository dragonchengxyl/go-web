package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Visibility represents who can see a post
type Visibility string

const (
	VisibilityPublic        Visibility = "public"
	VisibilityFollowersOnly Visibility = "followers_only"
	VisibilityPrivate       Visibility = "private"
)

// Post represents a community post/update
type Post struct {
	ID           uuid.UUID  `json:"id"`
	AuthorID     uuid.UUID  `json:"author_id"`
	Title        string     `json:"title,omitempty"` // optional, for long posts
	Content      string     `json:"content"`
	MediaURLs    []string   `json:"media_urls,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Visibility   Visibility `json:"visibility"`
	LikeCount    int        `json:"like_count"`
	CommentCount int        `json:"comment_count"`
	IsPinned     bool       `json:"is_pinned"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	// Joined fields
	AuthorUsername  string  `json:"author_username,omitempty"`
	AuthorAvatarKey *string `json:"author_avatar_key,omitempty"`
	IsLikedByMe     bool    `json:"is_liked_by_me,omitempty"`
}

// PostLike represents a like on a post
type PostLike struct {
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ListFilter holds filtering options for listing posts
type ListFilter struct {
	AuthorID   *uuid.UUID
	Tags       []string
	Visibility *Visibility
	Cursor     *time.Time // for cursor-based pagination
	Page       int
	PageSize   int
}

var (
	ErrNotFound  = errors.New("post not found")
	ErrForbidden = errors.New("not authorized to modify this post")
)
