# AI Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Integrate AI commit message generation using ai-commit core libraries, with streaming output, custom prompt templates, and local git commit support.

**Architecture:**
- Reuse ai-commit's proven packages (ai, provider, config, git, prompt)
- Wails Events for streaming AI responses to frontend
- Left-right split layout: project list (30%) | commit panel (70%)
- Only process staged changes (git diff --cached)

**Tech Stack:** Go (ai-commit libs), Wails v2, Vue3 Composition API, Pinia, SQLite

---

## Task 5: 集成 ai-commit 核心包

**Files:**
- Create: `pkg/ai/ai.go`
- Create: `pkg/ai/client.go`
- Create: `pkg/provider/registry/registry.go`
- Create: `pkg/provider/openai_compat/client.go`
- Create: `pkg/provider/openai/client.go`
- Create: `pkg/provider/openai/register.go`
- Create: `pkg/provider/anthropic/client.go`
- Create: `pkg/provider/anthropic/register.go`
- Create: `pkg/provider/deepseek/client.go`
- Create: `pkg/provider/deepseek/register.go`
- Create: `pkg/provider/ollama/client.go`
- Create: `pkg/provider/ollama/register.go`
- Create: `pkg/config/config.go`
- Create: `pkg/config/prompts.go`
- Create: `pkg/git/git.go`
- Create: `pkg/prompt/prompt.go`
- Create: `go.mod` (update)

**Step 1: Copy ai-commit packages**

Copy from `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit`:
- `pkg/ai/*` → `pkg/ai/`
- `pkg/provider/*` → `pkg/provider/`
- `pkg/config/*` → `pkg/config/`
- `pkg/git/*` → `pkg/git/`
- `pkg/prompt/*` → `pkg/prompt/`

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub

# Copy directories
cp -r /c/WorkSpace/Go2Hell/src/github.com/renatogalera/ai-commit/pkg/ai pkg/
cp -r /c/WorkSpace/Go2Hell/src/github.com/renatogalera/ai-commit/pkg/provider pkg/
cp -r /c/WorkSpace/Go2Hell/src/github.com/renatogalera/ai-commit/pkg/config pkg/
cp -r /c/WorkSpace/Go2Hell/src/github.com/renatogalera/ai-commit/pkg/git pkg/
cp -r /c/WorkSpace/Go2Hell/src/github.com/renatogalera/ai-commit/pkg/prompt pkg/
```

**Step 2: Update go.mod with required dependencies**

```bash
go get github.com/openai/openai-go/v2@latest
go get github.com/anthropics/anthropic-go-go/v11@latest
go get github.com/go-git/go-git/v5@latest
go get github.com/sergi/go-diff/diffmatchpatch@latest
go get github.com/go-playground/validator/v10@latest
go get gopkg.in/yaml.v3@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/dustin/go-humanize@latest
```

**Step 3: Update package imports**

Replace `github.com/renatogalera/ai-commit` with `github.com/allanpk716/ai-commit-hub` in all copied files:

```bash
find pkg/ -type f -name "*.go" -exec sed -i 's|github.com/renatogalera/ai-commit|github.com/allanpk716/ai-commit-hub|g' {} \;
```

**Step 4: Run go mod tidy**

```bash
go mod tidy
```

Expected: No errors, dependencies resolved.

**Step 5: Test imports compile**

```bash
go build ./...
```

Expected: SUCCESS or only missing main issues.

**Step 6: Commit**

```bash
git add pkg/ go.mod go.sum
git commit -m "feat: integrate ai-commit core packages

- Copy ai, provider, config, git, prompt packages from ai-commit
- Add dependencies: openai-go, anthropic-go, go-git, diffmatchpatch
- Update package imports to use ai-commit-hub module
"
```

---

## Task 6: 实现配置管理和 Provider 注册

**Files:**
- Modify: `app.go`
- Create: `pkg/service/config_service.go`
- Create: `.ai-commit-hub/config.yaml.example`

**Step 1: Create config service**

Create `pkg/service/config_service.go`:

```go
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
        Provider: "openai",
        Language: "zh",
        Providers: make(map[string]config.ProviderSettings),
    }

    if _, err := os.Stat(configPath); err == nil {
        data, _ := os.ReadFile(configPath)
        // Simple YAML parse (use gopkg.in/yaml.v3)
        // For now, return default if parse fails
    }

    return cfg, nil
}

