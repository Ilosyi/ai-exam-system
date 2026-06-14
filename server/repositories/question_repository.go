// ============================================================================
// repositories/question_repository.go - 题目数据访问层
// ============================================================================
//
// 本文件封装了所有与题目表（questions）相关的数据库操作。
//
// 提供的操作：
// - List:      分页查询题目列表（支持筛选）
// - Create:    创建题目
// - Update:    更新题目
// - Delete:    删除单个题目
// - FindByID:  根据 ID 查询题目
// - DeleteMany: 批量删除题目
//
// 学习要点：
// - 分页查询的标准实现
// - LIKE 模糊查询
// - 批量删除的实现
// ============================================================================

package repositories

import (
	"context"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// QuestionRepository 封装了题目相关的数据库操作。
type QuestionRepository struct {
	db *gorm.DB // 数据库连接实例
}

// NewQuestionRepository 创建一个新的 QuestionRepository 实例。
func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// QuestionFilters 定义了题目列表查询的筛选条件。
//
// 字段说明：
// - Keyword:  关键词（模糊匹配标题或内容）
// - Type:     题型筛选（single/multiple/coding）
// - Language: 语言筛选（go/cpp/java/javascript/python）
// - Page:     页码（从 1 开始）
// - PageSize: 每页数量
type QuestionFilters struct {
	Keyword  string // 关键词
	Type     string // 题型
	Language string // 语言
	Page     int    // 页码
	PageSize int    // 每页数量
}

// List 分页查询题目列表。
//
// 参数：
//   - ctx:     上下文
//   - filters: 筛选条件
//
// 返回值：
//   - []models.Question: 题目列表
//   - int64:             符合条件的总记录数
//   - error:             查询失败时返回错误
//
// 查询逻辑：
// 1. 根据 filters 构建 WHERE 条件
// 2. 用 Count 统计总记录数
// 3. 用 Offset + Limit 查询当前页数据
// 4. 按更新时间倒序排列
func (r *QuestionRepository) List(ctx context.Context, filters QuestionFilters) ([]models.Question, int64, error) {
	var items []models.Question
	var total int64

	// 构建查询条件
	query := r.db.WithContext(ctx).Model(&models.Question{})
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Language != "" {
		query = query.Where("language = ?", filters.Language)
	}
	if filters.Keyword != "" {
		// LIKE 模糊查询：标题或内容包含关键词
		like := "%" + filters.Keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	// 统计总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 处理分页参数默认值
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询当前页数据
	if err := query.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Create 创建一道新题目。
//
// 参数：
//   - ctx: 上下文
//   - q:   题目对象指针
//
// 返回值：
//   - error: 创建失败时返回错误
func (r *QuestionRepository) Create(ctx context.Context, q *models.Question) error {
	return r.db.WithContext(ctx).Create(q).Error
}

// Update 更新题目信息。
//
// 参数：
//   - ctx:     上下文
//   - id:      题目 ID
//   - updates: 要更新的字段（键值对）
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *QuestionRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Question{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除单个题目。
//
// 参数：
//   - ctx: 上下文
//   - id:  题目 ID
//
// 返回值：
//   - error: 删除失败时返回错误
func (r *QuestionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Question{}, id).Error
}

// FindByID 根据 ID 查询题目。
//
// 参数：
//   - ctx: 上下文
//   - id:  题目 ID
//
// 返回值：
//   - *models.Question: 题目指针
//   - error: 查询失败时返回错误（如题目不存在）
func (r *QuestionRepository) FindByID(ctx context.Context, id uint) (*models.Question, error) {
	var q models.Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

// DeleteMany 批量删除题目。
//
// 参数：
//   - ctx: 上下文
//   - ids: 要删除的题目 ID 列表
//
// 返回值：
//   - error: 删除失败时返回错误
//
// GORM 方法说明：
// - Delete(&models.Question{}, ids): 传入 ID 列表，批量删除
func (r *QuestionRepository) DeleteMany(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&models.Question{}, ids).Error
}
