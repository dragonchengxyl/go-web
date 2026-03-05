package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	statsService *usecase.StatsService
	userService  *usecase.UserService
	gameService  *usecase.GameService
	commentService *usecase.CommentService
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(
	statsService *usecase.StatsService,
	userService *usecase.UserService,
	gameService *usecase.GameService,
	commentService *usecase.CommentService,
) *AdminHandler {
	return &AdminHandler{
		statsService:   statsService,
		userService:    userService,
		gameService:    gameService,
		commentService: commentService,
	}
}

// GetDashboardStats returns main dashboard metrics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, stats)
}

// GetRevenueChart returns daily revenue chart data
func (h *AdminHandler) GetRevenueChart(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.statsService.GetRevenueChart(c.Request.Context(), days)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

// GetUserGrowthChart returns daily user registration chart data
func (h *AdminHandler) GetUserGrowthChart(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.statsService.GetUserGrowthChart(c.Request.Context(), days)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

// GetPopularGames returns top games by download count
func (h *AdminHandler) GetPopularGames(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data, err := h.statsService.GetPopularGames(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

// ListUsers returns paginated user list with filters
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	input := usecase.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		s := user.Status(statusStr)
		input.Status = &s
	}

	result, err := h.userService.ListUsers(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateUserRole changes a user's role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Role user.Role `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	u, err := h.userService.UpdateUserRole(c.Request.Context(), userID, input.Role)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// BanUser bans a user
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	u, err := h.userService.UpdateUserStatus(c.Request.Context(), userID, user.StatusBanned)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// UnbanUser unbans a user
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	u, err := h.userService.UpdateUserStatus(c.Request.Context(), userID, user.StatusActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListComments returns paginated comments for moderation
func (h *AdminHandler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.commentService.AdminListComments(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteComment deletes a comment (admin)
func (h *AdminHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	if err := h.commentService.AdminDeleteComment(c.Request.Context(), commentID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
