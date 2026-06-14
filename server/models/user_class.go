// ============================================================================
// models/user_class.go - 学生-班级关联数据模型
// ============================================================================
//
// 定义了学生与班级的多对多关系，对应数据库中的 user_classes 表。
//
// 为什么需要这张表？
// - 一个学生可以属于多个班级
// - 一个班级可以有多个学生
// - 这是典型的"多对多"关系，需要中间表来维护
//
// 与 users.class_id 的关系：
// - users 表中的 class_id 是旧设计（单班级），仍保留兼容
// - user_classes 是新设计（多班级），是主要的关联方式
// - 在查询学生所在班级时，需要同时考虑这两个来源
//
// 唯一约束：
// - (user_id, class_id) 组合必须唯一，防止重复加入
//
// 学习要点：
// - 多对多关系的数据库设计
// - 联合唯一索引的作用
// - GORM 的复合索引标签
// ============================================================================

package models

import "time"

// UserClass 代表学生与班级的关联关系。
//
// 字段说明：
// - ID:        关联记录唯一标识（自增主键）
// - UserID:    学生用户 ID（外键，关联 users 表）
// - ClassID:   班级 ID（外键，关联 classes 表）
// - CreatedAt: 加入时间
//
// 索引说明：
// - index:idx_user_class_unique,unique 表示 (user_id, class_id) 组合有唯一索引
//   即同一个学生不能重复加入同一个班级
type UserClass struct {
	ID        uint      `json:"id" gorm:"primaryKey"`                                                           // 关联 ID（主键）
	UserID    uint      `json:"userId" gorm:"column:user_id;not null;index:idx_user_class_unique,unique"`       // 学生 ID（联合唯一索引的一部分）
	ClassID   uint      `json:"classId" gorm:"column:class_id;not null;index:idx_user_class_unique,unique"`     // 班级 ID（联合唯一索引的一部分）
	CreatedAt time.Time `json:"createdAt"`                                                                       // 加入时间
}
