// ============================================================================
// repositories/exam_repository.go - 考试数据访问层
// ============================================================================
//
// 本文件封装了所有与考试答题相关的数据库操作，包括：
// - 答题记录（ExamAttempt）的 CRUD
// - 答案（ExamAnswer）的 Upsert
// - 发布记录（PaperPublication）的查询
// - 监考事件（ProctorEvent）的创建
//
// 学习要点：
// - Upsert 操作的实现（存在则更新，不存在则插入）
// - 预加载关联数据（Preload）
// - 时间窗口查询
// ============================================================================

package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/models"
)

// ExamRepository 封装了考试答题相关的数据库操作。
type ExamRepository struct {
	db *gorm.DB
}

// NewExamRepository 创建一个新的 ExamRepository 实例。
func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

// CreateAttempt 创建一个新的答题记录。
func (r *ExamRepository) CreateAttempt(ctx context.Context, attempt *models.ExamAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

// FindAttemptByID 根据 ID 查询答题记录（预加载答案）。
//
// Preload("Answers") 会自动查询 exam_answers 表，填充 ExamAttempt.Answers 字段。
func (r *ExamRepository) FindAttemptByID(ctx context.Context, id uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Preload("Answers").First(&attempt, id).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// FindActiveAttempt 查询学生对某张试卷的进行中的答题记录。
//
// 用于防止学生重复开始答题：
// - 如果已有 in_progress 的记录，直接返回该记录
// - 如果没有，返回 nil（可以开始新的答题）
func (r *ExamRepository) FindActiveAttempt(ctx context.Context, studentID, paperID uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Where("student_id = ? AND paper_id = ? AND status = ?", studentID, paperID, "in_progress").First(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// UpdateAttempt 更新答题记录。
func (r *ExamRepository) UpdateAttempt(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExamAttempt{}).Where("id = ?", id).Updates(updates).Error
}

// UpsertAnswer 创建或更新答案（Upsert 操作）。
//
// Upsert 逻辑：
// 1. 先查询是否存在该答案（通过 attempt_id + question_id 唯一确定）
// 2. 如果不存在，创建新记录
// 3. 如果存在，更新 answer_json 字段
//
// 这种方式避免了"先删除再插入"的开销，也避免了重复记录。
func (r *ExamRepository) UpsertAnswer(ctx context.Context, answer *models.ExamAnswer) error {
	var existing models.ExamAnswer
	err := r.db.WithContext(ctx).Where("attempt_id = ? AND question_id = ?", answer.AttemptID, answer.QuestionID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.db.WithContext(ctx).Create(answer).Error
	}
	if err != nil {
		return err
	}
	// 已存在，更新答案
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"answer_json": answer.AnswerJSON,
	}).Error
}

// FindAnswer 查询指定答题记录中某道题的答案。
func (r *ExamRepository) FindAnswer(ctx context.Context, attemptID, questionID uint) (*models.ExamAnswer, error) {
	var answer models.ExamAnswer
	if err := r.db.WithContext(ctx).Where("attempt_id = ? AND question_id = ?", attemptID, questionID).First(&answer).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

// ListPublishedPapers 查询学生可见的已发布试卷。
//
// 查询逻辑：
// 1. 通过 JOIN 关联 paper_publications 表
// 2. 筛选条件：is_published = true 且当前时间在时间窗口内
// 3. 班级筛选：公共试卷（class_id IS NULL）或学生所在班级的试卷
//
// 参数：
//   - classIDs: 学生所属班级的 ID 列表
func (r *ExamRepository) ListPublishedPapers(ctx context.Context, classIDs []uint) ([]models.Paper, error) {
	var papers []models.Paper
	now := time.Now()
	query := r.db.WithContext(ctx).
		Joins("JOIN paper_publications ON paper_publications.paper_id = papers.id").
		Where("paper_publications.is_published = ? AND unixepoch(paper_publications.start_time) <= unixepoch(?) AND unixepoch(paper_publications.end_time) >= unixepoch(?)", true, now, now)

	if len(classIDs) == 0 {
		// 学生不属于任何班级，只能看到公共试卷
		query = query.Where("paper_publications.class_id IS NULL")
	} else {
		// 学生可以看到公共试卷和自己班级的试卷
		query = query.Where("paper_publications.class_id IS NULL OR paper_publications.class_id IN ?", classIDs)
	}

	if err := query.Group("papers.id").Find(&papers).Error; err != nil {
		return nil, err
	}
	return papers, nil
}

// FindAttemptByStudentAndPaper 查询学生对某张试卷的答题记录（包括已提交的）。
func (r *ExamRepository) FindAttemptByStudentAndPaper(ctx context.Context, studentID, paperID uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.WithContext(ctx).Where("student_id = ? AND paper_id = ?", studentID, paperID).Order("created_at DESC").First(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

// CreateProctorEvent 创建监考事件记录。
func (r *ExamRepository) CreateProctorEvent(ctx context.Context, event *models.ProctorEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// UpdateAnswer 更新答案的阅卷结果。
func (r *ExamRepository) UpdateAnswer(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExamAnswer{}).Where("id = ?", id).Updates(updates).Error
}

// FindPublicationByPaperIDForExam 查询学生可见的试卷发布记录。
//
// 与 ListPublishedPapers 类似，但返回单条发布记录（包含详细的时间窗口和时长信息）。
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
