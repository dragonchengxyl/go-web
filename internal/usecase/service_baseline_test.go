package usecase_test

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/chat"
	"github.com/studio/platform/internal/domain/follow"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/domain/user"
	redisinfra "github.com/studio/platform/internal/infra/redis"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
)

func TestUserServiceRegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	tokenStore := newTestTokenStore(t)
	repo := newFakeUserRepo()

	svc := usecase.NewUserService(repo, tokenStore, configs.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Minute,
		RefreshExpiry: time.Hour,
	})

	registerOut, err := svc.Register(ctx, usecase.RegisterInput{
		Username: "fox_dev",
		Email:    "fox@example.com",
		Password: "wolf1234",
		IP:       "127.0.0.1",
		Device:   "integration-test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerOut.Tokens.AccessToken)
	require.NotEmpty(t, registerOut.Tokens.RefreshToken)

	storedUser, err := repo.GetByEmail(ctx, "fox@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, "wolf1234", storedUser.PasswordHash)

	loginOut, err := svc.Login(ctx, usecase.LoginInput{
		Email:    "fox@example.com",
		Password: "wolf1234",
		IP:       "127.0.0.2",
		Device:   "integration-test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginOut.Tokens.AccessToken)
	require.NotEmpty(t, loginOut.Tokens.RefreshToken)

	storedUser, err = repo.GetByEmail(ctx, "fox@example.com")
	require.NoError(t, err)
	require.NotNil(t, storedUser.LastLoginIP)
	assert.Equal(t, "127.0.0.2", *storedUser.LastLoginIP)
}

func TestUserServiceRegisterRejectsDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	tokenStore := newTestTokenStore(t)
	repo := newFakeUserRepo()

	svc := usecase.NewUserService(repo, tokenStore, configs.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Minute,
		RefreshExpiry: time.Hour,
	})

	_, err := svc.Register(ctx, usecase.RegisterInput{
		Username: "first_user",
		Email:    "dup@example.com",
		Password: "wolf1234",
	})
	require.NoError(t, err)

	_, err = svc.Register(ctx, usecase.RegisterInput{
		Username: "second_user",
		Email:    "dup@example.com",
		Password: "wolf1234",
	})
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeEmailExists, appErr.Code)
}

func TestPostServiceCreatePostRejectsUntrustedMediaHost(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostRepo()
	svc := usecase.NewPostService(repo, usecase.WithAllowedHosts([]string{"cdn.example.com"}))

	_, err := svc.CreatePost(ctx, usecase.CreatePostInput{
		AuthorID:   uuid.New(),
		Content:    "hello world",
		MediaURLs:  []string{"https://evil.example.com/cat.png"},
		Visibility: post.VisibilityPublic,
	})
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidParam, appErr.Code)
}

func TestFollowAndFeedFlow(t *testing.T) {
	ctx := context.Background()
	followRepo := newFakeFollowRepo()
	postRepo := newFakePostRepo()

	viewerID := uuid.New()
	followedAuthorID := uuid.New()
	otherAuthorID := uuid.New()

	postRepo.seed(&post.Post{
		ID:               uuid.New(),
		AuthorID:         followedAuthorID,
		Content:          "followed author post",
		Visibility:       post.VisibilityPublic,
		ModerationStatus: post.ModerationApproved,
		CreatedAt:        time.Now().Add(-time.Minute),
		UpdatedAt:        time.Now().Add(-time.Minute),
	})
	postRepo.seed(&post.Post{
		ID:               uuid.New(),
		AuthorID:         otherAuthorID,
		Content:          "other author post",
		Visibility:       post.VisibilityPublic,
		ModerationStatus: post.ModerationApproved,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})

	followSvc := usecase.NewFollowService(followRepo)
	require.NoError(t, followSvc.Follow(ctx, viewerID, followedAuthorID))

	followeeIDs, err := followSvc.GetFollowingIDs(ctx, viewerID)
	require.NoError(t, err)
	require.Len(t, followeeIDs, 1)
	assert.Equal(t, followedAuthorID, followeeIDs[0])

	postSvc := usecase.NewPostService(postRepo)
	feed, total, err := postSvc.ListFeed(ctx, followeeIDs, 1, 20)
	require.NoError(t, err)
	require.Len(t, feed, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, followedAuthorID, feed[0].AuthorID)
}

