<template>
  <button
    @click="openDialog"
    class="extension-status-btn"
    :class="statusClass"
    :title="statusTitle"
  >
    <span class="status-indicator"></span>
  </button>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
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

const statusTitle = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return '扩展未下载'
  if (pushoverStore.isUpdateAvailable) return `有更新可用 (${pushoverStore.extensionInfo.latest_version})`
  return `已更新到 ${pushoverStore.extensionInfo.current_version || '最新版本'}`
})

function openDialog() {
  emit('open')
}

onMounted(async () => {
  await pushoverStore.checkExtensionStatus()
})
</script>

<style scoped>
.extension-status-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid var(--border-default);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.extension-status-btn:hover {
  transform: scale(1.1);
  border-color: var(--border-hover);
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  position: relative;
}

.status-ok .status-indicator {
  background: var(--accent-success, #10b981);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.status-ok .status-indicator::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 8px;
  color: white;
}

.status-update .status-indicator {
  background: var(--accent-warning, #f59e0b);
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
}

.status-update .status-indicator::after {
  content: '↑';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 8px;
  color: white;
}

.status-error .status-indicator {
  background: var(--accent-error, #ef4444);
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.status-error .status-indicator::after {
  content: '!';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 10px;
  color: white;
  font-weight: bold;
}
</style>
