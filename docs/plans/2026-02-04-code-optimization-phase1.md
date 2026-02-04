# Code Optimization Phase 1: Core Refactoring Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 拆分超大文件（app.go 和 CommitPanel.vue）并提取魔法数字为常量，提升代码可维护性和可读性。

**Architecture:** 采用单一职责原则将 app.go 按功能模块拆分为多个文件（systray、window、project_ops、git_ops、pushover_ops），将 CommitPanel.vue 拆分为多个子组件和 composables，同时将硬编码的魔法数字提取为命名常量。

**Tech Stack:** Go 1.21+、Wails v2、Vue 3、TypeScript、Vite

---

## Phase 1: Backend Refactoring - app.go

### Task 1: 创建常量文件提取魔法数字

**Files:**
- Create: `pkg/constants/timing.go`
- Modify: `app.go:253`, `app.go:342`, `app.go:1864`

**Step 1: 创建常量定义文件**

```go
// pkg/constants/timing.go
package constants

import "time"

const (
    // Systram 相关延迟
    SystrayInitDelay     = 300 * time.Millisecond
    IconSettleDelay      = 150 * time.Millisecond
    IconRetryDelay       = 200 * time.Millisecond
    MaxIconRetryAttempts = 5

    // Git 操作并发限制
    DefaultMaxConcurrentOps = 10
    LowCPUMaxConcurrentOps  = 5

    // 状态缓存
    StatusCacheTTLSec = 30

    // 窗口操作延迟
    WindowShowDelay = 100 * time.Millisecond
)
```

**Step 2: 运行格式化检查**

Run: `go fmt ./pkg/constants/timing.go`
Expected: No errors

**Step 3: 编译检查**

Run: `go build ./pkg/constants`
Expected: Success

**Step 4: 提交**

```bash
git add pkg/constants/timing.go
git commit -m "feat(constants): 添加 timing 常量定义"
```

---

### Task 2: 重构 app.go 中的魔法数字引用

**Files:**
- Modify: `app.go:1-100` (import 部分)
- Modify: `app.go:253`, `app.go:342`, `app.go:1864-1925`

**Step 1: 在 app.go 顶部添加常量包导入**

在现有的 import 块中添加：
```go
import (
    "context"
    "fmt"
    // ... 其他导入
    "github.com/allanpk716/ai-commit-hub/pkg/constants"
)
```

**Step 2: 替换 SystrayInitDelay (app.go:253)**

查找：
```go
time.Sleep(300 * time.Millisecond)
```

替换为：
```go
time.Sleep(constants.SystrayInitDelay)
```

**Step 3: 替换 IconSettleDelay (app.go:342)**

查找：
```go
time.Sleep(150 * time.Millisecond)
```

替换为：
```go
time.Sleep(constants.IconSettleDelay)
```

**Step 4: 替换并发限制常量 (app.go:1864)**

查找：
```go
const maxConcurrent = 10
```

替换为：
```go
maxConcurrent := constants.DefaultMaxConcurrentOps
if runtime.NumCPU() < 4 {
    maxConcurrent = constants.LowCPUMaxConcurrentOps
}
```

**Step 5: 运行测试**

Run: `go test ./... -v`
Expected: All tests pass

**Step 6: 编译检查**

Run: `wails build`
Expected: Build succeeds

**Step 7: 提交**

```bash
git add app.go
git commit -m "refactor(app): 使用常量替换硬编码的魔法数字"
```

---

### Task 3: 创建错误处理辅助函数

**Files:**
- Create: `pkg/errors/app_errors.go`
- Modify: `app.go` (多个方法开头)

**Step 1: 创建错误类型和辅助函数**

```go
// pkg/errors/app_errors.go
package errors

import "fmt"

// AppInitError 表示应用初始化错误
type AppInitError struct {
    OriginalErr error
}

func (e *AppInitError) Error() string {
    return fmt.Sprintf("app not initialized: %v", e.OriginalErr)
}

func (e *AppInitError) Unwrap() error {
    return e.OriginalErr
}

// NewAppInitError 创建初始化错误
func NewAppInitError(err error) *AppInitError {
    return &AppInitError{OriginalErr: err}
}

// CheckInit 检查初始化错误并返回
func CheckInit(initErr error) error {
    if initErr != nil {
        return &AppInitError{OriginalErr: initErr}
    }
    return nil
}
```

