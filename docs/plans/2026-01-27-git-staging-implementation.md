# Git 暂存管理功能实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 为 AI Commit Hub 添加完整的 Git 暂存管理功能，支持文件级别的暂存/取消暂存操作，并提供直观的 diff 预览。

**架构:** 基于 Wails (Go + Vue3) 框架，后端通过 Git 命令操作暂存区，前端使用 Pinia 管理状态，v-code-diff 库展示 diff。

**技术栈:**
- 后端: Go 1.21+, git 命令行工具
- 前端: Vue 3 + TypeScript + Pinia + v-code-diff
- 绑定: Wails v2

---

## 前置条件

**已完成:**
- ✅ 后端 `pkg/git/staging.go` 和 `pkg/git/diff.go` 已实现
- ✅ `app.go` 中的 6 个导出方法已创建
- ✅ Wails 绑定已生成
- ✅ v-code-diff 库已安装测试
- ✅ 设计文档已完成 (`docs/plans/2026-01-27-git-staging-ui-design.md`)

---

## Task 1: 更新前端类型定义

**目的:** 添加 `ignored` 字段到 `StagedFile` 接口，确保与后端结构同步

**Files:**
- Modify: `frontend/src/types/index.ts:37-40`

**Step 1: 修改 StagedFile 接口**

```typescript
// 在 frontend/src/types/index.ts 中
export interface StagedFile {
  path: string
  status: string // 'Modified' | 'New' | 'Deleted' | 'Renamed'
  ignored: boolean // 是否被 .gitignore 忽略
}
```

**Step 2: 验证类型编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(types): 添加 StagedFile.ignored 字段"
```

---

## Task 2: 扩展 commitStore 状态管理

**目的:** 在 commitStore 中添加暂存区管理状态和方法

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 添加导入语句**

在文件顶部添加（第 12 行后）:
```typescript
import {
  GetStagingStatus,
  GetFileDiff,
  StageFile,
  StageAllFiles,
  UnstageFile,
  UnstageAllFiles
} from '../../wailsjs/go/main/App'
```

**Step 2: 添加新的状态变量**

在 `defineStore` 函数开始处（第 15 行后）添加:
```typescript
// 暂存区状态
const stagingStatus = ref<StagingStatus | null>(null)
const isLoadingStaging = ref(false)

// 文件选择状态
const selectedStagedFiles = ref<Set<string>>(new Set())
const selectedUnstagedFiles = ref<Set<string>>(new Set())

// Diff 预览
const selectedFile = ref<StagedFile | null>(null)
const fileDiff = ref<string | null>(null)
const isLoadingDiff = ref(false)
```

**Step 3: 添加状态管理方法**

在 `handleError` 函数后（第 185 行后）添加:
```typescript
// ========== 暂存区管理 ==========

async function loadStagingStatus(path: string) {
  isLoadingStaging.value = true
  error.value = null

  try {
    const result = await GetStagingStatus(path) as StagingStatus
    stagingStatus.value = result
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '加载暂存状态失败'
    error.value = message
  } finally {
    isLoadingStaging.value = false
  }
}

async function selectFile(file: StagedFile) {
  selectedFile.value = file
  await loadFileDiff(file.path, file.path !== '')
}

async function loadFileDiff(filePath: string, staged: boolean) {
  isLoadingDiff.value = true
  fileDiff.value = null

  try {
    const diff = await GetFileDiff(selectedProjectPath.value, filePath, staged)
    fileDiff.value = diff
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '加载 diff 失败'
    error.value = message
  } finally {
    isLoadingDiff.value = false
  }
}

async function stageFile(filePath: string) {
  if (!selectedProjectPath.value) return

  try {
    await StageFile(selectedProjectPath.value, filePath)
    await loadStagingStatus(selectedProjectPath.value)
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '暂存文件失败'
    error.value = message
    throw e
  }
}

async function unstageFile(filePath: string) {
  if (!selectedProjectPath.value) return

  try {
    await UnstageFile(selectedProjectPath.value, filePath)
    await loadStagingStatus(selectedProjectPath.value)
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '取消暂存失败'
    error.value = message
    throw e
  }
}

