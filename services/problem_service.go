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
		Select("problems.id, leetcode_id, title_slug, difficulty").
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
