package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type SearchHandler struct {
	musicService  *usecase.MusicService
	searchService *usecase.SearchService
	postService   *usecase.PostService
	userService   *usecase.UserService
}

func NewSearchHandler(musicService *usecase.MusicService, searchService *usecase.SearchService, postService *usecase.PostService, userService *usecase.UserService) *SearchHandler {
	return &SearchHandler{
		musicService:  musicService,
		searchService: searchService,
		postService:   postService,
		userService:   userService,
	}
}

// SearchAll 全局搜索（posts + users + albums）
func (h *SearchHandler) SearchAll(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "搜索关键词不能为空"})
		return
	}

	var posts, users, albums any

	if h.postService != nil {
		p, _ := h.postService.SearchPosts(c.Request.Context(), query, 20)
		posts = p
	}
	if h.userService != nil {
		u, _ := h.userService.SearchUsers(c.Request.Context(), query, 20)
		users = u
	}
	if h.musicService != nil {
		a, _ := h.musicService.SearchAlbums(c.Request.Context(), query, 10)
		albums = a
	}

	response.Success(c, gin.H{
		"posts":  posts,
		"users":  users,
		"albums": albums,
		"query":  query,
	})
}

// SearchAlbums 搜索音乐专辑
func (h *SearchHandler) SearchAlbums(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "搜索关键词不能为空"})
		return
	}

	if h.musicService == nil {
		response.Success(c, []any{})
		return
	}

	albums, err := h.musicService.SearchAlbums(c.Request.Context(), query, 20)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, albums)
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
	response.Success(c, searches)
}
