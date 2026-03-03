package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/usecase"
)

type AdminHandler struct {
	gameService *usecase.GameService
	userService *usecase.UserAdminService
}

func NewAdminHandler(gameService *usecase.GameService, userService *usecase.UserAdminService) *AdminHandler {
	return &AdminHandler{
		gameService: gameService,
		userService: userService,
	}
}

// GetDashboardStats 获取后台统计数据
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	// TODO: 实现真实的统计逻辑
	stats := map[string]interface{}{
		"online_users":     123,
		"today_new_users":  45,
		"today_downloads":  678,
		"today_revenue":    12345.67,
		"total_users":      10000,
		"total_games":      25,
		"total_orders":     5000,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetPopularGames 获取热门游戏
func (h *AdminHandler) GetPopularGames(c *gin.Context) {
	// TODO: 实现真实的热门游戏统计
	games := []map[string]interface{}{
		{
			"id":        1,
			"title":     "示例游戏 1",
			"downloads": 1000,
			"revenue":   5000.00,
		},
		{
			"id":        2,
			"title":     "示例游戏 2",
			"downloads": 800,
			"revenue":   4000.00,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    games,
	})
}

// ListGames 获取游戏列表（管理员）
func (h *AdminHandler) ListGames(c *gin.Context) {
	games, err := h.gameService.ListAllGames(c.Request.Context())
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

// GetGame 获取游戏详情（管理员）
func (h *AdminHandler) GetGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的游戏ID",
		})
		return
	}

	game, err := h.gameService.GetGameByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "游戏不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    game,
	})
}

// CreateGame 创建游戏
func (h *AdminHandler) CreateGame(c *gin.Context) {
	var req struct {
		Title            string   `json:"title" binding:"required"`
		Slug             string   `json:"slug" binding:"required"`
		Description      string   `json:"description"`
		ShortDescription string   `json:"short_description"`
		CoverImage       string   `json:"cover_image"`
		Price            float64  `json:"price"`
		DiscountPrice    float64  `json:"discount_price"`
		Status           string   `json:"status"`
		Tags             []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	game, err := h.gameService.CreateGame(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    game,
	})
}

// UpdateGame 更新游戏
func (h *AdminHandler) UpdateGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的游戏ID",
		})
		return
	}

	var req struct {
		Title            string   `json:"title"`
		Slug             string   `json:"slug"`
		Description      string   `json:"description"`
		ShortDescription string   `json:"short_description"`
		CoverImage       string   `json:"cover_image"`
		Price            float64  `json:"price"`
		DiscountPrice    float64  `json:"discount_price"`
		Status           string   `json:"status"`
		Tags             []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	err = h.gameService.UpdateGame(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
	})
}

// DeleteGame 删除游戏
func (h *AdminHandler) DeleteGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的游戏ID",
		})
		return
	}

	err = h.gameService.DeleteGame(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListUsers 获取用户列表
func (h *AdminHandler) ListUsers(c *gin.Context) {
	search := c.Query("search")

	users, err := h.userService.ListUsers(c.Request.Context(), search)
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
		"data":    users,
	})
}

// UpdateUserRole 更新用户角色
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	err = h.userService.UpdateUserRole(c.Request.Context(), id, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
	})
}

// BanUser 封禁用户
func (h *AdminHandler) BanUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	err = h.userService.BanUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "封禁成功",
	})
}

// UnbanUser 解封用户
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	err = h.userService.UnbanUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "解封成功",
	})
}
