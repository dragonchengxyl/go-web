package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *usecase.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *usecase.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile retrieves current user profile
// @Summary Get current user profile
// @Tags user
// @Security BearerAuth
// @Success 200 {object} response.Response{data=user.User}
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Get profile
	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile updates current user profile
// @Summary Update current user profile
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body usecase.UpdateProfileInput true "Profile update input"
// @Success 200 {object} response.Response{data=user.User}
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Parse input
	var input usecase.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update profile
	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// GetUserByID retrieves user by ID
// @Summary Get user by ID
// @Tags user
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=user.User}
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// ListUsers retrieves users with pagination and filters (Admin only)
// @Summary List users
// @Tags admin
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by username or email"
// @Success 200 {object} response.Response{data=usecase.ListUsersOutput}
// @Router /api/v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	var input usecase.ListUsersInput
	input.Page = 1
	input.PageSize = 20

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			input.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			input.PageSize = ps
		}
	}

	if role := c.Query("role"); role != "" {
		r := user.Role(role)
		input.Role = &r
	}

	if status := c.Query("status"); status != "" {
		st := user.Status(status)
		input.Status = &st
	}

	input.Search = c.Query("search")

	// List users
	output, err := h.userService.ListUsers(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, output)
}

// UpdateUserRole updates a user's role (Admin only)
// @Summary Update user role
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param input body usecase.UpdateUserRoleInput true "Role update input"
// @Success 200 {object} response.Response{data=user.User}
// @Router /api/v1/admin/users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Parse input
	var input usecase.UpdateUserRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update role
	user, err := h.userService.UpdateUserRole(c.Request.Context(), userID, input.Role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus updates a user's status (Admin only)
// @Summary Update user status
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param input body usecase.UpdateUserStatusInput true "Status update input"
// @Success 200 {object} response.Response{data=user.User}
// @Router /api/v1/admin/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Parse input
	var input usecase.UpdateUserStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update status
	user, err := h.userService.UpdateUserStatus(c.Request.Context(), userID, input.Status)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// DeleteUser deletes a user (Admin only)
// @Summary Delete user
// @Tags admin
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "用户已删除"})
}

