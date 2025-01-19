package routes

import (
	"ai_teach_system/controllers"
	"ai_teach_system/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api")

	// 需要鉴权的路由
	api.Use(AuthMiddleware())
	{
		// LeetCode 相关路由
		leetcodeService := services.NewLeetCodeService()
		leetcodeController := controllers.NewLeetCodeController(db, leetcodeService)
		leetcode := api.Group("/leetcode")
		{
			leetcode.GET("/problems/:id", leetcodeController.GetProblem)
			leetcode.POST("/interpret_solution", leetcodeController.RunTestCase)
			leetcode.POST("/submit", leetcodeController.Submit)
		}

		// AI 相关路由
		aiService := services.NewAIService()
		aiController := controllers.NewAIController(aiService)
		ai := api.Group("/ai")
		{
			ai.POST("/generate_code", aiController.GenerateCode)
		}
	}

	// 用户相关路由
	userService := services.NewUserService(db)
	userController := controllers.NewUserController(userService)
	users := api.Group("/users")
	{
		users.POST("/login", userController.Login)
	}
}
