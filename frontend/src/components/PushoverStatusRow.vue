<template>
  <div class="pushover-status-row">
    <!-- 状态指示器 -->
    <div class="status-indicator" :class="statusClass" :title="statusTitle">
      <span class="status-icon">{{ statusIcon }}</span>
    </div>

    <!-- 状态信息 -->
    <div class="status-info">
      <div class="status-main">
        <span class="status-label">{{ statusLabel }}</span>
        <span v-if="status?.version && status.version !== 'unknown'" class="status-version">
          v{{ status.version }}
        </span>
      </div>
      <div v-if="status?.installed" class="status-details">
        <span class="mode-badge" :class="modeClass">{{ modeLabel }}</span>
        <span v-if="updateAvailable" class="update-badge" title="有新版本可用">
          🔄
        </span>
      </div>
    </div>

    <!-- 快速操作按钮 -->
    <div class="status-actions">
      <!-- 未安装时显示安装按钮 -->
      <button
        v-if="!status?.installed"
        class="action-btn btn-install"
        :disabled="loading"
        @click="handleInstall"
        title="安装 Pushover Hook"
      >
        <span>{{ loading ? '安装中...' : '安装' }}</span>
      </button>

      <!-- 已安装时显示操作按钮 -->
      <template v-else>
        <!-- 切换通知按钮 -->
        <button
          class="action-btn btn-toggle"
          :disabled="loading"
          @click="handleToggle"
          :title="toggleTitle"
        >
          <span>{{ toggleIcon }}</span>
        </button>

        <!-- 更多操作菜单按钮 -->
        <button
          class="action-btn btn-more"
          :disabled="loading"
          @click="handleMoreClick"
          title="更多操作"
        >
          <span>⋮</span>
        </button>
      </template>
    </div>

    <!-- 更多操作菜单 -->
    <div v-if="showMenu" class="action-menu" @click.stop>
      <button
        v-if="updateAvailable"
        class="menu-item"
        :disabled="loading"
        @click="handleUpdate"
      >
        <span class="menu-icon">🔄</span>
        <span>更新 Hook</span>
      </button>
      <button
        class="menu-item"
        :disabled="loading"
        @click="handleConfigure"
      >
        <span class="menu-icon">⚙️</span>
        <span>配置通知模式</span>
      </button>
      <button
        class="menu-item menu-item-danger"
        :disabled="loading"
        @click="handleUninstall"
      >
        <span class="menu-icon">🗑️</span>
        <span>卸载 Hook</span>
      </button>
    </div>

    <!-- 遮罩层，点击关闭菜单 -->
    <div v-if="showMenu" class="menu-overlay" @click="showMenu = false"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { NOTIFICATION_MODES, type NotificationMode } from '../types/pushover'

interface Props {
  projectPath: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  configure: [projectPath: string]
}>()

const pushoverStore = usePushoverStore()
const loading = ref(false)
const showMenu = ref(false)
const updateCheckResult = ref<{ updateAvailable: boolean; currentVersion: string; latestVersion: string } | null>(null)

// 获取状态
const status = computed(() => {
  return pushoverStore.getCachedProjectStatus(props.projectPath)
})

// 是否有更新可用
const updateAvailable = computed(() => {
  return updateCheckResult.value?.updateAvailable ?? false
})

// 状态图标
const statusIcon = computed(() => {
  if (!status.value) return '🔕'
  if (!status.value.installed) return '🔕'
  return '🔔'
})

// 状态类名
const statusClass = computed(() => {
  if (!status.value) return 'status-unknown'
  if (!status.value.installed) return 'status-uninstalled'
  return 'status-installed'
})

// 状态标题
const statusTitle = computed(() => {
  if (!status.value) return 'Pushover Hook 状态未知'
  if (!status.value.installed) return 'Pushover Hook 未安装'
  return 'Pushover Hook 已安装'
})

// 状态标签
const statusLabel = computed(() => {
  if (!status.value) return '状态未知'
  if (!status.value.installed) return '未安装'
  return '已安装'
})

// 模式标签
const modeLabel = computed(() => {
  if (!status.value || !status.value.installed) return ''
  const mode = NOTIFICATION_MODES.find(m => m.value === status.value.mode)
  return mode?.label || '未知模式'
})

// 模式类名
const modeClass = computed(() => {
  if (!status.value || !status.value.installed) return ''
  return `mode-${status.value.mode}`
})

// 切换按钮图标
const toggleIcon = computed(() => {
  if (!status.value) return ''
  switch (status.value.mode) {
    case 'enabled':
      return '🔔'
    case 'pushover_only':
      return '📱'
    case 'windows_only':
      return '💻'
    case 'disabled':
      return '🔕'
    default:
      return '❓'
  }
})

// 切换按钮标题
const toggleTitle = computed(() => {
  if (!status.value) return ''
  const mode = NOTIFICATION_MODES.find(m => m.value === status.value.mode)
  return mode?.description || '切换通知模式'
})