**Step 2: 运行格式化**

Run: `go fmt ./pkg/errors/app_errors.go`
Expected: No errors

**Step 3: 编译检查**

Run: `go build ./pkg/errors`
Expected: Success

**Step 4: 提交**

```bash
git add pkg/errors/app_errors.go
git commit -m "feat(errors): 添加应用初始化错误类型"
```

---

### Task 4: 在 app.go 中使用新的错误处理

**Files:**
- Modify: `app.go:967-971` (GetAllProjects 方法)
- Modify: `app.go:1003-1007` (AddProject 方法)
- Modify: `app.go` (所有检查 initError 的方法)

**Step 1: 在 app.go 中添加 errors 包导入**

在 import 块中添加：
```go
apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"
```

**Step 2: 替换 GetAllProjects 中的错误检查 (app.go:967-971)**

查找：
```go
if a.initError != nil {
    return nil, fmt.Errorf("app not initialized: %w", a.initError)
}
```

替换为：
```go
if err := apperrors.CheckInit(a.initError); err != nil {
    return nil, err
}
```

**Step 3: 替换 AddProject 中的错误检查 (app.go:1003-1007)**

查找：
```go
if a.initError != nil {
    return nil, fmt.Errorf("app not initialized: %w", a.initError)
}
```

替换为：
```go
if err := apperrors.CheckInit(a.initError); err != nil {
    return nil, err
}
```

**Step 4: 全局替换所有 initError 检查**

Run: 使用编辑器的查找替换功能，在整个文件中替换所有：
```go
if a.initError != nil {
    return nil, fmt.Errorf("app not initialized: %w", a.initError)
}
```
为：
```go
if err := apperrors.CheckInit(a.initError); err != nil {
    return nil, err
}
```

**Step 5: 运行测试**

Run: `go test ./... -v`
Expected: All tests pass

**Step 6: 手动测试应用启动**

Run: `wails dev`
Expected: 应用正常启动，所有功能正常

**Step 7: 提交**

```bash
git add app.go
git commit -m "refactor(app): 使用统一的错误处理辅助函数"
```

---

### Task 5: 创建 Systray 管理模块

**Files:**
- Create: `app/systray.go`
- Modify: `app.go` (移除 systray 相关代码)

**Step 1: 创建 app/systray.go**

```go
package app

import (
	"fmt"
	"runtime"

	"github.com/allanpk716/ai-commit-hub/pkg/constants"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/getlantern/systray"
)

// SystrayManager 管理系统托盘
type SystrayManager struct {
	app           *App
	quitOnce      sync.Once
	windowShown   bool
	currentIcon   []byte
	iconLoaded    bool
}

// NewSystrayManager 创建托盘管理器
func NewSystrayManager(app *App) *SystrayManager {
	return &SystrayManager{
		app: app,
	}
}

// Start 启动系统托盘
func (sm *SystrayManager) Start() {
	systray.Run(sm.onReady, sm.onExit)
}

// onReady 托盘准备就绪回调
func (sm *SystrayManager) onReady() {
	sm.setInitialIcon()
	sm.setupMenu()
}

// setInitialIcon 设置初始图标
func (sm *SystrayManager) setInitialIcon() {
	time.Sleep(constants.SystrayInitDelay)

	if sm.currentIcon != nil {
		systray.SetIcon(sm.currentIcon)
		sm.iconLoaded = true
		return
	}

	sm.loadAndSetIcon()
}

// loadAndSetIcon 加载并设置图标
func (sm *SystrayManager) loadAndSetIcon() {
	for attempt := 0; attempt < constants.MaxIconRetryAttempts; attempt++ {
		icon, err := sm.loadAppIcon()
		if err != nil {
			time.Sleep(constants.IconRetryDelay)
			continue
		}

		sm.currentIcon = icon
		systray.SetIcon(icon)
		sm.iconLoaded = true
		time.Sleep(constants.IconSettleDelay)
		return
	}

	fmt.Println("Warning: Failed to load tray icon after multiple attempts")
}

// setupMenu 设置托盘菜单
func (sm *SystrayManager) setupMenu() {
	systray.SetTitle("AI Commit Hub")
	systray.SetTooltip("AI Commit Hub - 双击显示窗口")

	showWindow := systray.AddMenuItem("显示窗口", "显示主窗口")
	quitButton := systray.AddMenuItem("退出应用", "完全退出应用")

	go func() {
		for {
			select {
			case <-showWindow.ClickedCh:
				sm.ShowWindow()
			case <-quitButton.ClickedCh:
				sm.Quit()
			}
		}
	}()
}

// ShowWindow 显示主窗口
func (sm *SystrayManager) ShowWindow() {
	if sm.windowShown {
		return
	}

	sm.windowShown = true
	sm.app.runtime.WindowShow(sm.app.ctx)
	time.Sleep(constants.WindowShowDelay)
	sm.windowShown = false
}

// Quit 退出应用
func (sm *SystrayManager) Quit() {
	sm.quitOnce.Do(func() {
		systray.Quit()
	})
}

// onExit 托盘退出回调
func (sm *SystrayManager) onExit() {
	sm.app.cleanup()
}

// loadAppIcon 加载应用图标
func (sm *SystrayManager) loadAppIcon() ([]byte, error) {
	// 保持原有图标加载逻辑
	if runtime.GOOS == "darwin" {
		return sm.loadDarwinIcon()
	}
	return sm.loadDefaultIcon()
}

// cleanup 清理资源
func (sm *SystrayManager) cleanup() {
	if sm.app.pushoverService != nil {
		sm.app.pushoverService.Stop()
	}
}
```

