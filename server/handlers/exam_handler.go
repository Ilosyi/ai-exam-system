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

// ExamHandler exposes endpoints for the student exam experience.
type ExamHandler struct {
	examRepo     *repositories.ExamRepository
	paperRepo    *repositories.PaperRepository
	questionRepo *repositories.QuestionRepository
	classRepo    *repositories.ClassRepository
}

func NewExamHandler(examRepo *repositories.ExamRepository, paperRepo *repositories.PaperRepository, questionRepo *repositories.QuestionRepository, classRepo *repositories.ClassRepository) *ExamHandler {
	return &ExamHandler{examRepo: examRepo, paperRepo: paperRepo, questionRepo: questionRepo, classRepo: classRepo}
}

// --- Request/Response structs ---

type saveAnswersRequest struct {
	Answers []answerInput `json:"answers" binding:"required,min=1"`
}

type answerInput struct {
	QuestionID uint   `json:"questionId" binding:"required"`
	AnswerJSON string `json:"answerJson"`
}

type proctorEventRequest struct {
	EventType   string `json:"eventType" binding:"required"`
	PayloadJSON string `json:"payloadJson"`
}

// ExamAttemptResponse is the response for an exam attempt.
type ExamAttemptResponse struct {
	ID          uint                 `json:"id"`
	PaperID     uint                 `json:"paperId"`
	StudentID   uint                 `json:"studentId"`
	StartedAt   string               `json:"startedAt"`
	SubmittedAt *string              `json:"submittedAt"`
	Status      string               `json:"status"`
	TotalScore  *int                 `json:"totalScore"`
	Deadline    *string              `json:"deadline,omitempty"` // 答题截止时间(含答题时长)
	Paper       *PaperResponse       `json:"paper,omitempty"`
	Answers     []ExamAnswerResponse `json:"answers,omitempty"`
}

type ExamAnswerResponse struct {
	ID         uint   `json:"id"`
	AttemptID  uint   `json:"attemptId"`
	QuestionID uint   `json:"questionId"`
	AnswerJSON string `json:"answerJson"`
	IsCorrect  *bool  `json:"isCorrect"`
	Score      *int   `json:"score"`
}

