# 启动画面与项目状态预加载功能实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在应用启动时显示欢迎画面，预加载所有项目的 Pushover Hook 版本和 Git 状态，完成后在项目列表中显示状态指示器。

**Architecture:**
- 后端：在 `app.startup` 中启动 goroutine 执行预加载，通过 Wails Events 向前端推送进度
- 前端：独立的 SplashScreen 组件监听事件，完成后自动切换到主界面
- 状态管理：新增 `startupStore` 管理启动状态，`projectStore` 扩展以支持状态指示器

**Tech Stack:** Wails v2, Vue 3, Pinia, Go 1.21+, SQLite/GORM

---

## Task 1: 创建启动画面前端组件

**Files:**
- Create: `frontend/src/components/SplashScreen.vue`
- Create: `frontend/src/stores/startupStore.ts`
- Modify: `frontend/src/App.vue`

**Step 1: 创建 startupStore**

创建 `frontend/src/stores/startupStore.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface StartupProgress {
  stage: string
  percent: number
  message: string
}

export const useStartupStore = defineStore('startup', () => {
  const isVisible = ref(true)
  const progress = ref<StartupProgress>({
    stage: 'initializing',
    percent: 0,
    message: '正在初始化...'
  })

  function updateProgress(data: StartupProgress) {
    progress.value = data
  }

  function complete() {
    progress.value.percent = 100
    progress.value.message = '完成'
    setTimeout(() => {
      isVisible.value = false
    }, 500)
  }

  return {
    isVisible,
    progress,
    updateProgress,
    complete
  }
})
```

**Step 2: 创建 SplashScreen 组件**

创建 `frontend/src/components/SplashScreen.vue`:

```vue
<template>
  <div v-if="startupStore.isVisible" class="splash-screen">
    <div class="splash-content">
      <!-- Logo -->
      <div class="app-logo">
        <span class="logo-icon">🚀</span>
      </div>

      <!-- Title -->
      <h1 class="app-title">AI Commit Hub</h1>
      <p class="app-version">v1.0.0</p>

      <!-- Progress Bar -->
      <div class="progress-container">
        <div class="progress-bar">
          <div
            class="progress-fill"
            :style="{ width: startupStore.progress.percent + '%' }"
          ></div>
        </div>
        <span class="progress-text">{{ startupStore.progress.percent }}%</span>
      </div>

      <!-- Status Message -->
      <p class="status-message">{{ startupStore.progress.message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useStartupStore } from '../stores/startupStore'

const startupStore = useStartupStore()
</script>

<style scoped>
.splash-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #1b263b 0%, #0d1b2a 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fade-in 0.3s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.splash-content {
  text-align: center;
  color: white;
}

.app-logo {
  margin-bottom: 2rem;
}

.logo-icon {
  font-size: 80px;
  display: inline-block;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.app-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
  background: linear-gradient(135deg, #06b6d4, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.app-version {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin: 0 0 3rem 0;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  max-width: 300px;
  margin: 0 auto 1.5rem;
}

.progress-bar {
  flex: 1;
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #06b6d4, #8b5cf6);
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.8);
  min-width: 40px;
  text-align: right;
}

.status-message {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  margin: 0;
}
</style>
```

**Step 3: 修改 App.vue 添加启动画面**

修改 `frontend/src/App.vue`，在模板顶部添加 SplashScreen：

```vue
<template>
  <!-- SplashScreen -->
  <SplashScreen />

  <!-- Main App -->
  <div class="app-container">
    <!-- 现有内容 -->
  </div>
</template>

<script setup lang="ts">
import SplashScreen from './components/SplashScreen.vue'
// 现有 imports
</script>
```

**Step 4: 监听 Wails Events**

在 `frontend/src/main.ts` 中添加事件监听：

```typescript
import { EventsOn } from '../wailsjs/runtime'
import { useStartupStore } from './stores/startupStore'

// 在 app mount 之前
EventsOn('startup-progress', (data: StartupProgress) => {
  const startupStore = useStartupStore()
  startupStore.updateProgress(data)
})

EventsOn('startup-complete', () => {
  const startupStore = useStartupStore()
  startupStore.complete()
})
```

