package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/infra/moderation"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

// PostService handles post-related business logic
type PostService struct {
	postRepo     post.Repository
	groupRepo    group.Repository
	moderator    moderation.Moderator // may be nil (moderation disabled)
	logger       *zap.Logger
	allowedHosts []string           // OSS media URL whitelist; empty = skip validation
	publisher    eventbus.Publisher // may be nil (events disabled)
}

func NewPostService(postRepo post.Repository, opts ...PostServiceOption) *PostService {
	s := &PostService{postRepo: postRepo}
	for _, o := range opts {
		o(s)
	}
	return s
}

// PostServiceOption configures an optional PostService dependency.
type PostServiceOption func(*PostService)

// WithModerator enables async content moderation.
func WithModerator(m moderation.Moderator, logger *zap.Logger) PostServiceOption {
	return func(s *PostService) {
		s.moderator = m
		s.logger = logger
	}
}

// WithAllowedHosts enables media URL whitelist validation.
func WithAllowedHosts(hosts []string) PostServiceOption {
	return func(s *PostService) {
		s.allowedHosts = hosts
	}
}

// WithPublisher enables event publishing via the business event bus.
func WithPublisher(p eventbus.Publisher) PostServiceOption {
	return func(s *PostService) {
		s.publisher = p
	}
}

// WithGroupRepository enables group-aware post workflows.
func WithGroupRepository(repo group.Repository) PostServiceOption {
	return func(s *PostService) {
		s.groupRepo = repo
	}
}

// CreatePostInput represents input for creating a post
type CreatePostInput struct {
	AuthorID      uuid.UUID
	Title         string
	Content       string
	MediaURLs     []string
	Tags          []string
	ContentLabels map[string]bool
	Visibility    post.Visibility
	GroupID       *uuid.UUID
}

func (s *PostService) CreatePost(ctx context.Context, input CreatePostInput) (*post.Post, error) {
	if input.Content == "" {
		return nil, apperr.BadRequest("帖子内容不能为空")
	}
	if input.Visibility == "" {
		input.Visibility = post.VisibilityPublic
	}

	// Validate media URLs against OSS whitelist
	if len(s.allowedHosts) > 0 {
		for _, mediaURL := range input.MediaURLs {
			u, err := url.Parse(mediaURL)
			if err != nil {
				return nil, apperr.BadRequest("媒体URL格式无效")
			}
			allowed := false
			for _, h := range s.allowedHosts {
				if strings.EqualFold(u.Host, h) {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, apperr.BadRequest("媒体URL不属于允许的存储域名")
			}
		}
	}

	now := time.Now()
	p := &post.Post{
		ID:               uuid.New(),
		AuthorID:         input.AuthorID,
		GroupID:          input.GroupID,
		Title:            input.Title,
		Content:          input.Content,
		MediaURLs:        input.MediaURLs,
		Tags:             input.Tags,
		ContentLabels:    input.ContentLabels,
		Visibility:       input.Visibility,
		ModerationStatus: post.ModerationPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.postRepo.Create(ctx, p); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建帖子失败", err)
	}
	if input.GroupID != nil && s.groupRepo != nil {
		_ = s.groupRepo.IncrementPostCount(ctx, *input.GroupID)
	}

	// If moderation is configured, keep the post in pending state and push it
	// through either the async stream pipeline or the in-process fallback.
	if s.moderator != nil {
		if s.publisher != nil {
			err := s.publisher.Publish(ctx, eventbus.EventPostCreated, eventbus.PostCreatedPayload{
				PostID:    p.ID.String(),
				AuthorID:  p.AuthorID.String(),
				Content:   p.Content,
				MediaURLs: p.MediaURLs,
			})
			if err == nil {
				return p, nil
			}
			if s.logger != nil {
				s.logger.Error("publish post.created failed, falling back to local moderation",
					zap.Error(err),
					zap.String("post_id", p.ID.String()),
				)
			}
		}
		s.moderatePostAsync(p)
	}

	return p, nil
}

func (s *PostService) moderatePostAsync(p *post.Post) {
	if s.moderator == nil {
		return
	}

	postID := p.ID
	content := p.Content
	mediaURLs := append([]string(nil), p.MediaURLs...)
	authorID := p.AuthorID
	moderator := s.moderator
	repo := s.postRepo
	logger := s.logger
	publisher := s.publisher

	go func() {
		defer func() {
			if r := recover(); r != nil && logger != nil {
				logger.Error("moderation goroutine panic", zap.Any("recover", r))
			}
		}()

		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		decision, _, err := moderator.ReviewText(bgCtx, content)
		if err != nil {
			if logger != nil {
				logger.Error("moderation text review failed", zap.Error(err), zap.String("post_id", postID.String()))
			}
			decision = moderation.DecisionPass
		}

		if decision == moderation.DecisionPass && len(mediaURLs) > 0 {
			imgDecision, _, imgErr := moderator.ReviewImage(bgCtx, mediaURLs[0])
			if imgErr != nil {
				if logger != nil {
					logger.Error("moderation image review failed", zap.Error(imgErr), zap.String("post_id", postID.String()))
				}
			} else if imgDecision == moderation.DecisionBlock {
				decision = moderation.DecisionBlock
			}
		}

		status := post.ModerationApproved
		pubStatus := "approved"
		if decision == moderation.DecisionBlock {
			status = post.ModerationBlocked
			pubStatus = "blocked"
		}

		if err := repo.UpdateModerationStatus(bgCtx, postID, status); err != nil {
			if logger != nil {
				logger.Error("failed to update moderation_status", zap.Error(err), zap.String("post_id", postID.String()))
			}
			return
		}

		if publisher != nil {
			if err := publisher.Publish(bgCtx, eventbus.EventPostModerated, eventbus.PostModeratedPayload{
				PostID:   postID.String(),
				AuthorID: authorID.String(),
				Status:   pubStatus,
			}); err != nil && logger != nil {
				logger.Error("publish post.moderated failed", zap.Error(err), zap.String("post_id", postID.String()))
			}
		}
	}()
}

// UpdatePostInput represents input for updating a post
type UpdatePostInput struct {
	Title      string
	Content    string
	MediaURLs  []string
	Tags       []string
	Visibility post.Visibility
}

func (s *PostService) UpdatePost(ctx context.Context, userID, postID uuid.UUID, input UpdatePostInput) (*post.Post, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, post.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	if p.AuthorID != userID {
		return nil, apperr.New(apperr.CodeForbidden, "无权修改此帖子")
	}

	p.Title = input.Title
	p.Content = input.Content
	p.MediaURLs = input.MediaURLs
	p.Tags = input.Tags
	p.Visibility = input.Visibility
	p.UpdatedAt = time.Now()

	if err := s.postRepo.Update(ctx, p); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新帖子失败", err)
	}
	return p, nil
}

