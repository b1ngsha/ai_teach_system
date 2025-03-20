package models

type CourseClasses struct {
	CourseID uint `json:"course_id" gorm:"primaryKey;autoIncrement:false"`
	ClassID  uint `json:"class_id" gorm:"primaryKey;autoIncrement:false"`

	Course Course `json:"-" gorm:"foreignkey:CourseID"`
	Class  Class  `json:"-" gorm:"foreignkey:ClassID"`
}
