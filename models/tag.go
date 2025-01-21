package models

type Tag struct {
	ID       uint      `gorm:"primarykey"`
	Name     string    `json:"name" gorm:"unique"`
	Problems []Problem `json:"problems" gorm:"many2many:problem_tag;"`
}
