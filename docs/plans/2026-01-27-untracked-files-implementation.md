# 未跟踪文件管理和排除功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 AI Commit Hub 添加未跟踪文件显示和管理功能，支持右键菜单操作（暂存/排除/打开/复制），排除对话框支持精确/扩展名/目录三种模式，目录模式支持智能多层选择。

**Architecture:** 采用 Wails (Go + Vue3) 架构，后端使用 `Command()` 辅助函数执行 Git 命令避免控制台弹窗，前端通过 Wails 绑定调用后端 API，使用 Pinia 管理状态。

**Tech Stack:** Go 1.21+, Vue 3, TypeScript, Wails v2, Pinia, SQLite

---

## 前置知识

### 项目结构
- `pkg/git/`: Git 操作封装层，使用 `Command()` 函数避免 Windows 控制台弹窗
- `app.go`: Wails 应用入口，包含所有导出给前端的 API 方法
- `frontend/src/components/`: Vue 组件
- `frontend/src/stores/`: Pinia 状态管理
- `frontend/src/types/index.ts`: TypeScript 类型定义

### 关键约束
1. **所有 Git 命令必须使用** `pkg/git/cmdhelper.go` 中的 `Command()` 函数，**禁止直接使用** `exec.Command()`
2. 路径格式必须转换为 Git 标准（`/` 分隔符）
3. 规则追加到 `.gitignore`，不覆盖现有内容

---

## Task 1: 后端 - 添加未跟踪文件获取功能

**Files:**
- Modify: `pkg/git/status.go`
- Modify: `frontend/src/types/index.ts`

### Step 1: 在 status.go 添加 UntrackedFile 结构体

打开 `pkg/git/status.go`，在 `StagedFile` 结构体后添加：

```go
type UntrackedFile struct {
	Path string `json:"path"` // 相对于项目根目录的路径
}
```

位置: 第 15 行后（`StagedFile` 结构体定义后）

### Step 2: 在 status.go 添加 GetUntrackedFiles 函数

在 `status.go` 文件末尾添加：

```go
func GetUntrackedFiles(projectPath string) ([]UntrackedFile, error) {
	// 使用 Command() 而不是 exec.Command() 以避免控制台弹窗
	cmd := Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取未跟踪文件失败: %w", err)
	}

	var files []UntrackedFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, UntrackedFile{Path: line})
	}

	return files, nil
}
```

### Step 3: 在 types/index.ts 添加 TypeScript 类型

打开 `frontend/src/types/index.ts`，在 `StagedFile` 接口后添加：

```typescript
export interface UntrackedFile {
  path: string
}
```

位置: 第 41 行后

### Step 4: 提交

```bash
git add pkg/git/status.go frontend/src/types/index.ts
git commit -m "feat: 添加未跟踪文件类型定义和获取函数"
```

---

## Task 2: 后端 - 创建 GitIgnore 操作模块

**Files:**
- Create: `pkg/git/gitignore.go`

### Step 1: 创建 gitignore.go 文件

创建新文件 `pkg/git/gitignore.go`，添加以下内容：

