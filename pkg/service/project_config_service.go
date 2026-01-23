package service

import (
	"fmt"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// ProjectAIConfig 表示项目的 AI 配置
type ProjectAIConfig struct {
	Provider  string
	Language  string
	Model     string
	IsDefault bool // 是否使用默认配置
}

// GitProjectRepositoryInterface 定义项目存储库接口
type GitProjectRepositoryInterface interface {
	GetByID(id uint) (*models.GitProject, error)
	GetAll() ([]models.GitProject, error)
	Update(project *models.GitProject) error
}

// ProjectConfigService 管理项目级别的 AI 配置
type ProjectConfigService struct {
	projectRepo GitProjectRepositoryInterface
	config      *config.Config
}

// NewProjectConfigService 创建项目配置服务
func NewProjectConfigService(repo GitProjectRepositoryInterface, cfg *config.Config) *ProjectConfigService {
	return &ProjectConfigService{
		projectRepo: repo,
		config:      cfg,
	}
}

// GetProjectAIConfig 获取项目的有效 AI 配置
func (s *ProjectConfigService) GetProjectAIConfig(projectID uint) (*ProjectAIConfig, error) {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}

	result := &ProjectAIConfig{}

	// 获取默认值
	defaultProvider := s.config.Provider
	if defaultProvider == "" {
		defaultProvider = config.DefaultProvider
		logger.Warnf("配置文件中 Provider 为空，使用默认值: %s", defaultProvider)
	}

	defaultLanguage := s.config.Language
	if defaultLanguage == "" {
		defaultLanguage = "english"
		logger.Warnf("配置文件中 Language 为空，使用默认值: %s", defaultLanguage)
	}

	// 检查是否使用默认配置
	if project.UseDefault || (project.Provider == nil && project.Language == nil) {
		result.Provider = defaultProvider
		result.Language = defaultLanguage
		result.IsDefault = true
	} else {
		// 使用数据库中的配置
		if project.Provider != nil {
			result.Provider = *project.Provider
		} else {
			result.Provider = defaultProvider
		}

		if project.Language != nil {
			result.Language = *project.Language
		} else {
			result.Language = defaultLanguage
		}

		if project.Model != nil {
			result.Model = *project.Model
		}

		result.IsDefault = false
	}

	return result, nil
}

// isKnownProvider 检查是否是已知的 Provider
func isKnownProvider(provider string) bool {
	knownProviders := []string{"openai", "anthropic", "deepseek", "ollama", "google", "phind"}
	for _, p := range knownProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// ValidateProjectConfig 验证项目配置是否与配置文件一致
func (s *ProjectConfigService) ValidateProjectConfig(projectID uint) (valid bool, resetFields []string, suggestedConfig *ProjectAIConfig, err error) {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return false, nil, nil, err
	}

	// 如果使用默认配置，始终有效
	if project.UseDefault {
		return true, nil, nil, nil
	}

	var needsReset []string

	// 检查 Provider 是否存在
	if project.Provider != nil {
		provider := *project.Provider
		// 检查是否在配置文件的 Providers 中
		if _, exists := s.config.Providers[provider]; !exists {
			// 检查是否是已知的 Provider
			if !isKnownProvider(provider) {
				needsReset = append(needsReset, "provider")
			}
		}
	}

	// 检查 Language 是否有效
	if project.Language != nil {
		lang := *project.Language
		if lang != "zh" && lang != "en" && lang != "chinese" && lang != "english" {
			needsReset = append(needsReset, "language")
		}
	}

	if len(needsReset) > 0 {
		// 生成建议的默认配置
		suggestedConfig = &ProjectAIConfig{
			Provider:  s.config.Provider,
			Language:  s.config.Language,
			IsDefault: true,
		}
		if suggestedConfig.Provider == "" {
			suggestedConfig.Provider = config.DefaultProvider
		}
		if suggestedConfig.Language == "" {
			suggestedConfig.Language = "english"
		}

		return false, needsReset, suggestedConfig, nil
	}

	return true, nil, nil, nil
}

// ResetProjectToDefaults 将项目配置重置为默认值
func (s *ProjectConfigService) ResetProjectToDefaults(projectID uint) error {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}

	project.UseDefault = true
	project.Provider = nil
	project.Language = nil
	project.Model = nil

	return s.projectRepo.Update(project)
}
