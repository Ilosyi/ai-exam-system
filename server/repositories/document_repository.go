package repositories

import (
	"context"
	"encoding/json"
	"errors"
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

var documentSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

const courseMetadataFile = "course.json"

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

func NewDocumentRepository(root string) *DocumentRepository {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = filepath.Clean(root)
	}

	return &DocumentRepository{root: absRoot}
}

func (r *DocumentRepository) ListCourses(ctx context.Context) ([]CourseDocument, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := r.ensureRootSafe(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(r.root, 0o755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(r.root)
	if err != nil {
		return nil, err
	}

	courses := make([]CourseDocument, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, ErrInvalidDocumentSlug
		}

		coursePath, err := r.safePath(entry.Name(), courseMetadataFile)
		if err != nil {
			return nil, err
		}

		course, err := readCourseFile(coursePath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		sortDocuments(course.Documents)
		courses = append(courses, course)
	}

	sortCourses(courses)
	return courses, nil
}

func (r *DocumentRepository) GetCourse(ctx context.Context, courseID string) (CourseDocument, error) {
	if err := ctx.Err(); err != nil {
		return CourseDocument{}, err
	}
	if err := validateSlug(courseID); err != nil {
		return CourseDocument{}, err
	}

	coursePath, err := r.safePath(courseID, courseMetadataFile)
	if err != nil {
		return CourseDocument{}, err
	}

	course, err := readCourseFile(coursePath)
	if errors.Is(err, os.ErrNotExist) {
		return CourseDocument{}, ErrDocumentNotFound
	}
	if err != nil {
		return CourseDocument{}, err
	}

	sortDocuments(course.Documents)
	return course, nil
}

func (r *DocumentRepository) SaveCourse(ctx context.Context, course CourseDocument) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSlug(course.ID); err != nil {
		return err
	}

	courseDir, err := r.safePath(course.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(courseDir, 0o755); err != nil {
		return err
	}
	if err := r.rejectSymlinkPath(course.ID); err != nil {
		return err
	}

	// 只更新课程基础信息时，保留已有文档列表，避免误删目录中的文档索引。
	if course.Documents == nil {
		current, err := r.GetCourse(ctx, course.ID)
		if err == nil {
			course.Documents = current.Documents
		} else if !errors.Is(err, ErrDocumentNotFound) {
			return err
		}
	}

	return r.writeCourse(ctx, course)
}

func (r *DocumentRepository) DeleteCourse(ctx context.Context, courseID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSlug(courseID); err != nil {
		return err
	}

	courseDir, err := r.safePath(courseID)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(courseDir); err != nil {
		return err
	}
	return nil
}

func (r *DocumentRepository) GetDocument(ctx context.Context, courseID, docID string) (DocumentMeta, string, error) {
	if err := ctx.Err(); err != nil {
		return DocumentMeta{}, "", err
	}
	if err := validateSlug(courseID); err != nil {
		return DocumentMeta{}, "", err
	}
	if err := validateSlug(docID); err != nil {
		return DocumentMeta{}, "", err
	}

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
	if errors.Is(err, os.ErrNotExist) {
		return DocumentMeta{}, "", ErrDocumentNotFound
	}
	if err != nil {
		return DocumentMeta{}, "", err
	}

	return doc, string(content), nil
}

func (r *DocumentRepository) SaveDocument(ctx context.Context, courseID string, doc DocumentMeta, markdown string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSlug(courseID); err != nil {
		return err
	}
	if err := validateSlug(doc.ID); err != nil {
		return err
	}

	course, err := r.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	docPath, err := r.documentPath(courseID, doc.ID)
	if err != nil {
		return err
	}

	originalCourse := cloneCourseDocument(course)
	_, existed := findDocument(course.Documents, doc.ID)
	upsertDocument(&course, doc)
	if err := writeFileAtomically(docPath, []byte(markdown), 0o644); err != nil {
		return err
	}
	if err := r.writeCourse(ctx, course); err != nil {
		if !existed {
			_ = os.Remove(docPath)
		}
		if existed {
			_ = r.writeCourse(ctx, originalCourse)
		}
		return err
	}
	return nil
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, courseID, docID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSlug(courseID); err != nil {
		return err
	}
	if err := validateSlug(docID); err != nil {
		return err
	}

	course, err := r.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	originalCourse := cloneCourseDocument(course)
	documents := course.Documents[:0]
	found := false
	for _, doc := range course.Documents {
		if doc.ID == docID {
			found = true
			continue
		}
		documents = append(documents, doc)
	}
	if !found {
		return ErrDocumentNotFound
	}
	docPath, err := r.documentPath(courseID, docID)
	if err != nil {
		return err
	}

	tmpPath, err := renameFileAside(docPath)
	if err != nil {
		return err
	}
	if tmpPath != "" {
		defer os.Remove(tmpPath)
	}

	course.Documents = documents
	if err := r.writeCourse(ctx, course); err != nil {
		if tmpPath != "" {
			_ = os.Rename(tmpPath, docPath)
		}
		_ = r.writeCourse(ctx, originalCourse)
		return err
	}

	return nil
}

func (r *DocumentRepository) writeCourse(ctx context.Context, course CourseDocument) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSlug(course.ID); err != nil {
		return err
	}
	for _, doc := range course.Documents {
		if err := validateSlug(doc.ID); err != nil {
			return err
		}
	}
	sortDocuments(course.Documents)

	courseDir, err := r.safePath(course.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(courseDir, 0o755); err != nil {
		return err
	}
	if err := r.rejectSymlinkPath(course.ID); err != nil {
		return err
	}

	coursePath, err := r.safePath(course.ID, courseMetadataFile)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(course, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return writeFileAtomically(coursePath, data, 0o644)
}

func (r *DocumentRepository) documentPath(courseID, docID string) (string, error) {
	return r.safePath(courseID, docID+".md")
}

func (r *DocumentRepository) safePath(parts ...string) (string, error) {
	if err := r.ensureRootSafe(); err != nil {
		return "", err
	}
	targetParts := append([]string{r.root}, parts...)
	target := filepath.Join(targetParts...)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(r.root, absTarget)
	if err != nil {
		return "", err
	}
	if rel == ".." || rel == "."+string(filepath.Separator)+".." || startsWithParent(rel) || filepath.IsAbs(rel) {
		return "", ErrInvalidDocumentSlug
	}
	if err := r.rejectSymlinkPath(parts...); err != nil {
		return "", err
	}

	return absTarget, nil
}

func (r *DocumentRepository) ensureRootSafe() error {
	info, err := os.Lstat(r.root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrInvalidDocumentSlug
	}
	if !info.IsDir() {
		return ErrInvalidDocumentSlug
	}
	return nil
}

func (r *DocumentRepository) rejectSymlinkPath(parts ...string) error {
	current := r.root
	for _, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrInvalidDocumentSlug
		}
	}
	return nil
}

func validateSlug(slug string) error {
	if !documentSlugPattern.MatchString(slug) {
		return ErrInvalidDocumentSlug
	}
	return nil
}

func readCourseFile(path string) (CourseDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CourseDocument{}, err
	}

	var course CourseDocument
	if err := json.Unmarshal(data, &course); err != nil {
		return CourseDocument{}, err
	}
	return course, nil
}

