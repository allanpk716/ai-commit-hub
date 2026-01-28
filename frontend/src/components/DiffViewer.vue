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
import { watch } from 'vue'

const commitStore = useCommitStore()

// 监控 selectedFileDiff 的变化
watch(() => commitStore.selectedFileDiff, (newVal, oldVal) => {
  console.log('[DiffViewer watch] selectedFileDiff 变化:')
  console.log('  旧值:', oldVal ? `filePath=${oldVal.filePath}, diff长度=${oldVal.diff?.length}` : 'null')
  console.log('  新值:', newVal ? `filePath=${newVal.filePath}, diff长度=${newVal.diff?.length}` : 'null')
  console.log('  diff 内容预览:', newVal?.diff ? newVal.diff.substring(0, 100) : '空')
}, { deep: true })

// 监控 selectedFile 的变化
watch(() => commitStore.selectedFile, (newVal, oldVal) => {
  console.log('[DiffViewer watch] selectedFile 变化:')
  console.log('  旧值:', oldVal ? `path=${oldVal.path}` : 'null')
  console.log('  新值:', newVal ? `path=${newVal.path}, status=${newVal.status}` : 'null')
}, { deep: true })

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

  // 检测是否为标准 git diff 格式
  const isStandardDiff = commitStore.selectedFileDiff.diff.includes('diff --git') ||
                         commitStore.selectedFileDiff.diff.includes('@@')

  console.log('[DiffViewer getOldCode] isStandardDiff:', isStandardDiff)

  // 如果不是标准 diff 格式（如未跟踪文件的纯内容），返回空
  if (!isStandardDiff) {
    return ''
  }

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
  if (!commitStore.selectedFileDiff?.diff) {
    console.log('[DiffViewer getNewCode] selectedFileDiff.diff 为空')
    return ''
  }

  // 检测是否为标准 git diff 格式
  const isStandardDiff = commitStore.selectedFileDiff.diff.includes('diff --git') ||
                         commitStore.selectedFileDiff.diff.includes('@@')

  console.log('[DiffViewer getNewCode] isStandardDiff:', isStandardDiff)
  console.log('[DiffViewer getNewCode] diff 长度:', commitStore.selectedFileDiff.diff.length)
  console.log('[DiffViewer getNewCode] diff 内容预览:', commitStore.selectedFileDiff.diff.substring(0, 100))

  // 如果不是标准 diff 格式（如未跟踪文件的纯内容），直接返回完整内容
  if (!isStandardDiff) {
    console.log('[DiffViewer getNewCode] 返回完整内容，长度:', commitStore.selectedFileDiff.diff.length)
    return commitStore.selectedFileDiff.diff
  }

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
  min-height: 300px;
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
