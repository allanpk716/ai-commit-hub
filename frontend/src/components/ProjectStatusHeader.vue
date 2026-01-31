<template>
  <div class="project-status-header">
    <!-- 分支信息、操作按钮组和 Pushover 状态条 -->
    <div class="status-header-top">
      <!-- 左侧：分支和操作按钮 -->
      <div class="header-left">
        <!-- 分支徽章 + 同步状态 -->
        <div class="branch-badge-wrapper">
          <span class="branch-badge">
            <span class="icon">⑂</span>
            {{ branch }}
          </span>

          <!-- 同步状态徽章 -->
          <span v-if="syncStatus" class="sync-status-badge" :class="syncStatusClass">
            {{ syncStatusText }}
          </span>
        </div>

        <!-- 操作按钮组 -->
        <div class="action-buttons-inline">
          <!-- 文件夹按钮 -->
          <button @click="handleOpenInExplorer" class="icon-btn" title="在文件管理器中打开">
            <span class="icon">📁</span>
          </button>

          <!-- 终端按钮：复合设计 -->
          <div class="terminal-button-wrapper" ref="terminalButtonWrapper">
            <button @click="handleOpenInTerminalDirectly" class="icon-btn terminal-btn-main" title="在终端中打开">
              <span class="icon">_>_</span>
            </button>
            <button @click.stop="toggleTerminalMenu" class="icon-btn terminal-btn-dropdown" title="选择终端类型">
              <span class="dropdown-arrow">▼</span>
            </button>
            <!-- 下拉菜单 -->
            <div v-if="showTerminalMenu" class="dropdown-menu terminal-menu">
              <div class="menu-header">在终端中打开</div>
              <div
                v-for="terminal in availableTerminals"
                :key="terminal.id"
                @click="handleOpenInTerminal(terminal.id)"
                class="menu-item"
              >
                <span class="menu-icon">{{ terminal.icon }}</span>
                <span>{{ terminal.name }}</span>
                <span v-if="preferredTerminal === terminal.id" class="check-mark">✓</span>
              </div>
            </div>
          </div>

          <!-- 刷新按钮 -->
          <button @click="handleRefresh" class="icon-btn" title="刷新状态">
            <span class="icon">🔄</span>
          </button>
        </div>
      </div>

      <!-- Pushover 状态条 -->
      <div class="header-right">
        <PushoverStatusRow
          v-if="projectPath"
          :project-path="projectPath"
          :status="pushoverStatus"
          :loading="pushoverLoading"
          @install="handleInstallPushover"
          @update="handleUpdatePushover"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import PushoverStatusRow from './PushoverStatusRow.vue'
import type { HookStatus } from '../types/pushover'
import { useStatusCache } from '@/stores/statusCache'

// Props
interface Props {
  branch: string
  projectPath?: string
  pushoverStatus?: HookStatus | null
  pushoverLoading?: boolean
  availableTerminals: Array<{
    id: string
    name: string
    icon: string
  }>
  preferredTerminal: string
}

const props = withDefaults(defineProps<Props>(), {
  projectPath: undefined,
  pushoverStatus: null,
  pushoverLoading: false
})

// Emits
const emit = defineEmits<{
  openInExplorer: []
  openInTerminal: [terminalId: string]
  openInTerminalDirectly: []
  refresh: []
  installPushover: []
  updatePushover: []
}>()

// 终端菜单状态
const showTerminalMenu = ref(false)
const terminalButtonWrapper = ref<HTMLElement | null>(null)

// StatusCache
const statusCache = useStatusCache()

// 分支同步状态
const syncStatus = computed(() => {
  if (!props.projectPath) return null
  const pushStatus = statusCache.getPushStatus(props.projectPath)
  if (!pushStatus) return null

  const ahead = pushStatus.aheadCount || 0
  const behind = pushStatus.behindCount || 0

  // 如果同步了，不显示徽章
  if (ahead === 0 && behind === 0) return null

  return { ahead, behind }
})

// 同步状态文本
const syncStatusText = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  let text = ''
  if (ahead > 0) text += `↑${ahead}`
  if (behind > 0) text += (text ? ' ' : '') + `↓${behind}`
  return text
})

