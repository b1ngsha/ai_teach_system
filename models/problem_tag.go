package models

type ProblemTag struct {
	ProblemID uint    `json:"problem_id" gorm:"primaryKey;autoIncrement:false"`
	TagID     uint    `json:"tag_id" gorm:"primaryKey;autoIncrement:false"`
	Problem   Problem `json:"-" gorm:"foreignKey:ProblemID"`
	Tag       Tag     `json:"-" gorm:"foreignKey:TagID"`
}
