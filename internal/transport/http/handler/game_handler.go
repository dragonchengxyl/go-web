package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/game"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	gameService *usecase.GameService
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(gameService *usecase.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

// CreateGame creates a new game (Admin only)
// @Summary Create game
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body usecase.CreateGameInput true "Game creation input"
// @Success 200 {object} response.Response{data=game.Game}
// @Router /api/v1/games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	// Get developer ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	developerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Parse input
	var input usecase.CreateGameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Create game
	game, err := h.gameService.CreateGame(c.Request.Context(), developerID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, game)
}

// UpdateGame updates a game (Admin only)
// @Summary Update game
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Param input body usecase.UpdateGameInput true "Game update input"
// @Success 200 {object} response.Response{data=game.Game}
// @Router /api/v1/games/{id} [put]
func (h *GameHandler) UpdateGame(c *gin.Context) {
	// Parse game ID
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// Parse input
	var input usecase.UpdateGameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update game
	game, err := h.gameService.UpdateGame(c.Request.Context(), gameID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, game)
}

// GetGameByID retrieves a game by ID
// @Summary Get game by ID
// @Tags game
// @Param id path string true "Game ID"
// @Success 200 {object} response.Response{data=game.Game}
// @Router /api/v1/games/{id} [get]
func (h *GameHandler) GetGameByID(c *gin.Context) {
	// Parse game ID
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// Get game
	game, err := h.gameService.GetGameByID(c.Request.Context(), gameID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, game)
}

// GetGameBySlug retrieves a game by slug
// @Summary Get game by slug
// @Tags game
// @Param slug path string true "Game slug"
// @Success 200 {object} response.Response{data=game.Game}
// @Router /api/v1/games/slug/{slug} [get]
func (h *GameHandler) GetGameBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "游戏标识不能为空"))
		return
	}

	// Get game
	game, err := h.gameService.GetGameBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, game)
}

// ListGames retrieves games with pagination and filters
// @Summary List games
// @Tags game
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param genre query string false "Filter by genre"
// @Param tag query string false "Filter by tag"
// @Param search query string false "Search by title or slug"
// @Success 200 {object} response.Response{data=usecase.ListGamesOutput}
// @Router /api/v1/games [get]
func (h *GameHandler) ListGames(c *gin.Context) {
	// Parse query parameters
	var input usecase.ListGamesInput
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

	if status := c.Query("status"); status != "" {
		st := game.Status(status)
		input.Status = &st
	}

	input.Genre = c.Query("genre")
	input.Tag = c.Query("tag")
	input.Search = c.Query("search")

	// List games
	output, err := h.gameService.ListGames(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, output)
}

// DeleteGame deletes a game (Admin only)
// @Summary Delete game
// @Tags game
// @Security BearerAuth
// @Param id path string true "Game ID"
// @Success 200 {object} response.Response
// @Router /api/v1/games/{id} [delete]
func (h *GameHandler) DeleteGame(c *gin.Context) {
	// Parse game ID
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// Delete game
	if err := h.gameService.DeleteGame(c.Request.Context(), gameID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "游戏已删除"})
}
