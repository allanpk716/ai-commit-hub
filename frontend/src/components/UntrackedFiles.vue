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
