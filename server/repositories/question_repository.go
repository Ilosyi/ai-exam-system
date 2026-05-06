package repositories

import (
	"context"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// QuestionRepository wraps DB operations.
type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// QuestionFilters holds filter params for listing questions.
type QuestionFilters struct {
	Keyword  string
	Type     string
	Language string
	Page     int
	PageSize int
}

// List returns paginated questions with optional filters.
func (r *QuestionRepository) List(ctx context.Context, filters QuestionFilters) ([]models.Question, int64, error) {
	var items []models.Question
	var total int64

	// 开始构造 GORM 查询：先绑定上下文，便于在将来添加超时或 tracing
	query := r.db.WithContext(ctx).Model(&models.Question{})
	// 根据 filters 添加 where 条件
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Language != "" {
		query = query.Where("language = ?", filters.Language)
	}
	if filters.Keyword != "" {
		// 使用 LIKE 实现简单的模糊匹配（标题或内容包含关键词）
		like := "%" + filters.Keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	// 先统计总数，供分页使用
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 处理默认分页参数（防止前端传入不合理值）
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 最终查询：按更新时间倒序，应用偏移与限制
	if err := query.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *QuestionRepository) Create(ctx context.Context, q *models.Question) error {
	// Create 会在 q 中填充 ID/时间戳等字段
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *QuestionRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	// Updates 接受 map，可以进行部分字段更新，注意字段名需与数据库列名或 gorm 标签匹配
	return r.db.WithContext(ctx).Model(&models.Question{}).Where("id = ?", id).Updates(updates).Error
}

func (r *QuestionRepository) Delete(ctx context.Context, id uint) error {
	// Delete 指定主键删除单条记录
	return r.db.WithContext(ctx).Delete(&models.Question{}, id).Error
}

func (r *QuestionRepository) FindByID(ctx context.Context, id uint) (*models.Question, error) {
	var q models.Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepository) DeleteMany(ctx context.Context, ids []uint) error {
	// 批量删除，传入 id 列表
	return r.db.WithContext(ctx).Delete(&models.Question{}, ids).Error
}
