package models

// ExamAnswer represents a student's answer to a single question in an exam attempt.
type ExamAnswer struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	AttemptID  uint   `json:"attemptId" gorm:"column:attempt_id;not null;index"`
	QuestionID uint   `json:"questionId" gorm:"column:question_id;not null"`
	AnswerJSON string `json:"answerJson" gorm:"column:answer_json;type:text"`
	IsCorrect  *bool  `json:"isCorrect" gorm:"column:is_correct"`
	Score      *int   `json:"score" gorm:"column:score"`
}
