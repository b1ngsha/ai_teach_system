package services

import (
	"ai_teach_system/models"

	"gorm.io/gorm"
)

type ClassService struct {
	db *gorm.DB
}

func NewClassService(db *gorm.DB) *ClassService {
	return &ClassService{db: db}
}

func (s *ClassService) GetClassList() ([]string, error) {
	var classNames []string
	err := s.db.Model(&models.Class{}).Select("name").Find(&classNames).Error
	if err != nil {
		return nil, err
	}
	return classNames, nil
}
