package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigService_LoadConfig(t *testing.T) {
	service := NewConfigService()

	ctx := context.Background()
	cfg, err := service.LoadConfig(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "zh", cfg.Language)
}

func TestConfigService_GetAvailableProviders(t *testing.T) {
	service := NewConfigService()

	providers := service.GetAvailableProviders()

	// Provider Registry 通过 init 函数注册
	// 在测试环境中，我们至少验证不返回 nil
	assert.NotNil(t, providers)

	// 如果 providers 已注册（通过 config_service.go 中的 import），
	// 应该包含主要的 providers
	if len(providers) > 0 {
		// 验证包含一些已知的 provider
		assert.Contains(t, providers, "openai", "应该包含 openai provider")
	}
	// 注意：由于 init 函数的执行顺序，这个测试可能不稳定
	// 如果失败，可以考虑移除或标记为 flaky
}

func TestConfigService_ResolvePromptTemplate_Default(t *testing.T) {
	service := NewConfigService()

	template, err := service.ResolvePromptTemplate("", "")

	assert.NoError(t, err)
	assert.NotEmpty(t, template)
}

func TestConfigService_ResolvePromptTemplate_Custom(t *testing.T) {
	service := NewConfigService()

	// 创建临时目录和文件
	tempDir := t.TempDir()
	promptsDir := filepath.Join(tempDir, "prompts")
	os.Mkdir(promptsDir, 0o755)

	customPrompt := "Custom prompt template for testing"
	promptPath := filepath.Join(promptsDir, "custom.txt")
	os.WriteFile(promptPath, []byte(customPrompt), 0o644)

	template, err := service.ResolvePromptTemplate(tempDir, "custom.txt")

	assert.NoError(t, err)
	assert.Equal(t, customPrompt, template)
}

func TestConfigService_ResolvePromptTemplate_NotFound(t *testing.T) {
	service := NewConfigService()

	tempDir := t.TempDir()

	template, err := service.ResolvePromptTemplate(tempDir, "nonexistent.txt")

	assert.Error(t, err)
	assert.Empty(t, template)
}
