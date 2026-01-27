<template>
  <div class="project-status-header">
    <!-- 分支信息和操作按钮组 -->
    <div class="status-header-top">
      <div class="branch-badge">
        <span class="icon">⑂</span>
        {{ branch }}
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
    <PushoverStatusRow
      v-if="projectPath"
      :project-path="projectPath"
      :status="pushoverStatus"
      :loading="pushoverLoading"
      @install="handleInstallPushover"
      @update="handleUpdatePushover"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import PushoverStatusRow from './PushoverStatusRow.vue'

// Props
interface Props {
  branch: string
  projectPath?: string
  pushoverStatus: any
  pushoverLoading: boolean
  availableTerminals: Array<{
    id: string
    name: string
    icon: string
  }>
  preferredTerminal: string
}

const props = defineProps<Props>()

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

// 生命周期钩子
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
  gap: var(--space-sm);
  padding: var(--space-md);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
}

.status-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
}

.branch-badge {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-sm);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.branch-badge .icon {
  font-size: 14px;
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
  width: 32px;
  height: 32px;
  padding: 0;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

.icon-btn:active {
  transform: translateY(0);
}

.icon-btn .icon {
  font-size: 16px;
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
}

.terminal-btn-dropdown {
  width: 20px;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 2px;
  padding-right: 2px;
}

.dropdown-arrow {
  font-size: 8px;
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
    flex-wrap: wrap;
    gap: var(--space-sm);
  }

  .branch-badge {
    font-size: 12px;
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
