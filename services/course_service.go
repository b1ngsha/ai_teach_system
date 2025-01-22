package services

import (
	"ai_teach_system/models"
	"errors"

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
	ProblemCount int64  `json:"problem_count"` // 知识点下的总题目数
	SolvedCount  int64  `json:"solved_count"`  // 用户已解决的题目数
	Tags         []Tag  `json:"tags"`
}

type Tag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// 能力分析报告
type SkillAnalysis struct {
	KnowledgePoint string  `json:"knowledge_point"` // 知识点名称
	CorrectRate    float64 `json:"correct_rate"`    // 正确率
	TotalAttempts  int64   `json:"total_attempts"`  // 总尝试次数
	CorrectCount   int64   `json:"correct_count"`   // 正确次数
}

// 整体学习情况
type StudyOverview struct {
	TotalProblems     int64   `json:"total_problems"`     // 总题目数
	AttemptedProblems int64   `json:"attempted_problems"` // 已尝试题目数
	CorrectRate       float64 `json:"correct_rate"`       // 整体正确率
	WrongProblems     int64   `json:"wrong_problems"`     // 错题数
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
	if err := s.db.Where("course_id = ?", courseID).Find(&points).Error; err != nil {
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
			Distinct().
			Count(&problemCount)

		// 获取用户已解决的题目数
		var solvedCount int64
		s.db.Model(&models.Problem{}).
			Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
			Joins("JOIN tags ON problem_tag.tag_id = tags.id").
			Joins("JOIN user_problems ON problems.id = user_problems.problem_id").
			Where("tags.knowledge_point_id = ? AND user_problems.user_id = ? AND user_problems.status = ?",
				point.ID, userID, models.ProblemStatusSolved).
			Distinct().
			Count(&solvedCount)

		var tags []models.Tag
		s.db.Where("knowledge_point_id = ?", point.ID).Find(&tags)

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
			Select("COUNT(*) as total_attempts, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as correct_count", models.ProblemStatusSolved).
			Joins("JOIN problems ON user_problems.problem_id = problems.id").
			Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
			Joins("JOIN tags ON problem_tag.tag_id = tags.id").
			Where("tags.knowledge_point_id = ? AND user_problems.user_id = ?", point.ID, userID).
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
		Distinct().
		Count(&overview.TotalProblems)

	// 获取已尝试的题目数和正确率
	var correctCount int64
	s.db.Model(&models.UserProblem{}).
		Select("COUNT(DISTINCT problems.id) as attempted_problems, SUM(CASE WHEN user_problems.status = ? THEN 1 ELSE 0 END) as correct_count", models.ProblemStatusSolved).
		Joins("JOIN problems ON user_problems.problem_id = problems.id").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id").
		Joins("JOIN knowledge_points ON tags.knowledge_point_id = knowledge_points.id").
		Where("knowledge_points.course_id = ? AND user_problems.user_id = ?", courseID, userID).
		Row().Scan(&overview.AttemptedProblems, &correctCount)

	if overview.AttemptedProblems > 0 {
		overview.CorrectRate = float64(correctCount) / float64(overview.AttemptedProblems) * 100
	}

	// 获取错题数
	overview.WrongProblems = overview.AttemptedProblems - correctCount

	return &course, pointInfos, skillAnalysis, &overview, nil
}