async function stageAllFiles() {
  if (!selectedProjectPath.value) return

  try {
    await StageAllFiles(selectedProjectPath.value)
    await loadStagingStatus(selectedProjectPath.value)
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '暂存所有文件失败'
    error.value = message
    throw e
  }
}

async function unstageAllFiles() {
  if (!selectedProjectPath.value) return

  try {
    await UnstageAllFiles(selectedProjectPath.value)
    await loadStagingStatus(selectedProjectPath.value)
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '取消所有暂存失败'
    error.value = message
    throw e
  }
}

async function stageSelectedFiles() {
  const files = Array.from(selectedUnstagedFiles.value)
  for (const filePath of files) {
    await stageFile(filePath)
  }
  selectedUnstagedFiles.value.clear()
}

async function unstageSelectedFiles() {
  const files = Array.from(selectedStagedFiles.value)
  for (const filePath of files) {
    await unstageFile(filePath)
  }
  selectedStagedFiles.value.clear()
}

function toggleFileSelection(filePath: string, type: 'staged' | 'unstaged') {
  const set = type === 'staged' ? selectedStagedFiles.value : selectedUnstagedFiles.value
  if (set.has(filePath)) {
    set.delete(filePath)
  } else {
    set.add(filePath)
  }
}

function clearStagingState() {
  stagingStatus.value = null
  selectedStagedFiles.value.clear()
  selectedUnstagedFiles.value.clear()
  selectedFile.value = null
  fileDiff.value = null
}
```

**Step 4: 更新 return 对象**

在 `return` 对象中（第 187 行）添加新的导出:
```typescript
return {
  // ... 现有导出 ...
  stagingStatus,
  isLoadingStaging,
  selectedStagedFiles,
  selectedUnstagedFiles,
  selectedFile,
  fileDiff,
  isLoadingDiff,
  loadStagingStatus,
  selectFile,
  stageFile,
  unstageFile,
  stageAllFiles,
  unstageAllFiles,
  stageSelectedFiles,
  unstageSelectedFiles,
  toggleFileSelection,
  clearStagingState
}
```

**Step 5: 验证类型检查**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 6: 提交**

```bash
git add frontend/src/stores/commitStore.ts
git commit -m "feat(store): 添加暂存区状态管理"
```

---

## Task 3: 创建 StagedList.vue 组件

**目的:** 显示已暂存文件列表，支持取消暂存操作

**Files:**
- Create: `frontend/src/components/StagedList.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <div class="file-list-container staged-list">
    <div class="list-header">
      <h4>已暂存 ({{ commitStore.stagingStatus?.staged?.length ?? 0 }})</h4>
      <div class="bulk-actions" v-if="commitStore.stagingStatus?.staged?.length > 0">
        <label class="select-all">
          <input
            type="checkbox"
            :checked="isAllSelected"
            @change="toggleSelectAll"
          />
          <span>全选</span>
        </label>
        <button
          @click="unstageSelected"
          :disabled="selectedCount === 0"
          class="btn-bulk"
          title="取消暂存选中的文件"
        >
          [-] 取消选定
        </button>
        <button
          @click="unstageAll"
          class="btn-bulk btn-bulk-danger"
          title="取消暂存所有文件"
        >
          [═] 取消所有
        </button>
      </div>
    </div>

    <div class="file-list" v-if="commitStore.stagingStatus?.staged?.length > 0">
      <div
        v-for="file in commitStore.stagingStatus.staged"
        :key="file.path"
        :class="['file-item', 'staged', { 'selected': isSelected(file.path) }]"
        @click="handleFileClick(file)"
      >
        <label class="file-checkbox">
          <input
            type="checkbox"
            :checked="isSelected(file.path)"
            @change="toggleSelection(file.path)"
            @click.stop
          />
        </label>

        <span class="file-status" :class="getStatusClass(file.status)">
          {{ getStatusIcon(file.status) }}
        </span>

        <span class="file-path" :title="file.path">{{ file.path }}</span>

        <button
          @click.stop="handleUnstage(file.path)"
          class="btn-mini btn-unstage"
          title="取消暂存"
        >
          -
        </button>
      </div>
    </div>

    <div v-else class="empty-state">
      <span class="empty-icon">📭</span>
      <span>暂存区为空</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import type { StagedFile } from '../types'

