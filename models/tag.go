package models

type Tag struct {
	ID       uint      `gorm:"primarykey"`
	Name     string    `json:"name" gorm:"unique"`
	NameCn   string    `json:"name_cn"`
	Problems []Problem `json:"problems" gorm:"many2many:problem_tags;"`
}
