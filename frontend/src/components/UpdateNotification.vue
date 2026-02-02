<template>
  <div v-if="updateStore.hasUpdate" class="update-notification">
    <div class="notification-content">
      <div class="notification-icon">🔄</div>
      <div class="notification-text">
        发现新版本 {{ updateStore.displayVersion }}
      </div>
      <div class="notification-actions">
        <button @click="showDetails" class="btn-primary">查看详情</button>
        <button @click="ignore" class="btn-secondary">忽略</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUpdateStore } from '../stores/updateStore'

const updateStore = useUpdateStore()

const emit = defineEmits(['show-update-dialog'])

function showDetails() {
  emit('show-update-dialog')
}

function ignore() {
  updateStore.skipVersion(updateStore.updateInfo?.latestVersion || '')
}
</script>

<style scoped>
.update-notification {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.notification-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-icon {
  font-size: 24px;
}

.notification-text {
  flex: 1;
  font-weight: 500;
  font-size: 15px;
}

.notification-actions {
  display: flex;
  gap: 8px;
}

.btn-primary,
.btn-secondary {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
  font-weight: 500;
}

.btn-primary {
  background: white;
  color: #667eea;
}

.btn-primary:hover {
  background: #f0f0f0;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