func (s *ConfigService) GetAvailableProviders() []string {
    return registry.Names()
}

func (s *ConfigService) ResolvePromptTemplate(configDir, configFile string) (string, error) {
    if configFile == "" {
        return config.DefaultPromptTemplate, nil
    }

    promptPath := filepath.Join(configDir, "prompts", configFile)
    content, err := os.ReadFile(promptPath)
    if err != nil {
        return "", err
    }

    return string(content), nil
}
```

**Step 2: Update App struct**

Modify `app.go`:

```go
import "github.com/allanpk716/ai-commit-hub/pkg/service"

type App struct {
    ctx             context.Context
    dbPath          string
    gitProjectRepo  *repository.GitProjectRepository
    initError       error
    configService   *service.ConfigService  // Add this
}
```

**Step 3: Initialize config service in startup**

Modify `app.go` startup method:

```go
func (a *App) startup(ctx context.Context) {
    // ... existing code ...

    // Initialize config service
    a.configService = service.NewConfigService()

    fmt.Println("AI Commit Hub initialized successfully")
}
```

**Step 4: Create example config file**

Create `.ai-commit-hub/config.yaml.example`:

```yaml
# AI Provider 配置
provider: openai
language: zh

# Provider 详细配置
providers:
  openai:
    apiKey: your-openai-api-key-here
    model: gpt-4
    baseURL: https://api.openai.com/v1

  anthropic:
    apiKey: your-anthropic-api-key-here

  deepseek:
    apiKey: your-deepseek-api-key-here

  ollama:
    baseURL: http://localhost:11434
    model: llama2

# 自定义 Prompt 模板（可选）
prompts:
  commitMessage: custom-prompt.txt
```

**Step 5: Build to verify**

```bash
wails build
```

Expected: SUCCESS

**Step 6: Commit**

```bash
git add app.go pkg/service/ .ai-commit-hub/
git commit -m "feat: add config service and provider registration

- Create ConfigService for loading and managing AI provider config
- Register all AI providers (OpenAI, Anthropic, DeepSeek, Ollama)
- Add example config file with provider settings
- Support custom prompt templates from prompts/ directory
"
```

---

## Task 7: 实现暂存区状态查询

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/types/index.ts`
- Create: `pkg/git/status.go`

**Step 1: Create git status package**

Create `pkg/git/status.go`:

```go
package git

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

type StagedFile struct {
    Path   string `json:"path"`
    Status string `json:"status"` // Modified, New, Deleted, Renamed
}

type ProjectStatus struct {
    Branch      string      `json:"branch"`
    StagedFiles []StagedFile `json:"staged_files"`
    HasStaged   bool        `json:"has_staged"`
}

func GetProjectStatus(ctx context.Context, projectPath string) (*ProjectStatus, error) {
    // Check if it's a git repo
    _, err := os.Stat(filepath.Join(projectPath, ".git"))
    if os.IsNotExist(err) {
        return nil, fmt.Errorf("不是 git 仓库: %s", projectPath)
    }

    // Get current branch
    branch, _ := getCurrentBranch(projectPath)

    // Get staged files
    stagedFiles, err := getStagedFiles(projectPath)
    if err != nil {
        return nil, err
    }

    return &ProjectStatus{
        Branch:      branch,
        StagedFiles: stagedFiles,
        HasStaged:   len(stagedFiles) > 0,
    }, nil
}

func getCurrentBranch(projectPath string) (string, error) {
    cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
    cmd.Dir = projectPath
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

func getStagedFiles(projectPath string) ([]StagedFile, error) {
    cmd := exec.Command("git", "diff", "--cached", "--name-status")
    cmd.Dir = projectPath
    output, err := cmd.Output()
    if err != nil {
        return []StagedFile{}, nil // No staged files
    }

    var files []StagedFile
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")

    for _, line := range lines {
        if line == "" {
            continue
        }

        parts := strings.SplitN(line, "\t", 2)
        if len(parts) < 2 {
            continue
        }

        statusCode := parts[0]
        filePath := parts[1]

        var status string
        switch statusCode {
        case "M":
            status = "Modified"
        case "A":
            status = "New"
        case "D":
            status = "Deleted"
        case "R":
            status = "Renamed"
        default:
            status = "Modified"
        }

        files = append(files, StagedFile{
            Path:   filePath,
            Status: status,
        })
    }

    return files, nil
}
```

