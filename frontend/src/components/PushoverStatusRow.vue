<template>
  <div class="pushover-status-row">
    <div class="status-left">
      <span class="status-icon">{{ statusIcon }}</span>
      <span class="status-title">Pushover Hook</span>
      <span v-if="status?.version" class="status-version">{{ formatVersion(status.version) }}</span>
      <span v-if="!status?.installed" class="status-text">(未安装)</span>
    </div>

    <div v-if="status?.installed" class="notification-toggles">
      <button
        class="notify-btn"
        :class="{ active: isPushoverEnabled, disabled: !isPushoverEnabled }"
        :title="pushoverTooltip"
        :disabled="loading"
        @click="togglePushover"
      >
        <span class="notify-icon">📱</span>
      </button>
      <button
        class="notify-btn"
        :class="{ active: isWindowsEnabled, disabled: !isWindowsEnabled }"
        :title="windowsTooltip"
        :disabled="loading"
        @click="toggleWindows"
      >
        <span class="notify-icon">💻</span>
      </button>
    </div>

    <div class="status-right">
      <span v-if="isLatest && status?.installed" class="latest-badge">已是最新</span>
      <button
        v-else-if="!status?.installed"
        class="action-btn btn-primary"
        :disabled="loading"
        @click="handleInstall"
      >
        {{ loading ? '处理中...' : '安装 Hook' }}
      </button>
      <button
        v-else-if="needsUpdate"
        class="action-btn btn-update"
        :disabled="loading"
        @click="handleUpdate"
      >
        {{ loading ? '更新中...' : '更新 Hook' }}
      </button>
      <!-- 已是最新版本时显示重装按钮 -->
      <button
        v-else
        class="action-btn btn-reinstall"
        :disabled="loading"
        @click="handleReinstall"
      >
        {{ loading ? '重装中...' : '重装 Hook' }}
      </button>
    </div>
  </div>

  <!-- 重装确认对话框 -->
  <div v-if="showReinstallDialog" class="dialog-overlay" @click="closeReinstallDialog">
    <div class="dialog-content" @click.stop>
      <h3>重装 Pushover Hook</h3>
      <p class="dialog-description">
        这将重新安装 Pushover Hook 到该项目：
      </p>
      <ul class="dialog-list">
        <li>使用最新版本的 Hook 文件覆盖当前安装</li>
        <li>保留您的通知配置（Pushover/Windows 通知设置）</li>
      </ul>
      <div class="dialog-actions">
        <button
          class="dialog-btn btn-cancel"
          @click="closeReinstallDialog"
        >
          取消
        </button>
        <button
          class="dialog-btn btn-confirm"
          :disabled="localLoading"
          @click="confirmReinstall"
        >
          确定重装
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { formatVersion } from '../utils/versionFormat'
import type { HookStatus } from '../types/pushover'

interface Props {
  projectPath: string
  status?: HookStatus | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  install: []
  update: []
}>()

const pushoverStore = usePushoverStore()
const localLoading = ref(false)
const updateInfo = ref<{ updateAvailable: boolean; currentVersion: string; latestVersion: string } | null>(null)
const showReinstallDialog = ref(false)

const statusIcon = computed(() => {
  if (!props.status?.installed) return '🔴'
  if (needsUpdate.value) return '🟡'
  return '🟢'
})

const isPushoverEnabled = computed(() => {
  if (!props.status) return false
  return props.status.mode === 'enabled' || props.status.mode === 'pushover_only'
})

const isWindowsEnabled = computed(() => {
  if (!props.status) return false
  return props.status.mode === 'enabled' || props.status.mode === 'windows_only'
})

const needsUpdate = computed(() => {
  if (!props.status?.installed) return false
  return props.status.version === 'unknown' ||
         (updateInfo.value?.updateAvailable)
})

const isLatest = computed(() => {
  return props.status?.installed &&
         props.status.version !== 'unknown' &&
         !needsUpdate.value
})