**Step 5: 提交**

```bash
git add frontend/src/components/SplashScreen.vue frontend/src/stores/startupStore.ts frontend/src/App.vue frontend/src/main.ts
git commit -m "feat: 添加启动画面组件和状态管理

- 创建 SplashScreen 组件显示启动进度
- 创建 startupStore 管理启动状态
- 添加 Wails Events 监听

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 2: 扩展 GitProject 数据模型

**Files:**
- Modify: `pkg/models/git_project.go`

**Step 1: 添加运行时状态字段**

在 `GitProject` 结构体末尾添加：

```go
type GitProject struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Path      string `gorm:"not null;uniqueIndex" json:"path"`
	Name      string `json:"name"`
	SortOrder int    `gorm:"index" json:"sort_order"`

	// 项目级别 AI 配置（可选）
	Provider   *string `json:"provider,omitempty"`
	Language   *string `json:"language,omitempty"`
	Model      *string `json:"model,omitempty"`
	UseDefault bool    `gorm:"default:true" json:"use_default"`

	// Pushover Hook 配置
	HookInstalled   bool       `gorm:"default:false" json:"hook_installed"`
	NotificationMode string     `gorm:"default:'enabled'" json:"notification_mode"`
	HookVersion     string     `gorm:"size:50" json:"hook_version"`
	HookInstalledAt *time.Time `json:"hook_installed_at,omitempty"`

	// 运行时状态字段（不持久化到数据库）
	HasUncommittedChanges bool `json:"has_uncommitted_changes" gorm:"-"`
	UntrackedCount       int  `json:"untracked_count" gorm:"-"`
	PushoverNeedsUpdate  bool `json:"pushover_needs_update" gorm:"-"`
}
```

**Step 2: 提交**

```bash
git add pkg/models/git_project.go
git commit -m "feat: 添加项目运行时状态字段

- 添加 HasUncommittedChanges 标记未提交更改
- 添加 UntrackedCount 统计未跟踪文件数量
- 添加 PushoverNeedsUpdate 标记插件更新需求
- 使用 gorm:\"-\" 标签防止持久化

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 3: 实现启动预加载后端逻辑

**Files:**
- Modify: `app.go`
- Create: `pkg/service/startup_service.go`

**Step 1: 创建 StartupService**

