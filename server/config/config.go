package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig stores AI configuration loaded from config.json at week05/homework root.
type AppConfig struct {
	BaseURL               string   `json:"baseURL"`
	Model                 string   `json:"model"`
	RequestTimeoutSeconds int      `json:"requestTimeoutSeconds"`
	SystemPrompt          string   `json:"systemPrompt"`
	SystemPromptLines     []string `json:"systemPromptLines"`
	ServerPort            int      `json:"serverPort"`
	ClientPort            int      `json:"clientPort"`
}

var cfg *AppConfig

// Load reads config.json located at projectRoot (week05/homework).
func Load(projectRoot string) (*AppConfig, error) {
	if cfg != nil {
		return cfg, nil
	}

	configPath := filepath.Join(projectRoot, "config.json")
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config.json failed: %w", err)
	}

	var tmp AppConfig
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return nil, fmt.Errorf("parse config.json failed: %w", err)
	}

	if tmp.SystemPrompt == "" && len(tmp.SystemPromptLines) > 0 {
		tmp.SystemPrompt = joinLines(tmp.SystemPromptLines)
	}
	if tmp.RequestTimeoutSeconds <= 0 {
		tmp.RequestTimeoutSeconds = 120
	}
	if tmp.ServerPort <= 0 {
		tmp.ServerPort = 8080
	}
	if tmp.ClientPort <= 0 {
		tmp.ClientPort = 3000
	}

	if tmp.BaseURL == "" || tmp.Model == "" || tmp.SystemPrompt == "" {
		return nil, fmt.Errorf("config.json missing required fields")
	}

	cfg = &tmp
	return cfg, nil
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
