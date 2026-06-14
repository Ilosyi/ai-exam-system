// ============================================================================
// models/paper.go - 试卷数据模型
// ============================================================================
//
// 定义了试卷的数据结构，对应数据库中的 papers 表。
//
// 试卷状态说明：
// - draft:     草稿状态，只有创建者可见，可以编辑
// - published: 已发布状态，学生可以在考试列表中看到
// - closed:    已关闭状态，考试结束
//
// 试卷与题目的关系：
// - 一张试卷包含多个题目（通过 paper_items 表关联）
// - 使用 GORM 的 hasMany 关联：Paper.Items = []PaperItem
//
// 学习要点：
// - GORM 的关联标签：foreignKey 指定外键
// - omitempty 标签：JSON 序列化时如果字段为空则省略
// - 试卷的生命周期：draft → published → closed
// ============================================================================

package models

import "time"

// Paper 代表一张试卷。
//
// 字段说明：
// - ID:         试卷唯一标识（自增主键）
// - Title:      试卷标题
// - Language:   试卷涉及的编程语言
// - TotalScore: 试卷总分
// - Status:     试卷状态（draft/published/closed）
// - CreatedBy:  创建人用户 ID
// - CreatedAt:  创建时间
// - UpdatedAt:  更新时间
// - Items:      试卷包含的题目项（一对多关联）
type Paper struct {
	ID         uint      `json:"id" gorm:"primaryKey"`                // 试卷 ID（主键）
	Title      string    `json:"title" gorm:"not null"`               // 试卷标题
	Language   string    `json:"language"`                             // 编程语言
	TotalScore int       `json:"totalScore" gorm:"column:total_score"` // 总分
	Status     string    `json:"status" gorm:"default:draft"`          // 状态：draft/published/closed
	CreatedBy  uint      `json:"createdBy" gorm:"column:created_by"`  // 创建人 ID
	CreatedAt  time.Time `json:"createdAt"`                            // 创建时间
	UpdatedAt  time.Time `json:"updatedAt"`                            // 更新时间

	// Items 是试卷包含的题目项列表（一对多关联）
	// gorm:"foreignKey:PaperID" 表示 PaperItem 表中的 PaperID 字段是外键
	// json:"items,omitempty" 表示 JSON 序列化时，如果 Items 为空则省略
	Items []PaperItem `json:"items,omitempty" gorm:"foreignKey:PaperID"`
}
