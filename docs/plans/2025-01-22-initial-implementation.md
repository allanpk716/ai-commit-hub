# AI Commit Hub 初始实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个带界面的 Git Commit 自动生成工具，支持多项目管理和 AI 驱动的 commit 消息生成。

**Architecture:** 采用 Wails (Go 后端) + Vue3 (前端) 的桌面应用架构。复用 ai-commit 项目的核心 AI 逻辑，使用 SQLite 存储项目信息，左右分栏布局展示项目列表和 commit 详情。

**Tech Stack:** Go 1.22+, Wails v2, Vue 3, TypeScript, Vite, Pinia, GORM, SQLite, go-git

---

## 前置准备

### Task 0: 项目初始化

**Files:**
- Create: `go.mod`
- Create: `wails.json`
- Create: `main.go`
- Create: `app.go`
- Create: `frontend/wailsjs/go/main/App.js` (generated)
- Create: `frontend/src/App.vue`

**Step 1: 初始化 Go 模块**

运行:
```bash
cd "C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub"
go mod init github.com/allanpk716/ai-commit-hub
```

预期: 创建 `go.mod` 文件

**Step 2: 安装 Wails 依赖**

运行:
```bash
go get github.com/wailsapp/wails/v2@latest
go get github.com/wailsapp/wails/v2/pkg/options/mac
go get github.com/wailsapp/wails/v2/pkg/options/windows
go get github.com/wailsapp/wails/v2/pkg/options/linux
```

预期: 更新 `go.mod` 和 `go.sum`

**Step 3: 创建 wails.json 配置**

创建 `wails.json`:
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "ai-commit-hub",
  "outputfilename": "ai-commit-hub",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "allanpk716",
    "email": "allanpk716@example.com"
  },
  "info": {
    "companyName": "AI Commit Hub",
    "productName": "AI Commit Hub",
    "productVersion": "1.0.0",
    "copyright": "Copyright........",
    "comments": "AI-powered Git commit message generator"
  }
}
```

**Step 4: 创建 main.go 入口文件**

创建 `main.go`:
```go
package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "AI Commit Hub",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

**Step 5: 创建 app.go 主应用结构**

创建 `app.go`:
```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	dbPath  string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("AI Commit Hub starting up...")

	// Set database path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get home directory:", err)
		return
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Failed to create config directory:", err)
		return
	}

	a.dbPath = filepath.Join(configDir, "ai-commit-hub.db")
	fmt.Println("Database path:", a.dbPath)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("AI Commit Hub shutting down...")
}

// Greet returns a greeting
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, AI Commit Hub is ready!", name)
}

// OpenConfigFolder opens the config folder in system file manager
// @app.Method OpenConfigFolder
func (a *App) OpenConfigFolder() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", configDir)
	case "darwin":
		cmd = exec.Command("open", configDir)
	default:
		cmd = exec.Command("xdg-open", configDir)
	}

	return cmd.Start()
}
```

**Step 6: 初始化前端**

运行:
```bash
cd frontend
npm create vite@latest . -- --template vue-ts
npm install
npm install pinia
```

预期: 创建 Vue3 + TypeScript 项目结构

**Step 7: 配置 Vite 用于 Wails**

修改 `frontend/vite.config.ts`:
```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  clearScreen: false,
  server: {
    strictPort: true,
    hmr: {
      port: 5173,
    },
  },
  envPrefix: ['VITE_', 'WAILS_'],
  build: {
    outDir: '../frontend/dist',
    emptyOutDir: true,
  },
})
```

**Step 8: 创建基础 App.vue**

创建 `frontend/src/App.vue`:
```vue
<template>
  <div class="app">
    <h1>AI Commit Hub</h1>
    <p>{{ message }}</p>
    <button @click="greet">Greet</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Greet } from '../../wailsjs/go/main/App'

const message = ref('Click the button to greet!')

const greet = async () => {
  const result = await Greet('World')
  message.value = result
}
</script>

<style scoped>
.app {
  padding: 20px;
}
</style>
```

**Step 9: 验证项目运行**

运行:
```bash
wails dev
```

预期: 应用启动，显示 "AI Commit Hub" 标题，点击按钮显示问候消息

