package models

import "gorm.io/gorm"

type KnowledgePoint struct {
	gorm.Model
	Name     string `json:"name" gorm:"type:varchar(255);not null"`
	CourseID uint   `json:"course_id" gorm:"not null"`
	Course   Course `json:"-" gorm:"foreignKey:CourseID"`
	Tags     []Tag  `json:"tags" gorm:"foreignKey:KnowledgePointID"`
}