func cloneCourseDocument(course CourseDocument) CourseDocument {
	course.Documents = append([]DocumentMeta(nil), course.Documents...)
	return course
}

func writeFileAtomically(path string, data []byte, perm os.FileMode) error {
	tmpPath, err := writeTempFileForAtomicReplace(path, data, perm)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	return os.Rename(tmpPath, path)
}

func renameFileAside(path string) (string, error) {
	tmpPath, err := tempPathFor(path)
	if err != nil {
		return "", err
	}
	if err := os.Rename(path, tmpPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return tmpPath, nil
}

func tempPathFor(path string) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return "", err
	}
	tmpPath := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := os.Remove(tmpPath); err != nil {
		return "", err
	}
	return tmpPath, nil
}

func writeTempFileForAtomicReplace(path string, data []byte, perm os.FileMode) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	file, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return "", err
	}
	tmpPath := file.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Chmod(perm); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}

	cleanup = false
	return tmpPath, nil
}

func findDocument(documents []DocumentMeta, docID string) (DocumentMeta, bool) {
	for _, doc := range documents {
		if doc.ID == docID {
			return doc, true
		}
	}
	return DocumentMeta{}, false
}

func upsertDocument(course *CourseDocument, doc DocumentMeta) {
	for i := range course.Documents {
		if course.Documents[i].ID == doc.ID {
			course.Documents[i] = doc
			sortDocuments(course.Documents)
			return
		}
	}

	course.Documents = append(course.Documents, doc)
	sortDocuments(course.Documents)
}

func sortCourses(courses []CourseDocument) {
	sort.SliceStable(courses, func(i, j int) bool {
		if courses[i].Order != courses[j].Order {
			return courses[i].Order < courses[j].Order
		}
		return courses[i].Title < courses[j].Title
	})
}

func sortDocuments(documents []DocumentMeta) {
	sort.SliceStable(documents, func(i, j int) bool {
		if documents[i].Order != documents[j].Order {
			return documents[i].Order < documents[j].Order
		}
		return documents[i].Title < documents[j].Title
	})
}

func startsWithParent(path string) bool {
	return strings.HasPrefix(path, ".."+string(filepath.Separator))
}