**Step 10: 提交初始代码**

```bash
git add .
git commit -m "feat: initialize Wails project with Vue3 frontend

- Set up Go module with Wails v2
- Create basic app structure
- Initialize Vue3 + TypeScript frontend
- Add greeting test functionality
"
```

---

## 阶段 1: 数据层实现

### Task 1: 数据库初始化

**Files:**
- Create: `pkg/models/git_project.go`
- Create: `pkg/repository/db.go`
- Create: `pkg/repository/git_project_repository.go`

**Step 1: 创建 GitProject 模型**

创建 `pkg/models/git_project.go`:
```go
package models

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// GitProject represents a git repository project
type GitProject struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Path      string `gorm:"not null;uniqueIndex" json:"path"`
	Name      string `json:"name"`
	SortOrder int    `gorm:"index" json:"sort_order"`
}

// TableName specifies the table name for GitProject
func (GitProject) TableName() string {
	return "git_projects"
}

// Validate checks if the project is valid
func (gp *GitProject) Validate() error {
	if gp.Path == "" {
		return fmt.Errorf("项目路径不能为空")
	}

	// Check if path exists
	if _, err := os.Stat(gp.Path); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", gp.Path)
	}

	// Check if it's a git repository
	if _, err := git.PlainOpen(gp.Path); err != nil {
		return fmt.Errorf("不是有效的 git 仓库: %s", gp.Path)
	}

	return nil
}

// DetectName attempts to detect the project name from path or git config
func (gp *GitProject) DetectName() (string, error) {
	// Try folder name first
	folderName := filepath.Base(gp.Path)
	if folderName != "" && folderName != "." && folderName != "/" {
		return folderName, nil
	}

	// Try git config
	repo, err := git.PlainOpen(gp.Path)
	if err != nil {
		return "", fmt.Errorf("无法打开 git 仓库: %w", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return folderName, nil // fallback to folder name
	}

	// Try to get name from remote URL or use folder name
	if len(cfg.Remotes) > 0 {
		for _, remote := range cfg.Remotes {
			if len(remote.URLs) > 0 && remote.URLs[0] != "" {
				return folderName, nil // Use folder name for clarity
			}
		}
	}

	return folderName, nil
}
```

**Step 2: 创建数据库初始化代码**

创建 `pkg/repository/db.go`:
```go
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

var (
	db   *gorm.DB
	once sync.Once
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string
}

// NewDatabaseConfig creates a new database config
func NewDatabaseConfig() *DatabaseConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create config directory: %v", err))
	}

	return &DatabaseConfig{
		Path: filepath.Join(configDir, "ai-commit-hub.db"),
	}
}

// InitializeDatabase initializes the database connection
func InitializeDatabase(config *DatabaseConfig) error {
	var initErr error
	once.Do(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Auto migrate schemas
		if err := db.AutoMigrate(&models.GitProject{}); err != nil {
			initErr = fmt.Errorf("failed to migrate database: %w", err)
			return
		}

		fmt.Println("Database initialized:", config.Path)
	})

	return initErr
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
```

**Step 3: 创建 GitProject Repository**

创建 `pkg/repository/git_project_repository.go`:
```go
package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// GitProjectRepository handles git project data operations
type GitProjectRepository struct {
	db *gorm.DB
}

// NewGitProjectRepository creates a new GitProjectRepository
func NewGitProjectRepository() *GitProjectRepository {
	return &GitProjectRepository{
		db: GetDB(),
	}
}

// Create creates a new git project
func (r *GitProjectRepository) Create(project *models.GitProject) error {
	if err := r.db.Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetAll retrieves all projects ordered by sort_order
func (r *GitProjectRepository) GetAll() ([]models.GitProject, error) {
	var projects []models.GitProject
	if err := r.db.Order("sort_order asc").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}

// GetByID retrieves a project by ID
func (r *GitProjectRepository) GetByID(id uint) (*models.GitProject, error) {
	var project models.GitProject
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// Update updates a project
func (r *GitProjectRepository) Update(project *models.GitProject) error {
	if err := r.db.Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// Delete deletes a project by ID
func (r *GitProjectRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.GitProject{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// GetMaxSortOrder returns the maximum sort_order value
func (r *GitProjectRepository) GetMaxSortOrder() (int, error) {
	var maxOrder int
	if err := r.db.Model(&models.GitProject{}).
		Select("COALESCE(MAX(sort_order), -1)").
		Scan(&maxOrder).Error; err != nil {
		return 0, fmt.Errorf("failed to get max sort order: %w", err)
	}
	return maxOrder, nil
}
```

