package models

type KnowledgePointTag struct {
	KnowledgePointID uint           `json:"knowledge_point_id" gorm:"primaryKey;autoIncrement:false"`
	TagID            uint           `json:"tag_id" gorm:"primaryKey;autoIncrement:false"`
	KnowledgePoint   KnowledgePoint `json:"-" gorm:"foreignkey:KnowledgePointID"`
	Tag              Tag            `json:"-" gorm:"foreignkey:TagID"`
}