**Step 2: Add GetProjectStatus to App**

Add to `app.go`:

```go
func (a *App) GetProjectStatus(projectPath string) (map[string]interface{}, error) {
    if a.initError != nil {
        return nil, a.initError
    }

    status, err := git.GetProjectStatus(context.Background(), projectPath)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "branch":       status.Branch,
        "staged_files": status.StagedFiles,
        "has_staged":    status.HasStaged,
    }, nil
}
```

**Step 3: Add TypeScript types**

Add to `frontend/src/types/index.ts`:

```typescript
export interface StagedFile {
  path: string
  status: string // 'Modified' | 'New' | 'Deleted' | 'Renamed'
}

export interface ProjectStatus {
  branch: string
  staged_files: StagedFile[]
  has_staged: boolean
}
```

**Step 4: Build and regenerate bindings**

```bash
wails build
```

Expected: SUCCESS, wailsjs/go/main/App.js includes GetProjectStatus

**Step 5: Commit**

```bash
git add pkg/git/ app.go frontend/src/types/
git commit -m "feat: add project status query for staged files

- Implement GetProjectStatus to check git staged changes
- Get current branch name
- Parse git diff --cached --name-status output
- Return staged files with status (Modified, New, Deleted, Renamed)
- Add TypeScript types for ProjectStatus and StagedFile
"
```

---

## Task 8: 实现流式 Commit 生成 API

**Files:**
- Modify: `app.go`
- Create: `pkg/service/commit_service.go`
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: Create commit service**

Create `pkg/service/commit_service.go`:

```go
package service

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/allanpk716/ai-commit-hub/pkg/ai"
    "github.com/allanpk716/ai-commit-hub/pkg/config"
    "github.com/allanpk716/ai-commit-hub/pkg/git"
    "github.com/allanpk716/ai-commit-hub/pkg/prompt"
    "github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CommitService struct {
    ctx context.Context
}

func NewCommitService(ctx context.Context) *CommitService {
    return &CommitService{ctx: ctx}
}

func (s *CommitService) GenerateCommit(projectPath, providerName, language string) error {
    // Load config
    homeDir, _ := os.UserHomeDir()
    configDir := filepath.Join(homeDir, ".ai-commit-hub")
    cfg, _ := config.LoadOrCreateConfig()

    // Override provider if specified
    if providerName != "" {
        cfg.Provider = providerName
    }
    if language != "" {
        cfg.Language = language
    }

    // Get AI client
    factory, ok := registry.Get(cfg.Provider)
    if !ok {
        return fmt.Errorf("未知的 provider: %s", cfg.Provider)
    }

    client, err := factory(context.Background(), cfg.Provider, cfg.Providers[cfg.Provider])
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
```

**Step 2: Add GenerateCommit method to App**

Add to `app.go`:

```go
func (a *App) GenerateCommit(projectPath, provider, language string) error {
    if a.initError != nil {
        return a.initError
    }

    commitService := service.NewCommitService(a.ctx)
    return commitService.GenerateCommit(projectPath, provider, language)
}
```

**Step 3: Create commit store**

Create `frontend/src/stores/commitStore.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProjectStatus } from '../types'
import { GetProjectStatus, GenerateCommit } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useCommitStore = defineStore('commit', () => {
  const selectedProjectPath = ref<string>('')
  const projectStatus = ref<ProjectStatus | null>(null)
  const isGenerating = ref(false)
  const streamingMessage = ref('')
  const generatedMessage = ref('')
  const error = ref<string | null>(null)

  // Provider settings
  const provider = ref('openai')
  const language = ref('zh')

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
    projectStatus,
    isGenerating,
    streamingMessage,
    generatedMessage,
    error,
    provider,
    language,
    loadProjectStatus,
    generateCommit,
    clearMessage
  }
})
```

