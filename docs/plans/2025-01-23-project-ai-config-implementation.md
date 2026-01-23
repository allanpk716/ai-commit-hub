# 项目级别 AI 配置功能实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现为不同 Git 项目配置独立 AI Provider 和语言设置的功能，支持默认值继承、配置验证和自动重置。

**Architecture:** 数据库存储项目配置（优先），配置文件作为全局默认值（兜底）。切换项目时验证配置一致性，无效时提示用户确认重置。

**Tech Stack:** Go 1.21+, GORM, Vue 3, TypeScript, Pinia, Wails v2, SQLite

---

## Task 1: 数据库模型扩展

**Files:**
- Modify: `pkg/models/git_project.go`
- Create: `pkg/repository/migration.go`

### Step 1: 扩展 GitProject 结构体

在 `pkg/models/git_project.go` 中添加 AI 配置字段：

```go
// pkg/models/git_project.go
type GitProject struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Path      string `gorm:"not null;uniqueIndex" json:"path"`
	Name      string `json:"name"`
	SortOrder int    `gorm:"index" json:"sort_order"`

	// 项目级别 AI 配置（可选）
	Provider   *string `json:"provider,omitempty"`    // nil 表示使用默认
	Language   *string `json:"language,omitempty"`    // nil 表示使用默认
	Model      *string `json:"model,omitempty"`       // nil 表示使用默认
	UseDefault bool    `gorm:"default:true" json:"use_default"` // true=使用默认配置
}

// TableName specifies the table name for GitProject
func (GitProject) TableName() string {
	return "git_projects"
}
```

### Step 2: 创建数据库迁移文件

创建 `pkg/repository/migration.go`：

```go
// pkg/repository/migration.go
package repository

import (
	"fmt"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"gorm.io/gorm"
)

// MigrateAddProjectAIConfig 添加项目 AI 配置字段的迁移
func MigrateAddProjectAIConfig(db *gorm.DB) error {
	logger.Info("开始迁移：添加项目 AI 配置字段")

	// AutoMigrate 会自动添加新字段
	if err := db.AutoMigrate(&models.GitProject{}); err != nil {
		return fmt.Errorf("AutoMigrate 失败: %w", err)
	}

	// 将现有项目标记为使用默认配置
	result := db.Model(&models.GitProject{}).
		Where("use_default IS NULL OR use_default = false").
		Update("use_default", true)

	if result.Error != nil {
		return fmt.Errorf("更新现有项目失败: %w", result.Error)
	}

	logger.Infof("迁移完成：已更新 %d 个项目", result.RowsAffected)
	return nil
}
```

### Step 3: 运行迁移验证

运行: `wails dev` 或直接运行应用
Expected: 数据库自动迁移，新字段添加成功

### Step 4: 提交

```bash
git add pkg/models/git_project.go pkg/repository/migration.go
git commit -m "feat: 添加项目 AI 配置数据库字段"
```

---

## Task 2: 项目配置服务

**Files:**
- Create: `pkg/service/project_config_service.go`
- Create: `pkg/service/project_config_service_test.go`

### Step 1: 创建项目配置服务

创建 `pkg/service/project_config_service.go`：

```go
// pkg/service/project_config_service.go
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
```

### Step 2: 编写单元测试

创建 `pkg/service/project_config_service_test.go`：

```go
// pkg/service/project_config_service_test.go
package service_test

import (
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/service"
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

	svc := service.NewProjectConfigService(mockRepo, cfg)

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

	svc := service.NewProjectConfigService(mockRepo, cfg)

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

	svc := service.NewProjectConfigService(mockRepo, cfg)

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
	svc := service.NewProjectConfigService(mockRepo, cfg)

	err := svc.ResetProjectToDefaults(1)
	require.NoError(t, err)

	// 验证重置后的状态
	updated := mockRepo.projects[1]
	assert.True(t, updated.UseDefault)
	assert.Nil(t, updated.Provider)
	assert.Nil(t, updated.Language)
}
```

### Step 3: 运行测试

```bash
go test ./pkg/service -v -run TestGetProjectAIConfig
```

Expected: 所有测试通过

### Step 4: 提交

```bash
git add pkg/service/project_config_service.go pkg/service/project_config_service_test.go
git commit -m "feat: 添加项目配置服务和单元测试"
```

