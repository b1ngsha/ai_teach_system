package routes

import (
	"ai_teach_system/controllers"
	"ai_teach_system/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api")

	userService := services.NewUserService(db)
	ossService, _ := services.NewOSSService()
	userController := controllers.NewUserController(userService, ossService)

	leetcodeService := services.NewLeetCodeService(db)
	leetcodeController := controllers.NewLeetCodeController(leetcodeService)

	aiService := services.NewAIService()
	aiController := controllers.NewAIController(aiService)

	courseService := services.NewCourseService(db)
	courseController := controllers.NewCourseController(courseService)

	problemService := services.NewProblemService(db)
	problemController := controllers.NewProblemController(problemService)

	// 需要鉴权的路由
	auth := api.Group("")
	auth.Use(AuthMiddleware())
	{
		// LeetCode 相关路由
		leetcode := auth.Group("/leetcode")
		{
			leetcode.POST("/interpret_solution", leetcodeController.RunTestCase)
			leetcode.POST("/submit", leetcodeController.Submit)
		}

		// AI 相关路由
		ai := auth.Group("/ai")
		{
			ai.POST("/generate_code", aiController.GenerateCode)
		}

		// 用户相关路由
		users := auth.Group("/users")
		{
			users.GET("", userController.GetUserInfo)
		}

		// 课程相关路由
		courses := auth.Group("/courses")
		{
			courses.GET("/:id", courseController.GetCourseDetail)
		}

		// 题库相关路由
		problems := auth.Group("/problems")
		{
			problems.POST("", problemController.GetProblemList)
			problems.GET("/:id", problemController.GetProblemDetail)
		}
	}

	// 用户相关路由
	users := api.Group("/users")
	{
		users.POST("/login", userController.Login)
		users.POST("/register", userController.Register)
	}
}
