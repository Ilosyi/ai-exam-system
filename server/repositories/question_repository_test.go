package repositories

import (
	"context"
	"testing"

	"week05/homework/server/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuestionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQuestionRepository(db)
	ctx := context.Background()

	// 创建测试数据
	q1 := &models.Question{
		Title:    "Question 1",
		Type:     "single",
		Language: "go",
		Content:  "What is Go?",
	}
	q2 := &models.Question{
		Title:    "Question 2",
		Type:     "multiple",
		Language: "python",
		Content:  "Python basics",
	}

	db.Create(q1)
	db.Create(q2)

	// 测试：获取所有问题
	questions, total, err := repo.List(ctx, QuestionFilters{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, questions, 2)
}

func TestQuestionRepository_List_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQuestionRepository(db)
	ctx := context.Background()

	// 创建测试数据
	q1 := &models.Question{
		Title:    "Go Question",
		Type:     "single",
		Language: "go",
		Content:  "Go language",
	}
	q2 := &models.Question{
		Title:    "Python Question",
		Type:     "multiple",
		Language: "python",
		Content:  "Python basics",
	}

	db.Create(q1)
	db.Create(q2)

	// 测试：按语言过滤
	questions, total, err := repo.List(ctx, QuestionFilters{
		Language: "go",
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, questions, 1)
	assert.Equal(t, "Go Question", questions[0].Title)

	// 测试：按类型过滤
	questions, total, err = repo.List(ctx, QuestionFilters{
		Type:     "multiple",
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, questions, 1)
	assert.Equal(t, "Python Question", questions[0].Title)
}

func TestQuestionRepository_List_WithKeyword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQuestionRepository(db)
	ctx := context.Background()

	// 创建测试数据
	q1 := &models.Question{
		Title:   "String Methods in Go",
		Type:    "single",
		Content: "How to use strings in Go?",
	}
	q2 := &models.Question{
		Title:   "Array Operations",
		Type:    "multiple",
		Content: "Python list operations",
	}

	db.Create(q1)
	db.Create(q2)

	// 测试：按关键词搜索
	questions, total, err := repo.List(ctx, QuestionFilters{
		Keyword:  "Go",
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, questions, 1)
	assert.Equal(t, "String Methods in Go", questions[0].Title)
}

func TestQuestionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQuestionRepository(db)
	ctx := context.Background()

	q := &models.Question{
		Title:    "New Question",
		Type:     "coding",
		Language: "javascript",
		Content:  "Write a function",
	}

	err := repo.Create(ctx, q)
	assert.NoError(t, err)
	assert.NotZero(t, q.ID)

	// 验证数据被正确保存
	var retrieved models.Question
	db.First(&retrieved, q.ID)
	assert.Equal(t, "New Question", retrieved.Title)
	assert.Equal(t, "javascript", retrieved.Language)
}

func TestQuestionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQuestionRepository(db)
	ctx := context.Background()

	// 创建初始数据
	q := &models.Question{
		Title:    "Original Title",
		Type:     "single",
		Language: "python",
		Content:  "Original content",
	}
	db.Create(q)

	// 更新
	err := repo.Update(ctx, q.ID, map[string]interface{}{
		"title": "Updated Title",
		"type":  "multiple",
	})
	assert.NoError(t, err)

	// 验证更新
	var updated models.Question
	db.First(&updated, q.ID)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "multiple", updated.Type)
	assert.Equal(t, "python", updated.Language)
}

// setupTestDB 创建一个内存 SQLite 数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.Question{},
		&models.Paper{},
		&models.PaperItem{},
		&models.PaperPublication{},
		&models.ExamAttempt{},
		&models.ExamAnswer{},
		&models.User{},
		&models.Class{},
		&models.UserClass{},
		&models.ProctorEvent{},
	)
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	return db
}
