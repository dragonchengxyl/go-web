package streams

import "encoding/json"

// Event type constants for the furry:events Redis stream.
const (
	EventPostCreated    = "post.created"
	EventPostModerated  = "post.moderated"
	EventPostLiked      = "post.liked"
	EventUserFollowed   = "user.followed"
	EventTipSent        = "tip.sent"
	EventCommentCreated = "comment.created"
)

// StreamKey is the Redis stream key used for all platform events.
const StreamKey = "furry:events"

// Consumer group names.
const (
	GroupModeration   = "moderation-group"
	GroupNotification = "notification-group"
)

// StreamEvent is the envelope for all platform events.
type StreamEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// PostCreatedPayload is published when a post is created.
type PostCreatedPayload struct {
	PostID    string   `json:"post_id"`
	AuthorID  string   `json:"author_id"`
	Content   string   `json:"content"`
	MediaURLs []string `json:"media_urls"`
}

// PostModeratedPayload is published after moderation completes.
type PostModeratedPayload struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"` // "approved" | "blocked"
}

// PostLikedPayload is published when a post is liked.
type PostLikedPayload struct {
	PostID   string `json:"post_id"`
	ActorID  string `json:"actor_id"`
	AuthorID string `json:"author_id"`
}

// UserFollowedPayload is published when a user follows another.
type UserFollowedPayload struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

// TipSentPayload is published when a tip is sent.
type TipSentPayload struct {
	TipID       string `json:"tip_id"`
	SenderID    string `json:"sender_id"`
	ReceiverID  string `json:"receiver_id"`
	AmountCents int    `json:"amount_cents"`
}

// CommentCreatedPayload is published when a comment is created.
type CommentCreatedPayload struct {
	CommentID     string `json:"comment_id"`
	PostID        string `json:"post_id"`
	CommentableID string `json:"commentable_id"`
	AuthorID      string `json:"author_id"`
	TargetUserID  string `json:"target_user_id"`
}
