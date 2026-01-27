<template>
  <div class="file-list-container unstaged-list">
    <div class="list-header">
      <h4>未暂存 ({{ unstagedCount }})</h4>
      <div class="bulk-actions" v-if="hasUnstagedFiles">
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

    <!-- 未暂存文件列表 -->
    <div class="file-list" v-if="hasUnstagedFiles">
      <div
        v-for="file in commitStore.stagingStatus?.unstaged || []"
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

        <span class="status-badge" :class="getStatusClass(file.status)">
          {{ getStatusText(file.status) }}
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

    <!-- 未跟踪文件列表 -->
    <div v-if="hasUntrackedFiles" class="untracked-section">
      <div class="list-header">
        <h4>未跟踪 ({{ commitStore.stagingStatus?.untracked?.length || 0 }})</h4>
        <div class="bulk-actions" v-if="commitStore.stagingStatus?.untracked && commitStore.stagingStatus.untracked.length > 0">
          <button
            @click="stageUntrackedSelected"
            :disabled="selectedUntrackedCount === 0"
            class="btn-bulk"
            title="暂存选中的未跟踪文件"
          >
            [+] 暂存所选
          </button>
          <button
            @click="stageAllUntracked"
            class="btn-bulk btn-bulk-primary"
            title="暂存所有未跟踪文件"
          >
            [║] 全部暂存
          </button>
        </div>
      </div>

      <div class="file-list">
        <div
          v-for="file in commitStore.stagingStatus?.untracked || []"
          :key="file.path"
          :class="['file-item', 'untracked', { 'selected': isUntrackedSelected(file.path) }]"
          @click="handleUntrackedFileClick(file)"
          @contextmenu.prevent="handleUntrackedContextMenu($event, file)"
        >
          <label class="file-checkbox">
            <input
              type="checkbox"
              :checked="isUntrackedSelected(file.path)"
              @change="toggleUntrackedSelection(file.path)"
              @click.stop
            />
          </label>

          <span class="file-status">📄</span>

          <span class="status-badge untracked">未跟踪</span>

          <span class="file-path" :title="file.path">{{ file.path }}</span>

          <button
            @click.stop="handleStageUntracked(file)"
            class="btn-mini btn-stage"
            title="暂存文件"
          >
            +
          </button>
        </div>
      </div>
    </div>

    <div v-if="!hasUnstagedFiles && !hasUntrackedFiles" class="empty-state">
      <span class="empty-icon">✨</span>
      <span>工作区干净</span>
    </div>

    <!-- 右键菜单 -->
    <ContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuX"
      :y="contextMenuY"
      @copy-path="handleCopyPath"
      @stage-file="handleStageUntrackedFile"
      @exclude-file="handleExcludeUntrackedFile"
      @open-explorer="handleOpenExplorer"
      @close="closeContextMenu"
    />

    <!-- 排除对话框 -->
    <ExcludeDialog
      :visible="excludeDialogVisible"
      :file-path="selectedUntrackedFile?.path || ''"
      @close="excludeDialogVisible = false"
      @confirm="handleExcludeConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import type { StagedFile, UntrackedFile } from '../types'
import ContextMenu from './ContextMenu.vue'
import ExcludeDialog from './ExcludeDialog.vue'
import { OpenInFileExplorer } from '../../wailsjs/go/main/App'

const commitStore = useCommitStore()

// 计算属性
const unstagedCount = computed(() => commitStore.stagingStatus?.unstaged?.length || 0)
const hasUnstagedFiles = computed(() => unstagedCount.value > 0)
const hasUntrackedFiles = computed(() => commitStore.stagingStatus?.untracked?.length || 0 > 0)
const selectedUntrackedCount = computed(() => selectedUntrackedFiles.value.size)

// 选择状态管理
const selectedUntrackedFiles = ref<Set<string>>(new Set())

// 右键菜单状态
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const selectedUntrackedFile = ref<UntrackedFile | null>(null)

// 排除对话框状态
const excludeDialogVisible = ref(false)

// 计算属性
const isAllSelected = computed(() => {
  const unstaged = commitStore.stagingStatus?.unstaged ?? []
  return unstaged.length > 0 && unstaged.every((f: StagedFile) => commitStore.selectedUnstagedFiles.has(f.path))
})

const selectedCount = computed(() => commitStore.selectedUnstagedFiles.size)

// 未暂存文件选择
function isSelected(filePath: string): boolean {
  return commitStore.selectedUnstagedFiles.has(filePath)
}

function toggleSelection(filePath: string) {
  commitStore.toggleFileSelection(filePath, 'unstaged')
}

function toggleSelectAll() {
  const unstaged = commitStore.stagingStatus?.unstaged ?? []
  if (isAllSelected.value) {
    unstaged.forEach((f: StagedFile) => commitStore.selectedUnstagedFiles.delete(f.path))
  } else {
    unstaged.forEach((f: StagedFile) => commitStore.selectedUnstagedFiles.add(f.path))
  }
}

