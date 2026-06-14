package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/handlers"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

type dependencies struct {
	authService     *services.AuthService
	userRepo        *repositories.UserRepository
	questionHandler *handlers.QuestionHandler
	paperHandler    *handlers.PaperHandler
	examHandler     *handlers.ExamHandler
	aiHandler       *handlers.AIHandler
	noteHandler     *handlers.NoteHandler
	authHandler     *handlers.AuthHandler
	classHandler    *handlers.ClassHandler
	documentHandler *handlers.DocumentHandler
}

func initDependencies(projectRoot string, db *gorm.DB) (*dependencies, error) {
	questionRepo := repositories.NewQuestionRepository(db)
	paperRepo := repositories.NewPaperRepository(db)
	examRepo := repositories.NewExamRepository(db)
	userRepo := repositories.NewUserRepository(db)
	classRepo := repositories.NewClassRepository(db)
	documentRepo := repositories.NewDocumentRepository(filepath.Join(projectRoot, "course-docs"))
	documentService := services.NewDocumentService(documentRepo)

	aiService, err := services.NewAIService(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("ai service init failed: %w", err)
	}

	authService := services.NewAuthService(os.Getenv("JWT_SECRET"), 24*time.Hour)
	if err := seedDefaultUsers(context.Background(), userRepo, authService); err != nil {
		return nil, fmt.Errorf("seed users failed: %w", err)
	}

	return &dependencies{
		authService:     authService,
		userRepo:        userRepo,
		questionHandler: handlers.NewQuestionHandler(questionRepo, userRepo),
		paperHandler:    handlers.NewPaperHandler(paperRepo, questionRepo, classRepo),
		examHandler:     handlers.NewExamHandler(examRepo, paperRepo, questionRepo, classRepo),
		aiHandler:       handlers.NewAIHandler(aiService),
		noteHandler:     handlers.NewNoteHandler(projectRoot),
		authHandler:     handlers.NewAuthHandler(userRepo, authService),
		classHandler:    handlers.NewClassHandler(classRepo),
		documentHandler: handlers.NewDocumentHandler(documentService),
	}, nil
}

func seedDefaultUsers(ctx context.Context, userRepo *repositories.UserRepository, authService *services.AuthService) error {
	adminHash, err := authService.HashPassword("admin123")
	if err != nil {
		return err
	}
	teacherHash, err := authService.HashPassword("teacher123")
	if err != nil {
		return err
	}
	studentHash, err := authService.HashPassword("student123")
	if err != nil {
		return err
	}

	return userRepo.EnsureDefaults(ctx, []models.User{
		{Username: "admin", Role: "admin", PasswordHash: adminHash, Status: "active"},
		{Username: "teacher01", Role: "teacher", PasswordHash: teacherHash, Status: "active"},
		{Username: "student01", Role: "student", PasswordHash: studentHash, Status: "active"},
	})
}