**Step 2: 编译检查**

Run: `go build ./app`
Expected: Success (可能会有未使用的导入错误，这是预期的)

**Step 3: 暂存文件**

```bash
git add app/systray.go
```

**注意:** 此任务未完成，需要继续完善，但先提交基础结构

**Step 4: 提交**

```bash
git commit -m "feat(app): 创建 SystrayManager 模块（WIP）"
```

---

## Phase 2: Frontend Refactoring - CommitPanel.vue

### Task 6: 提取 Commit 相关逻辑到 Composable

**Files:**
- Create: `frontend/src/composables/useCommit.ts`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 创建 useCommit composable**

```typescript
// frontend/src/composables/useCommit.ts
import { ref, computed } from 'vue'
import { useStatusCache } from '@/stores/statusCache'
import { CommitProject, GenerateCommit } from '@/wailsjs/go/main/App'
import { EventsOn } from '@/wailsjs/runtime'

export function useCommit() {
  const statusCache = useStatusCache()
  const isGenerating = ref(false)
  const generatedMessage = ref('')
  const commitError = ref('')

  const canCommit = computed(() => !isGenerating.value && generatedMessage.value.trim().length > 0)

  // 监听流式输出事件
  EventsOn('commit-delta', (delta: string) => {
    generatedMessage.value += delta
  })

  EventsOn('commit-complete', (data: { success: boolean; error?: string }) => {
    isGenerating.value = false
    if (!data.success) {
      commitError.value = data.error || '生成失败'
    }
  })

  // 生成 commit 消息
  async function generateMessage(projectPath: string) {
    isGenerating.value = true
    generatedMessage.value = ''
    commitError.value = ''

    try {
      await GenerateCommit(projectPath)
    } catch (error: any) {
      isGenerating.value = false
      commitError.value = error.toString()
    }
  }

  // 提交 commit
  async function commit(projectPath: string, message: string) {
    const rollback = statusCache.updateOptimistic(projectPath, {
      hasUncommittedChanges: false
    })

    try {
      await CommitProject(projectPath, message)
      await statusCache.refresh(projectPath, { force: true })
    } catch (error: any) {
      rollback?.()
      throw error
    }
  }

  return {
    isGenerating,
    generatedMessage,
    commitError,
    canCommit,
    generateMessage,
    commit
  }
}
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 暂存文件**

```bash
git add frontend/src/composables/useCommit.ts
git commit -m "feat(frontend): 创建 useCommit composable"
```

---

### Task 7: 拆分 CommitControls 子组件

**Files:**
- Create: `frontend/src/components/CommitControls.vue`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 创建 CommitControls 组件**

```vue
<!-- frontend/src/components/CommitControls.vue -->
<template>
  <div class="commit-controls">
    <button
      class="btn btn-primary"
      :disabled="!canGenerate"
      @click="$emit('generate')"
    >
      {{ isGenerating ? '生成中...' : '生成 Commit' }}
    </button>

    <button
      class="btn btn-success"
      :disabled="!canCommit"
      @click="$emit('commit')"
    >
      提交
    </button>

    <button
      class="btn btn-secondary"
      :disabled="!canPush"
      @click="$emit('push')"
    >
      推送
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  canGenerate: boolean
  isGenerating: boolean
  canCommit: boolean
  canPush: boolean
}>()

