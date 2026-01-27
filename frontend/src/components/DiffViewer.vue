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
      <button @click="closeDiff" class="btn-close" title="close">×</button>
    </div>

    <div class="diff-content" v-if="commitStore.selectedFile">
      <div v-if="commitStore.selectedFile.path && !commitStore.selectedFileDiff" class="diff-loading">
        <span class="loading-spinner"></span>
        <span>Loading...</span>
      </div>

      <div v-else-if="commitStore.selectedFileDiff?.diff" class="diff-renderer">
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
        <span>No diff content</span>
      </div>
    </div>

    <div v-else class="diff-placeholder">
      <span class="placeholder-icon">👈</span>
      <span>Click file to view diff</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCommitStore } from '../stores/commitStore'
import { CodeDiff } from 'v-code-diff'

const commitStore = useCommitStore()

function closeDiff() {
  // 直接清空选中的文件和 diff，不触发加载
  commitStore.selectedFile = null
  commitStore.selectedFileDiff = null
}

function getStatusClass(status: string): string {
  return status.toLowerCase()
}

function getOldCode(): string {
  if (!commitStore.selectedFileDiff?.diff) return ''

  // Simple diff parsing, extract old code
  const lines = commitStore.selectedFileDiff.diff.split('\n')
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
  if (!commitStore.selectedFileDiff?.diff) return ''

  // Simple diff parsing, extract new code
  const lines = commitStore.selectedFileDiff.diff.split('\n')
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
