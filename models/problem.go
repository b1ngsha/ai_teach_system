package models

import (
	"gorm.io/gorm"
)

type Problem struct {
	gorm.Model
	LeetcodeID      int              `json:"leetcode_id" gorm:"unique;not null"`
	Title           string           `json:"title" gorm:"not null"`
	TitleSlug       string           `json:"title_slug" gorm:"not null"`
	Difficulty      string           `json:"difficulty" gorm:"type:ENUM('Easy', 'Medium', 'Hard')"`
	Content         string           `json:"content" gorm:"type:text"`
	SampleTestcases string           `json:"sample_testcases" gorm:"type:text"`
	Tags            []Tag            `json:"tags" gorm:"many2many:problem_tags;"`
	KnowledgePoints []KnowledgePoint `json:"knowledge_points" gorm:"many2many:problem_knowledge_points;"`
}

type Tag struct {
	gorm.Model
	Name     string    `json:"name" gorm:"unique"`
	Problems []Problem `json:"problems" gorm:"many2many:problem_tags;"`
}

type KnowledgePoint struct {
	gorm.Model
	Name        string    `json:"name" gorm:"unique"`
	Description string    `json:"description"`
	Problems    []Problem `json:"problems" gorm:"many2many:problem_knowledge_points;"`
}