func (s *PostService) DeletePost(ctx context.Context, userID, postID uuid.UUID, isAdmin bool) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, post.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}
	if !isAdmin && p.AuthorID != userID {
		return apperr.New(apperr.CodeForbidden, "无权删除此帖子")
	}
	if p.GroupID != nil && s.groupRepo != nil {
		_ = s.groupRepo.DecrementPostCount(ctx, *p.GroupID)
	}
	return s.postRepo.Delete(ctx, postID)
}

func (s *PostService) GetPost(ctx context.Context, postID uuid.UUID) (*post.Post, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, post.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// ListUserPosts lists posts by a specific author
type ListUserPostsInput struct {
	AuthorID   uuid.UUID
	Visibility *post.Visibility
	Page       int
	PageSize   int
}

func (s *PostService) ListUserPosts(ctx context.Context, input ListUserPostsInput) ([]*post.Post, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	return s.postRepo.List(ctx, post.ListFilter{
		AuthorID:   &input.AuthorID,
		Visibility: input.Visibility,
		Page:       input.Page,
		PageSize:   input.PageSize,
	})
}

// ListExplore lists approved public posts for the explore page, ranked by engagement score.
func (s *PostService) ListExplore(ctx context.Context, page, pageSize int, tag string) ([]*post.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	vis := post.VisibilityPublic
	approved := post.ModerationApproved
	filter := post.ListFilter{
		Visibility:       &vis,
		ModerationStatus: &approved,
		SortByScore:      true,
		Page:             page,
		PageSize:         pageSize,
	}
	if tag != "" {
		filter.Tags = []string{tag}
	}
	return s.postRepo.List(ctx, filter)
}

// GetHotTags returns the most used tags
func (s *PostService) GetHotTags(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.postRepo.GetHotTags(ctx, limit)
}

// ListGroupPosts returns approved public posts inside a group ordered by recency.
func (s *PostService) ListGroupPosts(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*post.Post, int64, error) {
	return s.ListGroupPostsWithOptions(ctx, groupID, page, pageSize, "", "latest")
}

// ListGroupPostsWithOptions returns approved public posts inside a group ordered by sort mode and tag.
func (s *PostService) ListGroupPostsWithOptions(ctx context.Context, groupID uuid.UUID, page, pageSize int, tag, sort string) ([]*post.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	vis := post.VisibilityPublic
	approved := post.ModerationApproved
	filter := post.ListFilter{
		GroupID:          &groupID,
		Visibility:       &vis,
		ModerationStatus: &approved,
		Page:             page,
		PageSize:         pageSize,
	}
	if tag != "" {
		filter.Tags = []string{tag}
	}
	if sort == "hot" {
		filter.SortByScore = true
	}
	return s.postRepo.List(ctx, filter)
}

// ListGroupHighlights returns the most engaging approved public posts inside a group.
func (s *PostService) ListGroupHighlights(ctx context.Context, groupID uuid.UUID, limit int) ([]*post.Post, int64, error) {
	if limit < 1 {
		limit = 3
	}
	vis := post.VisibilityPublic
	approved := post.ModerationApproved
	return s.postRepo.List(ctx, post.ListFilter{
		GroupID:          &groupID,
		Visibility:       &vis,
		ModerationStatus: &approved,
		SortByScore:      true,
		Page:             1,
		PageSize:         limit,
	})
}

// GetGroupHotTags returns the most used tags inside a group.
func (s *PostService) GetGroupHotTags(ctx context.Context, groupID uuid.UUID, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 12
	}
	return s.postRepo.GetGroupHotTags(ctx, groupID, limit)
}

