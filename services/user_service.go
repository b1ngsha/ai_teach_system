package services

import (
	"ai_teach_system/models"
	"ai_teach_system/utils"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Login(studentID, password string) (string, error) {
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

func (s *UserService) Register(username, password, name, studentId, className string) (*models.User, error) {
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

func (s *UserService) GetCourseTryRecords(courseID, userID uint) ([]map[string]interface{}, error) {
	knowledge_point_ids := []uint{}
	err := s.db.Select("id").
		Model(&models.KnowledgePoint{}).
		Where("course_id = ?", courseID).
		Scan(&knowledge_point_ids).
		Error
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	err = s.db.Select("user_problems.id, user_problems.problem_id, knowledge_points.name AS knowledge_point_name, problems.title_cn, problems.title, user_problems.status, user_problems.updated_at").
		Model(&models.UserProblem{}).
		Joins("JOIN knowledge_points ON user_problems.knowledge_point_id = knowledge_points.id").
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Where("knowledge_point_id in (?) AND user_id = ?", knowledge_point_ids, userID).
		Scan(&records).
		Error
	if err != nil {
		return nil, err
	}

	return records, err
}

func (s *UserService) GetTryRecordDetail(recordID uint) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.db.Select("problems.title, problems.title_cn, problems.content, problems.content_cn, problems.difficulty, user_problems.typed_code, user_problems.deepseek_wrong_reason_and_analyze, qwen_wrong_reason_and_analyze, user_problems.qwen_corrected_code, user_problems.deepseek_corrected_code").
		Model(&models.UserProblem{}).
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Where("user_problems.id = ?", recordID).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UserService) GetUserListByCourseAndClass(classID, courseID uint) ([]map[string]interface{}, error) {
	// 查询当前班级的用户列表
	var userList []models.User
	err := s.db.Model(&models.User{}).Where("class_id = ?", classID).Find(&userList).Error
	if err != nil {
		return nil, err
	}

	// 查询每个用户的答题数据
	result := make([]map[string]interface{}, len(userList))
	for i, user := range userList {
		// 先查出课程关联的知识点
		var courseKnowledgePointIDs []uint
		err = s.db.Select("id").
			Model(&models.KnowledgePoint{}).
			Where("course_id = ?", courseID).
			Scan(&courseKnowledgePointIDs).
			Error
		if err != nil {
			return nil, err
		}

		// 根据知识点-标签的关联关系查询出当前课程下的知识点关联的标签id
		var tagIDs []uint
		err = s.db.Model(&models.KnowledgePointTag{}).
			Select("tag_id").
			Where("knowledge_point_id IN (?)", courseKnowledgePointIDs).
			Find(&tagIDs).
			Error
		if err != nil {
			return nil, err
		}

		// 根据标签id查询关联的题目数量
		var totalProblemCount int64
		err = s.db.Model(&models.ProblemTag{}).
			Select("DISTINCT problem_id").
			Where("tag_id IN (?)", tagIDs).
			Count(&totalProblemCount).
			Error
		if err != nil {
			return nil, err
		}

		var solvedCount, wrongCount int64
		// 查询作答正确数量
		err := s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND knowledge_point_id in (?) AND status = ?", user.ID, courseKnowledgePointIDs, models.ProblemStatusSolved).
			Count(&solvedCount).
			Error
		if err != nil {
			return nil, err
		}

		// 查询作答错误数量
		err = s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND knowledge_point_id in (?) AND status = ?", user.ID, courseID, models.ProblemStatusFailed).
			Count(&wrongCount).
			Error
		if err != nil {
			return nil, err
		}

		// 进度
		var progress float64
		if totalProblemCount > 0 {
			progress = float64(solvedCount) / float64(totalProblemCount) * 100
		} else {
			progress = 0
		}

		// 正确率
		var correctRate float64
		totalResponses := solvedCount + wrongCount
		if totalResponses > 0 {
			correctRate = float64(solvedCount) / float64(totalResponses) * 100
		} else {
			correctRate = 0
		}

		result[i] = map[string]interface{}{
			"user_id":      user.ID,
			"student_id":   user.StudentID,
			"name":         user.Name,
			"solved_count": solvedCount,
			"wrong_count":  wrongCount,
			"correct_rate": correctRate,
			"progress":     progress,
		}
	}
	return result, nil
}

func (s *UserService) GetUserListByClass(classID uint) ([]map[string]interface{}, error) {
	var users []map[string]interface{}
	err := s.db.Model(&models.User{}).
		Select("id, name, username, student_id").
		Where("class_id = ?", classID).
		Find(&users).
		Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) ResetPassword(userID uint, password string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 手动设置密码并加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 使用Updates方法更新密码字段，直接使用加密后的密码
	return s.db.Model(&user).Updates(map[string]interface{}{
		"password": string(hashedPassword),
	}).Error
}

func (s *UserService) GetTryRecords(userID uint) ([]map[string]interface{}, error) {
	var records []map[string]interface{}
	err := s.db.Select("user_problems.id, user_problems.problem_id, problems.title_cn, problems.title, user_problems.status, user_problems.updated_at").
		Model(&models.UserProblem{}).
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Where("user_id = ?", userID).
		Scan(&records).
		Error
	if err != nil {
		return nil, err
	}

	return records, err
}