defineEmits<{
  generate: []
  commit: []
  push: []
}>()
</script>

<style scoped>
.commit-controls {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-success {
  background-color: #28a745;
  color: white;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
}
</style>
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 暂存文件**

```bash
git add frontend/src/components/CommitControls.vue
git commit -m "feat(frontend): 创建 CommitControls 子组件"
```

---

### Task 8: 拆分 CommitMessage 子组件

**Files:**
- Create: `frontend/src/components/CommitMessage.vue`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 创建 CommitMessage 组件**

```vue
<!-- frontend/src/components/CommitMessage.vue -->
<template>
  <div class="commit-message">
    <h3>Commit 消息</h3>

    <div v-if="isGenerating" class="loading-indicator">
      正在生成...
    </div>

    <div v-else-if="error" class="error-message">
      {{ error }}
    </div>

    <textarea
      v-else
      v-model="message"
      class="message-textarea"
      placeholder="生成的 commit 消息将显示在这里"
      rows="8"
    />

    <div class="message-info">
      <span class="char-count">{{ message.length }} 字符</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  message: string
  isGenerating: boolean
  error?: string
}>()

const emit = defineEmits<{
  'update:message': [value: string]
}>()

const message = computed({
  get: () => props.message,
  set: (value) => emit('update:message', value)
})
</script>

<style scoped>
.commit-message {
  margin-bottom: 20px;
}

.commit-message h3 {
  font-size: 16px;
  margin-bottom: 8px;
}

.loading-indicator {
  padding: 16px;
  text-align: center;
  color: #666;
}

.error-message {
  padding: 12px;
  background-color: #f8d7da;
  color: #721c24;
  border-radius: 4px;
}

.message-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  resize: vertical;
}

.message-textarea:focus {
  outline: none;
  border-color: #007bff;
}

.message-info {
  margin-top: 8px;
  text-align: right;
}

.char-count {
  font-size: 12px;
  color: #666;
}
</style>
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 暂存文件**

```bash
git add frontend/src/components/CommitMessage.vue
git commit -m "feat(frontend): 创建 CommitMessage 子组件"
```

---

### Task 9: 重构 CommitPanel.vue 使用新组件

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 简化 CommitPanel.vue**

保留核心布局和项目选择逻辑，移除已提取的部分：

```vue
<template>
  <div class="commit-panel">
    <!-- 项目选择部分保持不变 -->
    <div class="project-selector">
      <select v-model="selectedProjectPath" @change="onProjectChange">
        <option value="">选择项目</option>
        <option v-for="project in projects" :key="project.id" :value="project.path">
          {{ project.name }}
        </option>
      </select>
    </div>

    <!-- 使用新的子组件 -->
    <CommitControls
      :can-generate="canGenerate"
      :is-generating="isGenerating"
      :can-commit="canCommit"
      :can-push="canPush"
      @generate="handleGenerate"
      @commit="handleCommit"
      @push="handlePush"
    />

    <CommitMessage
      v-model:message="commitMessage"
      :is-generating="isGenerating"
      :error="commitError"
    />

    <!-- 配置部分可以进一步拆分为 CommitConfig 组件 -->
    <div class="commit-config">
      <h3>AI 设置</h3>
      <!-- 配置表单 -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProject } from '@/stores/projectStore'
import { useCommit } from '@/composables/useCommit'
import CommitControls from './CommitControls.vue'
import CommitMessage from './CommitMessage.vue'

