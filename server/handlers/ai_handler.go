// ============================================================================
// handlers/ai_handler.go - AI 出题接口处理器
// ============================================================================
//
// 本文件实现了 AI 出题相关的 HTTP 接口，包括：
// - Generate:       AI 生成题目
// - TestConnection: 测试 AI 连接
//
// AI 出题流程：
// 1. 前端发送出题参数（题型、数量、语言、关键词）
// 2. 后端构造 prompt，调用 AI 服务
// 3. AI 返回 JSON 格式的题目数组
// 4. 后端解析、清洗、规范化后返回给前端
//
// 学习要点：
// - prompt 工程的基本概念
// - AI 输出的后处理和规范化
// - 正则表达式的使用
// ============================================================================

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/services"
)

// AIHandler 处理 AI 相关的 HTTP 请求。
type AIHandler struct {
	ai *services.AIService
}

// NewAIHandler 创建一个新的 AIHandler 实例。
func NewAIHandler(ai *services.AIService) *AIHandler {
	return &AIHandler{ai: ai}
}

// AIRequest 定义了 AI 出题请求的参数格式。
type AIRequest struct {
	Type     string `json:"type" binding:"required,oneof=single multiple coding"` // 题型
	Count    int    `json:"count" binding:"required,gte=1,lte=10"`                // 数量（1-10）
	Language string `json:"language" binding:"required,oneof=go cpp java javascript python"` // 语言
	Keyword  string `json:"keyword" binding:"required"`                            // 关键词
}

// AIQuestion 表示从 AI 返回的单个题目。
type AIQuestion struct {
	Type     string   `json:"type"`     // 题型
	Language string   `json:"language"` // 语言
	Title    string   `json:"title"`    // 标题
	Content  string   `json:"content"`  // 内容
	Options  []string `json:"options"`  // 选项
	Answers  []int    `json:"answers"`  // 答案下标
}

// Generate 处理 AI 出题请求。
//
// 流程：
// 1. 解析并校验前端参数
// 2. 根据参数构造 prompt
// 3. 调用 AI 服务获取 JSON 文本
// 4. 解析 JSON 为题目数组
// 5. 规范化题目数据
// 6. 返回给前端
//
// POST /api/ai/generate
func (h *AIHandler) Generate(c *gin.Context) {
	var req AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 构造 prompt 并调用 AI
	prompt := buildPrompt(req)
	jsonText, err := h.ai.GenerateQuestions(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "AI 调用失败", "error": err.Error()})
		return
	}

	// 解析 AI 返回的 JSON
	var questions []AIQuestion
	if err := json.Unmarshal([]byte(jsonText), &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "AI 返回格式不符合要求", "error": err.Error(), "raw": jsonText})
		return
	}

	// 规范化题目数据
	questions = normalizeQuestions(req, questions)

	c.JSON(http.StatusOK, gin.H{
		"prompt":    prompt,
		"raw":       jsonText,
		"questions": questions,
	})
}

// TestConnection 测试 AI 连接。
//
// POST /api/ai/test
func (h *AIHandler) TestConnection(c *gin.Context) {
	diagnostic, err := h.ai.TestConnectionDiagnostic()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "AI 连接测试失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "连接成功",
		"data":    diagnostic,
	})
}

// buildPrompt 根据请求参数构造发送给 AI 的 prompt。
//
// prompt 工程的关键：
// - 明确指定输出格式（JSON 数组）
// - 详细描述每个字段的要求
// - 给出约束条件（如选项数量、答案格式）
func buildPrompt(req AIRequest) string {
	return fmt.Sprintf(strings.TrimSpace(`请根据如下参数批量生成题目，严格返回 JSON 数组：
参数：
- 题型: %s（single=单选，multiple=多选，coding=编程题）
- 数量: %d
- 语言: %s（go/cpp/java/javascript/python）
- 关键词: %s（题目需与此高度相关）

输出要求：
- 仅输出 JSON 数组（无任何多余文字、无 Markdown）。
- 每个元素都包含以下字段：
  {
	"type": "single|multiple|coding",
	"language": "go|cpp|java|javascript|python",
	"title": "题目标题",
	"content": "题干描述，含输入/输出说明",
	"options": ["A", "B", "C", "D"], // 单/多选题必须四个；编程题返回 []
	"answers": [0] // 单选为一个下标；多选为多个；编程题返回 []
  }
- 所有下标从 0 开始。
- 必须严格遵守"题型"参数：
  - 当题型为 single/multiple 时：一定返回 4 个选项，answers 仅包含 0..3 的下标；
  - 当题型为 coding 时：options 与 answers 必须为空数组；
- 确保生成的 JSON 可以被直接解析。`), req.Type, req.Count, req.Language, req.Keyword)
}

// normalizeQuestions 对 AI 返回的题目数据做规范化处理。
//
// 规范化逻辑：
// 1. 强制题型和语言为请求中指定的值
// 2. 对于单/多选题：确保有 4 个选项，清洗选项前缀
// 3. 过滤掉不合法的题目（如缺少标题、答案为空等）
func normalizeQuestions(req AIRequest, in []AIQuestion) []AIQuestion {
	out := make([]AIQuestion, 0, len(in))
	for _, q := range in {
		// 强制覆盖题型和语言
		q.Type = req.Type
		q.Language = req.Language

		switch req.Type {
		case "coding":
			// 编程题不需要选项和答案
			q.Options = nil
			q.Answers = nil
			if strings.TrimSpace(q.Title) == "" || strings.TrimSpace(q.Content) == "" {
				continue
			}
			out = append(out, q)

		case "single", "multiple":
			// 确保有 4 个选项
			if len(q.Options) < 4 {
				continue
			}
			// 清洗选项前缀（如 "A. xxx" → "xxx"）
			cleaned := make([]string, 4)
			for i := 0; i < 4; i++ {
				cleaned[i] = stripOptionLabel(q.Options[i])
			}
			q.Options = cleaned

			// 过滤和修正答案下标
			fixed := make([]int, 0, len(q.Answers))
			for _, idx := range q.Answers {
				if idx >= 0 && idx < 4 {
					fixed = append(fixed, idx)
				}
			}
			if len(fixed) == 0 {
				continue
			}
			// 单选题只保留第一个答案
			if req.Type == "single" && len(fixed) > 1 {
				fixed = fixed[:1]
			}
			q.Answers = fixed

			if strings.TrimSpace(q.Title) == "" || strings.TrimSpace(q.Content) == "" {
				continue
			}
			out = append(out, q)
		}
	}
	return out
}

// optionLabelRe 匹配选项前缀的正则表达式。
// 匹配 A.、A、A:、A）等格式。
var optionLabelRe = regexp.MustCompile(`(?i)^\s*[A-D][\.、:：\)\]）】]?\s*`)

// stripOptionLabel 去除选项前缀。
//
// 例如：
// - "A. Go语言" → "Go语言"
// - "B: Python" → "Python"
// - "C）Java"   → "Java"
func stripOptionLabel(s string) string {
	s = strings.TrimSpace(s)
	s = optionLabelRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
