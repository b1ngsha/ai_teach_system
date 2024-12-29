package services

import (
	"ai_teach_system/models"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type SchedulerService struct {
	db              *gorm.DB
	leetcodeService *LeetCodeService
	cron            *cron.Cron
}

func NewSchedulerService(db *gorm.DB, leetcodeService *LeetCodeService) *SchedulerService {
	return &SchedulerService{
		db:              db,
		leetcodeService: leetcodeService,
		cron:            cron.New(cron.WithSeconds()),
	}
}

func (s *SchedulerService) Start() {
	// 每天0点执行
	_, err := s.cron.AddFunc("0 0 0 * * *", s.syncLeetCodeProblems)
	if err != nil {
		log.Printf("添加定时任务失败: %v", err)
		return
	}
	s.cron.Start()
}

func (s *SchedulerService) Stop() {
	s.cron.Stop()
}

func (s *SchedulerService) syncLeetCodeProblems() {
	// 创建任务记录
	now := time.Now()
	taskRecord := &models.TaskRecord{
		TaskType:  "sync_leetcode_problems",
		Status:    models.TaskStatusPending,
		StartTime: &now,
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		log.Printf("创建任务记录失败: %v", err)
		return
	}

	// 更新任务状态为运行中
	taskRecord.Status = models.TaskStatusRunning
	s.db.Save(taskRecord)

	// 异步执行同步任务
	go func() {
		defer func() {
			endTime := time.Now()
			taskRecord.EndTime = &endTime
			s.db.Save(taskRecord)
		}()

		problems, err := s.leetcodeService.FetchAllProblems()
		if err != nil {
			taskRecord.Status = models.TaskStatusFailed
			taskRecord.ErrorMessage = err.Error()
			return
		}

		taskRecord.TotalCount = len(problems)

		// 增量更新题目
		for _, problem := range problems {
			var existingProblem models.Problem
			result := s.db.Where("leetcode_id = ?", problem.LeetcodeID).First(&existingProblem)

			if result.Error == gorm.ErrRecordNotFound {
				// 新题目，直接创建
				if err := s.db.Create(problem).Error; err != nil {
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
				if err := s.db.Model(&existingProblem).Association("Tags").Replace(problem.Tags); err != nil {
					log.Printf("更新题目标签失败 %d: %v", problem.LeetcodeID, err)
					continue
				}

				if err := s.db.Save(&existingProblem).Error; err != nil {
					log.Printf("更新题目失败 %d: %v", problem.LeetcodeID, err)
					continue
				}
			}

			taskRecord.SuccessCount++
		}

		taskRecord.Status = models.TaskStatusCompleted
	}()
}
