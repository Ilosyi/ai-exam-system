# Course Documents Home Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build file-backed course Markdown document management, replace the student exam entrance with `/home`, and add a Markdown reading experience.

**Architecture:** Store course metadata and Markdown content under project-level `course-docs/`. Add a Go document repository/service/handler stack with read routes for all authenticated users and write routes for teachers/admins. Add React API/types/hooks/pages for admin document management, student home, and Markdown reading while leaving exam take/result flows unchanged.

**Tech Stack:** Go 1.22, Gin, file system JSON/Markdown storage, React 18, TypeScript, Vite, Ant Design, react-markdown, remark-gfm, rehype-highlight, github-markdown-css.

---

## Files And Responsibilities

- Create `server/repositories/document_repository.go`: low-level safe file system operations for `course-docs/`.
- Create `server/repositories/document_repository_test.go`: repository unit tests for slug validation, CRUD, sorting, and traversal rejection.
- Create `server/services/document_service.go`: business layer for course/doc validation and orchestration.
- Create `server/handlers/document_handler.go`: HTTP request binding and JSON responses for document APIs.
- Modify `server/bootstrap/dependencies.go`: instantiate document repository/service/handler.
- Modify `server/bootstrap/router.go`: register authenticated read routes and teacher/admin write routes.
- Create `client/src/api/document.ts`: typed document API wrapper functions.
- Modify `client/src/api/api.test.ts`: assert document API exports.
- Create `client/src/hooks/useDocuments.ts`: shared loading/saving logic for course documents.
- Create `client/src/pages/DocumentManagePage.tsx`: teacher/admin management UI.
- Create `client/src/pages/HomePage.tsx`: student home page with profile, courses, exams.
- Create `client/src/pages/DocumentReaderPage.tsx`: Markdown reading page.
- Modify `client/src/App.tsx`: add routes and `/exam` redirect.
- Modify `client/src/components/AppLayout.tsx`: add “文档管理” menu item.
- Modify `client/src/hooks/useAuth.tsx`: change student default route to `/home`.
- Modify `client/src/index.css`: add student home and document reader styles.
- Create `course-docs/backend-go/course.json`: seed example course metadata.
- Create `course-docs/backend-go/day01-gin.md`: seed example Markdown content for smoke testing.
- Modify `AGENTS.md`: document the new module, routes, permissions, and storage directory.
- Modify `docs/backend-architecture.md`: document file-backed document module and routes.

## Task 1: Backend Document Repository

**Files:**
- Create: `server/repositories/document_repository.go`
- Create: `server/repositories/document_repository_test.go`

- [ ] **Step 1: Write failing repository tests**

Create `server/repositories/document_repository_test.go`:

```go
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
```

- [ ] **Step 2: Run repository tests and verify they fail**

Run:

```bash
cd server
go test ./repositories -run DocumentRepository -count=1
```

Expected: fail to compile because `NewDocumentRepository`, `CourseDocument`, `DocumentMeta`, and `ErrInvalidDocumentSlug` are undefined.

- [ ] **Step 3: Implement repository**

Create `server/repositories/document_repository.go`:

```go
package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrInvalidDocumentSlug = errors.New("invalid document slug")
	ErrDocumentNotFound    = errors.New("document not found")
)

type DocumentMeta struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type CourseDocument struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Order       int            `json:"order"`
	Documents   []DocumentMeta `json:"documents"`
}

type DocumentRepository struct {
	root string
}

var documentSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

func NewDocumentRepository(root string) *DocumentRepository {
	return &DocumentRepository{root: root}
}

func (r *DocumentRepository) ListCourses(ctx context.Context) ([]CourseDocument, error) {
	if err := os.MkdirAll(r.root, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return nil, err
	}
	courses := make([]CourseDocument, 0, len(entries))
	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !entry.IsDir() {
			continue
		}
		course, err := r.GetCourse(ctx, entry.Name())
		if err != nil {
			if errors.Is(err, ErrDocumentNotFound) {
				continue
			}
			return nil, err
		}
		courses = append(courses, course)
	}
	sortCourses(courses)
	return courses, nil
}

func (r *DocumentRepository) GetCourse(ctx context.Context, courseID string) (CourseDocument, error) {
	if ctx.Err() != nil {
		return CourseDocument{}, ctx.Err()
	}
	coursePath, err := r.coursePath(courseID)
	if err != nil {
		return CourseDocument{}, err
	}
	content, err := os.ReadFile(filepath.Join(coursePath, "course.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return CourseDocument{}, ErrDocumentNotFound
		}
		return CourseDocument{}, err
	}
	var course CourseDocument
	if err := json.Unmarshal(content, &course); err != nil {
		return CourseDocument{}, err
	}
	sortDocuments(course.Documents)
	return course, nil
}

func (r *DocumentRepository) SaveCourse(ctx context.Context, course CourseDocument) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !isValidDocumentSlug(course.ID) {
		return ErrInvalidDocumentSlug
	}
	coursePath, err := r.coursePath(course.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(coursePath, 0755); err != nil {
		return err
	}
	sortDocuments(course.Documents)
	return writeJSON(filepath.Join(coursePath, "course.json"), course)
}

func (r *DocumentRepository) DeleteCourse(ctx context.Context, courseID string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	coursePath, err := r.coursePath(courseID)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(coursePath); err != nil {
		return err
	}
	return nil
}

func (r *DocumentRepository) GetDocument(ctx context.Context, courseID string, docID string) (DocumentMeta, string, error) {
	course, err := r.GetCourse(ctx, courseID)
	if err != nil {
		return DocumentMeta{}, "", err
	}
	doc, ok := findDocument(course.Documents, docID)
	if !ok {
		return DocumentMeta{}, "", ErrDocumentNotFound
	}
	docPath, err := r.documentPath(courseID, docID)
	if err != nil {
		return DocumentMeta{}, "", err
	}
	content, err := os.ReadFile(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DocumentMeta{}, "", ErrDocumentNotFound
		}
		return DocumentMeta{}, "", err
	}
	return doc, string(content), nil
}

func (r *DocumentRepository) SaveDocument(ctx context.Context, courseID string, doc DocumentMeta, markdown string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !isValidDocumentSlug(doc.ID) {
		return ErrInvalidDocumentSlug
	}
	course, err := r.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}
	replaced := false
	for i := range course.Documents {
		if course.Documents[i].ID == doc.ID {
			course.Documents[i] = doc
			replaced = true
			break
		}
	}
	if !replaced {
		course.Documents = append(course.Documents, doc)
	}
	docPath, err := r.documentPath(courseID, doc.ID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(docPath, []byte(markdown), 0644); err != nil {
		return err
	}
	return r.SaveCourse(ctx, course)
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, courseID string, docID string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	course, err := r.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}
	next := course.Documents[:0]
	found := false
	for _, doc := range course.Documents {
		if doc.ID == docID {
			found = true
			continue
		}
		next = append(next, doc)
	}
	if !found {
		return ErrDocumentNotFound
	}
	course.Documents = next
	docPath, err := r.documentPath(courseID, docID)
	if err != nil {
		return err
	}
	if err := os.Remove(docPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return r.SaveCourse(ctx, course)
}

func (r *DocumentRepository) coursePath(courseID string) (string, error) {
	if !isValidDocumentSlug(courseID) {
		return "", ErrInvalidDocumentSlug
	}
	return safeJoin(r.root, courseID)
}

func (r *DocumentRepository) documentPath(courseID string, docID string) (string, error) {
	if !isValidDocumentSlug(docID) {
		return "", ErrInvalidDocumentSlug
	}
	coursePath, err := r.coursePath(courseID)
	if err != nil {
		return "", err
	}
	return safeJoin(coursePath, docID+".md")
}

func isValidDocumentSlug(value string) bool {
	return documentSlugPattern.MatchString(value)
}

func safeJoin(root string, parts ...string) (string, error) {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := filepath.Join(append([]string{cleanRoot}, parts...)...)
	cleanTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if cleanTarget != cleanRoot && !strings.HasPrefix(cleanTarget, cleanRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: path outside document root", ErrInvalidDocumentSlug)
	}
	return cleanTarget, nil
}

func writeJSON(path string, value interface{}) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0644)
}

func findDocument(documents []DocumentMeta, docID string) (DocumentMeta, bool) {
	for _, doc := range documents {
		if doc.ID == docID {
			return doc, true
		}
	}
	return DocumentMeta{}, false
}

func sortCourses(courses []CourseDocument) {
	sort.SliceStable(courses, func(i, j int) bool {
		if courses[i].Order == courses[j].Order {
			return courses[i].Title < courses[j].Title
		}
		return courses[i].Order < courses[j].Order
	})
}

func sortDocuments(documents []DocumentMeta) {
	sort.SliceStable(documents, func(i, j int) bool {
		if documents[i].Order == documents[j].Order {
			return documents[i].Title < documents[j].Title
		}
		return documents[i].Order < documents[j].Order
	})
}
```