```go
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExcludeMode 排除模式类型
type ExcludeMode string

const (
	ExcludeModeExact      ExcludeMode = "exact"      // 精确文件名
	ExcludeModeExtension  ExcludeMode = "extension"  // 扩展名
	ExcludeModeDirectory  ExcludeMode = "directory"  // 目录
)

// DirectoryOption 目录选项
type DirectoryOption struct {
	Pattern string `json:"pattern"` // .gitignore 模式
	Label   string `json:"label"`   // 显示标签
}

// GetDirectoryOptions 获取目录层级选项
func GetDirectoryOptions(filePath string) []DirectoryOption {
	// 转换为 Git 标准路径格式
	gitPath := toGitPath(filePath)
	parts := strings.Split(gitPath, "/")

	var options []DirectoryOption
	var pathBuilder strings.Builder

	// 构建层级选项（排除文件名）
	for i := 0; i < len(parts)-1; i++ {
		if i > 0 {
			pathBuilder.WriteString("/")
		}
		pathBuilder.WriteString(parts[i])

		pattern := pathBuilder.String()
		options = append(options, DirectoryOption{
			Pattern: pattern,
			Label:   pattern,
		})
	}

	// 添加"目录下所有扩展名"选项
	if len(parts) > 1 {
		dir := pathBuilder.String()
		ext := filepath.Ext(filePath)
		options = append(options, DirectoryOption{
			Pattern: dir + "/*" + ext,
			Label:   dir + "/*" + ext,
		})
	}

	return options
}

// GenerateGitIgnorePattern 生成 .gitignore 规则
func GenerateGitIgnorePattern(filePath string, mode ExcludeMode) (string, error) {
	gitPath := toGitPath(filePath)

	switch mode {
	case ExcludeModeExact:
		return gitPath, nil

	case ExcludeModeExtension:
		ext := filepath.Ext(filePath)
		if ext == "" {
			return "", fmt.Errorf("文件没有扩展名")
		}
		return "*" + ext, nil

	case ExcludeModeDirectory:
		dir := filepath.Dir(filePath)
		if dir == "." || dir == "" {
			return "/", nil
		}
		return toGitPath(dir), nil

	default:
		return "", fmt.Errorf("未知的排除模式: %s", mode)
	}
}

// AddToGitIgnoreFile 添加规则到 .gitignore 文件
func AddToGitIgnoreFile(projectPath, pattern string) error {
	gitIgnorePath := filepath.Join(projectPath, ".gitignore")

	// 读取现有内容
	var content []string
	if data, err := os.ReadFile(gitIgnorePath); err == nil {
		content = strings.Split(string(data), "\n")
	}

	// 检查是否已存在
	pattern = strings.TrimSpace(pattern)
	for _, line := range content {
		if strings.TrimSpace(line) == pattern {
			return nil // 已存在，不重复添加
		}
	}

	// 追加新规则
	content = append(content, pattern, "")
	return os.WriteFile(gitIgnorePath, []byte(strings.Join(content, "\n")), 0644)
}

// toGitPath 转换为 Git 标准路径格式
func toGitPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
```

### Step 2: 提交

```bash
git add pkg/git/gitignore.go
git commit -m "feat: 添加 GitIgnore 操作模块"
```

---

## Task 3: 后端 - 在 app.go 添加导出 API

**Files:**
- Modify: `app.go`

### Step 1: 在 app.go 添加 GetUntrackedFiles 方法

打开 `app.go`，找到导出的 API 方法区域（约第 100-400 行），在适当位置添加：

```go
// GetUntrackedFiles 获取未跟踪文件列表
func (a *App) GetUntrackedFiles(projectPath string) ([]git.UntrackedFile, error) {
	return git.GetUntrackedFiles(projectPath)
}
```

### Step 2: 在 app.go 添加 StageFiles 方法

```go
// StageFiles 添加文件到暂存区
func (a *App) StageFiles(projectPath string, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("文件列表为空")
	}

	// 使用 Command() 构建命令
	args := append([]string{"add"}, files...)
	cmd := git.Command("git", args...)
	cmd.Dir = projectPath

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("添加到暂存区失败: %s\n%w", string(output), err)
	}

	return nil
}
```

### Step 3: 在 app.go 添加 AddToGitIgnore 方法

```go
// AddToGitIgnore 添加到 .gitignore
func (a *App) AddToGitIgnore(projectPath, pattern, mode string) error {
	gitMode := git.ExcludeMode(mode)

	// 如果是目录模式，pattern 已经是最终规则
	// 否则需要根据文件路径生成规则
	var finalPattern string
	var err error

	if gitMode == git.ExcludeModeDirectory {
		finalPattern = pattern
	} else {
		finalPattern, err = git.GenerateGitIgnorePattern(pattern, gitMode)
		if err != nil {
			return fmt.Errorf("生成规则失败: %w", err)
		}
	}

	return git.AddToGitIgnoreFile(projectPath, finalPattern)
}
```

### Step 4: 在 app.go 添加 GetDirectoryOptions 方法

```go
// GetDirectoryOptions 获取目录层级选项
func (a *App) GetDirectoryOptions(filePath string) ([]git.DirectoryOption, error) {
	return git.GetDirectoryOptions(filePath), nil
}
```

### Step 5: 提交

```bash
git add app.go
git commit -m "feat: 添加未跟踪文件管理 API"
```

---

