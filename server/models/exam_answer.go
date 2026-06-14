// ============================================================================
// models/exam_answer.go - 考试答案数据模型
// ============================================================================
//
// 定义了学生每道题的答案数据结构，对应数据库中的 exam_answers 表。
//
// 答案存储方式：
// - AnswerJSON: 学生的答案，以 JSON 字符串存储
//   - 单选题：'[0]' 表示选择了第一个选项
//   - 多选题：'[0,2]' 表示选择了第一和第三个选项
//   - 编程题：代码文本（预留，当前未实现）
//
// 自动阅卷：
// - IsCorrect: 是否答对（true/false，可为空表示未阅卷）
// - Score:     该题得分（可为空表示未阅卷）
//
// 阅卷逻辑（在 exam_handler.go 的 autoSubmit 中）：
// 1. 遍历试卷的每个题目
// 2. 对比学生答案和正确答案
// 3. 如果完全一致，标记为正确并给满分
// 4. 否则标记为错误，得 0 分
//
// 学习要点：
// - *bool 和 *int 指针类型表示可空字段
// - Upsert 操作的实现（存在则更新，不存在则插入）
// - JSON 字符串在数据库中的灵活存储
// ============================================================================

package models

// ExamAnswer 代表学生在一次考试中对某道题的答案。
//
// 字段说明：
// - ID:         答案唯一标识（自增主键）
// - AttemptID:  所属答题记录 ID（外键，关联 exam_attempts 表）
// - QuestionID: 关联的题目 ID（外键，关联 questions 表）
// - AnswerJSON: 学生的答案（JSON 字符串）
// - IsCorrect:  是否答对（可为空，表示未阅卷）
// - Score:      该题得分（可为空，表示未阅卷）
type ExamAnswer struct {
	ID         uint   `json:"id" gorm:"primaryKey"`                              // 答案 ID（主键）
	AttemptID  uint   `json:"attemptId" gorm:"column:attempt_id;not null;index"` // 答题记录 ID（外键，索引）
	QuestionID uint   `json:"questionId" gorm:"column:question_id;not null"`     // 题目 ID（外键）
	AnswerJSON string `json:"answerJson" gorm:"column:answer_json;type:text"`    // 学生答案（JSON 字符串）
	IsCorrect  *bool  `json:"isCorrect" gorm:"column:is_correct"`                // 是否答对（可为空）
	Score      *int   `json:"score" gorm:"column:score"`                         // 得分（可为空）
}