- [ ] **Step 4: Run repository tests and verify they pass**

Run:

```bash
cd server
go test ./repositories -run DocumentRepository -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit backend repository**

```bash
git add server/repositories/document_repository.go server/repositories/document_repository_test.go
git commit -m "feat: add course document repository"
```

## Task 2: Backend Service, Handler, And Routes

**Files:**
- Create: `server/services/document_service.go`
- Create: `server/handlers/document_handler.go`
- Modify: `server/bootstrap/dependencies.go`
- Modify: `server/bootstrap/router.go`

- [ ] **Step 1: Write service implementation**

Create `server/services/document_service.go`:

```go
package services

import (
	"context"
	"errors"
	"strings"

	"week05/homework/server/repositories"
)

var (
	ErrDocumentTitleRequired = errors.New("document title is required")
	ErrDocumentSlugConflict  = errors.New("document slug already exists")
)

type CourseInput struct {
	ID          string
	Title       string
	Description string
	Order       int
}

type DocumentInput struct {
	ID       string
	Title    string
	Order    int
	Markdown string
}

type DocumentDetail struct {
	repositories.DocumentMeta
	Markdown string `json:"markdown"`
}

type DocumentService struct {
	repo *repositories.DocumentRepository
}

func NewDocumentService(repo *repositories.DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) ListCourses(ctx context.Context) ([]repositories.CourseDocument, error) {
	return s.repo.ListCourses(ctx)
}

func (s *DocumentService) GetCourse(ctx context.Context, courseID string) (repositories.CourseDocument, error) {
	return s.repo.GetCourse(ctx, courseID)
}

func (s *DocumentService) CreateCourse(ctx context.Context, input CourseInput) (repositories.CourseDocument, error) {
	course, err := buildCourse(input, nil)
	if err != nil {
		return repositories.CourseDocument{}, err
	}
	if _, err := s.repo.GetCourse(ctx, course.ID); err == nil {
		return repositories.CourseDocument{}, ErrDocumentSlugConflict
	} else if !errors.Is(err, repositories.ErrDocumentNotFound) {
		return repositories.CourseDocument{}, err
	}
	if err := s.repo.SaveCourse(ctx, course); err != nil {
		return repositories.CourseDocument{}, err
	}
	return course, nil
}

func (s *DocumentService) UpdateCourse(ctx context.Context, courseID string, input CourseInput) (repositories.CourseDocument, error) {
	existing, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return repositories.CourseDocument{}, err
	}
	input.ID = courseID
	course, err := buildCourse(input, existing.Documents)
	if err != nil {
		return repositories.CourseDocument{}, err
	}
	if err := s.repo.SaveCourse(ctx, course); err != nil {
		return repositories.CourseDocument{}, err
	}
	return course, nil
}

func (s *DocumentService) DeleteCourse(ctx context.Context, courseID string) error {
	return s.repo.DeleteCourse(ctx, courseID)
}

func (s *DocumentService) GetDocument(ctx context.Context, courseID string, docID string) (DocumentDetail, error) {
	meta, markdown, err := s.repo.GetDocument(ctx, courseID, docID)
	if err != nil {
		return DocumentDetail{}, err
	}
	return DocumentDetail{DocumentMeta: meta, Markdown: markdown}, nil
}

func (s *DocumentService) CreateDocument(ctx context.Context, courseID string, input DocumentInput) (DocumentDetail, error) {
	doc, err := buildDocument(input)
	if err != nil {
		return DocumentDetail{}, err
	}
	course, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return DocumentDetail{}, err
	}
	for _, existing := range course.Documents {
		if existing.ID == doc.ID {
			return DocumentDetail{}, ErrDocumentSlugConflict
		}
	}
	if err := s.repo.SaveDocument(ctx, courseID, doc, input.Markdown); err != nil {
		return DocumentDetail{}, err
	}
	return DocumentDetail{DocumentMeta: doc, Markdown: input.Markdown}, nil
}

func (s *DocumentService) UpdateDocument(ctx context.Context, courseID string, docID string, input DocumentInput) (DocumentDetail, error) {
	input.ID = docID
	doc, err := buildDocument(input)
	if err != nil {
		return DocumentDetail{}, err
	}
	if _, _, err := s.repo.GetDocument(ctx, courseID, docID); err != nil {
		return DocumentDetail{}, err
	}
	if err := s.repo.SaveDocument(ctx, courseID, doc, input.Markdown); err != nil {
		return DocumentDetail{}, err
	}
	return DocumentDetail{DocumentMeta: doc, Markdown: input.Markdown}, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, courseID string, docID string) error {
	return s.repo.DeleteDocument(ctx, courseID, docID)
}