// 未暂存文件暂存操作
async function handleStage(file: StagedFile) {
  try {
    await commitStore.stageFile(file.path)
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageSelected() {
  if (commitStore.selectedUnstagedFiles.size === 0) return

  try {
    const files: string[] = Array.from(commitStore.selectedUnstagedFiles)
    await commitStore.stageFiles(files)
    commitStore.selectedUnstagedFiles.clear()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageAll() {
  const unstagedFiles = commitStore.stagingStatus?.unstaged || []
  if (unstagedFiles.length === 0) return

  try {
    const files = unstagedFiles.filter(f => !f.ignored).map(f => f.path)
    if (files.length > 0) {
      await commitStore.stageFiles(files)
    }
    commitStore.selectedUnstagedFiles.clear()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

// 未跟踪文件选择
function isUntrackedSelected(filePath: string): boolean {
  return selectedUntrackedFiles.value.has(filePath)
}

function toggleUntrackedSelection(filePath: string) {
  if (selectedUntrackedFiles.value.has(filePath)) {
    selectedUntrackedFiles.value.delete(filePath)
  } else {
    selectedUntrackedFiles.value.add(filePath)
  }
}

// 未跟踪文件操作
async function handleStageUntracked(file: UntrackedFile) {
  try {
    await commitStore.stageFile(file.path)
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageUntrackedSelected() {
  if (selectedUntrackedFiles.value.size === 0) return

  try {
    const files: string[] = Array.from(selectedUntrackedFiles.value)
    await commitStore.stageFiles(files)
    selectedUntrackedFiles.value.clear()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

async function stageAllUntracked() {
  const untrackedFiles = commitStore.stagingStatus?.untracked || []
  if (untrackedFiles.length === 0) return

  try {
    const files = untrackedFiles.map(f => f.path)
    await commitStore.stageFiles(files)
    selectedUntrackedFiles.value.clear()
  } catch (e) {
    // 错误已在 store 中处理
  }
}

// 文件点击处理
function handleFileClick(file: StagedFile) {
  commitStore.selectFile(file)
}

function handleUntrackedFileClick(_file: UntrackedFile) {
  // 未跟踪文件不需要显示 diff
  // 可以选择弹出提示或者不做任何操作
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

function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    'Modified': '修改',
    'New': '新增',
    'Deleted': '删除',
    'Renamed': '重命名'
  }
  return texts[status] || '未知'
}

function getStatusClass(status: string): string {
  return status.toLowerCase()
}

// 未跟踪文件右键菜单
function handleUntrackedContextMenu(event: MouseEvent, file: UntrackedFile) {
  selectedUntrackedFile.value = file
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function closeContextMenu() {
  contextMenuVisible.value = false
}

async function handleCopyPath() {
  if (!selectedUntrackedFile.value) return
  try {
    await navigator.clipboard.writeText(selectedUntrackedFile.value.path)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('复制失败:', e)
  }
  closeContextMenu()
}

async function handleStageUntrackedFile() {
  if (!selectedUntrackedFile.value) return
  try {
    await commitStore.stageFiles([selectedUntrackedFile.value.path])
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('添加到暂存区失败:', e)
  }
  closeContextMenu()
}

function handleExcludeUntrackedFile() {
  closeContextMenu()
  excludeDialogVisible.value = true
}

async function handleExcludeConfirm(mode: 'exact' | 'extension' | 'directory', _pattern: string) {
  if (!selectedUntrackedFile.value) return
  try {
    await commitStore.addToGitIgnore(selectedUntrackedFile.value.path, mode)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('添加到排除列表失败:', e)
  }
  excludeDialogVisible.value = false
}

async function handleOpenExplorer() {
  if (!selectedUntrackedFile.value || !commitStore.selectedProjectPath) return
  try {
    const fullPath = `${commitStore.selectedProjectPath}/${selectedUntrackedFile.value.path}`
    await OpenInFileExplorer(fullPath)
    // TODO: 显示 Toast 提示
  } catch (e) {
    console.error('打开失败:', e)
  }
  closeContextMenu()
}
</script>

<style scoped>
.file-list-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
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

/* 已忽略文件的特殊样式 */
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

/* 未跟踪文件样式 */
.file-item.untracked {
  border-left: 3px solid var(--color-info);
}

.status-badge.untracked {
  background: #e0e7ff;
  color: #3730a3;
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

/* 状态徽章样式 */
.status-badge {
  padding: 2px 8px;
  font-size: 10px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
}

.status-badge.modified {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.new {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.deleted {
  background: #fee2e2;
  color: #dc2626;
}

.status-badge.renamed {
  background: #dbeafe;
  color: #2563eb;
}

/* 已忽略徽章样式 */
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

/* 未跟踪部分样式 */
.untracked-section {
  margin-top: var(--space-md);
}
</style>
