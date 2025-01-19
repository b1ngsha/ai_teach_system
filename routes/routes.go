package routes

import (
	"ai_teach_system/controllers"
	"ai_teach_system/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// API 路由组
	api := r.Group("/api")

	// LeetCode 相关路由
	leetcodeController := controllers.NewLeetCodeController(db)
	leetcode := api.Group("/leetcode")
	{
		leetcode.GET("/problems/:id", leetcodeController.GetProblem)
	}

	// AI 相关路由
	aiController := controllers.NewAIController()
	ai := api.Group("/ai")
	{
		ai.POST("/generate_code", aiController.GenerateCode)
	}

	// 用户相关路由
	userService := services.NewUserService(db)
	userController := controllers.NewUserController(userService)
	users := api.Group("/users")
	{
		users.POST("/login", userController.Login)
	}
}