**Step 4: 添加 GORM 依赖**

运行:
```bash
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get github.com/go-git/go-git/v5
```

**Step 5: 测试数据库功能**

创建 `pkg/repository/git_project_repository_test.go`:
```go
package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

func TestGitProjectRepository(t *testing.T) {
	// Use temp database
	tempDir := t.TempDir()
	testDBPath := filepath.Join(tempDir, "test.db")

	// Initialize test database
	config := &DatabaseConfig{Path: testDBPath}
	if err := InitializeDatabase(config); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	repo := NewGitProjectRepository()

	t.Run("Create project", func(t *testing.T) {
		project := &models.GitProject{
			Path:      "/test/path",
			Name:      "test-project",
			SortOrder: 0,
		}
		if err := repo.Create(project); err != nil {
			t.Errorf("Failed to create project: %v", err)
		}
		if project.ID == 0 {
			t.Error("Project ID should be set after creation")
		}
	})

	t.Run("GetAll projects", func(t *testing.T) {
		projects, err := repo.GetAll()
		if err != nil {
			t.Errorf("Failed to get all projects: %v", err)
		}
		if len(projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(projects))
		}
	})

	t.Run("GetMaxSortOrder", func(t *testing.T) {
		maxOrder, err := repo.GetMaxSortOrder()
		if err != nil {
			t.Errorf("Failed to get max sort order: %v", err)
		}
		if maxOrder != 0 {
			t.Errorf("Expected max sort order 0, got %d", maxOrder)
		}
	})
}
```

**Step 6: 运行测试**

运行:
```bash
go test ./pkg/repository/... -v
```

预期: 测试通过

**Step 7: 提交数据层代码**

```bash
git add .
git commit -m "feat: implement data layer with GORM and SQLite

- Add GitProject model with validation
- Create database initialization with auto-migration
- Implement GitProjectRepository CRUD operations
- Add unit tests for repository
"
```

---

## 阶段 2: 项目管理功能

### Task 2: 项目添加与列表

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/App.vue`
- Create: `frontend/src/stores/projectStore.ts`
- Create: `frontend/src/types/index.ts`

**Step 1: 定义 TypeScript 类型**

创建 `frontend/src/types/index.ts`:
```typescript
export interface GitProject {
  id: number
  path: string
  name: string
  sort_order: number
  created_at?: string
  updated_at?: string
}

export interface ProjectInfo {
  branch: string
  files_changed: number
  has_staged: boolean
  path: string
  name: string
}
```

**Step 2: 在 app.go 中添加项目管理方法**

修改 `app.go`，添加以下方法:

```go
// 在 import 部分添加
import "github.com/allanpk716/ai-commit-hub/pkg/repository"
import "github.com/allanpk716/ai-commit-hub/pkg/models"

// 在 App struct 中添加字段
type App struct {
	ctx             context.Context
	dbPath          string
	gitProjectRepo  *repository.GitProjectRepository
}

// 在 startup 方法中初始化 repository
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("AI Commit Hub starting up...")

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get home directory:", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "错误",
			Message: fmt.Sprintf("无法获取用户目录: %v", err),
		})
		return
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Failed to create config directory:", err)
		return
	}

	a.dbPath = filepath.Join(configDir, "ai-commit-hub.db")

	// Initialize database
	dbConfig := &repository.DatabaseConfig{Path: a.dbPath}
	if err := repository.InitializeDatabase(dbConfig); err != nil {
		fmt.Println("Failed to initialize database:", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "错误",
			Message: fmt.Sprintf("数据库初始化失败: %v", err),
		})
		return
	}

	// Initialize repositories
	a.gitProjectRepo = repository.NewGitProjectRepository()

	fmt.Println("AI Commit Hub initialized successfully")
}
```

**Step 3: 添加 GetAllProjects 方法**

在 `app.go` 中添加:

```go
// GetAllProjects retrieves all projects
// @app.Method GetAllProjects
func (a *App) GetAllProjects() ([]models.GitProject, error) {
	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}
