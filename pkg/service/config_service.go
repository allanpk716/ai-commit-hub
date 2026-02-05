package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/prompt"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/anthropic"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/deepseek"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/google"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/ollama"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/openai"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/openrouter"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/phind"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
	"gopkg.in/yaml.v3"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) LoadConfig(ctx context.Context) (*config.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		defaultCfg := s.getDefaultConfig()
		if err := s.saveConfig(configPath, defaultCfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultCfg, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func (s *ConfigService) getDefaultConfig() *config.Config {
	return &config.Config{
		Provider:         "openai",
		Language:         "zh",
		AuthorName:       config.DefaultAuthorName,
		AuthorEmail:      config.DefaultAuthorEmail,
		EnableEmoji:      true,
		SemanticRelease:  false,
		InteractiveSplit: false,
		LockFiles:        []string{"go.mod", "go.sum", "package-lock.json", "yarn.lock"},
		CommitTypes: []config.CommitTypeConfig{
			{Type: "feat", Emoji: "✨"},
			{Type: "fix", Emoji: "🐛"},
			{Type: "docs", Emoji: "📚"},
			{Type: "style", Emoji: "💎"},
			{Type: "refactor", Emoji: "♻️"},
			{Type: "test", Emoji: "🧪"},
			{Type: "chore", Emoji: "🔧"},
			{Type: "perf", Emoji: "🚀"},
			{Type: "build", Emoji: "📦"},
			{Type: "ci", Emoji: "👷"},
		},
		Limits: config.Limits{
			Diff:   config.LimitSettings{Enabled: false, MaxChars: 0},
			Prompt: config.LimitSettings{Enabled: false, MaxChars: 0},
		},
		Providers: map[string]config.ProviderSettings{
			"openai": {
				APIKey:  "",
				Model:   "gpt-4",
				BaseURL: "https://api.openai.com/v1",
			},
			"anthropic": {
				APIKey:  "",
				Model:   "claude-3-opus-20240229",
				BaseURL: "https://api.anthropic.com",
			},
			"deepseek": {
				APIKey:  "",
				Model:   "deepseek-chat",
				BaseURL: "https://api.deepseek.com",
			},
			"ollama": {
				Model:   "llama2",
				BaseURL: "http://localhost:11434",
			},
			"google": {
				APIKey:  "",
				Model:   "gemini-pro",
				BaseURL: "https://generativelanguage.googleapis.com",
			},
			"openrouter": {
				APIKey:  "",
				Model:   "mistralai/mistral-7b-instruct",
				BaseURL: "https://openrouter.ai/api/v1",
			},
			"phind": {
				APIKey:  "",
				Model:   "phind-34b-v2",
				BaseURL: "",
			},
		},
		Prompts: config.PromptFiles{
			CommitMessage: "commit-message.txt",
			CodeReview:    "code-review.txt",
			StyleReview:   "style-review.txt",
		},
	}
}

func (s *ConfigService) saveConfig(path string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *ConfigService) GetAvailableProviders() []string {
	return registry.Names()
}

func (s *ConfigService) ResolvePromptTemplate(configDir, configFile string) (string, error) {
	if configFile == "" {
		// Return default prompt from prompt package
		return prompt.DefaultPromptTemplate, nil
	}

	promptPath := filepath.Join(configDir, "prompts", configFile)
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (s *ConfigService) GetConfiguredProviders(cfg *config.Config) []models.ProviderInfo {
	// 获取所有已注册的 providers
	registeredProviders := registry.Names()

	result := make([]models.ProviderInfo, 0, len(registeredProviders))

	for _, name := range registeredProviders {
		info := models.ProviderInfo{
			Name: name,
		}

		// 检查该 provider 是否在 config 中配置
		if cfg.Providers == nil {
			info.Configured = false
			info.Reason = "未在配置文件中添加"
			result = append(result, info)
			continue
		}

		providerSettings, exists := cfg.Providers[name]
		if !exists {
			info.Configured = false
			info.Reason = "未在配置文件中添加"
			result = append(result, info)
			continue
		}

		// 检查是否需要 API Key
		requiresKey := registry.RequiresAPIKey(name)

		// 验证配置完整性
		var reason string
		configured := true

		if requiresKey && providerSettings.APIKey == "" {
			configured = false
			reason = "缺少 API Key"
		}

		info.Configured = configured
		if !configured {
			info.Reason = reason
		}

		result = append(result, info)
	}

	return result
}