const commitStore = useCommitStore()

const isAllSelected = computed(() => {
  const staged = commitStore.stagingStatus?.staged ?? []
  return staged.length > 0 && staged.every(f => commitStore.selectedStagedFiles.has(f.path))
})

const selectedCount = computed(() => commitStore.selectedStagedFiles.size)

function isSelected(filePath: string): boolean {
  return commitStore.selectedStagedFiles.has(filePath)
}

function toggleSelection(filePath: string) {
  commitStore.toggleFileSelection(filePath, 'staged')
}

function toggleSelectAll() {
  const staged = commitStore.stagingStatus?.staged ?? []
  if (isAllSelected.value) {
    staged.forEach(f => commitStore.selectedStagedFiles.delete(f.path))
  } else {
    staged.forEach(f => commitStore.selectedStagedFiles.add(f.path))
  }
}

async function handleUnstage(filePath: string) {
  try {
    await commitStore.unstageFile(filePath)
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function unstageSelected() {
  try {
    await commitStore.unstageSelectedFiles()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function unstageAll() {
  try {
    await commitStore.unstageAllFiles()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

function handleFileClick(file: StagedFile) {
  commitStore.selectFile(file)
}

function getStatusIcon(status: string): string {
  const icons: Record<string, string> = {
    'Modified': '📝',
    'New': '✨',
    'Deleted': '🗑️',
    'Renamed': '📛'
  }
  return icons[status] || '📄'
}

function getStatusClass(status: string): string {
  return status.toLowerCase()
}
</script>

<style scoped>
.file-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm) var(--space-md);
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-secondary);
}

.list-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.bulk-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.select-all {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.select-all input[type="checkbox"] {
  cursor: pointer;
}

.btn-bulk {
  padding: 4px 10px;
  font-size: 11px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-bulk:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--border-hover);
}

.btn-bulk:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-bulk-danger:hover:not(:disabled) {
  background: #fee2e2;
  border-color: #f87171;
  color: #dc2626;
}

.file-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-xs);
}

.file-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-xs);
  cursor: pointer;
  transition: all 0.2s;
}

.file-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
}

.file-item.selected {
  background: var(--bg-selected);
  border-color: var(--color-primary);
}

.file-item.staged {
  border-left: 3px solid var(--color-success);
}

.file-checkbox {
  display: flex;
  align-items: center;
}

.file-checkbox input[type="checkbox"] {
  cursor: pointer;
}

.file-status {
  font-size: 14px;
  flex-shrink: 0;
}

