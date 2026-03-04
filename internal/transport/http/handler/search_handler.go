package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/usecase"
)

type SearchHandler struct {
	gameService  *usecase.GameService
	musicService *usecase.MusicService
}

func NewSearchHandler(gameService *usecase.GameService, musicService *usecase.MusicService) *SearchHandler {
	return &SearchHandler{
		gameService:  gameService,
		musicService: musicService,
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
	// TODO: 从Redis或数据库获取热门搜索词
	popularSearches := []string{
		"动作游戏",
		"独立游戏",
		"冒险游戏",
		"原声音乐",
		"钢琴曲",
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    popularSearches,
	})
}
