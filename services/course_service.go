package services

import (
	"ai_teach_system/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type CourseService struct {
	db *gorm.DB
}

func NewCourseService(db *gorm.DB) *CourseService {
	return &CourseService{db: db}
}

type KnowledgePointInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ProblemCount int64  `json:"problem_count"`
	SolvedCount  int64  `json:"solved_count"`
	Tags         []Tag  `json:"tags"`
}

type Tag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// 能力分析报告
type SkillAnalysis struct {
	KnowledgePoint string  `json:"knowledge_point"`
	CorrectRate    float64 `json:"correct_rate"`
	TotalAttempts  int64   `json:"total_attempts"`
	CorrectCount   int64   `json:"correct_count"`
}

// 整体学习情况
type StudyOverview struct {
	TotalProblems     int64   `json:"total_problems"`
	AttemptedProblems int64   `json:"attempted_problems"`
	CorrectRate       float64 `json:"correct_rate"`
	WrongProblems     int64   `json:"wrong_problems"`
}

func (s *CourseService) GetCourseDetail(courseID, userID uint) (*models.Course, []KnowledgePointInfo, []SkillAnalysis, *StudyOverview, error) {
	var course models.Course
	if err := s.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil, errors.New("课程不存在")
		}
		return nil, nil, nil, nil, err
	}

	var points []models.KnowledgePoint
	if err := s.db.Model(&models.KnowledgePoint{}).Where("course_id = ?", courseID).Find(&points).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var pointInfos []KnowledgePointInfo
	for _, point := range points {
		// 获取知识点下的总题目数
		var problemCount int64
		s.db.Model(&models.Problem{}).
			Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
			Joins("JOIN tags ON problem_tag.tag_id = tags.id").
			Where("tags.knowledge_point_id = ?", point.ID).
			Distinct("problems.id").
			Count(&problemCount)

		// 获取用户已解决的题目数
		var solvedCount int64
		s.db.Model(&models.UserProblem{}).
			Joins("JOIN problems ON user_problems.problem_id = problems.id").
			Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
			Joins("JOIN tags ON problem_tag.tag_id = tags.id").
			Where("tags.knowledge_point_id = ? AND user_problems.user_id = ? AND user_problems.status = ?",
				point.ID, userID, models.ProblemStatusSolved).
			Distinct("user_problems.problem_id").
			Count(&solvedCount)

		var tags []models.Tag
		s.db.Model(&models.Tag{}).Where("knowledge_point_id = ?", point.ID).Find(&tags)

		tagInfos := make([]Tag, len(tags))
		for i, tag := range tags {
			tagInfos[i] = Tag{
				ID:   tag.ID,
				Name: tag.Name,
			}
		}

		pointInfos = append(pointInfos, KnowledgePointInfo{
			ID:           point.ID,
			Name:         point.Name,
			ProblemCount: problemCount,
			SolvedCount:  solvedCount,
			Tags:         tagInfos,
		})
	}

	// 获取能力分析报告
	var skillAnalysis []SkillAnalysis
	for _, point := range points {
		var totalAttempts, correctCount int64

		// 获取该知识点下的所有提交记录
		err := s.db.Model(&models.UserProblem{}).
			Select("COUNT(*) as total_attempts, COUNT(DISTINCT(problems.id)) as correct_count").
			Joins("JOIN problems ON user_problems.problem_id = problems.id").
			Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
			Joins("JOIN tags ON problem_tag.tag_id = tags.id").
			Where("tags.knowledge_point_id = ? AND user_problems.user_id = ? AND user_problems.status = ?", point.ID, userID, models.ProblemStatusSolved).
			Row().Scan(&totalAttempts, &correctCount)

		if err != nil {
			return nil, nil, nil, nil, err
		}

		var correctRate float64
		if totalAttempts > 0 {
			correctRate = float64(correctCount) / float64(totalAttempts) * 100
		}

		skillAnalysis = append(skillAnalysis, SkillAnalysis{
			KnowledgePoint: point.Name,
			CorrectRate:    correctRate,
			TotalAttempts:  totalAttempts,
			CorrectCount:   correctCount,
		})
	}

	// 获取整体学习情况
	var overview StudyOverview

	// 获取课程总题目数
	s.db.Model(&models.Problem{}).
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id").
		Joins("JOIN knowledge_points ON tags.knowledge_point_id = knowledge_points.id").
		Where("knowledge_points.course_id = ?", courseID).
		Distinct("problems.id").
		Count(&overview.TotalProblems)

	// 获取已尝试的题目数和正确率
	var correctCount int64
	s.db.Model(&models.UserProblem{}).
		Select("COUNT(*) as attempted_problems, COUNT(DISTINCT(problems.id)) as correct_count").
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id").
		Joins("JOIN knowledge_points ON tags.knowledge_point_id = knowledge_points.id").
		Where("knowledge_points.course_id = ? AND user_problems.user_id = ? AND user_problems.status = ?", courseID, userID, models.ProblemStatusSolved).
		Row().Scan(&overview.AttemptedProblems, &correctCount)

	if overview.AttemptedProblems > 0 {
		overview.CorrectRate = float64(correctCount) / float64(overview.AttemptedProblems) * 100
	}

	// 获取错题数
	overview.WrongProblems = overview.AttemptedProblems - correctCount

	return &course, pointInfos, skillAnalysis, &overview, nil
}