**Step 4: Build to test**

```bash
wails build
```

Expected: SUCCESS

**Step 5: Commit**

```bash
git add pkg/service/ app.go frontend/src/stores/
git commit -m "feat: implement streaming commit message generation

- Create CommitService for AI-powered commit generation
- Use Wails Events for real-time streaming to frontend
- Support provider and language selection
- Add commit store with Pinia for state management
- Handle commit-delta, commit-complete, and commit-error events
"
```

---

## Task 9: 创建 CommitPanel 组件

**Files:**
- Create: `frontend/src/components/CommitPanel.vue`

**Step 1: Create CommitPanel component**

Create `frontend/src/components/CommitPanel.vue`:

```vue
<template>
  <div class="commit-panel">
    <!-- Project Info Section -->
    <div class="section" v-if="commitStore.projectStatus">
      <h3>{{ commitStore.projectStatus.branch }}</h3>
      <div class="staged-files">
        <div v-if="!commitStore.projectStatus.has_staged" class="empty-state">
          暂存区为空，请先 git add 文件
        </div>
        <div v-else>
          <div
            v-for="file in commitStore.projectStatus.staged_files"
            :key="file.path"
            class="file-item"
          >
            <span class="file-status" :class="file.status.toLowerCase()">
              {{ file.status }}
            </span>
            <span class="file-path">{{ file.path }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div class="section empty" v-else>
      <p>请从左侧选择一个项目</p>
    </div>

    <!-- AI Settings -->
    <div class="section" v-if="commitStore.projectStatus">
      <h3>AI 设置</h3>
      <div class="settings">
        <div class="setting-row">
          <label>Provider:</label>
          <select v-model="commitStore.provider">
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="deepseek">DeepSeek</option>
            <option value="ollama">Ollama</option>
          </select>
        </div>
        <div class="setting-row">
          <label>语言:</label>
          <select v-model="commitStore.language">
            <option value="zh">中文</option>
            <option value="english">English</option>
          </select>
        </div>
      </div>
      <button
        @click="handleGenerate"
        :disabled="!commitStore.projectStatus.has_staged || commitStore.isGenerating"
        class="btn-primary"
      >
        {{ commitStore.isGenerating ? '生成中...' : '生成 Commit 消息' }}
      </button>
    </div>

    <!-- Generated Message -->
    <div class="section" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
      <h3>生成结果</h3>
      <div class="message-area">
        <pre class="message-content">{{ commitStore.streamingMessage || commitStore.generatedMessage }}</pre>
      </div>
      <div class="actions">
        <button @click="handleCopy" class="btn-secondary">复制</button>
        <button @click="handleCommit" class="btn-primary">提交到本地</button>
        <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-secondary">重新生成</button>
      </div>
    </div>

    <!-- Error -->
    <div class="section error" v-if="commitStore.error">
      {{ commitStore.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCommitStore } from '../stores/commitStore'

const commitStore = useCommitStore()

async function handleGenerate() {
  await commitStore.generateCommit()
}

async function handleCopy() {
  const text = commitStore.streamingMessage.value || commitStore.generatedMessage.value
  await navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

async function handleCommit() {
  // TODO: Implement in next task
  alert('提交功能将在下一步实现')
}

async function handleRegenerate() {
  commitStore.clearMessage()
  await commitStore.generateCommit()
}
</script>

<style scoped>
.commit-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
  overflow-y: auto;
}

.section {
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.section h3 {
  margin-top: 0;
  margin-bottom: 15px;
}

.empty {
  text-align: center;
  color: #999;
}

.staged-files {
  max-height: 200px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 6px 0;
  font-size: 14px;
}

.file-status {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  margin-right: 8px;
  font-weight: bold;
}

.file-status.modified { background: #fff3cd; color: #856404; }
.file-status.new { background: #d1e7dd; color: #0f5132; }
.file-status.deleted { background: #f8d7da; color: #842029; }
.file-status.renamed { background: #cff4fc; color: #055160; }

.file-path {
  flex: 1;
  font-family: monospace;
  word-break: break-all;
}

.settings {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 15px;
}

.setting-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.setting-row label {
  min-width: 80px;
}

.setting-row select {
  flex: 1;
  padding: 6px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.message-area {
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  max-height: 300px;
  overflow-y: auto;
}

.message-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.btn-primary, .btn-secondary {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: #2196f3;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #1976d2;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #5a6268;
}

.error {
  background: #f8d7da;
  color: #842029;
  border: 1px solid #f5c2c7;
}
</style>
```

