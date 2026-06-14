// ============================================================================
// repositories/paper_repository.go - 试卷数据访问层
// ============================================================================
//
// 本文件封装了所有与试卷相关的数据库操作，包括：
// - 试卷本身的 CRUD
// - 试卷题目项（PaperItem）的管理
// - 试卷发布记录（PaperPublication）的管理
// - 随机组卷的题目查询
// - 提交统计的复杂查询
//
// 这是整个后端最复杂的 Repository，因为试卷涉及多个关联表。
//
// 学习要点：
// - GORM 的 Preload 预加载关联数据
// - 事务（Transaction）的使用
// - 复杂的 SQL 查询（JOIN、子查询、聚合）
// - RANDOM() 随机排序
// ============================================================================

package repositories

import (
	"context"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// PaperRepository 封装了试卷相关的数据库操作。
type PaperRepository struct {
	db *gorm.DB
}

// NewPaperRepository 创建一个新的 PaperRepository 实例。
func NewPaperRepository(db *gorm.DB) *PaperRepository {
	return &PaperRepository{db: db}
}

// PaperFilters 定义了试卷列表查询的筛选条件。
type PaperFilters struct {
	Keyword   string // 关键词（模糊匹配标题）
	Status    string // 状态筛选
	CreatedBy *uint  // 创建人筛选
	Page      int    // 页码
	PageSize  int    // 每页数量
}

// PaperSubmittedStudent 代表已提交试卷的学生信息。
type PaperSubmittedStudent struct {
	StudentID   uint       `json:"studentId"`   // 学生 ID
	Username    string     `json:"username"`     // 学生用户名
	Status      string     `json:"status"`       // 答题状态
	TotalScore  *int       `json:"totalScore"`   // 总分
	SubmittedAt *time.Time `json:"submittedAt"`  // 交卷时间
}

// PaperSubmissionStats 代表试卷的提交统计信息。
type PaperSubmissionStats struct {
	PaperID           uint                    `json:"paperId"`           // 试卷 ID
	ClassID           *uint                   `json:"classId"`           // 班级 ID（可为空）
	ExpectedCount     int64                   `json:"expectedCount"`     // 应提交人数
	SubmittedCount    int64                   `json:"submittedCount"`    // 已提交人数
	UnsubmittedCount  int64                   `json:"unsubmittedCount"`  // 未提交人数
	SubmittedStudents []PaperSubmittedStudent `json:"submittedStudents"` // 已提交学生列表
}

// List 分页查询试卷列表。
func (r *PaperRepository) List(ctx context.Context, filters PaperFilters) ([]models.Paper, int64, error) {
	var items []models.Paper
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Paper{})
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}
	if filters.Keyword != "" {
		like := "%" + filters.Keyword + "%"
		query = query.Where("title LIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	if err := query.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// FindByID 根据 ID 查询试卷（预加载题目项）。
//
// Preload 说明：
// - Preload("Items") 会自动查询 paper_items 表，填充 Paper.Items 字段
// - 回调函数用于自定义预加载的排序方式（按 sort_no 升序）
func (r *PaperRepository) FindByID(ctx context.Context, id uint) (*models.Paper, error) {
	var paper models.Paper
	if err := r.db.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_no ASC")
	}).First(&paper, id).Error; err != nil {
		return nil, err
	}
	return &paper, nil
}

// Create 创建试卷及其题目项（事务操作）。
//
// 事务保证：
// - 如果创建 PaperItem 失败，整个操作回滚
// - 不会出现"试卷创建了但题目项没创建"的不一致状态
func (r *PaperRepository) Create(ctx context.Context, paper *models.Paper) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range paper.Items {
			paper.Items[i].PaperID = paper.ID
		}
		return tx.Create(paper).Error
	})
}

// Update 更新试卷基本信息。
func (r *PaperRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Paper{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除试卷及其题目项（事务操作）。
//
// 先删除关联的 PaperItem，再删除 Paper 本身。
// 注意：没有删除 PaperPublication（发布记录保留用于审计）。
func (r *PaperRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("paper_id = ?", id).Delete(&models.PaperItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Paper{}, id).Error
	})
}

// ReplaceItem 替换试卷中的一个题目。
func (r *PaperRepository) ReplaceItem(ctx context.Context, itemID uint, newQuestionID uint) error {
	return r.db.WithContext(ctx).Model(&models.PaperItem{}).Where("id = ?", itemID).
		Updates(map[string]interface{}{"question_id": newQuestionID}).Error
}

// DeleteItem 删除试卷中的一个题目项。
func (r *PaperRepository) DeleteItem(ctx context.Context, itemID uint) error {
	return r.db.WithContext(ctx).Delete(&models.PaperItem{}, itemID).Error
}

