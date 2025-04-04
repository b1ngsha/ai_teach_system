package models

type ProblemDifficulty string

const (
	ProblemDifficultyEasy   ProblemDifficulty = "Easy"
	ProblemDifficultyMedium ProblemDifficulty = "Medium"
	ProblemDifficultyHard   ProblemDifficulty = "Hard"
)

type Problem struct {
	ID              uint              `gorm:"primarykey"`
	LeetcodeID      int               `json:"leetcode_id"`
	Title           string            `json:"title" gorm:"type:varchar(255);not null"`
	TitleCn         string            `json:"title_cn" gorm:"not null"`
	TitleSlug       string            `json:"title_slug" gorm:"not null"`
	Difficulty      ProblemDifficulty `json:"difficulty" gorm:"type:ENUM('Easy', 'Medium', 'Hard')"`
	Content         string            `json:"content" gorm:"type:text;not null"`
	ContentCn       string            `json:"content_cn" gorm:"type:text"`
	SampleTestcases string            `json:"sample_testcases" gorm:"type:text"`
	Tags            []Tag             `json:"tags" gorm:"many2many:problem_tags;"`
	Users           []User            `json:"-" gorm:"many2many:user_problems;"`
	KnowledgePoints []KnowledgePoint  `json:"knowledge_points" gorm:"many2many:knowledge_point_problems;"`
	IsCustom        bool              `json:"is_custom" gorm:"default:false"`
	TestCases       string            `json:"test_cases" gorm:"type:text"`
	TimeLimit       int               `json:"time_limit" gorm:"type:int;default:1000"`
	MemoryLimit     int               `json:"memory_limit" gorm:"type:int;default:128"`
}
