package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin 要求管理员权限
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户角色
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireModerator 要求版主或管理员权限
func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// 检查是否为版主或管理员
		if role != "admin" && role != "moderator" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要版主或管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