// 处理安装
async function handleInstall() {
  loading.value = true
  try {
    const result = await pushoverStore.installHook(props.projectPath, false)
    if (!result.success) {
      console.error('安装失败:', result.message)
    }
  } catch (e) {
    console.error('安装失败:', e)
  } finally {
    loading.value = false
  }
}

// 处理切换通知模式
async function handleToggle() {
  if (!status.value || !status.value.installed) return

  loading.value = true
  try {
    // 循环切换模式: enabled -> pushover_only -> windows_only -> disabled -> enabled
    const modes: NotificationMode[] = ['enabled', 'pushover_only', 'windows_only', 'disabled']
    const currentIndex = modes.indexOf(status.value.mode)
    const nextMode = modes[(currentIndex + 1) % modes.length]

    await pushoverStore.setNotificationMode(props.projectPath, nextMode)
  } catch (e) {
    console.error('切换模式失败:', e)
  } finally {
    loading.value = false
  }
}

// 处理更多操作点击
function handleMoreClick() {
  showMenu.value = !showMenu.value
}

// 处理更新
async function handleUpdate() {
  showMenu.value = false
  if (!confirm('确定要更新此项目的 Pushover Hook 吗？')) return

  loading.value = true
  try {
    const result = await pushoverStore.updateHook(props.projectPath)
    if (!result.success) {
      console.error('更新失败:', result.message)
    } else {
      // 重新检查更新
      await checkUpdates()
    }
  } catch (e) {
    console.error('更新失败:', e)
  } finally {
    loading.value = false
  }
}

// 处理配置
function handleConfigure() {
  showMenu.value = false
  emit('configure', props.projectPath)
}

// 处理卸载
async function handleUninstall() {
  showMenu.value = false
  if (!confirm('确定要卸载 Pushover Hook 吗？')) return

  loading.value = true
  try {
    // TODO: 实现卸载功能
    await new Promise(resolve => setTimeout(resolve, 1000))
    console.log('卸载功能待实现')
  } catch (e) {
    console.error('卸载失败:', e)
  } finally {
    loading.value = false
  }
}

// 检查更新
async function checkUpdates() {
  try {
    const result = await pushoverStore.checkForUpdates(props.projectPath)
    updateCheckResult.value = result
  } catch (e) {
    console.error('检查更新失败:', e)
    updateCheckResult.value = null
  }
}

// 点击外部关闭菜单
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  const menu = document.querySelector('.action-menu')
  const moreBtn = document.querySelector('.btn-more')

  if (menu && moreBtn && !menu.contains(target) && !moreBtn.contains(target)) {
    showMenu.value = false
  }
}

// 组件挂载
onMounted(async () => {
  // 加载项目状态
  await pushoverStore.getProjectHookStatus(props.projectPath)

  // 如果已安装，检查更新
  if (status.value?.installed) {
    await checkUpdates()
  }

  // 添加全局点击监听
  document.addEventListener('click', handleClickOutside)
})

// 组件卸载
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.pushover-status-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  position: relative;
  transition: all var(--transition-normal);
}

.pushover-status-row:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

/* 状态指示器 */
.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.status-icon {
  font-size: 16px;
}

.status-installed {
  background: rgba(16, 185, 129, 0.1);
}

.status-uninstalled {
  background: rgba(156, 163, 175, 0.1);
}

.status-unknown {
  background: rgba(239, 68, 68, 0.1);
}

/* 状态信息 */
.status-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.status-main {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.status-label {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.status-version {
  font-size: 11px;
  color: var(--text-muted);
  background: var(--bg-tertiary);
  padding: 1px 4px;
  border-radius: var(--radius-xs);
}

.status-details {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.mode-badge {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  font-weight: 500;
}

.mode-enabled {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.mode-pushover_only {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.mode-windows_only {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}

.mode-disabled {
  background: rgba(156, 163, 175, 0.15);
  color: #9ca3af;
}

.update-badge {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  font-weight: 500;
}

/* 操作按钮 */
.status-actions {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  flex-shrink: 0;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-xs);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-normal);
  font-size: 14px;
  min-width: 32px;
  height: 32px;
}

.action-btn:hover:not(:disabled) {
  background: var(--bg-elevated);
  border-color: var(--accent-primary);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-install {
  font-size: 12px;
  padding: 0 var(--space-sm);
  background: var(--accent-primary);
  color: white;
  border: none;
}

.btn-install:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.btn-toggle {
  font-size: 16px;
}

.btn-more {
  font-size: 18px;
  font-weight: bold;
}

/* 更多操作菜单 */
.action-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 4px);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 100;
  min-width: 160px;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background: none;
  border: none;
  cursor: pointer;
  transition: all var(--transition-normal);
  font-size: 13px;
  color: var(--text-primary);
  text-align: left;
}

.menu-item:hover:not(:disabled) {
  background: var(--bg-tertiary);
}

.menu-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.menu-item-danger {
  color: #ef4444;
}

.menu-item-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.1);
}

.menu-icon {
  font-size: 14px;
  width: 20px;
  text-align: center;
}

/* 遮罩层 */
.menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99;
}
</style>
