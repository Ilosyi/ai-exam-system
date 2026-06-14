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
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": courses, "total": len(courses)})
}

func (h *DocumentHandler) GetCourse(c *gin.Context) {
	course, err := h.service.GetCourse(c.Request.Context(), c.Param("courseId"))
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": course})
}

func (h *DocumentHandler) CreateCourse(c *gin.Context) {
	var payload coursePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	course, err := h.service.CreateCourse(c.Request.Context(), services.CourseInput{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Order:       payload.Order,
	})
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "data": course})
}

func (h *DocumentHandler) UpdateCourse(c *gin.Context) {
	var payload coursePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	course, err := h.service.UpdateCourse(c.Request.Context(), c.Param("courseId"), services.CourseInput{
		ID:          payload.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Order:       payload.Order,
	})
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": course})
}

func (h *DocumentHandler) DeleteCourse(c *gin.Context) {
	if err := h.service.DeleteCourse(c.Request.Context(), c.Param("courseId")); err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	doc, err := h.service.GetDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId"))
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": doc})
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var payload documentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	doc, err := h.service.CreateDocument(c.Request.Context(), c.Param("courseId"), services.DocumentInput{
		ID:       payload.ID,
		Title:    payload.Title,
		Order:    payload.Order,
		Markdown: payload.Markdown,
	})
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "data": doc})
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	var payload documentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	doc, err := h.service.UpdateDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId"), services.DocumentInput{
		ID:       payload.ID,
		Title:    payload.Title,
		Order:    payload.Order,
		Markdown: payload.Markdown,
	})
	if err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": doc})
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	if err := h.service.DeleteDocument(c.Request.Context(), c.Param("courseId"), c.Param("docId")); err != nil {
		respondDocumentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func respondDocumentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": "文档不存在", "error": err.Error()})
	case errors.Is(err, repositories.ErrInvalidDocumentSlug),
		errors.Is(err, services.ErrDocumentTitleRequired),
		errors.Is(err, services.ErrDocumentSlugConflict):
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
	}
}