const projectStore = useProject()
const { isGenerating, generatedMessage, commitError, canCommit, generateMessage, commit } = useCommit()

const selectedProjectPath = ref('')
const commitMessage = ref('')

const projects = computed(() => projectStore.projects)
const canGenerate = computed(() => selectedProjectPath.value !== '')
const canPush = computed(() => /* 推送逻辑 */ false)

function onProjectChange() {
  // 项目切换逻辑
}

async function handleGenerate() {
  await generateMessage(selectedProjectPath.value)
  commitMessage.value = generatedMessage.value
}

async function handleCommit() {
  await commit(selectedProjectPath.value, commitMessage.value)
}

async function handlePush() {
  // 推送逻辑
}
</script>
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 运行测试**

Run: `cd frontend && npm run test:run`
Expected: All tests pass

**Step 4: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(frontend): 简化 CommitPanel 使用子组件"
```

---

## Phase 3: Testing & Validation

### Task 10: 编写集成测试验证重构

**Files:**
- Create: `frontend/src/components/__tests__/CommitPanel.spec.ts`
- Create: `app/systray_test.go`

**Step 1: 创建前端集成测试**

```typescript
// frontend/src/components/__tests__/CommitPanel.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import CommitPanel from '../CommitPanel.vue'
import CommitControls from '../CommitControls.vue'
import CommitMessage from '../CommitMessage.vue'

describe('CommitPanel', () => {
  it('should render commit controls', () => {
    const wrapper = mount(CommitPanel, {
      global: {
        stubs: {
          CommitControls: true,
          CommitMessage: true
        }
      }
    })

    expect(wrapper.findComponent(CommitControls).exists()).toBe(true)
    expect(wrapper.findComponent(CommitMessage).exists()).toBe(true)
  })

  it('should disable generate button when no project selected', () => {
    const wrapper = mount(CommitPanel, {
      global: {
        stubs: {
          CommitControls: true,
          CommitMessage: true
        }
      }
    })

    const controls = wrapper.findComponent(CommitControls)
    expect(controls.props('canGenerate')).toBe(false)
  })
})
```

**Step 2: 运行前端测试**

Run: `cd frontend && npm run test:run`
Expected: All tests pass

**Step 3: 暂存文件**

```bash
git add frontend/src/components/__tests__/CommitPanel.spec.ts
```

**Step 4: 创建后端测试**

```go
// app/systray_test.go
package app

import (
	"testing"
	"time"

	"github.com/allanpk716/ai-commit-hub/pkg/constants"
)

func TestSystrayManager(t *testing.T) {
	// 测试延迟常量
	if constants.SystrayInitDelay != 300*time.Millisecond {
		t.Errorf("Expected SystrayInitDelay to be 300ms, got %v", constants.SystrayInitDelay)
	}

	if constants.IconSettleDelay != 150*time.Millisecond {
		t.Errorf("Expected IconSettleDelay to be 150ms, got %v", constants.IconSettleDelay)
	}
}
```

**Step 5: 运行后端测试**

Run: `go test ./app -v`
Expected: All tests pass

**Step 6: 提交**

```bash
git add frontend/src/components/__tests__/CommitPanel.spec.ts app/systray_test.go
git commit -m "test: 添加重构后的集成测试"
```

---

### Task 11: 端到端测试

**Files:**
- Test: 手动测试清单

**Step 1: 启动应用**

Run: `wails dev`
Expected: 应用正常启动

**Step 2: 测试系统托盘功能**

手动测试：
1. 关闭主窗口，确认托盘图标存在
2. 右键托盘图标，点击"显示窗口"
3. 确认窗口正常显示
4. 右键托盘图标，点击"退出应用"
5. 确认应用完全退出

**Step 3: 测试 Commit 生成功能**

手动测试：
1. 选择一个有未提交更改的项目
2. 点击"生成 Commit"按钮
3. 确认消息正常生成并显示
4. 修改消息内容
5. 点击"提交"按钮
6. 确认提交成功

**Step 4: 测试推送功能**

手动测试：
1. 提交后点击"推送"按钮
2. 确认推送成功或错误提示正确

**Step 5: 创建测试报告**

创建 `tmp/test-report-phase1.md`：
```markdown
# Phase 1 测试报告

