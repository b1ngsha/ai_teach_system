package models

type ProblemDifficulty string

const (
	ProblemDifficultyEasy   ProblemDifficulty = "Easy"
	ProblemDifficultyMedium ProblemDifficulty = "Medium"
	ProblemDifficultyHard   ProblemDifficulty = "Hard"
)

type Problem struct {
	ID              uint              `gorm:"primarykey"`
	LeetcodeID      int               `json:"leetcode_id" gorm:"unique;not null"`
	Title           string            `json:"title" gorm:"not null"`
	TitleCn         string            `json:"title_cn" gorm:"not null"`
	TitleSlug       string            `json:"title_slug" gorm:"not null"`
	Difficulty      ProblemDifficulty `json:"difficulty" gorm:"type:ENUM('Easy', 'Medium', 'Hard')"`
	Content         string            `json:"content" gorm:"type:text"`
	SampleTestcases string            `json:"sample_testcases" gorm:"type:text"`
	Tags            []Tag             `json:"tags" gorm:"many2many:problem_tag;"`
	Users           []User            `json:"-" gorm:"many2many:user_problems;"`
	Course          []Course          `json:"-" gorm:"many2many:course_problems"`
}
