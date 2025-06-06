package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	gorm.Model
	Username  string    `json:"username" gorm:"unique;not null"`
	Name      string    `json:"name" gorm:"unique;not null"`
	StudentID string    `json:"student_id" gorm:"unique;not null"`
	Password  string    `json:"-"`
	Role      Role      `json:"role" gorm:"type:ENUM('USER', 'ADMIN');default:'USER'"`
	Problems  []Problem `json:"-" gorm:"many2many:user_problems;"`
	Class     Class     `json:"class,omitempty" gorm:"foreignKey:ClassID"`
	ClassID   uint      `json:"class_id" gorm:"index"`
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
