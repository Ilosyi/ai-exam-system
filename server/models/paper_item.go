package models

// PaperItem represents a question item within a paper.
// type: single | multiple | coding
type PaperItem struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PaperID    uint   `json:"paperId" gorm:"column:paper_id;not null;index"`
	QuestionID uint   `json:"questionId" gorm:"column:question_id;not null"`
	Type       string `json:"type"`
	Score      int    `json:"score"`
	SortNo     int    `json:"sortNo" gorm:"column:sort_no"`
}
