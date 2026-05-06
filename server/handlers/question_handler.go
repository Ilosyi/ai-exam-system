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

// QuestionHandler exposes CRUD endpoints for questions stored in SQLite.
// QuestionHandler exposes CRUD endpoints for questions stored in SQLite.
//
// 责任说明（面向初学者）：
// - 该 handler 将 HTTP 请求解析为结构化数据（如 JSON），
// - 校验参数的合法性，
// - 调用 repository 完成数据库操作，
// - 将结果序列化为标准 JSON 响应返回给客户端。
type QuestionHandler struct {
	// repo 是实际的数据访问层（Repository），它封装了与数据库的所有交互。
	repo     *repositories.QuestionRepository
	userRepo *repositories.UserRepository
}

// NewQuestionHandler 创建一个 QuestionHandler，并注入对应的 repository。
// 这种“依赖注入”方式便于在测试时替换为 mock，实现单元测试。
func NewQuestionHandler(repo *repositories.QuestionRepository, userRepo *repositories.UserRepository) *QuestionHandler {
	return &QuestionHandler{repo: repo, userRepo: userRepo}
}

// questionPayload describes create/update payloads.
// questionPayload 描述来自前端创建/更新请求的 JSON 结构。
// 使用 Gin 的 binding tag 做简单校验，保证接收到的数据基本满足要求。
type questionPayload struct {
	Type     string   `json:"type" binding:"required,oneof=single multiple coding"`
	Language string   `json:"language" binding:"required,oneof=go cpp java javascript python"`
	Title    string   `json:"title" binding:"required"`
	Content  string   `json:"content" binding:"required"`
	Options  []string `json:"options"`
	Answers  []int    `json:"answers"`
}

// QuestionResponse is returned to the client.
// QuestionResponse 是返回给前端的题目信息结构，字段都已做了 JSON 标签处理，便于序列化。
// CreatedAt 使用字符串格式（RFC3339），便于前端直接显示或格式化。
type QuestionResponse struct {
	ID          uint     `json:"id"`
	CreatedBy   uint     `json:"createdBy"`
	CreatorName string   `json:"creatorName"`
	Type        string   `json:"type"`
	Language    string   `json:"language"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Options     []string `json:"options"`
	Answers     []int    `json:"answers"`
	CreatedAt   string   `json:"createdAt"`
}

// List handles GET /api/questions
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

	// 使用空的 context（可以在未来添加超时或请求追踪信息）
	ctx := context.Background()
	// 从 repository 拉取数据，注意：repository 处理具体的 SQL/ORM 逻辑
	items, total, err := h.repo.List(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	// 将数据库模型转换为响应结构（包含 JSON 反序列化选项/答案）
	responses := make([]QuestionResponse, 0, len(items))
	for _, q := range items {
		resp, convErr := toQuestionResponse(q, h.creatorName(ctx, q.CreatedBy))
		if convErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "解析题目失败", "error": convErr.Error()})
			return
		}
		responses = append(responses, resp)
	}

	// 返回：data（列表）和 total（用于分页）
	c.JSON(http.StatusOK, gin.H{"data": responses, "total": total})
}

// Create handles POST /api/questions
func (h *QuestionHandler) Create(c *gin.Context) {
	var payload questionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}
	if middleware.GetCurrentUserRole(c) == "student" {
		c.JSON(http.StatusForbidden, gin.H{"message": "学生无权创建题目"})
		return
	}

	// 将 payload 转换为数据库模型（包含将 options/answers 序列化为 JSON 字符串）
	model, err := payload.toModel()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	model.CreatedBy = middleware.GetCurrentUserID(c)

	// 调用 repository 完成持久化
	if err := h.repo.Create(context.Background(), model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存失败", "error": err.Error()})
		return
	}

	// 将插入后的模型返回给客户端（方便前端展示新创建的记录）
	resp, err := toQuestionResponse(*model, h.creatorName(context.Background(), model.CreatedBy))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "解析失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功", "data": resp})
}

// Update handles PUT /api/questions/:id
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

	// 将 payload 转换成 Updates map，便于部分更新（数据库层使用 map 更新字段）
	updates, err := payload.toMap()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// 执行更新，若不存在将返回错误
	if err := h.repo.Update(context.Background(), uint(idVal), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "error": err.Error()})
		return
	}

	// 查询更新后的记录并返回给客户端
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

// Delete handles DELETE /api/questions/:id
func (h *QuestionHandler) Delete(c *gin.Context) {
	idVal, err := strconv.Atoi(c.Param("id"))
	if err != nil || idVal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	// 删除记录（若删除失败则返回 500）
	if err := h.repo.Delete(context.Background(), uint(idVal)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}
	// 删除成功返回 204 No Content
	c.Status(http.StatusNoContent)
}

// DeleteMany handles DELETE /api/questions with body {"ids":[1,2]}
func (h *QuestionHandler) DeleteMany(c *gin.Context) {
	var payload struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}
	// 批量删除
	if err := h.repo.DeleteMany(context.Background(), payload.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func toQuestionResponse(q models.Question, creatorName string) (QuestionResponse, error) {
	var options []string
	var answers []int
	if q.OptionsJSON != "" {
		// 将数据库中保存的 JSON 字符串反序列化为切片
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

func marshalOptionsAnswers(p questionPayload) (string, string, error) {
	if p.Type != "coding" {
		// 对于单选/多选题，强制要求 4 个选项，否则返回绑定错误（binding error）
		if len(p.Options) != 4 {
			return "", "", gin.Error{Err: nil, Type: gin.ErrorTypeBind, Meta: "单选/多选题必须提供4个选项"}
		}
	} else {
		p.Options = nil
		p.Answers = nil
	}
	// 将 options/answers 序列化为 JSON 字符串，存储在数据库中
	optsBytes, _ := json.Marshal(p.Options)
	ansBytes, _ := json.Marshal(p.Answers)
	return string(optsBytes), string(ansBytes), nil
}

func parseIntWithDefault(val string, def int) int {
	if v, err := strconv.Atoi(val); err == nil && v > 0 {
		return v
	}
	return def
}