func TestGroupServiceUpdateGroupCreatesAnnouncement(t *testing.T) {
	ctx := context.Background()
	repo := newFakeGroupRepo()
	svc := usecase.NewGroupService(repo)

	ownerID := uuid.New()
	created, err := svc.CreateGroup(ctx, usecase.CreateGroupInput{
		OwnerID:     ownerID,
		Name:        "测试圈子",
		Description: "desc",
	})
	require.NoError(t, err)

	updated, err := svc.UpdateGroup(ctx, ownerID, created.ID, usecase.UpdateGroupInput{
		Announcement: "新的公告内容",
	})
	require.NoError(t, err)
	assert.Equal(t, "新的公告内容", updated.Announcement)

	items, total, err := svc.ListAnnouncements(ctx, created.ID, 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "新的公告内容", items[0].Content)
}

func TestNotificationServiceNotifyPersistsAndPushes(t *testing.T) {
	ctx := context.Background()
	repo := newFakeNotificationRepo()
	hub := &fakeHub{}
	svc := usecase.NewNotificationService(repo, hub)

	userID := uuid.New()
	err := svc.Notify(ctx, &notification.Notification{
		UserID:     userID,
		Type:       notification.TypeFollow,
		TargetType: "user",
	})
	require.NoError(t, err)

	items, total, err := svc.ListNotifications(ctx, usecase.ListNotificationsInput{
		UserID:   userID,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
	assert.False(t, items[0].IsRead)

	require.Len(t, hub.sent, 1)
	assert.Equal(t, userID, hub.sent[0].userID)
	assert.Equal(t, ws.MessageTypeNotification, hub.sent[0].msg.Type)
}

func TestChatServiceCreateConversationAndSendMessage(t *testing.T) {
	ctx := context.Background()
	repo := newFakeChatRepo()
	svc := usecase.NewChatService(repo)

	userA := uuid.New()
	userB := uuid.New()

	conversation, err := svc.CreateDirectConversation(ctx, userA, userB)
	require.NoError(t, err)
	require.Len(t, conversation.Members, 2)

	sameConversation, err := svc.CreateDirectConversation(ctx, userA, userB)
	require.NoError(t, err)
	assert.Equal(t, conversation.ID, sameConversation.ID)

	msg, err := svc.SendMessage(ctx, usecase.SendMessageInput{
		ConversationID: conversation.ID,
		SenderID:       userA,
		Content:        "你好，世界",
	})
	require.NoError(t, err)
	assert.Equal(t, "你好，世界", msg.Content)

	messages, total, err := svc.ListMessages(ctx, conversation.ID, userB, 1, 20)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, msg.ID, messages[0].ID)
}

func newTestTokenStore(t *testing.T) *redisinfra.TokenStore {
	t.Helper()

	addr := os.Getenv("STUDIO_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redisclient.NewClient(&redisclient.Options{
		Addr: addr,
		DB:   15,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available for auth integration test: %v", err)
	}
	require.NoError(t, client.FlushDB(ctx).Err())
	t.Cleanup(func() {
		_ = client.FlushDB(ctx).Err()
		_ = client.Close()
	})

	return redisinfra.NewTokenStore(client)
}

type fakeUserRepo struct {
	byID       map[uuid.UUID]*user.User
	byEmail    map[string]uuid.UUID
	byUsername map[string]uuid.UUID
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:       make(map[uuid.UUID]*user.User),
		byEmail:    make(map[string]uuid.UUID),
		byUsername: make(map[string]uuid.UUID),
	}
}

func (r *fakeUserRepo) Create(_ context.Context, entity *user.User) error {
	if _, exists := r.byEmail[entity.Email]; exists {
		return user.ErrEmailExists
	}
	if _, exists := r.byUsername[entity.Username]; exists {
		return user.ErrUsernameExists
	}

	cloned := *entity
	r.byID[entity.ID] = &cloned
	r.byEmail[entity.Email] = entity.ID
	r.byUsername[entity.Username] = entity.ID
	return nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	entity, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cloned := *entity
	return &cloned, nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*user.User, error) {
	id, ok := r.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return r.GetByID(context.Background(), id)
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*user.User, error) {
	id, ok := r.byUsername[username]
	if !ok {
		return nil, user.ErrNotFound
	}
	return r.GetByID(context.Background(), id)
}

