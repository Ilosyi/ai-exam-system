// ============================================================================
// handlers/paper_handler.go - 试卷管理接口处理器
// ============================================================================
//
// 本文件实现了试卷管理相关的 HTTP 接口，包括：
// - Generate:        智能组卷（随机选题）
// - Create:          保存试卷
// - List:            试卷列表
// - Get:             试卷详情
// - Update:          更新试卷
// - Delete:          删除试卷
// - ReplaceQuestion: 替换试卷中的题目
// - DeleteItem:      删除试卷中的题目项
// - Publish:         发布试卷
// - Unpublish:       取消发布
// - GetSubmissions:  查看提交统计
//
// 试卷生命周期：
//   draft（草稿）→ published（已发布）→ closed（已关闭）
//
// 学习要点：
// - 复杂的业务逻辑处理
// - 时间格式的解析和验证
// - 关联数据的查询和组装
// ============================================================================

package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
)

// PaperHandler 处理试卷相关的 HTTP 请求。
type PaperHandler struct {
	paperRepo    *repositories.PaperRepository
	questionRepo *repositories.QuestionRepository
	classRepo    *repositories.ClassRepository
}

// NewPaperHandler 创建一个新的 PaperHandler 实例。
func NewPaperHandler(paperRepo *repositories.PaperRepository, questionRepo *repositories.QuestionRepository, classRepo *repositories.ClassRepository) *PaperHandler {
	return &PaperHandler{paperRepo: paperRepo, questionRepo: questionRepo, classRepo: classRepo}
}

// ---- 请求体结构体 ----

// generateRequest 智能组卷请求体
type generateRequest struct {
	Language      string `json:"language" binding:"required,oneof=go cpp java javascript python"` // 语言
	SingleCount   int    `json:"singleCount"`    // 单选题数量
	MultipleCount int    `json:"multipleCount"`  // 多选题数量
	CodingCount   int    `json:"codingCount"`    // 编程题数量
	TotalScore    int    `json:"totalScore" binding:"required,min=1"` // 总分
}

// savePaperRequest 保存试卷请求体
type savePaperRequest struct {
	Title      string          `json:"title" binding:"required"`                                         // 标题
	Language   string          `json:"language" binding:"required,oneof=go cpp java javascript python"` // 语言
	Items      []savePaperItem `json:"items" binding:"required,min=1"`                                   // 题目项
	TotalScore int             `json:"totalScore" binding:"required,min=1"`                              // 总分
}

// savePaperItem 保存试卷的题目项
type savePaperItem struct {
	QuestionID uint   `json:"questionId" binding:"required"`                                       // 题目 ID
	Type       string `json:"type" binding:"required,oneof=single multiple coding"`                // 题型
	Score      int    `json:"score" binding:"required,min=1"`                                       // 分值
	SortNo     int    `json:"sortNo"`                                                               // 排序号
}

// updatePaperRequest 更新试卷请求体
type updatePaperRequest struct {
	Title      string `json:"title"`       // 标题
	Language   string `json:"language"`    // 语言
	TotalScore *int   `json:"totalScore"` // 总分
}

// replaceQuestionRequest 替换题目请求体
type replaceQuestionRequest struct {
	ItemID     uint `json:"itemId" binding:"required"`     // 要替换的题目项 ID
	QuestionID uint `json:"questionId"`                    // 新的题目 ID（0 表示随机选一个）
}

// publishRequest 发布试卷请求体
type publishRequest struct {
	StartTime string `json:"startTime" binding:"required"` // 开始时间（RFC3339 格式）
	EndTime   string `json:"endTime" binding:"required"`   // 结束时间（RFC3339 格式）
	Duration  int    `json:"duration"`                      // 答题时长（分钟），0=不限时
	ClassID   *uint  `json:"classId"`                       // 目标班级 ID（空表示公共试卷）
}

// ---- 响应体结构体 ----

// PaperItemResponse 试卷题目项响应
type PaperItemResponse struct {
	ID         uint              `json:"id"`         // 题目项 ID
	PaperID    uint              `json:"paperId"`    // 试卷 ID
	QuestionID uint              `json:"questionId"` // 题目 ID
	Type       string            `json:"type"`       // 题型
	Score      int               `json:"score"`      // 分值
	SortNo     int               `json:"sortNo"`     // 排序号
	Question   *QuestionResponse `json:"question,omitempty"` // 题目详情
}