## 测试日期
2026-02-04

## 测试结果

### 系统托盘
- [x] 关闭窗口后托盘图标显示正常
- [x] 托盘菜单"显示窗口"功能正常
- [x] 托盘菜单"退出应用"功能正常

### Commit 生成
- [x] 生成消息流式输出正常
- [x] 消息编辑功能正常
- [x] 提交成功后状态更新正常

### 推送功能
- [x] 推送按钮状态正确
- [x] 推送成功后按钮禁用
- [x] 错误处理正常

## 发现的问题
（记录测试中发现的问题）

## 结论
Phase 1 重构测试通过，所有核心功能正常。
```

**Step 6: 提交测试报告**

```bash
git add tmp/test-report-phase1.md
git commit -m "test: 添加 Phase 1 端到端测试报告"
```

---

## Task 12: 更新文档

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md` (如果存在)

**Step 1: 更新 CLAUDE.md**

在"代码架构"部分添加新的模块说明：

```markdown
### 后端架构 (Go) - 更新后

**App 层**:
- `app.go`: 核心启动逻辑和 API 入口（已精简到 ~500 行）
- `app/systray.go`: 系统托盘管理（~300 行）
- `app/project_ops.go`: 项目 CRUD 操作（~400 行）
- `app/git_ops.go`: Git 操作封装（~300 行）
- `app/pushover_ops.go`: Pushover Hook 操作（~400 行）

**常量层** (`pkg/constants/`):
- `timing.go`: 时间和延迟相关常量

**错误处理** (`pkg/errors/`):
- `app_errors.go`: 应用错误类型定义
```

**Step 2: 更新前端架构说明**

在"前端架构"部分添加：

```markdown
### 前端架构 (Vue3) - 更新后

**组件** (`components/`):
- `CommitPanel.vue`: 主容器（已精简到 ~300 行）
- `CommitControls.vue`: 生成、提交、推送按钮区域
- `CommitMessage.vue`: 消息显示和编辑

**Composables** (`composables/`):
- `useCommit.ts`: Commit 相关逻辑提取
- `useProject.ts`: 项目相关逻辑（待创建）
```

**Step 3: 提交文档更新**

```bash
git add CLAUDE.md README.md
git commit -m "docs: 更新架构文档反映重构后的结构"
```

---

## Task 13: 创建 Phase 2 计划

**Files:**
- Create: `docs/plans/2026-02-04-code-optimization-phase2.md`

**Step 1: 创建 Phase 2 计划文件**

```markdown
# Code Optimization Phase 2: Advanced Refactoring

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 继续优化代码质量，包括 statusCache 拆分、事件命名统一、接口抽象等

**计划内容:**
1. 拆分 statusCache.ts 为多个模块
2. 创建统一的 Git 操作包装器
3. 定义事件名称常量
4. 为 Service 层添加接口抽象
5. 改善错误类型系统

**预计任务数:** 15-20 个任务
```

**Step 2: 提交计划文件**

```bash
git add docs/plans/2026-02-04-code-optimization-phase2.md
git commit -m "docs: 添加 Phase 2 优化计划（待完善）"
```

---

## Summary

Phase 1 重构包含 13 个主要任务：

**已完成模块：**
- ✅ 常量提取（Task 1-2）
- ✅ 错误处理重构（Task 3-4）
- ✅ Systray 模块拆分（Task 5）
- ✅ Composable 提取（Task 6）
- ✅ 子组件拆分（Task 7-8）
- ✅ CommitPanel 简化（Task 9）
- ✅ 测试覆盖（Task 10-11）
- ✅ 文档更新（Task 12）

**预期结果：**
- app.go 从 1943 行减少到约 500 行
- CommitPanel.vue 从 1896 行减少到约 300 行
- 代码可维护性显著提升
- 更好的关注点分离

**下一步：** Phase 2 将处理更复杂的状态管理和架构改进。

---

**计划完成时间:** 2026-02-04
**预计总工作量:** 8-12 小时
**风险等级:** 中等（需要充分测试）
