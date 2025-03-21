package services

import (
	"ai_teach_system/models"
	"ai_teach_system/utils"
	"errors"
	"fmt"

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
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) Register(username string, password string, name string, studentId string, className string) (*models.User, error) {
	var count int64
	s.db.Model(&models.User{}).Where("username = ?", username).Or("student_id = ?", studentId).Or("name = ?", name).Count(&count)
	if count > 0 {
		return nil, errors.New("用户已存在")
	}

	var class models.Class
	err := s.db.Model(&models.Class{}).Where("name = ?", className).First(&class).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("请先创建班级: %s", className)
		}
		return nil, fmt.Errorf("查询班级信息失败: %v", err)
	}

	s.db.Model(&models.Class{}).Where("name = ")
	user := models.User{
		Username:  username,
		Password:  password,
		Name:      name,
		StudentID: studentId,
		Class:     class,
		ClassID:   class.ID,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("创建用户失败")
	}

	return &user, nil
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
		"username":        user.Username,
		"learn_progress":  completionRate,
		"solved_problems": solvedProblems,
	}, nil
}

func (s *UserService) CreateAdminIfNotExists() error {
	var count int64
	s.db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	if count > 0 {
		return nil
	}

	admin := models.User{
		Username:  "admin",
		Password:  "szu_admin",
		Name:      "系统管理员",
		StudentID: "admin",
		Role:      models.RoleAdmin,
	}

	return s.db.Create(&admin).Error
}

func (s *UserService) GetTryRecords(userID uint) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := s.db.Select("user_problems.id, user_problems.problem_id, courses.name AS course_name, knowledge_points.name AS knowledge_point_name, problems.title_cn, problems.title, user_problems.status, user_problems.updated_at").
		Model(&models.UserProblem{}).
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Joins("JOIN knowledge_points ON user_problems.knowledge_point_id = knowledge_points.id").
		Joins("JOIN courses ON knowledge_points.course_id = courses.id").
		Where("user_problems.user_id = ?", userID).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}
	return result, err
}

func (s *UserService) GetTryRecordDetail(userID uint, recordID uint) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.db.Select("problems.title, problems.title_cn, problems.content, user_problems.typed_code, user_problems.wrong_reason_and_analyze, user_problems.corrected_code").
		Model(&models.UserProblem{}).
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Where("user_problems.user_id = ? AND user_problems.id = ?", userID, recordID).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
