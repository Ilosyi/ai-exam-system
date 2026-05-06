package models

import "time"

// PaperPublication represents a publication of a paper to a class.
type PaperPublication struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PaperID     uint      `json:"paperId" gorm:"column:paper_id;not null;index"`
	ClassID     *uint     `json:"classId" gorm:"column:class_id"`
	StartTime   time.Time `json:"startTime" gorm:"column:start_time"`
	EndTime     time.Time `json:"endTime" gorm:"column:end_time"`
	Duration    int       `json:"duration" gorm:"column:duration;default:0"` // 答题时长(分钟), 0=不限时
	IsPublished bool      `json:"isPublished" gorm:"column:is_published;default:false"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
