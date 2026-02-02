<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="update-dialog">
      <div class="dialog-header">
        <h2>发现新版本</h2>
        <button @click="close" class="close-btn">&times;</button>
      </div>

      <div class="dialog-body">
        <div class="version-comparison">
          <div class="version-item">
            <span class="label">当前版本:</span>
            <span class="value">{{ updateInfo?.currentVersion }}</span>
          </div>
          <div class="version-item">
            <span class="label">最新版本:</span>
            <span class="value highlight">{{ updateInfo?.latestVersion }}</span>
          </div>
        </div>

        <div class="release-notes">
          <h3>更新内容</h3>
          <div class="notes-content" v-html="formattedReleaseNotes"></div>
        </div>

        <div class="file-info">
          <span>文件大小: {{ formattedSize }}</span>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="download" class="btn-download" :disabled="isDownloading">
          {{ isDownloading ? '下载中...' : '立即更新' }}
        </button>
        <button @click="skip" class="btn-skip">跳过此版本</button>
        <button @click="close" class="btn-cancel">稍后提醒</button>
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
const isDownloading = computed(() => updateStore.isDownloading)

const updateInfo = computed(() => updateStore.updateInfo)

const formattedReleaseNotes = computed(() => {
  if (!updateInfo.value?.releaseNotes) return ''
  // 简单的 Markdown 转换
  return formatMarkdown(updateInfo.value.releaseNotes)
})

const formattedSize = computed(() => updateStore.formattedSize)

function formatMarkdown(markdown: string): string {
  if (!markdown) return ''

  // 转换标题
  markdown = markdown.replace(/^### (.*$)/gim, '<h3>$1</h3>')
  markdown = markdown.replace(/^## (.*$)/gim, '<h2>$1</h2>')
  markdown = markdown.replace(/^# (.*$)/gim, '<h1>$1</h1>')

  // 转换列表
  markdown = markdown.replace(/^\* (.*$)/gim, '<li>$1</li>')
  markdown = markdown.replace(/^- (.*$)/gim, '<li>$1</li>')

  // 转换代码块
  markdown = markdown.replace(/`([^`]+)`/g, '<code>$1</code>')

  // 转换换行
  markdown = markdown.replace(/\n\n/g, '</p><p>')
  markdown = markdown.replace(/\n/g, '<br>')

  return `<p>${markdown}</p>`
}

function close() {
  emit('close')
}

function skip() {
  updateStore.skipVersion(updateInfo.value?.latestVersion || '')
  close()
}

async function download() {
  // TODO: 实现下载逻辑
  console.log('开始下载更新')
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

.update-dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 32px;
  cursor: pointer;
  color: #6b7280;
  line-height: 1;
  transition: color 0.2s;
}

.close-btn:hover {
  color: #1f2937;
}

.dialog-body {
  padding: 24px;
}

.version-comparison {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-item .label {
  color: #6b7280;
  font-weight: 500;
}

.version-item .value {
  font-weight: 600;
  color: #1f2937;
}

.version-item .value.highlight {
  color: #667eea;
  font-size: 18px;
}

.release-notes {
  margin-bottom: 24px;
}

.release-notes h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
}

.notes-content {
  color: #374151;
  line-height: 1.6;
  max-height: 300px;
  overflow-y: auto;
}

.notes-content :deep(h1),
.notes-content :deep(h2),
.notes-content :deep(h3) {
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: 600;
}

.notes-content :deep(h1) {
  font-size: 20px;
}

.notes-content :deep(h2) {
  font-size: 18px;
}

.notes-content :deep(h3) {
  font-size: 16px;
}

.notes-content :deep(ul),
.notes-content :deep(ol) {
  margin: 8px 0;
  padding-left: 24px;
}

.notes-content :deep(li) {
  margin: 4px 0;
}

.notes-content :deep(code) {
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 14px;
  font-family: 'Courier New', monospace;
}

.file-info {
  color: #6b7280;
  font-size: 14px;
  text-align: center;
}

.dialog-footer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn-download,
.btn-skip,
.btn-cancel {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-download {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-download:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-download:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-skip,
.btn-cancel {
  background: white;
  color: #6b7280;
  border: 1px solid #d1d5db;
}

.btn-skip:hover,
.btn-cancel:hover {
  background: #f9fafb;
}
</style>
