package models

import "time"

// Paper represents a test paper / exam paper.
// status: draft | published | closed
type Paper struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title" gorm:"not null"`
	Language   string    `json:"language"`
	TotalScore int       `json:"totalScore" gorm:"column:total_score"`
	Status     string    `json:"status" gorm:"default:draft"`
	CreatedBy  uint      `json:"createdBy" gorm:"column:created_by"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	Items []PaperItem `json:"items,omitempty" gorm:"foreignKey:PaperID"`
}
