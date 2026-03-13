package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/bookmark"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/pkg/apperr"
)

type BookmarkService struct {
	repo     bookmark.Repository
	postSvc  *PostService
	groupSvc *GroupService
	eventSvc *EventService
}

func NewBookmarkService(repo bookmark.Repository, postSvc *PostService, groupSvc *GroupService, eventSvc *EventService) *BookmarkService {
	return &BookmarkService{
		repo:     repo,
		postSvc:  postSvc,
		groupSvc: groupSvc,
		eventSvc: eventSvc,
	}
}

func (s *BookmarkService) Add(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetID uuid.UUID) error {
	if err := s.ensureTargetExists(ctx, targetType, targetID); err != nil {
		return err
	}
	return s.repo.Create(ctx, &bookmark.Bookmark{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  time.Now(),
	})
}

func (s *BookmarkService) Remove(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, targetType, targetID)
}

func (s *BookmarkService) Exists(ctx context.Context, userID uuid.UUID, targetType bookmark.TargetType, targetID uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, userID, targetType, targetID)
}

func (s *BookmarkService) MarkPosts(ctx context.Context, userID uuid.UUID, posts []*post.Post) error {
	for _, item := range posts {
		exists, err := s.repo.Exists(ctx, userID, bookmark.TargetPost, item.ID)
		if err != nil {
			return apperr.Wrap(apperr.CodeInternalError, "查询帖子收藏状态失败", err)
		}
		item.IsBookmarkedByMe = exists
	}
	return nil
}

func (s *BookmarkService) ListPosts(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*post.Post, int64, error) {
	items, total, err := s.repo.List(ctx, userID, bookmark.TargetPost, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取收藏帖子失败", err)
	}
	posts := make([]*post.Post, 0, len(items))
	for _, item := range items {
		p, err := s.postSvc.GetPost(ctx, item.TargetID)
		if err != nil {
			continue
		}
		p.IsBookmarkedByMe = true
		posts = append(posts, p)
	}
	return posts, total, nil
}

func (s *BookmarkService) ListGroups(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error) {
	items, total, err := s.repo.List(ctx, userID, bookmark.TargetGroup, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取收藏圈子失败", err)
	}
	groups := make([]*group.Group, 0, len(items))
	for _, item := range items {
		g, err := s.groupSvc.GetGroup(ctx, item.TargetID)
		if err != nil {
			continue
		}
		groups = append(groups, g)
	}
	return groups, total, nil
}

func (s *BookmarkService) ListEvents(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*event.Event, int64, error) {
	items, total, err := s.repo.List(ctx, userID, bookmark.TargetEvent, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "读取收藏活动失败", err)
	}
	events := make([]*event.Event, 0, len(items))
	for _, item := range items {
		e, err := s.eventSvc.GetEvent(ctx, item.TargetID)
		if err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, total, nil
}

func (s *BookmarkService) ensureTargetExists(ctx context.Context, targetType bookmark.TargetType, targetID uuid.UUID) error {
	switch targetType {
	case bookmark.TargetPost:
		_, err := s.postSvc.GetPost(ctx, targetID)
		return err
	case bookmark.TargetGroup:
		_, err := s.groupSvc.GetGroup(ctx, targetID)
		return err
	case bookmark.TargetEvent:
		_, err := s.eventSvc.GetEvent(ctx, targetID)
		return err
	default:
		return apperr.New(apperr.CodeInvalidParam, "不支持的收藏类型")
	}
}
