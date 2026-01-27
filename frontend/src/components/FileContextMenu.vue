<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="context-menu"
      :style="{ left: x + 'px', top: y + 'px' }"
      @click="close"
    >
      <!-- 复制文件路径 -->
      <div v-if="hasMenuItem('copy-path')" class="menu-item" @click="handleClick('copy-path')">
        <span class="icon">📋</span>
        复制文件路径
      </div>

      <!-- 添加到暂存区 -->
      <div v-if="hasMenuItem('stage')" class="menu-item" @click="handleClick('stage')">
        <span class="icon">✓</span>
        添加到暂存区
      </div>

      <!-- 取消暂存 -->
      <div v-if="hasMenuItem('unstage')" class="menu-item" @click="handleClick('unstage')">
        <span class="icon">✕</span>
        取消暂存
      </div>

      <!-- 还原文件更改 -->
      <div v-if="hasMenuItem('discard')" class="menu-item danger" @click="handleClick('discard')">
        <span class="icon">↩️</span>
        还原文件更改
      </div>

      <!-- 分隔线和排除项（仅未跟踪文件） -->
      <div v-if="hasMenuItem('exclude')" class="menu-divider"></div>
      <div v-if="hasMenuItem('exclude')" class="menu-item" @click="handleClick('exclude')">
        <span class="icon">🚫</span>
        添加到排除列表...
      </div>

      <div class="menu-divider"></div>

      <!-- 在文件管理器中打开 -->
      <div v-if="hasMenuItem('open-explorer')" class="menu-item" @click="handleClick('open-explorer')">
        <span class="icon">📁</span>
        在文件管理器中打开
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

// 直接定义菜单项类型
export type MenuItemType =
  | 'copy-path'
  | 'stage'
  | 'unstage'
  | 'discard'
  | 'exclude'
  | 'open-explorer'

type MenuEvent = {
  (e: MenuItemType): void
  (e: 'close'): void
}

const emit = defineEmits<MenuEvent>()

const props = defineProps<{
  visible: boolean
  x: number
  y: number
  menuItems: MenuItemType[]
}>()

function hasMenuItem(item: MenuItemType): boolean {
  return props.menuItems?.includes(item) || false
}

function handleClick(action: MenuItemType) {
  emit(action)
}

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

.menu-item.danger {
  color: var(--color-error);
}

.menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.1);
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
