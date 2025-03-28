package services

import (
	"ai_teach_system/models"

	"gorm.io/gorm"
)

type ProblemService struct {
	db *gorm.DB
}

func NewProblemService(db *gorm.DB) *ProblemService {
	return &ProblemService{db: db}
}

func (s *ProblemService) GetCourseProblemList(courseID, userID uint, difficulty models.ProblemDifficulty, knowledgePointID uint) ([]map[string]interface{}, error) {
	var problems []map[string]interface{}

	// 获取课程所关联的知识点id列表
	var knowledgePointIDs []uint
	err := s.db.Select("id").
		Model(&models.KnowledgePoint{}).
		Where("course_id = ?", courseID).
		Scan(&knowledgePointIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 获取所有关联的知识点下的课程id列表
	var problemIDs []uint
	err = s.db.Select("problem_id").
		Model(&models.KnowledgePointProblems{}).
		Where("knowledge_point_id IN (?)", knowledgePointIDs).
		Scan(&problemIDs).
		Error
	if err != nil {
		return nil, err
	}

	query := s.db.Model(&models.Problem{}).
		Select("problems.id, leetcode_id, title_slug, title_cn, difficulty").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id").
		Where("problems.id IN (?)", problemIDs)
	if knowledgePointID != 0 {
		query = query.Where("knowledge_point_id = ?", knowledgePointID)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err = query.Find(&problems).Error
	if err != nil {
		return nil, err
	}

	problemStatus := make([]map[string]interface{}, 0, 10)
	err = s.db.Model(&models.UserProblem{}).
		Select("problem_id, status").
		Where("user_id = ?", userID).
		Find(&problemStatus).Error
	if err != nil {
		return nil, err
	}

outer:
	for _, problem := range problems {
		for _, status := range problemStatus {
			problemID := status["problem_id"].(uint)
			status := status["status"].(models.ProblemStatus)
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
		"title_slug":   problem.TitleSlug,
		"difficulty":   problem.Difficulty,
		"content":      problem.Content,
		"sample_cases": problem.SampleTestcases,
		"tags":         problem.Tags,
	}

	var knowledgePointInfo []map[string]interface{}
	err = s.db.Model(&models.KnowledgePoint{}).
		Select("id, name").
		Joins("JOIN knowledge_point_problems ON knowledge_point_problems.problem_id = problems.id").
		Where("problems.id = ?", problemID).
		Scan(&knowledgePointInfo).
		Error
	if err != nil {
		return nil, err
	}

	problemMap["knowledge_point_info"] = knowledgePointInfo
	return problemMap, nil
}

func (s *ProblemService) SetKnowledgePointProblems(knowledgePointID uint, problemIDs []uint) (map[string]interface{}, error) {
	// 先查出原来选中的题目
	var existProblemIDs []uint
	err := s.db.Select("problem_id").
		Model(&models.KnowledgePointProblems{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Scan(&existProblemIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 存到map里提高查询效率
	existProblemIDMap := make(map[uint]int)
	for _, problemID := range existProblemIDs {
		existProblemIDMap[problemID] = 1
	}
	newProblemIDMap := make(map[uint]int)
	for _, id := range problemIDs {
		newProblemIDMap[id] = 1
	}

	// 考虑三种情况:
	// 新旧集合中都存在的保持不变
	// 新集合中存在旧集合中不存在则新增
	// 旧集合中存在新集合中不存在则删除
	createList := make([]uint, 0)
	deleteList := make([]uint, 0)

	// 找出需要新增的题目
	for _, id := range problemIDs {
		if _, exist := existProblemIDMap[id]; !exist {
			createList = append(createList, id)
		}
	}

	// 找出需要删除的题目
	for _, id := range existProblemIDs {
		if _, exist := newProblemIDMap[id]; !exist {
			deleteList = append(deleteList, id)
		}
	}

	// 开事务处理创建和删除操作
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if len(createList) > 0 {
			knowledgePointProblems := make([]models.KnowledgePointProblems, 0, len(createList))
			for _, problemID := range createList {
				knowledgePointProblems = append(knowledgePointProblems, models.KnowledgePointProblems{
					KnowledgePointID: knowledgePointID,
					ProblemID:        problemID,
				})
			}

			if err := tx.Create(&knowledgePointProblems).Error; err != nil {
				return err
			}
		}

		if len(deleteList) > 0 {
			if err := tx.Where("knowledge_point_id = ? AND problem_id IN ?", knowledgePointID, deleteList).
				Delete(&models.KnowledgePointProblems{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 查询更新后的总题目数
	var totalCount int64
	err = s.db.Model(&models.KnowledgePointProblems{}).
		Where("knowledge_point_id = ?", knowledgePointID).
		Count(&totalCount).
		Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"knowledge_point_id": knowledgePointID,
		"total_count":        int(totalCount),
		"added_count":        len(createList),
		"removed_count":      len(deleteList),
	}, nil
}

func (s *ProblemService) GetKnowledgePointProblems(knowledgePointID uint) ([]map[string]interface{}, error) {
	// 查询该知识点下的所有题目ID
	var problemIDs []uint
	err := s.db.Model(&models.KnowledgePointProblems{}).
		Select("problem_id").
		Where("knowledge_point_id = ?", knowledgePointID).
		Find(&problemIDs).
		Error
	if err != nil {
		return nil, err
	}

	// 根据ID查询具体信息
	var problemInfos []map[string]interface{}
	err = s.db.Model(&models.Problem{}).
		Select("id, title, content").
		Where("id in (?)", problemIDs).
		Find(&problemInfos).
		Error
	if err != nil {
		return nil, err
	}
	return problemInfos, nil
}

func (s *ProblemService) GetProblemList(difficulty string, tagID uint) ([]map[string]interface{}, error) {
	var problems []map[string]interface{}
	query := s.db.Model(&models.Problem{}).
		Select("problems.id, leetcode_id, title_slug, title_cn, difficulty, tags.name AS tag_name").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id")

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
