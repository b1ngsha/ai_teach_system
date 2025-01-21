package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Avatar    string `json:"avatar"`
	Username  string `json:"username" gorm:"unique;not null"`
	Name      string `json:"name" gorm:"unique;not null"`
	StudentID string `json:"student_id" gorm:"unique;not null"`
	Class     string `json:"class"`
	Password  string `json:"-"`
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