// ListFeed returns posts from followed users
func (s *PostService) ListFeed(ctx context.Context, followeeIDs []uuid.UUID, page, pageSize int) ([]*post.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return s.postRepo.ListFeed(ctx, followeeIDs, post.ListFilter{
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *PostService) LikePost(ctx context.Context, userID, postID uuid.UUID) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, post.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}

	hasLiked, err := s.postRepo.HasLiked(ctx, userID, postID)
	if err != nil {
		return err
	}
	if hasLiked {
		return apperr.New(apperr.CodeInvalidParam, "已点赞")
	}

	like := &post.PostLike{
		PostID:    postID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	if err := s.postRepo.LikePost(ctx, like); err != nil {
		if errors.Is(err, post.ErrAlreadyLiked) {
			return apperr.New(apperr.CodeInvalidParam, "已点赞")
		}
		return err
	}
	_ = s.postRepo.IncrementLikeCount(ctx, postID)

	if s.publisher != nil && p.AuthorID != userID {
		go func() {
			_ = s.publisher.Publish(context.Background(), eventbus.EventPostLiked, eventbus.PostLikedPayload{
				PostID:   postID.String(),
				ActorID:  userID.String(),
				AuthorID: p.AuthorID.String(),
			})
		}()
	}
	return nil
}

func (s *PostService) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if err := s.postRepo.UnlikePost(ctx, userID, postID); err != nil {
		if errors.Is(err, post.ErrNotLiked) {
			return apperr.New(apperr.CodeInvalidParam, "未点赞")
		}
		return err
	}
	_ = s.postRepo.DecrementLikeCount(ctx, postID)
	return nil
}

func (s *PostService) PinPost(ctx context.Context, userID, postID uuid.UUID, pin bool) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if p.AuthorID != userID {
		return apperr.New(apperr.CodeForbidden, "无权操作此帖子")
	}
	p.IsPinned = pin
	p.UpdatedAt = time.Now()
	return s.postRepo.Update(ctx, p)
}

// PinGroupPost pins or unpins a post inside a group; only owner/moderator may do this.
func (s *PostService) PinGroupPost(ctx context.Context, actorID, groupID, postID uuid.UUID, pin bool) error {
	if s.groupRepo == nil {
		return apperr.New(apperr.CodeForbidden, "圈子置顶未启用")
	}
	member, err := s.groupRepo.GetMember(ctx, groupID, actorID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "查询圈子成员失败", err)
	}
	if member == nil || (member.Role != group.GroupRoleOwner && member.Role != group.GroupRoleModerator) {
		return apperr.New(apperr.CodeForbidden, "只有圈主或管理员可以置顶帖子")
	}

	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if p.GroupID == nil || *p.GroupID != groupID {
		return apperr.New(apperr.CodeInvalidParam, "该帖子不属于这个圈子")
	}
	p.IsPinned = pin
	p.UpdatedAt = time.Now()
	return s.postRepo.Update(ctx, p)
}

// SearchPosts searches public posts by keyword
func (s *PostService) SearchPosts(ctx context.Context, query string, limit int) ([]*post.Post, error) {
	vis := post.VisibilityPublic
	approved := post.ModerationApproved
	posts, _, err := s.postRepo.List(ctx, post.ListFilter{
		Search:           query,
		Visibility:       &vis,
		ModerationStatus: &approved,
		Page:             1,
		PageSize:         limit,
	})
	return posts, err
}

// AdminListPosts returns paginated posts filtered by moderation status (admin use only).
func (s *PostService) AdminListPosts(ctx context.Context, status string, page, pageSize int) ([]*post.Post, int64, error) {
	filter := post.ListFilter{Page: page, PageSize: pageSize}
	if status != "" {
		ms := post.ModerationStatus(status)
		filter.ModerationStatus = &ms
	}
	return s.postRepo.List(ctx, filter)
}

// AdminUpdateModerationStatus updates a post's moderation_status (admin use only).
func (s *PostService) AdminUpdateModerationStatus(ctx context.Context, postID uuid.UUID, status post.ModerationStatus) error {
	item, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, post.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}

	prevStatus := item.ModerationStatus
	if err := s.postRepo.UpdateModerationStatus(ctx, postID, status); err != nil {
		return err
	}
	if s.publisher != nil && prevStatus != status && (status == post.ModerationApproved || status == post.ModerationBlocked) {
		pubStatus := "blocked"
		if status == post.ModerationApproved {
			pubStatus = "approved"
		}
		if err := s.publisher.Publish(ctx, eventbus.EventPostModerated, eventbus.PostModeratedPayload{
			PostID:   item.ID.String(),
			AuthorID: item.AuthorID.String(),
			Status:   pubStatus,
		}); err != nil && s.logger != nil {
			s.logger.Error("publish admin post.moderated failed", zap.Error(err), zap.String("post_id", item.ID.String()))
		}
	}
	return nil
}
