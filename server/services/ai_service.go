// ============================================================================
// services/ai_service.go - AI 服务
// ============================================================================
//
// 本文件封装了与外部 AI 服务（DashScope 兼容接口）的交互逻辑。
//
// 主要职责：
// 1. 从 .env 加载 API Key
// 2. 从 config.json 加载 AI 配置（BaseURL、模型名、系统提示词等）
// 3. 发送 Chat Completion 请求到 AI 服务
// 4. 返回 AI 的回复文本
//
// DashScope 是什么？
// DashScope 是阿里云提供的 AI 模型服务平台，兼容 OpenAI 的 API 格式。
// 这意味着我们可以使用 OpenAI 的官方 SDK 来调用 DashScope 的服务。
//
// 学习要点：
// - OpenAI Chat Completion API 的基本用法
// - 环境变量的加载（.env 文件）
// - 超时控制和错误处理
// - 上下文（context）的使用
// ============================================================================

package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"              // .env 文件加载库
	openai "github.com/openai/openai-go/v3" // OpenAI 官方 Go SDK
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"week05/homework/server/config"
)

// AIService 封装了与外部 AI 服务的交互逻辑。
//
// 它持有 OpenAI 客户端和应用配置，用于发送 AI 请求。
type AIService struct {
	client openai.Client    // OpenAI 客户端（兼容 DashScope）
	cfg    *config.AppConfig // 应用配置
}

// ConnectionDiagnostic 代表 AI 连接测试的诊断信息。
//
// 用于前端展示连接测试结果，包含：
// - 当前使用的模型和 Base URL
// - 请求耗时
// - AI 的回复内容
type ConnectionDiagnostic struct {
	Model       string `json:"model"`       // AI 模型名称
	BaseURL     string `json:"baseURL"`     // API 基础 URL
	ElapsedMs   int64  `json:"elapsedMs"`   // 请求耗时（毫秒）
	Reply       string `json:"reply"`       // AI 的完整回复
	ReplyDigest string `json:"replyDigest"` // AI 回复的摘要（前 120 字符）
}

// NewAIService 创建一个新的 AIService 实例。
//
// 初始化流程：
// 1. 从 .env 文件加载环境变量（主要是 DASHSCOPE_API_KEY）
// 2. 从 config.json 加载应用配置
// 3. 创建 OpenAI 客户端
//
// 参数：
//   - projectRoot: 项目根目录
//
// 返回值：
//   - *AIService: AI 服务实例
//   - error: 初始化失败时返回错误
func NewAIService(projectRoot string) (*AIService, error) {
	// 加载 .env 文件
	// godotenv.Load 会读取 .env 文件并将其中的键值对设置为环境变量
	// 如果 .env 文件不存在，不会报错（返回的错误被忽略）
	_ = godotenv.Load(filepath.Join(projectRoot, ".env"))

	// 获取 API Key
	key := os.Getenv("DASHSCOPE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("missing DASHSCOPE_API_KEY in .env")
	}

	// 加载应用配置
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	// 创建 OpenAI 客户端
	return &AIService{
		client: openai.NewClient(
			option.WithAPIKey(key),        // API Key
			option.WithBaseURL(cfg.BaseURL), // API 基础 URL（如 DashScope 地址）
			option.WithMaxRetries(0),      // 不重试（避免重复消费 token）
		),
		cfg: cfg,
	}, nil
}

// GenerateQuestions 将 prompt 发送给 AI 服务，返回 AI 生成的题目文本。
//
// 请求流程：
// 1. 构造消息列表（系统提示词 + 用户 prompt）
// 2. 调用 Chat Completion API
// 3. 返回 AI 的回复文本
//
// 参数：
//   - prompt: 用户的出题请求
//
// 返回值：
//   - string: AI 返回的 JSON 文本（题目数组）
//   - error: 调用失败时返回错误
func (s *AIService) GenerateQuestions(prompt string) (string, error) {
	return s.runChatCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(s.cfg.SystemPrompt),      // 系统提示词（定义 AI 的行为）
		openai.UserMessage(appendNoThinkSuffix(prompt)), // 用户请求（追加 /no_think 后缀）
	})
}