// PaperResponse 试卷响应
type PaperResponse struct {
	ID          uint                 `json:"id"`          // 试卷 ID
	Title       string               `json:"title"`       // 标题
	Language    string               `json:"language"`    // 语言
	TotalScore  int                  `json:"totalScore"`  // 总分
	Status      string               `json:"status"`      // 状态
	CreatedBy   uint                 `json:"createdBy"`   // 创建人 ID
	CreatedAt   string               `json:"createdAt"`   // 创建时间
	UpdatedAt   string               `json:"updatedAt"`   // 更新时间
	Items       []PaperItemResponse  `json:"items,omitempty"`       // 题目项列表
	Publication *PublicationResponse `json:"publication,omitempty"` // 发布信息
}

// PublicationResponse 发布信息响应
type PublicationResponse struct {
	ID          uint   `json:"id"`          // 发布记录 ID
	PaperID     uint   `json:"paperId"`     // 试卷 ID
	StartTime   string `json:"startTime"`   // 开始时间
	EndTime     string `json:"endTime"`     // 结束时间
	Duration    int    `json:"duration"`    // 答题时长（分钟）
	IsPublished bool   `json:"isPublished"` // 是否已发布
}

// ---- Handler 方法 ----

// Generate 处理智能组卷请求。
//
// 流程：
// 1. 解析请求参数
// 2. 按题型和数量随机选取题目
// 3. 计算每题分值
// 4. 返回题目列表
//
// POST /api/papers/generate
func (h *PaperHandler) Generate(c *gin.Context) {
	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	ctx := context.Background()
	var items []PaperItemResponse
	sortNo := 0
	totalRequested := 0

	type countReq struct {
		qType string
		count int
	}

	counts := []countReq{
		{"single", req.SingleCount},
		{"multiple", req.MultipleCount},
		{"coding", req.CodingCount},
	}

	// 计算总题目数
	for _, cr := range counts {
		totalRequested += cr.count
	}
	if totalRequested == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "至少需要选择一种题型"})
		return
	}

	// 计算每题分值
	scorePerQuestion := req.TotalScore / totalRequested
	remainder := req.TotalScore % totalRequested

	for _, cr := range counts {
		if cr.count <= 0 {
			continue
		}

		// 随机选取题目
		questions, err := h.paperRepo.RandomQuestions(ctx, cr.qType, req.Language, cr.count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "查询题目失败", "error": err.Error()})
			return
		}
		if len(questions) < cr.count {
			available, _ := h.paperRepo.CountQuestionsByType(ctx, cr.qType, req.Language)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "题目数量不足",
				"error":   "题型 " + cr.qType + " 需要 " + strconv.Itoa(cr.count) + " 题，仅剩 " + strconv.FormatInt(available, 10) + " 题",
			})
			return
		}

		for i, q := range questions {
			sortNo++
			s := scorePerQuestion
			// 将余数分配给第一题
			if i == 0 && remainder > 0 {
				s += remainder
			}
			qResp, _ := toQuestionResponse(q, "")
			items = append(items, PaperItemResponse{
				QuestionID: q.ID,
				Type:       q.Type,
				Score:      s,
				SortNo:     sortNo,
				Question:   &qResp,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items":      items,
			"language":   req.Language,
			"totalScore": req.TotalScore,
		},
	})
}

// Create 处理保存试卷请求。
//
// POST /api/papers
func (h *PaperHandler) Create(c *gin.Context) {
	var req savePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	paper := models.Paper{
		Title:      req.Title,
		Language:   req.Language,
		TotalScore: req.TotalScore,
		Status:     "draft",
		CreatedBy:  middleware.GetCurrentUserID(c),
	}

	for _, item := range req.Items {
		paper.Items = append(paper.Items, models.PaperItem{
			QuestionID: item.QuestionID,
			Type:       item.Type,
			Score:      item.Score,
			SortNo:     item.SortNo,
		})
	}

	if err := h.paperRepo.Create(context.Background(), &paper); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功", "data": toPaperResponse(paper)})
}