## Task 4: 前端 - 扩展 commitStore 状态管理

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`

### Step 1: 在 commitStore.ts 添加 UntrackedFile 状态导入

打开 `frontend/src/stores/commitStore.ts`，在 imports 部分添加：

```typescript
import type { UntrackedFile } from '../types'
```

### Step 2: 添加状态定义

在 `CommitState` 接口中添加：

```typescript
untrackedFiles: UntrackedFile[]
untrackedFilesLoading: boolean
```

在状态初始化部分（`return {}` 对象中）添加：

```typescript
untrackedFiles: [],
untrackedFilesLoading: false,
```

### Step 3: 添加 loadUntrackedFiles 方法

```typescript
async loadUntrackedFiles(projectPath: string) {
  this.untrackedFilesLoading = true
  try {
    const files = await GetUntrackedFiles(projectPath)
    this.untrackedFiles = files
  } catch (e) {
    console.error('加载未跟踪文件失败:', e)
    this.untrackedFiles = []
  } finally {
    this.untrackedFilesLoading = false
  }
}
```

### Step 4: 添加 stageFiles 方法

```typescript
async stageFiles(files: string[]) {
  if (!this.selectedProjectPath) return

  try {
    await StageFiles(this.selectedProjectPath, files)
    // 刷新暂存区和未跟踪文件
    await Promise.all([
      this.loadStagingStatus(this.selectedProjectPath),
      this.loadUntrackedFiles(this.selectedProjectPath)
    ])
  } catch (e) {
    const msg = e instanceof Error ? e.message : '操作失败'
    console.error('添加到暂存区失败:', e)
    throw e
  }
}
```

### Step 5: 添加 addToGitIgnore 方法

```typescript
async addToGitIgnore(file: string, mode: 'exact' | 'extension' | 'directory') {
  if (!this.selectedProjectPath) return

  try {
    await AddToGitIgnore(this.selectedProjectPath, file, mode)
    // 刷新未跟踪文件列表
    await this.loadUntrackedFiles(this.selectedProjectPath)
  } catch (e) {
    console.error('添加到排除列表失败:', e)
    throw e
  }
}
```

### Step 6: 提交

```bash
git add frontend/src/stores/commitStore.ts
git commit -m "feat: 扩展 commitStore 支持未跟踪文件管理"
```

---

## Task 5: 前端 - 创建 UntrackedFiles 组件

**Files:**
- Create: `frontend/src/components/UntrackedFiles.vue`

### Step 1: 创建 UntrackedFiles.vue 组件

创建 `frontend/src/components/UntrackedFiles.vue`：

```vue
<template>
  <div class="untracked-files-section">
    <div class="section-header">
      <div class="header-left">
        <span class="icon">📄</span>
        <h3>未跟踪文件 ({{ files.length }})</h3>
      </div>
      <button @click="toggleCollapse" class="icon-btn">
        {{ collapsed ? '▼' : '▲' }}
      </button>
    </div>

    <div v-if="!collapsed" class="files-list">
      <div v-if="files.length === 0" class="empty-state">
        <span>无未跟踪文件</span>
      </div>
      <div
        v-for="file in files"
        :key="file.path"
        class="file-item"
        @contextmenu.prevent="$emit('context-menu', $event, file)"
      >
        <span class="file-icon">📝</span>
        <span class="file-name">{{ file.path }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { UntrackedFile } from '../types'

defineProps<{
  files: UntrackedFile[]
}>()

const emit = defineEmits<{
  (e: 'context-menu', event: MouseEvent, file: UntrackedFile): void
}>()

const collapsed = ref(false)

function toggleCollapse() {
  collapsed.value = !collapsed.value
}
</script>

<style scoped>
.untracked-files-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm) 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.header-left h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.icon-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  font-size: 12px;
}

.icon-btn:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  max-height: 300px;
  overflow-y: auto;
}

.empty-state {
  padding: var(--space-md);
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: context-menu;
  transition: all var(--transition-fast);
}

.file-item:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