.file-path {
  flex: 1;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-mini {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;
}

.btn-unstage {
  background: var(--color-danger);
  color: white;
}

.btn-unstage:hover {
  background: #dc2626;
  transform: scale(1.1);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  color: var(--text-tertiary);
  gap: var(--space-sm);
}

.empty-icon {
  font-size: 32px;
  opacity: 0.5;
}
</style>
```

**Step 2: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/components/StagedList.vue
git commit -m "feat(component): 创建 StagedList 已暂存文件列表组件"
```

---

## Task 4: 创建 UnstagedList.vue 组件

**目的:** 显示未暂存文件列表，支持暂存操作，包含忽略文件的特殊处理

**Files:**
- Create: `frontend/src/components/UnstagedList.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <div class="file-list-container unstaged-list">
    <div class="list-header">
      <h4>未暂存 ({{ unstagedCount }})</h4>
      <div class="bulk-actions" v-if="commitStore.stagingStatus?.unstaged?.length > 0">
        <label class="select-all">
          <input
            type="checkbox"
            :checked="isAllSelected"
            @change="toggleSelectAll"
          />
          <span>全选</span>
        </label>
        <button
          @click="stageSelected"
          :disabled="selectedCount === 0"
          class="btn-bulk"
          title="暂存选中的文件"
        >
          [+] 暂存所选
        </button>
        <button
          @click="stageAll"
          class="btn-bulk btn-bulk-primary"
          title="暂存所有未忽略文件"
        >
          [║] 暂存所有
        </button>
      </div>
    </div>

    <div class="file-list" v-if="commitStore.stagingStatus?.unstaged?.length > 0">
      <div
        v-for="file in commitStore.stagingStatus.unstaged"
        :key="file.path"
        :class="['file-item', 'unstaged', { 'selected': isSelected(file.path), 'ignored': file.ignored }]"
        @click="handleFileClick(file)"
      >
        <label class="file-checkbox">
          <input
            type="checkbox"
            :checked="isSelected(file.path)"
            @change="toggleSelection(file.path)"
            @click.stop
          />
        </label>

        <span class="file-status" :class="getStatusClass(file.status)">
          {{ getStatusIcon(file.status) }}
        </span>

        <span class="ignored-badge" v-if="file.ignored">已忽略</span>

        <span class="file-path" :title="file.path">{{ file.path }}</span>

        <button
          @click.stop="handleStage(file)"
          class="btn-mini btn-stage"
          :disabled="file.ignored"
          :title="file.ignored ? '此文件被 .gitignore 忽略' : '暂存文件'"
        >
          +
        </button>
      </div>
    </div>

    <div v-else class="empty-state">
      <span class="empty-icon">✨</span>
      <span>工作区干净</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import type { StagedFile } from '../types'

const commitStore = useCommitStore()

const unstagedCount = computed(() => {
  return commitStore.stagingStatus?.unstaged?.length ?? 0
})

const isAllSelected = computed(() => {
  const unstaged = commitStore.stagingStatus?.unstaged ?? []
  return unstaged.length > 0 && unstaged.every(f => commitStore.selectedUnstagedFiles.has(f.path))
})

const selectedCount = computed(() => commitStore.selectedUnstagedFiles.size)

function isSelected(filePath: string): boolean {
  return commitStore.selectedUnstagedFiles.has(filePath)
}

function toggleSelection(filePath: string) {
  commitStore.toggleFileSelection(filePath, 'unstaged')
}

function toggleSelectAll() {
  const unstaged = commitStore.stagingStatus?.unstaged ?? []
  if (isAllSelected.value) {
    unstaged.forEach(f => commitStore.selectedUnstagedFiles.delete(f.path))
  } else {
    unstaged.forEach(f => commitStore.selectedUnstagedFiles.add(f.path))
  }
}

async function handleStage(file: StagedFile) {
  if (file.ignored) return

  try {
    await commitStore.stageFile(file.path)
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageSelected() {
  try {
    await commitStore.stageSelectedFiles()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageAll() {
  try {
    await commitStore.stageAllFiles()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

function handleFileClick(file: StagedFile) {
  commitStore.selectFile(file)
}

function getStatusIcon(status: string): string {
  const icons: Record<string, string> = {
    'Modified': '📝',
    'New': '✨',
    'Deleted': '🗑️',
    'Renamed': '📛'
  }
  return icons[status] || '📄'
}

function getStatusClass(status: string): string {
  return status.toLowerCase()
}
</script>

<style scoped>
.file-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm) var(--space-md);
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-secondary);
}

.list-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.bulk-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.select-all {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.select-all input[type="checkbox"] {
  cursor: pointer;
}

.btn-bulk {
  padding: 4px 10px;
  font-size: 11px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-bulk:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--border-hover);
}

.btn-bulk:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-bulk-primary:hover:not(:disabled) {
  background: #dcfce7;
  border-color: var(--color-success);
  color: #16a34a;
}

.file-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-xs);
}

.file-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-xs);
  cursor: pointer;
  transition: all 0.2s;
}

.file-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
}

.file-item.selected {
  background: var(--bg-selected);
  border-color: var(--color-primary);
}

.file-item.unstaged {
  border-left: 3px solid var(--color-warning);
}

.file-item.ignored {
  opacity: 0.6;
  background: #2a2a2a;
  border-color: #666;
}

.file-item.ignored .file-path {
  color: #888;
  text-decoration: line-through;
}

.file-item.ignored .btn-stage {
  opacity: 0.5;
  cursor: not-allowed;
}

.file-checkbox {
  display: flex;
  align-items: center;
}

.file-checkbox input[type="checkbox"] {
  cursor: pointer;
}

.file-status {
  font-size: 14px;
  flex-shrink: 0;
}

.ignored-badge {
  padding: 2px 6px;
  font-size: 9px;
  border-radius: var(--radius-sm);
  background: #666;
  color: #aaa;
  white-space: nowrap;
}

.file-path {
  flex: 1;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-mini {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;
}

.btn-stage {
  background: var(--color-success);
  color: white;
}

.btn-stage:hover:not(:disabled) {
  background: #16a34a;
  transform: scale(1.1);
}

.btn-stage:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  color: var(--text-tertiary);
  gap: var(--space-sm);
}

.empty-icon {
  font-size: 32px;
  opacity: 0.5;
}
</style>
```

**Step 2: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/components/UnstagedList.vue
git commit -m "feat(component): 创建 UnstagedList 未暂存文件列表组件"
```

---

## Task 5: 创建 DiffViewer.vue 组件

**目的:** 使用 v-code-diff 库显示文件 diff 内容

**Files:**
- Create: `frontend/src/components/DiffViewer.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <div class="diff-viewer">
    <div class="diff-header" v-if="commitStore.selectedFile">
      <div class="file-info">
        <span class="file-icon">📄</span>
        <span class="file-name">{{ commitStore.selectedFile.path }}</span>
        <span class="file-status" :class="getStatusClass(commitStore.selectedFile.status)">
          {{ commitStore.selectedFile.status }}
        </span>
      </div>
      <button @click="closeDiff" class="btn-close" title="关闭">×</button>
    </div>

    <div class="diff-content" v-if="commitStore.selectedFile">
      <div v-if="commitStore.isLoadingDiff" class="diff-loading">
        <span class="loading-spinner"></span>
        <span>加载中...</span>
      </div>

      <div v-else-if="commitStore.fileDiff" class="diff-renderer">
        <CodeDiff
          :old-string="getOldCode()"
          :new-string="getNewCode()"
          :output-format="'line-by-line'"
          :context="10"
          language="plaintext"
        />
      </div>

      <div v-else class="diff-empty">
        <span class="empty-icon">📭</span>
        <span>无 diff 内容</span>
      </div>
    </div>

    <div v-else class="diff-placeholder">
      <span class="placeholder-icon">👈</span>
      <span>点击文件查看 diff</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCommitStore } from '../stores/commitStore'
