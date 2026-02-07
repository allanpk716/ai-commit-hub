<template>
  <div v-if="visible" class="modal-overlay">
    <div class="download-dialog">
      <div class="dialog-header">
        <h2>正在下载更新</h2>
      </div>

      <div class="dialog-body">
        <!-- 进度条 -->
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>

        <!-- 百分比 -->
        <div class="progress-text">{{ progress.toFixed(1) }}%</div>

        <!-- 已下载/总大小 -->
        <div class="size-info">{{ formatBytes(downloaded) }} / {{ formatBytes(total) }}</div>

        <!-- 速度 -->
        <div class="speed-info">速度: {{ formatSpeed(speed) }}</div>

        <!-- 剩余时间 -->
        <div class="eta-info">预计剩余: {{ eta }}</div>
      </div>

      <div class="dialog-footer">
        <button @click="cancel" class="btn-cancel">取消</button>
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

// 从 store 获取下载状态
const progress = computed(() => updateStore.downloadProgress)
const downloaded = computed(() => updateStore.downloadedSize)
const total = computed(() => updateStore.totalSize)
const speed = computed(() => updateStore.downloadSpeed)
const eta = computed(() => updateStore.downloadETA)

// 使用 store 的工具方法
function formatBytes(bytes: number): string {
  return updateStore.formatBytes(bytes)
}

function formatSpeed(bytesPerSecond: number): string {
  return updateStore.formatSpeed(bytesPerSecond)
}

function cancel() {
  updateStore.cancelDownload()
  emit('close')
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

.download-dialog {
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
  padding: 32px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.progress-bar {
  width: 100%;
  height: 12px;
  background: #f3f4f6;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s ease;
  border-radius: 6px;
}

.progress-text {
  font-size: 32px;
  font-weight: 700;
  color: #667eea;
  margin: 8px 0;
}

.size-info,
.speed-info,
.eta-info {
  font-size: 16px;
  color: #6b7280;
  text-align: center;
}

.size-info {
  font-weight: 600;
  color: #1f2937;
}

.speed-info {
  font-family: 'Courier New', monospace;
}

.dialog-footer {
  display: flex;
  justify-content: center;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn-cancel {
  padding: 12px 32px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
  color: #6b7280;
}

.btn-cancel:hover {
  background: #f9fafb;
  border-color: #9ca3af;
}

.btn-cancel:active {
  transform: translateY(0);
}
</style>