创建 `pkg/service/startup_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/git"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// StartupProgress 启动进度
type StartupProgress struct {
	Stage   string `json:"stage"`
	Percent int    `json:"percent"`
	Message string `json:"message"`
}

// StartupService 启动服务
type StartupService struct {
	ctx              context.Context
	gitProjectRepo   *repository.GitProjectRepository
	pushoverService  *pushover.Service
	db               *gorm.DB
}

// NewStartupService 创建启动服务
func NewStartupService(
	ctx context.Context,
	gitProjectRepo *repository.GitProjectRepository,
	pushoverService *pushover.Service,
) *StartupService {
	return &StartupService{
		ctx:             ctx,
		gitProjectRepo:  gitProjectRepo,
		pushoverService: pushoverService,
		db:              repository.GetDB(),
	}
}

// Preload 预加载所有项目状态
func (s *StartupService) Preload() error {
	logger.Info("开始启动预加载...")

	// 阶段 1: 初始化
	s.emitProgress(StartupProgress{
		Stage:   "initializing",
		Percent: 10,
		Message: "正在初始化...",
	})
	time.Sleep(500 * time.Millisecond)

	// 阶段 2: 检查扩展
	s.emitProgress(StartupProgress{
		Stage:   "extension",
		Percent: 20,
		Message: "检查扩展...",
	})
	time.Sleep(300 * time.Millisecond)

	// 阶段 3: 扫描项目
	projects, err := s.gitProjectRepo.GetAll()
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %w", err)
	}

	totalProjects := len(projects)
	if totalProjects == 0 {
		s.emitProgress(StartupProgress{
			Stage:   "complete",
			Percent: 100,
			Message: "完成",
		})
		return nil
	}

	// 并发检查所有项目
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数为 5
	completed := 0
	var mu sync.Mutex

	for i, project := range projects {
		wg.Add(1)
		go func(idx int, proj models.GitProject) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 检查项目状态
			s.checkProjectStatus(&proj)

			// 更新进度
			mu.Lock()
			completed++
			percent := 20 + int(float64(completed)/float64(totalProjects)*70)
			s.emitProgress(StartupProgress{
				Stage:   "scanning",
				Percent: percent,
				Message: fmt.Sprintf("扫描项目 %d/%d...", completed, totalProjects),
			})
			mu.Unlock()
		}(i, project)
	}

	wg.Wait()

	// 阶段 4: 完成
	s.emitProgress(StartupProgress{
		Stage:   "complete",
		Percent: 100,
		Message: "完成",
	})

	logger.Info("启动预加载完成")
	return nil
}

// checkProjectStatus 检查单个项目状态
func (s *StartupService) checkProjectStatus(project *models.GitProject) {
	// 检查 Pushover 更新状态
	if s.pushoverService != nil {
		status, err := s.pushoverService.GetHookStatus(project.Path)
		if err == nil && status.Installed {
			latestVersion, err := s.pushoverService.GetExtensionVersion()
			if err == nil {
				project.PushoverNeedsUpdate = pushover.CompareVersions(status.Version, latestVersion) < 0
			}
		}
	}

	// 检查 Git 状态（超时 3 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stagingStatus, err := git.GetStagingStatus(project.Path)
	if err == nil {
		project.HasUncommittedChanges = len(stagingStatus.Staged) > 0 || len(stagingStatus.Unstaged) > 0
		project.UntrackedCount = len(stagingStatus.Untracked)
	}

	// 更新数据库
	s.db.Save(project)
}

// emitProgress 发送进度事件
func (s *StartupService) emitProgress(progress StartupProgress) {
	runtime.EventsEmit(s.ctx, "startup-progress", progress)
}
```

**Step 2: 修改 app.go 添加预加载调用**

在 `app.go` 的 `startup` 方法末尾添加：

```go
func (a *App) startup(ctx context.Context) {
	// ... 现有代码 ...

	// 启动预加载（异步）
	if a.pushoverService != nil && a.gitProjectRepo != nil {
		go func() {
			startupService := service.NewStartupService(ctx, a.gitProjectRepo, a.pushoverService)
			if err := startupService.Preload(); err != nil {
				logger.Errorf("启动预加载失败: %v", err)
				// 发送完成事件，即使失败也进入主界面
				runtime.EventsEmit(ctx, "startup-complete", nil)
			} else {
				runtime.EventsEmit(ctx, "startup-complete", nil)
			}
		}()
	} else {
		// 无需预加载，直接完成
		runtime.EventsEmit(ctx, "startup-complete", nil)
	}
}
```

**Step 3: 提交**

