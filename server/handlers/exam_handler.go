// ============================================================================
// handlers/exam_handler.go - 考试接口处理器
// ============================================================================
//
// 本文件实现了学生考试相关的 HTTP 接口，包括：
// - ListPublished:  已发布考试与历史记录
// - StartAttempt:   开始答题
// - GetAttempt:     获取答题详情
// - SaveAnswers:    保存答案（自动保存）
// - SubmitAttempt:  交卷
// - GetResult:      查看考试结果
// - RecordEvent:    记录监考事件
//
// 答题流程：
//   查看已发布考试与历史记录 → 开始答题 → 自动保存答案 → 交卷 → 查看结果
//
// 学习要点：
// - 考试时间窗口的验证
// - 自动保存的实现
// - 自动阅卷的实现
// - 截止时间的计算
// ============================================================================

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
)

// ExamHandler 处理考试相关的 HTTP 请求。
type ExamHandler struct {
	examRepo     *repositories.ExamRepository
	paperRepo    *repositories.PaperRepository
	questionRepo *repositories.QuestionRepository
	classRepo    *repositories.ClassRepository
}

// NewExamHandler 创建一个新的 ExamHandler 实例。
func NewExamHandler(examRepo *repositories.ExamRepository, paperRepo *repositories.PaperRepository, questionRepo *repositories.QuestionRepository, classRepo *repositories.ClassRepository) *ExamHandler {
	return &ExamHandler{examRepo: examRepo, paperRepo: paperRepo, questionRepo: questionRepo, classRepo: classRepo}
}

// ---- 请求体结构体 ----

// saveAnswersRequest 保存答案请求体
type saveAnswersRequest struct {
	Answers []answerInput `json:"answers" binding:"required,min=1"` // 答案列表
}

// answerInput 单个答案
type answerInput struct {
	QuestionID uint   `json:"questionId" binding:"required"` // 题目 ID
	AnswerJSON string `json:"answerJson"`                    // 答案（JSON 字符串）
}

// proctorEventRequest 监考事件请求体
type proctorEventRequest struct {
	EventType   string `json:"eventType" binding:"required"` // 事件类型
	PayloadJSON string `json:"payloadJson"`                  // 事件数据
}

// ---- 响应体结构体 ----

// ExamAttemptResponse 答题记录响应
type ExamAttemptResponse struct {
	ID          uint                 `json:"id"`                 // 答题记录 ID。未参加试卷详情为 0
	PaperID     uint                 `json:"paperId"`            // 试卷 ID
	StudentID   uint                 `json:"studentId"`          // 学生 ID
	StartedAt   string               `json:"startedAt"`          // 开始时间
	SubmittedAt *string              `json:"submittedAt"`        // 交卷时间
	Status      string               `json:"status"`             // 状态
	TotalScore  *int                 `json:"totalScore"`         // 总分
	Deadline    *string              `json:"deadline,omitempty"` // 截止时间
	Paper       *PaperResponse       `json:"paper,omitempty"`    // 试卷详情
	Answers     []ExamAnswerResponse `json:"answers,omitempty"`  // 答案列表
}

// ExamAnswerResponse 答案响应
type ExamAnswerResponse struct {
	ID         uint   `json:"id"`         // 答案 ID
	AttemptID  uint   `json:"attemptId"`  // 答题记录 ID
	QuestionID uint   `json:"questionId"` // 题目 ID
	AnswerJSON string `json:"answerJson"` // 答案
	IsCorrect  *bool  `json:"isCorrect"`  // 是否正确
	Score      *int   `json:"score"`      // 得分
}

// PublishedPaperResponse 已发布试卷响应（学生视角）
type PublishedPaperResponse struct {
	PaperID            uint    `json:"paperId"`                      // 试卷 ID
	Title              string  `json:"title"`                        // 标题
	Language           string  `json:"language"`                     // 语言
	TotalScore         int     `json:"totalScore"`                   // 试卷总分
	StartTime          string  `json:"startTime"`                    // 开始时间
	EndTime            string  `json:"endTime"`                      // 结束时间
	Duration           int     `json:"duration"`                     // 答题时长（分钟）
	AttemptID          *uint   `json:"attemptId,omitempty"`          // 最近一次答题记录 ID
	AttemptStatus      *string `json:"attemptStatus,omitempty"`      // 最近一次答题状态
	AttemptScore       *int    `json:"attemptScore,omitempty"`       // 最近一次答题得分
	AttemptStartedAt   *string `json:"attemptStartedAt,omitempty"`   // 最近一次开始时间
	AttemptSubmittedAt *string `json:"attemptSubmittedAt,omitempty"` // 最近一次提交时间
}

