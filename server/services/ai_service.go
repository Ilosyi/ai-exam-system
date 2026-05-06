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

	"github.com/joho/godotenv"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"week05/homework/server/config"
)

// AIService 封装了与外部 AI 服务（兼容 DashScope 的 API）的调用逻辑。
// 主要职责：
// - 从 .env 加载 API key
// - 读取本地 config（包含 BaseURL、模型名、system prompt 等）
// - 发送 HTTP 请求并返回原始文本结果（通常为 JSON 字符串）
type AIService struct {
	client openai.Client
	cfg    *config.AppConfig
}

type ConnectionDiagnostic struct {
	Model       string `json:"model"`
	BaseURL     string `json:"baseURL"`
	ElapsedMs   int64  `json:"elapsedMs"`
	Reply       string `json:"reply"`
	ReplyDigest string `json:"replyDigest"`
}

// NewAIService 从 projectRoot 加载 .env 与 config.json，构造 AIService
// 注意：如果 .env 中没有 DASHSCOPE_API_KEY，会返回错误，提示用户配置。
func NewAIService(projectRoot string) (*AIService, error) {
	// 使用 godotenv 仅在本地加载 .env 文件，生产环境推荐使用系统环境变量
	_ = godotenv.Load(filepath.Join(projectRoot, ".env"))
	key := os.Getenv("DASHSCOPE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("missing DASHSCOPE_API_KEY in .env")
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}
	return &AIService{
		client: openai.NewClient(
			option.WithAPIKey(key),
			option.WithBaseURL(cfg.BaseURL),
			option.WithMaxRetries(0),
		),
		cfg: cfg,
	}, nil
}

// GenerateQuestions 将 prompt 发送给外部 AI 接口，并返回 AI 回复的文本内容（通常为 JSON 字符串）
func (s *AIService) GenerateQuestions(prompt string) (string, error) {
	return s.runChatCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(s.cfg.SystemPrompt),
		openai.UserMessage(appendNoThinkSuffix(prompt)),
	})
}

func (s *AIService) TestConnection() (string, error) {
	return s.runChatCompletion([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("who are you"),
	})
}

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

func (s *AIService) runChatCompletion(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	timeout := time.Duration(s.cfg.RequestTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    shared.ChatModel(s.cfg.Model),
		},
		option.WithRequestTimeout(timeout),
	)
	if err != nil {
		if isTimeoutError(err) {
			return "", fmt.Errorf(
				"ai upstream timeout after %ds, request may have been accepted by provider and token may already be consumed",
				s.cfg.RequestTimeoutSeconds,
			)
		}
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("ai response choices empty")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	return trimCodeBlock(content), nil
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func appendNoThinkSuffix(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasSuffix(trimmed, "/no_think") {
		return trimmed
	}
	return trimmed + "\n\n/no_think"
}

func summarizeReply(reply string) string {
	trimmed := strings.Join(strings.Fields(strings.TrimSpace(reply)), " ")
	runes := []rune(trimmed)
	if len(runes) <= 120 {
		return trimmed
	}
	return string(runes[:120]) + "..."
}

// trimCodeBlock 用于去除 AI 返回中可能存在的 Markdown 代码块包裹（例如 ```json ... ```），方便后续 JSON 解析
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
