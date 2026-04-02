package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/infra/eventbus"
)

type capturedEvent struct {
	eventType string
	payload   json.RawMessage
}

type fakeEventPublisher struct {
	events []capturedEvent
}

func (p *fakeEventPublisher) Publish(_ context.Context, eventType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	p.events = append(p.events, capturedEvent{
		eventType: eventType,
		payload:   raw,
	})
	return nil
}

func (p *fakeEventPublisher) PublishEvent(ctx context.Context, event eventbus.Event) error {
	return p.Publish(ctx, event.Type, json.RawMessage(event.Payload))
}

func (p *fakeEventPublisher) Close() error {
	return nil
}

type fakeModerationPostRepo struct {
	items map[uuid.UUID]*post.Post
}

func newFakeModerationPostRepo() *fakeModerationPostRepo {
	return &fakeModerationPostRepo{
		items: make(map[uuid.UUID]*post.Post),
	}
}

func (r *fakeModerationPostRepo) seed(item *post.Post) {
	clone := *item
	r.items[item.ID] = &clone
}

func (r *fakeModerationPostRepo) Create(_ context.Context, p *post.Post) error {
	clone := *p
	r.items[p.ID] = &clone
	return nil
}

func (r *fakeModerationPostRepo) GetByID(_ context.Context, id uuid.UUID) (*post.Post, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, post.ErrNotFound
	}
	clone := *item
	return &clone, nil
}

func (r *fakeModerationPostRepo) Update(_ context.Context, p *post.Post) error {
	if _, ok := r.items[p.ID]; !ok {
		return post.ErrNotFound
	}
	clone := *p
	r.items[p.ID] = &clone
	return nil
}

func (r *fakeModerationPostRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.items, id)
	return nil
}

func (r *fakeModerationPostRepo) List(_ context.Context, _ post.ListFilter) ([]*post.Post, int64, error) {
	return nil, 0, nil
}

func (r *fakeModerationPostRepo) ListFeed(_ context.Context, _ []uuid.UUID, _ post.ListFilter) ([]*post.Post, int64, error) {
	return nil, 0, nil
}

func (r *fakeModerationPostRepo) GetHotTags(_ context.Context, _ int) ([]string, error) {
	return nil, nil
}

func (r *fakeModerationPostRepo) GetGroupHotTags(_ context.Context, _ uuid.UUID, _ int) ([]string, error) {
	return nil, nil
}

func (r *fakeModerationPostRepo) LikePost(_ context.Context, _ *post.PostLike) error {
	return nil
}

func (r *fakeModerationPostRepo) UnlikePost(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (r *fakeModerationPostRepo) HasLiked(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeModerationPostRepo) IncrementLikeCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *fakeModerationPostRepo) DecrementLikeCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *fakeModerationPostRepo) IncrementCommentCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *fakeModerationPostRepo) DecrementCommentCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *fakeModerationPostRepo) UpdateModerationStatus(_ context.Context, id uuid.UUID, status post.ModerationStatus) error {
	item, ok := r.items[id]
	if !ok {
		return errors.New("post not found")
	}
	item.ModerationStatus = status
	return nil
}

func TestCreatePostStaysPendingWithoutModerator(t *testing.T) {
	ctx := context.Background()
	repo := newFakeModerationPostRepo()
	publisher := &fakeEventPublisher{}
	svc := NewPostService(repo, WithPublisher(publisher))

	item, err := svc.CreatePost(ctx, CreatePostInput{
		AuthorID:   uuid.New(),
		Content:    "需要人工审核的帖子",
		Visibility: post.VisibilityPublic,
	})
	if err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}
	if item.ModerationStatus != post.ModerationPending {
		t.Fatalf("moderation status = %q, want pending", item.ModerationStatus)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected no moderation event without moderator, got %d", len(publisher.events))
	}
}

func TestAdminUpdateModerationStatusPublishesEvent(t *testing.T) {
	ctx := context.Background()
	repo := newFakeModerationPostRepo()
	publisher := &fakeEventPublisher{}
	svc := NewPostService(repo, WithPublisher(publisher))

	item := &post.Post{
		ID:               uuid.New(),
		AuthorID:         uuid.New(),
		Content:          "等待审核",
		Visibility:       post.VisibilityPublic,
		ModerationStatus: post.ModerationPending,
	}
	repo.seed(item)

	if err := svc.AdminUpdateModerationStatus(ctx, item.ID, post.ModerationApproved); err != nil {
		t.Fatalf("AdminUpdateModerationStatus error: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected one event, got %d", len(publisher.events))
	}
	if publisher.events[0].eventType != eventbus.EventPostModerated {
		t.Fatalf("event type = %q, want %q", publisher.events[0].eventType, eventbus.EventPostModerated)
	}

	var payload eventbus.PostModeratedPayload
	if err := json.Unmarshal(publisher.events[0].payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.PostID != item.ID.String() {
		t.Fatalf("payload post_id = %q, want %q", payload.PostID, item.ID.String())
	}
	if payload.AuthorID != item.AuthorID.String() {
		t.Fatalf("payload author_id = %q, want %q", payload.AuthorID, item.AuthorID.String())
	}
	if payload.Status != "approved" {
		t.Fatalf("payload status = %q, want approved", payload.Status)
	}
}
