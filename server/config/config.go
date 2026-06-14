// ============================================================================
// config/config.go - 应用配置管理
// ============================================================================
//
// 本文件负责从 config.json 加载应用配置，并提供合理的默认值。
//
// 配置文件结构 (config.json)：
// {
//   "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",  // AI 服务地址
//   "model": "qwen-plus",                                            // AI 模型名称
//   "requestTimeoutSeconds": 120,                                     // AI 请求超时时间（秒）
//   "systemPromptLines": ["你是一个..."],                              // AI 系统提示词（多行）
//   "serverPort": 8080,                                               // 后端服务端口
//   "clientPort": 3000                                                // 前端开发服务端口
// }
//
// 设计要点：
// - 使用单例模式（包级变量 cfg），配置只加载一次
// - 提供合理的默认值，减少配置项
// - 必填字段缺失时返回明确的错误信息
//
// 学习要点：
// - Go 的 JSON 反序列化 (json.Unmarshal)
// - 单例模式的简单实现
// - 配置文件的最佳实践
// ============================================================================

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig 存储应用的所有配置项。
//
// JSON 标签 (json:"xxx") 指定了 JSON 字段名与结构体字段的映射关系。
// 例如 `json:"baseURL"` 表示 JSON 中的 "baseURL" 字段映射到这个 Go 字段。
type AppConfig struct {
	BaseURL               string   `json:"baseURL"`               // AI 服务的 API 地址（如 DashScope 兼容接口）
	Model                 string   `json:"model"`                 // AI 模型名称（如 qwen-plus）
	RequestTimeoutSeconds int      `json:"requestTimeoutSeconds"` // AI 请求超时时间（秒），默认 120
	SystemPrompt          string   `json:"systemPrompt"`          // AI 系统提示词（单字符串形式）
	SystemPromptLines     []string `json:"systemPromptLines"`     // AI 系统提示词（多行数组形式，会自动拼接）
	ServerPort            int      `json:"serverPort"`            // 后端服务监听端口，默认 8080
	ClientPort            int      `json:"clientPort"`            // 前端开发服务端口，默认 3000（仅开发时使用）
}

// cfg 是全局单例，保存已加载的配置。
// 使用包级变量 + nil 检查实现简单的单例模式。
var cfg *AppConfig

// Load 从项目根目录加载 config.json 配置文件。
//
// 这是一个单例函数：第一次调用时加载配置，后续调用直接返回缓存的配置。
// 这种设计避免了重复读取文件和解析 JSON 的开销。
//
// 参数：
//   - projectRoot: 项目根目录的绝对路径
//
// 返回值：
//   - *AppConfig: 配置对象指针
//   - error: 加载失败时返回错误（如文件不存在、JSON 格式错误、必填字段缺失）
//
// 加载流程：
// 1. 检查是否已加载（缓存命中）
// 2. 读取 config.json 文件内容
// 3. 反序列化 JSON 为 AppConfig 结构体
// 4. 处理多行提示词拼接
// 5. 填充默认值
// 6. 校验必填字段
func Load(projectRoot string) (*AppConfig, error) {
	// 单例检查：如果已经加载过，直接返回缓存的配置
	if cfg != nil {
		return cfg, nil
	}

	// 拼接配置文件的完整路径
	configPath := filepath.Join(projectRoot, "config.json")

	// 读取文件内容
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config.json failed: %w", err)
	}

	// 将 JSON 字节反序列化为 AppConfig 结构体
	var tmp AppConfig
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return nil, fmt.Errorf("parse config.json failed: %w", err)
	}

	// 处理多行提示词：如果 SystemPrompt 为空但 SystemPromptLines 有内容，
	// 则将多行拼接为单个字符串
	if tmp.SystemPrompt == "" && len(tmp.SystemPromptLines) > 0 {
		tmp.SystemPrompt = joinLines(tmp.SystemPromptLines)
	}

	// 填充默认值（如果配置文件中没有指定）
	if tmp.RequestTimeoutSeconds <= 0 {
		tmp.RequestTimeoutSeconds = 120 // 默认 120 秒超时
	}
	if tmp.ServerPort <= 0 {
		tmp.ServerPort = 8080 // 默认后端端口
	}
	if tmp.ClientPort <= 0 {
		tmp.ClientPort = 3000 // 默认前端端口
	}

	// 校验必填字段
	if tmp.BaseURL == "" || tmp.Model == "" || tmp.SystemPrompt == "" {
		return nil, fmt.Errorf("config.json missing required fields")
	}

	// 缓存配置（单例）
	cfg = &tmp
	return cfg, nil
}

// joinLines 将多行字符串用换行符拼接为单个字符串。
//
// 例如：["第一行", "第二行"] → "第一行\n第二行"
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
