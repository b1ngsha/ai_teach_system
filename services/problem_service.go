package services

import (
	"ai_teach_system/models"
	"fmt"

	"gorm.io/gorm"
)

type ProblemService struct {
	db *gorm.DB
}

func NewProblemService(db *gorm.DB) *ProblemService {
	return &ProblemService{db: db}
}

func (s *ProblemService) GetCourseProblemList(courseID, userID uint, difficulty models.ProblemDifficulty, knowledgePointID uint, tagID uint) ([]map[string]interface{}, error) {
	// 获取课程所关联的知识点id列表
	var knowledgePointIDs []uint
	query := s.db.Select("id").
		Model(&models.KnowledgePoint{}).
		Where("course_id = ?", courseID)

	if knowledgePointID != 0 {
		query = query.Where("id = ?", knowledgePointID)
	}

	err := query.Scan(&knowledgePointIDs).Error
	if err != nil {
		return nil, err
	}

	// 获取知识点关联的标签
	var tagIDs []uint
	if tagID != 0 {
		tagIDs = append(tagIDs, tagID)
	} else {
		err = s.db.Model(&models.KnowledgePointTag{}).
			Select("tag_id").
			Where("knowledge_point_id IN ?", knowledgePointIDs).
			Scan(&tagIDs).Error
		if err != nil {
			return nil, err
		}
	}

	// 获取标签关联的题目
	var problems []map[string]interface{}
	query = s.db.Model(&models.Problem{}).
		Select("DISTINCT problems.id, leetcode_id, title_slug, title_cn, difficulty").
		Joins("JOIN problem_tags ON problems.id = problem_tags.problem_id").
		Joins("JOIN tags ON problem_tags.tag_id = tags.id").
		Where("tags.id IN ?", tagIDs)

	if difficulty != "" {
		query = query.Where("problems.difficulty = ?", difficulty)
	}

	err = query.Scan(&problems).Error
	if err != nil {
		return nil, err
	}

	// 获取用户题目状态
	var problemStatus []map[string]interface{}
	err = s.db.Model(&models.UserProblem{}).
		Select("problem_id, status").
		Where("user_id = ?", userID).
		Scan(&problemStatus).Error
	if err != nil {
		return nil, err
	}

outer:
	for _, problem := range problems {
		for _, status := range problemStatus {
			problemID := status["problem_id"].(uint64)
			status := status["status"].(string)
			if problemID == problem["id"] {
				problem["status"] = status
				continue outer
			}
		}
		problem["status"] = models.ProblemStatusUntried
	}
	return problems, nil
}

func (s *ProblemService) GetProblemDetail(problemID uint) (map[string]interface{}, error) {
	var problem models.Problem
	err := s.db.Preload("Tags").Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return nil, err
	}

	problemMap := map[string]interface{}{
		"id":           problem.ID,
		"title":        problem.Title,
		"title_cn":     problem.TitleCn,
		"title_slug":   problem.TitleSlug,
		"difficulty":   problem.Difficulty,
		"content":      problem.Content,
		"content_cn":   problem.ContentCn,
		"sample_cases": problem.SampleTestcases,
		"tags":         problem.Tags,
		"is_custom":    problem.IsCustom,
	}

	// 获取关联的知识点信息
	var knowledgePointInfo []map[string]interface{}
	err = s.db.Model(&models.KnowledgePoint{}).
		Select("DISTINCT knowledge_points.id, knowledge_points.name, knowledge_points.course_id").
		Joins("JOIN knowledge_point_tags ON knowledge_point_tags.knowledge_point_id = knowledge_points.id").
		Joins("JOIN tags ON tags.id = knowledge_point_tags.tag_id").
		Joins("JOIN problem_tags ON problem_tags.tag_id = tags.id").
		Where("problem_tags.problem_id = ?", problemID).
		Scan(&knowledgePointInfo).
		Error
	if err != nil {
		return nil, err
	}

	problemMap["knowledge_point_info"] = knowledgePointInfo
	return problemMap, nil
}