.file-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.file-name {
  font-size: 13px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  word-break: break-all;
}
</style>
```

### Step 2: 提交

```bash
git add frontend/src/components/UntrackedFiles.vue
git commit -m "feat: 创建未跟踪文件列表组件"
```

---

## Task 6: 前端 - 创建 ContextMenu 组件

**Files:**
- Create: `frontend/src/components/ContextMenu.vue`

### Step 1: 创建 ContextMenu.vue 组件

创建 `frontend/src/components/ContextMenu.vue`：

```vue
<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="context-menu"
      :style="{ left: x + 'px', top: y + 'px' }"
      @click="close"
    >
      <div class="menu-item" @click="emit('copy-path')">
        <span class="icon">📋</span>
        复制文件路径
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="emit('stage-file')">
        <span class="icon">✓</span>
        添加到暂存区
      </div>
      <div class="menu-item" @click="emit('exclude-file')">
        <span class="icon">🚫</span>
        添加到排除列表...
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="emit('open-explorer')">
        <span class="icon">📁</span>
        在文件管理器中打开
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

defineProps<{
  visible: boolean
  x: number
  y: number
}>()

const emit = defineEmits<{
  (e: 'copy-path'): void
  (e: 'stage-file'): void
  (e: 'exclude-file'): void
  (e: 'open-explorer'): void
  (e: 'close'): void
}>()

function close() {
  emit('close')
}

function handleClickOutside() {
  close()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.context-menu {
  position: fixed;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  z-index: var(--z-modal);
  min-width: 200px;
  padding: var(--space-xs) 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: 13px;
  color: var(--text-secondary);
}

.menu-item:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.menu-item .icon {
  font-size: 14px;
  flex-shrink: 0;
}

.menu-divider {
  height: 1px;
  background: var(--border-default);
  margin: var(--space-xs) 0;
}
</style>
```

### Step 2: 提交

```bash
git add frontend/src/components/ContextMenu.vue
git commit -m "feat: 创建右键菜单组件"
```

---

## Task 7: 前端 - 创建 ExcludeDialog 组件

**Files:**
- Create: `frontend/src/components/ExcludeDialog.vue`
- Modify: `frontend/src/types/index.ts`

### Step 1: 在 types/index.ts 添加类型

在 `frontend/src/types/index.ts` 末尾添加：

```typescript
// 排除模式
export type ExcludeMode = 'exact' | 'extension' | 'directory'

// 目录选项
export interface DirectoryOption {
  pattern: string
  label: string
}
```

### Step 2: 创建 ExcludeDialog.vue 组件

创建 `frontend/src/components/ExcludeDialog.vue`：

```vue
<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="exclude-dialog">
      <div class="dialog-header">
        <h3>添加到排除列表</h3>
        <button @click="close" class="close-btn">×</button>
      </div>

      <div class="dialog-body">
        <label class="input-label">忽略文件名或模式:</label>
        <input v-model="pattern" class="pattern-input" />

        <div class="radio-group">
          <label class="radio-option">
            <input type="radio" value="exact" v-model="mode" />
            <span>忽略精确的文件名</span>
          </label>

          <label class="radio-option">
            <input type="radio" value="extension" v-model="mode" />
            <span>忽略所有文件的扩展名 ({{ extension }})</span>
          </label>

          <label class="radio-option">
            <input type="radio" value="directory" v-model="mode" :disabled="!hasDirectory" />
            <span>忽略下列所有:</span>
          </label>

          <select
            v-if="mode === 'directory'"
            v-model="selectedDirectory"
            class="directory-select"
            :disabled="!hasDirectory"
          >
            <option v-for="opt in directoryOptions" :key="opt.pattern" :value="opt.pattern">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="close" class="btn-secondary">取消</button>
        <button @click="confirm" class="btn-primary">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { DirectoryOption, ExcludeMode } from '../types'
import { GetDirectoryOptions } from '../../wailsjs/go/main/App'

const props = defineProps<{
  visible: boolean
  filePath: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm', mode: ExcludeMode, pattern: string): void
}>()

const pattern = ref(props.filePath)
const mode = ref<ExcludeMode>('exact')
const selectedDirectory = ref('')
const directoryOptions = ref<DirectoryOption[]>([])

const extension = computed(() => {
  const ext = props.filePath.split('.').pop()
  return ext ? `.${ext}` : ''
})

const hasDirectory = computed(() => {
  return props.filePath.includes('/') || props.filePath.includes('\\')
})

watch(() => props.filePath, async (newPath) => {
  pattern.value = newPath

  // 自动选择默认模式
  if (!hasDirectory.value) {
    mode.value = 'exact'
  } else {
    mode.value = 'directory'
  }

  // 加载目录选项
  if (hasDirectory.value) {
    try {
      const opts = await GetDirectoryOptions(newPath)
      directoryOptions.value = opts
      if (opts.length > 0) {
        selectedDirectory.value = opts[0].pattern
      }
    } catch (e) {
      console.error('加载目录选项失败:', e)
    }
  } else {
    directoryOptions.value = []
  }
}, { immediate: true })

function close() {
  emit('close')
}

function confirm() {
  let finalPattern = pattern.value
  if (mode.value === 'directory' && selectedDirectory.value) {
    finalPattern = selectedDirectory.value
  }
  emit('confirm', mode.value, finalPattern)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}

.exclude-dialog {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg);
  border-bottom: 1px solid var(--border-default);
}

