package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"week05/homework/server/models"
)

// ClassRepository wraps DB operations for classes.
type ClassRepository struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

type ClassFilters struct {
	Keyword   string
	TeacherID *uint
	Page      int
	PageSize  int
}

type ClassStudentFilters struct {
	Keyword  string
	Status   string
	Scope    string
	Page     int
	PageSize int
}

type ClassStudentRow struct {
	ID       uint
	Username string
	Status   string
	ClassID  *uint
	InClass  bool
}

type StudentExamRecord struct {
	AttemptID   uint
	PaperID     uint
	PaperTitle  string
	Status      string
	TotalScore  *int
	StartedAt   time.Time
	SubmittedAt *time.Time
}

func (r *ClassRepository) Create(ctx context.Context, item *models.Class) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ClassRepository) FindByID(ctx context.Context, id uint) (*models.Class, error) {
	var item models.Class
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

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

func (r *ClassRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Class{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ClassRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Class{}, id).Error
}

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

func (r *ClassRepository) ListClassIDsByStudent(ctx context.Context, studentID uint) ([]uint, error) {
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})

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

	var extra []uint
	if err := r.db.WithContext(ctx).
		Model(&models.UserClass{}).
		Where("user_id = ?", studentID).
		Pluck("class_id", &extra).Error; err != nil {
		return nil, err
	}
	for _, id := range extra {
		if _, ok := seen[id]; ok {
			continue
		}
		ids = append(ids, id)
		seen[id] = struct{}{}
	}

	return ids, nil
}

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

func (r *ClassRepository) BatchAddStudentsToClass(ctx context.Context, classID uint, studentIDs []uint) error {
	if len(studentIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userClasses := make([]models.UserClass, 0, len(studentIDs))
		for _, studentID := range studentIDs {
			userClasses = append(userClasses, models.UserClass{UserID: studentID, ClassID: classID})
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&userClasses).Error; err != nil {
			return err
		}

		if err := tx.
			Model(&models.User{}).
			Where("id IN ? AND role = ? AND class_id IS NULL", studentIDs, "student").
			Update("class_id", classID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ClassRepository) BatchRemoveStudentsFromClass(ctx context.Context, classID uint, studentIDs []uint) error {
	if len(studentIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ? AND user_id IN ?", classID, studentIDs).Delete(&models.UserClass{}).Error; err != nil {
			return err
		}

		var needSync []uint
		if err := tx.
			Model(&models.User{}).
			Where("id IN ? AND role = ? AND class_id = ?", studentIDs, "student", classID).
			Pluck("id", &needSync).Error; err != nil {
			return err
		}

		for _, studentID := range needSync {
			var nextClassID *uint
			var uc models.UserClass
			err := tx.Where("user_id = ?", studentID).Order("id ASC").First(&uc).Error
			if err == nil {
				nextClassID = &uc.ClassID
			} else if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err := tx.Model(&models.User{}).Where("id = ?", studentID).Update("class_id", nextClassID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

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
