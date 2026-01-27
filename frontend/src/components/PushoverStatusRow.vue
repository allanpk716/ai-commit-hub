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
  padding: var(--space-md) var(--space-lg);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  gap: var(--space-md);
  margin-bottom: var(--space-md);
  transition: all var(--transition-fast);
}

.pushover-status-row:hover {
  border-color: var(--border-hover);
}

.status-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex: 1;
  min-width: 0;
}

.status-icon {
  font-size: 16px;
  line-height: 1;
  flex-shrink: 0;
}

.status-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

.status-version {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.status-text {
  font-size: 13px;
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
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 13px;
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
</style>
