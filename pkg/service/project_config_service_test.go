package service

import (
	"fmt"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGitProjectRepository 用于测试的 mock repository
type MockGitProjectRepository struct {
	projects map[uint]*models.GitProject
}

func (m *MockGitProjectRepository) GetByID(id uint) (*models.GitProject, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("项目不存在")
}

func (m *MockGitProjectRepository) GetAll() ([]models.GitProject, error) {
	var result []models.GitProject
	for _, p := range m.projects {
		result = append(result, *p)
	}
	return result, nil
}

func (m *MockGitProjectRepository) Update(project *models.GitProject) error {
	m.projects[project.ID] = project
	return nil
}

func TestGetProjectAIConfig_UseDefault(t *testing.T) {
	project := &models.GitProject{
		ID:         1,
		Path:       "/test/project",
		Name:       "Test Project",
		UseDefault: true,
	}

	mockRepo := &MockGitProjectRepository{
		projects: map[uint]*models.GitProject{1: project},
	}

	cfg := &config.Config{
		Provider: "deepseek",
		Language: "chinese",
	}

	svc := NewProjectConfigService(mockRepo, cfg)

	result, err := svc.GetProjectAIConfig(1)
	require.NoError(t, err)
	assert.True(t, result.IsDefault)
	assert.Equal(t, "deepseek", result.Provider)
	assert.Equal(t, "chinese", result.Language)
}

func TestGetProjectAIConfig_UseCustom(t *testing.T) {
	provider := "openai"
	language := "english"

	project := &models.GitProject{
		ID:         1,
		Path:       "/test/project",
		Name:       "Test Project",
		UseDefault: false,
		Provider:   &provider,
		Language:   &language,
	}

	mockRepo := &MockGitProjectRepository{
		projects: map[uint]*models.GitProject{1: project},
	}

	cfg := &config.Config{
		Provider: "deepseek",
		Language: "chinese",
	}

	svc := NewProjectConfigService(mockRepo, cfg)

	result, err := svc.GetProjectAIConfig(1)
	require.NoError(t, err)
	assert.False(t, result.IsDefault)
	assert.Equal(t, "openai", result.Provider)
	assert.Equal(t, "english", result.Language)
}

func TestValidateProjectConfig_InvalidProvider(t *testing.T) {
	provider := "invalid-provider"

	project := &models.GitProject{
		ID:         1,
		Path:       "/test/project",
		Name:       "Test Project",
		UseDefault: false,
		Provider:   &provider,
	}

	mockRepo := &MockGitProjectRepository{
		projects: map[uint]*models.GitProject{1: project},
	}

	cfg := &config.Config{
		Provider: "deepseek",
		Providers: map[string]config.ProviderSettings{
			"deepseek": {},
		},
	}

	svc := NewProjectConfigService(mockRepo, cfg)

	valid, resetFields, suggested, err := svc.ValidateProjectConfig(1)
	require.NoError(t, err)
	assert.False(t, valid)
	assert.Contains(t, resetFields, "provider")
	assert.NotNil(t, suggested)
	assert.Equal(t, "deepseek", suggested.Provider)
}

func TestResetProjectToDefaults(t *testing.T) {
	provider := "openai"

	project := &models.GitProject{
		ID:         1,
		Path:       "/test/project",
		Name:       "Test Project",
		UseDefault: false,
		Provider:   &provider,
	}

	mockRepo := &MockGitProjectRepository{
		projects: map[uint]*models.GitProject{1: project},
	}

	cfg := &config.Config{}
	svc := NewProjectConfigService(mockRepo, cfg)

	err := svc.ResetProjectToDefaults(1)
	require.NoError(t, err)

	// 验证重置后的状态
	updated := mockRepo.projects[1]
	assert.True(t, updated.UseDefault)
	assert.Nil(t, updated.Provider)
	assert.Nil(t, updated.Language)
}
