package repositories

import (
	"context"
	"testing"
	"time"

	"week05/homework/server/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamRepository_ListPublishedPapersIncludesVisibleHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewExamRepository(db)
	ctx := context.Background()
	now := time.Now()
	classID := uint(1)
	otherClassID := uint(2)

	papers := []models.Paper{
		{Title: "历史公共考试", Language: "go", TotalScore: 100, Status: "published"},
		{Title: "未来公共考试", Language: "go", TotalScore: 100, Status: "published"},
		{Title: "当前班级考试", Language: "go", TotalScore: 100, Status: "published"},
		{Title: "其他班级考试", Language: "go", TotalScore: 100, Status: "published"},
		{Title: "未发布考试", Language: "go", TotalScore: 100, Status: "draft"},
	}
	require.NoError(t, db.Create(&papers).Error)

	publicPast := models.PaperPublication{
		PaperID:     papers[0].ID,
		StartTime:   now.Add(-48 * time.Hour),
		EndTime:     now.Add(-24 * time.Hour),
		IsPublished: true,
	}
	publicFuture := models.PaperPublication{
		PaperID:     papers[1].ID,
		StartTime:   now.Add(24 * time.Hour),
		EndTime:     now.Add(48 * time.Hour),
		IsPublished: true,
	}
	classCurrent := models.PaperPublication{
		PaperID:     papers[2].ID,
		ClassID:     &classID,
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now.Add(1 * time.Hour),
		IsPublished: true,
	}
	otherClassCurrent := models.PaperPublication{
		PaperID:     papers[3].ID,
		ClassID:     &otherClassID,
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now.Add(1 * time.Hour),
		IsPublished: true,
	}
	unpublished := models.PaperPublication{
		PaperID:     papers[4].ID,
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now.Add(1 * time.Hour),
		IsPublished: false,
	}
	require.NoError(t, db.Create(&[]models.PaperPublication{publicPast, publicFuture, classCurrent, otherClassCurrent, unpublished}).Error)

	got, err := repo.ListPublishedPapers(ctx, []uint{classID})
	require.NoError(t, err)

	titles := make([]string, 0, len(got))
	for _, paper := range got {
		titles = append(titles, paper.Title)
	}

	assert.ElementsMatch(t, []string{"历史公共考试", "未来公共考试", "当前班级考试"}, titles)
}

func TestExamRepository_FindAttemptByStudentAndPaperUsesLatestStartedAttempt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewExamRepository(db)
	ctx := context.Background()
	now := time.Now()
	score := 88

	attempts := []models.ExamAttempt{
		{
			PaperID:    10,
			StudentID:  3,
			StartedAt:  now.Add(-2 * time.Hour),
			Status:     "submitted",
			TotalScore: nil,
		},
		{
			PaperID:    10,
			StudentID:  3,
			StartedAt:  now.Add(-1 * time.Hour),
			Status:     "submitted",
			TotalScore: &score,
		},
	}
	require.NoError(t, db.Create(&attempts).Error)

	got, err := repo.FindAttemptByStudentAndPaper(ctx, 3, 10)
	require.NoError(t, err)

	assert.Equal(t, attempts[1].ID, got.ID)
	require.NotNil(t, got.TotalScore)
	assert.Equal(t, score, *got.TotalScore)
}
