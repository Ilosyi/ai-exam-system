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

// AIHandler 处理与 AI 交互相关的 HTTP 请求（例如生成试题）。
// 设计目标：将 AI 的输出转换为前端可消费的 JSON 数据，并对数据做必要的校验/清洗。
type AIHandler struct {
	ai *services.AIService
}

func NewAIHandler(ai *services.AIService) *AIHandler {
	return &AIHandler{ai: ai}
}

// AIRequest 定义了前端发起 AI 生成请求时的参数格式。
// binding 标签用于在请求抵达时进行基础校验。
type AIRequest struct {
	Type     string `json:"type" binding:"required,oneof=single multiple coding"`
	Count    int    `json:"count" binding:"required,gte=1,lte=10"`
	Language string `json:"language" binding:"required,oneof=go cpp java javascript python"`
	Keyword  string `json:"keyword" binding:"required"`
}

// AIQuestion 表示从 AI 返回给前端的单个题目结构（用于预览）。
type AIQuestion struct {
	Type     string   `json:"type"`
	Language string   `json:"language"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Options  []string `json:"options"`
	Answers  []int    `json:"answers"`
}

// Generate 处理 POST /api/ai/generate，请求流程：
// 1. 解析并校验前端参数（AIRequest）
// 2. 根据参数构造 prompt，调用 AI 服务获取 JSON 文本
// 3. 将 AI 返回的文本解析为结构化数据（[]AIQuestion），并做清洗（normalize）
// 4. 将结果返回给前端：包含 prompt/raw（用于调试）和 questions（预览用）
func (h *AIHandler) Generate(c *gin.Context) {
	var req AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}
	// 构造 prompt 并调用 AI 服务
	prompt := buildPrompt(req)
	jsonText, err := h.ai.GenerateQuestions(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "AI 调用失败", "error": err.Error()})
		return
	}

	// 将 AI 返回的 JSON 文本解析为 []AIQuestion
	var questions []AIQuestion
	if err := json.Unmarshal([]byte(jsonText), &questions); err != nil {
		// 若解析失败，将原始文本返回给前端以便排查问题
		c.JSON(http.StatusInternalServerError, gin.H{"message": "AI 返回格式不符合要求", "error": err.Error(), "raw": jsonText})
		return
	}

	// 规范化/约束返回数据，确保与前端请求的题型一致并清洗内容
	questions = normalizeQuestions(req, questions)

	c.JSON(http.StatusOK, gin.H{
		"prompt":    prompt,
		"raw":       jsonText,
		"questions": questions,
	})
}

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

// buildPrompt 根据请求参数构造发送给 AI 的 prompt（指令），要求 AI 严格返回可解析的 JSON 数组。
// 这里通过模板指定输出字段和约束，目的是减少 AI 返回不可解析文本的概率。
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
- 必须严格遵守“题型”参数：
  - 当题型为 single/multiple 时：一定返回 4 个选项，answers 仅包含 0..3 的下标；
  - 当题型为 coding 时：options 与 answers 必须为空数组；
- 确保生成的 JSON 可以被直接解析。`), req.Type, req.Count, req.Language, req.Keyword)
}

// normalizeQuestions 对 AI 返回的数据做逐条检查与修正：
// - 强制题型与语言为请求中指定的值
// - 对于单/多选题确保有 4 个选项并清洗选项前缀（如 "A. xxx" -> "xxx"）
// - 过滤掉不合法或缺失关键信息（如 title/content 为空或 answers 不合法）的题目
func normalizeQuestions(req AIRequest, in []AIQuestion) []AIQuestion {
	out := make([]AIQuestion, 0, len(in))
	for _, q := range in {
		// 强制覆盖题型/语言，避免 AI 返回不符合要求的类型
		q.Type = req.Type
		q.Language = req.Language
		switch req.Type {
		case "coding":
			// 编程题不应该有 options/answers
			q.Options = nil
			q.Answers = nil
			// 确保题目包含基本信息
			if strings.TrimSpace(q.Title) == "" || strings.TrimSpace(q.Content) == "" {
				continue
			}
			out = append(out, q)
		case "single", "multiple":
			// 单/多选题需要 4 个选项，否则跳过
			if len(q.Options) < 4 {
				continue
			}
			// 截断并清洗选项前缀（去掉 A./A 等前缀，避免前端重复）
			cleaned := make([]string, 4)
			for i := 0; i < 4; i++ {
				cleaned[i] = stripOptionLabel(q.Options[i])
			}
			q.Options = cleaned
			// 过滤并修正答案下标，移除越界或非法下标
			fixed := make([]int, 0, len(q.Answers))
			for _, idx := range q.Answers {
				if idx >= 0 && idx < 4 {
					fixed = append(fixed, idx)
				}
			}
			if len(fixed) == 0 {
				// 无答案则跳过
				continue
			}
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

// optionLabelRe 用于匹配选项前缀（如 A.、A、A: 等），便于去掉这些标识只保留文本内容
var optionLabelRe = regexp.MustCompile(`(?i)^\s*[A-D][\.、:：\)\]）】]?\s*`)

// stripOptionLabel 去除选项前缀并修剪空白
func stripOptionLabel(s string) string {
	s = strings.TrimSpace(s)
	s = optionLabelRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
