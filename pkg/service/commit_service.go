package service

import (
	"context"
	"fmt"
	"os"

	"github.com/allanpk716/ai-commit-hub/pkg/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/git"
	"github.com/allanpk716/ai-commit-hub/pkg/prompt"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
	aicommitconfig "github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type CommitService struct {
	ctx            context.Context
	configService  *ConfigService
}

func NewCommitService(ctx context.Context) *CommitService {
	return &CommitService{
		ctx:           ctx,
		configService: NewConfigService(),
	}
}

func (s *CommitService) GenerateCommit(projectPath, providerName, language string) error {
	// Load config
	cfg, _ := config.LoadOrCreateConfig()

	// Override provider if specified
	if providerName != "" {
		cfg.Provider = providerName
	}
	if language != "" {
		cfg.Language = language
	}

	// 加载配置检查 provider 是否已配置
	configuredCfg, err := s.configService.LoadConfig(s.ctx)
	if err != nil {
		runtime.EventsEmit(s.ctx, "commit-error", fmt.Sprintf("加载配置失败: %v", err))
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 检查 provider 是否已配置
	providers := s.configService.GetConfiguredProviders(configuredCfg)
	providerConfigured := false
	for _, p := range providers {
		if p.Name == cfg.Provider && p.Configured {
			providerConfigured = true
			break
		}
	}

	if !providerConfigured {
		errMsg := fmt.Sprintf("Provider '%s' 未配置，请先在配置文件中添加", cfg.Provider)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("provider not configured: %s", cfg.Provider)
	}

	// Get AI client from registry (imports provider packages for side effects)
	// The providers are already registered via their init() functions
	factory, ok := registry.Get(cfg.Provider)
	if !ok {
		return fmt.Errorf("未知的 provider: %s", cfg.Provider)
	}

	// Convert our config.ProviderSettings to ai-commit's config.ProviderSettings
	providerSettings := cfg.Providers[cfg.Provider]
	ps := aicommitconfig.ProviderSettings{
		APIKey:  providerSettings.APIKey,
		Model:   providerSettings.Model,
		BaseURL: providerSettings.BaseURL,
	}

	client, err := factory(context.Background(), cfg.Provider, ps)
	if err != nil {
		return fmt.Errorf("创建 AI client 失败: %w", err)
	}

	// Get diff
	originalDir, _ := os.Getwd()
	os.Chdir(projectPath)
	defer os.Chdir(originalDir)

	diff, err := git.GetGitDiffIgnoringMoves(context.Background())
	if err != nil {
		return fmt.Errorf("获取 diff 失败: %w", err)
	}

	if diff == "" {
		runtime.EventsEmit(s.ctx, "commit-error", "暂存区没有变更")
		return nil
	}

	// Build prompt
	promptText := prompt.BuildCommitPrompt(diff, cfg.Language, "", "", "")

	// Stream commit message
	if sc, ok := client.(ai.StreamingAIClient); ok {
		go func() {
			final, err := sc.StreamCommitMessage(context.Background(), promptText, func(delta string) {
				runtime.EventsEmit(s.ctx, "commit-delta", delta)
			})

			if err != nil {
				runtime.EventsEmit(s.ctx, "commit-error", err.Error())
			} else {
				runtime.EventsEmit(s.ctx, "commit-complete", final)
			}
		}()
		return nil
	}

	// Fallback: non-streaming
	msg, err := client.GetCommitMessage(context.Background(), promptText)
	if err != nil {
		return err
	}

	runtime.EventsEmit(s.ctx, "commit-complete", msg)
	return nil
}

// SaveHistory is a placeholder for history saving functionality
// History saving is handled at the App layer via SaveCommitHistory API
func (s *CommitService) SaveHistory(projectID uint, message, provider, language string) error {
	// Placeholder - actual history saving happens via App.SaveCommitHistory
	return nil
}