func (r *fakeUserRepo) Update(_ context.Context, entity *user.User) error {
	if _, ok := r.byID[entity.ID]; !ok {
		return user.ErrNotFound
	}
	cloned := *entity
	r.byID[entity.ID] = &cloned
	r.byEmail[entity.Email] = entity.ID
	r.byUsername[entity.Username] = entity.ID
	return nil
}

func (r *fakeUserRepo) UpdateLastLogin(_ context.Context, id uuid.UUID, ip string) error {
	entity, ok := r.byID[id]
	if !ok {
		return user.ErrNotFound
	}
	now := time.Now()
	entity.LastLoginAt = &now
	entity.LastLoginIP = &ip
	entity.UpdatedAt = now
	return nil
}

func (r *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, exists := r.byEmail[email]
	return exists, nil
}

func (r *fakeUserRepo) ExistsByUsername(_ context.Context, username string) (bool, error) {
	_, exists := r.byUsername[username]
	return exists, nil
}

func (r *fakeUserRepo) List(_ context.Context, _ user.ListFilter) ([]*user.User, int64, error) {
	items := make([]*user.User, 0, len(r.byID))
	for _, entity := range r.byID {
		cloned := *entity
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	entity, ok := r.byID[id]
	if !ok {
		return user.ErrNotFound
	}
	delete(r.byEmail, entity.Email)
	delete(r.byUsername, entity.Username)
	delete(r.byID, id)
	return nil
}

type fakePostRepo struct {
	posts map[uuid.UUID]*post.Post
	likes map[string]struct{}
}

func newFakePostRepo() *fakePostRepo {
	return &fakePostRepo{
		posts: make(map[uuid.UUID]*post.Post),
		likes: make(map[string]struct{}),
	}
}

func (r *fakePostRepo) seed(item *post.Post) {
	cloned := clonePost(item)
	r.posts[cloned.ID] = cloned
}

func (r *fakePostRepo) Create(_ context.Context, item *post.Post) error {
	r.posts[item.ID] = clonePost(item)
	return nil
}

func (r *fakePostRepo) GetByID(_ context.Context, id uuid.UUID) (*post.Post, error) {
	item, ok := r.posts[id]
	if !ok {
		return nil, post.ErrNotFound
	}
	return clonePost(item), nil
}

func (r *fakePostRepo) Update(_ context.Context, item *post.Post) error {
	if _, ok := r.posts[item.ID]; !ok {
		return post.ErrNotFound
	}
	r.posts[item.ID] = clonePost(item)
	return nil
}

func (r *fakePostRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.posts, id)
	return nil
}

