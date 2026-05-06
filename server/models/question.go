package models

import "time"

// Question represents a single question row saved in SQLite database.
// type: single | multiple | coding
// language: go | cpp | java | javascript | python
type Question struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedBy   uint      `json:"createdBy" gorm:"column:created_by;default:0;index"`
	Type        string    `json:"type"`
	Language    string    `json:"language"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	OptionsJSON string    `json:"options" gorm:"column:options"`
	AnswerJSON  string    `json:"answers" gorm:"column:answers"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