---

## Task 3: App API 方法

**Files:**
- Modify: `app.go`

### Step 1: 添加 API 方法

在 `app.go` 中添加新的方法。首先找到 `App` 结构体，添加新字段：

```go
// app.go
type App struct {
	ctx              context.Context
	config           *config.Config
	// ... 现有字段 ...
	projectConfigService *service.ProjectConfigService
}
```

在 `startup()` 方法中初始化服务：

```go
func (a *App) startup(ctx context.Context) error {
	// ... 现有初始化代码 ...

	// 初始化项目配置服务
	a.projectConfigService = service.NewProjectConfigService(a.projectRepo, a.config)

	return nil
}
```

添加新的 API 方法：

```go
// app.go

// GetProjectAIConfig 获取项目的 AI 配置
func (a *App) GetProjectAIConfig(projectID int) (*service.ProjectAIConfig, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	config, err := a.projectConfigService.GetProjectAIConfig(uint(projectID))
	if err != nil {
		logger.Errorf("获取项目 AI 配置失败: %v", err)
		return nil, err
	}

	return config, nil
}

// UpdateProjectAIConfig 更新项目的 AI 配置
func (a *App) UpdateProjectAIConfig(projectID int, provider, language, model string, useDefault bool) error {
	if a.initError != nil {
		return a.initError
	}

	project, err := a.projectRepo.GetByID(uint(projectID))
	if err != nil {
		return err
	}

	project.UseDefault = useDefault

	if useDefault {
		project.Provider = nil
		project.Language = nil
		project.Model = nil
	} else {
		if provider != "" {
			project.Provider = &provider
		}
		if language != "" {
			project.Language = &language
		}
		if model != "" {
			project.Model = &model
		}
	}

	if err := a.projectRepo.Update(project); err != nil {
		logger.Errorf("更新项目配置失败: %v", err)
		return err
	}

	return nil
}

// ValidateProjectConfig 验证项目配置
func (a *App) ValidateProjectConfig(projectID int) (valid bool, resetFields []string, suggestedConfig map[string]interface{}, err error) {
	if a.initError != nil {
		return false, nil, nil, a.initError
	}

	valid, fields, config, err := a.projectConfigService.ValidateProjectConfig(uint(projectID))
	if err != nil {
		return false, nil, nil, err
	}

	if config != nil {
		suggestedConfig = map[string]interface{}{
			"provider":  config.Provider,
			"language":  config.Language,
			"isDefault": config.IsDefault,
		}
	}

	return valid, fields, suggestedConfig, nil
}

// ConfirmResetProjectConfig 确认并重置项目配置
func (a *App) ConfirmResetProjectConfig(projectID int) error {
	if a.initError != nil {
		return a.initError
	}

	if err := a.projectConfigService.ResetProjectToDefaults(uint(projectID)); err != nil {
		logger.Errorf("重置项目配置失败: %v", err)
		return err
	}

	return nil
}
```

### Step 2: 运行应用验证

```bash
wails dev
```

Expected: 应用正常启动，无编译错误

### Step 3: 提交

```bash
git add app.go
git commit -m "feat: 添加项目 AI 配置 API 方法"
```

---

## Task 4: 前端类型定义

**Files:**
- Modify: `frontend/src/types/index.ts`

### Step 1: 更新 GitProject 接口

```typescript
// frontend/src/types/index.ts

export interface GitProject {
  id: number
  path: string
  name: string
  sort_order: number
  created_at?: string
  updated_at?: string

  // 项目 AI 配置（可选）
  provider?: string | null      // null 表示使用默认
  language?: string | null      // null 表示使用默认
  model?: string | null         // null 表示使用默认
  use_default?: boolean         // true 表示使用默认配置
}

export interface ProjectAIConfig {
  provider: string
  language: string
  model?: string
  isDefault: boolean
}

// ... 现有类型保持不变 ...
```

### Step 2: 提交

```bash
git add frontend/src/types/index.ts
git commit -m "feat: 更新前端类型定义支持项目 AI 配置"
```

---