**Step 2: Update App.vue to use CommitPanel**

Modify `frontend/src/App.vue`:

```vue
<template>
  <div class="app">
    <div class="toolbar">
      <h1>AI Commit Hub</h1>
      <div class="toolbar-actions">
        <button @click="openAddProject">+ 添加项目</button>
        <button @click="openConfigFolder">⚙ 设置</button>
      </div>
    </div>

    <div class="content">
      <ProjectList
        :selected-id="selectedProjectId"
        @select="handleSelectProject"
      />
      <CommitPanel v-if="selectedProjectId" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { useCommitStore } from './stores/commitStore'
import { OpenConfigFolder, SelectProjectFolder } from './wailsjs/go/main/App'
import ProjectList from './components/ProjectList.vue'
import CommitPanel from './components/CommitPanel.vue'
import type { GitProject } from './types'

const projectStore = useProjectStore()
const commitStore = useCommitStore()
const selectedProjectId = ref<number>()

onMounted(() => {
  projectStore.loadProjects()
})

async function openAddProject() {
  try {
    const path = await SelectProjectFolder()
    if (path) {
      await projectStore.addProject(path)
    }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '添加项目失败'
    alert('添加项目失败: ' + message)
  }
}

function handleSelectProject(project: GitProject) {
  selectedProjectId.value = project.id
  // Load project status for commit panel
  commitStore.loadProjectStatus(project.path)
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '打开配置文件夹失败'
    alert('打开配置文件夹失败: ' + message)
  }
}
</script>

<style scoped>
/* ... existing styles ... */
.content {
  display: flex;
  gap: 20px;
  height: calc(100vh - 70px);
}
</style>
```

**Step 3: Build to verify**

```bash
wails build
```

Expected: SUCCESS

**Step 4: Commit**

```bash
git add frontend/src/components/CommitPanel.vue frontend/src/App.vue
git commit -m "feat: create CommitPanel component with streaming AI

- Display project branch and staged files list
- Add AI settings (provider, language selection)
- Implement streaming message display with typewriter effect
- Add copy, regenerate, and commit buttons (commit in next task)
- Integrate with commit store for state management
"
```

---

## Task 10: 实现本地提交功能

**Files:**
- Modify: `app.go`
- Modify: `pkg/git/status.go`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: Add CommitChanges to git package**

Add to `pkg/git/status.go`:

```go
func CommitChanges(projectPath, message string) error {
    cmd := exec.Command("git", "commit", "-m", message)
    cmd.Dir = projectPath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("git commit 失败: %s\nOutput: %s", err, string(output))
    }
    return nil
}
```

**Step 2: Add CommitLocally to App**

Add to `app.go`:

```go
func (a *App) CommitLocally(projectPath, message string) error {
    if a.initError != nil {
        return a.initError
    }

    if message == "" {
        return fmt.Errorf("commit 消息不能为空")
    }

    return git.CommitChanges(projectPath, message)
}
```

**Step 3: Update CommitPanel to implement commit**

Modify `frontend/src/components/CommitPanel.vue` script section:

```typescript
import { useCommitStore } from '../stores/commitStore'
import { CommitLocally } from '../../wailsjs/go/main/App'

const commitStore = useCommitStore()

async function handleCopy() {
  const text = commitStore.streamingMessage.value || commitStore.generatedMessage.value
  await navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

async function handleCommit() {
  if (!commitStore.selectedProjectPath) {
    alert('请先选择项目')
    return
  }

  const message = commitStore.streamingMessage.value || commitStore.generatedMessage.value
  if (!message) {
    alert('请先生成 commit 消息')
    return
  }

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)
    alert('提交成功!')

    // Reload project status
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)

    // Clear message
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
```

