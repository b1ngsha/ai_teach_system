package models

import "gorm.io/gorm"

type ProblemStatus string

const (
	ProblemStatusUntried ProblemStatus = "UNTRIED"
	ProblemStatusTried   ProblemStatus = "TRIED"
	ProblemStatusSolved  ProblemStatus = "SOLVED"
)

type UserProblem struct {
	gorm.Model
	UserID    uint          `json:"user_id" gorm:"index:idx_user_problem"`
	ProblemID uint          `json:"problem_id" gorm:"index:idx_user_problem"`
	Status    ProblemStatus `json:"status" gorm:"type:ENUM('UNTRIED', 'TRIED', 'SOLVED');default:'UNTRIED'"`

	User    User    `json:"-" gorm:"foreignkey:UserID"`
	Problem Problem `json:"-" gorm:"foreignkey:ProblemID"`
}

func (up *UserProblem) BeforeCreate(tx *gorm.DB) error {
	if up.Status == "" {
		up.Status = ProblemStatusUntried
	}
	return nil
}
