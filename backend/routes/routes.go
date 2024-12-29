package routes

import (
	"ai_teach_system/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// API 路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		api.GET("/users", controllers.GetUsers)
		api.POST("/users", controllers.CreateUser)
	}

	// 创建控制器实例
	leetcodeController := controllers.NewLeetCodeController(db)

	// LeetCode 相关路由
	leetcode := api.Group("/leetcode")
	{
		leetcode.GET("/problems/:id", leetcodeController.GetProblem)
		leetcode.POST("/problems/:titleSlug/fetch", leetcodeController.FetchProblem)
		leetcode.POST("/problems/fetch-all", leetcodeController.FetchAllProblems)
	}
}
