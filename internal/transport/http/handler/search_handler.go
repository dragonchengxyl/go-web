package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/usecase"
)

type SearchHandler struct {
	gameService   *usecase.GameService
	musicService  *usecase.MusicService
	searchService *usecase.SearchService
}

func NewSearchHandler(gameService *usecase.GameService, musicService *usecase.MusicService, searchService *usecase.SearchService) *SearchHandler {
	return &SearchHandler{
		gameService:   gameService,
		musicService:  musicService,
		searchService: searchService,
	}
}

// SearchAll 全局搜索
func (h *SearchHandler) SearchAll(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	// 搜索游戏
	games, _ := h.gameService.SearchGames(c.Request.Context(), query, 5)

	// 搜索音乐
	albums, _ := h.musicService.SearchAlbums(c.Request.Context(), query, 5)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"games":  games,
			"albums": albums,
			"query":  query,
		},
	})
}

// SearchGames 搜索游戏
func (h *SearchHandler) SearchGames(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	games, err := h.gameService.SearchGames(c.Request.Context(), query, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    games,
	})
}

// SearchAlbums 搜索音乐专辑
func (h *SearchHandler) SearchAlbums(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	albums, err := h.musicService.SearchAlbums(c.Request.Context(), query, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    albums,
	})
}

// GetPopularSearches 获取热门搜索
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	searches, err := h.searchService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    searches,
	})
}
