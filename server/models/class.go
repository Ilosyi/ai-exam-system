package models

// Class represents a class of students (reserved for future).
type Class struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"not null"`
	TeacherID uint   `json:"teacherId" gorm:"column:teacher_id;not null"`
}