import { CodeDiff } from 'v-code-diff'

const commitStore = useCommitStore()

function closeDiff() {
  commitStore.selectFile({ path: '', status: '', ignored: false } as any)
}

function getStatusClass(status: string): string {
  return status.toLowerCase()
}

function getOldCode(): string {
  if (!commitStore.fileDiff) return ''

  // 简单解析 diff，提取旧代码
  const lines = commitStore.fileDiff.split('\n')
  const oldLines: string[] = []

  for (const line of lines) {
    if (line.startsWith('-') && !line.startsWith('---')) {
      oldLines.push(line.substring(1))
    } else if (line.startsWith(' ')) {
      oldLines.push(line.substring(1))
    }
  }

  return oldLines.join('\n')
}

function getNewCode(): string {
  if (!commitStore.fileDiff) return ''

  // 简单解析 diff，提取新代码
  const lines = commitStore.fileDiff.split('\n')
  const newLines: string[] = []

  for (const line of lines) {
    if (line.startsWith('+') && !line.startsWith('+++')) {
      newLines.push(line.substring(1))
    } else if (line.startsWith(' ')) {
      newLines.push(line.substring(1))
    }
  }

  return newLines.join('\n')
}
</script>

<style scoped>
.diff-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
}

.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm) var(--space-md);
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-tertiary);
}

.file-info {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex: 1;
  overflow: hidden;
}