// ---- Handler 方法 ----

// ListPublished 处理获取学生可见考试列表请求。
//
// 查询逻辑：
// 1. 获取学生所属的所有班级 ID
// 2. 查询已发布且对学生可见的试卷，包含历史、当前、未来考试
// 3. 筛选条件：公共试卷 或 学生所在班级的试卷
//
// GET /api/exam/published
func (h *ExamHandler) ListPublished(c *gin.Context) {
	var classIDs []uint
	if user, ok := middleware.GetCurrentUser(c); ok {
		ids, err := h.classRepo.ListClassIDsByStudent(context.Background(), user.ID)
		if err == nil {
			classIDs = ids
		}
	}

	papers, err := h.examRepo.ListPublishedPapers(context.Background(), classIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	responses := make([]PublishedPaperResponse, 0, len(papers))
	for _, p := range papers {
		pub, err := h.examRepo.FindPublicationByPaperIDForExam(context.Background(), p.ID, classIDs)
		if err != nil {
			continue
		}
		response := PublishedPaperResponse{
			PaperID:    p.ID,
			Title:      p.Title,
			Language:   p.Language,
			TotalScore: p.TotalScore,
			StartTime:  pub.StartTime.Format(time.RFC3339),
			EndTime:    pub.EndTime.Format(time.RFC3339),
			Duration:   pub.Duration,
		}

		if user, ok := middleware.GetCurrentUser(c); ok {
			attempt, attemptErr := h.examRepo.FindAttemptByStudentAndPaper(context.Background(), user.ID, p.ID)
			if attemptErr == nil && attempt != nil {
				response.AttemptID = &attempt.ID
				response.AttemptStatus = &attempt.Status
				response.AttemptScore = attempt.TotalScore
				startedAt := attempt.StartedAt.Format(time.RFC3339)
				response.AttemptStartedAt = &startedAt
				if attempt.SubmittedAt != nil {
					submittedAt := attempt.SubmittedAt.Format(time.RFC3339)
					response.AttemptSubmittedAt = &submittedAt
				}
			}
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// StartAttempt 处理开始答题请求。
//
// 流程：
// 1. 验证试卷已发布且在时间窗口内
// 2. 检查是否已有进行中的答题
// 3. 检查是否已提交过
// 4. 创建答题记录
// 5. 计算截止时间
//
// POST /api/exam/papers/:id/start
func (h *ExamHandler) StartAttempt(c *gin.Context) {
	paperID, err := strconv.Atoi(c.Param("id"))
	if err != nil || paperID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}

	studentID := middleware.GetCurrentUserID(c)
	currentUser, _ := middleware.GetCurrentUser(c)
	ctx := context.Background()

	// 获取学生所属班级
	var classIDs []uint
	if currentUser != nil {
		ids, err := h.classRepo.ListClassIDsByStudent(ctx, currentUser.ID)
		if err == nil {
			classIDs = ids
		}
	}

	// 验证试卷发布状态和时间窗口
	pub, err := h.examRepo.FindPublicationByPaperIDForExam(ctx, uint(paperID), classIDs)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "试卷未发布或不在考试时间"})
		return
	}
	now := time.Now()
	if now.Before(pub.StartTime) || now.After(pub.EndTime) {
		c.JSON(http.StatusForbidden, gin.H{"message": "不在考试时间范围内"})
		return
	}

	// 检查是否有进行中的答题
	existing, err := h.examRepo.FindActiveAttempt(ctx, studentID, uint(paperID))
	if err == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{"message": "已有进行中的答题", "data": toAttemptResponse(*existing, nil, nil)})
		return
	}

	// 检查是否已提交过
	submitted, _ := h.examRepo.FindAttemptByStudentAndPaper(ctx, studentID, uint(paperID))
	if submitted != nil && submitted.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"message": "已提交过答卷，不能重复答题"})
		return
	}

	// 创建答题记录
	attempt := models.ExamAttempt{
		PaperID:   uint(paperID),
		StudentID: studentID,
		StartedAt: time.Now(),
		Status:    "in_progress",
	}

	if err := h.examRepo.CreateAttempt(ctx, &attempt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建答题记录失败", "error": err.Error()})
		return
	}

	// 计算截止时间
	resp := toAttemptResponse(attempt, nil, nil)
	deadline := calcDeadline(attempt.StartedAt, pub)
	if deadline != nil {
		dlStr := deadline.Format(time.RFC3339)
		resp.Deadline = &dlStr
	}

	c.JSON(http.StatusOK, gin.H{"message": "开始答题", "data": resp})
}

