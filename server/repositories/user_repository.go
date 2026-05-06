package repositories

import (
	"context"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// UserRepository wraps user-related DB operations.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

type UserFilters struct {
	Keyword  string
	Role     string
	ClassID  *uint
	Status   string
	Page     int
	PageSize int
}

func (r *UserRepository) List(ctx context.Context, filters UserFilters) ([]models.User, int64, error) {
	var users []models.User
	var total int64

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

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (r *UserRepository) EnsureDefaults(ctx context.Context, users []models.User) error {
	for _, candidate := range users {
		var count int64
		if err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", candidate.Username).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := r.db.WithContext(ctx).Create(&candidate).Error; err != nil {
			return err
		}
	}
	return nil
}
