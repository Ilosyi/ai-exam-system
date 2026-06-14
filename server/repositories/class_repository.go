// ============================================================================
// repositories/class_repository.go - 班级数据访问层
// ============================================================================
//
// 本文件封装了所有与班级相关的数据库操作，包括：
// - 班级本身的 CRUD
// - 学生-班级关联的管理（批量加入/移出）
// - 学生考试记录的查询
// - 复杂的多班级关联查询
//
// 本 Repository 是最复杂的 Repository 之一，因为涉及：
// - 两套班级关联机制（users.class_id 和 user_classes 表）
// - 复杂的 SQL 子查询
// - 事务操作
//
// 学习要点：
// - 多对多关系的查询和维护
// - 事务中的批量操作
// - CASE WHEN 条件表达式
// - EXISTS 子查询
// ============================================================================

package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"week05/homework/server/models"
)

// ClassRepository 封装了班级相关的数据库操作。
type ClassRepository struct {
	db *gorm.DB
}

// NewClassRepository 创建一个新的 ClassRepository 实例。
func NewClassRepository(db *gorm.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

// ClassFilters 定义了班级列表查询的筛选条件。
type ClassFilters struct {
	Keyword   string // 关键词（模糊匹配班级名称）
	TeacherID *uint  // 教师 ID 筛选
	Page      int    // 页码
	PageSize  int    // 每页数量
}

// ClassStudentFilters 定义了班级学生列表的筛选条件。
type ClassStudentFilters struct {
	Keyword  string // 关键词（模糊匹配用户名）
	Status   string // 状态筛选
	Scope    string // 范围：class（班级内学生）或 all（所有学生）
	Page     int    // 页码
	PageSize int    // 每页数量
}

// ClassStudentRow 代表班级学生列表中的一行数据。
type ClassStudentRow struct {
	ID       uint   // 学生 ID
	Username string // 学生用户名
	Status   string // 学生状态
	ClassID  *uint  // 学生的默认班级 ID
	InClass  bool   // 是否在当前班级中
}

// StudentExamRecord 代表学生的考试记录。
type StudentExamRecord struct {
	AttemptID   uint       // 答题记录 ID
	PaperID     uint       // 试卷 ID
	PaperTitle  string     // 试卷标题
	Status      string     // 答题状态
	TotalScore  *int       // 总分
	StartedAt   time.Time  // 开始时间
	SubmittedAt *time.Time // 交卷时间
}

// Create 创建班级。
func (r *ClassRepository) Create(ctx context.Context, item *models.Class) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// FindByID 根据 ID 查询班级。
func (r *ClassRepository) FindByID(ctx context.Context, id uint) (*models.Class, error) {
	var item models.Class
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// List 分页查询班级列表。
func (r *ClassRepository) List(ctx context.Context, filters ClassFilters) ([]models.Class, int64, error) {
	var items []models.Class
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Class{})
	if filters.Keyword != "" {
		like := "%" + filters.Keyword + "%"
		query = query.Where("name LIKE ?", like)
	}
	if filters.TeacherID != nil {
		query = query.Where("teacher_id = ?", *filters.TeacherID)
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

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Update 更新班级信息。
func (r *ClassRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Class{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除班级。
func (r *ClassRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Class{}, id).Error
}

// ExistsByTeacher 检查班级是否属于指定教师。
//
// 用于权限校验：教师只能操作自己创建的班级。
func (r *ClassRepository) ExistsByTeacher(ctx context.Context, classID uint, teacherID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Class{}).
		Where("id = ? AND teacher_id = ?", classID, teacherID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListClassIDsByStudent 查询学生所属的所有班级 ID。
//
// 查询逻辑（兼容两套关联机制）：
// 1. 从 users.class_id 获取学生的默认班级
// 2. 从 user_classes 表获取学生的其他班级
// 3. 合并去重后返回
func (r *ClassRepository) ListClassIDsByStudent(ctx context.Context, studentID uint) ([]uint, error) {
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})

	// 查询 users.class_id
	var primary struct {
		ClassID *uint `gorm:"column:class_id"`
	}
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("class_id").
		Where("id = ? AND role = ?", studentID, "student").
		Scan(&primary).Error; err != nil {
		return nil, err
	}
	if primary.ClassID != nil {
		ids = append(ids, *primary.ClassID)
		seen[*primary.ClassID] = struct{}{}
	}

	// 查询 user_classes 表
	var extra []uint
	if err := r.db.WithContext(ctx).
		Model(&models.UserClass{}).
		Where("user_id = ?", studentID).
		Pluck("class_id", &extra).Error; err != nil {
		return nil, err
	}
	// 合并去重
	for _, id := range extra {
		if _, ok := seen[id]; ok {
			continue
		}
		ids = append(ids, id)
		seen[id] = struct{}{}
	}

	return ids, nil
}

// ListStudentsByClass 分页查询班级学生列表。
//
// 这是一个复杂的查询，使用了 CASE WHEN 条件表达式来判断学生是否在班级中。
//
// 参数：
//   - classID: 班级 ID
//   - filters: 筛选条件
//
// 返回值：
//   - []ClassStudentRow: 学生列表（包含 InClass 标志）
//   - int64:             总记录数
//   - error:             查询失败时返回错误
//
// SQL 说明：
// - CASE WHEN ... THEN 1 ELSE 0 END AS in_class
//   这是一个条件表达式，如果学生在班级中则 in_class = 1，否则为 0
func (r *ClassRepository) ListStudentsByClass(ctx context.Context, classID uint, filters ClassStudentFilters) ([]ClassStudentRow, int64, error) {
	var items []ClassStudentRow
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ?", "student")

	if filters.Keyword != "" {
		like := "%" + filters.Keyword + "%"
		query = query.Where("username LIKE ?", like)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Scope != "all" {
		query = query.Where(
			"class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?)",
			classID,
			classID,
		)
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

	// 使用 CASE WHEN 计算 in_class 字段
	if err := query.
		Select(`
			users.id,
			users.username,
			users.status,
			users.class_id,
			CASE
				WHEN users.class_id = ? OR EXISTS (
					SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?
				) THEN 1
				ELSE 0
			END AS in_class
		`, classID, classID).
		Order("users.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// IsStudentInClass 检查学生是否在指定班级中。
func (r *ClassRepository) IsStudentInClass(ctx context.Context, classID uint, studentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND role = ?", studentID, "student").
		Where("class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?)", classID, classID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchAddStudentsToClass 批量将学生加入班级（事务操作）。
//
// 操作逻辑：
// 1. 向 user_classes 表插入关联记录（使用 ON CONFLICT DO NOTHING 避免重复）
// 2. 更新 users.class_id（仅对 class_id 为空的学生）
//
// ON CONFLICT DO NOTHING 说明：
// - 如果 (user_id, class_id) 组合已存在，不报错，跳过
// - 这是 PostgreSQL/SQLite 的语法，MySQL 使用 ON DUPLICATE KEY UPDATE
func (r *ClassRepository) BatchAddStudentsToClass(ctx context.Context, classID uint, studentIDs []uint) error {
	if len(studentIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 第一步：向 user_classes 表插入关联记录
		userClasses := make([]models.UserClass, 0, len(studentIDs))
		for _, studentID := range studentIDs {
			userClasses = append(userClasses, models.UserClass{UserID: studentID, ClassID: classID})
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&userClasses).Error; err != nil {
			return err
		}

		// 第二步：更新 users.class_id（仅对 class_id 为空的学生）
		if err := tx.
			Model(&models.User{}).
			Where("id IN ? AND role = ? AND class_id IS NULL", studentIDs, "student").
			Update("class_id", classID).Error; err != nil {
			return err
		}

		return nil
	})
}

// BatchRemoveStudentsFromClass 批量将学生移出班级（事务操作）。
//
// 操作逻辑：
// 1. 从 user_classes 表删除关联记录
// 2. 对于 users.class_id 等于当前班级的学生，将其 class_id 更新为下一个班级
//    （如果学生还有其他班级），或清空（如果没有其他班级）
func (r *ClassRepository) BatchRemoveStudentsFromClass(ctx context.Context, classID uint, studentIDs []uint) error {
	if len(studentIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 第一步：从 user_classes 表删除关联记录
		if err := tx.Where("class_id = ? AND user_id IN ?", classID, studentIDs).Delete(&models.UserClass{}).Error; err != nil {
			return err
		}

		// 第二步：同步 users.class_id
		var needSync []uint
		if err := tx.
			Model(&models.User{}).
			Where("id IN ? AND role = ? AND class_id = ?", studentIDs, "student", classID).
			Pluck("id", &needSync).Error; err != nil {
			return err
		}

		for _, studentID := range needSync {
			// 查找学生的下一个班级
			var nextClassID *uint
			var uc models.UserClass
			err := tx.Where("user_id = ?", studentID).Order("id ASC").First(&uc).Error
			if err == nil {
				nextClassID = &uc.ClassID
			} else if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			// 更新 users.class_id
			if err := tx.Model(&models.User{}).Where("id = ?", studentID).Update("class_id", nextClassID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// CountStudentsByClass 统计班级学生数量。
func (r *ClassRepository) CountStudentsByClass(ctx context.Context, classID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ?", "student").
		Where("class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?)", classID, classID).
		Distinct("users.id").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListStudentExamRecords 查询学生在指定班级中的考试记录。
func (r *ClassRepository) ListStudentExamRecords(ctx context.Context, classID uint, studentID uint, page int, pageSize int) ([]StudentExamRecord, int64, error) {
	var rows []StudentExamRecord
	var total int64

	baseQuery := r.db.WithContext(ctx).
		Table("exam_attempts AS ea").
		Joins("JOIN papers p ON p.id = ea.paper_id").
		Where("ea.student_id = ?", studentID).
		Where("EXISTS (SELECT 1 FROM users u WHERE u.id = ea.student_id AND (u.class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = u.id AND uc.class_id = ?)))", classID, classID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if err := baseQuery.
		Select("ea.id AS attempt_id, ea.paper_id, p.title AS paper_title, ea.status, ea.total_score, ea.started_at, ea.submitted_at").
		Order("ea.started_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