**Step 4: Build and test**

```bash
wails build
```

Expected: SUCCESS

**Step 5: Commit**

```bash
git add pkg/git/ app.go frontend/src/components/CommitPanel.vue
git commit -m "feat: implement local git commit functionality

- Add CommitChanges function to execute git commit -m
- Add CommitLocally API method
- Implement handleCommit in CommitPanel
- Reload project status after successful commit
- Show success/error alerts for commit operations
"
```

---

## Task 11: 实现历史记录功能

**Files:**
- Create: `pkg/models/commit_history.go`
- Modify: `app.go`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: Create CommitHistory model**

Create `pkg/models/commit_history.go`:

```go
package models

import (
    "time"
)

type CommitHistory struct {
    ID        uint   `gorm:"primaryKey" json:"id"`
    ProjectID uint   `gorm:"index" json:"project_id"`
    Message   string `gorm:"type:text" json:"message"`
    Provider  string `json:"provider"`
    Language  string `json:"language"`
    CreatedAt time.Time `json:"created_at"`

    Project   GitProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (CommitHistory) TableName() string {
    return "commit_histories"
}
```

**Step 2: Add migration**

Modify `app.go` startup to include CommitHistory:

```go
import "github.com/allanpk716/ai-commit-hub/pkg/models"

// In startup function, after InitializeDatabase:
if err := db.AutoMigrate(&models.GitProject{}, &models.CommitHistory{}); err != nil {
```

**Step 3: Create repository**

Create `pkg/repository/commit_history_repository.go`:

```go
package repository

import (
    "fmt"

    "github.com/allanpk716/ai-commit-hub/pkg/models"
    "gorm.io/gorm"
)

type CommitHistoryRepository struct {
    db *gorm.DB
}

func NewCommitHistoryRepository() *CommitHistoryRepository {
    return &CommitHistoryRepository{db: GetDB()}
}

func (r *CommitHistoryRepository) Create(history *models.CommitHistory) error {
    if err := r.db.Create(history).Error; err != nil {
        return fmt.Errorf("failed to create commit history: %w", err)
    }
    return nil
}

func (r *CommitHistoryRepository) GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error) {
    var histories []models.CommitHistory
    err := r.db.Where("project_id = ?", projectID).
        Order("created_at DESC").
        Limit(limit).
        Find(&histories).Error
    return histories, err
}

func (r *CommitHistoryRepository) GetRecent(limit int) ([]models.CommitHistory, error) {
    var histories []models.CommitHistory
    err := r.db.Preload("Project").
        Order("created_at DESC").
        Limit(limit).
        Find(&histories).Error
    return histories, err
}
```

**Step 4: Add API methods**

Add to `app.go`:

```go
import "github.com/allanpk716/ai-commit-hub/pkg/repository"

type App struct {
    // ... existing fields ...
    commitHistoryRepo *repository.CommitHistoryRepository
}

func (a *App) startup(ctx context.Context) {
    // ... existing code ...

    // Initialize commit history repository
    a.commitHistoryRepo = repository.NewCommitHistoryRepository()
}

func (a *App) SaveCommitHistory(projectID uint, message, provider, language string) error {
    history := &models.CommitHistory{
        ProjectID: projectID,
        Message:   message,
        Provider:  provider,
        Language:  language,
    }

    if err := a.commitHistoryRepo.Create(history); err != nil {
        return fmt.Errorf("保存历史记录失败: %w", err)
    }
    return nil
}

func (a *App) GetProjectHistory(projectID uint) ([]models.CommitHistory, error) {
    histories, err := a.commitHistoryRepo.GetByProjectID(projectID, 10)
    if err != nil {
        return nil, err
    }
    return histories, nil
}
```

**Step 5: Add TypeScript types**

Add to `frontend/src/types/index.ts`:

```typescript
export interface CommitHistory {
  id: number
  project_id: number
  message: string
  provider: string
  language: string
  created_at: string
  project?: GitProject
}
```