func buildCourse(input CourseInput, documents []repositories.DocumentMeta) (repositories.CourseDocument, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return repositories.CourseDocument{}, ErrDocumentTitleRequired
	}
	return repositories.CourseDocument{
		ID:          strings.TrimSpace(input.ID),
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Order:       input.Order,
		Documents:   documents,
	}, nil
}

func buildDocument(input DocumentInput) (repositories.DocumentMeta, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return repositories.DocumentMeta{}, ErrDocumentTitleRequired
	}
	return repositories.DocumentMeta{
		ID:    strings.TrimSpace(input.ID),
		Title: title,
		Order: input.Order,
	}, nil
}
```

- [ ] **Step 2: Write handler implementation**

Create `server/handlers/document_handler.go`:

```go
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

type DocumentHandler struct {
	service *services.DocumentService
}

type coursePayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type documentPayload struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Order    int    `json:"order"`
	Markdown string `json:"markdown"`
}

func NewDocumentHandler(service *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

func (h *DocumentHandler) ListCourses(c *gin.Context) {
	courses, err := h.service.ListCourses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "读取课程文档失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": courses, "total": len(courses)})
}

func (h *DocumentHandler) GetCourse(c *gin.Context) {
	course, err := h.service.GetCourse(c.Request.Context(), c.Param("courseId"))
	if err != nil {
		respondDocumentError(c, err, "读取课程失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": course})
}

