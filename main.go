package main

import (
	"ai_teach_system/config"
	"ai_teach_system/routes"
	"ai_teach_system/services"
	"ai_teach_system/tasks"
	"ai_teach_system/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	db := utils.InitDB()

	// 定时任务
	tasksManager := tasks.NewTasksManager(db, services.NewLeetCodeService(db))
	tasksManager.Start()
	defer tasksManager.Stop()

	r := gin.Default()
	routes.SetupRoutes(r, db)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败：", err)
	}
}