**Step 6: Update commit service to save history**

Modify `pkg/service/commit_service.go`:

```go
func (s *CommitService) SaveHistory(projectID uint, message, provider, language string) error {
    // Will be called from App after generation completes
    return nil
}
```

**Step 7: Build to test**

```bash
wails build
```

Expected: SUCCESS

**Step 8: Commit**

```bash
git add pkg/models/ pkg/repository/ app.go frontend/src/types/
git commit -m "feat: add commit history tracking

- Create CommitHistory model with GORM
- Implement CommitHistoryRepository with CRUD operations
- Add SaveCommitHistory and GetProjectHistory APIs
- Store generated messages with provider and language
- Support retrieving recent history per project
"
```

---

## Task 12: 完善 UI 和错误处理

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`
- Modify: `frontend/src/stores/commitStore.ts`
- Create: `frontend/src/components/HistoryPanel.vue`

**Step 1: Add history panel to CommitPanel**

Add to CommitPanel.vue template:

```vue
<template>
  <!-- ... existing code ... -->

  <!-- History Section -->
  <div class="section" v-if="history.length > 0">
    <h3>历史记录</h3>
    <div class="history-list">
      <div
        v-for="item in history"
        :key="item.id"
        class="history-item"
        @click="loadHistory(item)"
      >
        <div class="history-meta">
          <span class="history-provider">{{ item.provider }}</span>
          <span class="history-time">{{ formatTime(item.created_at) }}</span>
        </div>
        <div class="history-message">{{ item.message.substring(0, 100) }}...</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import { GetProjectHistory } from '../../wailsjs/go/main/App'

const commitStore = useCommitStore()
const history = ref<any[]>([])

watch(() => commitStore.selectedProjectPath, async (path) => {
  if (path) {
    try {
      // Get project from store to find ID
      const result = await GetProjectHistory(1) // TODO: get actual project ID
      history.value = result || []
    } catch (e) {
      console.error('Failed to load history:', e)
    }
  }
})

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return date.toLocaleDateString()
}

function loadHistory(item: any) {
  commitStore.generatedMessage = item.message
  commitStore.streamingMessage = item.message
}
</script>

<style scoped>
/* ... existing styles ... */

.history-list {
  max-height: 200px;
  overflow-y: auto;
}

.history-item {
  padding: 10px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  margin-bottom: 8px;
  cursor: pointer;
}

.history-item:hover {
  background: #f5f5f5;
}

.history-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

.history-provider {
  padding: 2px 6px;
  background: #e3f2fd;
  border-radius: 3px;
}