.dialog-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 24px;
  line-height: 1;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.close-btn:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.dialog-body {
  padding: var(--space-lg);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.input-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.pattern-input {
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.pattern-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.radio-option {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

.radio-option input[type="radio"] {
  cursor: pointer;
}

.radio-option input[type="radio"]:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.directory-select {
  width: 100%;
  margin-top: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
  padding: var(--space-lg);
  border-top: 1px solid var(--border-default);
}

.btn-secondary,
.btn-primary {
  padding: var(--space-sm) var(--space-lg);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-secondary {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

.btn-primary {
  background: var(--accent-success);
  color: white;
  border-color: var(--accent-success);
}

.btn-primary:hover {
  background: #059669;
}
</style>
```

### Step 3: 提交

```bash
git add frontend/src/components/ExcludeDialog.vue frontend/src/types/index.ts
git commit -m "feat: 创建排除对话框组件"
```

---

## Task 8: 前端 - 在 StagingArea 集成 UntrackedFiles 组件

**Files:**
- Modify: `frontend/src/components/StagingArea.vue`

### Step 1: 在 StagingArea.vue 引入 UntrackedFiles

打开 `frontend/src/components/StagingArea.vue`，在 `<script setup>` 部分添加：

```typescript
import UntrackedFiles from './UntrackedFiles.vue'
```

### Step 2: 在 template 添加 UntrackedFiles 组件

在 StagingArea 的 template 末尾（在 `</div>` 闭合标签前）添加：

```vue
<!-- 未跟踪文件区域 -->
<UntrackedFiles
  v-if="commitStore.untrackedFiles.length > 0"
  :files="commitStore.untrackedFiles"
  @context-menu="handleContextMenu"
/>

<!-- 右键菜单 -->
<ContextMenu
  :visible="contextMenuVisible"
  :x="contextMenuX"
  :y="contextMenuY"
  @copy-path="handleCopyPath"
  @stage-file="handleStageFile"
  @exclude-file="handleExcludeFile"
  @open-explorer="handleOpenExplorer"
  @close="closeContextMenu"
/>

<!-- 排除对话框 -->
<ExcludeDialog
  :visible="excludeDialogVisible"
  :file-path="selectedFile?.path || ''"
  @close="excludeDialogVisible = false"
  @confirm="handleExcludeConfirm"
/>
```

### Step 3: 添加响应式变量和方法

在 `<script setup>` 部分添加：

```typescript
import { ref } from 'vue'
import ContextMenu from './ContextMenu.vue'
import ExcludeDialog from './ExcludeDialog.vue'
import { useCommitStore } from '../stores/commitStore'
import type { UntrackedFile } from '../types'

const commitStore = useCommitStore()

// 右键菜单状态
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const selectedFile = ref<UntrackedFile | null>(null)

// 排除对话框状态
const excludeDialogVisible = ref(false)

function handleContextMenu(event: MouseEvent, file: UntrackedFile) {
  selectedFile.value = file
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function closeContextMenu() {
  contextMenuVisible.value = false
}

async function handleCopyPath() {
  if (!selectedFile.value) return
  try {
    await navigator.clipboard.writeText(selectedFile.value.path)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('复制失败:', e)
  }
  closeContextMenu()
}

async function handleStageFile() {
  if (!selectedFile.value) return
  try {
    await commitStore.stageFiles([selectedFile.value.path])
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('添加到暂存区失败:', e)
  }
  closeContextMenu()
}

function handleExcludeFile() {
  closeContextMenu()
  excludeDialogVisible.value = true
}

async function handleExcludeConfirm(mode: 'exact' | 'extension' | 'directory', pattern: string) {
  if (!selectedFile.value) return
  try {
    await commitStore.addToGitIgnore(selectedFile.value.path, mode)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('添加到排除列表失败:', e)
  }
  excludeDialogVisible.value = false
}

async function handleOpenExplorer() {
  if (!selectedFile.value || !commitStore.selectedProjectPath) return
  try {
    const fullPath = `${commitStore.selectedProjectPath}/${selectedFile.value.path}`
    await OpenInFileExplorer(fullPath)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('打开失败:', e)
  }
  closeContextMenu()
}
```

### Step 4: 在 CommitPanel 加载未跟踪文件

修改 `frontend/src/components/CommitPanel.vue`，在 `watch(() => projectStore.selectedProject, ...)` 回调中添加：

```typescript
await commitStore.loadUntrackedFiles(project.path)
```

### Step 5: 提交

```bash
git add frontend/src/components/StagingArea.vue frontend/src/components/CommitPanel.vue
git commit -m "feat: 集成未跟踪文件管理功能"
```

---

## Task 9: 测试和验证

### Step 1: 启动开发服务器

```bash
wails dev
```

### Step 2: 测试未跟踪文件显示

1. 创建一个新文件（如 `test.txt`）
2. 在应用中选择项目
3. 验证未跟踪文件区域显示 `test.txt`

### Step 3: 测试添加到暂存区

1. 右键点击未跟踪文件
2. 选择"添加到暂存区"
3. 验证文件从未跟踪区域消失，出现在暂存区

### Step 4: 测试排除功能 - 精确文件名

1. 右键点击 `docs/test.md`
2. 选择"添加到排除列表"
3. 选择"忽略精确的文件名"
4. 验证 `.gitignore` 包含 `docs/test.md`

### Step 5: 测试排除功能 - 扩展名

1. 右键点击 `test.log`
2. 选择"添加到排除列表"
3. 选择"忽略所有文件的扩展名"
4. 验证 `.gitignore` 包含 `*.log`

### Step 6: 测试排除功能 - 目录层级

1. 右键点击 `docs/plans/test.md`
2. 选择"添加到排除列表"
3. 选择"忽略下列所有"
4. 验证下拉菜单显示：
   - `docs`
   - `docs/plans`
   - `docs/plans/*.md`
5. 选择不同选项，验证 `.gitignore` 规则正确

### Step 7: 测试路径复制

1. 右键点击文件
2. 选择"复制文件路径"
3. 粘贴验证路径正确

### Step 8: 测试在文件管理器中打开

1. 右键点击文件
2. 选择"在文件管理器中打开"
3. 验证文件管理器正确打开

### Step 9: 测试边界情况

- 测试无未跟踪文件时的空状态
- 测试根目录文件（如 `config.json`）的目录选项应被禁用
- 测试中文路径
- 测试 Windows 路径分隔符转换

### Step 10: 提交测试修复（如有）

```bash
git add .
git commit -m "fix: 修复测试发现的问题"
```

---

## Task 10: 生成 Wails 绑定

### Step 1: 重新生成绑定

```bash
wails generate module
```

### Step 2: 检查生成的绑定

验证 `frontend/wailsjs/go/main/App.js` 包含新增的方法：
- `GetUntrackedFiles`
- `StageFiles`
- `AddToGitIgnore`
- `GetDirectoryOptions`

### Step 3: 提交

```bash
git add frontend/wailsjs
git commit -m "chore: 更新 Wails 绑定"
```

---

## 实现完成检查清单

- [ ] 后端 API 方法全部实现
- [ ] 前端组件全部创建
- [ ] 状态管理扩展完成
- [ ] 组件集成完成
- [ ] 功能测试通过
- [ ] 边界情况处理正确
- [ ] 无控制台弹窗（Windows）
- [ ] 路径格式转换正确
- [ ] .gitignore 规则生成正确

---

## 预期结果

完成后，用户可以：

1. 在暂存区下方看到所有未跟踪文件
2. 右键点击文件显示操作菜单
3. 快速添加文件到暂存区
4. 通过排除对话框灵活配置 `.gitignore` 规则
5. 复制文件路径或在文件管理器中打开
6. 所有操作无控制台弹窗，体验流畅
