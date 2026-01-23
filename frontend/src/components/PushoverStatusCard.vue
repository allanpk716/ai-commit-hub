<template>
  <div class="pushover-status-card">
    <div class="card-header" @click="collapsed = !collapsed">
      <div class="header-left">
        <span class="header-icon">{{ statusIcon }}</span>
        <span class="header-title">{{ headerTitle }}</span>
      </div>
      <button class="collapse-btn" :class="{ collapsed }">
        <span>{{ collapsed ? '▶' : '▼' }}</span>
      </button>
    </div>

    <div v-if="!collapsed" class="card-body">
      <!-- 未安装状态 -->
      <div v-if="!status || !status.installed" class="status-section not-installed">
        <p class="status-message">Pushover Hook 未安装到此项目</p>
        <button
          class="btn btn-primary"
          :disabled="loading"
          @click="handleInstall"
        >
          {{ loading ? '安装中...' : '安装 Hook' }}
        </button>
      </div>

      <!-- 已安装状态 -->
      <div v-else class="status-section installed">
        <div class="status-info">
          <div class="info-row">
            <span class="info-label">状态:</span>
            <span class="info-value">{{ statusText }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">模式:</span>
            <span class="info-value">{{ modeLabel }}</span>
          </div>
          <div v-if="status.version" class="info-row">
            <span class="info-label">版本:</span>
            <span class="info-value">v{{ status.version }}</span>
          </div>
          <div v-if="status.installed_at" class="info-row">
            <span class="info-label">安装时间:</span>
            <span class="info-value">{{ formatDate(status.installed_at) }}</span>
          </div>
        </div>

        <div class="mode-selector">
          <h4>通知模式</h4>
          <div class="mode-options">
            <button
              v-for="mode in notificationModes"
              :key="mode.value"
              class="mode-btn"
              :class="{ active: status.mode === mode.value }"
              :disabled="loading"
              @click="handleSetMode(mode.value)"
            >
              <span class="mode-icon">{{ mode.icon }}</span>
              <div class="mode-text">
                <span class="mode-label">{{ mode.label }}</span>
                <span class="mode-description">{{ mode.description }}</span>
              </div>
            </button>
          </div>
        </div>

        <div class="card-actions">
          <button
            class="btn btn-secondary"
            :disabled="loading"
            @click="handleUninstall"
          >
            卸载 Hook
          </button>
        </div>
      </div>

      <!-- 错误信息 -->
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { NOTIFICATION_MODES, type NotificationMode } from '../types/pushover'

interface Props {
  projectPath: string
}

const props = defineProps<Props>()

const pushoverStore = usePushoverStore()
const collapsed = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)

// 获取状态
const status = computed(() =>
  pushoverStore.getCachedProjectStatus(props.projectPath)
)

// 状态图标
const statusIcon = computed(() => {
  if (!status.value) return '🔔'
  if (!status.value.installed) return '🔕'
  return '✅'
})

// 标题
const headerTitle = computed(() => {
  if (!status.value) return 'Pushover Hook'
  if (!status.value.installed) return 'Pushover Hook (未安装)'
  return 'Pushover Hook'
})

// 状态文本
const statusText = computed(() => {
  if (!status.value) return '未知'
  if (!status.value.installed) return '未安装'
  return '已安装'
})

// 模式标签
const modeLabel = computed(() => {
  if (!status.value) return ''
  const modeValue = status.value.mode
  if (!modeValue) return ''
  const mode = NOTIFICATION_MODES.find(m => m.value === modeValue)
  return mode?.label || '未知'
})

// 通知模式列表
const notificationModes = NOTIFICATION_MODES

// 格式化日期
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 安装 Hook
async function handleInstall() {
  console.log('[PushoverStatusCard] handleInstall called', props.projectPath)
  error.value = null
  loading.value = true

  try {
    const result = await pushoverStore.installHook(props.projectPath, false)
    console.log('[PushoverStatusCard] installHook result:', result)
    if (!result.success) {
      error.value = result.message || '安装失败'
    }
  } catch (e) {
    console.error('[PushoverStatusCard] installHook error:', e)
    error.value = '安装失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

// 设置通知模式
async function handleSetMode(mode: NotificationMode) {
  if (!status.value || status.value.mode === mode) return

  error.value = null
  loading.value = true

  try {
    await pushoverStore.setNotificationMode(props.projectPath, mode)
  } catch (e) {
    error.value = '设置失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

// 卸载 Hook
async function handleUninstall() {
  if (!confirm('确定要卸载 Pushover Hook 吗？')) return

  error.value = null
  loading.value = true

  try {
    // TODO: 实现卸载功能
    await new Promise(resolve => setTimeout(resolve, 1000))
  } catch (e) {
    error.value = '卸载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.pushover-status-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md);
  cursor: pointer;
  user-select: none;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-default);
  transition: background var(--transition-normal);
}

.card-header:hover {
  background: var(--bg-tertiary);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.header-icon {
  font-size: 18px;
}

.header-title {
  font-weight: 600;
  color: var(--text-primary);
}

.collapse-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: var(--space-xs);
  transition: transform var(--transition-normal);
}

.collapse-btn.collapsed {
  transform: rotate(-90deg);
}

.card-body {
  padding: var(--space-md);
}

.status-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.status-message {
  color: var(--text-muted);
  margin: 0;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  padding: var(--space-sm);
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-label {
  color: var(--text-secondary);
  font-size: 13px;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
}

.mode-selector h4 {
  margin: 0 0 var(--space-sm) 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.mode-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.mode-btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  background: var(--bg-tertiary);
  border: 2px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-normal);
}

.mode-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  background: var(--bg-elevated);
}

.mode-btn.active {
  border-color: var(--accent-primary);
  background: rgba(6, 182, 212, 0.1);
}

.mode-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.mode-icon {
  font-size: 20px;
}

.mode-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.mode-label {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.mode-description {
  font-size: 11px;
  color: var(--text-muted);
}

.card-actions {
  display: flex;
  gap: var(--space-sm);
}

.btn {
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-normal);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.error-message {
  padding: var(--space-sm);
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: var(--radius-sm);
  font-size: 13px;
}
</style>
