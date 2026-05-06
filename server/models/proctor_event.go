package models

import "time"

// ProctorEvent represents a monitoring event during an exam (reserved for future).
type ProctorEvent struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	AttemptID  uint      `json:"attemptId" gorm:"column:attempt_id;not null;index"`
	EventType  string    `json:"eventType" gorm:"column:event_type;not null"`
	EventTime  time.Time `json:"eventTime" gorm:"column:event_time"`
	PayloadJSON string   `json:"payloadJson" gorm:"column:payload_json;type:text"`
}
