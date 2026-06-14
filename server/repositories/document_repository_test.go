package repositories

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestDocumentRepository_RejectsSymlinkCourseDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	outside := t.TempDir()
	repo := NewDocumentRepository(root)

	if err := os.Symlink(outside, filepath.Join(root, "backend-go")); err != nil {
		t.Fatalf("Symlink failed: %v", err)
	}

	err := repo.SaveCourse(ctx, CourseDocument{ID: "backend-go", Title: "服务端训练营"})
	if !errors.Is(err, ErrInvalidDocumentSlug) {
		t.Fatalf("expected ErrInvalidDocumentSlug for symlink course directory, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "course.json")); !os.IsNotExist(err) {
		t.Fatalf("expected outside course.json to remain absent, stat err=%v", err)
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

func TestDocumentRepository_DeleteDocumentRemovesCourseIndex(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewDocumentRepository(root)

	if err := repo.SaveCourse(ctx, CourseDocument{ID: "backend-go", Title: "服务端训练营"}); err != nil {
		t.Fatalf("SaveCourse failed: %v", err)
	}
	if err := repo.SaveDocument(ctx, "backend-go", DocumentMeta{ID: "day01-gin", Title: "DAY01", Order: 1}, "hello"); err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}
	if err := repo.DeleteDocument(ctx, "backend-go", "day01-gin"); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	course, err := repo.GetCourse(ctx, "backend-go")
	if err != nil {
		t.Fatalf("GetCourse failed: %v", err)
	}
	if len(course.Documents) != 0 {
		t.Fatalf("expected document index to be empty, got %#v", course.Documents)
	}
}

func TestDocumentRepository_SaveDocumentLeavesIndexedMarkdownReadable(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewDocumentRepository(root)

	if err := repo.SaveCourse(ctx, CourseDocument{ID: "backend-go", Title: "服务端训练营"}); err != nil {
		t.Fatalf("SaveCourse failed: %v", err)
	}
	if err := repo.SaveDocument(ctx, "backend-go", DocumentMeta{ID: "day01-gin", Title: "DAY01", Order: 1}, "hello"); err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	course, err := repo.GetCourse(ctx, "backend-go")
	if err != nil {
		t.Fatalf("GetCourse failed: %v", err)
	}
	if len(course.Documents) != 1 {
		t.Fatalf("expected one indexed document, got %#v", course.Documents)
	}
	if _, err := os.Stat(filepath.Join(root, "backend-go", course.Documents[0].ID+".md")); err != nil {
		t.Fatalf("expected indexed markdown to exist: %v", err)
	}
}

func TestDocumentRepository_DeleteDocumentRemovesMarkdownBody(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewDocumentRepository(root)

	if err := repo.SaveCourse(ctx, CourseDocument{ID: "backend-go", Title: "服务端训练营"}); err != nil {
		t.Fatalf("SaveCourse failed: %v", err)
	}
	if err := repo.SaveDocument(ctx, "backend-go", DocumentMeta{ID: "day01-gin", Title: "DAY01", Order: 1}, "hello"); err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}
	if err := repo.DeleteDocument(ctx, "backend-go", "day01-gin"); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "backend-go", "day01-gin.md")); !os.IsNotExist(err) {
		t.Fatalf("expected markdown body to be removed, stat err=%v", err)
	}
}

func TestWriteFileAtomicallyReplacesFileAndCleansTemp(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "course.json")

	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed target failed: %v", err)
	}
	if err := writeFileAtomically(target, []byte("new"), 0o644); err != nil {
		t.Fatalf("writeFileAtomically failed: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("expected replaced content, got %q", content)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Fatalf("expected temp files to be cleaned, found %s", entry.Name())
		}
	}
}
