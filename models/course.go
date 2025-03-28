package models

import "gorm.io/gorm"

type Course struct {
	gorm.Model
	Name     string           `json:"name" gorm:"unique;not null"`
	Points   []KnowledgePoint `json:"points,omitempty" gorm:"foreignKey:CourseID"`
}
