package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ModerationStatus tracks the content safety review state of a post.
type ModerationStatus string

const (
	ModerationPending  ModerationStatus = "pending"
	ModerationApproved ModerationStatus = "approved"
	ModerationBlocked  ModerationStatus = "blocked"
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
	ID               uuid.UUID        `json:"id"`
	AuthorID         uuid.UUID        `json:"author_id"`
	GroupID          *uuid.UUID       `json:"group_id,omitempty"`
	Title            string           `json:"title,omitempty"` // optional, for long posts
	Content          string           `json:"content"`
	MediaURLs        []string         `json:"media_urls,omitempty"`
	Tags             []string         `json:"tags,omitempty"`
	ContentLabels    map[string]bool  `json:"content_labels,omitempty"` // e.g. {"is_ai_generated": true}
	Visibility       Visibility       `json:"visibility"`
	ModerationStatus ModerationStatus `json:"moderation_status"`
	LikeCount        int              `json:"like_count"`
	CommentCount     int              `json:"comment_count"`
	IsPinned         bool             `json:"is_pinned"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        *time.Time       `json:"deleted_at,omitempty"`

	// Joined fields
	AuthorUsername   string  `json:"author_username,omitempty"`
	AuthorAvatarKey  *string `json:"author_avatar_key,omitempty"`
	GroupName        *string `json:"group_name,omitempty"`
	IsBookmarkedByMe bool    `json:"is_bookmarked_by_me,omitempty"`
	IsLikedByMe      bool    `json:"is_liked_by_me,omitempty"`
}

// PostLike represents a like on a post
type PostLike struct {
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ListFilter holds filtering options for listing posts
type ListFilter struct {
	AuthorID         *uuid.UUID
	GroupID          *uuid.UUID
	Tags             []string
	Search           string // full-text search on title+content
	Visibility       *Visibility
	ModerationStatus *ModerationStatus // if set, filter by this status
	Cursor           *time.Time        // for cursor-based pagination
	Page             int
	PageSize         int
	SortByScore      bool // if true, sort by engagement score (BIZ-02)
}

var (
	ErrNotFound  = errors.New("post not found")
	ErrForbidden = errors.New("not authorized to modify this post")
)
