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

	func() {
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

		// 先处理所有标签
		allTags := make(map[string]*models.Tag)
		for _, problem := range problems {
			for _, tag := range problem.Tags {
				if _, exists := allTags[tag.Name]; !exists {
					allTags[tag.Name] = &tag
				}
			}
		}

		// 批量创建或更新标签
		for _, tag := range allTags {
			var existingTag models.Tag
			result := tm.db.Where("name = ?", tag.Name).First(&existingTag)
			if result.Error == gorm.ErrRecordNotFound {
				if err := tm.db.Create(tag).Error; err != nil {
					log.Printf("创建标签失败 %s: %v", tag.Name, err)
					continue
				}
			} else {
				tag.ID = existingTag.ID
			}
		}

		// 增量更新题目
		for _, problem := range problems {
			var existingProblem models.Problem
			result := tm.db.Where("leetcode_id = ?", problem.LeetcodeID).First(&existingProblem)

			// 更新标签的ID
			for i, tag := range problem.Tags {
				var existingTag models.Tag
				if err := tm.db.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
					log.Printf("获取标签失败 %s: %v", tag.Name, err)
					continue
				}
				problem.Tags[i].ID = existingTag.ID
			}

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
