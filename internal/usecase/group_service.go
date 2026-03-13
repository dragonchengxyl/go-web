package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/pkg/apperr"
)

// GroupService handles group business logic.
type GroupService struct {
	groupRepo group.Repository
}

// NewGroupService creates a new GroupService.
func NewGroupService(groupRepo group.Repository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

// CreateGroupInput holds data needed to create a group.
type CreateGroupInput struct {
	OwnerID      uuid.UUID
	Name         string
	Description  string
	Announcement string
	Rules        string
	Tags         []string
	Privacy      group.GroupPrivacy
}

// CreateGroup creates a new group and adds the owner as a member.
func (s *GroupService) CreateGroup(ctx context.Context, input CreateGroupInput) (*group.Group, error) {
	now := time.Now()
	g := &group.Group{
		ID:           uuid.New(),
		OwnerID:      input.OwnerID,
		Name:         input.Name,
		Description:  input.Description,
		Announcement: input.Announcement,
		Rules:        input.Rules,
		Tags:         input.Tags,
		Privacy:      input.Privacy,
		MemberCount:  1,
		PostCount:    0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if g.Privacy == "" {
		g.Privacy = group.GroupPrivacyPublic
	}

	if err := s.groupRepo.Create(ctx, g); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建圈子失败", err)
	}

	// Owner becomes first member.
	_ = s.groupRepo.AddMember(ctx, &group.GroupMember{
		GroupID:  g.ID,
		UserID:   input.OwnerID,
		Role:     group.GroupRoleOwner,
		JoinedAt: now,
	})

	return g, nil
}

// GetGroup returns a group by ID.
func (s *GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*group.Group, error) {
	g, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}
	return g, nil
}

// UpdateGroupInput holds fields that can be updated.
type UpdateGroupInput struct {
	Name         string
	Description  string
	Announcement string
	Rules        string
	Tags         []string
	Privacy      group.GroupPrivacy
}

// UpdateGroup updates a group; only owner/moderator may update.
func (s *GroupService) UpdateGroup(ctx context.Context, callerID, groupID uuid.UUID, input UpdateGroupInput) (*group.Group, error) {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}

	member, err := s.groupRepo.GetMember(ctx, groupID, callerID)
	if err != nil || member == nil || (member.Role != group.GroupRoleOwner && member.Role != group.GroupRoleModerator) {
		return nil, apperr.New(apperr.CodeForbidden, "无权修改此圈子")
	}

	if input.Name != "" {
		g.Name = input.Name
	}
	if input.Description != "" {
		g.Description = input.Description
	}
	announcementChanged := input.Announcement != g.Announcement
	if input.Announcement != "" || g.Announcement != "" {
		g.Announcement = input.Announcement
	}
	if input.Rules != "" || g.Rules != "" {
		g.Rules = input.Rules
	}
	if len(input.Tags) > 0 {
		g.Tags = input.Tags
	}
	if input.Privacy != "" {
		g.Privacy = input.Privacy
	}
	g.UpdatedAt = time.Now()

	if err := s.groupRepo.Update(ctx, g); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新圈子失败", err)
	}
	if announcementChanged && strings.TrimSpace(input.Announcement) != "" {
		_ = s.groupRepo.CreateAnnouncement(ctx, &group.GroupAnnouncement{
			ID:        uuid.New(),
			GroupID:   groupID,
			AuthorID:  callerID,
			Content:   input.Announcement,
			CreatedAt: time.Now(),
		})
	}
	return g, nil
}

// ListGroupsInput holds filtering options for discovering groups.
type ListGroupsInput struct {
	Privacy  *group.GroupPrivacy
	Search   string
	Page     int
	PageSize int
}

// ListGroups returns a paginated list of groups.
func (s *GroupService) ListGroups(ctx context.Context, input ListGroupsInput) ([]*group.Group, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 50 {
		input.PageSize = 20
	}

	filter := group.ListFilter{
		Privacy:  input.Privacy,
		Search:   input.Search,
		Page:     input.Page,
		PageSize: input.PageSize,
	}
	groups, total, err := s.groupRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询圈子列表失败", err)
	}
	return groups, total, nil
}

// SetFeaturedPost sets or clears the featured post for a group; only owner/moderator may do this.
func (s *GroupService) SetFeaturedPost(ctx context.Context, callerID, groupID uuid.UUID, postID *uuid.UUID) (*group.Group, error) {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}

	member, err := s.groupRepo.GetMember(ctx, groupID, callerID)
	if err != nil || member == nil || (member.Role != group.GroupRoleOwner && member.Role != group.GroupRoleModerator) {
		return nil, apperr.New(apperr.CodeForbidden, "只有圈主或管理员可以设置精选内容")
	}

	g.FeaturedPostID = postID
	g.UpdatedAt = time.Now()
	if err := s.groupRepo.Update(ctx, g); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新圈子精选失败", err)
	}
	return g, nil
}

