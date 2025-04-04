package models

type KnowledgePointProblem struct {
	KnowledgePointID uint           `json:"knowledge_point_id" gorm:"primaryKey;autoIncrement:false"`
	ProblemID        uint           `json:"problem_id" gorm:"primaryKey;autoIncrement:false"`
	KnowledgePoint   KnowledgePoint `json:"-" gorm:"foreignkey:KnowledgePointID"`
	Problem          Problem        `json:"-" gorm:"foreignkey:ProblemID"`
}