const pushoverTooltip = computed(() => {
  return isPushoverEnabled.value ? '点击禁用 Pushover 通知' : '点击启用 Pushover 通知'
})

const windowsTooltip = computed(() => {
  return isWindowsEnabled.value ? '点击禁用 Windows 通知' : '点击启用 Windows 通知'
})

async function togglePushover() {
  if (localLoading.value) return
  localLoading.value = true
  try {
    await pushoverStore.toggleNotification(props.projectPath, 'pushover')
  } finally {
    localLoading.value = false
  }
}

async function toggleWindows() {
  if (localLoading.value) return
  localLoading.value = true
  try {
    await pushoverStore.toggleNotification(props.projectPath, 'windows')
  } finally {
    localLoading.value = false
  }
}

function handleInstall() {
  emit('install')
}

function handleUpdate() {
  emit('update')
}

function handleReinstall() {
  showReinstallDialog.value = true
}

function closeReinstallDialog() {
  showReinstallDialog.value = false
}

async function confirmReinstall() {
  if (localLoading.value) return

  localLoading.value = true
  try {
    const result = await pushoverStore.reinstallHook(props.projectPath)

    if (result.success) {
      // 关闭对话框
      closeReinstallDialog()
      // 可选：显示成功提示
      console.log('重装成功:', result.message)
    } else {
      // 显示错误信息
      console.error('重装失败:', result.message)
    }
  } catch (e) {
    console.error('重装 Hook 失败:', e)
  } finally {
    localLoading.value = false
  }
}

async function checkForUpdates() {
  if (!props.status?.installed) return
  try {
    updateInfo.value = await pushoverStore.checkForUpdates(props.projectPath)
  } catch (e) {
    console.error('检查更新失败:', e)
  }
}

watch(() => props.status, (newStatus) => {
  if (newStatus?.installed) {
    checkForUpdates()
  }
}, { immediate: true })
</script>

<style scoped>
.pushover-status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2px 8px;
  background: transparent;
  border: none;
  gap: var(--space-sm);
  margin-bottom: 0;
  transition: all var(--transition-fast);
}

.pushover-status-row:hover {
  /* 移除 hover 效果 */
}

.status-left {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

.status-icon {
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}

.status-title {
  font-weight: 500;
  font-size: 12px;
  color: var(--text-primary);
  white-space: nowrap;
}

.status-version {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 1px 4px;
  border-radius: 3px;
  white-space: nowrap;
}

.status-text {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
}

.notification-toggles {
  display: flex;
  gap: var(--space-xs);
  flex-shrink: 0;
}

.notify-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 2px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
  padding: 0;
}

.notify-btn:hover:not(:disabled) {
  transform: scale(1.1);
  border-color: var(--accent-primary);
}

.notify-btn.active {
  border-color: var(--accent-primary);
  background: rgba(6, 182, 212, 0.15);
}

.notify-btn.disabled {
  opacity: 0.4;
  filter: grayscale(1);
}

.notify-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.notify-icon {
  font-size: 18px;
  line-height: 1;
}

.status-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.latest-badge {
  font-size: 12px;
  color: var(--text-muted);
  padding: var(--space-xs) var(--space-sm);
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.action-btn {
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.btn-update {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.btn-update:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.3);
}

/* 重装按钮样式 */
.btn-reinstall {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.btn-reinstall:hover:not(:disabled) {
  background: rgba(6, 182, 212, 0.25);
}

/* 对话框样式 */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  max-width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.dialog-content h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.dialog-description {
  margin: 0 0 var(--space-sm) 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-list {
  margin: 0 0 var(--space-md) 0;
  padding-left: var(--space-lg);
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-list li {
  margin-bottom: var(--space-xs);
}

.dialog-actions {
  display: flex;
  gap: var(--space-sm);
  justify-content: flex-end;
}

.dialog-btn {
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-fast);
}

.dialog-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.btn-confirm {
  background: var(--accent-primary);
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  background: var(--accent-secondary);
}
</style>