func (r *fakePostRepo) List(_ context.Context, filter post.ListFilter) ([]*post.Post, int64, error) {
	items := make([]*post.Post, 0, len(r.posts))
	for _, item := range r.posts {
		if filter.AuthorID != nil && item.AuthorID != *filter.AuthorID {
			continue
		}
		if filter.GroupID != nil {
			if item.GroupID == nil || *item.GroupID != *filter.GroupID {
				continue
			}
		}
		if filter.Visibility != nil && item.Visibility != *filter.Visibility {
			continue
		}
		if filter.ModerationStatus != nil && item.ModerationStatus != *filter.ModerationStatus {
			continue
		}
		if filter.Search != "" && !strings.Contains(item.Content, filter.Search) && !strings.Contains(item.Title, filter.Search) {
			continue
		}
		items = append(items, clonePost(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return paginatePosts(items, filter.Page, filter.PageSize)
}

func (r *fakePostRepo) ListFeed(_ context.Context, followeeIDs []uuid.UUID, filter post.ListFilter) ([]*post.Post, int64, error) {
	allowed := make(map[uuid.UUID]struct{}, len(followeeIDs))
	for _, id := range followeeIDs {
		allowed[id] = struct{}{}
	}

	items := make([]*post.Post, 0)
	for _, item := range r.posts {
		if _, ok := allowed[item.AuthorID]; !ok {
			continue
		}
		items = append(items, clonePost(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return paginatePosts(items, filter.Page, filter.PageSize)
}

func (r *fakePostRepo) GetHotTags(_ context.Context, _ int) ([]string, error) {
	return nil, nil
}

func (r *fakePostRepo) GetGroupHotTags(_ context.Context, _ uuid.UUID, _ int) ([]string, error) {
	return nil, nil
}

func (r *fakePostRepo) LikePost(_ context.Context, like *post.PostLike) error {
	r.likes[like.UserID.String()+":"+like.PostID.String()] = struct{}{}
	return nil
}

func (r *fakePostRepo) UnlikePost(_ context.Context, userID, postID uuid.UUID) error {
	delete(r.likes, userID.String()+":"+postID.String())
	return nil
}

func (r *fakePostRepo) HasLiked(_ context.Context, userID, postID uuid.UUID) (bool, error) {
	_, ok := r.likes[userID.String()+":"+postID.String()]
	return ok, nil
}

func (r *fakePostRepo) IncrementLikeCount(_ context.Context, postID uuid.UUID) error {
	r.posts[postID].LikeCount++
	return nil
}

func (r *fakePostRepo) DecrementLikeCount(_ context.Context, postID uuid.UUID) error {
	r.posts[postID].LikeCount--
	return nil
}

func (r *fakePostRepo) IncrementCommentCount(_ context.Context, postID uuid.UUID) error {
	r.posts[postID].CommentCount++
	return nil
}

func (r *fakePostRepo) DecrementCommentCount(_ context.Context, postID uuid.UUID) error {
	r.posts[postID].CommentCount--
	return nil
}

func (r *fakePostRepo) UpdateModerationStatus(_ context.Context, id uuid.UUID, status post.ModerationStatus) error {
	item, ok := r.posts[id]
	if !ok {
		return post.ErrNotFound
	}
	item.ModerationStatus = status
	return nil
}

type fakeFollowRepo struct {
	following map[uuid.UUID]map[uuid.UUID]*follow.UserFollow
}

func newFakeFollowRepo() *fakeFollowRepo {
	return &fakeFollowRepo{
		following: make(map[uuid.UUID]map[uuid.UUID]*follow.UserFollow),
	}
}

func (r *fakeFollowRepo) Follow(_ context.Context, entity *follow.UserFollow) error {
	if r.following[entity.FollowerID] == nil {
		r.following[entity.FollowerID] = make(map[uuid.UUID]*follow.UserFollow)
	}
	cloned := *entity
	r.following[entity.FollowerID][entity.FolloweeID] = &cloned
	return nil
}

func (r *fakeFollowRepo) Unfollow(_ context.Context, followerID, followeeID uuid.UUID) error {
	if r.following[followerID] != nil {
		delete(r.following[followerID], followeeID)
	}
	return nil
}

func (r *fakeFollowRepo) IsFollowing(_ context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	if r.following[followerID] == nil {
		return false, nil
	}
	_, ok := r.following[followerID][followeeID]
	return ok, nil
}

func (r *fakeFollowRepo) ListFollowers(_ context.Context, userID uuid.UUID, _, _ int) ([]*follow.UserFollow, int64, error) {
	items := make([]*follow.UserFollow, 0)
	for followerID, followees := range r.following {
		if item, ok := followees[userID]; ok {
			cloned := *item
			cloned.FollowerID = followerID
			items = append(items, &cloned)
		}
	}
	return items, int64(len(items)), nil
}

func (r *fakeFollowRepo) ListFollowing(_ context.Context, userID uuid.UUID, _, _ int) ([]*follow.UserFollow, int64, error) {
	items := make([]*follow.UserFollow, 0)
	for _, item := range r.following[userID] {
		cloned := *item
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeFollowRepo) GetFollowingIDs(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	items := make([]uuid.UUID, 0, len(r.following[userID]))
	for followeeID := range r.following[userID] {
		items = append(items, followeeID)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].String() < items[j].String()
	})
	return items, nil
}

func (r *fakeFollowRepo) GetStats(_ context.Context, userID uuid.UUID) (*follow.FollowStats, error) {
	followerCount := int64(0)
	for _, followees := range r.following {
		if _, ok := followees[userID]; ok {
			followerCount++
		}
	}
	return &follow.FollowStats{
		UserID:         userID,
		FollowerCount:  followerCount,
		FollowingCount: int64(len(r.following[userID])),
	}, nil
}

type fakeGroupRepo struct {
	groups        map[uuid.UUID]*group.Group
	members       map[uuid.UUID]map[uuid.UUID]*group.GroupMember
	announcements map[uuid.UUID][]*group.GroupAnnouncement
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{
		groups:        make(map[uuid.UUID]*group.Group),
		members:       make(map[uuid.UUID]map[uuid.UUID]*group.GroupMember),
		announcements: make(map[uuid.UUID][]*group.GroupAnnouncement),
	}
}

func (r *fakeGroupRepo) Create(_ context.Context, item *group.Group) error {
	cloned := *item
	r.groups[item.ID] = &cloned
	return nil
}

func (r *fakeGroupRepo) GetByID(_ context.Context, id uuid.UUID) (*group.Group, error) {
	item, ok := r.groups[id]
	if !ok {
		return nil, group.ErrNotFound
	}
	cloned := *item
	return &cloned, nil
}

func (r *fakeGroupRepo) Update(_ context.Context, item *group.Group) error {
	if _, ok := r.groups[item.ID]; !ok {
		return group.ErrNotFound
	}
	cloned := *item
	r.groups[item.ID] = &cloned
	return nil
}

func (r *fakeGroupRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.groups, id)
	delete(r.members, id)
	delete(r.announcements, id)
	return nil
}

func (r *fakeGroupRepo) List(_ context.Context, filter group.ListFilter) ([]*group.Group, int64, error) {
	items := make([]*group.Group, 0)
	for _, item := range r.groups {
		if filter.Privacy != nil && item.Privacy != *filter.Privacy {
			continue
		}
		if filter.Search != "" && !strings.Contains(item.Name, filter.Search) && !strings.Contains(item.Description, filter.Search) {
			continue
		}
		cloned := *item
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeGroupRepo) AddMember(_ context.Context, member *group.GroupMember) error {
	if r.members[member.GroupID] == nil {
		r.members[member.GroupID] = make(map[uuid.UUID]*group.GroupMember)
	}
	cloned := *member
	r.members[member.GroupID][member.UserID] = &cloned
	return nil
}

func (r *fakeGroupRepo) RemoveMember(_ context.Context, groupID, userID uuid.UUID) error {
	if r.members[groupID] != nil {
		delete(r.members[groupID], userID)
	}
	return nil
}

func (r *fakeGroupRepo) GetMember(_ context.Context, groupID, userID uuid.UUID) (*group.GroupMember, error) {
	if r.members[groupID] == nil {
		return nil, nil
	}
	member, ok := r.members[groupID][userID]
	if !ok {
		return nil, nil
	}
	cloned := *member
	return &cloned, nil
}

func (r *fakeGroupRepo) UpdateMemberRole(_ context.Context, groupID, userID uuid.UUID, role group.GroupRole) error {
	if r.members[groupID] == nil || r.members[groupID][userID] == nil {
		return group.ErrNotMember
	}
	r.members[groupID][userID].Role = role
	return nil
}

func (r *fakeGroupRepo) ListMembers(_ context.Context, groupID uuid.UUID, _, _ int) ([]*group.GroupMember, int64, error) {
	items := make([]*group.GroupMember, 0, len(r.members[groupID]))
	for _, member := range r.members[groupID] {
		cloned := *member
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeGroupRepo) CreateAnnouncement(_ context.Context, item *group.GroupAnnouncement) error {
	cloned := *item
	r.announcements[item.GroupID] = append(r.announcements[item.GroupID], &cloned)
	return nil
}

func (r *fakeGroupRepo) ListAnnouncements(_ context.Context, groupID uuid.UUID, _, _ int) ([]*group.GroupAnnouncement, int64, error) {
	items := make([]*group.GroupAnnouncement, 0, len(r.announcements[groupID]))
	for _, item := range r.announcements[groupID] {
		cloned := *item
		items = append(items, &cloned)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, int64(len(items)), nil
}

func (r *fakeGroupRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, _, _ int) ([]*group.Group, int64, error) {
	items := make([]*group.Group, 0)
	for _, item := range r.groups {
		if item.OwnerID != ownerID {
			continue
		}
		cloned := *item
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeGroupRepo) ListByRole(_ context.Context, userID uuid.UUID, role group.GroupRole, _, _ int) ([]*group.Group, int64, error) {
	items := make([]*group.Group, 0)
	for groupID, members := range r.members {
		member := members[userID]
		if member == nil || member.Role != role {
			continue
		}
		cloned := *r.groups[groupID]
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeGroupRepo) ListRecentActiveMembers(_ context.Context, groupID uuid.UUID, limit int) ([]*group.GroupMember, error) {
	items := make([]*group.GroupMember, 0, len(r.members[groupID]))
	for _, member := range r.members[groupID] {
		cloned := *member
		items = append(items, &cloned)
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (r *fakeGroupRepo) IncrementMemberCount(_ context.Context, groupID uuid.UUID) error {
	r.groups[groupID].MemberCount++
	return nil
}

func (r *fakeGroupRepo) DecrementMemberCount(_ context.Context, groupID uuid.UUID) error {
	r.groups[groupID].MemberCount--
	return nil
}

func (r *fakeGroupRepo) IncrementPostCount(_ context.Context, groupID uuid.UUID) error {
	r.groups[groupID].PostCount++
	return nil
}

func (r *fakeGroupRepo) DecrementPostCount(_ context.Context, groupID uuid.UUID) error {
	r.groups[groupID].PostCount--
	return nil
}

type fakeNotificationRepo struct {
	items map[uuid.UUID][]*notification.Notification
}

func newFakeNotificationRepo() *fakeNotificationRepo {
	return &fakeNotificationRepo{
		items: make(map[uuid.UUID][]*notification.Notification),
	}
}

func (r *fakeNotificationRepo) Create(_ context.Context, item *notification.Notification) error {
	cloned := *item
	r.items[item.UserID] = append(r.items[item.UserID], &cloned)
	return nil
}

func (r *fakeNotificationRepo) List(_ context.Context, userID uuid.UUID, _, _ int) ([]*notification.Notification, int64, error) {
	items := make([]*notification.Notification, 0, len(r.items[userID]))
	for _, item := range r.items[userID] {
		cloned := *item
		items = append(items, &cloned)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, int64(len(items)), nil
}

func (r *fakeNotificationRepo) MarkRead(_ context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	targets := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		targets[id] = struct{}{}
	}
	for _, item := range r.items[userID] {
		if _, ok := targets[item.ID]; ok {
			item.IsRead = true
		}
	}
	return nil
}

func (r *fakeNotificationRepo) MarkAllRead(_ context.Context, userID uuid.UUID) error {
	for _, item := range r.items[userID] {
		item.IsRead = true
	}
	return nil
}

func (r *fakeNotificationRepo) CountUnread(_ context.Context, userID uuid.UUID) (int64, error) {
	count := int64(0)
	for _, item := range r.items[userID] {
		if !item.IsRead {
			count++
		}
	}
	return count, nil
}

type fakeChatRepo struct {
	conversations map[uuid.UUID]*chat.Conversation
	directIndex   map[string]uuid.UUID
	members       map[uuid.UUID]map[uuid.UUID]struct{}
	messages      map[uuid.UUID][]*chat.Message
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{
		conversations: make(map[uuid.UUID]*chat.Conversation),
		directIndex:   make(map[string]uuid.UUID),
		members:       make(map[uuid.UUID]map[uuid.UUID]struct{}),
		messages:      make(map[uuid.UUID][]*chat.Message),
	}
}

func (r *fakeChatRepo) CreateConversation(_ context.Context, item *chat.Conversation) error {
	cloned := *item
	cloned.Members = append([]uuid.UUID(nil), item.Members...)
	r.conversations[item.ID] = &cloned
	if item.Type == chat.ConversationTypeDirect && len(item.Members) == 2 {
		r.directIndex[directConversationKey(item.Members[0], item.Members[1])] = item.ID
	}
	if r.members[item.ID] == nil {
		r.members[item.ID] = make(map[uuid.UUID]struct{})
	}
	for _, memberID := range item.Members {
		r.members[item.ID][memberID] = struct{}{}
	}
	return nil
}

func (r *fakeChatRepo) GetConversationByID(_ context.Context, id uuid.UUID) (*chat.Conversation, error) {
	item, ok := r.conversations[id]
	if !ok {
		return nil, chat.ErrNotFound
	}
	cloned := *item
	cloned.Members = append([]uuid.UUID(nil), item.Members...)
	return &cloned, nil
}

func (r *fakeChatRepo) GetDirectConversation(_ context.Context, userA, userB uuid.UUID) (*chat.Conversation, error) {
	id, ok := r.directIndex[directConversationKey(userA, userB)]
	if !ok {
		return nil, nil
	}
	return r.GetConversationByID(context.Background(), id)
}

func (r *fakeChatRepo) ListConversations(_ context.Context, userID uuid.UUID, _, _ int) ([]*chat.Conversation, int64, error) {
	items := make([]*chat.Conversation, 0)
	for conversationID, members := range r.members {
		if _, ok := members[userID]; !ok {
			continue
		}
		cloned := *r.conversations[conversationID]
		cloned.Members = append([]uuid.UUID(nil), r.conversations[conversationID].Members...)
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

func (r *fakeChatRepo) IsMember(_ context.Context, conversationID, userID uuid.UUID) (bool, error) {
	if r.members[conversationID] == nil {
		return false, nil
	}
	_, ok := r.members[conversationID][userID]
	return ok, nil
}

func (r *fakeChatRepo) AddMember(_ context.Context, member *chat.ConversationMember) error {
	if r.members[member.ConversationID] == nil {
		r.members[member.ConversationID] = make(map[uuid.UUID]struct{})
	}
	r.members[member.ConversationID][member.UserID] = struct{}{}
	return nil
}

func (r *fakeChatRepo) CreateMessage(_ context.Context, item *chat.Message) error {
	cloned := *item
	r.messages[item.ConversationID] = append(r.messages[item.ConversationID], &cloned)
	return nil
}

func (r *fakeChatRepo) GetMessageByID(_ context.Context, id uuid.UUID) (*chat.Message, error) {
	for _, items := range r.messages {
		for _, item := range items {
			if item.ID == id {
				cloned := *item
				return &cloned, nil
			}
		}
	}
	return nil, errors.New("message not found")
}

func (r *fakeChatRepo) ListMessages(_ context.Context, conversationID uuid.UUID, page, pageSize int) ([]*chat.Message, int64, error) {
	items := make([]*chat.Message, 0, len(r.messages[conversationID]))
	for _, item := range r.messages[conversationID] {
		cloned := *item
		items = append(items, &cloned)
	}
	total := int64(len(items))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = len(items)
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []*chat.Message{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (r *fakeChatRepo) MarkRead(_ context.Context, conversationID, _ uuid.UUID) error {
	for _, item := range r.messages[conversationID] {
		item.IsRead = true
	}
	return nil
}

type sentWSMessage struct {
	userID uuid.UUID
	msg    ws.WSMessage
}

type fakeHub struct {
	sent []sentWSMessage
}

func (h *fakeHub) SendToUser(userID uuid.UUID, msg ws.WSMessage) {
	h.sent = append(h.sent, sentWSMessage{userID: userID, msg: msg})
}

func (h *fakeHub) Register(_ *ws.Client)     {}
func (h *fakeHub) Unregister(_ *ws.Client)   {}
func (h *fakeHub) Run(_ context.Context)     {}
func (h *fakeHub) ConnCount(_ uuid.UUID) int { return 0 }

func paginatePosts(items []*post.Post, page, pageSize int) ([]*post.Post, int64, error) {
	total := int64(len(items))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = len(items)
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []*post.Post{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func clonePost(item *post.Post) *post.Post {
	cloned := *item
	cloned.MediaURLs = append([]string(nil), item.MediaURLs...)
	cloned.Tags = append([]string(nil), item.Tags...)
	if item.ContentLabels != nil {
		cloned.ContentLabels = make(map[string]bool, len(item.ContentLabels))
		for key, value := range item.ContentLabels {
			cloned.ContentLabels[key] = value
		}
	}
	return &cloned
}

func directConversationKey(userA, userB uuid.UUID) string {
	ids := []string{userA.String(), userB.String()}
	sort.Strings(ids)
	return ids[0] + ":" + ids[1]
}

var _ user.Repository = (*fakeUserRepo)(nil)
var _ post.Repository = (*fakePostRepo)(nil)
var _ follow.Repository = (*fakeFollowRepo)(nil)
var _ group.Repository = (*fakeGroupRepo)(nil)
var _ notification.Repository = (*fakeNotificationRepo)(nil)
var _ chat.Repository = (*fakeChatRepo)(nil)
var _ ws.HubInterface = (*fakeHub)(nil)
