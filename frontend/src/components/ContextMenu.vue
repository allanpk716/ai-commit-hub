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