func (s *ProblemService) SetKnowledgePointTags(knowledgePointID uint, tagIDs []uint) (map[string]interface{}, error) {
	// 先查出原来选中的标签
	var existTagIDs []uint
	err := s.db.Select("tag_id").
		Model(&models.KnowledgePointTag{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Scan(&existTagIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 存到map里提高查询效率
	existTagIDMap := make(map[uint]int)
	for _, tagID := range existTagIDs {
		existTagIDMap[tagID] = 1
	}
	newTagIDMap := make(map[uint]int)
	for _, id := range tagIDs {
		newTagIDMap[id] = 1
	}

	// 考虑三种情况:
	// 新旧集合中都存在的保持不变
	// 新集合中存在旧集合中不存在则新增
	// 旧集合中存在新集合中不存在则删除
	createList := make([]uint, 0)
	deleteList := make([]uint, 0)

	// 找出需要新增的标签
	for _, id := range tagIDs {
		if _, exist := existTagIDMap[id]; !exist {
			createList = append(createList, id)
		}
	}

	// 找出需要删除的标签
	for _, id := range existTagIDs {
		if _, exist := newTagIDMap[id]; !exist {
			deleteList = append(deleteList, id)
		}
	}

	// 开事务处理创建和删除操作
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if len(createList) > 0 {
			knowledgePointTags := make([]models.KnowledgePointTag, 0, len(createList))
			for _, tagID := range createList {
				knowledgePointTags = append(knowledgePointTags, models.KnowledgePointTag{
					KnowledgePointID: knowledgePointID,
					TagID:            tagID,
				})
			}

			if err := tx.Create(&knowledgePointTags).Error; err != nil {
				return err
			}
		}

		if len(deleteList) > 0 {
			if err := tx.Where("knowledge_point_id = ? AND tag_id IN ?", knowledgePointID, deleteList).
				Delete(&models.KnowledgePointTag{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 查询更新后的总标签数
	var totalCount int64
	err = s.db.Model(&models.KnowledgePointTag{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Count(&totalCount).
		Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"knowledge_point_id": knowledgePointID,
		"total_count":        totalCount,
		"added_count":        len(createList),
		"removed_count":      len(deleteList),
	}, nil
}

func (s *ProblemService) GetKnowledgePointProblems(userID, knowledgePointID uint) ([]map[string]interface{}, error) {
	// 查询该知识点下的所有标签ID
	var tagIDs []uint
	err := s.db.Model(&models.KnowledgePointTag{}).
		Select("tag_id").
		Where("knowledge_point_id = ?", knowledgePointID).
		Find(&tagIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 根据标签ID查询题目ID
	var problemIDs []uint
	err = s.db.Model(&models.ProblemTag{}).
		Select("problem_id").
		Where("tag_id IN (?)", tagIDs).
		Find(&problemIDs).
		Error

	// 根据ID查询具体信息
	var problemInfos []map[string]interface{}
	err = s.db.Model(&models.Problem{}).
		Select("id, title, title_cn, content, content_cn, difficulty").
		Where("id in (?)", problemIDs).
		Find(&problemInfos).
		Error
	if err != nil {
		return nil, err
	}

	// 如果当前用户为学生，则需要查询这些题目的完成状态
	var user models.User
	err = s.db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	if user.Role == models.RoleUser {
		var userProblems []models.UserProblem
		err = s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND problem_id IN (?)", userID, problemIDs).
			Find(&userProblems).
			Error
		if err != nil {
			return nil, err
		}

		// 将用户答题状态添加到题目信息中
		userProblemMap := make(map[uint]models.UserProblem)
		for _, userProblem := range userProblems {
			userProblemMap[userProblem.ProblemID] = userProblem
		}

		for _, info := range problemInfos {
			problemID := info["id"].(uint)
			userProblem, exists := userProblemMap[problemID]
			if exists {
				info["status"] = userProblem.Status
			} else {
				info["status"] = models.ProblemStatusUntried
			}
		}
	}
	return problemInfos, nil
}

func (s *ProblemService) GetProblemList(difficulty string, tagID uint) ([]map[string]interface{}, error) {
	var problems []map[string]interface{}
	query := s.db.Model(&models.Problem{}).
		Select("problems.id, leetcode_id, title_slug, title_cn, difficulty, tags.id AS tag_id, tags.name AS tag_name").
		Joins("JOIN problem_tags ON problems.id = problem_tags.problem_id").
		Joins("JOIN tags ON problem_tags.tag_id = tags.id")

	if tagID != 0 {
		query = query.Where("tags.id = ?", tagID)
	}

	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	err := query.Scan(&problems).Error
	if err != nil {
		return nil, err
	}

	return problems, nil
}

func (s *ProblemService) GetAllTags() ([]models.Tag, error) {
	var tags []models.Tag
	if err := s.db.Order("name").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("获取标签列表失败: %v", err)
	}
	return tags, nil
}

func (s *ProblemService) CreateCustomProblem(problem *models.Problem, tagIDs []uint) (*models.Problem, error) {
	// 验证标签是否存在
	var count int64
	err := s.db.Model(&models.Tag{}).Where("id IN ?", tagIDs).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("验证标签失败: %v", err)
	}
	if int(count) != len(tagIDs) {
		return nil, fmt.Errorf("部分标签不存在")
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建题目
		if err := tx.Create(problem).Error; err != nil {
			return fmt.Errorf("创建题目失败: %v", err)
		}

		// 2. 创建题目与标签的关联
		for _, tagID := range tagIDs {
			if err := tx.Create(&models.ProblemTag{
				ProblemID: problem.ID,
				TagID:     tagID,
			}).Error; err != nil {
				return fmt.Errorf("创建题目标签关联失败: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return problem, nil
}
