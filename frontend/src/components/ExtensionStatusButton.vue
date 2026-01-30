<template>
  <button
    @click="openDialog"
    class="extension-status-btn"
    :class="statusClass"
    :title="statusTitle"
  >
    <span class="btn-icon">🔔</span>
    <span class="btn-text">Pushover 扩展</span>
    <span class="status-badge">{{ statusBadge }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'

const emit = defineEmits<{
  open: []
}>()

const pushoverStore = usePushoverStore()

const statusClass = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return 'status-error'
  if (pushoverStore.isUpdateAvailable) return 'status-update'
  return 'status-ok'
})

const statusBadge = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return '!'
  if (pushoverStore.isUpdateAvailable) return '↑'
  return '✓'
})

const statusTitle = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return 'cc-pushover-hook 扩展未下载'
  if (pushoverStore.isUpdateAvailable)
    return `cc-pushover-hook 有更新可用 (v${pushoverStore.extensionInfo.latest_version})`
  return `cc-pushover-hook 已是最新版本 (v${pushoverStore.extensionInfo.current_version})`
})

function openDialog() {
  emit('open')
}
</script>

<style scoped>
.extension-status-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-md);
  border: none;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  color: white;
  min-width: 120px;
}

.extension-status-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.btn-icon {
  font-size: 14px;
}

.btn-text {
  flex: 1;
}

.status-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  font-size: 11px;
  font-weight: bold;
}

/* 状态变体 */
.status-ok {
  background: linear-gradient(135deg, #10b981, #059669);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.status-update {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.status-error {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}
</style>
