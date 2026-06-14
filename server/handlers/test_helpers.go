// ============================================================================
// handlers/test_helpers.go - 测试辅助工具
// ============================================================================
//
// 本文件提供了用于单元测试的辅助工具，包括：
// - TestContext: 测试上下文，包含所有依赖对象
// - SetupTestContext: 初始化测试环境（内存数据库、默认用户）
// - MakeRequest: 创建 HTTP 请求并获取响应
// - GetTokenForUser: 为用户生成 JWT 令牌
// - ParseResponse: 解析 JSON 响应
// - CreateTestQuestion: 创建测试题目
// - CreateTestClass: 创建测试班级
//
// 为什么需要测试辅助工具？
// - 避免在每个测试中重复初始化代码
// - 提供统一的测试环境
// - 简化测试用例的编写
//
// 学习要点：
// - Go 的 testing 包
// - 内存数据库的使用
// - httptest 的使用
// ============================================================================

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestContext 持有测试所需的所有依赖对象。
type TestContext struct {
	DB              *gorm.DB
	UserRepo        *repositories.UserRepository
	QuestionRepo    *repositories.QuestionRepository
	PaperRepo       *repositories.PaperRepository
	ExamRepo        *repositories.ExamRepository
	ClassRepo       *repositories.ClassRepository
	AuthService     *services.AuthService
	AuthHandler     *AuthHandler
	QuestionHandler *QuestionHandler
	PaperHandler    *PaperHandler
	ExamHandler     *ExamHandler
	ClassHandler    *ClassHandler
}

// SetupTestContext 初始化测试环境。
//
// 创建内存数据库、迁移表结构、初始化所有依赖、创建默认用户。
func SetupTestContext(t *testing.T) *TestContext {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Question{},
		&models.Paper{},
		&models.PaperItem{},
		&models.PaperPublication{},
		&models.Class{},
		&models.UserClass{},
		&models.ExamAttempt{},
		&models.ExamAnswer{},
		&models.ProctorEvent{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	questionRepo := repositories.NewQuestionRepository(db)
	paperRepo := repositories.NewPaperRepository(db)
	examRepo := repositories.NewExamRepository(db)
	classRepo := repositories.NewClassRepository(db)

	authService := services.NewAuthService("test-secret", 24*time.Hour)

	adminHash, _ := authService.HashPassword("admin123")
	studentHash, _ := authService.HashPassword("student123")
	teacherHash, _ := authService.HashPassword("teacher123")

	db.Create(&models.User{Username: "admin", Role: "admin", PasswordHash: adminHash, Status: "active"})
	db.Create(&models.User{Username: "student", Role: "student", PasswordHash: studentHash, Status: "active"})
	db.Create(&models.User{Username: "teacher", Role: "teacher", PasswordHash: teacherHash, Status: "active"})

	return &TestContext{
		DB:              db,
		UserRepo:        userRepo,
		QuestionRepo:    questionRepo,
		PaperRepo:       paperRepo,
		ExamRepo:        examRepo,
		ClassRepo:       classRepo,
		AuthService:     authService,
		AuthHandler:     NewAuthHandler(userRepo, authService),
		QuestionHandler: NewQuestionHandler(questionRepo, userRepo),
		PaperHandler:    NewPaperHandler(paperRepo, questionRepo, classRepo),
		ExamHandler:     NewExamHandler(examRepo, paperRepo, questionRepo, classRepo),
		ClassHandler:    NewClassHandler(classRepo),
	}
}

// MakeRequest 创建 HTTP 请求并获取响应。
func MakeRequest(t *testing.T, method string, path string, body interface{}, handler gin.HandlerFunc, token *string) *httptest.ResponseRecorder {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req, err = http.NewRequest(method, path, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, nil)
	}

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)

	return w
}

// GetTokenForUser 为指定用户生成 JWT 令牌。
func (tc *TestContext) GetTokenForUser(t *testing.T, username string) string {
	var user models.User
	tc.DB.Where("username = ?", username).First(&user)
	if user.ID == 0 {
		t.Fatalf("User %s not found", username)
	}

	token, err := tc.AuthService.GenerateToken(&user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	return token
}

// ParseResponse 解析 JSON 响应。
func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), v)
	if err != nil {
		t.Fatalf("Failed to parse response: %v, body: %s", err, w.Body.String())
	}
}

// CreateTestQuestion 创建一个测试题目。
func (tc *TestContext) CreateTestQuestion(t *testing.T, title string, qType string) *models.Question {
	q := &models.Question{
		Title:    title,
		Type:     qType,
		Language: "go",
		Content:  "Test content",
	}
	if err := tc.QuestionRepo.Create(context.Background(), q); err != nil {
		t.Fatalf("Failed to create question: %v", err)
	}
	return q
}

// CreateTestClass 创建一个测试班级。
func (tc *TestContext) CreateTestClass(t *testing.T, teacherID uint, name string) *models.Class {
	c := &models.Class{
		Name:      name,
		TeacherID: teacherID,
	}
	if err := tc.ClassRepo.Create(context.Background(), c); err != nil {
		t.Fatalf("Failed to create class: %v", err)
	}
	return c
}
