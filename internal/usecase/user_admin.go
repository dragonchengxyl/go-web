package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
)

// ListUsersInput represents input for listing users
type ListUsersInput struct {
	Page     int
	PageSize int
	Role     *user.Role
	Status   *user.Status
	Search   string
}

// ListUsersOutput represents output for listing users
type ListUsersOutput struct {
	Users []*user.User `json:"users"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// ListUsers retrieves users with pagination and filters
func (s *UserService) ListUsers(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	// Set default pagination
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := user.ListFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Role:     input.Role,
		Status:   input.Status,
		Search:   input.Search,
	}

	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户列表失败", err)
	}

	return &ListUsersOutput{
		Users: users,
		Total: total,
		Page:  input.Page,
		Size:  len(users),
	}, nil
}

// UpdateUserRoleInput represents input for updating user role
type UpdateUserRoleInput struct {
	Role user.Role `json:"role"`
}

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole user.Role) (*user.User, error) {
	// Get user
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Validate role
	validRoles := []user.Role{
		user.RoleSuperAdmin, user.RoleAdmin, user.RoleModerator,
		user.RoleCreator, user.RolePremium, user.RolePlayer, user.RoleGuest,
	}
	valid := false
	for _, r := range validRoles {
		if newRole == r {
			valid = true
			break
		}
	}
	if !valid {
		return nil, apperr.New(apperr.CodeInvalidParam, "无效的角色")
	}

	// Update role
	u.Role = newRole

	// Save to database
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新用户角色失败", err)
	}

	return u, nil
}

// UpdateUserStatusInput represents input for updating user status
type UpdateUserStatusInput struct {
	Status user.Status `json:"status"`
}

// UpdateUserStatus updates a user's status
func (s *UserService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, newStatus user.Status) (*user.User, error) {
	// Get user
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Validate status
	validStatuses := []user.Status{
		user.StatusActive, user.StatusInactive, user.StatusSuspended, user.StatusBanned,
	}
	valid := false
	for _, s := range validStatuses {
		if newStatus == s {
			valid = true
			break
		}
	}
	if !valid {
		return nil, apperr.New(apperr.CodeInvalidParam, "无效的状态")
	}

	// Update status
	u.Status = newStatus

	// Save to database
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新用户状态失败", err)
	}

	return u, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询用户失败", err)
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "删除用户失败", err)
	}

	return nil
}
