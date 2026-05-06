package models

// User represents a system user (reserved for future auth).
// role: admin | teacher | student
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"uniqueIndex;not null"`
	Role         string `json:"role" gorm:"default:student"`
	ClassID      *uint  `json:"classId" gorm:"column:class_id"`
	PasswordHash string `json:"-" gorm:"column:password_hash"`
	Status       string `json:"status" gorm:"default:active"`
}