// TestConnection 测试 AI 连接。
//
// 发送一个简单的 "who are you" 请求，验证：
// - API Key 是否有效
// - Base URL 是否可达
// - AI 服务是否正常响应
func (s *AIService) TestConnection() (string, error) {
	return s.runChatCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("who are you"),
	})
}

// TestConnectionDiagnostic 测试 AI 连接并返回诊断信息。
//
// 与 TestConnection 类似，但会额外记录：
// - 请求耗时
// - AI 回复的摘要
func (s *AIService) TestConnectionDiagnostic() (*ConnectionDiagnostic, error) {
	startedAt := time.Now()
	reply, err := s.TestConnection()
	if err != nil {
		return nil, err
	}

	return &ConnectionDiagnostic{
		Model:       s.cfg.Model,
		BaseURL:     s.cfg.BaseURL,
		ElapsedMs:   time.Since(startedAt).Milliseconds(),
		Reply:       reply,
		ReplyDigest: summarizeReply(reply),
	}, nil
}

// runChatCompletion 执行 Chat Completion 请求。
//
// 这是所有 AI 调用的核心方法，负责：
// 1. 设置超时控制
// 2. 发送请求到 AI 服务
// 3. 处理响应和错误
// 4. 清理 AI 返回的文本（去除 Markdown 代码块包裹）
//
// 参数：
//   - messages: 消息列表（系统提示词 + 用户消息）
//
// 返回值：
//   - string: AI 的回复文本
//   - error: 调用失败时返回错误
func (s *AIService) runChatCompletion(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	// 设置超时
	timeout := time.Duration(s.cfg.RequestTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 确保在函数返回时取消 context

	// 发送请求
	resp, err := s.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    shared.ChatModel(s.cfg.Model),
		},
		option.WithRequestTimeout(timeout),
	)
	if err != nil {
		// 检查是否是超时错误
		if isTimeoutError(err) {
			return "", fmt.Errorf(
				"ai upstream timeout after %ds, request may have been accepted by provider and token may already be consumed",
				s.cfg.RequestTimeoutSeconds,
			)
		}
		return "", err
	}

	// 检查响应是否为空
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("ai response choices empty")
	}

	// 提取回复内容并清理
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	return trimCodeBlock(content), nil
}

// isTimeoutError 检查错误是否是超时错误。
func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// appendNoThinkSuffix 在 prompt 末尾追加 "/no_think" 后缀。
//
// 这是 DashScope 的一个特殊指令，告诉 AI 不要输出思考过程，
// 直接输出最终结果。这可以减少 token 消耗并加快响应速度。
func appendNoThinkSuffix(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasSuffix(trimmed, "/no_think") {
		return trimmed
	}
	return trimmed + "\n\n/no_think"
}

// summarizeReply 将 AI 回复截断为前 120 个字符的摘要。
func summarizeReply(reply string) string {
	trimmed := strings.Join(strings.Fields(strings.TrimSpace(reply)), " ")
	runes := []rune(trimmed)
	if len(runes) <= 120 {
		return trimmed
	}
	return string(runes[:120]) + "..."
}

// trimCodeBlock 去除 AI 返回中可能存在的 Markdown 代码块包裹。
//
// AI 有时会返回这样的格式：
//
//	```json
//	[{"type": "single", ...}]
//	```
//
// 我们需要提取中间的 JSON 部分，去掉 ```json 和 ``` 标记。
func trimCodeBlock(content string) string {
	trim := strings.TrimSpace(content)
	if strings.HasPrefix(trim, "```json") {
		trim = strings.TrimPrefix(trim, "```json")
		trim = strings.TrimSpace(trim)
	}
	if strings.HasPrefix(trim, "```") {
		trim = strings.TrimPrefix(trim, "```")
		trim = strings.TrimSpace(trim)
	}
	if strings.HasSuffix(trim, "```") {
		trim = strings.TrimSuffix(trim, "```")
		trim = strings.TrimSpace(trim)
	}
	return trim
}
