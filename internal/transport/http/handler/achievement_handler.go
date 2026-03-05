package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AchievementHandler handles achievement HTTP requests
type AchievementHandler struct {
	service *usecase.AchievementService
}

// NewAchievementHandler creates a new AchievementHandler
func NewAchievementHandler(service *usecase.AchievementService) *AchievementHandler {
	return &AchievementHandler{service: service}
}

// ListAchievements returns all public achievements
// GET /api/v1/achievements
func (h *AchievementHandler) ListAchievements(c *gin.Context) {
	list, err := h.service.ListAchievements(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

// GetMyAchievements returns achievements unlocked by the current user
// GET /api/v1/users/me/achievements
func (h *AchievementHandler) GetMyAchievements(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	list, err := h.service.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

// GetUserAchievements returns achievements unlocked by any user (public)
// GET /api/v1/users/:id/achievements
func (h *AchievementHandler) GetUserAchievements(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	list, err := h.service.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

// GetMyPoints returns the current user's point balance
// GET /api/v1/users/me/points
func (h *AchievementHandler) GetMyPoints(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	up, err := h.service.GetUserPoints(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, up)
}

// GetMyPointHistory returns the current user's point transaction history
// GET /api/v1/users/me/points/history
func (h *AchievementHandler) GetMyPointHistory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	txs, err := h.service.GetPointTransactions(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, txs)
}

// GetLeaderboard returns the overall points leaderboard
// GET /api/v1/leaderboard
func (h *AchievementHandler) GetLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	entries, err := h.service.GetTopLeaderboard(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, entries)
}

// GetWeeklyLeaderboard returns the weekly points leaderboard
// GET /api/v1/leaderboard/weekly
func (h *AchievementHandler) GetWeeklyLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	entries, err := h.service.GetWeeklyLeaderboard(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, entries)
}

// AdminUnlockAchievement manually unlocks an achievement for a user (admin)
// POST /api/v1/admin/users/:id/achievements
func (h *AchievementHandler) AdminUnlockAchievement(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Slug string `json:"slug" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	ua, err := h.service.UnlockBySlug(c.Request.Context(), userID, input.Slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ua)
}

// AdminAwardPoints manually awards points to a user (admin)
// POST /api/v1/admin/users/:id/points
func (h *AchievementHandler) AdminAwardPoints(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Amount int    `json:"amount" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	if err := h.service.AwardPoints(c.Request.Context(), userID, input.Amount, "admin", input.Note); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// getUserID is a helper to extract the authenticated user's UUID from context
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	idStr, ok := val.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
