package eventbus

import (
	"context"

	"github.com/studio/platform/internal/infra/events"
)

type HandlerFunc = events.HandlerFunc

type Event = events.Event

const (
	EventPostCreated     = events.EventPostCreated
	EventPostModerated   = events.EventPostModerated
	EventPostLiked       = events.EventPostLiked
	EventUserFollowed    = events.EventUserFollowed
	EventTipSent         = events.EventTipSent
	EventCommentCreated  = events.EventCommentCreated
	EventAudioJobCreated = events.EventAudioJobCreated

	GroupModeration   = events.GroupModeration
	GroupNotification = events.GroupNotification
	GroupAudioJobs    = events.GroupAudioJobs
)

type PostCreatedPayload = events.PostCreatedPayload
type PostModeratedPayload = events.PostModeratedPayload
type PostLikedPayload = events.PostLikedPayload
type UserFollowedPayload = events.UserFollowedPayload
type TipSentPayload = events.TipSentPayload
type CommentCreatedPayload = events.CommentCreatedPayload
type AudioJobCreatedPayload = events.AudioJobCreatedPayload

type Publisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
	PublishEvent(ctx context.Context, event Event) error
	Close() error
}

type Consumer interface {
	Start(ctx context.Context, group string, handler HandlerFunc) error
	Close() error
}
