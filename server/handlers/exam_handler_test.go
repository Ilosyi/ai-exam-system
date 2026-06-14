package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"week05/homework/server/middleware"
	"week05/homework/server/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamHandler_GetPaperDetailAllowsUnattemptedVisiblePaper(t *testing.T) {
	tc := SetupTestContext(t)
	paper := createPublishedPaperForExamDetail(t, tc, nil)
	student := getTestUser(t, tc, "student")

	w := makeAuthenticatedRequest(t, tc, "GET", "/api/exam/papers/1/detail", nil, func(c *gin.Context) {
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		tc.ExamHandler.GetPaperDetail(c)
	}, student)

	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Data ExamAttemptResponse `json:"data"`
	}
	ParseResponse(t, w, &body)

	assert.Equal(t, paper.ID, body.Data.PaperID)
	assert.Equal(t, student.ID, body.Data.StudentID)
	assert.Equal(t, "not_started", body.Data.Status)
	assert.Nil(t, body.Data.TotalScore)
	require.NotNil(t, body.Data.Paper)
	assert.Equal(t, paper.Title, body.Data.Paper.Title)
	assert.Len(t, body.Data.Paper.Items, 1)
}

func TestExamHandler_GetPaperDetailRejectsInvisibleClassPaper(t *testing.T) {
	tc := SetupTestContext(t)
	otherClassID := uint(99)
	createPublishedPaperForExamDetail(t, tc, &otherClassID)
	student := getTestUser(t, tc, "student")

	w := makeAuthenticatedRequest(t, tc, "GET", "/api/exam/papers/1/detail", nil, func(c *gin.Context) {
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		tc.ExamHandler.GetPaperDetail(c)
	}, student)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestExamHandler_GetPaperDetailRejectsUnattemptedPaperBeforeEndTime(t *testing.T) {
	tc := SetupTestContext(t)
	createPublishedPaperForExamDetailWindow(t, tc, nil, time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	student := getTestUser(t, tc, "student")

	w := makeAuthenticatedRequest(t, tc, "GET", "/api/exam/papers/1/detail", nil, func(c *gin.Context) {
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		tc.ExamHandler.GetPaperDetail(c)
	}, student)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func createPublishedPaperForExamDetail(t *testing.T, tc *TestContext, classID *uint) models.Paper {
	return createPublishedPaperForExamDetailWindow(t, tc, classID, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
}

func createPublishedPaperForExamDetailWindow(t *testing.T, tc *TestContext, classID *uint, startTime time.Time, endTime time.Time) models.Paper {
	t.Helper()

	question := models.Question{
		Title:       "Go 单选题",
		Type:        "single",
		Language:    "go",
		Content:     "请选择正确声明方式",
		OptionsJSON: `["var x int","let x","const x","dim x"]`,
		AnswerJSON:  `[0]`,
	}
	require.NoError(t, tc.DB.Create(&question).Error)

	paper := models.Paper{
		Title:      "Go 历史考试",
		Language:   "go",
		TotalScore: 10,
		Status:     "published",
	}
	require.NoError(t, tc.DB.Create(&paper).Error)

	item := models.PaperItem{
		PaperID:    paper.ID,
		QuestionID: question.ID,
		Type:       question.Type,
		Score:      10,
		SortNo:     1,
	}
	require.NoError(t, tc.DB.Create(&item).Error)

	publication := models.PaperPublication{
		PaperID:     paper.ID,
		ClassID:     classID,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    30,
		IsPublished: true,
	}
	require.NoError(t, tc.DB.Create(&publication).Error)

	return paper
}

func getTestUser(t *testing.T, tc *TestContext, username string) models.User {
	t.Helper()

	var user models.User
	require.NoError(t, tc.DB.Where("username = ?", username).First(&user).Error)
	return user
}

func makeAuthenticatedRequest(t *testing.T, tc *TestContext, method string, path string, body interface{}, handler gin.HandlerFunc, user models.User) *httptest.ResponseRecorder {
	t.Helper()

	w := MakeRequest(t, method, path, body, func(c *gin.Context) {
		c.Set(middleware.CurrentUserKey, user)
		handler(c)
	}, nil)
	return w
}
