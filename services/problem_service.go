package services

import (
	"ai_teach_system/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	// 验证知识点是否存在
	var knowledgePoint models.KnowledgePoint
	if err := s.db.First(&knowledgePoint, knowledgePointID).Error; err != nil {
		return nil, fmt.Errorf("知识点不存在: %v", err)
	}

	// 验证标签是否存在
	var count int64
	err := s.db.Model(&models.Tag{}).Where("id IN ?", tagIDs).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("验证标签失败: %v", err)
	}
	if int(count) != len(tagIDs) {
		return nil, fmt.Errorf("部分标签不存在")
	}

	// 先查出原来选中的标签
	var existTagIDs []uint
	err = s.db.Select("tag_id").
		Model(&models.KnowledgePointTag{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Scan(&existTagIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 存到map里提高查询效率
	existTagIDMap := make(map[uint]bool)
	for _, tagID := range existTagIDs {
		existTagIDMap[tagID] = true
	}
	newTagIDMap := make(map[uint]bool)
	for _, id := range tagIDs {
		newTagIDMap[id] = true
	}

	// 找出需要新增和删除的标签
	var toAddTags, toRemoveTags []uint
	for _, id := range tagIDs {
		if !existTagIDMap[id] {
			toAddTags = append(toAddTags, id)
		}
	}
	for _, id := range existTagIDs {
		if !newTagIDMap[id] {
			toRemoveTags = append(toRemoveTags, id)
		}
	}

	// 开启事务处理所有操作
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 处理标签关联
		if len(toAddTags) > 0 {
			knowledgePointTags := make([]models.KnowledgePointTag, 0, len(toAddTags))
			for _, tagID := range toAddTags {
				knowledgePointTags = append(knowledgePointTags, models.KnowledgePointTag{
					KnowledgePointID: knowledgePointID,
					TagID:            tagID,
				})
			}

			if err := tx.Create(&knowledgePointTags).Error; err != nil {
				return fmt.Errorf("创建标签关联失败: %v", err)
			}

			// 获取新增标签关联的所有题目
			var problemIDs []uint
			err = tx.Model(&models.ProblemTag{}).
				Select("DISTINCT problem_id").
				Where("tag_id IN ?", toAddTags).
				Scan(&problemIDs).Error
			if err != nil {
				return fmt.Errorf("获取标签关联题目失败: %v", err)
			}

			// 创建知识点与题目的关联
			if len(problemIDs) > 0 {
				knowledgePointProblems := make([]models.KnowledgePointProblem, 0, len(problemIDs))
				for _, problemID := range problemIDs {
					knowledgePointProblems = append(knowledgePointProblems, models.KnowledgePointProblem{
						KnowledgePointID: knowledgePointID,
						ProblemID:        problemID,
					})
				}

				// 使用 ON CONFLICT DO NOTHING 避免重复插入
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&knowledgePointProblems).Error; err != nil {
					return fmt.Errorf("创建题目关联失败: %v", err)
				}
			}
		}

		if len(toRemoveTags) > 0 {
			// 删除标签关联
			if err := tx.Where("knowledge_point_id = ? AND tag_id IN ?", knowledgePointID, toRemoveTags).
				Delete(&models.KnowledgePointTag{}).Error; err != nil {
				return fmt.Errorf("删除标签关联失败: %v", err)
			}

			// 获取要删除的标签独有的题目（不被其他保留的标签关联的题目）
			var problemsToRemove []uint
			err = tx.Model(&models.ProblemTag{}).
				Select("DISTINCT pt1.problem_id").
				Table("problem_tags pt1").
				Where("pt1.tag_id IN ? AND NOT EXISTS (SELECT 1 FROM problem_tags pt2 WHERE pt2.problem_id = pt1.problem_id AND pt2.tag_id IN ?)",
					toRemoveTags, tagIDs).
				Scan(&problemsToRemove).Error
			if err != nil {
				return fmt.Errorf("获取要删除的题目失败: %v", err)
			}

			// 删除知识点与这些题目的关联
			if len(problemsToRemove) > 0 {
				if err := tx.Where("knowledge_point_id = ? AND problem_id IN ?", knowledgePointID, problemsToRemove).
					Delete(&models.KnowledgePointProblem{}).Error; err != nil {
					return fmt.Errorf("删除题目关联失败: %v", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回更新后的统计信息
	var tagCount, problemCount int64
	err = s.db.Model(&models.KnowledgePointTag{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Count(&tagCount).Error
	if err != nil {
		return nil, fmt.Errorf("统计标签数量失败: %v", err)
	}

	err = s.db.Model(&models.KnowledgePointProblem{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Count(&problemCount).Error
	if err != nil {
		return nil, fmt.Errorf("统计题目数量失败: %v", err)
	}

	return map[string]interface{}{
		"knowledge_point_id": knowledgePointID,
		"total_tags":         tagCount,
		"total_problems":     problemCount,
		"added_tags":         len(toAddTags),
		"removed_tags":       len(toRemoveTags),
	}, nil
}

func (s *ProblemService) SetKnowledgePointProblems(knowledgePointID uint, problemIDs []uint) (map[string]interface{}, error) {
	// 验证知识点是否存在
	var knowledgePoint models.KnowledgePoint
	if err := s.db.First(&knowledgePoint, knowledgePointID).Error; err != nil {
		return nil, fmt.Errorf("知识点不存在: %v", err)
	}

	// 验证题目是否存在
	var count int64
	err := s.db.Model(&models.Problem{}).Where("id IN ?", problemIDs).Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("验证题目失败: %v", err)
	}
	if int(count) != len(problemIDs) {
		return nil, fmt.Errorf("部分题目不存在")
	}

	// 获取当前知识点关联的题目
	var existingProblemIDs []uint
	err = s.db.Model(&models.KnowledgePointProblem{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Pluck("problem_id", &existingProblemIDs).Error
	if err != nil {
		return nil, fmt.Errorf("获取现有关联失败: %v", err)
	}

	// 将现有题目ID和新题目ID转换为map，方便比较
	existingMap := make(map[uint]bool)
	for _, id := range existingProblemIDs {
		existingMap[id] = true
	}
	newMap := make(map[uint]bool)
	for _, id := range problemIDs {
		newMap[id] = true
	}

	// 找出需要添加和删除的题目ID
	var toAdd, toRemove []uint
	for _, id := range problemIDs {
		if !existingMap[id] {
			toAdd = append(toAdd, id)
		}
	}
	for _, id := range existingProblemIDs {
		if !newMap[id] {
			toRemove = append(toRemove, id)
		}
	}

	// 使用事务处理添加和删除操作
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 添加新关联
		if len(toAdd) > 0 {
			associations := make([]models.KnowledgePointProblem, 0, len(toAdd))
			for _, problemID := range toAdd {
				associations = append(associations, models.KnowledgePointProblem{
					KnowledgePointID: knowledgePointID,
					ProblemID:        problemID,
				})
			}
			if err := tx.Create(&associations).Error; err != nil {
				return fmt.Errorf("创建关联失败: %v", err)
			}
		}

		// 删除旧关联
		if len(toRemove) > 0 {
			if err := tx.Where("knowledge_point_id = ? AND problem_id IN ?", knowledgePointID, toRemove).
				Delete(&models.KnowledgePointProblem{}).Error; err != nil {
				return fmt.Errorf("删除关联失败: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回更新后的统计信息
	var totalCount int64
	err = s.db.Model(&models.KnowledgePointProblem{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Count(&totalCount).Error
	if err != nil {
		return nil, fmt.Errorf("统计关联数量失败: %v", err)
	}

	return map[string]interface{}{
		"knowledge_point_id": knowledgePointID,
		"total_problems":     totalCount,
		"added_count":        len(toAdd),
		"removed_count":      len(toRemove),
	}, nil
}

func (s *ProblemService) GetKnowledgePointProblems(knowledgePointID uint) ([]map[string]interface{}, error) {
	var problems []map[string]interface{}
	err := s.db.Model(&models.Problem{}).
		Select("problems.id, problems.title, problems.title_cn, problems.difficulty").
		Joins("JOIN knowledge_point_problems ON problems.id = knowledge_point_problems.problem_id").
		Where("knowledge_point_problems.knowledge_point_id = ?", knowledgePointID).
		Scan(&problems).Error
	if err != nil {
		return nil, fmt.Errorf("获取知识点题目失败: %v", err)
	}

	return problems, nil
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