```

**Step 4: 添加 AddProject 方法**

在 `app.go` 中添加:

```go
// AddProject adds a new project
// @app.Method AddProject
func (a *App) AddProject(path string) (models.GitProject, error) {
	// Validate path
	project := &models.GitProject{Path: path}
	if err := project.Validate(); err != nil {
		return models.GitProject{}, fmt.Errorf("项目验证失败: %w", err)
	}

	// Detect name
	name, err := project.DetectName()
	if err != nil {
		return models.GitProject{}, fmt.Errorf("无法检测项目名称: %w", err)
	}
	project.Name = name

	// Get next sort order
	maxOrder, err := a.gitProjectRepo.GetMaxSortOrder()
	if err != nil {
		return models.GitProject{}, fmt.Errorf("无法获取排序: %w", err)
	}
	project.SortOrder = maxOrder + 1

	// Save to database
	if err := a.gitProjectRepo.Create(project); err != nil {
		return models.GitProject{}, fmt.Errorf("保存项目失败: %w", err)
	}

	return *project, nil
}
```

**Step 5: 添加 DeleteProject 方法**

在 `app.go` 中添加:

```go
// DeleteProject deletes a project
// @app.Method DeleteProject
func (a *App) DeleteProject(id uint) error {
	if err := a.gitProjectRepo.Delete(id); err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}
	return nil
}
```

**Step 6: 创建 Pinia Store**

创建 `frontend/src/stores/projectStore.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GitProject } from '../types'
import { GetAllProjects, AddProject, DeleteProject } from '../../wailsjs/go/main/App'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<GitProject[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadProjects() {
    loading.value = true
    error.value = null
    try {
      const result = await GetAllProjects()
      projects.value = result
    } catch (e: any) {
      error.value = e.message || '加载项目失败'
      console.error('Failed to load projects:', e)
    } finally {
      loading.value = false
    }
  }

  async function addProject(path: string) {
    loading.value = true
    error.value = null
    try {
      const result = await AddProject(path)
      projects.value.push(result)
      return result
    } catch (e: any) {
      error.value = e.message || '添加项目失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteProject(id: number) {
    loading.value = true
    error.value = null
    try {
      await DeleteProject(id)
      projects.value = projects.value.filter(p => p.id !== id)
    } catch (e: any) {
      error.value = e.message || '删除项目失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    projects,
    loading,
    error,
    loadProjects,
    addProject,
    deleteProject
  }
})
```

**Step 7: 更新 App.vue 实现项目列表**

修改 `frontend/src/App.vue`:
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
      <div class="project-list">
        <h2>项目列表</h2>
        <div v-if="projectStore.loading">加载中...</div>
        <div v-else-if="projectStore.error" class="error">
          {{ projectStore.error }}
        </div>
        <div v-else-if="projectStore.projects.length === 0" class="empty">
          暂无项目，请添加项目
        </div>
        <div v-else class="projects">
          <div
            v-for="project in projectStore.projects"
            :key="project.id"
            class="project-item"
          >
            <span class="project-name">{{ project.name }}</span>
            <span class="project-path">{{ project.path }}</span>
            <button @click="handleDelete(project)" class="delete-btn">✕</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { OpenConfigFolder } from './wailsjs/go/main/App'

const projectStore = useProjectStore()

onMounted(() => {
  projectStore.loadProjects()
})

async function openAddProject() {
  // TODO: Open file dialog to select project path
  const path = prompt('请输入 Git 仓库路径:')
  if (path) {
    try {
      await projectStore.addProject(path)
      alert('项目添加成功!')
    } catch (e: any) {
      alert('添加失败: ' + e.message)
    }
  }
}

async function handleDelete(project: any) {
  if (confirm(`确定要删除项目 "${project.name}" 吗?`)) {
    try {
      await projectStore.deleteProject(project.id)
    } catch (e: any) {
      alert('删除失败: ' + e.message)
    }
  }
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: any) {
    alert('打开配置文件夹失败: ' + e.message)
  }
}
</script>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #e0e0e0;
}

.toolbar h1 {
  margin: 0;
  font-size: 20px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.toolbar-actions button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.toolbar-actions button:hover {
  background: #f5f5f5;
}

.content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.project-list h2 {
  margin-top: 0;
}

.project-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  margin-bottom: 8px;
}

.project-name {
  font-weight: bold;
  margin-right: 10px;
}

.project-path {
  flex: 1;
  color: #666;
  font-size: 14px;
}

.delete-btn {
  padding: 4px 8px;
  border: 1px solid #ff4444;
  color: #ff4444;
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.delete-btn:hover {
  background: #fff5f5;
}

.error {
  color: #ff4444;
}

.empty {
  color: #999;
  text-align: center;
  padding: 40px;
}
</style>
```

**Step 8: 重新生成 Wails bindings**

运行:
```bash
wails dev
```

预期: 应用启动，显示项目列表，可以添加和删除项目

**Step 9: 提交项目管理功能**

```bash
git add .
git commit -m "feat: implement project management features

- Add GetAllProjects, AddProject, DeleteProject methods
- Create projectStore with Pinia for state management
- Update App.vue with project list UI
- Add TypeScript types for GitProject
"
```

---

## 阶段 3: 文件夹选择与拖拽排序

### Task 3: 添加文件夹选择对话框

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/App.vue`

**Step 1: 在 app.go 中添加选择文件夹方法**

```go
// SelectProjectFolder opens a folder selection dialog
// @app.Method SelectProjectFolder
func (a *App) SelectProjectFolder() (string, error) {
	selectedFile, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Git 仓库",
	})
	if err != nil {
		return "", fmt.Errorf("取消选择: %w", err)
	}

	return selectedFile, nil
}
```

**Step 2: 更新 App.vue 使用文件夹选择对话框**

修改 `openAddProject` 函数:

```typescript
async function openAddProject() {
  try {
    const path = await SelectProjectFolder()
    if (path) {
      await projectStore.addProject(path)
    }
  } catch (e: any) {
    if (e.message !== 'cancel') {
      alert('添加项目失败: ' + e.message)
    }
  }
}
```

在 import 部分添加:
```typescript
import { SelectProjectFolder } from './wailsjs/go/main/App'
```

**Step 4: 测试文件夹选择**

运行 `wails dev`，点击"添加项目"按钮，应该打开文件夹选择对话框。

### Task 4: 实现拖拽排序

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/stores/projectStore.ts`
- Modify: `frontend/src/components/ProjectList.vue`