## Task 5: 前端 Store 更新

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`
- Modify: `frontend/src/stores/projectStore.ts`

### Step 1: 更新 commitStore

```typescript
// frontend/src/stores/commitStore.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProjectStatus, ProjectAIConfig } from '../types'
import {
  GetProjectStatus,
  GenerateCommit,
  GetProjectAIConfig,
  UpdateProjectAIConfig,
  ValidateProjectConfig,
  ConfirmResetProjectConfig
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useCommitStore = defineStore('commit', () => {
  const selectedProjectPath = ref<string>('')
  const selectedProjectId = ref<number>(0)
  const projectStatus = ref<ProjectStatus | null>(null)
  const isGenerating = ref(false)
  const streamingMessage = ref('')
  const generatedMessage = ref('')
  const error = ref<string | null>(null)

  // Provider settings
  const provider = ref('openai')
  const language = ref('zh')
  const isDefaultConfig = ref(true)  // 标记是否使用默认配置
  const isSavingConfig = ref(false)  // 保存状态

  // 配置验证状态
  const configValidation = ref<{
    valid: boolean
    resetFields: string[]
    suggestedConfig?: ProjectAIConfig
  } | null>(null)

  async function loadProjectStatus(path: string) {
    selectedProjectPath.value = path
    error.value = null

    try {
      const result = await GetProjectStatus(path)
      projectStatus.value = result as ProjectStatus
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载项目状态失败'
      error.value = message
    }
  }

  async function loadProjectAIConfig(projectId: number) {
    selectedProjectId.value = projectId

    try {
      const config = await GetProjectAIConfig(projectId) as ProjectAIConfig
      provider.value = config.provider
      language.value = config.language
      isDefaultConfig.value = config.isDefault

      // 验证配置
      const [valid, resetFields, suggestedConfig] = await ValidateProjectConfig(projectId)

      if (!valid && resetFields.length > 0) {
        configValidation.value = {
          valid: false,
          resetFields,
          suggestedConfig: suggestedConfig as ProjectAIConfig
        }
      } else {
        configValidation.value = null
      }
    } catch (e: unknown) {
      console.error('加载项目配置失败:', e)
      // 失败时使用默认配置
      provider.value = 'openai'
      language.value = 'zh'
      isDefaultConfig.value = true
    }
  }

  async function saveProjectConfig(projectId: number) {
    if (isSavingConfig.value) {
      return
    }

    isSavingConfig.value = true

    try {
      await UpdateProjectAIConfig(
        projectId,
        isDefaultConfig.value ? '' : provider.value,
        isDefaultConfig.value ? '' : language.value,
        '',
        isDefaultConfig.value
      )
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '保存配置失败'
      error.value = message
      throw e
    } finally {
      isSavingConfig.value = false
    }
  }

  async function confirmResetConfig(projectId: number) {
    try {
      await ConfirmResetProjectConfig(projectId)

      // 重新加载配置
      await loadProjectAIConfig(projectId)

      configValidation.value = null
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '重置配置失败'
      error.value = message
      throw e
    }
  }

  async function generateCommit() {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    isGenerating.value = true
    streamingMessage.value = ''
    generatedMessage.value = ''
    error.value = null

    try {
      await GenerateCommit(
        selectedProjectPath.value,
        provider.value,
        language.value
      )
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '生成失败'
      error.value = message
      isGenerating.value = false
    }
  }

  function clearMessage() {
    streamingMessage.value = ''
    generatedMessage.value = ''
  }

  // Setup event listeners
  EventsOn('commit-delta', (delta: string) => {
    streamingMessage.value += delta
  })

  EventsOn('commit-complete', (message: string) => {
    generatedMessage.value = message
    streamingMessage.value = message
    isGenerating.value = false
  })

  EventsOn('commit-error', (err: string) => {
    error.value = err
    isGenerating.value = false
  })

  return {
    selectedProjectPath,
    selectedProjectId,
    projectStatus,
    isGenerating,
    streamingMessage,
    generatedMessage,
    error,
    provider,
    language,
    isDefaultConfig,
    isSavingConfig,
    configValidation,
    loadProjectStatus,
    loadProjectAIConfig,
    saveProjectConfig,
    confirmResetConfig,
    generateCommit,
    clearMessage
  }
})
```

### Step 2: 更新 projectStore

在 `frontend/src/stores/projectStore.ts` 中添加 `selectedProject` 计算属性。找到现有的 store 定义，添加：

```typescript
// frontend/src/stores/projectStore.ts

// 在现有的 ref 定义之后添加
const selectedProject = computed(() => {
  return projects.value.find(p => p.path === selectedPath.value)
})

// 在 return 中添加 selectedProject
return {
  // ... 现有返回值 ...
  selectedProject  // 新增
}
```

确保导入了 `computed`：

```typescript
import { ref, computed } from 'vue'
```

### Step 3: 提交

```bash
git add frontend/src/stores/commitStore.ts frontend/src/stores/projectStore.ts
git commit -m "feat: 更新 store 支持项目 AI 配置管理"
```

---

## Task 6: CommitPanel UI 更新

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

### Step 1: 更新模板部分

找到 `<section class="panel-section" v-if="commitStore.projectStatus">` 中的 AI Settings 部分，替换为：

```vue
<!-- AI Settings -->
<section class="panel-section" v-if="commitStore.projectStatus">
  <div class="section-header">
    <div class="section-title">
      <span class="icon">🤖</span>
      <h3>AI 配置</h3>
      <span v-if="!commitStore.isDefaultConfig" class="config-badge">自定义</span>
    </div>
    <button
      v-if="!commitStore.isDefaultConfig"
      @click="handleResetToDefault"
      class="btn-reset"
      title="重置为默认配置"
    >
      <span class="icon">↺</span>
      恢复默认
    </button>
  </div>

  <!-- 配置不一致警告 -->
  <div
    v-if="commitStore.configValidation && !commitStore.configValidation.valid"
    class="config-warning-banner"
  >
    <div class="warning-content">
      <span class="icon">⚠️</span>
      <div class="warning-text">
        <strong>配置已过时</strong>
        <p>该项目配置的 {{ formatResetFields(commitStore.configValidation.resetFields) }} 在配置文件中不存在</p>
      </div>
    </div>
    <button @click="handleConfirmReset" class="btn-confirm-reset">
      确认重置
    </button>
  </div>

  <div class="settings-grid">
    <div class="setting-group">
      <label class="setting-label">
        <span class="icon">🌐</span>
        Provider
        <span v-if="commitStore.isSavingConfig" class="saving-indicator">保存中...</span>
      </label>
      <select
        v-model="commitStore.provider"
        class="setting-select"
        @change="handleConfigChange"
        :disabled="commitStore.isSavingConfig"
      >
        <option value="openai">OpenAI</option>
        <option value="anthropic">Anthropic</option>
        <option value="deepseek">DeepSeek</option>
        <option value="ollama">Ollama</option>
        <option value="google">Google</option>
        <option value="phind">Phind</option>
      </select>
    </div>

    <div class="setting-group">
      <label class="setting-label">
        <span class="icon">🌍</span>
        语言
      </label>
      <select
        v-model="commitStore.language"
        class="setting-select"
        @change="handleConfigChange"
        :disabled="commitStore.isSavingConfig"
      >
        <option value="zh">中文</option>
        <option value="en">English</option>
      </select>
    </div>
  </div>

  <button
    @click="handleGenerate"
    :disabled="!commitStore.projectStatus.has_staged || commitStore.isGenerating"
    class="btn-generate"
    :class="{ generating: commitStore.isGenerating }"
  >
    <span class="icon" v-if="!commitStore.isGenerating">⚡</span>
    <span class="icon spin" v-else>⏳</span>
    {{ commitStore.isGenerating ? '生成中...' : '生成 Commit 消息' }}
  </button>
</section>
```

### Step 2: 更新 script 部分

更新 script setup 部分：

```vue
<script setup lang="ts">
import { ref, watch } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import { useProjectStore } from '../stores/projectStore'
import { GetProjectHistory, SaveCommitHistory, CommitLocally } from '../../wailsjs/go/main/App'
import type { CommitHistory } from '../types'

const commitStore = useCommitStore()
const projectStore = useProjectStore()
const history = ref<CommitHistory[]>([])

const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// 监听选中的项目变化
watch(() => projectStore.selectedProject, async (project) => {
  if (project) {
    await commitStore.loadProjectAIConfig(project.id)
    await commitStore.loadProjectStatus(project.path)
    await loadHistoryForProject()
  }
}, { immediate: true })

async function loadHistoryForProject() {
  const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
  if (!project) return

  try {
    const result = await GetProjectHistory(project.id)
    history.value = result || []
  } catch (e) {
    console.error('Failed to load history:', e)
  }
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < MINUTE) return '刚刚'
  if (diff < HOUR) return `${Math.floor(diff / MINUTE)} 分钟前`
  if (diff < DAY) return `${Math.floor(diff / HOUR)} 小时前`
  return date.toLocaleDateString()
}

