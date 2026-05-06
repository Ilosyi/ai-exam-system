package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// ExamRepository wraps DB operations for exam attempts and answers.
type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

// CreateAttempt creates a new exam attempt.
func (r *ExamRepository) CreateAttempt(ctx context.Context, attempt *models.ExamAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

// FindAttemptByID returns an attempt with answers preloaded.
func (r *ExamRepository) FindAttemptByID(ctx context.Context, id uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Preload("Answers").First(&attempt, id).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// FindActiveAttempt returns the in-progress attempt for a student+paper.
func (r *ExamRepository) FindActiveAttempt(ctx context.Context, studentID, paperID uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Where("student_id = ? AND paper_id = ? AND status = ?", studentID, paperID, "in_progress").First(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// UpdateAttempt updates an attempt's fields.
func (r *ExamRepository) UpdateAttempt(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExamAttempt{}).Where("id = ?", id).Updates(updates).Error
}

// UpsertAnswer creates or updates an answer for a question in an attempt.
func (r *ExamRepository) UpsertAnswer(ctx context.Context, answer *models.ExamAnswer) error {
	var existing models.ExamAnswer
	err := r.db.WithContext(ctx).Where("attempt_id = ? AND question_id = ?", answer.AttemptID, answer.QuestionID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(answer).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"answer_json": answer.AnswerJSON,
	}).Error
}

// FindAnswer returns the answer for a specific question in an attempt.
func (r *ExamRepository) FindAnswer(ctx context.Context, attemptID, questionID uint) (*models.ExamAnswer, error) {
	var answer models.ExamAnswer
	if err := r.db.WithContext(ctx).Where("attempt_id = ? AND question_id = ?", attemptID, questionID).First(&answer).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

// ListPublishedPapers returns papers that are currently published with active time windows.
func (r *ExamRepository) ListPublishedPapers(ctx context.Context, classIDs []uint) ([]models.Paper, error) {
	var papers []models.Paper
	now := time.Now()
	query := r.db.WithContext(ctx).
		Joins("JOIN paper_publications ON paper_publications.paper_id = papers.id").
		Where("paper_publications.is_published = ? AND unixepoch(paper_publications.start_time) <= unixepoch(?) AND unixepoch(paper_publications.end_time) >= unixepoch(?)", true, now, now)

	if len(classIDs) == 0 {
		query = query.Where("paper_publications.class_id IS NULL")
	} else {
		query = query.Where("paper_publications.class_id IS NULL OR paper_publications.class_id IN ?", classIDs)
	}

	if err := query.Group("papers.id").Find(&papers).Error; err != nil {
		return nil, err
	}
	return papers, nil
}

// FindAttemptByStudentAndPaper returns any attempt (including submitted) for a student+paper.
func (r *ExamRepository) FindAttemptByStudentAndPaper(ctx context.Context, studentID, paperID uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Where("student_id = ? AND paper_id = ?", studentID, paperID).Order("created_at DESC").First(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// CreateProctorEvent creates a proctor event record.
func (r *ExamRepository) CreateProctorEvent(ctx context.Context, event *models.ProctorEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// UpdateAnswer updates an answer's grading fields.
func (r *ExamRepository) UpdateAnswer(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExamAnswer{}).Where("id = ?", id).Updates(updates).Error
}

// FindPublicationByPaperIDForExam returns the publication for exam lookup.
func (r *ExamRepository) FindPublicationByPaperIDForExam(ctx context.Context, paperID uint, classIDs []uint) (*models.PaperPublication, error) {
	var pub models.PaperPublication
	query := r.db.WithContext(ctx).Where("paper_id = ? AND is_published = ?", paperID, true)
	if len(classIDs) == 0 {
		query = query.Where("class_id IS NULL")
	} else {
		query = query.Where("class_id IS NULL OR class_id IN ?", classIDs)
	}

	if err := query.Order("created_at DESC").First(&pub).Error; err != nil {
		return nil, err
	}
	return &pub, nil
}
