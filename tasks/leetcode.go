package tasks

import (
	"ai_teach_system/models"
	"log"
	"time"

	"gorm.io/gorm"
)

func (tm *TasksManager) SyncLeetCodeProblems() {
	now := time.Now()
	taskRecord := &models.TaskRecord{
		TaskType:  "sync_leetcode_problems",
		Status:    models.TaskStatusPending,
		StartTime: &now,
	}

	if err := tm.db.Create(taskRecord).Error; err != nil {
		log.Printf("创建任务记录失败: %v", err)
		return
	}

	taskRecord.Status = models.TaskStatusRunning
	tm.db.Save(taskRecord)

	// 异步执行同步任务
	go func() {
		defer func() {
			endTime := time.Now()
			taskRecord.EndTime = &endTime
			tm.db.Save(taskRecord)
		}()

		problems, err := tm.leetcodeService.FetchAllProblems()
		if err != nil {
			taskRecord.Status = models.TaskStatusFailed
			taskRecord.ErrorMessage = err.Error()
			return
		}

		taskRecord.TotalCount = len(problems)

		// 增量更新题目
		for _, problem := range problems {
			var existingProblem models.Problem
			result := tm.db.Where("leetcode_id = ?", problem.LeetcodeID).First(&existingProblem)

			if result.Error == gorm.ErrRecordNotFound {
				// 新题目，直接创建
				if err := tm.db.Create(problem).Error; err != nil {
					log.Printf("创建题目失败 %d: %v", problem.LeetcodeID, err)
					continue
				}
			} else {
				// 已存在的题目，更新内容
				existingProblem.Title = problem.Title
				existingProblem.Content = problem.Content
				existingProblem.Difficulty = problem.Difficulty
				existingProblem.SampleTestcases = problem.SampleTestcases

				// 更新标签
				if err := tm.db.Model(&existingProblem).Association("Tags").Replace(problem.Tags); err != nil {
					log.Printf("更新题目标签失败 %d: %v", problem.LeetcodeID, err)
					continue
				}

				if err := tm.db.Save(&existingProblem).Error; err != nil {
					log.Printf("更新题目失败 %d: %v", problem.LeetcodeID, err)
					continue
				}
			}

			taskRecord.SuccessCount++
		}

		taskRecord.Status = models.TaskStatusCompleted
	}()
}
