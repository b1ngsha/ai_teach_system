package tests

import (
	"ai_teach_system/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SetupTestDB() (*gorm.DB, func()) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbName := "test_" + os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL server:", err)
	}

	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))

	// 创建测试数据库
	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error; err != nil {
		log.Fatal("Failed to create test database:", err)
	}

	// 切换到测试数据库
	if err := db.Exec(fmt.Sprintf("USE %s", dbName)).Error; err != nil {
		log.Fatal("Failed to switch to test database:", err)
	}

	// 自动迁移数据库结构
	err = db.AutoMigrate(
		&models.Problem{},
		&models.Tag{},
		&models.TaskRecord{},
		&models.User{},
		&models.UserProblem{},
	)
	if err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}

	// 返回删除测试数据库函数
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Failed to get current DB: %v", err)
			return
		}
		sqlDB.Close()

		// 重新连接以删除测试数据库
		cleanupDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Failed to connect for cleanup: %v", err)
			return
		}
		if err := cleanupDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error; err != nil {
			log.Printf("Failed to drop test database: %v", err)
		}
	}

	return db, cleanup
}
