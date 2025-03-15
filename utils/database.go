package utils

import (
	"ai_teach_system/config"
	"ai_teach_system/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.DBUser,
		config.DB.DBPassword,
		config.DB.DBHost,
		config.DB.DBPort,
		config.DB.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal("数据库连接失败：", err)
	}

	// 自动迁移数据库结构
	err = db.AutoMigrate(
		&models.Problem{},
		&models.Tag{},
		&models.TaskRecord{},
		&models.User{},
		&models.Course{},
		&models.Class{},
		&models.KnowledgePoint{},
		&models.UserProblem{},
		&models.CourseProblem{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败：", err)
	}

	return db
}
