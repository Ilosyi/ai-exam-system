package models

import "time"

// UserClass maps users to classes, enabling multi-class membership for students.
type UserClass struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"column:user_id;not null;index:idx_user_class_unique,unique"`
	ClassID   uint      `json:"classId" gorm:"column:class_id;not null;index:idx_user_class_unique,unique"`
	CreatedAt time.Time `json:"createdAt"`
}
