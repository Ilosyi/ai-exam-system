package repositories

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentRepository_CreateListAndGetDocument(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewDocumentRepository(root)

	course := CourseDocument{
		ID:          "backend-go",
		Title:       "服务端训练营",
		Description: "Go 后端基础课",
		Order:       2,
	}
	if err := repo.SaveCourse(ctx, course); err != nil {
		t.Fatalf("SaveCourse failed: %v", err)
	}

	doc := DocumentMeta{ID: "day01-gin", Title: "DAY01 - Gin 基础", Order: 1}
	if err := repo.SaveDocument(ctx, "backend-go", doc, "# Gin 基础\n\nHello"); err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	courses, err := repo.ListCourses(ctx)
	if err != nil {
		t.Fatalf("ListCourses failed: %v", err)
	}
	if len(courses) != 1 {
		t.Fatalf("expected 1 course, got %d", len(courses))
	}
	if courses[0].ID != "backend-go" || len(courses[0].Documents) != 1 {
		t.Fatalf("unexpected course: %#v", courses[0])
	}

	loaded, markdown, err := repo.GetDocument(ctx, "backend-go", "day01-gin")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if loaded.Title != "DAY01 - Gin 基础" {
		t.Fatalf("unexpected title: %s", loaded.Title)
	}
	if markdown != "# Gin 基础\n\nHello" {
		t.Fatalf("unexpected markdown: %q", markdown)
	}
}

func TestDocumentRepository_RejectsUnsafeSlug(t *testing.T) {
	ctx := context.Background()
	repo := NewDocumentRepository(t.TempDir())

	err := repo.SaveCourse(ctx, CourseDocument{ID: "../escape", Title: "bad"})
	if !errors.Is(err, ErrInvalidDocumentSlug) {
		t.Fatalf("expected ErrInvalidDocumentSlug, got %v", err)
	}

	err = repo.SaveDocument(ctx, "backend-go", DocumentMeta{ID: "../escape", Title: "bad"}, "x")
	if !errors.Is(err, ErrInvalidDocumentSlug) {
		t.Fatalf("expected ErrInvalidDocumentSlug for document, got %v", err)
	}
}

func TestDocumentRepository_DeleteCourseRemovesDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewDocumentRepository(root)

	if err := repo.SaveCourse(ctx, CourseDocument{ID: "backend-go", Title: "服务端训练营"}); err != nil {
		t.Fatalf("SaveCourse failed: %v", err)
	}
	if err := repo.DeleteCourse(ctx, "backend-go"); err != nil {
		t.Fatalf("DeleteCourse failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "backend-go")); !os.IsNotExist(err) {
		t.Fatalf("expected directory to be removed, stat err=%v", err)
	}
}
