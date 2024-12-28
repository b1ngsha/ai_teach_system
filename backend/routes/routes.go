package routes

import (
	"ai_teach_system/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// API 路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		api.GET("/users", controllers.GetUsers)
		api.POST("/users", controllers.CreateUser)
	}
}
