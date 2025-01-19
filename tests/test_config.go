package tests

import (
	"ai_teach_system/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SetupTestLeetCodeServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回模拟的 GraphQL 响应
		req, _ := io.ReadAll(r.Body)
		var reqJson map[string]interface{}
		json.Unmarshal(req, &reqJson)
		w.Header().Set("Content-Type", "application/json")
		if reqJson["operationName"] == "problemsetQuestionList" {
			w.Write([]byte(`{
				"data": {
					"problemsetQuestionList": {
						"hasMore": false,
						"questions": [
							{
								"titleSlug": "two-sum",
								"title": "Two Sum",
								"difficulty": "Easy",
								"topicTags": [
									{"name": "Array"},
									{"name": "Hash Table"}
								]
							}
						]
					}
				}
			}`))
		} else if reqJson["operationName"] == "questionData" {
			w.Write([]byte(`{
				"data": {
					"question": {
						"questionId": "1",
						"title": "Two Sum",
						"titleSlug": "two-sum",
						"difficulty": "Easy",
						"content": "Given an array of integers...",
						"sampleTestCase": "[2,7,11,15]\n9"
					}
				}
			}`))
		}
	}))
}

func SetupTestDB() (*gorm.DB, func()) {
	err := godotenv.Load("../.env")
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
		&models.KnowledgePoint{},
		&models.TaskRecord{},
		&models.User{},
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
