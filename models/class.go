package models

import "gorm.io/gorm"

type Class struct {
	gorm.Model
	Name string `json:"name" gorm:"unique;not null;index"`
	Users []User `json:"users,omitempty" gorm:"foreignKey:ClassID"`
}