// RandomQuestions 随机选取指定数量的题目（用于智能组卷）。
//
// 参数：
//   - ctx:      上下文
//   - qType:    题型（single/multiple/coding）
//   - language: 编程语言
//   - count:    需要的题目数量
//
// 返回值：
//   - []models.Question: 题目列表（如果数量不足，返回空列表）
//   - error: 查询失败时返回错误
//
// SQL 说明：
// - ORDER BY RANDOM() 是 SQLite 的随机排序语法
// - LIMIT count 限制返回数量
func (r *PaperRepository) RandomQuestions(ctx context.Context, qType string, language string, count int) ([]models.Question, error) {
	var questions []models.Question
	query := r.db.WithContext(ctx).Model(&models.Question{}).Where("type = ?", qType)
	if language != "" {
		query = query.Where("language = ?", language)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 如果可用题目数量不足，返回空列表
	if total < int64(count) {
		return questions, nil
	}

	if err := query.Order("RANDOM()").Limit(count).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

// CountQuestionsByType 统计指定类型和语言的题目数量。
func (r *PaperRepository) CountQuestionsByType(ctx context.Context, qType string, language string) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Question{}).Where("type = ?", qType)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindQuestionByID 根据 ID 查询题目。
func (r *PaperRepository) FindQuestionByID(ctx context.Context, id uint) (*models.Question, error) {
	var q models.Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

// RandomQuestionByTypeLanguage 随机选取一道题目（排除指定 ID）。
//
// 用于"替换题目"功能：随机选一道同类型、同语言、且不在当前试卷中的题目。
func (r *PaperRepository) RandomQuestionByTypeLanguage(ctx context.Context, qType string, language string, excludeIDs []uint) (*models.Question, error) {
	var q models.Question
	query := r.db.WithContext(ctx).Model(&models.Question{}).Where("type = ?", qType)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	if err := query.Order("RANDOM()").Limit(1).Find(&q).Error; err != nil {
		return nil, err
	}
	if q.ID == 0 {
		return nil, nil
	}
	return &q, nil
}

// GetPaperItemIDs 获取试卷中所有题目的 ID 列表（用于排除逻辑）。
func (r *PaperRepository) GetPaperItemIDs(ctx context.Context, paperID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&models.PaperItem{}).Where("paper_id = ?", paperID).Pluck("question_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// CreatePublication 创建试卷发布记录。
func (r *PaperRepository) CreatePublication(ctx context.Context, pub *models.PaperPublication) error {
	return r.db.WithContext(ctx).Create(pub).Error
}

// FindPublicationByPaperID 查询试卷的最新发布记录。
func (r *PaperRepository) FindPublicationByPaperID(ctx context.Context, paperID uint) (*models.PaperPublication, error) {
	var pub models.PaperPublication
	if err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).Order("created_at DESC").First(&pub).Error; err != nil {
		return nil, err
	}
	return &pub, nil
}

// FindActivePublicationByPaperID 查询试卷的最新有效发布记录（已发布状态）。
func (r *PaperRepository) FindActivePublicationByPaperID(ctx context.Context, paperID uint) (*models.PaperPublication, error) {
	var pub models.PaperPublication
	if err := r.db.WithContext(ctx).
		Where("paper_id = ? AND is_published = ?", paperID, true).
		Order("created_at DESC").
		First(&pub).Error; err != nil {
		return nil, err
	}
	return &pub, nil
}

// UpdatePublication 更新发布记录。
func (r *PaperRepository) UpdatePublication(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.PaperPublication{}).Where("id = ?", id).Updates(updates).Error
}

// ListPublications 查询试卷的所有发布记录。
func (r *PaperRepository) ListPublications(ctx context.Context, paperID uint) ([]models.PaperPublication, error) {
	var pubs []models.PaperPublication
	if err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).Order("created_at DESC").Find(&pubs).Error; err != nil {
		return nil, err
	}
	return pubs, nil
}

// GetSubmissionStats 查询试卷的提交统计。
//
// 统计逻辑：
// 1. 查询应提交人数（指定班级的所有活跃学生）
// 2. 查询已提交人数（状态为 submitted 或 timeout 的学生）
// 3. 计算未提交人数
// 4. 查询已提交学生列表
//
// 参数：
//   - paperID: 试卷 ID
//   - classID: 班级 ID（可为空，表示统计所有班级）
func (r *PaperRepository) GetSubmissionStats(ctx context.Context, paperID uint, classID *uint) (*PaperSubmissionStats, error) {
	stats := &PaperSubmissionStats{PaperID: paperID, ClassID: classID}

	// 统计应提交人数
	studentScope := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ? AND status = ?", "student", "active")
	if classID != nil {
		// 学生在班级中的判断：users.class_id 匹配 或 user_classes 表中有记录
		studentScope = studentScope.Where(
			"class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?)",
			*classID,
			*classID,
		)
	}
	if err := studentScope.Distinct("users.id").Count(&stats.ExpectedCount).Error; err != nil {
		return nil, err
	}

	// 统计已提交人数
	attemptScope := r.db.WithContext(ctx).
		Table("exam_attempts AS ea").
		Joins("JOIN users u ON u.id = ea.student_id").
		Where("ea.paper_id = ? AND ea.status IN ?", paperID, []string{"submitted", "timeout"})
	if classID != nil {
		attemptScope = attemptScope.Where(
			"u.class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = u.id AND uc.class_id = ?)",
			*classID,
			*classID,
		)
	}
	if err := attemptScope.Distinct("ea.student_id").Count(&stats.SubmittedCount).Error; err != nil {
		return nil, err
	}

	// 计算未提交人数
	if stats.ExpectedCount > stats.SubmittedCount {
		stats.UnsubmittedCount = stats.ExpectedCount - stats.SubmittedCount
	}

	// 查询已提交学生列表
	rows := make([]PaperSubmittedStudent, 0)
	if err := attemptScope.
		Select("u.id AS student_id, u.username, ea.status, ea.total_score, ea.submitted_at").
		Order("ea.submitted_at DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 去重（一个学生可能有多次提交，只保留最新的）
	seen := make(map[uint]struct{})
	for _, row := range rows {
		if _, ok := seen[row.StudentID]; ok {
			continue
		}
		stats.SubmittedStudents = append(stats.SubmittedStudents, row)
		seen[row.StudentID] = struct{}{}
	}

	return stats, nil
}

// GetRandomInt 返回一个随机整数（工具函数，预留使用）。
func GetRandomInt(max int) int {
	return rand.Intn(max)
}