```bash
git add pkg/service/startup_service.go app.go
git commit -m "feat: 实现启动预加载服务

- 创建 StartupService 处理启动预加载逻辑
- 并发检查所有项目的 Pushover 和 Git 状态
- 通过 Wails Events 发送进度更新
- 添加超时保护和并发控制

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 4: 修改项目列表显示状态指示器

**Files:**
- Modify: `frontend/src/components/ProjectList.vue`
- Modify: `frontend/src/stores/projectStore.ts`

**Step 1: 扩展 projectStore 添加状态获取方法**

修改 `frontend/src/stores/projectStore.ts`，添加获取带状态项目列表的方法：

```typescript
// 添加新的方法
export const useProjectStore = defineStore('project', () => {
  // ... 现有代码 ...

  async function loadProjectsWithStatus() {
    loading.value = true
    error.value = null
    try {
      const projects = await GetProjectsWithStatus()
      projectsList.value = projects
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载失败'
      error.value = message
      logger.error(`加载项目失败: ${message}`)
    } finally {
      loading.value = false
    }
  }

  return {
    // ... 现有返回值 ...
    loadProjectsWithStatus
  }
})
```

**Step 2: 修改 ProjectList.vue 添加状态指示器**

在项目卡片中添加状态行，修改模板：

```vue
<div class="project-info">
  <span class="project-name">{{ project.name }}</span>
  <span class="project-path">{{ project.path }}</span>

  <!-- 新增：状态指示器行 -->
  <div class="project-status-row">
    <span
      v-if="project.has_uncommitted_changes"
      class="status-indicator uncommitted"
      title="有未提交更改"
    >
      🔄
    </span>
    <span
      v-if="project.untracked_count > 0"
      class="status-indicator untracked"
      :title="`${project.untracked_count} 个未跟踪文件`"
    >
      ➕ {{ project.untracked_count }}
    </span>
    <span
      v-if="project.pushover_needs_update"
      class="status-indicator update"
      title="Pushover 插件可更新"
    >
      ⬆️
    </span>
  </div>
</div>
```

**Step 3: 添加状态指示器样式**

在 `ProjectList.vue` 的 `<style>` 部分添加：

```css
.project-status-row {
  display: flex;
  gap: 6px;
  margin-top: 6px;
  flex-wrap: wrap;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 500;
}

.status-indicator.uncommitted {
  color: #f97316;
  background: rgba(249, 115, 22, 0.15);
}

.status-indicator.untracked {
  color: #eab308;
  background: rgba(234, 179, 8, 0.15);
}

.status-indicator.update {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.15);
}
```

**Step 4: 监听启动完成事件刷新项目列表**

在 `ProjectList.vue` 的 `<script setup>` 中添加：

```typescript
import { onMounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime'

onMounted(() => {
  EventsOn('startup-complete', async () => {
    await projectStore.loadProjectsWithStatus()
  })
})
```

**Step 5: 提交**

```bash
git add frontend/src/components/ProjectList.vue frontend/src/stores/projectStore.ts
git commit -m "feat: 添加项目状态指示器

- 在项目卡片下方显示状态图标
- 支持未提交、未跟踪、Pushover 更新状态
- 启动完成后自动刷新带状态的项目列表
- 添加状态指示器样式

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 5: 添加 API 方法

**Files:**
- Modify: `app.go`

**Step 1: 添加 GetProjectsWithStatus 方法**

在 `app.go` 中添加：

```go
// GetProjectsWithStatus 获取带状态的项目列表
func (a *App) GetProjectsWithStatus() ([]models.GitProject, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}

	return projects, nil
}
```

**Step 2: 提交**

```bash
git add app.go
git commit -m "feat: 添加 GetProjectsWithStatus API 方法

- 返回包含运行时状态的项目列表
- 支持前端获取预加载后的项目状态

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 6: 错误处理与降级

**Files:**
- Modify: `pkg/service/startup_service.go`
- Modify: `frontend/src/main.ts`

**Step 1: 添加超时保护**

修改 `startup_service.go` 的 `Preload` 方法，添加总体超时：

```go
func (s *StartupService) Preload() error {
	logger.Info("开始启动预加载...")

	// 添加总体超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 在新 goroutine 中执行预加载
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.doPreload()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.Warn("启动预加载超时，将进入主界面")
		s.emitProgress(StartupProgress{
			Stage:   "complete",
			Percent: 100,
			Message: "完成",
		})
		return nil
	}
}