.file-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.file-name {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-status {
  padding: 2px 8px;
  font-size: 10px;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  text-transform: uppercase;
  flex-shrink: 0;
}

.file-status.modified {
  background: #fef3c7;
  color: #d97706;
}

.file-status.new {
  background: #dcfce7;
  color: #16a34a;
}

.file-status.deleted {
  background: #fee2e2;
  color: #dc2626;
}

.btn-close {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-sm);
  border: none;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}

.btn-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.diff-content {
  flex: 1;
  overflow: auto;
}

.diff-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  gap: var(--space-md);
  color: var(--text-secondary);
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-default);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.diff-renderer {
  padding: var(--space-md);
}

.diff-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  gap: var(--space-md);
  color: var(--text-tertiary);
}

.empty-icon {
  font-size: 32px;
  opacity: 0.5;
}

.diff-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  gap: var(--space-md);
  color: var(--text-tertiary);
  height: 100%;
}

.placeholder-icon {
  font-size: 48px;
  opacity: 0.3;
}
</style>
```

**Step 2: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/components/DiffViewer.vue
git commit -m "feat(component): 创建 DiffViewer diff 预览组件"
```

---

## Task 6: 创建 StagingArea.vue 容器组件

**目的:** 组合 StagedList、UnstagedList 和 DiffViewer 组件

**Files:**
- Create: `frontend/src/components/StagingArea.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <div class="staging-area">
    <div class="staging-panels">
      <div class="file-lists-panel">
        <StagedList />
        <div class="list-divider"></div>
        <UnstagedList />
      </div>

      <div class="diff-panel">
        <DiffViewer />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import StagedList from './StagedList.vue'
import UnstagedList from './UnstagedList.vue'
import DiffViewer from './DiffViewer.vue'
</script>

<style scoped>
.staging-area {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.staging-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.file-lists-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  overflow: hidden;
  min-height: 0;
}

.list-divider {
  height: 1px;
  background: var(--border-default);
  flex-shrink: 0;
}

.diff-panel {
  overflow: hidden;
  min-height: 0;
}

@media (max-width: 1024px) {
  .staging-panels {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }
}
</style>
```

**Step 2: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/components/StagingArea.vue
git commit -m "feat(component): 创建 StagingArea 容器组件"
```

---

## Task 7: 重构 CommitPanel.vue 集成新组件

**目的:** 在 CommitPanel 中集成 StagingArea，替换原有的简单文件列表

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 添加导入语句**

在 `<script setup>` 部分顶部添加（第 1 行后）:
```vue
<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import { useProjectStore } from '../stores/projectStore'
import { usePushoverStore } from '../stores/pushoverStore'
import PushoverStatusRow from './PushoverStatusRow.vue'
import StagingArea from './StagingArea.vue'

// ... 其余代码保持不变
```

**Step 2: 替换当前状态显示区域**

找到 "Project Info Section" 中的文件列表部分（约第 53-67 行），替换为:
```vue
      <!-- Pushover Status Row -->
      <PushoverStatusRow v-if="currentProject" :project-path="currentProject.path" :status="pushoverStatus"
        :loading="pushoverStore.loading" @install="handleInstallPushover" @update="handleUpdatePushover" />

      <!-- Staging Area -->
      <StagingArea v-if="commitStore.projectStatus" />
