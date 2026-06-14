// ============================================================================
// handlers/class_handler.go - 班级管理接口处理器
// ============================================================================
//
// 本文件实现了班级管理相关的 HTTP 接口，包括：
// - Create:              创建班级
// - List:                班级列表
// - Update:              更新班级
// - Delete:              删除班级
// - ListStudents:        班级学生列表
// - BatchEditStudents:   批量加入/移出学生
// - GetStudentExams:     学生考试记录
//
// 权限控制：
// - 教师只能管理自己创建的班级
// - 管理员可以管理所有班级
//
// 学习要点：
// - 资源归属的权限校验
// - 批量操作的实现
// - 复杂查询的使用
// ============================================================================

package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
)

// ClassHandler 处理班级相关的 HTTP 请求。
type ClassHandler struct {
	classRepo *repositories.ClassRepository
}

// classPayload 创建/更新班级的请求体
type classPayload struct {
	Name      string `json:"name" binding:"required,min=2,max=64"` // 班级名称（2-64 字符）
	TeacherID *uint  `json:"teacherId"`                             // 教师 ID（仅管理员可指定）
}

// NewClassHandler 创建一个新的 ClassHandler 实例。
func NewClassHandler(classRepo *repositories.ClassRepository) *ClassHandler {
	return &ClassHandler{classRepo: classRepo}
}

// Create 处理创建班级请求。
//
// POST /api/classes
func (h *ClassHandler) Create(c *gin.Context) {
	var req classPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 默认使用当前用户作为教师
	teacherID := middleware.GetCurrentUserID(c)
	// 管理员可以指定其他教师
	if req.TeacherID != nil && middleware.GetCurrentUserRole(c) == "admin" {
		teacherID = *req.TeacherID
	}

	item := models.Class{
		Name:      strings.TrimSpace(req.Name),
		TeacherID: teacherID,
	}
	if item.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "班级名称不能为空"})
		return
	}

	if err := h.classRepo.Create(context.Background(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "data": item})
}

// List 处理班级列表请求。
//
// GET /api/classes?page=1&pageSize=10&keyword=xxx&teacherId=1
func (h *ClassHandler) List(c *gin.Context) {
	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)

	filters := repositories.ClassFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     page,
		PageSize: pageSize,
	}

	// 教师只能看到自己的班级
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" {
		filters.TeacherID = &currentID
	} else if teacherVal := strings.TrimSpace(c.Query("teacherId")); teacherVal != "" {
		id, err := strconv.Atoi(teacherVal)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "teacherId 参数无效"})
			return
		}
		tid := uint(id)
		filters.TeacherID = &tid
	}

	items, total, err := h.classRepo.List(context.Background(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// Update 处理更新班级请求。
//
// PUT /api/classes/:id
func (h *ClassHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的班级ID"})
		return
	}

	item, err := h.classRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "班级不存在"})
		return
	}

	// 权限校验：教师只能修改自己的班级
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" && item.TeacherID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权限修改该班级"})
		return
	}

	var req classPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	name := strings.TrimSpace(req.Name)
	if name != "" {
		updates["name"] = name
	}
	// 只有管理员可以修改教师归属
	if req.TeacherID != nil && role == "admin" {
		updates["teacher_id"] = *req.TeacherID
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无更新内容"})
		return
	}

	if err := h.classRepo.Update(context.Background(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "error": err.Error()})
		return
	}

	updated, _ := h.classRepo.FindByID(context.Background(), uint(id))
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updated})
}

// Delete 处理删除班级请求。
//
// DELETE /api/classes/:id
func (h *ClassHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的班级ID"})
		return
	}

	item, err := h.classRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "班级不存在"})
		return
	}

	// 权限校验
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" && item.TeacherID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权限删除该班级"})
		return
	}

	if err := h.classRepo.Delete(context.Background(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListStudents 处理班级学生列表请求。
//
// GET /api/classes/:id/students?page=1&pageSize=20&keyword=xxx&scope=class
//
// scope 参数：
// - class: 只显示班级内的学生
// - all:   显示所有学生（标注是否在班级内）
func (h *ClassHandler) ListStudents(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的班级ID"})
		return
	}

	item, err := h.classRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "班级不存在"})
		return
	}

	// 权限校验
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" && item.TeacherID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权查看该班级学生"})
		return
	}

	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 20)
	filters := repositories.ClassStudentFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		Scope:    strings.TrimSpace(c.DefaultQuery("scope", "class")),
		Page:     page,
		PageSize: pageSize,
	}
	if filters.Scope != "class" && filters.Scope != "all" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "scope 参数无效"})
		return
	}

	students, total, err := h.classRepo.ListStudentsByClass(context.Background(), uint(id), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	type studentRow struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Status   string `json:"status"`
		ClassID  *uint  `json:"classId"`
		InClass  bool   `json:"inClass"`
	}
	rows := make([]studentRow, 0, len(students))
	for _, s := range students {
		rows = append(rows, studentRow{
			ID:       s.ID,
			Username: s.Username,
			Status:   s.Status,
			ClassID:  s.ClassID,
			InClass:  s.InClass,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total})
}

// batchEditRequest 批量编辑学生请求体
type batchEditRequest struct {
	Action     string `json:"action" binding:"required,oneof=add remove"` // 操作：add=加入，remove=移出
	StudentIDs []uint `json:"studentIds" binding:"required,min=1"`         // 学生 ID 列表
}

// BatchEditStudents 处理批量加入/移出学生请求。
//
// POST /api/classes/:id/students/batch-edit
// Body: {"action": "add", "studentIds": [1, 2, 3]}
func (h *ClassHandler) BatchEditStudents(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的班级ID"})
		return
	}

	item, err := h.classRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "班级不存在"})
		return
	}

	// 权限校验
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" && item.TeacherID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权管理该班级成员"})
		return
	}

	var req batchEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	ctx := context.Background()
	if req.Action == "add" {
		if err := h.classRepo.BatchAddStudentsToClass(ctx, uint(id), req.StudentIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "批量加入失败", "error": err.Error()})
			return
		}
	} else {
		if err := h.classRepo.BatchRemoveStudentsFromClass(ctx, uint(id), req.StudentIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "批量移出失败", "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// GetStudentExams 处理学生考试记录请求。
//
// GET /api/classes/:id/students/:studentId/exams?page=1&pageSize=10
func (h *ClassHandler) GetStudentExams(c *gin.Context) {
	classID, err := strconv.Atoi(c.Param("id"))
	if err != nil || classID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的班级ID"})
		return
	}
	studentID, err := strconv.Atoi(c.Param("studentId"))
	if err != nil || studentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的学生ID"})
		return
	}

	item, err := h.classRepo.FindByID(context.Background(), uint(classID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "班级不存在"})
		return
	}

	// 权限校验
	role := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)
	if role == "teacher" && item.TeacherID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权查看该班级学生考试记录"})
		return
	}

	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)

	records, total, err := h.classRepo.ListStudentExamRecords(context.Background(), uint(classID), uint(studentID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records, "total": total})
}