// ListAnnouncements returns paginated announcement history entries for a group.
func (s *GroupService) ListAnnouncements(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*group.GroupAnnouncement, int64, error) {
	items, total, err := s.groupRepo.ListAnnouncements(ctx, groupID, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询圈子公告历史失败", err)
	}
	return items, total, nil
}

// JoinGroup adds a user to a group.
func (s *GroupService) JoinGroup(ctx context.Context, groupID, userID uuid.UUID) error {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}
	_ = g

	existing, _ := s.groupRepo.GetMember(ctx, groupID, userID)
	if existing != nil {
		return apperr.New(apperr.CodeInvalidParam, "已是圈子成员")
	}

	if err := s.groupRepo.AddMember(ctx, &group.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     group.GroupRoleMember,
		JoinedAt: time.Now(),
	}); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "加入圈子失败", err)
	}

	_ = s.groupRepo.IncrementMemberCount(ctx, groupID)
	return nil
}

// LeaveGroup removes a user from a group. The owner cannot leave.
func (s *GroupService) LeaveGroup(ctx context.Context, groupID, userID uuid.UUID) error {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}
	if g.OwnerID == userID {
		return apperr.New(apperr.CodeInvalidParam, "圈主不能退出圈子")
	}

	member, _ := s.groupRepo.GetMember(ctx, groupID, userID)
	if member == nil {
		return apperr.New(apperr.CodeInvalidParam, "未加入此圈子")
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "退出圈子失败", err)
	}
	_ = s.groupRepo.DecrementMemberCount(ctx, groupID)
	return nil
}

// ListMembers returns paginated members of a group.
func (s *GroupService) ListMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*group.GroupMember, int64, error) {
	members, total, err := s.groupRepo.ListMembers(ctx, groupID, page, pageSize)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询成员列表失败", err)
	}
	return members, total, nil
}

// GetMember returns the caller's membership info in a group.
func (s *GroupService) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*group.GroupMember, error) {
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询圈子成员失败", err)
	}
	return member, nil
}

// CanViewGroup reports whether the caller can view content inside the group.
func (s *GroupService) CanViewGroup(ctx context.Context, groupID, viewerID uuid.UUID, isAuthenticated bool) (bool, *group.Group, error) {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return false, nil, apperr.ErrNotFound
		}
		return false, nil, apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}
	if g.Privacy == group.GroupPrivacyPublic {
		return true, g, nil
	}
	if !isAuthenticated {
		return false, g, nil
	}
	member, err := s.groupRepo.GetMember(ctx, groupID, viewerID)
	if err != nil {
		return false, g, apperr.Wrap(apperr.CodeInternalError, "查询圈子成员失败", err)
	}
	return member != nil, g, nil
}

// UpdateMemberRole changes a member's role; only the owner may do this.
func (s *GroupService) UpdateMemberRole(ctx context.Context, callerID, groupID, targetUserID uuid.UUID, role group.GroupRole) error {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}
	if g.OwnerID != callerID {
		return apperr.New(apperr.CodeForbidden, "只有圈主可以修改成员角色")
	}
	if err := s.groupRepo.UpdateMemberRole(ctx, groupID, targetUserID, role); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "更新成员角色失败", err)
	}
	return nil
}

// KickMember removes a member from the group; owner/mod may kick.
func (s *GroupService) KickMember(ctx context.Context, callerID, groupID, targetUserID uuid.UUID) error {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, group.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询圈子失败", err)
	}

	callerMember, _ := s.groupRepo.GetMember(ctx, groupID, callerID)
	if callerMember == nil || (callerMember.Role != group.GroupRoleOwner && callerMember.Role != group.GroupRoleModerator) {
		return apperr.New(apperr.CodeForbidden, "无权踢出成员")
	}
	if g.OwnerID == targetUserID {
		return apperr.New(apperr.CodeInvalidParam, "不能踢出圈主")
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, targetUserID); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "踢出成员失败", err)
	}
	_ = s.groupRepo.DecrementMemberCount(ctx, groupID)
	return nil
}

// MyGroups returns groups the user has joined.
func (s *GroupService) MyGroups(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error) {
	type lister interface {
		ListByMember(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error)
	}
	if l, ok := s.groupRepo.(lister); ok {
		groups, total, err := l.ListByMember(ctx, userID, page, pageSize)
		if err != nil {
			return nil, 0, apperr.Wrap(apperr.CodeInternalError, "查询我的圈子失败", err)
		}
		return groups, total, nil
	}
	// Fallback: list all and filter (not used in practice).
	return nil, 0, nil
}
