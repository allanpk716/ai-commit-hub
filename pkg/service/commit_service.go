package service

import (
	"context"
	"fmt"
	"os"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/ai"
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
	logger.Info("开始生成 Commit 消息")
	logger.Infof("项目路径: %s", projectPath)
	logger.Infof("请求的 Provider: %s", providerName)
	logger.Infof("请求的语言: %s", language)

	// 加载配置检查 provider 是否已配置
	logger.Info("正在加载配置...")
	cfg, err := s.configService.LoadConfig(s.ctx)
	if err != nil {
		errMsg := fmt.Sprintf("加载配置失败: %v", err)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("加载配置失败: %w", err)
	}
	logger.Info("配置加载成功")
	logger.Infof("当前默认 Provider: %s", cfg.Provider)

	// Override provider if specified
	if providerName != "" {
		cfg.Provider = providerName
		logger.Infof("使用指定的 Provider: %s", providerName)
	}
	if language != "" {
		cfg.Language = language
		logger.Infof("使用指定的语言: %s", language)
	}

	// 检查 provider 是否已配置
	logger.Info("检查 Provider 配置状态...")
	providers := s.configService.GetConfiguredProviders(cfg)
	logger.Infof("找到 %d 个已配置的 Provider", len(providers))

	providerConfigured := false
	for _, p := range providers {
		logger.Debugf("Provider: %s, 配置状态: %v", p.Name, p.Configured)
		if p.Name == cfg.Provider && p.Configured {
			providerConfigured = true
			logger.Infof("找到已配置的 Provider: %s", cfg.Provider)
			break
		}
	}

	if !providerConfigured {
		errMsg := fmt.Sprintf("Provider '%s' 未配置或不可用，请先在设置中配置 API Key", cfg.Provider)
		logger.Error(errMsg)
		logger.Infof("可用的 Provider: %v", providers)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("provider not configured: %s", cfg.Provider)
	}

	// Get AI client from registry (imports provider packages for side effects)
	// The providers are already registered via their init() functions
	logger.Infof("从注册表获取 Provider: %s", cfg.Provider)
	factory, ok := registry.Get(cfg.Provider)
	if !ok {
		errMsg := fmt.Sprintf("未知的 provider: %s", cfg.Provider)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("未知的 provider: %s", cfg.Provider)
	}

	// Convert our config.ProviderSettings to ai-commit's config.ProviderSettings
	providerSettings := cfg.Providers[cfg.Provider]
	ps := aicommitconfig.ProviderSettings{
		APIKey:  providerSettings.APIKey,
		Model:   providerSettings.Model,
		BaseURL: providerSettings.BaseURL,
	}
	logger.Infof("Provider 配置 - Model: %s, BaseURL: %s", providerSettings.Model, providerSettings.BaseURL)

	logger.Info("创建 AI Client...")
	client, err := factory(context.Background(), cfg.Provider, ps)
	if err != nil {
		errMsg := fmt.Sprintf("创建 AI client 失败: %v", err)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("创建 AI client 失败: %w", err)
	}
	logger.Info("AI Client 创建成功")

	// Get diff - 使用 GetStagedDiff 读取暂存区变更（匹配 ai-commit 项目行为）
	logger.Info("获取暂存区 Diff（使用 git diff --cached）...")
	originalDir, _ := os.Getwd()
	err = os.Chdir(projectPath)
	if err != nil {
		errMsg := fmt.Sprintf("切换到项目目录失败: %v", err)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("切换目录失败: %w", err)
	}
	defer os.Chdir(originalDir)

	diff, err := git.GetStagedDiff(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("获取暂存区 diff 失败: %v", err)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return fmt.Errorf("获取暂存区 diff 失败: %w", err)
	}
	logger.Infof("暂存区 Diff 获取成功，长度: %d 字符", len(diff))

	if diff == "" {
		errMsg := "暂存区没有变更"
		logger.Warn(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return nil
	}

	// Build prompt
	logger.Info("构建 Prompt...")
	promptText := prompt.BuildCommitPrompt(diff, cfg.Language, "", "", "")
	logger.Debugf("Prompt 长度: %d 字符", len(promptText))

	// Stream commit message
	if sc, ok := client.(ai.StreamingAIClient); ok {
		logger.Info("使用流式生成模式")
		go func() {
			logger.Info("开始流式生成...")
			final, err := sc.StreamCommitMessage(context.Background(), promptText, func(delta string) {
				runtime.EventsEmit(s.ctx, "commit-delta", delta)
			})

			if err != nil {
				errMsg := fmt.Sprintf("生成失败: %v", err)
				logger.Error(errMsg)
				runtime.EventsEmit(s.ctx, "commit-error", errMsg)
			} else {
				logger.Info("Commit 消息生成成功")
				runtime.EventsEmit(s.ctx, "commit-complete", final)
			}
		}()
		return nil
	}

	// Fallback: non-streaming
	logger.Info("使用非流式生成模式")
	msg, err := client.GetCommitMessage(context.Background(), promptText)
	if err != nil {
		errMsg := fmt.Sprintf("生成失败: %v", err)
		logger.Error(errMsg)
		runtime.EventsEmit(s.ctx, "commit-error", errMsg)
		return err
	}

	logger.Info("Commit 消息生成成功")
	runtime.EventsEmit(s.ctx, "commit-complete", msg)
	return nil
}

// SaveHistory is a placeholder for history saving functionality
// History saving is handled at the App layer via SaveCommitHistory API
func (s *CommitService) SaveHistory(projectID uint, message, provider, language string) error {
	// Placeholder - actual history saving happens via App.SaveCommitHistory
	return nil
}
