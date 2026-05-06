package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// NoteHandler returns markdown content for the study notes page.
type NoteHandler struct {
	readmePath string
}

func NewNoteHandler(projectRoot string) *NoteHandler {
	return &NoteHandler{readmePath: filepath.Join(projectRoot, "README.md")}
}

func (h *NoteHandler) Get(c *gin.Context) {
	content, err := os.ReadFile(h.readmePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "读取 README 失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"markdown": string(content)})
}