func (s *CourseService) GetKnowledgePoints(courseID uint) ([]map[string]interface{}, error) {
	var points []map[string]interface{}
	err := s.db.Model(&models.KnowledgePoint{}).Select("name").Where("course_id = ?", courseID).Find(&points).Error
	if err != nil {
		return nil, err
	}
	return points, nil
}

func (s *CourseService) GetCourseList() ([]string, error) {
	var courseNames []string
	err := s.db.Model(&models.Course{}).Select("name").Find(&courseNames).Error
	if err != nil {
		return nil, err
	}
	return courseNames, nil
}

func (s *CourseService) GetUserListByCourseAndClass(classID uint, courseID uint) ([]map[string]interface{}, error) {
	// 查询当前班级的用户列表
	var userList []models.User
	err := s.db.Model(&models.User{}).Where("class_id = ?", classID).Find(&userList).Error
	if err != nil {
		return nil, err
	}

	// 查询每个用户的答题数据
	result := make([]map[string]interface{}, len(userList))
	for _, user := range userList {
		var solvedCount, wrongCount int64
		// 查询作答正确数量
		err := s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND course_id = ? AND status = ?", user.ID, courseID, models.ProblemStatusSolved).
			Count(&solvedCount).
			Error
		if err != nil {
			return nil, err
		}

		// 查询作答错误数量
		err = s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND course_id = ? AND status = ?", user.ID, courseID, models.ProblemStatusTried).
			Count(&wrongCount).
			Error
		if err != nil {
			return nil, err
		}

		// 正确率
		correctRate := float64(solvedCount) / float64(solvedCount+wrongCount) * 100

		// 进度
		var totalProblemCount int64
		err = s.db.Model(&models.CourseProblem{}).
			Where("course_id = ?", courseID).
			Count(&totalProblemCount).
			Error
		progress := float64(solvedCount) / float64(totalProblemCount) * 100
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"student_id":   user.StudentID,
			"name":         user.Name,
			"solved_count": solvedCount,
			"wrong_count":  wrongCount,
			"correct_rate": correctRate,
			"progress":     progress,
		})
	}
	return result, nil
}

func (s *CourseService) AddCourse(courseName string, pointNames []string) (map[string]interface{}, error) {
	// 检查课程是否已存在
	var courseCount int64
	err := s.db.Model(&models.Course{}).Where("name = ?", courseName).Count(&courseCount).Error
	if err != nil {
		return nil, err
	}
	if courseCount > 0 {
		return nil, fmt.Errorf("课程: %s已存在", courseName)
	}

	// 开启事务
	course := models.Course{Name: courseName}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建课程
		err = s.db.Model(&models.Course{}).Create(&course).Error
		if err != nil {
			return err
		}

		// 创建关联的知识点
		knowledgePoints := make([]models.KnowledgePoint, len(pointNames))
		for i, name := range pointNames {
			knowledgePoint := models.KnowledgePoint{
				Name:     name,
				CourseID: course.ID,
				Course:   course,
			}
			knowledgePoints[i] = knowledgePoint
		}
		err = s.db.Model(&models.KnowledgePoint{}).Create(&knowledgePoints).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 重新加载课程数据，包括知识点
	if err := s.db.Preload("Points").First(&course, course.ID).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"course_id":   course.ID,
		"course_name": course.Name,
		"points":      course.Points,
	}, nil
}
