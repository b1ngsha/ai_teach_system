package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type dbConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

type jwtConfig struct {
	SecretKey string
}

type ossConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

type leetcodeConfig struct {
	LeetcodeSession string
}

var DB dbConfig
var JWT jwtConfig
var OSS ossConfig
var Leetcode leetcodeConfig

func LoadConfig() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DB = dbConfig{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "mydb"),
	}

	JWT = jwtConfig{
		SecretKey: getEnv("JWT_SECRET_KEY", ""),
	}

	OSS = ossConfig{
		Endpoint:        getEnv("ALIYUN_OSS_ENDPOINT", ""),
		AccessKeyID:     getEnv("ALIYUN_ACCESS_KEY", ""),
		AccessKeySecret: getEnv("ALIYUN_ACCESS_SECRET", ""),
		BucketName:      getEnv("ALIYUN_OSS_BUCKET_NAME", ""),
	}

	Leetcode = leetcodeConfig{
		LeetcodeSession: getEnv("LEETCODE_SESSION", ""),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
