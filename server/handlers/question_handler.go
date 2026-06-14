// ============================================================================
// handlers/question_handler.go - 题库管理接口处理器
// ============================================================================
//
// 本文件实现了题库管理相关的 HTTP 接口，包括：
// - List:       题目列表（支持分页、筛选）
// - Create:     创建题目
// - Update:     更新题目
// - Delete:     删除单个题目
// - DeleteMany: 批量删除题目
//
// 题目类型：
// - single:   单选题（4 个选项，1 个正确答案）
// - multiple: 多选题（4 个选项，多个正确答案）
// - coding:   编程题（无选项，代码评测）
//
// 学习要点：
// - JSON 序列化/反序列化
// - 参数校验（binding 标签）
// - 数据库模型与响应结构的转换
// ============================================================================

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
)

// QuestionHandler 处理题库相关的 HTTP 请求。
//
// 依赖：
// - repo: 题目数据访问层
// - userRepo: 用户数据访问层（用于查询创建人名称）
type QuestionHandler struct {
	repo     *repositories.QuestionRepository
	userRepo *repositories.UserRepository
}

// NewQuestionHandler 创建一个新的 QuestionHandler 实例。
func NewQuestionHandler(repo *repositories.QuestionRepository, userRepo *repositories.UserRepository) *QuestionHandler {
	return &QuestionHandler{repo: repo, userRepo: userRepo}
}

// questionPayload 创建/更新题目的请求体。
//
// binding 标签说明：
// - required: 字段必填
// - oneof=x y z: 字段值必须是给定值之一
type questionPayload struct {
	Type     string   `json:"type" binding:"required,oneof=single multiple coding"` // 题型
	Language string   `json:"language" binding:"required,oneof=go cpp java javascript python"` // 语言
	Title    string   `json:"title" binding:"required"`    // 标题
	Content  string   `json:"content" binding:"required"`  // 内容
	Options  []string `json:"options"`                      // 选项（单选/多选题必须 4 个）
	Answers  []int    `json:"answers"`                      // 答案下标
}

// QuestionResponse 返回给前端的题目信息。
type QuestionResponse struct {
	ID          uint     `json:"id"`          // 题目 ID
	CreatedBy   uint     `json:"createdBy"`   // 创建人 ID
	CreatorName string   `json:"creatorName"` // 创建人名称
	Type        string   `json:"type"`        // 题型
	Language    string   `json:"language"`    // 语言
	Title       string   `json:"title"`       // 标题
	Content     string   `json:"content"`     // 内容
	Options     []string `json:"options"`     // 选项
	Answers     []int    `json:"answers"`     // 答案下标
	CreatedAt   string   `json:"createdAt"`   // 创建时间
}

// List 处理题目列表请求。
//
// GET /api/questions?page=1&pageSize=10&keyword=xxx&type=single&language=go
func (h *QuestionHandler) List(c *gin.Context) {
	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)
	filters := repositories.QuestionFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Type:     strings.TrimSpace(c.Query("type")),
		Language: strings.TrimSpace(c.Query("language")),
		Page:     page,
		PageSize: pageSize,
	}

	ctx := context.Background()
	items, total, err := h.repo.List(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	// 将数据库模型转换为响应结构
	responses := make([]QuestionResponse, 0, len(items))
	for _, q := range items {
		resp, convErr := toQuestionResponse(q, h.creatorName(ctx, q.CreatedBy))
		if convErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "解析题目失败", "error": convErr.Error()})
			return
		}
		responses = append(responses, resp)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses, "total": total})
}

// Create 处理创建题目请求。
//
// POST /api/questions
func (h *QuestionHandler) Create(c *gin.Context) {
	var payload questionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 学生不能创建题目
	if middleware.GetCurrentUserRole(c) == "student" {
		c.JSON(http.StatusForbidden, gin.H{"message": "学生无权创建题目"})
		return
	}

	// 将 payload 转换为数据库模型
	model, err := payload.toModel()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	model.CreatedBy = middleware.GetCurrentUserID(c)

	if err := h.repo.Create(context.Background(), model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存失败", "error": err.Error()})
		return
	}

	resp, err := toQuestionResponse(*model, h.creatorName(context.Background(), model.CreatedBy))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "解析失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功", "data": resp})
}

