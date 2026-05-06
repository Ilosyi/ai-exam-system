package models

import "time"

// ExamAttempt represents a student's attempt at an exam.
// status: in_progress | submitted | timeout
type ExamAttempt struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	PaperID     uint       `json:"paperId" gorm:"column:paper_id;not null;index"`
	StudentID   uint       `json:"studentId" gorm:"column:student_id;not null;index"`
	StartedAt   time.Time  `json:"startedAt" gorm:"column:started_at"`
	SubmittedAt *time.Time `json:"submittedAt" gorm:"column:submitted_at"`
	Status      string     `json:"status" gorm:"default:in_progress"`
	TotalScore  *int       `json:"totalScore" gorm:"column:total_score"`

	Answers []ExamAnswer `json:"answers,omitempty" gorm:"foreignKey:AttemptID"`
}
