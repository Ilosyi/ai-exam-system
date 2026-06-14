// ============================================================================
// models/exam_attempt.go - 考试答题记录数据模型
// ============================================================================
//
// 定义了学生考试答题记录的数据结构，对应数据库中的 exam_attempts 表。
//
// 答题流程：
// 1. 学生点击"开始答题" → 创建 ExamAttempt（status = in_progress）
// 2. 学生答题过程中 → 自动保存答案到 exam_answers 表
// 3. 学生点击"交卷"或超时 → 更新 status 为 submitted/timeout，自动阅卷
// 4. 学生查看结果 → 查询 ExamAttempt 和关联的 ExamAnswer
//
// 状态说明：
// - in_progress: 答题进行中
// - submitted:   已交卷（正常交卷）
// - timeout:     超时交卷（自动提交）
//
// 防重复答题：
// - 每个学生对每张试卷只能有一个答题记录
// - 如果已有 submitted/timeout 的记录，不允许再次开始
//
// 学习要点：
// - *time.Time 指针类型表示可空的时间字段
// - *int 指针类型表示可空的整数字段
// - GORM 的 Preload 预加载关联数据
// ============================================================================

package models

import "time"

// ExamAttempt 代表学生的一次考试答题记录。
//
// 字段说明：
// - ID:          答题记录唯一标识（自增主键）
// - PaperID:     关联的试卷 ID（外键）
// - StudentID:   学生用户 ID（外键）
// - StartedAt:   开始答题时间
// - SubmittedAt: 交卷时间（可为空，表示尚未交卷）
// - Status:      答题状态（in_progress/submitted/timeout）
// - TotalScore:  总分（可为空，表示尚未阅卷）
// - Answers:     答案列表（一对多关联）
type ExamAttempt struct {
	ID          uint       `json:"id" gorm:"primaryKey"`                          // 答题记录 ID（主键）
	PaperID     uint       `json:"paperId" gorm:"column:paper_id;not null;index"` // 试卷 ID（外键，索引）
	StudentID   uint       `json:"studentId" gorm:"column:student_id;not null;index"` // 学生 ID（外键，索引）
	StartedAt   time.Time  `json:"startedAt" gorm:"column:started_at"`            // 开始答题时间
	SubmittedAt *time.Time `json:"submittedAt" gorm:"column:submitted_at"`        // 交卷时间（可为空）
	Status      string     `json:"status" gorm:"default:in_progress"`             // 状态：in_progress/submitted/timeout
	TotalScore  *int       `json:"totalScore" gorm:"column:total_score"`          // 总分（可为空）

	// Answers 是该答题记录包含的所有答案（一对多关联）
	Answers []ExamAnswer `json:"answers,omitempty" gorm:"foreignKey:AttemptID"`
}
