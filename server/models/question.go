// ============================================================================
// models/question.go - 题目数据模型
// ============================================================================
//
// 定义了题库中题目的数据结构，对应数据库中的 questions 表。
//
// 题型说明：
// - single:   单选题，答案为一个下标（如 [0]）
// - multiple: 多选题，答案为多个下标（如 [0, 2]）
// - coding:   编程题，无选项，答案由评测系统判定
//
// 语言说明：
// - go, cpp, java, javascript, python
//
// 选项和答案的存储方式：
// - OptionsJSON: 选项数组序列化为 JSON 字符串存储（如 '["A选项","B选项","C选项","D选项"]'）
// - AnswerJSON:  答案数组序列化为 JSON 字符串存储（如 '[0,2]' 表示第1和第3个选项正确）
//
// 为什么用 JSON 字符串而不是单独的表？
// - 简化查询：不需要 JOIN 多张表
// - 选项数量固定（4个），结构简单
// - 适合教学演示，降低复杂度
//
// 学习要点：
// - GORM 的时间字段自动管理（CreatedAt、UpdatedAt）
// - JSON 字符串在数据库中的存储方式
// - 索引的作用（index 标签）
// ============================================================================

package models

import "time"

// Question 代表题库中的一道题目。
//
// 字段说明：
// - ID:          题目唯一标识（自增主键）
// - CreatedBy:   创建人用户 ID（关联 users 表）
// - Type:        题型（single/multiple/coding）
// - Language:    编程语言（go/cpp/java/javascript/python）
// - Title:       题目标题
// - Content:     题目内容/描述
// - OptionsJSON: 选项数组，以 JSON 字符串存储
// - AnswerJSON:  答案数组，以 JSON 字符串存储
// - CreatedAt:   创建时间（GORM 自动管理）
// - UpdatedAt:   更新时间（GORM 自动管理）
type Question struct {
	ID          uint      `json:"id" gorm:"primaryKey"`                   // 题目 ID（主键）
	CreatedBy   uint      `json:"createdBy" gorm:"column:created_by;default:0;index"` // 创建人 ID（索引）
	Type        string    `json:"type"`                                    // 题型：single/multiple/coding
	Language    string    `json:"language"`                                // 语言：go/cpp/java/javascript/python
	Title       string    `json:"title"`                                   // 题目标题
	Content     string    `json:"content"`                                 // 题目内容
	OptionsJSON string    `json:"options" gorm:"column:options"`           // 选项（JSON 字符串）
	AnswerJSON  string    `json:"answers" gorm:"column:answers"`           // 答案（JSON 字符串）
	CreatedAt   time.Time `json:"createdAt"`                               // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`                               // 更新时间
}