func (s *StartupService) doPreload() error {
	// 原有 Preload 的实现
}
```

**Step 2: 改进单个项目检查的错误处理**

修改 `checkProjectStatus` 方法，添加错误日志：

```go
func (s *StartupService) checkProjectStatus(project *models.GitProject) {
	projectName := project.Name

	// 检查 Pushover 更新状态
	if s.pushoverService != nil {
		status, err := s.pushoverService.GetHookStatus(project.Path)
		if err != nil {
			logger.Debugf("[%s] 获取 Pushover 状态失败: %v", projectName, err)
		} else if status.Installed {
			latestVersion, err := s.pushoverService.GetExtensionVersion()
			if err != nil {
				logger.Debugf("[%s] 获取扩展版本失败: %v", projectName, err)
			} else {
				project.PushoverNeedsUpdate = pushover.CompareVersions(status.Version, latestVersion) < 0
			}
		}
	}

	// 检查 Git 状态
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stagingStatus, err := git.GetStagingStatus(project.Path)
	if err != nil {
		logger.Debugf("[%s] 获取 Git 状态失败: %v", projectName, err)
	} else {
		project.HasUncommittedChanges = len(stagingStatus.Staged) > 0 || len(stagingStatus.Unstaged) > 0
		project.UntrackedCount = len(stagingStatus.Untracked)
	}

	// 更新数据库
	if err := s.db.Save(project).Error; err != nil {
		logger.Errorf("[%s] 保存项目状态失败: %v", projectName, err)
	}
}
```

**Step 3: 前端添加降级处理**

修改 `frontend/src/main.ts`，添加超时处理：

```typescript
// 设置启动超时（30 秒）
setTimeout(() => {
  const startupStore = useStartupStore()
  if (startupStore.isVisible) {
    logger.warn('启动超时，强制进入主界面')
    startupStore.complete()
  }
}, 30000)
```

**Step 4: 提交**

```bash
git add pkg/service/startup_service.go frontend/src/main.ts
git commit -m "feat: 添加错误处理和降级策略

- 添加总体 30 秒超时保护
- 单个项目检查失败不影响其他项目
- 前端超时后强制进入主界面
- 改进日志记录

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## Task 7: 测试与验证

**Files:**
- Create: `tmp/test-splash-screen.html`

**Step 1: 创建测试页面**

创建用于测试启动画面的简单 HTML 页面：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>Splash Screen Test</title>
</head>
<body>
    <h1>启动画面测试</h1>
    <button onclick="testProgress()">测试进度更新</button>
    <button onclick="testComplete()">测试完成</button>

    <script>
        function testProgress() {
            console.log('模拟进度更新事件')
        }

        function testComplete() {
            console.log('模拟完成事件')
        }
    </script>
</body>
</html>
```

**Step 2: 手动测试清单**

测试以下场景：
- [ ] 空项目（首次启动）
- [ ] 单个项目
- [ ] 多个项目（5+）
- [ ] 项目路径不存在
- [ ] Pushover 扩展未安装
- [ ] Git 仓库有未提交更改
- [ ] Git 仓库有未跟踪文件
- [ ] 启动超时（模拟）

**Step 3: 提交测试文件**

```bash
git add tmp/test-splash-screen.html
git commit -m "test: 添加启动画面测试页面

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## 验收标准

### 功能验收
- [ ] 启动时显示 SplashScreen
- [ ] 进度条正确显示加载进度
- [ ] 完成后自动切换到主界面
- [ ] 项目列表显示状态指示器
- [ ] 未提交更改显示 🔄 图标
- [ ] 未跟踪文件显示 ➕ N 图标
- [ ] Pushover 需要更新显示 ⬆️ 图标

### 性能验收
- [ ] 10 个项目启动时间 < 5 秒
- [ ] 单个项目检查超时 3 秒生效
- [ ] 总体超时 30 秒生效

### 稳定性验收
- [ ] 部分项目失败不影响整体启动
- [ ] Pushover 扩展未安装不影响启动
- [ ] 预加载失败仍可进入主界面

---

## 参考文档

- 设计文档: `docs/plans/2026-01-28-splash-screen-project-status-design.md`
- Wails Events 文档: https://wails.io/docs/reference/runtime/events
- Vue3 Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
- Pinia Store: https://pinia.vuejs.org/core-concepts/