// Update 处理更新题目请求。
//
// PUT /api/questions/:id
func (h *QuestionHandler) Update(c *gin.Context) {
	idVal, err := strconv.Atoi(c.Param("id"))
	if err != nil || idVal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}
	var payload questionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	updates, err := payload.toMap()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.repo.Update(context.Background(), uint(idVal), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "error": err.Error()})
		return
	}

	item, err := h.repo.FindByID(context.Background(), uint(idVal))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	resp, err := toQuestionResponse(*item, h.creatorName(context.Background(), item.CreatedBy))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "解析失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": resp})
}

// Delete 处理删除单个题目请求。
//
// DELETE /api/questions/:id
func (h *QuestionHandler) Delete(c *gin.Context) {
	idVal, err := strconv.Atoi(c.Param("id"))
	if err != nil || idVal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	if err := h.repo.Delete(context.Background(), uint(idVal)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteMany 处理批量删除题目请求。
//
// DELETE /api/questions
// Body: {"ids": [1, 2, 3]}
func (h *QuestionHandler) DeleteMany(c *gin.Context) {
	var payload struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	if err := h.repo.DeleteMany(context.Background(), payload.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// toQuestionResponse 将数据库模型转换为响应结构体。
//
// 主要工作是将 JSON 字符串（OptionsJSON、AnswerJSON）反序列化为 Go 切片。
func toQuestionResponse(q models.Question, creatorName string) (QuestionResponse, error) {
	var options []string
	var answers []int
	if q.OptionsJSON != "" {
		if err := json.Unmarshal([]byte(q.OptionsJSON), &options); err != nil {
			return QuestionResponse{}, err
		}
	}
	if q.AnswerJSON != "" {
		if err := json.Unmarshal([]byte(q.AnswerJSON), &answers); err != nil {
			return QuestionResponse{}, err
		}
	}
	return QuestionResponse{
		ID:          q.ID,
		CreatedBy:   q.CreatedBy,
		CreatorName: creatorName,
		Type:        q.Type,
		Language:    q.Language,
		Title:       q.Title,
		Content:     q.Content,
		Options:     options,
		Answers:     answers,
		CreatedAt:   q.CreatedAt.Format(time.RFC3339),
	}, nil
}

// creatorName 查询创建人的用户名。
func (h *QuestionHandler) creatorName(ctx context.Context, userID uint) string {
	if userID == 0 || h.userRepo == nil {
		return ""
	}
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ""
	}
	return user.Username
}

// toModel 将请求体转换为数据库模型。
func (p questionPayload) toModel() (*models.Question, error) {
	opts, ans, err := marshalOptionsAnswers(p)
	if err != nil {
		return nil, err
	}
	return &models.Question{
		Type:        p.Type,
		Language:    p.Language,
		Title:       p.Title,
		Content:     p.Content,
		OptionsJSON: opts,
		AnswerJSON:  ans,
	}, nil
}

// toMap 将请求体转换为更新 map。
func (p questionPayload) toMap() (map[string]interface{}, error) {
	opts, ans, err := marshalOptionsAnswers(p)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"type":     p.Type,
		"language": p.Language,
		"title":    p.Title,
		"content":  p.Content,
		"options":  opts,
		"answers":  ans,
	}, nil
}

// marshalOptionsAnswers 将选项和答案序列化为 JSON 字符串。
//
// 校验规则：
// - 单选/多选题必须有 4 个选项
// - 编程题不需要选项和答案
func marshalOptionsAnswers(p questionPayload) (string, string, error) {
	if p.Type != "coding" {
		if len(p.Options) != 4 {
			return "", "", gin.Error{Err: nil, Type: gin.ErrorTypeBind, Meta: "单选/多选题必须提供4个选项"}
		}
	} else {
		p.Options = nil
		p.Answers = nil
	}
	optsBytes, _ := json.Marshal(p.Options)
	ansBytes, _ := json.Marshal(p.Answers)
	return string(optsBytes), string(ansBytes), nil
}

// parseIntWithDefault 将字符串解析为整数，如果解析失败返回默认值。
func parseIntWithDefault(val string, def int) int {
	if v, err := strconv.Atoi(val); err == nil && v > 0 {
		return v
	}
	return def
}
