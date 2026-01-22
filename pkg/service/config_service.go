package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
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
		// Return default prompt from config package
		return getDefaultCommitPrompt(), nil
	}

	promptPath := filepath.Join(configDir, "prompts", configFile)
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// getDefaultCommitPrompt returns the default commit message prompt template
// This is a simplified version - the full version is in config package
func getDefaultCommitPrompt() string {
	return `You are an expert software engineer.
Analyze the provided git diff and generate a commit message following the Conventional Commits specification.

### 1. INPUT
- **Diff Data**: See below.
- **Target Language**: {LANGUAGE}

### 2. RULES
- **Intent**: Focus on *why* the change was made.
- **Noise**: Ignore pure formatting changes or lockfiles unless specific intent exists.
- **Types**: Keep standard types (feat, fix, chore, etc.) in English.

### 3. OUTPUT FORMAT (Strictly Follow)
You must generate the message in the structure below, using **{LANGUAGE}** for the description and body:

<type>(<scope>): <concise description in {LANGUAGE}>

[Optional Body in {LANGUAGE}, bullet points]

### 4. EXAMPLES (For format reference only - Translate your output to {LANGUAGE})

Input: (Logic change)
Output: fix(auth): resolve nil pointer in token validation
*(If Language is Chinese, you should output: fix(auth): 修复令牌验证中的空指针问题)*

Input: (Ignore file)
Output: chore(gitignore): ignore .DS_Store files

---

### 5. FINAL INSTRUCTION
Analyze the diff below and write the commit message.
**CRITICAL**: The <description> and <body> MUST be in **{LANGUAGE}**.

### DIFF TO ANALYZE:
{DIFF}`
}
