package services_test

import (
	"ai_teach_system/models"
	"ai_teach_system/tasks"
	"ai_teach_system/tests"
	"ai_teach_system/tests/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerService_syncLeetCodeProblems(t *testing.T) {
	db, cleanup := tests.SetupTestDB()
	defer cleanup()

	manager := tasks.NewTasksManager(db, mocks.NewMockLeetCodeService())

	manager.SyncLeetCodeProblems()

	// 等待异步任务完成
	time.Sleep(time.Second)

	// 验证任务记录
	var taskRecord models.TaskRecord
	err := db.Order("created_at desc").First(&taskRecord).Error
	assert.NoError(t, err)
	assert.Equal(t, models.TaskStatusCompleted, taskRecord.Status)
	assert.Equal(t, 2, taskRecord.SuccessCount)

	// 验证问题是否正确保存
	var problems []models.Problem
	err = db.Preload("Tags").Find(&problems).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, len(problems))
	assert.Equal(t, "Two Sum", problems[0].Title)
	assert.Equal(t, "Easy", problems[0].Difficulty)
	assert.Equal(t, "Array", problems[0].Tags[0].Name)
}
