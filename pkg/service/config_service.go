package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/prompt"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/anthropic"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/deepseek"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/ollama"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/openai"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) LoadConfig(ctx context.Context) (*config.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Load or create default config
	cfg := &config.Config{
		Provider:  "openai",
		Language:  "zh",
		Providers: make(map[string]config.ProviderSettings),
	}

	if _, err := os.Stat(configPath); err == nil {
		data, _ := os.ReadFile(configPath)
		// Simple YAML parse (use gopkg.in/yaml.v3)
		// For now, return default if parse fails
		_ = data // TODO: Implement YAML parsing
	}

	return cfg, nil
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
