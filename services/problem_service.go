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

func (s *ProblemService) GetProblemList(userID uint, difficulty models.ProblemDifficulty, knowledgePointID uint) ([]map[string]interface{}, error) {
	problems := make([]map[string]interface{}, 0, 10)
	query := s.db.Model(&models.Problem{}).
		Select("problems.id, leetcode_id, title_slug, title_cn, difficulty").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id")
	if knowledgePointID != 0 {
		query = query.Where("knowledge_point_id = ?", knowledgePointID)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Find(&problems).Error
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
	tags, ok := problemMap["tags"].([]models.Tag)
	knowledgePointIDs := make([]uint, len(tags))
	if ok {
		for i, tag := range tags {
			knowledgePointIDs[i] = tag.KnowledgePointID
		}
	}

	// 去重
	var uniqueKnowledgePointIds []uint
	uniqueMap := make(map[uint]bool)
	for _, id := range knowledgePointIDs {
		if !uniqueMap[id] {
			uniqueMap[id] = true
			uniqueKnowledgePointIds = append(uniqueKnowledgePointIds, id)
		}
	}

	var knowledgePointInfo []map[string]interface{}
	err = s.db.Model(&models.KnowledgePoint{}).Select("id, name").Where("id IN (?)", uniqueKnowledgePointIds).Find(&knowledgePointInfo).Error

	if err != nil {
		return nil, err
	}

	problemMap["knowledge_point_info"] = knowledgePointInfo
	return problemMap, nil
}
