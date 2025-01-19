package routes

import (
	"ai_teach_system/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "未登录，请先登录",
			})
			return
		}

		// Bearer token格式验证
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效的Token认证格式",
			})
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效的token",
			})
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
