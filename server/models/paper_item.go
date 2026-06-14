// ============================================================================
// models/paper_item.go - 试卷题目项数据模型
// ============================================================================
//
// 定义了试卷中每个题目的数据结构，对应数据库中的 paper_items 表。
//
// 这是试卷（Paper）和题目（Question）之间的"中间表"，
// 记录了"哪张试卷包含哪道题目，分值是多少，排列顺序是什么"。
//
// 为什么不直接用多对多关联？
// - 因为我们需要额外存储"分值"和"排序号"
// - 标准的多对多关联表只能存储两个外键
// - 所以使用独立的中间表，可以存储更多关联属性
//
// 学习要点：
// - 多对多关系的两种实现方式：GORM Many2Many vs 独立中间表
// - SortNo 排序字段的作用
// - 外键索引的性能意义
// ============================================================================

package models

// PaperItem 代表试卷中的一个题目项。
//
// 它连接了试卷（Paper）和题目（Question），并记录了：
// - 这道题在试卷中的位置（SortNo）
// - 这道题的分值（Score）
// - 这道题的类型（Type，冗余存储便于查询）
//
// 字段说明：
// - ID:         题目项唯一标识（自增主键）
// - PaperID:    所属试卷 ID（外键，关联 papers 表）
// - QuestionID: 关联的题目 ID（外键，关联 questions 表）
// - Type:       题型（single/multiple/coding，冗余存储）
// - Score:      该题的分值
// - SortNo:     排序号（决定题目在试卷中的显示顺序）
type PaperItem struct {
	ID         uint   `json:"id" gorm:"primaryKey"`                    // 题目项 ID（主键）
	PaperID    uint   `json:"paperId" gorm:"column:paper_id;not null;index"` // 所属试卷 ID（外键，索引）
	QuestionID uint   `json:"questionId" gorm:"column:question_id;not null"` // 关联题目 ID（外键）
	Type       string `json:"type"`                                     // 题型：single/multiple/coding
	Score      int    `json:"score"`                                    // 分值
	SortNo     int    `json:"sortNo" gorm:"column:sort_no"`            // 排序号
}
