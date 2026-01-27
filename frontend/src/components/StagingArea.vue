<template>
  <div class="staging-area">
    <div class="staging-panels" ref="panelsRef">
      <div class="file-lists-panel" :style="{ width: leftPanelWidth + 'px' }">
        <StagedList />
        <div class="list-divider"></div>
        <UnstagedList />
      </div>

      <div
        class="resizer"
        @mousedown="startResize"
        :class="{ 'resizing': isResizing }"
      ></div>

      <div class="diff-panel">
        <DiffViewer />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import StagedList from './StagedList.vue'
import UnstagedList from './UnstagedList.vue'
import DiffViewer from './DiffViewer.vue'

const STORAGE_KEY = 'ai-commit-hub:staging-area-left-width'
const MIN_WIDTH = 250
const MAX_WIDTH = 600
const DEFAULT_WIDTH = 350

const leftPanelWidth = ref(DEFAULT_WIDTH)
const isResizing = ref(false)

// 从 localStorage 加载保存的宽度
onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) {
    const width = parseInt(saved, 10)
    if (!isNaN(width) && width >= MIN_WIDTH && width <= MAX_WIDTH) {
      leftPanelWidth.value = width
    }
  }

  // 添加全局事件监听
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
})

function startResize(e: MouseEvent) {
  isResizing.value = true
  e.preventDefault()
}

function handleMouseMove(e: MouseEvent) {
  if (!isResizing.value) return

  const panels = panelsRef.value
  if (!panels) return

  const rect = panels.getBoundingClientRect()
  const newWidth = e.clientX - rect.left

  // 限制在最小和最大宽度之间
  if (newWidth >= MIN_WIDTH && newWidth <= MAX_WIDTH) {
    leftPanelWidth.value = newWidth
  }
}

function handleMouseUp() {
  if (isResizing.value) {
    isResizing.value = false
    // 保存到 localStorage
    localStorage.setItem(STORAGE_KEY, leftPanelWidth.value.toString())
  }
}

const panelsRef = ref<HTMLElement | null>(null)
</script>

<style scoped>
.staging-area {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.staging-panels {
  display: flex;
  flex: 1;
  overflow: hidden;
  min-height: 0;
  position: relative;
}

.file-lists-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  overflow: hidden;
  min-height: 0;
  flex-shrink: 0;
}

/* 可拖拽的分割条 */
.resizer {
  width: 4px;
  background: var(--border-default);
  cursor: col-resize;
  flex-shrink: 0;
  transition: background 0.2s;
  position: relative;
  z-index: 10;
}

.resizer:hover {
  background: var(--accent-primary);
}

.resizer.resizing {
  background: var(--accent-primary);
}

/* 添加拖拽时的视觉反馈 */
.resizer::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 12px;
  transform: translateX(-50%);
  background: transparent;
}

.resizer:hover::after,
.resizer.resizing::after {
  background: rgba(6, 182, 212, 0.1);
}

.list-divider {
  height: 1px;
  background: var(--border-default);
  flex-shrink: 0;
}

.diff-panel {
  flex: 1;
  overflow: hidden;
  min-height: 0;
  min-width: 0;
}

@media (max-width: 1024px) {
  .staging-panels {
    flex-direction: column;
  }

  .resizer {
    width: 100%;
    height: 4px;
    cursor: row-resize;
  }

  .file-lists-panel {
    width: 100% !important;
  }

  .diff-panel {
    min-height: 200px;
  }
}
</style>