// PublishedPaperResponse is for the student-facing published paper list.
type PublishedPaperResponse struct {
	PaperID    uint   `json:"paperId"`
	Title      string `json:"title"`
	Language   string `json:"language"`
	TotalScore int    `json:"totalScore"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Duration   int    `json:"duration"` // 答题时长(分钟), 0=不限时
}

// --- Handlers ---

// ptrUintToSlice converts an optional *uint class ID into the []uint form expected by repositories.
// A nil pointer means "no class" (public papers only); a non-nil value means the student belongs to that class.
func ptrUintToSlice(id *uint) []uint {
	if id == nil {
		return nil
	}
	return []uint{*id}
}

// ListPublished handles GET /api/exam/published
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
		responses = append(responses, PublishedPaperResponse{
			PaperID:    p.ID,
			Title:      p.Title,
			Language:   p.Language,
			TotalScore: p.TotalScore,
			StartTime:  pub.StartTime.Format(time.RFC3339),
			EndTime:    pub.EndTime.Format(time.RFC3339),
			Duration:   pub.Duration,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// StartAttempt handles POST /api/exam/papers/:id/start
func (h *ExamHandler) StartAttempt(c *gin.Context) {
	paperID, err := strconv.Atoi(c.Param("id"))
	if err != nil || paperID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的试卷ID"})
		return
	}

	studentID := middleware.GetCurrentUserID(c)
	currentUser, _ := middleware.GetCurrentUser(c)
	ctx := context.Background()

	// Get all class IDs the student belongs to
	var classIDs []uint
	if currentUser != nil {
		ids, err := h.classRepo.ListClassIDsByStudent(ctx, currentUser.ID)
		if err == nil {
			classIDs = ids
		}
	}

	// Check if paper is published and within time window
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

	// Check for existing in-progress attempt
	existing, err := h.examRepo.FindActiveAttempt(ctx, studentID, uint(paperID))
	if err == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{"message": "已有进行中的答题", "data": toAttemptResponse(*existing, nil, nil)})
		return
	}

	// Check for already submitted attempt
	submitted, _ := h.examRepo.FindAttemptByStudentAndPaper(ctx, studentID, uint(paperID))
	if submitted != nil && submitted.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"message": "已提交过答卷，不能重复答题"})
		return
	}

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

	// Calculate deadline: min(startedAt + duration, endTime)
	resp := toAttemptResponse(attempt, nil, nil)
	deadline := calcDeadline(attempt.StartedAt, pub)
	if deadline != nil {
		dlStr := deadline.Format(time.RFC3339)
		resp.Deadline = &dlStr
	}

	c.JSON(http.StatusOK, gin.H{"message": "开始答题", "data": resp})
}

// GetAttempt handles GET /api/exam/attempts/:id
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
	// Add deadline for in-progress attempts
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

// SaveAnswers handles PUT /api/exam/attempts/:id/answers
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

	// Check deadline (considers both exam time window and answer duration)
	var classID *uint
	if user, ok := middleware.GetCurrentUser(c); ok {
		classID = user.ClassID
	}
	pub, pubErr := h.examRepo.FindPublicationByPaperIDForExam(ctx, attempt.PaperID, ptrUintToSlice(classID))
	if pubErr == nil {
		deadline := calcDeadline(attempt.StartedAt, pub)
		if deadline != nil && time.Now().After(*deadline) {
			// Auto-submit on timeout
			attempt.Status = "timeout"
			_ = h.autoSubmit(ctx, attempt)
			c.JSON(http.StatusForbidden, gin.H{"message": "答题时间已结束"})
			return
		}
	}

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

// SubmitAttempt handles POST /api/exam/attempts/:id/submit
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

// GetResult handles GET /api/exam/attempts/:id/result
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

	// Only show result if submitted or timeout
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

// RecordEvent handles POST /api/exam/attempts/:id/events
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

// --- Helper functions ---

func (h *ExamHandler) autoSubmit(ctx context.Context, attempt *models.ExamAttempt) error {
	now := time.Now()
	updates := map[string]interface{}{
		"submitted_at": now,
		"status":       "submitted",
	}

	// Grade the attempt
	paper, err := h.paperRepo.FindByID(ctx, attempt.PaperID)
	if err != nil {
		return h.examRepo.UpdateAttempt(ctx, attempt.ID, updates)
	}

	totalScore := 0
	for _, item := range paper.Items {
		answer, err := h.examRepo.FindAnswer(ctx, attempt.ID, item.QuestionID)
		if err != nil {
			continue
		}

		isCorrect := false
		score := 0

		// Grade the answer
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

// calcDeadline returns the answer deadline for an attempt.
// deadline = min(startedAt + duration, endTime); nil if no limit.
func calcDeadline(startedAt time.Time, pub *models.PaperPublication) *time.Time {
	if pub.Duration <= 0 {
		// No duration limit, deadline is the exam end time
		return &pub.EndTime
	}
	durationDeadline := startedAt.Add(time.Duration(pub.Duration) * time.Minute)
	// If duration extends beyond exam end time, use exam end time
	if durationDeadline.After(pub.EndTime) {
		return &pub.EndTime
	}
	return &durationDeadline
}

func gradeAnswer(question *models.Question, answer *models.ExamAnswer, maxScore int) (bool, int) {
	if answer.AnswerJSON == "" {
		return false, 0
	}

	// Parse the student's answer
	var studentAnswers []int
	if err := json.Unmarshal([]byte(answer.AnswerJSON), &studentAnswers); err != nil {
		return false, 0
	}

	// Parse the correct answer
	var correctAnswers []int
	if err := json.Unmarshal([]byte(question.AnswerJSON), &correctAnswers); err != nil {
		return false, 0
	}

	// Compare
	if len(studentAnswers) != len(correctAnswers) {
		return false, 0
	}

	// Sort both for comparison
	sortInts(studentAnswers)
	sortInts(correctAnswers)

	for i := range studentAnswers {
		if studentAnswers[i] != correctAnswers[i] {
			return false, 0
		}
	}

	return true, maxScore
}

func sortInts(arr []int) {
	for i := 0; i < len(arr)-1; i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}

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
func canAccessAttempt(c *gin.Context, studentID uint) bool {
	role := middleware.GetCurrentUserRole(c)
	if role == "admin" {
		return true
	}
	return middleware.GetCurrentUserID(c) == studentID
}
