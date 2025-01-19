package services

import (
	"ai_teach_system/config"
	"ai_teach_system/models"
	"errors"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type LoginRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Name      string `json:"name" binding:"required"`
	StudentID string `json:"student_id" binding:"required"`
	Class     string `json:"class" binding:"required"`
}

func (s *UserService) Login(req *LoginRequest) (string, error) {
	var user models.User
	if err := s.db.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户不存在")
		}
		return "", err
	}

	if !user.ValidatePassword(req.Password) {
		return "", errors.New("密码错误")
	}

	// 生成 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.JWT.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *UserService) Register(req *RegisterRequest) error {
	var count int64
	s.db.Model(&models.User{}).Where("username = ?", req.Username).Or("student_id = ?", req.StudentID).Count(&count)
	if count > 0 {
		return errors.New("用户已存在")
	}

	user := models.User{
		Username:  req.Username,
		Password:  req.Password,
		Name:      req.Name,
		StudentID: req.StudentID,
		Class:     req.Class,
	}

	return s.db.Create(&user).Error
}