func (h *DocumentHandler) CreateCourse(c *gin.Context) {
	var payload coursePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "课程参数错误", "error": err.Error()})
		return
	}
	course, err := h.service.CreateCourse(c.Request.Context(), services.CourseInput{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Order:       payload.Order,
	})
	if err != nil {
		respondDocumentError(c, err, "创建课程失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "课程创建成功", "data": course})
}

func (h *DocumentHandler) UpdateCourse(c *gin.Context) {
	var payload coursePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "课程参数错误", "error": err.Error()})
		return
	}
	course, err := h.service.UpdateCourse(c.Request.Context(), c.Param("courseId"), services.CourseInput{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Order:       payload.Order,
	})
	if err != nil {
		respondDocumentError(c, err, "更新课程失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "课程更新成功", "data": course})
}

func (h *DocumentHandler) DeleteCourse(c *gin.Context) {
	if err := h.service.DeleteCourse(c.Request.Context(), c.Param("courseId")); err != nil {
		respondDocumentError(c, err, "删除课程失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "课程删除成功"})
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	doc, err := h.service.GetDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId"))
	if err != nil {
		respondDocumentError(c, err, "读取文档失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": doc})
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var payload documentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文档参数错误", "error": err.Error()})
		return
	}
	doc, err := h.service.CreateDocument(c.Request.Context(), c.Param("courseId"), services.DocumentInput{
		ID:       payload.ID,
		Title:    payload.Title,
		Order:    payload.Order,
		Markdown: payload.Markdown,
	})
	if err != nil {
		respondDocumentError(c, err, "创建文档失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "文档创建成功", "data": doc})
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	var payload documentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文档参数错误", "error": err.Error()})
		return
	}
	doc, err := h.service.UpdateDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId"), services.DocumentInput{
		ID:       payload.ID,
		Title:    payload.Title,
		Order:    payload.Order,
		Markdown: payload.Markdown,
	})
	if err != nil {
		respondDocumentError(c, err, "更新文档失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "文档更新成功", "data": doc})
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	if err := h.service.DeleteDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId")); err != nil {
		respondDocumentError(c, err, "删除文档失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "文档删除成功"})
}

func respondDocumentError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, repositories.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": message, "error": err.Error()})
	case errors.Is(err, repositories.ErrInvalidDocumentSlug),
		errors.Is(err, services.ErrDocumentTitleRequired),
		errors.Is(err, services.ErrDocumentSlugConflict):
		c.JSON(http.StatusBadRequest, gin.H{"message": message, "error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": message, "error": err.Error()})
	}
}
```

- [ ] **Step 3: Wire dependencies**

Modify `server/bootstrap/dependencies.go`:

1. Add a field to `dependencies`:

```go
documentHandler *handlers.DocumentHandler
```

2. In `initDependencies`, after `classRepo := ...`, add:

```go
documentRepo := repositories.NewDocumentRepository(filepath.Join(projectRoot, "course-docs"))
documentService := services.NewDocumentService(documentRepo)
```

3. Add `path/filepath` to imports if missing.

4. Add handler construction in the returned struct:

```go
documentHandler: handlers.NewDocumentHandler(documentService),
```

- [ ] **Step 4: Register routes**

Modify `server/bootstrap/router.go` inside `protected` after auth routes:

```go
protected.GET("/documents/courses", deps.documentHandler.ListCourses)
protected.GET("/documents/courses/:courseId", deps.documentHandler.GetCourse)
protected.GET("/documents/courses/:courseId/docs/:docId", deps.documentHandler.GetDocument)
```

Inside `teacherRoutes`, add:

```go
teacherRoutes.POST("/documents/courses", deps.documentHandler.CreateCourse)
teacherRoutes.PUT("/documents/courses/:courseId", deps.documentHandler.UpdateCourse)
teacherRoutes.DELETE("/documents/courses/:courseId", deps.documentHandler.DeleteCourse)
teacherRoutes.POST("/documents/courses/:courseId/docs", deps.documentHandler.CreateDocument)
teacherRoutes.PUT("/documents/courses/:courseId/docs/:docId", deps.documentHandler.UpdateDocument)
teacherRoutes.DELETE("/documents/courses/:courseId/docs/:docId", deps.documentHandler.DeleteDocument)
```

- [ ] **Step 5: Build backend**

Run:

```bash
cd server
go test ./repositories -run DocumentRepository -count=1
go build ./...
```

Expected: both commands pass.

- [ ] **Step 6: Commit backend API**

```bash
git add server/services/document_service.go server/handlers/document_handler.go server/bootstrap/dependencies.go server/bootstrap/router.go
git commit -m "feat: add document management api"
```

## Task 3: Frontend Document API And Shared Hook

**Files:**
- Create: `client/src/api/document.ts`
- Create: `client/src/hooks/useDocuments.ts`
- Modify: `client/src/api/api.test.ts`

- [ ] **Step 1: Add document API module**

Create `client/src/api/document.ts`:

```ts
import { apiDelete, apiGet, apiPost, apiPut } from "./client";

export interface DocumentMeta {
  id: string;
  title: string;
  order: number;
}

export interface CourseDocument {
  id: string;
  title: string;
  description: string;
  order: number;
  documents: DocumentMeta[];
}

export interface DocumentDetail extends DocumentMeta {
  markdown: string;
}

export interface CourseInput {
  id: string;
  title: string;
  description: string;
  order: number;
}

export interface DocumentInput {
  id: string;
  title: string;
  order: number;
  markdown: string;
}

export async function fetchDocumentCourses(): Promise<{ data: CourseDocument[]; total: number }> {
  return apiGet<{ data: CourseDocument[]; total: number }>("/documents/courses");
}

export async function fetchDocumentCourse(courseId: string): Promise<{ data: CourseDocument }> {
  return apiGet<{ data: CourseDocument }>(`/documents/courses/${courseId}`);
}

export async function createDocumentCourse(input: CourseInput): Promise<{ message: string; data: CourseDocument }> {
  return apiPost<{ message: string; data: CourseDocument }>("/documents/courses", input);
}

export async function updateDocumentCourse(courseId: string, input: CourseInput): Promise<{ message: string; data: CourseDocument }> {
  return apiPut<{ message: string; data: CourseDocument }>(`/documents/courses/${courseId}`, input);
}

export async function deleteDocumentCourse(courseId: string): Promise<{ message: string }> {
  return apiDelete<{ message: string }>(`/documents/courses/${courseId}`);
}

export async function fetchDocumentDetail(courseId: string, docId: string): Promise<{ data: DocumentDetail }> {
  return apiGet<{ data: DocumentDetail }>(`/documents/courses/${courseId}/docs/${docId}`);
}

export async function createDocument(courseId: string, input: DocumentInput): Promise<{ message: string; data: DocumentDetail }> {
  return apiPost<{ message: string; data: DocumentDetail }>(`/documents/courses/${courseId}/docs`, input);
}

export async function updateDocument(courseId: string, docId: string, input: DocumentInput): Promise<{ message: string; data: DocumentDetail }> {
  return apiPut<{ message: string; data: DocumentDetail }>(`/documents/courses/${courseId}/docs/${docId}`, input);
}

export async function deleteDocument(courseId: string, docId: string): Promise<{ message: string }> {
  return apiDelete<{ message: string }>(`/documents/courses/${courseId}/docs/${docId}`);
}
```

- [ ] **Step 2: Add shared hook**

Create `client/src/hooks/useDocuments.ts`:

```ts
import { useCallback, useEffect, useState } from "react";
import { message } from "antd";
import {
  fetchDocumentCourses,
  fetchDocumentDetail,
  type CourseDocument,
  type DocumentDetail,
} from "../api/document";

export function useDocumentCourses() {
  const [courses, setCourses] = useState<CourseDocument[]>([]);
  const [loading, setLoading] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchDocumentCourses();
      setCourses(res.data || []);
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "读取课程文档失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { courses, loading, reload, setCourses };
}

export function useDocumentDetail(courseId?: string, docId?: string) {
  const [detail, setDetail] = useState<DocumentDetail | null>(null);
  const [loading, setLoading] = useState(false);

  const reload = useCallback(async () => {
    if (!courseId || !docId) {
      setDetail(null);
      return;
    }
    setLoading(true);
    try {
      const res = await fetchDocumentDetail(courseId, docId);
      setDetail(res.data);
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "读取文档失败");
      setDetail(null);
    } finally {
      setLoading(false);
    }
  }, [courseId, docId]);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { detail, loading, reload };
}
```

- [ ] **Step 3: Extend API export smoke test**

Modify `client/src/api/api.test.ts` by adding:

```ts
    it('document module exports expected functions', async () => {
        const documentModule = await import('../api/document')
        expect(documentModule.fetchDocumentCourses).toBeDefined()
        expect(documentModule.fetchDocumentDetail).toBeDefined()
        expect(documentModule.createDocumentCourse).toBeDefined()
    })
```

- [ ] **Step 4: Run frontend API tests**

Run:

```bash
cd client
pnpm test -- api.test.ts --run
```

Expected: PASS.

- [ ] **Step 5: Commit frontend API**

```bash
git add client/src/api/document.ts client/src/hooks/useDocuments.ts client/src/api/api.test.ts
git commit -m "feat: add document api client"
```

## Task 4: Backend Admin Document Management Page

**Files:**
- Create: `client/src/pages/DocumentManagePage.tsx`
- Modify: `client/src/App.tsx`
- Modify: `client/src/components/AppLayout.tsx`

- [ ] **Step 1: Create management page**

Create `client/src/pages/DocumentManagePage.tsx`:

```tsx
import { useMemo, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import { Button, Card, Col, Drawer, Empty, Form, Input, InputNumber, Modal, Row, Space, Spin, Tree, Typography, message } from "antd";
import { DeleteOutlined, EyeOutlined, FileMarkdownOutlined, FolderAddOutlined, PlusOutlined, SaveOutlined } from "@ant-design/icons";
import type { DataNode } from "antd/es/tree";
import {
  createDocument,
  createDocumentCourse,
  deleteDocument,
  deleteDocumentCourse,
  fetchDocumentDetail,
  updateDocument,
  updateDocumentCourse,
  type CourseDocument,
  type DocumentDetail,
} from "../api/document";
import { useDocumentCourses } from "../hooks/useDocuments";

type Selection =
  | { type: "course"; courseId: string }
  | { type: "document"; courseId: string; docId: string };

interface CourseFormValue {
  id: string;
  title: string;
  description: string;
  order: number;
}

interface DocumentFormValue {
  id: string;
  title: string;
  order: number;
  markdown: string;
}

const emptyDocumentMarkdown = "# 新文档\n\n请在这里编写课程内容。";

export function DocumentManagePage() {
  const { courses, loading, reload } = useDocumentCourses();
  const [selection, setSelection] = useState<Selection | null>(null);
  const [detail, setDetail] = useState<DocumentDetail | null>(null);
  const [saving, setSaving] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [courseForm] = Form.useForm<CourseFormValue>();
  const [docForm] = Form.useForm<DocumentFormValue>();
  const previewMarkdown = Form.useWatch("markdown", docForm);

  const selectedCourse = useMemo(
    () => courses.find((course) => course.id === selection?.courseId) ?? null,
    [courses, selection],
  );

  const treeData = useMemo<DataNode[]>(
    () =>
      courses.map((course) => ({
        key: `course:${course.id}`,
        title: course.title,
        icon: <FolderAddOutlined />,
        children: course.documents.map((doc) => ({
          key: `doc:${course.id}:${doc.id}`,
          title: doc.title,
          icon: <FileMarkdownOutlined />,
        })),
      })),
    [courses],
  );

  const handleSelect = async (keys: React.Key[]) => {
    const key = String(keys[0] ?? "");
    if (!key) return;
    const parts = key.split(":");
    if (parts[0] === "course") {
      const course = courses.find((item) => item.id === parts[1]);
      if (!course) return;
      setSelection({ type: "course", courseId: course.id });
      setDetail(null);
      courseForm.setFieldsValue({
        id: course.id,
        title: course.title,
        description: course.description,
        order: course.order,
      });
      return;
    }
    if (parts[0] === "doc") {
      setSelection({ type: "document", courseId: parts[1], docId: parts[2] });
      const res = await fetchDocumentDetail(parts[1], parts[2]);
      setDetail(res.data);
      docForm.setFieldsValue({
        id: res.data.id,
        title: res.data.title,
        order: res.data.order,
        markdown: res.data.markdown,
      });
    }
  };

  const handleNewCourse = () => {
    setSelection({ type: "course", courseId: "" });
    setDetail(null);
    courseForm.setFieldsValue({ id: "", title: "", description: "", order: courses.length + 1 });
  };

  const handleNewDocument = () => {
    if (!selectedCourse) {
      message.warning("请先选择一个课程");
      return;
    }
    setSelection({ type: "document", courseId: selectedCourse.id, docId: "" });
    setDetail(null);
    docForm.setFieldsValue({
      id: "",
      title: "",
      order: selectedCourse.documents.length + 1,
      markdown: emptyDocumentMarkdown,
    });
  };

  const saveCourse = async (values: CourseFormValue) => {
    setSaving(true);
    try {
      if (selection?.courseId) {
        await updateDocumentCourse(selection.courseId, values);
        message.success("课程已保存");
      } else {
        await createDocumentCourse(values);
        message.success("课程已创建");
      }
      await reload();
      setSelection({ type: "course", courseId: values.id });
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "保存课程失败");
    } finally {
      setSaving(false);
    }
  };

  const saveDocument = async (values: DocumentFormValue) => {
    if (!selection || selection.type !== "document") return;
    setSaving(true);
    try {
      if (selection.docId) {
        const res = await updateDocument(selection.courseId, selection.docId, values);
        setDetail(res.data);
        message.success("文档已保存");
      } else {
        const res = await createDocument(selection.courseId, values);
        setDetail(res.data);
        setSelection({ type: "document", courseId: selection.courseId, docId: values.id });
        message.success("文档已创建");
      }
      await reload();
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "保存文档失败");
    } finally {
      setSaving(false);
    }
  };

  const removeSelected = () => {
    if (!selection) return;
    Modal.confirm({
      title: selection.type === "course" ? "确认删除课程？" : "确认删除文档？",
      content: selection.type === "course" ? "删除课程会同时删除该课程下的所有 Markdown 文档。" : "删除后该 Markdown 文档无法恢复。",
      okText: "确认删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: async () => {
        if (selection.type === "course") {
          await deleteDocumentCourse(selection.courseId);
        } else {
          await deleteDocument(selection.courseId, selection.docId);
        }
        message.success("删除成功");
        setSelection(null);
        setDetail(null);
        await reload();
      },
    });
  };

  return (
    <div className="page-shell">
      <div className="dashboard-toolbar">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>文档管理</Typography.Title>
          <Typography.Text type="secondary">按课程维护 Markdown 课件，学生端会在首页同步展示。</Typography.Text>
        </div>
        <Space>
          <Button icon={<FolderAddOutlined />} onClick={handleNewCourse}>新增课程</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleNewDocument}>新增文档</Button>
        </Space>
      </div>

      <Row gutter={16}>
        <Col xs={24} md={8}>
          <Card className="section-card" title="课程文档">
            {loading ? (
              <Spin />
            ) : courses.length === 0 ? (
              <Empty description="暂无课程文档" />
            ) : (
              <Tree showIcon defaultExpandAll treeData={treeData} onSelect={handleSelect} />
            )}
          </Card>
        </Col>
        <Col xs={24} md={16}>
          <Card
            className="section-card"
            title={selection?.type === "document" ? "编辑文档" : "编辑课程"}
            extra={
              selection ? (
                <Space>
                  {selection.type === "document" && <Button icon={<EyeOutlined />} onClick={() => setPreviewOpen(true)}>预览</Button>}
                  <Button danger icon={<DeleteOutlined />} onClick={removeSelected}>删除</Button>
                </Space>
              ) : null
            }
          >
            {!selection ? (
              <Empty description="请选择课程或文档" />
            ) : selection.type === "course" ? (
              <Form form={courseForm} layout="vertical" onFinish={saveCourse}>
                <Form.Item name="id" label="课程 ID" rules={[{ required: true, message: "请输入课程 ID" }]}>
                  <Input placeholder="backend-go" disabled={Boolean(selection.courseId)} />
                </Form.Item>
                <Form.Item name="title" label="课程标题" rules={[{ required: true, message: "请输入课程标题" }]}>
                  <Input placeholder="服务端训练营" />
                </Form.Item>
                <Form.Item name="description" label="课程描述">
                  <Input.TextArea rows={3} placeholder="课程描述会展示在学生首页课程卡片中" />
                </Form.Item>
                <Form.Item name="order" label="排序" rules={[{ required: true, message: "请输入排序" }]}>
                  <InputNumber min={0} style={{ width: 160 }} />
                </Form.Item>
                <Button type="primary" htmlType="submit" loading={saving} icon={<SaveOutlined />}>保存课程</Button>
              </Form>
            ) : (
              <Form form={docForm} layout="vertical" onFinish={saveDocument}>
                <Form.Item name="id" label="文档 ID" rules={[{ required: true, message: "请输入文档 ID" }]}>
                  <Input placeholder="day01-gin" disabled={Boolean(selection.docId)} />
                </Form.Item>
                <Form.Item name="title" label="文档标题" rules={[{ required: true, message: "请输入文档标题" }]}>
                  <Input placeholder="DAY01 - Gin 基础" />
                </Form.Item>
                <Form.Item name="order" label="排序" rules={[{ required: true, message: "请输入排序" }]}>
                  <InputNumber min={0} style={{ width: 160 }} />
                </Form.Item>
                <Form.Item name="markdown" label="Markdown 正文" rules={[{ required: true, message: "请输入 Markdown 正文" }]}>
                  <Input.TextArea rows={18} style={{ fontFamily: "Menlo, Consolas, monospace" }} />
                </Form.Item>
                <Button type="primary" htmlType="submit" loading={saving} icon={<SaveOutlined />}>保存文档</Button>
              </Form>
            )}
          </Card>
        </Col>
      </Row>

      <Drawer title="Markdown 预览" open={previewOpen} width={760} onClose={() => setPreviewOpen(false)}>
        <div className="markdown-body" style={{ background: "#fff" }}>
          <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
            {previewMarkdown || detail?.markdown || ""}
          </ReactMarkdown>
        </div>
      </Drawer>
    </div>
  );
}
```

- [ ] **Step 2: Add route**

Modify `client/src/App.tsx`:

1. Add import:

```ts
import { DocumentManagePage } from "./pages/DocumentManagePage";
```

2. Add route inside the `AppLayout` route group:

```tsx
<Route path="/documents" element={<DocumentManagePage />} />
```

- [ ] **Step 3: Add menu item**

Modify `client/src/components/AppLayout.tsx`:

1. Add `FolderOpenOutlined` to icon imports.
2. Add menu item after learning notes:

```tsx
{ key: "/documents", label: <Link to="/documents">文档管理</Link>, icon: <FolderOpenOutlined /> },
```

3. Update `getSelectedKey`:

```ts
if (pathname.startsWith("/documents")) return "/documents";
```

- [ ] **Step 4: Type-check frontend**

Run:

```bash
cd client
pnpm build
```

Expected: build succeeds.

- [ ] **Step 5: Commit management UI**

```bash
git add client/src/pages/DocumentManagePage.tsx client/src/App.tsx client/src/components/AppLayout.tsx
git commit -m "feat: add document management page"
```

## Task 5: Student Home And Document Reader

**Files:**
- Create: `client/src/pages/HomePage.tsx`
- Create: `client/src/pages/DocumentReaderPage.tsx`
- Modify: `client/src/App.tsx`
- Modify: `client/src/hooks/useAuth.tsx`
- Modify: `client/src/index.css`

- [ ] **Step 1: Create student home page**

Create `client/src/pages/HomePage.tsx`:

```tsx
import { useEffect, useState } from "react";
import { Button, Empty, Spin, Tag, Typography, message } from "antd";
import { CalendarOutlined, LogoutOutlined, ReadOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import dayjs from "dayjs";
import { fetchPublishedPapers, startAttempt, type PublishedPaper } from "../api/exam";
import { useAuth } from "../hooks/useAuth";
import { useDocumentCourses } from "../hooks/useDocuments";

export function HomePage() {
  const { user, logout } = useAuth();
  const { courses, loading: coursesLoading } = useDocumentCourses();
  const [papers, setPapers] = useState<PublishedPaper[]>([]);
  const [papersLoading, setPapersLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const load = async () => {
      setPapersLoading(true);
      try {
        const res = await fetchPublishedPapers();
        setPapers(res.data || []);
      } catch (err: unknown) {
        message.error(err instanceof Error ? err.message : "读取考试列表失败");
      } finally {
        setPapersLoading(false);
      }
    };
    void load();
  }, []);

  const handleStart = async (paperId: number) => {
    try {
      const res = await startAttempt(paperId);
      navigate(`/exam/${res.data.id}/take`);
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "开始答题失败");
    }
  };

  return (
    <div className="student-home">
      <header className="student-home__top">
        <Typography.Title level={2}>个人中心</Typography.Title>
        <Button
          type="text"
          icon={<LogoutOutlined />}
          onClick={() => {
            logout();
            navigate("/login", { replace: true });
          }}
        >
          退出登录
        </Button>
      </header>

      <section className="student-profile-card">
        <div className="student-profile-card__avatar">{user?.username?.slice(0, 2) || "学员"}</div>
        <div>
          <Typography.Title level={2} style={{ margin: 0 }}>
            <UserOutlined /> {user?.username ?? "未登录"}
          </Typography.Title>
          <div className="student-profile-card__meta">
            <span>ID: {user?.id ?? "-"}</span>
            <span>角色: {user?.role ?? "-"}</span>
            <span>状态: {user?.status ?? "-"}</span>
            <span>班级 ID: {user?.classId ?? "未分配"}</span>
          </div>
        </div>
        <div className="student-profile-card__plane" aria-hidden />
      </section>

      <section className="student-section">
        <h2><ReadOutlined /> 课程列表</h2>
        {coursesLoading ? (
          <div className="student-loading"><Spin /></div>
        ) : courses.length === 0 ? (
          <Empty description="暂无课程资料" />
        ) : (
          <div className="course-card-grid">
            {courses.map((course) => (
              <article className="student-course-card" key={course.id}>
                <h3>📄 {course.title}</h3>
                <div className="student-course-card__divider" />
                <div className="student-course-card__desc">📘 课程描述：{course.description || "暂无描述"}</div>
                <div className="student-course-card__docs">
                  {course.documents.length === 0 ? (
                    <span className="student-muted">暂无文档</span>
                  ) : (
                    course.documents.map((doc, index) => (
                      <button
                        key={doc.id}
                        className="student-doc-link"
                        onClick={() => navigate(`/home/courses/${course.id}/docs/${doc.id}`)}
                      >
                        <span>{index + 1}</span>
                        {doc.title}
                      </button>
                    ))
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="student-section">
        <h2>🎓 我的考试</h2>
        {papersLoading ? (
          <div className="student-loading"><Spin /></div>
        ) : papers.length === 0 ? (
          <Empty description="暂无可参加的考试" />
        ) : (
          <div className="student-exam-grid">
            {papers.map((paper) => {
              const start = dayjs(paper.startTime);
              const end = dayjs(paper.endTime);
              const now = dayjs();
              const isActive = now.isAfter(start) && now.isBefore(end);
              return (
                <article className="student-exam-card" key={paper.paperId}>
                  <h3>🖊 {paper.title}</h3>
                  <div className="student-course-card__divider" />
                  <p>
                    <strong>考试得分</strong>
                    <span className="student-score">总分 {paper.totalScore}</span>
                  </p>
                  <p>
                    <strong>考试时间</strong>
                    <span><CalendarOutlined /> {start.format("YYYY-MM-DD")}</span>
                  </p>
                  <Tag color={isActive ? "green" : "default"}>{isActive ? "可进入" : "不在考试时间"}</Tag>
                  <Button type="primary" block disabled={!isActive} onClick={() => handleStart(paper.paperId)}>
                    {isActive ? "进入答题" : `${start.format("MM-DD HH:mm")} 开始`}
                  </Button>
                </article>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
```

- [ ] **Step 2: Create document reader page**

Create `client/src/pages/DocumentReaderPage.tsx`:

```tsx
import { useMemo } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import "github-markdown-css/github-markdown-light.css";
import "highlight.js/styles/github.css";
import { Button, Empty, Spin } from "antd";
import { ArrowLeftOutlined, MenuFoldOutlined } from "@ant-design/icons";
import { useNavigate, useParams } from "react-router-dom";
import { useDocumentCourses, useDocumentDetail } from "../hooks/useDocuments";

interface TocItem {
  id: string;
  text: string;
  level: number;
}

function slugifyHeading(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-|-$/g, "");
}

function extractToc(markdown: string): TocItem[] {
  return markdown
    .split("\n")
    .map((line) => /^(#{1,3})\s+(.+)$/.exec(line.trim()))
    .filter((match): match is RegExpExecArray => Boolean(match))
    .map((match) => ({
      id: slugifyHeading(match[2]),
      text: match[2],
      level: match[1].length,
    }));
}

export function DocumentReaderPage() {
  const { courseId, docId } = useParams();
  const navigate = useNavigate();
  const { courses, loading: coursesLoading } = useDocumentCourses();
  const { detail, loading: detailLoading } = useDocumentDetail(courseId, docId);

  const course = courses.find((item) => item.id === courseId);
  const toc = useMemo(() => extractToc(detail?.markdown || ""), [detail?.markdown]);

  return (
    <div className="document-reader">
      <header className="document-reader__top">
        <div>
          <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate("/home")} />
          <strong>{detail?.title || "课程文档"}</strong>
          <MenuFoldOutlined />
        </div>
        <Button type="text" onClick={() => navigate("/home")}>切换课件</Button>
      </header>

      <main className="document-reader__layout">
        <aside className="document-reader__sidebar">
          {coursesLoading ? (
            <Spin />
          ) : !course ? (
            <Empty description="课程不存在" />
          ) : (
            <>
              <div className="document-reader__active">{detail?.title || "正在加载"}</div>
              <h3>{course.title}</h3>
              <nav>
                {course.documents.map((doc) => (
                  <button
                    key={doc.id}
                    className={doc.id === docId ? "is-current" : ""}
                    onClick={() => navigate(`/home/courses/${course.id}/docs/${doc.id}`)}
                  >
                    {doc.title}
                  </button>
                ))}
              </nav>
              {toc.length > 0 && (
                <div className="document-reader__toc">
                  {toc.map((item) => (
                    <a key={`${item.level}-${item.id}`} style={{ paddingLeft: (item.level - 1) * 14 }} href={`#${item.id}`}>
                      {item.text}
                    </a>
                  ))}
                </div>
              )}
            </>
          )}
        </aside>
        <article className="document-reader__content markdown-body">
          {detailLoading ? (
            <Spin />
          ) : detail ? (
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              rehypePlugins={[rehypeHighlight]}
              components={{
                h1: ({ children }) => <h1 id={slugifyHeading(String(children))}>{children}</h1>,
                h2: ({ children }) => <h2 id={slugifyHeading(String(children))}>{children}</h2>,
                h3: ({ children }) => <h3 id={slugifyHeading(String(children))}>{children}</h3>,
              }}
            >
              {detail.markdown}
            </ReactMarkdown>
          ) : (
            <Empty description="文档不存在或读取失败" />
          )}
        </article>
      </main>
    </div>
  );
}
```

- [ ] **Step 3: Add student routes and redirect**

Modify `client/src/App.tsx`:

1. Add imports:

```ts
import { HomePage } from "./pages/HomePage";
import { DocumentReaderPage } from "./pages/DocumentReaderPage";
```

2. Replace the `/exam` route with:

```tsx
<Route
  path="/home"
  element={
    <RequireAuth roles={["admin", "student"]}>
      <HomePage />
    </RequireAuth>
  }
/>
<Route
  path="/home/courses/:courseId/docs/:docId"
  element={
    <RequireAuth roles={["admin", "student"]}>
      <DocumentReaderPage />
    </RequireAuth>
  }
/>
<Route
  path="/exam"
  element={
    <RequireAuth roles={["admin", "student"]}>
      <Navigate to="/home" replace />
    </RequireAuth>
  }
/>
```

Keep `/exam/:id/take` and `/exam/:id/result` unchanged.

- [ ] **Step 4: Update default student route**

Modify `client/src/hooks/useAuth.tsx`:

```ts
if (role === "student") {
  return "/home";
}
```

- [ ] **Step 5: Add CSS**

Append to `client/src/index.css`:

```css
.student-home {
  min-height: 100vh;
  background: #f4f4f5;
  color: #111827;
  padding: 28px 24px 64px;
}

.student-home__top,
.student-section,
.student-profile-card {
  width: min(1180px, 100%);
  margin: 0 auto;
}

.student-home__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
}

.student-profile-card {
  position: relative;
  overflow: hidden;
  min-height: 132px;
  padding: 28px 42px;
  display: flex;
  align-items: center;
  gap: 24px;
  background: #fff;
  border: 1px solid #dfe4ec;
  border-radius: 8px;
  box-shadow: 0 10px 18px rgba(15, 23, 42, 0.14);
}

.student-profile-card__avatar {
  width: 86px;
  height: 86px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eef2f7;
  color: #111827;
  font-weight: 800;
}

.student-profile-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 18px;
  margin-top: 12px;
  color: #6b7280;
  font-weight: 600;
}

.student-profile-card__plane {
  position: absolute;
  right: 38px;
  top: 28px;
  width: 130px;
  height: 82px;
  background: linear-gradient(135deg, #d7efff 0%, #7fc6f4 70%);
  clip-path: polygon(0 56%, 100% 0, 74% 72%, 46% 88%);
  opacity: 0.82;
}

.student-section {
  margin-top: 34px;
}

.student-section h2 {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 900;
}

.course-card-grid,
.student-exam-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
}

.student-course-card,
.student-exam-card {
  min-height: 294px;
  background: #fff;
  border: 1px solid #dfe4ec;
  border-radius: 8px;
  box-shadow: 0 7px 14px rgba(15, 23, 42, 0.12);
  padding: 24px 28px;
}

.student-course-card h3,
.student-exam-card h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 900;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.student-course-card__divider {
  height: 1px;
  background: #d5dbe5;
  margin: 16px 0 12px;
}

.student-course-card__desc {
  background: #f0f2f5;
  border-radius: 6px;
  color: #6b7280;
  padding: 12px;
  min-height: 54px;
  font-weight: 700;
}

.student-course-card__docs {
  margin-top: 14px;
  min-height: 142px;
  background: #f8fafc;
  border-radius: 6px;
  padding: 10px 12px;
}

.student-doc-link {
  width: 100%;
  border: 0;
  background: transparent;
  padding: 4px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
  font-weight: 700;
  color: #111827;
  cursor: pointer;
}

.student-doc-link span {
  width: 20px;
  height: 20px;
  border-radius: 5px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #e5e7eb;
  color: #4b5563;
  flex: 0 0 auto;
}

.student-doc-link:hover {
  color: #2563eb;
}

.student-exam-card {
  min-height: 236px;
}

.student-exam-card p {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: #4b5563;
}

.student-score {
  color: #35a852;
  font-size: 20px;
  font-weight: 900;
}

.student-exam-card .ant-btn {
  margin-top: 12px;
  height: 42px;
  border-radius: 6px;
}

.student-loading {
  padding: 48px;
  text-align: center;
}

.student-muted {
  color: #9ca3af;
}

.document-reader {
  min-height: 100vh;
  background: #f3f6fb;
}

.document-reader__top {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 22px;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
}

.document-reader__top > div {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
}

.document-reader__layout {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 18px;
  padding: 14px;
}

.document-reader__sidebar,
.document-reader__content {
  background: #fff;
  border-radius: 8px;
}

.document-reader__sidebar {
  min-height: calc(100vh - 92px);
  padding: 14px;
  overflow: auto;
}

.document-reader__active {
  background: #dceeff;
  color: #3b82f6;
  border-radius: 5px;
  padding: 10px 12px;
  font-weight: 900;
}

.document-reader__sidebar nav button,
.document-reader__toc a {
  width: 100%;
  display: block;
  border: 0;
  border-left: 2px solid #e5e7eb;
  background: transparent;
  color: #4b5563;
  padding: 8px 10px;
  text-align: left;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
}

.document-reader__sidebar nav button.is-current {
  color: #2563eb;
  border-left-color: #2563eb;
}

.document-reader__toc {
  margin-top: 18px;
}

.document-reader__content {
  min-height: calc(100vh - 92px);
  padding: 34px;
  overflow: auto;
}

@media (max-width: 980px) {
  .course-card-grid,
  .student-exam-grid {
    grid-template-columns: 1fr;
  }

  .document-reader__layout {
    grid-template-columns: 1fr;
  }

  .student-profile-card__plane {
    display: none;
  }
}
```

- [ ] **Step 6: Build frontend**

Run:

```bash
cd client
pnpm build
```

Expected: build succeeds.

- [ ] **Step 7: Commit student pages**

```bash
git add client/src/pages/HomePage.tsx client/src/pages/DocumentReaderPage.tsx client/src/App.tsx client/src/hooks/useAuth.tsx client/src/index.css
git commit -m "feat: add student home and document reader"
```

## Task 6: Seed Course Docs And Project Documentation

**Files:**
- Create: `course-docs/backend-go/course.json`
- Create: `course-docs/backend-go/day01-gin.md`
- Modify: `AGENTS.md`
- Modify: `docs/backend-architecture.md`

- [ ] **Step 1: Add seed course metadata**

Create `course-docs/backend-go/course.json`:

```json
{
  "id": "backend-go",
  "title": "服务端训练营",
  "description": "Go 后端基础课",
  "order": 1,
  "documents": [
    {
      "id": "day01-gin",
      "title": "DAY01 - Gin 基础",
      "order": 1
    }
  ]
}
```

- [ ] **Step 2: Add seed Markdown**

Create `course-docs/backend-go/day01-gin.md`:

```markdown
# DAY01 - Gin 基础

## Gin 是什么

Gin 是 Go 生态中常用的 Web 框架，适合快速编写 REST API。

## 最小示例

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.Run(":8080")
}
```

## 知识点

- `gin.Default()` 会创建带日志和恢复中间件的路由引擎。
- `GET` 用来注册 GET 请求路由。
- `c.JSON` 用来返回 JSON 响应。
```

- [ ] **Step 3: Update AGENTS.md**

Modify `AGENTS.md`:

1. In frontend pages list, replace `/exam` item with `/home` and add:

```markdown
- `/documents`
  文档管理
- `/home/courses/:courseId/docs/:docId`
  课程文档阅读页
```

2. In backend route groups, add:

```markdown
- 已登录接口
  - `/api/documents/courses`
  - `/api/documents/courses/:courseId`
  - `/api/documents/courses/:courseId/docs/:docId`
- 教师/管理员接口
  - `/api/documents/courses*`
```

3. Add a new core module section:

```markdown
### 4.6 课程文档

已完成：

- 文件型课程文档存储
- 课程与 Markdown 文档管理
- 学生首页课程列表
- Markdown 阅读页

当前约束：

- 文档根目录为 `course-docs/`
- 课程按目录分组，课程元数据写入 `course.json`
- 管理员和教师均可管理全部课程文档
- 登录用户均可查看全部课程文档
- 当前不支持图片或附件上传
```

4. In data model overview, add that course docs are file-backed and not in SQLite.

5. In regression baseline, add document management and `/home` smoke scenarios.

- [ ] **Step 4: Update backend architecture doc**

Modify `docs/backend-architecture.md` by adding a “课程文档模块” section:

```markdown
## 课程文档模块

课程文档不进入 SQLite 主业务表，而是存储在项目根目录的 `course-docs/` 下。

- `server/repositories/document_repository.go` 负责安全读写 `course-docs/`
- `server/services/document_service.go` 负责 slug、标题、冲突等业务校验
- `server/handlers/document_handler.go` 负责 HTTP 参数和响应

读接口要求登录即可访问；写接口只允许 `admin` 和 `teacher`。
```

- [ ] **Step 5: Run final build checks**

Run:

```bash
cd server
go build ./...
```

Run:

```bash
cd client
pnpm build
```

Expected: both pass.

- [ ] **Step 6: Commit docs and seed content**

```bash
git add course-docs AGENTS.md docs/backend-architecture.md
git commit -m "docs: document course materials module"
```

## Task 7: Browser Smoke Verification

**Files:**
- No source files expected. This task verifies the integrated app.

- [ ] **Step 1: Start backend server**

Run:

```bash
cd server
go run .
```

Expected: server starts and prints the listening address.

- [ ] **Step 2: Start frontend dev server**

Run in another shell:

```bash
cd client
pnpm dev -- --host 127.0.0.1
```

Expected: Vite prints a localhost URL.

- [ ] **Step 3: Verify student home**

Open the Vite URL in the browser, login with the seeded student account:

```text
username: student01
password: student123
```

Verify:

- Login redirects to `/home`.
- Personal information card is visible.
- Course list shows “服务端训练营”.
- Exam list section is visible.
- Visiting `/exam` redirects to `/home`.

- [ ] **Step 4: Verify document reader**

Click “DAY01 - Gin 基础”.

Verify:

- URL is `/home/courses/backend-go/docs/day01-gin`.
- Top title, left sidebar, and Markdown body are visible.
- Code block renders with highlighting.
- Back/switch control returns to `/home`.

- [ ] **Step 5: Verify teacher document management**

Logout, login with:

```text
username: teacher01
password: teacher123
```

Verify:

- Sidebar contains “文档管理”.
- `/documents` opens the management page.
- Selecting “服务端训练营” loads course form.
- Selecting “DAY01 - Gin 基础” loads Markdown editor.
- Editing and saving the description or markdown shows success.

- [ ] **Step 6: Stop dev servers**

Stop both terminal sessions with `Ctrl+C`.

Expected: no source changes from smoke verification except intentional content edits from the management page. If smoke edits changed `course-docs/`, inspect the diff and keep only changes that are useful as seed content.

## Final Verification

Run:

```bash
cd server
go test ./...
go build ./...
```

Run:

```bash
cd client
pnpm test -- --run
pnpm build
```

Expected:

- Go tests pass.
- Go build passes.
- Frontend tests pass.
- Frontend build passes.
