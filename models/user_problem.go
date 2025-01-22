package models

import (
	"time"

	"gorm.io/gorm"
)

type ProblemStatus string

const (
	ProblemStatusUntried ProblemStatus = "UNTRIED"
	ProblemStatusTried   ProblemStatus = "TRIED"
	ProblemStatusSolved  ProblemStatus = "SOLVED"
)

type UserProblem struct {
	UserID    uint          `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	ProblemID uint          `json:"problem_id" gorm:"primaryKey;autoIncrement:false"`
	Status    ProblemStatus `json:"status" gorm:"type:ENUM('UNTRIED', 'TRIED', 'SOLVED');default:'UNTRIED'"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`

	User    User    `json:"-" gorm:"foreignkey:UserID"`
	Problem Problem `json:"-" gorm:"foreignkey:ProblemID"`
}

func (up *UserProblem) BeforeCreate(tx *gorm.DB) error {
	if up.Status == "" {
		up.Status = ProblemStatusUntried
	}
	return nil
}
