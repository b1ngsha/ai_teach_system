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

func (s *CourseService) GetCourseDetail(courseID, userID uint) (*models.Course, []KnowledgePointInfo, error) {
	var course models.Course
	if err := s.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("课程不存在")
		}
		return nil, nil, err
	}

	var points []models.KnowledgePoint
	if err := s.db.Where("course_id = ?", courseID).Find(&points).Error; err != nil {
		return nil, nil, err
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

	return &course, pointInfos, nil
}
