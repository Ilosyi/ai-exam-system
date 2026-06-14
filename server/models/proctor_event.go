// ============================================================================
// models/proctor_event.go - 监考事件数据模型
// ============================================================================
//
// 定义了监考事件的数据结构，对应数据库中的 proctor_events 表。
//
// 监考机制说明：
// - 学生在答题过程中，前端会记录各种异常事件并上报
// - 事件类型包括：切屏、复制粘贴、窗口失焦等
// - 教师可以通过这些事件监控学生的考试行为
//
// 当前状态：
// - 后端已实现事件记录接口
// - 前端监考看板（教师端）尚未实现
//
// 学习要点：
// - 事件驱动架构的简单应用
// - PayloadJSON 的灵活存储（不同事件类型有不同的数据结构）
// - 为未来功能预留的扩展点
// ============================================================================

package models

import "time"

// ProctorEvent 代表考试过程中的一个监考事件。
//
// 字段说明：
// - ID:          事件唯一标识（自增主键）
// - AttemptID:   关联的答题记录 ID（外键）
// - EventType:   事件类型（如 "tab_switch", "copy_paste", "window_blur"）
// - EventTime:   事件发生时间
// - PayloadJSON: 事件附加数据（JSON 字符串，不同事件类型有不同的结构）
type ProctorEvent struct {
	ID          uint      `json:"id" gorm:"primaryKey"`                            // 事件 ID（主键）
	AttemptID   uint      `json:"attemptId" gorm:"column:attempt_id;not null;index"` // 答题记录 ID（外键，索引）
	EventType   string    `json:"eventType" gorm:"column:event_type;not null"`      // 事件类型
	EventTime   time.Time `json:"eventTime" gorm:"column:event_time"`               // 事件时间
	PayloadJSON string    `json:"payloadJson" gorm:"column:payload_json;type:text"` // 事件数据（JSON）
}