.history-message {
  font-size: 13px;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
```

**Step 2: Add loading states**

Update CommitPanel to show loading:

```vue
<template>
  <div class="message-area" v-if="commitStore.isGenerating">
    <div class="loading-indicator">
      <span class="spinner"></span>
      AI 正在生成...
    </div>
    <pre class="message-content">{{ commitStore.streamingMessage }}</pre>
  </div>

  <div class="message-area" v-else-if="commitStore.generatedMessage">
    <pre class="message-content">{{ commitStore.generatedMessage }}</pre>
  </div>
</template>

<style scoped>
.loading-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  color: #2196f3;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #f3f3f3;
  border-top: 2px solid #2196f3;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
```

**Step 3: Build to verify**

```bash
wails build
```

Expected: SUCCESS

**Step 4: Commit**

```bash
git add frontend/src/components/CommitPanel.vue frontend/src/stores/
git commit -m "feat: add history panel and loading states

- Display commit history with provider and timestamp
- Click history item to reload message
- Add loading spinner during AI generation
- Format relative time (刚刚, X分钟前, etc.)
- Improve UX with better error display
"
```

---

## Task 13: 最终测试和优化

**Files:**
- Create: `README.md`
- Create: `.ai-commit-hub/config.yaml` (default config template)

**Step 1: Create default config template**

Create `.ai-commit-hub/config.yaml`:

```yaml
# AI Commit Hub 配置文件

# 默认 AI Provider
provider: openai

# 默认语言
language: zh

# Provider 配置
providers:
  openai:
    # 请在下方填写您的 OpenAI API Key
    apiKey:
    model: gpt-4
    baseURL: https://api.openai.com/v1

  # 如果使用 Ollama（本地运行）
  # ollama:
  #   baseURL: http://localhost:11434
  #   model: llama2

# 自定义 Prompt 模板（可选）
# 将模板文件放在 ~/.config/ai-commit-hub/prompts/ 目录下
prompts:
  # commitMessage: my-custom-prompt.txt
```

**Step 2: Update README**

Create/update `README.md`:

```markdown
# AI Commit Hub

一个桌面应用，用于为多个 Git 项目生成 AI 驱动的 commit 消息。

## 功能特性

- 📁 管理多个 Git 项目
- 🤖 支持多种 AI Provider (OpenAI, Anthropic, DeepSeek, Ollama)
- ⚡ 实时流式生成 commit 消息
- 📝 支持自定义 Prompt 模板
- 📋 一键复制或直接提交到本地
- 📜 历史记录功能

## 快速开始

### 安装

\`\`\`bash
# 克隆仓库
git clone https://github.com/allanpk716/ai-commit-hub.git
cd ai-commit-hub

# 安装依赖
go mod tidy
cd frontend && npm install

# 构建
wails build
\`\`\`

### 配置

1. 复制配置文件模板：
\`\`\`bash
mkdir -p ~/.config/ai-commit-hub
cp .ai-commit-hub/config.yaml ~/.config/ai-commit-hub/
\`\`\`

2. 编辑配置文件，填入您的 API Key

3. (可选) 添加自定义 prompt 模板到 ~/.config/ai-commit-hub/prompts/

### 使用

\`\`\`bash
# 开发模式
wails dev

# 生产构建
wails build
\`\`\`

## 工作流

1. 在 Git 客户端/IDE 中: `git add` 文件到暂存区
2. 在 AI Commit Hub 中选择项目
3. 选择 Provider 和语言，点击"生成 Commit 消息"
4. 查看流式生成结果
5. 点击"复制"或"提交到本地"

## 开发

\`\`\`bash
# 运行 Go 测试
go test ./...

# 运行前端测试
cd frontend && npm test
\`\`\`
```

**Step 3: Run full test**

```bash
# Backend tests
go test ./... -v

# Build
wails build

# Run application
./build/bin/ai-commit-hub.exe
```

Expected: Application starts, no errors

**Step 4: Test workflow manually**

1. Add a project
2. Make changes in that project
3. `git add` some files
4. Select project in AI Commit Hub
5. Verify staged files show up
6. Click "生成 Commit 消息"
7. Verify streaming output
8. Click "提交到本地"
9. Verify commit was created

**Step 5: Final commit**

```bash
git add README.md .ai-commit-hub/
git commit -m "docs: add README and default config template

- Document features, installation, and usage
- Add config.yaml template with provider setup
- Include step-by-step workflow guide
- Document development and testing instructions
"
```

---

## 测试指南

### 运行测试

\`\`\`bash
# Go 后端测试
go test ./... -v

# 前端测试 (如果配置了)
cd frontend
npm test
\`\`\`

### 手动测试清单

- [ ] 添加项目（选择文件夹）
- [ ] 删除项目
- [ ] 项目上移/下移
- [ ] 拖拽排序
- [ ] 搜索过滤
- [ ] 查看暂存区状态
- [ ] 生成 commit 消息（流式）
- [ ] 复制消息到剪贴板
- [ ] 提交到本地 git
- [ ] 查看历史记录
- [ ] 重新生成消息

---

## 注意事项

1. **每次修改 Go 代码后需要重启 wails dev**
2. **Wails bindings 自动生成在 frontend/wailsjs/ 目录**
3. **TypeScript 类型需要与 Go 结构体保持一致**
4. **错误处理要友好，使用中文提示**
5. **遵循 TDD 原则，先写测试再实现功能**

---

**计划完成！** 下一步可以选择：

1. **开始执行计划** - 使用 superpowers:subagent-driven-development (this session)
2. **在新会话中执行** - 使用 superpowers:executing-plans
