package main

import (
	"ai_teach_system/config"
	"ai_teach_system/routes"
	"ai_teach_system/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	db := utils.InitDB()

	// 获取通用数据库对象 sql.DB，然后使用其提供的功能
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库实例失败：", err)
	}
	defer sqlDB.Close()

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r, db)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败：", err)
	}
}