// List 处理试卷列表请求。
//
// GET /api/papers?page=1&pageSize=10&keyword=xxx&status=draft
func (h *PaperHandler) List(c *gin.Context) {
	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)
	filters := repositories.PaperFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}
	// 教师只能看到自己创建的试卷
	if middleware.GetCurrentUserRole(c) == "teacher" {
		createdBy := middleware.GetCurrentUserID(c)
		filters.CreatedBy = &createdBy
	}

	items, total, err := h.paperRepo.List(context.Background(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	responses := make([]PaperResponse, 0, len(items))
	for _, p := range items {
		responses = append(responses, toPaperResponse(p))
	}

	c.JSON(http.StatusOK, gin.H{"data": responses, "total": total})
}

// Get 处理获取试卷详情请求。
//
// GET /api/papers/:id
func (h *PaperHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	paper, err := h.paperRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "试卷不存在"})
		return
	}

	resp := toPaperResponseWithItems(*paper, h.questionRepo)
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Update 处理更新试卷请求。
//
// PUT /api/papers/:id
func (h *PaperHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	var req updatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}
	if req.TotalScore != nil {
		updates["total_score"] = *req.TotalScore
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无更新内容"})
		return
	}

	if err := h.paperRepo.Update(context.Background(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "error": err.Error()})
		return
	}

	paper, _ := h.paperRepo.FindByID(context.Background(), uint(id))
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": toPaperResponseWithItems(*paper, h.questionRepo)})
}

// Delete 处理删除试卷请求。
//
// DELETE /api/papers/:id
func (h *PaperHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	if err := h.paperRepo.Delete(context.Background(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReplaceQuestion 处理替换题目请求。
//
// POST /api/papers/:id/replace-question
func (h *PaperHandler) ReplaceQuestion(c *gin.Context) {
	paperID, err := strconv.Atoi(c.Param("id"))
	if err != nil || paperID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}

	var req replaceQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	ctx := context.Background()

	// 如果没有指定新题目 ID，随机选一个同类型的
	if req.QuestionID == 0 {
		paper, err := h.paperRepo.FindByID(ctx, uint(paperID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "试卷不存在"})
			return
		}
		var targetType string
		for _, item := range paper.Items {
			if item.ID == req.ItemID {
				targetType = item.Type
				break
			}
		}
		if targetType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "未找到指定的题目项"})
			return
		}

		existingIDs, _ := h.paperRepo.GetPaperItemIDs(ctx, uint(paperID))
		newQ, err := h.paperRepo.RandomQuestionByTypeLanguage(ctx, targetType, paper.Language, existingIDs)
		if err != nil || newQ == nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "没有可替换的题目"})
			return
		}
		req.QuestionID = newQ.ID
	}

	if err := h.paperRepo.ReplaceItem(ctx, req.ItemID, req.QuestionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "替换失败", "error": err.Error()})
		return
	}

	paper, _ := h.paperRepo.FindByID(ctx, uint(paperID))
	c.JSON(http.StatusOK, gin.H{"message": "替换成功", "data": toPaperResponseWithItems(*paper, h.questionRepo)})
}

// DeleteItem 处理删除题目项请求。
//
// DELETE /api/papers/:id/items/:itemId
func (h *PaperHandler) DeleteItem(c *gin.Context) {
	paperID, err := strconv.Atoi(c.Param("id"))
	if err != nil || paperID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}
	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的题目项ID"})
		return
	}

	ctx := context.Background()

	paper, err := h.paperRepo.FindByID(ctx, uint(paperID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "试卷不存在"})
		return
	}

	// 找到被删除题目的分值
	var deletedScore int
	for _, item := range paper.Items {
		if item.ID == uint(itemID) {
			deletedScore = item.Score
			break
		}
	}

	if err := h.paperRepo.DeleteItem(ctx, uint(itemID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}

	// 重新计算总分
	newTotal := paper.TotalScore - deletedScore
	_ = h.paperRepo.Update(ctx, uint(paperID), map[string]interface{}{"total_score": newTotal})

	paper, _ = h.paperRepo.FindByID(ctx, uint(paperID))
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "data": toPaperResponseWithItems(*paper, h.questionRepo)})
}

