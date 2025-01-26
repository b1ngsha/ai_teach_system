package main

import (
	"ai_teach_system/config"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tasks"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	config.LoadConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.DBUser,
		config.DB.DBPassword,
		config.DB.DBHost,
		config.DB.DBPort,
		config.DB.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	now := time.Now()
	taskRecord := &models.TaskRecord{
		TaskType:  "leetcode_sync",
		Status:    models.TaskStatusPending,
		StartTime: &now,
	}
	if err := db.Create(taskRecord).Error; err != nil {
		log.Fatalf("创建任务记录失败: %v", err)
	}

	leetcodeService := services.NewLeetCodeService(db)
	task := tasks.NewTasksManager(db, leetcodeService)
	task.SyncLeetCodeProblems()

	if err != nil {
		log.Printf("同步任务失败: %v", err)
		db.Model(taskRecord).Updates(map[string]interface{}{
			"status":        models.TaskStatusFailed,
			"error_message": err.Error(),
			"end_time":      time.Now(),
		})
	} else {
		log.Println("同步任务完成")
		db.Model(taskRecord).Updates(map[string]interface{}{
			"status":   models.TaskStatusCompleted,
			"end_time": time.Now(),
		})
	}
}
