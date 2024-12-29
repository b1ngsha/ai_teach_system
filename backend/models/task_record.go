package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type TaskRecord struct {
	gorm.Model
	TaskType     string     `json:"task_type" gorm:"not null"`
	Status       TaskStatus `json:"status" gorm:"type:ENUM('pending', 'running', 'completed', 'failed');not null"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	TotalCount   int        `json:"total_count"`
	SuccessCount int        `json:"success_count"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
}