```

**Step 3: 在 onMounted 中加载暂存状态**

找到 `onMounted` 钩子（约第 350 行），在加载项目状态后添加:
```typescript
onMounted(async () => {
  // ... 现有代码 ...

  // 监听 commit-delta 事件
  EventsOn('commit-delta', commitStore.handleDelta)

  // 监听 commit-complete 事件
  EventsOn('commit-complete', commitStore.handleComplete)

  // 监听 commit-error 事件
  EventsOn('commit-error', commitStore.handleError)

  // 新增：加载暂存区状态
  if (commitStore.selectedProjectPath) {
    await commitStore.loadStagingStatus(commitStore.selectedProjectPath)
  }
})
```

**Step 4: 在 handleRefresh 中刷新暂存状态**

找到 `handleRefresh` 函数（约第 250 行），在刷新项目状态后添加:
```typescript
async function handleRefresh() {
  if (!currentProject.value) return

  try {
    await commitStore.loadProjectStatus(currentProject.value.path)
    // 新增：刷新暂存区状态
    await commitStore.loadStagingStatus(currentProject.value.path)
  } catch (e) {
    console.error('刷新失败:', e)
  }
}
```

**Step 5: 在提交成功后刷新暂存状态**

找到 `handleCommit` 函数（约第 270 行），在提交成功后添加:
```typescript
async function handleCommit() {
  // ... 现有提交逻辑 ...

  try {
    await DoCommit(commitStore.selectedProjectPath, commitStore.generatedMessage)

    // 新增：刷新暂存区状态
    await commitStore.loadStagingStatus(commitStore.selectedProjectPath)

    // ... 其余处理 ...
  } catch (e) {
    // ... 错误处理 ...
  }
}
```

**Step 6: 在 onUnmounted 中清理状态**

找到 `onUnmounted` 钩子（约第 365 行），添加清理逻辑:
```typescript
onUnmounted(() => {
  EventsOff('commit-delta', commitStore.handleDelta)
  EventsOff('commit-complete', commitStore.handleComplete)
  EventsOff('commit-error', commitStore.handleError)

  // 新增：清理暂存区状态
  commitStore.clearStagingState()
})
```

**Step 7: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 8: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(panel): 集成 StagingArea 到 CommitPanel"
```

---

## Task 8: 添加 StagingStatus 类型定义

**目的:** 确保前端类型与后端同步

**Files:**
- Modify: `frontend/src/types/index.ts`

**Step 1: 添加 StagingStatus 接口**

在 `StagedFile` 接口后（第 41 行后）添加:
```typescript
export interface StagingStatus {
  staged: StagedFile[]
  unstaged: StagedFile[]
}
```

**Step 2: 验证类型编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(types): 添加 StagingStatus 接口"
```

---

## Task 9: 端到端测试

**目的:** 验证完整功能流程

**Step 1: 启动开发服务器**

Run: `wails dev`
Expected: 服务器启动成功，无编译错误

**Step 2: 测试暂存功能**

1. 选择一个 Git 项目
2. 验证已暂存和未暂存文件正确显示
3. 点击未暂存文件的 `+` 按钮
4. 验证文件移动到已暂存列表

**Step 3: 测试取消暂存功能**

1. 点击已暂存文件的 `-` 按钮
2. 验证文件移动到未暂存列表

**Step 4: 测试批量操作**

1. 勾选多个未暂存文件
2. 点击"暂存所选"按钮
3. 验证所有选中文件都被暂存

**Step 5: 测试忽略文件处理**

1. 确保有 .gitignore 文件
2. 验证被忽略的文件显示灰色样式
3. 验证忽略文件的暂存按钮被禁用

**Step 6: 测试 diff 预览**

1. 点击任意文件
2. 验证右侧显示 diff 内容
3. 验证已暂存和未暂存文件的 diff 都正确显示

**Step 7: 测试提交流程**

1. 暂存一些文件
2. 生成 commit 消息
3. 提交到本地
4. 验证暂存区状态更新

**Step 8: 修复发现的问题**

记录并修复测试中发现的所有问题

**Step 9: 最终提交**

```bash
git add -A
git commit -m "test: 完成端到端测试和问题修复"
```

---

## 验收标准

- [ ] 所有组件编译无错误
- [ ] 暂存/取消暂存单个文件功能正常
- [ ] 批量暂存/取消暂存功能正常
- [ ] 忽略文件正确显示且按钮禁用
- [ ] Diff 预览正确显示
- [ ] 提交后状态正确刷新
- [ ] UI 样式符合设计规范
- [ ] 无控制台错误或警告

---

## 参考资料

- 设计文档: `docs/plans/2026-01-27-git-staging-ui-design.md`
- v-code-diff 文档: https://github.com/Shimada666/v-code-diff
- Wails 文档: https://wails.io/docs/next/introduction
