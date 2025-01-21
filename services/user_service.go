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

func (s *UserService) Login(studentID string, password string) (string, error) {
	var user models.User
	if err := s.db.Where("student_id = ?", studentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户不存在")
		}
		return "", err
	}

	if !user.ValidatePassword(password) {
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

func (s *UserService) Register(username string, password string, name string, studentId string, class string, avatar string) error {
	var count int64
	s.db.Model(&models.User{}).Where("username = ?", username).Or("student_id = ?", studentId).Count(&count)
	if count > 0 {
		return errors.New("用户已存在")
	}

	user := models.User{
		Username:  username,
		Password:  password,
		Name:      name,
		StudentID: studentId,
		Class:     class,
		Avatar:    avatar,
	}

	return s.db.Create(&user).Error
}

func (s *UserService) GetUserInfo(userID uint) (map[string]interface{}, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var totalProblems int64
	s.db.Model(&models.Problem{}).Count(&totalProblems)

	if totalProblems == 0 {
		return nil, errors.New("当前系统中题库为空，请先同步题目数据")
	}

	var solvedProblems int64
	s.db.Model(&models.UserProblem{}).
		Where("user_id = ? AND status = ?", userID, models.ProblemStatusSolved).
		Count(&solvedProblems)

	completionRate := float64(solvedProblems) / float64(totalProblems) * 100

	return map[string]interface{}{
		"avatar":          user.Avatar,
		"username":        user.Username,
		"learn_progress":  completionRate,
		"solved_problems": solvedProblems,
	}, nil
}