// Publish 处理发布试卷请求。
//
// POST /api/papers/:id/publish
func (h *PaperHandler) Publish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	var req publishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "开始时间格式错误", "error": err.Error()})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "结束时间格式错误", "error": err.Error()})
		return
	}

	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "结束时间必须晚于开始时间"})
		return
	}

	ctx := context.Background()
	currentRole := middleware.GetCurrentUserRole(c)
	currentID := middleware.GetCurrentUserID(c)

	// 权限校验：教师只能发布到自己管理的班级
	if req.ClassID != nil {
		if h.classRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "班级仓储未初始化"})
			return
		}
		if currentRole == "teacher" {
			exists, err := h.classRepo.ExistsByTeacher(ctx, *req.ClassID, currentID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "校验班级失败", "error": err.Error()})
				return
			}
			if !exists {
				c.JSON(http.StatusForbidden, gin.H{"message": "无权限发布到该班级"})
				return
			}
		}
	}

	// 创建发布记录
	pub := models.PaperPublication{
		PaperID:     uint(id),
		ClassID:     req.ClassID,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    req.Duration,
		IsPublished: true,
	}
	if err := h.paperRepo.CreatePublication(ctx, &pub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "发布失败", "error": err.Error()})
		return
	}

	// 更新试卷状态为 published
	if err := h.paperRepo.Update(ctx, uint(id), map[string]interface{}{"status": "published"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新状态失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "发布成功"})
}

// Unpublish 处理取消发布请求。
//
// POST /api/papers/:id/unpublish
func (h *PaperHandler) Unpublish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	ctx := context.Background()

	pub, err := h.paperRepo.FindPublicationByPaperID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "未找到发布记录"})
		return
	}

	if err := h.paperRepo.UpdatePublication(ctx, pub.ID, map[string]interface{}{"is_published": false}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "取消发布失败", "error": err.Error()})
		return
	}

	// 更新试卷状态为 draft
	if err := h.paperRepo.Update(ctx, uint(id), map[string]interface{}{"status": "draft"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新状态失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消发布成功"})
}

// GetSubmissions 处理查看提交统计请求。
//
// GET /api/papers/:id/submissions?classId=1
func (h *PaperHandler) GetSubmissions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}

	var classID *uint
	if classIDStr := c.Query("classId"); classIDStr != "" {
		cid, err := strconv.Atoi(classIDStr)
		if err != nil || cid <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "classId 参数无效"})
			return
		}
		ucid := uint(cid)
		classID = &ucid
	}

	stats, err := h.paperRepo.GetSubmissionStats(context.Background(), uint(id), classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ---- 辅助函数 ----

// toPaperResponse 将 Paper 模型转换为响应结构体（不含题目详情）。
func toPaperResponse(p models.Paper) PaperResponse {
	return PaperResponse{
		ID:         p.ID,
		Title:      p.Title,
		Language:   p.Language,
		TotalScore: p.TotalScore,
		Status:     p.Status,
		CreatedBy:  p.CreatedBy,
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt.Format(time.RFC3339),
	}
}

// toPaperResponseWithItems 将 Paper 模型转换为响应结构体（含题目详情）。
func toPaperResponseWithItems(p models.Paper, questionRepo *repositories.QuestionRepository) PaperResponse {
	resp := toPaperResponse(p)

	itemResps := make([]PaperItemResponse, 0, len(p.Items))
	for _, item := range p.Items {
		itemResp := PaperItemResponse{
			ID:         item.ID,
			PaperID:    item.PaperID,
			QuestionID: item.QuestionID,
			Type:       item.Type,
			Score:      item.Score,
			SortNo:     item.SortNo,
		}
		// 查询题目详情
		q, err := questionRepo.FindByID(context.Background(), item.QuestionID)
		if err == nil {
			qResp, _ := toQuestionResponse(*q, "")
			itemResp.Question = &qResp
		}
		itemResps = append(itemResps, itemResp)
	}
	resp.Items = itemResps

	return resp
}