// GetAttempt 处理获取答题详情请求。
//
// GET /api/exam/attempts/:id
func (h *ExamHandler) GetAttempt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的答题ID"})
		return
	}

	attempt, err := h.examRepo.FindAttemptByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "答题记录不存在"})
		return
	}
	if !canAccessAttempt(c, attempt.StudentID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权查看该答题记录"})
		return
	}

	paper, _ := h.paperRepo.FindByID(context.Background(), attempt.PaperID)
	var paperResp *PaperResponse
	if paper != nil {
		pr := toPaperResponseWithItems(*paper, h.questionRepo)
		paperResp = &pr
	}

	answerResps := make([]ExamAnswerResponse, 0, len(attempt.Answers))
	for _, a := range attempt.Answers {
		answerResps = append(answerResps, toAnswerResponse(a))
	}

	resp := toAttemptResponse(*attempt, paperResp, answerResps)

	// 为进行中的答题添加截止时间
	if attempt.Status == "in_progress" {
		var classID *uint
		if user, ok := middleware.GetCurrentUser(c); ok {
			classID = user.ClassID
		}
		pub, pubErr := h.examRepo.FindPublicationByPaperIDForExam(context.Background(), attempt.PaperID, ptrUintToSlice(classID))
		if pubErr == nil {
			deadline := calcDeadline(attempt.StartedAt, pub)
			if deadline != nil {
				dlStr := deadline.Format(time.RFC3339)
				resp.Deadline = &dlStr
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// SaveAnswers 处理保存答案请求（自动保存）。
//
// PUT /api/exam/attempts/:id/answers
func (h *ExamHandler) SaveAnswers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的答题ID"})
		return
	}

	var req saveAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	ctx := context.Background()

	attempt, err := h.examRepo.FindAttemptByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "答题记录不存在"})
		return
	}
	if !canAccessAttempt(c, attempt.StudentID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权修改该答题记录"})
		return
	}

	if attempt.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"message": "答题已结束"})
		return
	}

	// 检查截止时间
	var classID *uint
	if user, ok := middleware.GetCurrentUser(c); ok {
		classID = user.ClassID
	}
	pub, pubErr := h.examRepo.FindPublicationByPaperIDForExam(ctx, attempt.PaperID, ptrUintToSlice(classID))
	if pubErr == nil {
		deadline := calcDeadline(attempt.StartedAt, pub)
		if deadline != nil && time.Now().After(*deadline) {
			// 超时自动提交
			attempt.Status = "timeout"
			_ = h.autoSubmit(ctx, attempt)
			c.JSON(http.StatusForbidden, gin.H{"message": "答题时间已结束"})
			return
		}
	}

	// 保存答案（Upsert 操作）
	for _, ans := range req.Answers {
		answer := models.ExamAnswer{
			AttemptID:  uint(id),
			QuestionID: ans.QuestionID,
			AnswerJSON: ans.AnswerJSON,
		}
		if err := h.examRepo.UpsertAnswer(ctx, &answer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "保存答案失败", "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

// SubmitAttempt 处理交卷请求。
//
// POST /api/exam/attempts/:id/submit
func (h *ExamHandler) SubmitAttempt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的答题ID"})
		return
	}

	ctx := context.Background()

	attempt, err := h.examRepo.FindAttemptByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "答题记录不存在"})
		return
	}
	if !canAccessAttempt(c, attempt.StudentID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权提交该答题记录"})
		return
	}

	if attempt.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"message": "答题已结束"})
		return
	}

	if err := h.autoSubmit(ctx, attempt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交卷成功"})
}

