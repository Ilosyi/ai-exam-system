// ============================================================================
// handlers/note_handler.go - 学习笔记接口处理器
// ============================================================================
//
// 本文件实现了学习笔记相关的 HTTP 接口。
//
// 当前实现非常简单：
// - Get: 读取项目根目录的 README.md 文件，返回其内容
//
// 这个功能的用途是在前端的"学习心得"页面展示项目文档。
//
// 学习要点：
// - 文件读取操作 (os.ReadFile)
// - 路径拼接 (filepath.Join)
// ============================================================================

package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// NoteHandler 处理学习笔记相关的 HTTP 请求。
type NoteHandler struct {
	readmePath string // README.md 文件的绝对路径
}

// NewNoteHandler 创建一个新的 NoteHandler 实例。
//
// 参数：
//   - projectRoot: 项目根目录
func NewNoteHandler(projectRoot string) *NoteHandler {
	return &NoteHandler{readmePath: filepath.Join(projectRoot, "README.md")}
}

// Get 处理获取学习笔记请求。
//
// GET /api/notes
//
// 返回格式：
// {"markdown": "# 项目标题\n\n项目描述..."}
func (h *NoteHandler) Get(c *gin.Context) {
	content, err := os.ReadFile(h.readmePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "读取 README 失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"markdown": string(content)})
}
