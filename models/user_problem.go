package models

import (
	"gorm.io/gorm"
)

type ProblemStatus string

const (
	ProblemStatusUntried ProblemStatus = "UNTRIED"
	ProblemStatusTried   ProblemStatus = "TRIED"
	ProblemStatusSolved  ProblemStatus = "SOLVED"
	ProblemStatusFailed  ProblemStatus = "FAILED"
)

// 用户作答记录表（课程间隔离）
type UserProblem struct {
	gorm.Model
	UserID                        uint          `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	ProblemID                     uint          `json:"problem_id" gorm:"primaryKey;autoIncrement:false"`
	KnowledgePointID              uint          `json:"knowledge_point_id" gorm:"primaryKey;autoIncrement:false"`
	Status                        ProblemStatus `json:"status" gorm:"type:ENUM('UNTRIED', 'TRIED', 'FAILED', 'SOLVED');default:'UNTRIED'"`
	TypedCode                     string        `json:"typed_code"`
	QwenWrongReasonAndAnalyze     string        `json:"wrong_reason_and_analyze"`
	DeepseekWrongReasonAndAnalyze string        `json:"deepseek_wrong_reason_and_analyze"`
	QwenCorrectedCode             string        `json:"qwen_corrected_code"`
	DeepseekCorrected_code        string        `json:"deepseek_corrected_code"`
	SubmissionID                  float64       `json:"submission_id" gorm:"index"`

	User           User           `json:"-" gorm:"foreignkey:UserID"`
	Problem        Problem        `json:"-" gorm:"foreignkey:ProblemID"`
	KnowledgePoint KnowledgePoint `json:"-" gorm:"foreignkey:KnowledgePointID"`
}

func (up *UserProblem) BeforeCreate(tx *gorm.DB) error {
	if up.Status == "" {
		up.Status = ProblemStatusUntried
	}
	return nil
}
