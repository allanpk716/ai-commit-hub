<template>
  <div class="staging-area">
    <div class="staging-panels" ref="panelsRef">
      <div class="file-lists-panel" ref="fileListsPanelRef" :style="{ width: leftPanelWidth + 'px' }">
        <StagedList />

        <div
          class="vertical-resizer"
          @mousedown="startVerticalResize"
          :class="{ 'resizing': isVerticalResizing }"
        ></div>

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
import { useCommitStore } from '../stores/commitStore'

const STORAGE_KEY = 'ai-commit-hub:staging-area-left-width'
const VERTICAL_STORAGE_KEY = 'ai-commit-hub:staged-list-height-ratio'
const MIN_WIDTH = 250
const MAX_WIDTH = 600
const DEFAULT_WIDTH = 350
const MIN_LIST_HEIGHT = 150

const leftPanelWidth = ref(DEFAULT_WIDTH)
const isResizing = ref(false)
const stagedListHeightRatio = ref(0.5)
const isVerticalResizing = ref(false)

const panelsRef = ref<HTMLElement | null>(null)
const fileListsPanelRef = ref<HTMLElement | null>(null)

const commitStore = useCommitStore()

// 从 localStorage 加载保存的宽度
onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) {
    const width = parseInt(saved, 10)
    if (!isNaN(width) && width >= MIN_WIDTH && width <= MAX_WIDTH) {
      leftPanelWidth.value = width
    }
  }

  const savedRatio = localStorage.getItem(VERTICAL_STORAGE_KEY)
  if (savedRatio) {
    const ratio = parseFloat(savedRatio)
    if (!isNaN(ratio) && ratio >= 0.2 && ratio <= 0.8) {
      stagedListHeightRatio.value = ratio
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

function startVerticalResize(e: MouseEvent) {
  isVerticalResizing.value = true
  e.preventDefault()
  e.stopPropagation()
}

function handleMouseMove(e: MouseEvent) {
  // 水平拖拽
  if (isResizing.value) {
    const panels = panelsRef.value
    if (!panels) return
    const rect = panels.getBoundingClientRect()
    const newWidth = e.clientX - rect.left

    // 限制在最小和最大宽度之间
    if (newWidth >= MIN_WIDTH && newWidth <= MAX_WIDTH) {
      leftPanelWidth.value = newWidth
    }
  }

  // 垂直拖拽
  if (isVerticalResizing.value) {
    const panel = fileListsPanelRef.value
    if (!panel) return
    const rect = panel.getBoundingClientRect()
    const relativeY = e.clientY - rect.top
    const newRatio = relativeY / rect.height

    // 限制在 20%-80% 之间（考虑最小高度）
    const minHeightRatio = MIN_LIST_HEIGHT / rect.height
    const clampedRatio = Math.max(minHeightRatio, Math.min(1 - minHeightRatio, newRatio))

    if (clampedRatio >= 0.2 && clampedRatio <= 0.8) {
      stagedListHeightRatio.value = clampedRatio
    }
  }
}

function handleMouseUp() {
  if (isResizing.value) {
    isResizing.value = false
    // 保存到 localStorage
    localStorage.setItem(STORAGE_KEY, leftPanelWidth.value.toString())
  }

  if (isVerticalResizing.value) {
    isVerticalResizing.value = false
    localStorage.setItem(VERTICAL_STORAGE_KEY, stagedListHeightRatio.value.toString())
  }
}
</script>

<style scoped>
.staging-area {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 300px;
}

.staging-panels {
  display: flex;
  flex: 1;
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.file-lists-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  overflow: hidden;
  min-height: 0;
  flex: 1;
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

.vertical-resizer {
  height: 4px;
  background: var(--border-default);
  cursor: row-resize;
  flex-shrink: 0;
  transition: background 0.2s;
  position: relative;
  z-index: 10;
  margin: 2px 0;
}

.vertical-resizer:hover {
  background: var(--accent-primary);
}

.vertical-resizer.resizing {
  background: var(--accent-primary);
}

.diff-panel {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
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

  .vertical-resizer {
    height: 1px;
    background: var(--border-default);
    cursor: default;
    pointer-events: none;
    margin: 0;
  }

  .diff-panel {
    min-height: 200px; /* 小屏幕时保持最小高度 */
  }
}
</style>