**Step 1: 在 app.go 中添加排序方法**

```go
// MoveProject moves a project up or down
// @app.Method MoveProject
func (a *App) MoveProject(id uint, direction string) error {
	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %w", err)
	}

	// Find current project index
	var currentIndex int = -1
	for i, p := range projects {
		if p.ID == id {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return fmt.Errorf("项目不存在")
	}

	// Calculate new index
	newIndex := currentIndex
	if direction == "up" && currentIndex > 0 {
		newIndex = currentIndex - 1
	} else if direction == "down" && currentIndex < len(projects)-1 {
		newIndex = currentIndex + 1
	} else {
		return nil // No change needed
	}

	// Swap sort orders
	projects[currentIndex].SortOrder, projects[newIndex].SortOrder =
		projects[newIndex].SortOrder, projects[currentIndex].SortOrder

	// Save both projects
	if err := a.gitProjectRepo.Update(&projects[currentIndex]); err != nil {
		return fmt.Errorf("更新项目失败: %w", err)
	}
	if err := a.gitProjectRepo.Update(&projects[newIndex]); err != nil {
		return fmt.Errorf("更新项目失败: %w", err)
	}

	return nil
}

// ReorderProjects reorders projects based on new order
// @app.Method ReorderProjects
func (a *App) ReorderProjects(projects []models.GitProject) error {
	for i, project := range projects {
		project.SortOrder = i
		if err := a.gitProjectRepo.Update(&project); err != nil {
			return fmt.Errorf("更新项目排序失败: %w", err)
		}
	}
	return nil
}
```

