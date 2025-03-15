package models

type CourseProblem struct {
	CourseID  uint `json:"course_id" gorm:"primaryKey;autoIncrement:false"`
	ProblemID uint `json:"problem_id" gorm:"primaryKey;autoIncrement:false"`

	Course  Course  `json:"-" gorm:"foreignkey:CourseID"`
	Problem Problem `json:"-" gorm:"foreignkey:ProblemID"`
}
