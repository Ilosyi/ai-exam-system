// ============================================================================
// repositories/user_repository.go - 用户数据访问层
// ============================================================================
//
// 本文件封装了所有与用户表（users）相关的数据库操作。
//
// Repository 模式的作用：
// - 将数据库操作从业务逻辑中分离出来
// - Handler 不直接操作数据库，而是通过 Repository
// - 便于测试时替换为 mock 实现
//
// 本 Repository 提供以下操作：
// - FindByID:       根据 ID 查询用户
// - FindByUsername:  根据用户名查询用户
// - Create:          创建用户
// - List:            分页查询用户列表（支持筛选）
// - Update:          更新用户信息
// - Delete:          删除用户
// - EnsureDefaults:  确保默认用户存在（幂等操作）
//
// 学习要点：
// - GORM 的基本 CRUD 操作
// - 分页查询的实现方式
// - map[string]interface{} 用于部分字段更新
// ============================================================================

package repositories

import (
	"context"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// UserRepository 封装了用户相关的数据库操作。
//
// 它持有一个 gorm.DB 实例，所有数据库操作都通过这个实例执行。
type UserRepository struct {
	db *gorm.DB // 数据库连接实例
}

// NewUserRepository 创建一个新的 UserRepository 实例。
//
// 参数：
//   - db: GORM 数据库连接实例
//
// 返回值：
//   - *UserRepository: 用户数据访问层实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID 根据用户 ID 查询用户。
//
// 参数：
//   - ctx: 上下文，用于控制超时/取消
//   - id:  用户 ID
//
// 返回值：
//   - *models.User: 用户指针
//   - error: 查询失败时返回错误（如用户不存在）
//
// GORM 方法说明：
// - WithContext(ctx): 将上下文传递给 GORM，支持超时控制
// - First(&user, id): 根据主键查询第一条记录
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户。
//
// 参数：
//   - ctx:      上下文
//   - username: 用户名
//
// 返回值：
//   - *models.User: 用户指针
//   - error: 查询失败时返回错误（如用户不存在）
//
// GORM 方法说明：
// - Where("username = ?", username): 添加 WHERE 条件
// - First(&user): 查询第一条匹配的记录
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建一个新用户。
//
// 参数：
//   - ctx:  上下文
//   - user: 用户对象指针（Create 会填充 ID、CreatedAt 等字段）
//
// 返回值：
//   - error: 创建失败时返回错误
//
// GORM 方法说明：
// - Create(user): 插入一条新记录，GORM 会自动填充 ID 和时间戳
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// UserFilters 定义了用户列表查询的筛选条件。
//
// 字段说明：
// - Keyword:  关键词（模糊匹配用户名）
// - Role:     角色筛选
// - ClassID:  班级 ID 筛选
// - Status:   状态筛选
// - Page:     页码（从 1 开始）
// - PageSize: 每页数量
type UserFilters struct {
	Keyword  string // 关键词（模糊匹配用户名）
	Role     string // 角色筛选
	ClassID  *uint  // 班级 ID 筛选（可为空）
	Status   string // 状态筛选
	Page     int    // 页码
	PageSize int    // 每页数量
}

// List 分页查询用户列表。
//
// 参数：
//   - ctx:     上下文
//   - filters: 筛选条件
//
// 返回值：
//   - []models.User: 用户列表
//   - int64:         符合条件的总记录数
//   - error:         查询失败时返回错误
//
// 分页实现原理：
// 1. 先用 Count 统计总记录数
// 2. 再用 Offset + Limit 查询当前页的数据
// 3. Offset = (page - 1) * pageSize，表示跳过前面的记录
// 4. Limit = pageSize，表示最多返回多少条记录
//
// LIKE 查询说明：
// - "%" + keyword + "%" 表示关键词可以在任意位置
// - 例如 keyword = "test"，会匹配 "test123"、"abc_test" 等
func (r *UserRepository) List(ctx context.Context, filters UserFilters) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 构建查询条件
	query := r.db.WithContext(ctx).Model(&models.User{})
	if filters.Keyword != "" {
		like := "%" + filters.Keyword + "%"
		query = query.Where("username LIKE ?", like)
	}
	if filters.Role != "" {
		query = query.Where("role = ?", filters.Role)
	}
	if filters.ClassID != nil {
		query = query.Where("class_id = ?", *filters.ClassID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	// 统计总记录数（用于分页）
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

	// 查询当前页的数据
	// Order("id DESC"): 按 ID 降序排列（最新的在前面）
	// Offset: 跳过前面的记录
	// Limit: 最多返回 pageSize 条记录
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户信息。
//
// 参数：
//   - ctx:     上下文
//   - id:      用户 ID
//   - updates: 要更新的字段（键值对）
//
// 返回值：
//   - error: 更新失败时返回错误
//
// 使用 map[string]interface{} 的好处：
// - 可以只更新部分字段（不需要先查询再更新）
// - 字段名使用数据库列名（如 "password_hash"），不是 Go 字段名
// - 空值（如 0、""、nil）也会被更新
//
// 示例：
//
//	repo.Update(ctx, 1, map[string]interface{}{"role": "admin", "status": "active"})
func (r *UserRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除用户。
//
// 参数：
//   - ctx: 上下文
//   - id:  用户 ID
//
// 返回值：
//   - error: 删除失败时返回错误
//
// 注意：这是物理删除，不是软删除。删除后数据无法恢复。
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// EnsureDefaults 确保默认用户存在（幂等操作）。
//
// 这是一个"幂等操作"——无论调用多少次，结果都是一样的：
// - 如果用户已存在（通过 username 判断），跳过
// - 如果用户不存在，创建
//
// 参数：
//   - ctx:   上下文
//   - users: 要确保存在的用户列表
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用场景：系统首次启动时，创建默认的 admin、teacher、student 账号
func (r *UserRepository) EnsureDefaults(ctx context.Context, users []models.User) error {
	for _, candidate := range users {
		// 检查用户是否已存在
		var count int64
		if err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", candidate.Username).Count(&count).Error; err != nil {
			return err
		}
		// 如果已存在，跳过
		if count > 0 {
			continue
		}
		// 如果不存在，创建
		if err := r.db.WithContext(ctx).Create(&candidate).Error; err != nil {
			return err
		}
	}
	return nil
}
