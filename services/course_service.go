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
	err := s.db.Model(&models.KnowledgePoint{}).Select("id, name").Where("course_id = ?", courseID).Find(&points).Error
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

func (s *CourseService) SetCourseClasses(courseID uint, classIDs []uint) (map[string]interface{}, error) {
	// 先查出原来的班级
	var existClassIDs []uint
	err := s.db.Select("class_id").
		Model(&models.CourseClasses{}).
		Where("course_id = ?", courseID).
		Scan(&existClassIDs).
		Error
	if err != nil {
		return nil, err
	}
	// 存到map里提高查询效率
	existClassIDMap := make(map[uint]int)
	for _, classID := range existClassIDs {
		existClassIDMap[classID] = 1
	}
	newClassIDMap := make(map[uint]int)
	for _, id := range classIDs {
		newClassIDMap[id] = 1
	}
	// 考虑三种情况:
	// 新旧集合中都存在的保持不变
	// 新集合中存在旧集合中不存在则新增
	// 旧集合中存在新集合中不存在则删除
	createList := make([]uint, 0)
	deleteList := make([]uint, 0)
	// 找出需要新增的班级
	for _, id := range classIDs {
		if _, exist := existClassIDMap[id]; !exist {
			createList = append(createList, id)
		}
	}
	// 找出需要删除的班级
	for _, id := range existClassIDs {
		if _, exist := newClassIDMap[id]; !exist {
			deleteList = append(deleteList, id)
		}
	}
	// 开事务处理创建和删除操作
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if len(createList) > 0 {
			courseClasses := make([]models.CourseClasses, 0, len(createList))
			for _, classID := range createList {
				courseClasses = append(courseClasses, models.CourseClasses{
					CourseID: courseID,
					ClassID:  classID,
				})
			}
			if err := tx.Create(&courseClasses).Error; err != nil {
				return err
			}
		}
		if len(deleteList) > 0 {
			if err := tx.Where("course_id = ? AND class_id IN?", courseID, deleteList).
				Delete(&models.CourseClasses{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 查询更新后的总班级数
	var totalCount int64
	err = s.db.Model(&models.CourseClasses{}).
		Where("course_id = ?", courseID).
		Count(&totalCount).
		Error
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"course_id":     courseID,
		"total_count":   int(totalCount),
		"added_count":   len(createList),
		"removed_count": len(deleteList),
	}, nil
}

func (s *CourseService) GetCourseClasses(courseID uint) ([]map[string]interface{}, error) {
	// 查询该课程下的所有班级ID
	var classIDs []uint
	err := s.db.Model(&models.CourseClasses{}).
		Select("class_id").
		Where("course_id = ?", courseID).
		Find(&classIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 根据课程ID查询课程信息
	var classInfos []map[string]interface{}
	err = s.db.Model(&models.Class{}).
		Select("id, name").
		Where("id IN (?)", classIDs).
		Find(&classInfos).
		Error
	if err != nil {
		return nil, err
	}
	return classInfos, nil
}