**Step 2: 更新 projectStore 添加排序方法**

修改 `frontend/src/stores/projectStore.ts`:

```typescript
import { MoveProject, ReorderProjects } from '../../wailsjs/go/main/App'

// 在 store 中添加
async function moveProject(id: number, direction: 'up' | 'down') {
  loading.value = true
  error.value = null
  try {
    await MoveProject(id, direction)
    await loadProjects()
  } catch (e: any) {
    error.value = e.message || '移动项目失败'
    throw e
  } finally {
    loading.value = false
  }
}

async function reorderProjects(projects: GitProject[]) {
  loading.value = true
  error.value = null
  try {
    await ReorderProjects(projects)
    projects.value = projects
  } catch (e: any) {
    error.value = e.message || '重新排序失败'
    throw e
  } finally {
    loading.value = false
  }
}

return {
  // ... existing returns
  moveProject,
  reorderProjects
}
```

**Step 3: 创建 ProjectList 组件支持拖拽**

创建 `frontend/src/components/ProjectList.vue`:

```vue
<template>
  <div class="project-list">
    <div class="list-header">
      <h3>项目列表</h3>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="🔍 搜索..."
        class="search-input"
      />
    </div>

    <div class="projects">
      <div
        v-for="(project, index) in filteredProjects"
        :key="project.id"
        class="project-item"
        :class="{ selected: selectedId === project.id }"
        draggable="true"
        @dragstart="handleDragStart(project, index, $event)"
        @dragover.prevent="handleDragOver"
        @drop="handleDrop(project, index)"
        @click="selectProject(project)"
      >
        <span class="drag-handle">⋮⋮</span>
        <span class="project-index">{{ index + 1 }}.</span>
        <span class="project-name">{{ project.name }}</span>
        <div class="project-actions">
          <button
            @click.stop="moveUp(project, index)"
            :disabled="index === 0"
            title="上移"
          >↑</button>
          <button
            @click.stop="moveDown(project, index)"
            :disabled="index === filteredProjects.length - 1"
            title="下移"
          >↓</button>
          <button
            @click.stop="handleDelete(project)"
            title="删除"
          >✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { GitProject } from '../types'
import { useProjectStore } from '../stores/projectStore'

const props = defineProps<{
  selectedId?: number
}>()

const emit = defineEmits<{
  select: [project: GitProject]
}>()

const projectStore = useProjectStore()
const searchQuery = ref('')
const draggedItem = ref<{ project: GitProject; index: number } | null>(null)

const filteredProjects = computed(() => {
  if (!searchQuery.value) {
    return projectStore.projects
  }
  const query = searchQuery.value.toLowerCase()
  return projectStore.projects.filter(p =>
    p.name.toLowerCase().includes(query) ||
    p.path.toLowerCase().includes(query)
  )
})

function selectProject(project: GitProject) {
  emit('select', project)
}

async function moveUp(project: GitProject, index: number) {
  if (index > 0) {
    try {
      await projectStore.moveProject(project.id, 'up')
    } catch (e: any) {
      alert('移动失败: ' + e.message)
    }
  }
}

async function moveDown(project: GitProject, index: number) {
  if (index < filteredProjects.value.length - 1) {
    try {
      await projectStore.moveProject(project.id, 'down')
    } catch (e: any) {
      alert('移动失败: ' + e.message)
    }
  }
}

async function handleDelete(project: GitProject) {
  if (confirm(`确定要删除项目 "${project.name}" 吗?`)) {
    try {
      await projectStore.deleteProject(project.id)
    } catch (e: any) {
      alert('删除失败: ' + e.message)
    }
  }
}

function handleDragStart(project: GitProject, index: number, event: DragEvent) {
  draggedItem.value = { project, index }
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

function handleDragOver(event: DragEvent) {
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

async function handleDrop(targetProject: GitProject, targetIndex: number) {
  if (!draggedItem.value) return

  const { project: draggedProject, index: draggedIndex } = draggedItem.value

  if (draggedProject.id === targetProject.id) {
    draggedItem.value = null
    return
  }

  // Reorder projects
  const newProjects = [...filteredProjects.value]
  newProjects.splice(draggedIndex, 1)
  newProjects.splice(targetIndex, 0, draggedProject)

  // Update sort orders
  const reorderedProjects = newProjects.map((p, i) => ({
    ...p,
    sort_order: i
  }))

  try {
    await projectStore.reorderProjects(reorderedProjects as GitProject[])
  } catch (e: any) {
    alert('排序失败: ' + e.message)
  }

  draggedItem.value = null
}
</script>

<style scoped>
.project-list {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.list-header {
  padding: 15px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  gap: 10px;
  align-items: center;
}

.list-header h3 {
  margin: 0;
  white-space: nowrap;
}

.search-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.projects {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.project-item {
  display: flex;
  align-items: center;
  padding: 10px;
  margin-bottom: 5px;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.project-item:hover {
  background-color: #f5f5f5;
}

.project-item.selected {
  background-color: #e3f2fd;
  border-color: #2196f3;
}

.drag-handle {
  cursor: grab;
  color: #999;
  margin-right: 8px;
}

.project-index {
  color: #666;
  font-size: 12px;
  min-width: 30px;
}

.project-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.project-actions {
  display: none;
  gap: 4px;
}

.project-item:hover .project-actions {
  display: flex;
}

.project-actions button {
  padding: 4px 8px;
  font-size: 14px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 3px;
  cursor: pointer;
}

.project-actions button:hover:not(:disabled) {
  background-color: #f0f0f0;
}

.project-actions button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
</style>
```

