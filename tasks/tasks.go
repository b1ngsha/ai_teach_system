package tasks

import (
	"ai_teach_system/services"
	"log"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type TasksManager struct {
	db   *gorm.DB
	cron *cron.Cron

	leetcodeService services.LeetCodeServiceInterface
}

func NewTasksManager(db *gorm.DB, s services.LeetCodeServiceInterface) *TasksManager {
	return &TasksManager{
		db:              db,
		cron:            cron.New(cron.WithSeconds()),
		leetcodeService: s,
	}
}

func (tm *TasksManager) Start() {
	_, err := tm.cron.AddFunc("0 0 0 * * *", tm.SyncLeetCodeProblems) // 每天0点执行
	if err != nil {
		log.Printf("添加定时任务失败: %v", err)
		return
	}
	tm.cron.Start()
}

func (tm *TasksManager) Stop() {
	tm.cron.Stop()
}
