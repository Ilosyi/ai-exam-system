// ============================================================================
// models/user.go - 用户数据模型
// ============================================================================
//
// 定义了系统用户的数据结构，对应数据库中的 users 表。
//
// 角色说明：
// - admin:   管理员，拥有所有权限
// - teacher: 教师，可以管理题库、试卷、班级
// - student: 学生，只能参加考试、查看结果
//
// 状态说明：
// - active:   正常状态，可以登录和使用系统
// - disabled:  停用状态，无法登录
//
// 数据库表结构：
//   CREATE TABLE users (
//     id            INTEGER PRIMARY KEY AUTOINCREMENT,
//     username      TEXT NOT NULL UNIQUE,
//     role          TEXT DEFAULT 'student',
//     class_id      INTEGER,
//     password_hash TEXT,
//     status        TEXT DEFAULT 'active'
//   );
//
// 学习要点：
// - GORM 标签的含义：json（JSON 序列化）、gorm（数据库映射）
// - json:"-" 表示该字段在 JSON 序列化时被忽略（密码哈希不应返回给前端）
// - *uint 是指针类型，表示该字段可以为 NULL
// ============================================================================

package models

// User 代表系统中的一个用户。
//
// GORM 标签说明：
// - json:"id"          → JSON 序列化时字段名为 "id"
// - gorm:"primaryKey"  → 该字段是数据库表的主键
// - gorm:"uniqueIndex" → 该字段有唯一索引（不允许重复）
// - gorm:"not null"    → 该字段不允许为空
// - gorm:"default:xxx" → 数据库层面的默认值
// - gorm:"column:xxx"  → 指定数据库列名（默认与字段名的小写形式相同）
//
// 字段说明：
// - ID:           用户唯一标识（自增主键）
// - Username:     用户名（唯一，用于登录）
// - Role:         角色（admin/teacher/student）
// - ClassID:      所属班级 ID（可为空，兼容旧的单班级设计）
// - PasswordHash: 密码的 bcrypt 哈希值（json:"-" 表示不返回给前端）
// - Status:       账号状态（active/disabled）
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`                             // 用户 ID（主键）
	Username     string `json:"username" gorm:"uniqueIndex;not null"`             // 用户名（唯一索引）
	Role         string `json:"role" gorm:"default:student"`                      // 角色：admin/teacher/student
	ClassID      *uint  `json:"classId" gorm:"column:class_id"`                   // 所属班级 ID（可为空）
	PasswordHash string `json:"-" gorm:"column:password_hash"`                    // 密码哈希（json:"-" 不返回给前端）
	Status       string `json:"status" gorm:"default:active"`                     // 账号状态：active/disabled
}