function loadHistory(item: CommitHistory) {
  commitStore.generatedMessage = item.message
}

// 配置变更时立即保存
async function handleConfigChange() {
  if (commitStore.selectedProjectId) {
    commitStore.isDefaultConfig = false
    await commitStore.saveProjectConfig(commitStore.selectedProjectId)
  }
}

// 重置为默认配置
async function handleResetToDefault() {
  if (confirm('确定要重置为默认配置吗？')) {
    commitStore.isDefaultConfig = true
    await commitStore.saveProjectConfig(commitStore.selectedProjectId)
    // 重新加载配置
    await commitStore.loadProjectAIConfig(commitStore.selectedProjectId)
  }
}

// 确认重置过时的配置
async function handleConfirmReset() {
  if (commitStore.selectedProjectId) {
    await commitStore.confirmResetConfig(commitStore.selectedProjectId)
  }
}

function formatResetFields(fields: string[]): string {
  const fieldNames: Record<string, string> = {
    provider: '服务商',
    language: '语言'
  }
  return fields.map(f => fieldNames[f] || f).join('、')
}

async function handleGenerate() {
  await commitStore.generateCommit()
}

async function handleCopy() {
  const text = commitStore.streamingMessage || commitStore.generatedMessage
  await navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

async function handleCommit() {
  if (!commitStore.selectedProjectPath) {
    alert('请先选择项目')
    return
  }

  const message = commitStore.streamingMessage || commitStore.generatedMessage
  if (!message) {
    alert('请先生成 commit 消息')
    return
  }

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)

    const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
    if (project) {
      await SaveCommitHistory(project.id, message, commitStore.provider, commitStore.language)
    }

    alert('提交成功!')
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)
    await loadHistoryForProject()
    commitStore.clearMessage()
  } catch (e: unknown) {
    const errMessage = e instanceof Error ? e.message : '提交失败'
    alert('提交失败: ' + errMessage)
  }
}

