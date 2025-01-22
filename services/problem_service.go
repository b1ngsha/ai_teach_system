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

func (s *ProblemService) GetProblemList(difficulty models.ProblemDifficulty, knowledgePointID uint) ([]map[string]interface{}, error) {
	response := make([]map[string]interface{}, 0, 10)
	query := s.db.Model(&models.Problem{}).
		Select("leetcode_id, title_slug, difficulty, status").
		Joins("JOIN user_problems ON problems.id = user_problems.problem_id").
		Joins("JOIN problem_tag ON problems.id = problem_tag.problem_id").
		Joins("JOIN tags ON problem_tag.tag_id = tags.id")
	if knowledgePointID != 0 {
		query = query.Where("knowledge_point_id = ?", knowledgePointID)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Find(&response).Error

	for index, item := range response {
		response[index]["status"] = string(item["status"].([]uint8))
	}
	if err != nil {
		return nil, err
	}
	return response, nil
}
