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

// PaperHandler exposes endpoints for paper generation and management.
type PaperHandler struct {
	paperRepo    *repositories.PaperRepository
	questionRepo *repositories.QuestionRepository
	classRepo    *repositories.ClassRepository
}

func NewPaperHandler(paperRepo *repositories.PaperRepository, questionRepo *repositories.QuestionRepository, classRepo *repositories.ClassRepository) *PaperHandler {
	return &PaperHandler{paperRepo: paperRepo, questionRepo: questionRepo, classRepo: classRepo}
}

// --- Request/Response structs ---

type generateRequest struct {
	Language      string `json:"language" binding:"required,oneof=go cpp java javascript python"`
	SingleCount   int    `json:"singleCount"`
	MultipleCount int    `json:"multipleCount"`
	CodingCount   int    `json:"codingCount"`
	TotalScore    int    `json:"totalScore" binding:"required,min=1"`
}

type savePaperRequest struct {
	Title      string          `json:"title" binding:"required"`
	Language   string          `json:"language" binding:"required,oneof=go cpp java javascript python"`
	Items      []savePaperItem `json:"items" binding:"required,min=1"`
	TotalScore int             `json:"totalScore" binding:"required,min=1"`
}

type savePaperItem struct {
	QuestionID uint   `json:"questionId" binding:"required"`
	Type       string `json:"type" binding:"required,oneof=single multiple coding"`
	Score      int    `json:"score" binding:"required,min=1"`
	SortNo     int    `json:"sortNo"`
}

type updatePaperRequest struct {
	Title      string `json:"title"`
	Language   string `json:"language"`
	TotalScore *int   `json:"totalScore"`
}

type replaceQuestionRequest struct {
	ItemID     uint `json:"itemId" binding:"required"`
	QuestionID uint `json:"questionId"`
}

type publishRequest struct {
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
	Duration  int    `json:"duration"` // 答题时长(分钟), 0=不限时
	ClassID   *uint  `json:"classId"`
}

// PaperItemResponse is a paper item with question details.
type PaperItemResponse struct {
	ID         uint              `json:"id"`
	PaperID    uint              `json:"paperId"`
	QuestionID uint              `json:"questionId"`
	Type       string            `json:"type"`
	Score      int               `json:"score"`
	SortNo     int               `json:"sortNo"`
	Question   *QuestionResponse `json:"question,omitempty"`
}

// PaperResponse is the paper detail response.
type PaperResponse struct {
	ID          uint                 `json:"id"`
	Title       string               `json:"title"`
	Language    string               `json:"language"`
	TotalScore  int                  `json:"totalScore"`
	Status      string               `json:"status"`
	CreatedBy   uint                 `json:"createdBy"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	Items       []PaperItemResponse  `json:"items,omitempty"`
	Publication *PublicationResponse `json:"publication,omitempty"`
}

type PublicationResponse struct {
	ID          uint   `json:"id"`
	PaperID     uint   `json:"paperId"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Duration    int    `json:"duration"` // 答题时长(分钟), 0=不限时
	IsPublished bool   `json:"isPublished"`
}

// --- Handlers ---

// Generate handles POST /api/papers/generate
// Generates a draft paper by randomly selecting questions.
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

	// Calculate per-question score
	for _, cr := range counts {
		totalRequested += cr.count
	}
	if totalRequested == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "至少需要选择一种题型"})
		return
	}
	scorePerQuestion := req.TotalScore / totalRequested
	remainder := req.TotalScore % totalRequested

	for _, cr := range counts {
		if cr.count <= 0 {
			continue
		}
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

// Create handles POST /api/papers
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

// List handles GET /api/papers
func (h *PaperHandler) List(c *gin.Context) {
	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)
	filters := repositories.PaperFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}
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

// Get handles GET /api/papers/:id
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

// Update handles PUT /api/papers/:id
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

// Delete handles DELETE /api/papers/:id
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

// ReplaceQuestion handles POST /api/papers/:id/replace-question
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

	// If no specific question ID provided, randomly pick one of the same type
	if req.QuestionID == 0 {
		// Find the item to get its type
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

		// Get all current question IDs in the paper for exclusion
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

// DeleteItem handles DELETE /api/papers/:id/items/:itemId
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

	// Get the item to find its score for recalculating total
	paper, err := h.paperRepo.FindByID(ctx, uint(paperID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "试卷不存在"})
		return
	}

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

	// Recalculate total score
	newTotal := paper.TotalScore - deletedScore
	_ = h.paperRepo.Update(ctx, uint(paperID), map[string]interface{}{"total_score": newTotal})

	paper, _ = h.paperRepo.FindByID(ctx, uint(paperID))
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "data": toPaperResponseWithItems(*paper, h.questionRepo)})
}

// Publish handles POST /api/papers/:id/publish
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

	// Update paper status to published
	if err := h.paperRepo.Update(ctx, uint(id), map[string]interface{}{"status": "published"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新状态失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "发布成功"})
}

// Unpublish handles POST /api/papers/:id/unpublish
func (h *PaperHandler) Unpublish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的ID"})
		return
	}

	ctx := context.Background()

	// Find and update the publication
	pub, err := h.paperRepo.FindPublicationByPaperID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "未找到发布记录"})
		return
	}

	if err := h.paperRepo.UpdatePublication(ctx, pub.ID, map[string]interface{}{"is_published": false}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "取消发布失败", "error": err.Error()})
		return
	}

	// Update paper status back to draft
	if err := h.paperRepo.Update(ctx, uint(id), map[string]interface{}{"status": "draft"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新状态失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消发布成功"})
}

// GetSubmissions handles GET /api/papers/:id/submissions
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

// --- Helper functions ---

func toPaperResponse(p models.Paper) PaperResponse {
	resp := PaperResponse{
		ID:         p.ID,
		Title:      p.Title,
		Language:   p.Language,
		TotalScore: p.TotalScore,
		Status:     p.Status,
		CreatedBy:  p.CreatedBy,
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt.Format(time.RFC3339),
	}
	return resp
}

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
		// Try to load question details
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
