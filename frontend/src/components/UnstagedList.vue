<template>
  <div class="file-list-container unstaged-list">
    <div class="list-header">
      <h4>未暂存 ({{ commitStore.stagingStatus?.unstaged?.length || 0 }})</h4>
      <div class="bulk-actions" v-if="commitStore.stagingStatus?.unstaged && commitStore.stagingStatus.unstaged.length > 0">
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

    <div class="file-list" v-if="commitStore.stagingStatus?.unstaged && commitStore.stagingStatus.unstaged.length > 0">
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

const isAllSelected = computed(() => {
  const unstaged = commitStore.stagingStatus?.unstaged ?? []
  return unstaged.length > 0 && unstaged.every((f: StagedFile) => commitStore.selectedUnstagedFiles.has(f.path))
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
    unstaged.forEach((f: StagedFile) => commitStore.selectedUnstagedFiles.delete(f.path))
  } else {
    unstaged.forEach((f: StagedFile) => commitStore.selectedUnstagedFiles.add(f.path))
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
</style>
