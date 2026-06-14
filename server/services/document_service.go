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
	course, err := courseFromInput(input, nil)
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
	return s.repo.GetCourse(ctx, course.ID)
}

func (s *DocumentService) UpdateCourse(ctx context.Context, courseID string, input CourseInput) (repositories.CourseDocument, error) {
	current, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return repositories.CourseDocument{}, err
	}

	course, err := courseFromInput(CourseInput{
		ID:          courseID,
		Title:       input.Title,
		Description: input.Description,
		Order:       input.Order,
	}, current.Documents)
	if err != nil {
		return repositories.CourseDocument{}, err
	}

	if err := s.repo.SaveCourse(ctx, course); err != nil {
		return repositories.CourseDocument{}, err
	}
	return s.repo.GetCourse(ctx, courseID)
}

func (s *DocumentService) DeleteCourse(ctx context.Context, courseID string) error {
	if _, err := s.repo.GetCourse(ctx, courseID); err != nil {
		return err
	}
	return s.repo.DeleteCourse(ctx, courseID)
}

func (s *DocumentService) GetDocument(ctx context.Context, courseID, docID string) (DocumentDetail, error) {
	doc, markdown, err := s.repo.GetDocument(ctx, courseID, docID)
	if err != nil {
		return DocumentDetail{}, err
	}
	return DocumentDetail{DocumentMeta: doc, Markdown: markdown}, nil
}

func (s *DocumentService) CreateDocument(ctx context.Context, courseID string, input DocumentInput) (DocumentDetail, error) {
	doc, markdown, err := documentFromInput(input)
	if err != nil {
		return DocumentDetail{}, err
	}

	if _, err := s.GetDocument(ctx, courseID, doc.ID); err == nil {
		return DocumentDetail{}, ErrDocumentSlugConflict
	} else if !errors.Is(err, repositories.ErrDocumentNotFound) {
		return DocumentDetail{}, err
	}

	if err := s.repo.SaveDocument(ctx, courseID, doc, markdown); err != nil {
		return DocumentDetail{}, err
	}
	return s.GetDocument(ctx, courseID, doc.ID)
}

func (s *DocumentService) UpdateDocument(ctx context.Context, courseID, docID string, input DocumentInput) (DocumentDetail, error) {
	if _, err := s.GetDocument(ctx, courseID, docID); err != nil {
		return DocumentDetail{}, err
	}

	doc, markdown, err := documentFromInput(DocumentInput{
		ID:       docID,
		Title:    input.Title,
		Order:    input.Order,
		Markdown: input.Markdown,
	})
	if err != nil {
		return DocumentDetail{}, err
	}

	if err := s.repo.SaveDocument(ctx, courseID, doc, markdown); err != nil {
		return DocumentDetail{}, err
	}
	return s.GetDocument(ctx, courseID, docID)
}

func (s *DocumentService) DeleteDocument(ctx context.Context, courseID, docID string) error {
	return s.repo.DeleteDocument(ctx, courseID, docID)
}

func courseFromInput(input CourseInput, documents []repositories.DocumentMeta) (repositories.CourseDocument, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return repositories.CourseDocument{}, ErrDocumentTitleRequired
	}

	return repositories.CourseDocument{
		ID:          strings.TrimSpace(input.ID),
		Title:       title,
		Description: input.Description,
		Order:       input.Order,
		Documents:   documents,
	}, nil
}

func documentFromInput(input DocumentInput) (repositories.DocumentMeta, string, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return repositories.DocumentMeta{}, "", ErrDocumentTitleRequired
	}

	return repositories.DocumentMeta{
		ID:    strings.TrimSpace(input.ID),
		Title: title,
		Order: input.Order,
	}, input.Markdown, nil
}
