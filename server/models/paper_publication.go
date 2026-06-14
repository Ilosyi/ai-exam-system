// ============================================================================
// models/paper_publication.go - 试卷发布记录数据模型
// ============================================================================
//
// 定义了试卷发布记录的数据结构，对应数据库中的 paper_publications 表。
//
// 发布机制说明：
// - 试卷创建后处于 draft（草稿）状态
// - 教师通过"发布"操作，设置考试的时间窗口和目标班级
// - 发布后，学生可以在考试列表中看到这张试卷
// - 教师可以"取消发布"，试卷回到草稿状态
//
// 时间窗口：
// - StartTime: 考试开始时间，学生在此时间之前无法开始答题
// - EndTime:   考试结束时间，学生在此时间之后无法开始答题
// - Duration:  答题时长（分钟），0 表示不限时
//
// 截止时间计算：
//   deadline = min(开始答题时间 + Duration, EndTime)
//
// 班级限制：
// - ClassID 为空：公共试卷，所有学生可见
// - ClassID 不为空：仅指定班级的学生可见
//
// 学习要点：
// - *uint 指针类型表示可空字段
// - 时间窗口的设计模式
// - 布尔字段的数据库存储
// ============================================================================

package models

import "time"

// PaperPublication 代表一次试卷发布记录。
//
// 字段说明：
// - ID:          发布记录唯一标识（自增主键）
// - PaperID:     关联的试卷 ID（外键）
// - ClassID:     目标班级 ID（可为空，表示公共试卷）
// - StartTime:   考试开始时间
// - EndTime:     考试结束时间
// - Duration:    答题时长（分钟），0 表示不限时
// - IsPublished: 是否已发布（true=发布，false=取消发布）
// - CreatedAt:   创建时间
// - UpdatedAt:   更新时间
type PaperPublication struct {
	ID          uint      `json:"id" gorm:"primaryKey"`                             // 发布记录 ID（主键）
	PaperID     uint      `json:"paperId" gorm:"column:paper_id;not null;index"`    // 试卷 ID（外键，索引）
	ClassID     *uint     `json:"classId" gorm:"column:class_id"`                   // 班级 ID（可为空）
	StartTime   time.Time `json:"startTime" gorm:"column:start_time"`               // 考试开始时间
	EndTime     time.Time `json:"endTime" gorm:"column:end_time"`                   // 考试结束时间
	Duration    int       `json:"duration" gorm:"column:duration;default:0"`         // 答题时长（分钟），0=不限时
	IsPublished bool      `json:"isPublished" gorm:"column:is_published;default:false"` // 是否已发布
	CreatedAt   time.Time `json:"createdAt"`                                         // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`                                         // 更新时间
}