// 同步状态样式类
const syncStatusClass = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  if (ahead > 0 && behind === 0) return 'status-ahead'
  if (behind > 0 && ahead === 0) return 'status-behind'
  return 'status-diverged'
})

// 切换终端菜单
function toggleTerminalMenu() {
  showTerminalMenu.value = !showTerminalMenu.value
}

// 点击外部关闭菜单
function closeTerminalMenu() {
  showTerminalMenu.value = false
}

// 处理点击外部区域
function handleClickOutside(event: MouseEvent) {
  if (
    terminalButtonWrapper.value &&
    !terminalButtonWrapper.value.contains(event.target as Node)
  ) {
    closeTerminalMenu()
  }
}

// 事件处理函数
function handleOpenInExplorer() {
  emit('openInExplorer')
}

function handleOpenInTerminal(terminalId: string) {
  emit('openInTerminal', terminalId)
  closeTerminalMenu()
}

function handleOpenInTerminalDirectly() {
  emit('openInTerminalDirectly')
}

function handleRefresh() {
  emit('refresh')
}

function handleInstallPushover() {
  emit('installPushover')
}

function handleUpdatePushover() {
  emit('updatePushover')
}

// 只在菜单打开时添加监听器
watch(showTerminalMenu, (newValue) => {
  if (newValue) {
    document.addEventListener('click', handleClickOutside)
  } else {
    document.removeEventListener('click', handleClickOutside)
  }
})

// 暴露关闭菜单方法供父组件调用
defineExpose({
  closeTerminalMenu
})
</script>

<style scoped>
.project-status-header {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
}

.status-header-top {
  display: flex;
  align-items: center;
  padding: var(--space-xs) var(--space-sm);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  min-height: 36px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  flex-shrink: 0;
}

.header-right {
  margin-left: auto;
  flex-shrink: 0;
}

.branch-badge-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.branch-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.branch-badge .icon {
  font-size: 12px;
}

.sync-status-badge {
  padding: 2px 6px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.sync-status-badge.status-ahead {
  background: rgba(16, 185, 129, 0.2);
  color: var(--accent-success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.sync-status-badge.status-behind {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.sync-status-badge.status-diverged {
  background: rgba(239, 68, 68, 0.2);
  color: var(--accent-error);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-buttons-inline {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

/* 终端按钮 hover 特殊样式 */
.terminal-btn-main:hover,
.terminal-btn-dropdown:hover {
  background: var(--accent-primary-bg);
  border-color: var(--accent-primary);
  transform: translateY(-1px);
}

.icon-btn:active {
  transform: translateY(0);
}

.icon-btn .icon {
  font-size: 14px;
}

/* 终端按钮图标颜色 */
.terminal-btn-main .icon {
  color: var(--accent-primary);
}

.terminal-btn-dropdown .dropdown-arrow {
  color: var(--accent-primary);
}

/* 终端按钮复合样式 */
.terminal-button-wrapper {
  display: flex;
  position: relative;
}

.terminal-btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  border-color: var(--accent-primary-border);
}

.terminal-btn-dropdown {
  width: 18px;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 2px;
  padding-right: 2px;
  border-color: var(--accent-primary-border);
}

.dropdown-arrow {
  font-size: 7px;
  color: var(--text-secondary);
}

/* 下拉菜单样式 */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  z-index: 100;
  min-width: 180px;
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.terminal-menu {
  right: 0;
}

.menu-header {
  padding: var(--space-sm) var(--space-md);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  border-bottom: 1px solid var(--border-default);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: background 0.2s;
}

.menu-item:hover {
  background: var(--bg-hover);
}

.menu-icon {
  font-size: 14px;
  width: 20px;
  text-align: center;
}

.check-mark {
  margin-left: auto;
  color: var(--color-primary);
  font-weight: bold;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .status-header-top {
    gap: var(--space-xs);
  }

  .branch-badge {
    font-size: 12px;
    padding: var(--space-xs) var(--space-xs);
  }

  .branch-badge .icon {
    font-size: 13px;
  }

  .icon-btn {
    width: 28px;
    height: 28px;
  }

  .icon-btn .icon {
    font-size: 14px;
  }

  .terminal-btn-dropdown {
    width: 18px;
  }

  .dropdown-menu {
    min-width: 160px;
  }
}
</style>