async function handleRegenerate() {
  commitStore.clearMessage()
  await commitStore.generateCommit()
}
</script>
```

### Step 3: 添加新样式

在 style 部分末尾添加：

```vue
<style scoped>
/* ... 现有样式保持不变 ... */

/* 新增样式 */
.config-badge {
  padding: 2px 8px;
  background: rgba(6, 182, 212, 0.2);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.btn-reset {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 11px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-reset:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.config-warning-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md);
  margin-bottom: var(--space-md);
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-md);
}

.warning-content {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  flex: 1;
}

.warning-content .icon {
  font-size: 18px;
  line-height: 1;
  flex-shrink: 0;
}

.warning-text strong {
  display: block;
  font-size: 13px;
  color: var(--accent-warning);
  margin-bottom: 2px;
}

.warning-text p {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.btn-confirm-reset {
  padding: var(--space-sm) var(--space-md);
  background: var(--accent-warning);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
}

.btn-confirm-reset:hover {
  filter: brightness(1.1);
}

.saving-indicator {
  margin-left: auto;
  font-size: 10px;
  color: var(--accent-primary);
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.setting-select:disabled {
  opacity: 0.6;
  cursor: wait;
}
</style>
```

### Step 4: 提交

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "feat: 更新 CommitPanel UI 支持项目 AI 配置"
```

---

## Task 7: 运行迁移脚本

**Files:**
- Modify: `main.go` 或启动文件

### Step 1: 在应用启动时运行迁移

找到应用的启动入口（通常是 `main.go` 或 `app.go` 的 `startup` 方法），添加迁移调用：

```go
// 在 startup 方法中添加
func (a *App) startup(ctx context.Context) error {
	// ... 现有代码 ...

	// 运行数据库迁移
	if err := repository.MigrateAddProjectAIConfig(a.db); err != nil {
		logger.Errorf("数据库迁移失败: %v", err)
		return err
	}

	return nil
}
```

### Step 2: 运行应用验证

```bash
wails dev
```

Expected: 数据库迁移成功，新字段已添加

### Step 3: 提交

```bash
git add app.go
git commit -m "feat: 添加数据库迁移到启动流程"
```

---

## Task 8: 端到端测试

### Step 1: 手动测试场景

1. **添加新项目**
   - 操作: 添加一个新项目
   - 预期: 新项目默认使用配置文件的 Provider 和 Language

2. **修改项目配置**
   - 操作: 修改项目的 Provider 或 Language
   - 预期:
     - 立即保存到数据库
     - 显示"自定义"标记
     - 切换到其他项目后再切换回来，配置保持不变

3. **恢复默认配置**
   - 操作: 点击"恢复默认"按钮并确认
   - 预期:
     - 配置重置为默认值
     - "自定义"标记消失
     - "恢复默认"按钮消失

4. **配置不一致验证**
   - 操作: 在数据库中手动设置一个无效的 Provider（如 "invalid"）
   - 预期:
     - 切换到该项目时显示警告横幅
     - 点击"确认重置"后配置恢复为默认

### Step 2: 自动化测试

如果需要，可以添加 Playwright 端到端测试。

### Step 3: 提交

```bash
git add .
git commit -m "test: 添加端到端测试验证"
```

---

## Task 9: 文档更新

**Files:**
- Create: `docs/features/project-ai-config.md`

### Step 1: 创建功能文档

创建 `docs/features/project-ai-config.md`：

```markdown
# 项目级别 AI 配置功能

## 概述

AI Commit Hub 支持为不同的 Git 项目配置独立的 AI Provider 和语言设置。

## 功能特性

### 默认配置

新添加的项目默认使用配置文件中的全局设置：
- Provider: 配置文件中 `provider` 字段
- Language: 配置文件中 `language` 字段

### 自定义配置

可以为特定项目设置独立的 AI 配置：
- 打开项目详情
- 在 AI 配置区域选择 Provider 和 Language
- 配置会立即保存到数据库

### 恢复默认

点击"恢复默认"按钮可将项目配置重置为全局默认值。

### 配置验证

当项目的配置与配置文件不一致时（如 Provider 已删除），会显示警告提示用户确认重置。

## 使用示例

### 场景 1: 开源项目用英文

1. 选择开源项目
2. 将 Language 设置为 "English"
3. 后续所有 commit 消息都使用英文

### 场景 2: 个人项目用中文

1. 选择个人项目
2. 将 Language 设置为 "中文"
3. 后续所有 commit 消息都使用中文

## 配置文件位置

- Windows: `C:\Users\<username>\.ai-commit-hub\config.yaml`
- macOS/Linux: `~/.ai-commit-hub/config.yaml`
```

### Step 2: 提交

```bash
git add docs/features/project-ai-config.md
git commit -m "docs: 添加项目 AI 配置功能文档"
```

---

## Task 10: 最终验证和发布

### Step 1: 运行所有测试

```bash
# 后端测试
go test ./... -v

# 前端测试（如果有）
cd frontend && npm test
```

Expected: 所有测试通过

### Step 2: 构建验证

```bash
wails build
```

Expected: 构建成功，无错误

### Step 3: 创建最终提交

```bash
git add .
git commit -m "feat: 完成项目级别 AI 配置功能"
```

### Step 4: 创建 Pull Request

```bash
# 如果使用 Git flow
git flow feature finish project-ai-config

# 或直接推送
git push origin feature/project-ai-config
```

---

## 完成检查清单

- [ ] 数据库模型已扩展
- [ ] 数据库迁移成功
- [ ] 后端服务层已实现
- [ ] 单元测试全部通过
- [ ] App API 方法已添加
- [ ] 前端类型定义已更新
- [ ] Store 状态管理已更新
- [ ] CommitPanel UI 已更新
- [ ] 手动测试全部场景通过
- [ ] 文档已更新
- [ ] 代码已提交

---

## 相关文件

- 设计文档: `docs/plans/2025-01-23-project-ai-config-design.md`
- 实施计划: `docs/plans/2025-01-23-project-ai-config-implementation.md` (本文件)
- 功能文档: `docs/features/project-ai-config.md`
