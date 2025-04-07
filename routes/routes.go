package routes

import (
	"ai_teach_system/controllers"
	"ai_teach_system/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/healthz", healthCheckHandler)

	r.Use(CORSMiddleware())

	api := r.Group("/api")

	userService := services.NewUserService(db)
	ossService, _ := services.NewOSSService()
	userController := controllers.NewUserController(userService, ossService)

	leetcodeService := services.NewLeetCodeService(db)
	leetcodeController := controllers.NewLeetCodeController(leetcodeService)

	aiService := services.NewAIService(db)
	aiController := controllers.NewAIController(aiService)

	courseService := services.NewCourseService(db)
	courseController := controllers.NewCourseController(courseService)

	problemService := services.NewProblemService(db)
	problemController := controllers.NewProblemController(problemService)

	classService := services.NewClassService(db)
	classController := controllers.NewClassController(classService)

	// 需要鉴权的路由
	auth := api.Group("")
	auth.Use(AuthMiddleware())
	{
		// LeetCode 相关路由
		leetcode := auth.Group("/leetcode")
		{
			leetcode.POST("/interpret_solution/", leetcodeController.RunTestCase)
			leetcode.POST("/submit/", leetcodeController.Submit)
			leetcode.POST("/check/", leetcodeController.Check)
		}

		// AI 相关路由
		ai := auth.Group("/ai")
		{
			ai.POST("/generate_hint/", aiController.GenerateHint)
			ai.POST("/correct_code/", aiController.CorrectCode)
			ai.POST("/analyze_code/", aiController.AnalyzeCode)
			ai.POST("/chat/", aiController.Chat)
			ai.POST("/judge/", aiController.JudgeCode)
		}

		// 用户相关路由
		users := auth.Group("/users")
		{
			users.GET("/", userController.GetUserInfo)
			users.POST("/reset_password/", userController.ResetPassword)
		}

		// 课程相关路由
		courses := auth.Group("/courses")
		{
			// 获取所有课程列表
			courses.GET("/", courseController.GetCourseList)
			// 添加新课程
			courses.POST("/", courseController.AddCourse)
			// 获取课程详情
			courses.GET("/:course_id/", courseController.GetCourseDetail)

			// 获取课程下的班级统计数据
			courses.GET("/:course_id/stats/", courseController.GetCourseClassStats)

			// 知识点相关路由
			knowledgePoints := courses.Group("/:course_id/knowledge_points")
			{
				// 获取课程下的知识点列表
				knowledgePoints.GET("/", courseController.GetKnowledgePoints)

				// 标签相关路由
				tags := knowledgePoints.Group("/:knowledge_point_id/tags")
				{
					// 设置知识点下的标签列表
					tags.POST("/", problemController.SetKnowledgePointTags)
				}

				// 题目相关路由
				problems := knowledgePoints.Group("/:knowledge_point_id/problems")
				{
					// 获取知识点下的题目列表
					problems.GET("/", problemController.GetKnowledgePointProblems)
					// 设置知识点下的题目列表
					problems.POST("/", problemController.SetKnowledgePointProblems)
				}

				// AI 相关路由
				ai := knowledgePoints.Group("/:knowledge_point_id/ai")
				{
					ai.GET("/suggest_tags/", aiController.SuggestKnowledgePointTags)
				}
			}

			// 班级相关路由
			classes := courses.Group("/:course_id/classes")
			{
				// 获取课程下的班级列表
				classes.GET("/", courseController.GetCourseClasses)
				// 设置课程下的班级列表
				classes.POST("/", courseController.SetCourseClasses)

				// 用户相关路由
				users := classes.Group("/:class_id/users")
				{
					// 获取某个课程和班级下的用户列表
					users.GET("/", userController.GetUserListByCourseAndClass)
				}
			}

			// 作答记录相关路由
			records := courses.Group("/:course_id/records")
			{
				records.GET("/", userController.GetCourseTryRecords)
				records.GET("/:id/", userController.GetTryRecordDetail)
			}

			// 题库相关路由
			problems := courses.Group("/:course_id/problems")
			{
				problems.POST("/", problemController.GetCourseProblemList)
			}
		}

		// 题库相关路由
		problems := auth.Group("/problems")
		{
			problems.GET("/:id/", problemController.GetProblemDetail)
			problems.POST("/", problemController.GetProblemList)
			problems.POST("/custom/", problemController.CreateCustomProblem)
			// 标签相关路由
			tags := problems.Group("/tags")
			{
				tags.GET("/", problemController.GetAllTags)
			}
		}

		// 课程相关路由
		classes := auth.Group("/classes")
		{
			classes.GET("/", classController.GetClassList)
			classes.POST("/", classController.AddClass)

			// 用户相关路由
			users := classes.Group("/:class_id/users")
			{
				users.GET("/", userController.GetUserListByClass)
			}
		}

		// 作答记录相关路由
		records := auth.Group("/records")
		{
			records.GET("/", userController.GetTryRecords)
		}
	}

	// 用户相关路由
	users := api.Group("/users")
	{
		users.POST("/login/", userController.Login)
		users.POST("/register/", userController.Register)
		users.POST("/select_course/", userController.SelectCourse)
		users.PUT("/:id/", userController.UpdateUser)
		users.DELETE("/:id/", userController.DeleteUser)
	}
}
