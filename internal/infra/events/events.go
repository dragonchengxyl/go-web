package events

import (
	"context"
	"encoding/json"
)

const (
	EventPostCreated     = "post.created"
	EventPostModerated   = "post.moderated"
	EventPostLiked       = "post.liked"
	EventUserFollowed    = "user.followed"
	EventTipSent         = "tip.sent"
	EventCommentCreated  = "comment.created"
	EventAudioJobCreated = "audio.job.created"
)

const (
	GroupModeration   = "moderation-group"
	GroupNotification = "notification-group"
	GroupAudioJobs    = "audio-job-group"
)

type Event struct {
	EventID string          `json:"event_id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type HandlerFunc func(ctx context.Context, event Event) error

type PostCreatedPayload struct {
	PostID    string   `json:"post_id"`
	AuthorID  string   `json:"author_id"`
	Content   string   `json:"content"`
	MediaURLs []string `json:"media_urls"`
}

type PostModeratedPayload struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type PostLikedPayload struct {
	PostID   string `json:"post_id"`
	ActorID  string `json:"actor_id"`
	AuthorID string `json:"author_id"`
}

type UserFollowedPayload struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

type TipSentPayload struct {
	TipID       string `json:"tip_id"`
	SenderID    string `json:"sender_id"`
	ReceiverID  string `json:"receiver_id"`
	AmountCents int    `json:"amount_cents"`
}

type CommentCreatedPayload struct {
	CommentID     string `json:"comment_id"`
	PostID        string `json:"post_id"`
	CommentableID string `json:"commentable_id"`
	AuthorID      string `json:"author_id"`
	TargetUserID  string `json:"target_user_id"`
}

type AudioJobCreatedPayload struct {
	JobID string `json:"job_id"`
}
