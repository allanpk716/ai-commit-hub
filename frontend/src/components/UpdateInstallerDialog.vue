<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="install-dialog">
      <div class="dialog-header">
        <h2>准备安装更新</h2>
      </div>
      <div class="dialog-body">
        <div class="warning-box">
          <div class="warning-icon">!</div>
          <div class="warning-text">
            <p>安装更新时，应用将关闭并重新启动。</p>
            <p>请确保已保存当前的工作。</p>
          </div>
        </div>
        <div class="update-info">
          <div class="info-row">
            <span class="label">版本:</span>
            <span class="value">{{ updateInfo?.latestVersion }}</span>
          </div>
          <div class="info-row">
            <span class="label">文件大小:</span>
            <span class="value">{{ formattedSize }}</span>
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button @click="confirm" class="btn-confirm" :disabled="isInstalling">
          {{ isInstalling ? '正在启动...' : '立即安装' }}
        </button>
        <button @click="close" class="btn-later" :disabled="isInstalling">
          稍后
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUpdateStore } from '../stores/updateStore'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits(['close'])

const updateStore = useUpdateStore()
const updateInfo = computed(() => updateStore.updateInfo)
const isInstalling = computed(() => updateStore.isInstalling)
const formattedSize = computed(() => updateStore.formattedSize)

function close() {
  emit('close')
}

async function confirm() {
  try {
    await updateStore.confirmInstall()
  } catch (error) {
    // 错误已在 store 中处理
  }
}
</script>

<style scoped>
.modal-overlay {
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

.install-dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.dialog-header {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.dialog-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.warning-box {
  display: flex;
  gap: 16px;
  padding: 16px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
}

.warning-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f59e0b;
  color: white;
  border-radius: 50%;
  font-size: 18px;
  font-weight: 700;
}

.warning-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.warning-text p {
  margin: 0;
  font-size: 14px;
  color: #92400e;
  line-height: 1.5;
}

.update-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row .label {
  color: #6b7280;
  font-weight: 500;
  font-size: 14px;
}

.info-row .value {
  font-weight: 600;
  color: #1f2937;
  font-size: 14px;
}

.dialog-footer {
  display: flex;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn-confirm,
.btn-later {
  flex: 1;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-later {
  background: white;
  color: #6b7280;
  border: 1px solid #d1d5db;
}

.btn-later:hover:not(:disabled) {
  background: #f9fafb;
}

.btn-later:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