// GetResult 处理查看考试结果请求。
//
// GET /api/exam/attempts/:id/result
func (h *ExamHandler) GetResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的答题ID"})
		return
	}

	ctx := context.Background()

	attempt, err := h.examRepo.FindAttemptByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "答题记录不存在"})
		return
	}
	if !canAccessAttempt(c, attempt.StudentID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权查看该答卷结果"})
		return
	}

	// 只有已交卷才能查看结果
	if attempt.Status == "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"message": "尚未交卷，无法查看结果"})
		return
	}

	paper, _ := h.paperRepo.FindByID(ctx, attempt.PaperID)
	var paperResp *PaperResponse
	if paper != nil {
		pr := toPaperResponseWithItems(*paper, h.questionRepo)
		paperResp = &pr
	}

	answerResps := make([]ExamAnswerResponse, 0, len(attempt.Answers))
	for _, a := range attempt.Answers {
		answerResps = append(answerResps, toAnswerResponse(a))
	}

	c.JSON(http.StatusOK, gin.H{"data": toAttemptResponse(*attempt, paperResp, answerResps)})
}

// GetPaperDetail 处理按试卷查看详情请求。
//
// 这个接口用于学生考试历史页的“查看详情”：
// - 已交卷/超时：返回最近一次答题记录、分数、答案和试卷题目
// - 未参加/未交卷：也返回试卷题目，答题记录字段用于表达当前状态
//
// GET /api/exam/papers/:id/detail
func (h *ExamHandler) GetPaperDetail(c *gin.Context) {
	paperID, err := strconv.Atoi(c.Param("id"))
	if err != nil || paperID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}

	ctx := context.Background()
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var classIDs []uint
	if currentUser.Role != "admin" {
		ids, err := h.classRepo.ListClassIDsByStudent(ctx, currentUser.ID)
		if err == nil {
			classIDs = ids
		}
	}

	pub, err := h.examRepo.FindPublicationByPaperIDForExam(ctx, uint(paperID), classIDs)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "试卷未发布或无权查看"})
		return
	}

	paper, err := h.paperRepo.FindByID(ctx, uint(paperID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "试卷不存在"})
		return
	}
	paperResp := toPaperResponseWithItems(*paper, h.questionRepo)

	attempt, err := h.examRepo.FindAttemptByStudentAndPaper(ctx, currentUser.ID, uint(paperID))
	if err == nil && attempt != nil {
		answerResps := make([]ExamAnswerResponse, 0, len(attempt.Answers))
		fullAttempt, fullErr := h.examRepo.FindAttemptByID(ctx, attempt.ID)
		if fullErr == nil {
			attempt = fullAttempt
			for _, a := range fullAttempt.Answers {
				answerResps = append(answerResps, toAnswerResponse(a))
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": toAttemptResponse(*attempt, &paperResp, answerResps)})
		return
	}

	if time.Now().Before(pub.EndTime) {
		c.JSON(http.StatusForbidden, gin.H{"message": "考试结束后才能查看试卷详情"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ExamAttemptResponse{
		PaperID:   uint(paperID),
		StudentID: currentUser.ID,
		Status:    "not_started",
		Paper:     &paperResp,
		Answers:   []ExamAnswerResponse{},
	}})
}

// RecordEvent 处理记录监考事件请求。
//
// POST /api/exam/attempts/:id/events
func (h *ExamHandler) RecordEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的答题ID"})
		return
	}

	var req proctorEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	event := models.ProctorEvent{
		AttemptID:   uint(id),
		EventType:   req.EventType,
		EventTime:   time.Now(),
		PayloadJSON: req.PayloadJSON,
	}

	attempt, err := h.examRepo.FindAttemptByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "答题记录不存在"})
		return
	}
	if !canAccessAttempt(c, attempt.StudentID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "无权记录该答题事件"})
		return
	}

	if err := h.examRepo.CreateProctorEvent(context.Background(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "记录事件失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "记录成功"})
}

// ---- 辅助函数 ----

// autoSubmit 自动提交并阅卷。
//
// 阅卷逻辑：
// 1. 遍历试卷的每个题目
// 2. 获取学生的答案
// 3. 对比正确答案
// 4. 计算总分
func (h *ExamHandler) autoSubmit(ctx context.Context, attempt *models.ExamAttempt) error {
	now := time.Now()
	updates := map[string]interface{}{
		"submitted_at": now,
		"status":       "submitted",
	}

	// 获取试卷信息
	paper, err := h.paperRepo.FindByID(ctx, attempt.PaperID)
	if err != nil {
		return h.examRepo.UpdateAttempt(ctx, attempt.ID, updates)
	}

	// 阅卷
	totalScore := 0
	for _, item := range paper.Items {
		answer, err := h.examRepo.FindAnswer(ctx, attempt.ID, item.QuestionID)
		if err != nil {
			continue
		}

		isCorrect := false
		score := 0

		q, qErr := h.questionRepo.FindByID(ctx, item.QuestionID)
		if qErr == nil {
			isCorrect, score = gradeAnswer(q, answer, item.Score)
		}

		isCorrectPtr := isCorrect
		scorePtr := score
		_ = h.examRepo.UpdateAnswer(ctx, answer.ID, map[string]interface{}{
			"is_correct": isCorrectPtr,
			"score":      scorePtr,
		})
		totalScore += score
	}

	updates["total_score"] = totalScore
	return h.examRepo.UpdateAttempt(ctx, attempt.ID, updates)
}

// calcDeadline 计算答题截止时间。
//
// 公式：deadline = min(开始答题时间 + Duration, EndTime)
// 如果 Duration 为 0（不限时），则截止时间为 EndTime。
func calcDeadline(startedAt time.Time, pub *models.PaperPublication) *time.Time {
	if pub.Duration <= 0 {
		return &pub.EndTime
	}
	durationDeadline := startedAt.Add(time.Duration(pub.Duration) * time.Minute)
	if durationDeadline.After(pub.EndTime) {
		return &pub.EndTime
	}
	return &durationDeadline
}

// gradeAnswer 阅卷：对比学生答案和正确答案。
//
// 阅卷规则：
// - 答案为空：0 分
// - 答案格式错误：0 分
// - 答案长度不一致：0 分
// - 答案内容不一致：0 分
// - 完全一致：满分
func gradeAnswer(question *models.Question, answer *models.ExamAnswer, maxScore int) (bool, int) {
	if answer.AnswerJSON == "" {
		return false, 0
	}

	var studentAnswers []int
	if err := json.Unmarshal([]byte(answer.AnswerJSON), &studentAnswers); err != nil {
		return false, 0
	}

	var correctAnswers []int
	if err := json.Unmarshal([]byte(question.AnswerJSON), &correctAnswers); err != nil {
		return false, 0
	}

	if len(studentAnswers) != len(correctAnswers) {
		return false, 0
	}

	// 排序后比较
	sortInts(studentAnswers)
	sortInts(correctAnswers)

	for i := range studentAnswers {
		if studentAnswers[i] != correctAnswers[i] {
			return false, 0
		}
	}

	return true, maxScore
}

// sortInts 对整数切片进行冒泡排序。
//
// 注意：这是一个简单的冒泡排序实现，仅用于教学目的。
// 生产环境应该使用标准库的 sort.Ints()。
func sortInts(arr []int) {
	for i := 0; i < len(arr)-1; i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}

// toAttemptResponse 将 ExamAttempt 模型转换为响应结构体。
func toAttemptResponse(a models.ExamAttempt, paper *PaperResponse, answers []ExamAnswerResponse) ExamAttemptResponse {
	resp := ExamAttemptResponse{
		ID:         a.ID,
		PaperID:    a.PaperID,
		StudentID:  a.StudentID,
		StartedAt:  a.StartedAt.Format(time.RFC3339),
		Status:     a.Status,
		TotalScore: a.TotalScore,
		Paper:      paper,
		Answers:    answers,
	}
	if a.SubmittedAt != nil {
		s := a.SubmittedAt.Format(time.RFC3339)
		resp.SubmittedAt = &s
	}
	return resp
}

// toAnswerResponse 将 ExamAnswer 模型转换为响应结构体。
func toAnswerResponse(a models.ExamAnswer) ExamAnswerResponse {
	return ExamAnswerResponse{
		ID:         a.ID,
		AttemptID:  a.AttemptID,
		QuestionID: a.QuestionID,
		AnswerJSON: a.AnswerJSON,
		IsCorrect:  a.IsCorrect,
		Score:      a.Score,
	}
}

// canAccessAttempt 检查当前用户是否有权访问答题记录。
// 管理员可以访问所有记录，学生只能访问自己的记录。
func canAccessAttempt(c *gin.Context, studentID uint) bool {
	role := middleware.GetCurrentUserRole(c)
	if role == "admin" {
		return true
	}
	return middleware.GetCurrentUserID(c) == studentID
}

// ptrUintToSlice 将 *uint 转换为 []uint。
// nil 转换为 nil，非 nil 转换为包含单个元素的切片。
func ptrUintToSlice(id *uint) []uint {
	if id == nil {
		return nil
	}
	return []uint{*id}
}
