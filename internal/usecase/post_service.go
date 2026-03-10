package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/infra/moderation"
	"github.com/studio/platform/internal/pkg/apperr"
	"go.uber.org/zap"
)

// PostService handles post-related business logic
type PostService struct {
	postRepo     post.Repository
	moderator    moderation.Moderator // may be nil (moderation disabled)
	logger       *zap.Logger
	allowedHosts []string // OSS media URL whitelist; empty = skip validation
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

// CreatePostInput represents input for creating a post
type CreatePostInput struct {
	AuthorID      uuid.UUID
	Title         string
	Content       string
	MediaURLs     []string
	Tags          []string
	ContentLabels map[string]bool
	Visibility    post.Visibility
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

	// Trigger async content moderation if moderator is configured.
	if s.moderator != nil {
		postID := p.ID
		content := p.Content
		mediaURLs := p.MediaURLs
		moderator := s.moderator
		repo := s.postRepo
		logger := s.logger
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if logger != nil {
						logger.Error("moderation goroutine panic", zap.Any("recover", r))
					}
				}
			}()
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Check text content
			decision, _, err := moderator.ReviewText(bgCtx, content)
			if err != nil && logger != nil {
				logger.Error("moderation text review failed", zap.Error(err), zap.String("post_id", postID.String()))
			}

			// If text passes, check first image (if any)
			if decision == moderation.DecisionPass && len(mediaURLs) > 0 {
				imgDecision, _, imgErr := moderator.ReviewImage(bgCtx, mediaURLs[0])
				if imgErr != nil && logger != nil {
					logger.Error("moderation image review failed", zap.Error(imgErr), zap.String("post_id", postID.String()))
				}
				if imgDecision == moderation.DecisionBlock {
					decision = moderation.DecisionBlock
				}
			}

			status := post.ModerationApproved
			if decision == moderation.DecisionBlock {
				status = post.ModerationBlocked
			}
			if err := repo.UpdateModerationStatus(bgCtx, postID, status); err != nil && logger != nil {
				logger.Error("failed to update moderation_status", zap.Error(err), zap.String("post_id", postID.String()))
			}
		}()
	} else {
		// No moderator: auto-approve immediately
		p.ModerationStatus = post.ModerationApproved
		if err := s.postRepo.UpdateModerationStatus(ctx, p.ID, post.ModerationApproved); err != nil {
			return nil, apperr.Wrap(apperr.CodeInternalError, "更新帖子状态失败", err)
		}
	}

	return p, nil
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
		return err
	}
	_ = s.postRepo.IncrementLikeCount(ctx, postID)
	return nil
}

func (s *PostService) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if err := s.postRepo.UnlikePost(ctx, userID, postID); err != nil {
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

// SearchPosts searches public posts by keyword
func (s *PostService) SearchPosts(ctx context.Context, query string, limit int) ([]*post.Post, error) {
	vis := post.VisibilityPublic
	posts, _, err := s.postRepo.List(ctx, post.ListFilter{
		Search:     query,
		Visibility: &vis,
		Page:       1,
		PageSize:   limit,
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
	return s.postRepo.UpdateModerationStatus(ctx, postID, status)
}