**Step 4: 更新 App.vue 使用 ProjectList 组件**

修改 `frontend/src/App.vue`，简化为使用组件:

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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { OpenConfigFolder, SelectProjectFolder } from './wailsjs/go/main/App'
import ProjectList from './components/ProjectList.vue'
import type { GitProject } from './types'

const projectStore = useProjectStore()
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
  } catch (e: any) {
    console.error('Failed to add project:', e)
  }
}

function handleSelectProject(project: GitProject) {
  selectedProjectId.value = project.id
  console.log('Selected project:', project)
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: any) {
    console.error('Failed to open config folder:', e)
  }
}
</script>

<style scoped>
/* ... existing styles ... */
</style>
```

**Step 5: 测试拖拽排序**

运行 `wails dev`，测试：
- 点击上移/下移按钮
- 拖拽项目到新位置
- 刷新页面，确认排序保存

**Step 6: 提交排序功能**

```bash
git add .
git commit -m "feat: add project sorting with drag-and-drop

- Implement MoveProject and ReorderProjects backend methods
- Create ProjectList component with drag-and-drop support
- Add up/down buttons for alternative sorting
- Implement search filter for projects
"
```

---

## 后续阶段概述

由于计划篇幅较长，以下是后续阶段的简要概述：

### 阶段 4: AI 集成
- 集成 ai-commit 的核心逻辑
- 实现 commit 消息生成
- 添加重新生成功能

### 阶段 5: Commit 详情面板
- 创建左右分栏布局
- 显示项目信息和 diff
- 实现 commit 预览

### 阶段 6: Git 操作
- 实现 git commit 执行
- 添加分支和文件信息显示
- 处理错误情况

### 阶段 7: 完善与优化
- 添加配置文件支持
- 优化错误处理
- 添加 loading 状态

---

## 测试指南

### 运行测试

```bash
# Go 后端测试
go test ./... -v

# 前端测试 (如果配置了)
cd frontend
npm test
```

### 手动测试清单

- [ ] 添加项目（选择文件夹）
- [ ] 删除项目
- [ ] 项目上移/下移
- [ ] 拖拽排序
- [ ] 搜索过滤
- [ ] 刷新页面，数据持久化

---

## 注意事项

1. **每次修改 Go 代码后需要重启 wails dev**
2. **Wails bindings 自动生成在 frontend/wailsjs/ 目录**
3. **TypeScript 类型需要与 Go 结构体保持一致**
4. **错误处理要友好，使用中文提示**
5. **遵循 TDD 原则，先写测试再实现功能**

---

**计划完成！** 下一步可以选择：

1. **继续编写详细实施步骤** (阶段 4-7)
2. **开始执行计划** - 使用 superpowers:executing-plans
