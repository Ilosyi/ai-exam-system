package repositories

import (
	"context"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// PaperRepository wraps DB operations for papers.
type PaperRepository struct {
	db *gorm.DB
}

func NewPaperRepository(db *gorm.DB) *PaperRepository {
	return &PaperRepository{db: db}
}

// PaperFilters holds filter params for listing papers.
type PaperFilters struct {
	Keyword   string
	Status    string
	CreatedBy *uint
	Page      int
	PageSize  int
}

type PaperSubmittedStudent struct {
	StudentID   uint       `json:"studentId"`
	Username    string     `json:"username"`
	Status      string     `json:"status"`
	TotalScore  *int       `json:"totalScore"`
	SubmittedAt *time.Time `json:"submittedAt"`
}

type PaperSubmissionStats struct {
	PaperID           uint                    `json:"paperId"`
	ClassID           *uint                   `json:"classId"`
	ExpectedCount     int64                   `json:"expectedCount"`
	SubmittedCount    int64                   `json:"submittedCount"`
	UnsubmittedCount  int64                   `json:"unsubmittedCount"`
	SubmittedStudents []PaperSubmittedStudent `json:"submittedStudents"`
}

// List returns paginated papers with optional filters.
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

// FindByID returns a paper with its items preloaded.
func (r *PaperRepository) FindByID(ctx context.Context, id uint) (*models.Paper, error) {
	var paper models.Paper
	if err := r.db.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_no ASC")
	}).First(&paper, id).Error; err != nil {
		return nil, err
	}
	return &paper, nil
}

// Create creates a paper with its items in a transaction.
func (r *PaperRepository) Create(ctx context.Context, paper *models.Paper) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set PaperID on items before creation so GORM creates them in one pass
		for i := range paper.Items {
			paper.Items[i].PaperID = paper.ID
		}
		// GORM Create with association will insert both Paper and Items together
		return tx.Create(paper).Error
	})
}

// Update updates a paper's basic fields.
func (r *PaperRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Paper{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a paper and its items.
func (r *PaperRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("paper_id = ?", id).Delete(&models.PaperItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Paper{}, id).Error
	})
}

// ReplaceItem replaces a paper item with a new question.
func (r *PaperRepository) ReplaceItem(ctx context.Context, itemID uint, newQuestionID uint) error {
	return r.db.WithContext(ctx).Model(&models.PaperItem{}).Where("id = ?", itemID).
		Updates(map[string]interface{}{"question_id": newQuestionID}).Error
}

// DeleteItem removes a paper item.
func (r *PaperRepository) DeleteItem(ctx context.Context, itemID uint) error {
	return r.db.WithContext(ctx).Delete(&models.PaperItem{}, itemID).Error
}

// RandomQuestions randomly selects questions by type and language.
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

	if total < int64(count) {
		return questions, nil // return empty to signal insufficient count
	}

	if err := query.Order("RANDOM()").Limit(count).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

// CountQuestionsByType counts questions by type and optional language.
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

// FindQuestionByID returns a question by ID.
func (r *PaperRepository) FindQuestionByID(ctx context.Context, id uint) (*models.Question, error) {
	var q models.Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

// RandomQuestionByTypeLanguage returns one random question of the given type/language, excluding certain IDs.
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

// GetPaperItemIDs returns all question IDs in a paper for exclude logic.
func (r *PaperRepository) GetPaperItemIDs(ctx context.Context, paperID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&models.PaperItem{}).Where("paper_id = ?", paperID).Pluck("question_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// CreatePublication creates a paper publication.
func (r *PaperRepository) CreatePublication(ctx context.Context, pub *models.PaperPublication) error {
	return r.db.WithContext(ctx).Create(pub).Error
}

// FindPublicationByPaperID returns the latest publication for a paper.
func (r *PaperRepository) FindPublicationByPaperID(ctx context.Context, paperID uint) (*models.PaperPublication, error) {
	var pub models.PaperPublication
	if err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).Order("created_at DESC").First(&pub).Error; err != nil {
		return nil, err
	}
	return &pub, nil
}

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

// UpdatePublication updates a publication.
func (r *PaperRepository) UpdatePublication(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.PaperPublication{}).Where("id = ?", id).Updates(updates).Error
}

// ListPublications returns all publications for a paper.
func (r *PaperRepository) ListPublications(ctx context.Context, paperID uint) ([]models.PaperPublication, error) {
	var pubs []models.PaperPublication
	if err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).Order("created_at DESC").Find(&pubs).Error; err != nil {
		return nil, err
	}
	return pubs, nil
}

func (r *PaperRepository) GetSubmissionStats(ctx context.Context, paperID uint, classID *uint) (*PaperSubmissionStats, error) {
	stats := &PaperSubmissionStats{PaperID: paperID, ClassID: classID}

	studentScope := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ? AND status = ?", "student", "active")
	if classID != nil {
		studentScope = studentScope.Where(
			"class_id = ? OR EXISTS (SELECT 1 FROM user_classes uc WHERE uc.user_id = users.id AND uc.class_id = ?)",
			*classID,
			*classID,
		)
	}
	if err := studentScope.Distinct("users.id").Count(&stats.ExpectedCount).Error; err != nil {
		return nil, err
	}

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

	if stats.ExpectedCount > stats.SubmittedCount {
		stats.UnsubmittedCount = stats.ExpectedCount - stats.SubmittedCount
	}

	rows := make([]PaperSubmittedStudent, 0)
	if err := attemptScope.
		Select("u.id AS student_id, u.username, ea.status, ea.total_score, ea.submitted_at").
		Order("ea.submitted_at DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

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

// GetRandomInt returns a random int (utility for future use).
func GetRandomInt(max int) int {
	return rand.Intn(max)
}
