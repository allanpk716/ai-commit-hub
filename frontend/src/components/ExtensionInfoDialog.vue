<template>
  <div v-if="open" class="dialog-overlay" @click.self="handleClose">
    <div class="dialog">
      <!-- Header -->
      <div class="dialog-header">
        <h2>cc-pushover-hook 扩展信息</h2>
        <button class="close-btn" @click="handleClose">×</button>
      </div>

      <!-- Body -->
      <div class="dialog-body">
        <!-- 状态卡片 -->
        <div class="status-card">
          <h3>扩展状态</h3>
          <div class="info-row">
            <span class="label">下载状态:</span>
            <span class="value" :class="statusClass">
              {{ pushoverStore.isExtensionDownloaded ? '已下载' : '未下载' }}
            </span>
          </div>
          <div class="info-row">
            <span class="label">扩展路径:</span>
            <span
              v-if="pushoverStore.extensionInfo.path"
              class="value value-path clickable"
              @click="handleOpenExtensionFolder"
            >
              {{ pushoverStore.extensionInfo.path }}
            </span>
            <span v-else class="value value-path">-</span>
          </div>
        </div>

        <!-- 版本卡片 -->
        <div v-if="pushoverStore.isExtensionDownloaded" class="version-card">
          <h3>版本信息</h3>
          <div class="info-row">
            <span class="label">当前版本:</span>
            <span class="value">{{ pushoverStore.extensionInfo.current_version || pushoverStore.extensionInfo.version }}</span>
          </div>
          <div v-if="pushoverStore.isUpdateAvailable" class="info-row">
            <span class="label">最新版本:</span>
            <span class="value text-accent">{{ pushoverStore.extensionInfo.latest_version }}</span>
          </div>
          <div v-if="pushoverStore.isUpdateAvailable" class="update-hint">
            有新版本可用，建议更新扩展
          </div>
          <div v-if="!pushoverStore.isUpdateAvailable && pushoverStore.isExtensionDownloaded" class="latest-hint">
            ✅ 已是最新版本
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="actions-card">
          <h3>操作</h3>
          <div class="actions-grid">
            <button
              v-if="!pushoverStore.isExtensionDownloaded"
              class="btn btn-primary"
              :disabled="pushoverStore.loading"
              @click="handleDownload"
            >
              下载扩展
            </button>

            <button
              v-if="pushoverStore.isExtensionDownloaded"
              class="btn btn-secondary"
              :disabled="pushoverStore.loading"
              @click="handleCheckUpdate"
            >
              检查更新
            </button>

            <button
              v-if="pushoverStore.isExtensionDownloaded && pushoverStore.isUpdateAvailable"
              class="btn btn-primary"
              :disabled="pushoverStore.loading"
              @click="handleUpdate"
            >
              更新扩展
            </button>

            <button
              v-if="pushoverStore.isExtensionDownloaded"
              class="btn btn-danger"
              :disabled="pushoverStore.loading"
              @click="handleReclone"
            >
              重新下载
            </button>

            <button
              class="btn btn-secondary"
              @click="handleOpenGitHub"
            >
              GitHub 仓库
            </button>

            <button
              class="btn btn-secondary"
              @click="handleOpenConfigFolder"
            >
              打开配置文件夹
            </button>
          </div>
        </div>

        <!-- 错误信息 -->
        <div v-if="pushoverStore.error" class="error-message">
          {{ pushoverStore.error }}
        </div>

        <!-- Loading 状态 -->
        <div v-if="pushoverStore.loading" class="loading-overlay">
          <div class="loading-spinner"></div>
          <span>处理中...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { OpenConfigFolder } from '../../wailsjs/go/main/App'

interface Props {
  open: boolean
}

interface Emits {
  (e: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const pushoverStore = usePushoverStore()

// 状态样式类
const statusClass = {
  get: () => {
    if (!pushoverStore.isExtensionDownloaded) return 'text-error'
    if (pushoverStore.isUpdateAvailable) return 'text-warning'
    return 'text-success'
  }
}

// 监听 open 变化，打开时刷新状态
watch(() => props.open, async (newValue) => {
  if (newValue) {
    await pushoverStore.checkExtensionStatus()
  }
})

// 关闭对话框
function handleClose() {
  emit('close')
}

// 下载扩展
async function handleDownload() {
  try {
    await pushoverStore.cloneExtension()
  } catch (e) {
    // Error handled in store
  }
}

// 检查更新
async function handleCheckUpdate() {
  try {
    await pushoverStore.checkForExtensionUpdates()
    await pushoverStore.checkExtensionStatus()
  } catch (e) {
    // Error handled in store
  }
}

// 更新扩展
async function handleUpdate() {
  try {
    await pushoverStore.updateExtension()
  } catch (e) {
    // Error handled in store
  }
}

// 重新下载扩展
async function handleReclone() {
  if (!confirm('确定要重新下载扩展吗？这将覆盖现有文件。')) return
  try {
    await pushoverStore.recloneExtension()
  } catch (e) {
    // Error handled in store
  }
}

// 打开 GitHub 仓库
function handleOpenGitHub() {
  window.open('https://github.com/allanpk716/cc-pushover-hook', '_blank')
}

// 打开配置文件夹
async function handleOpenConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    console.error('打开配置文件夹失败:', message)
  }
}

// 打开扩展文件夹
async function handleOpenExtensionFolder() {
  try {
    const { OpenExtensionFolder } = await import('../../wailsjs/go/main/App')
    await OpenExtensionFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    pushoverStore.error = `打开扩展文件夹失败: ${message}`
  }
}
</script>

<style scoped>
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
  backdrop-filter: blur(4px);
}

.dialog {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  width: 600px;
  max-width: 90vw;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  border: 1px solid var(--border-default);
  position: relative;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg) var(--space-xl);
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-elevated);
}

.dialog-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-normal);
}

.close-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.dialog-body {
  padding: var(--space-xl);
  overflow-y: auto;
  flex: 1;
  position: relative;
}

.status-card,
.version-card,
.actions-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  margin-bottom: var(--space-md);
}

.status-card h3,
.version-card h3,
.actions-card h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
}

.info-row .label {
  color: var(--text-secondary);
  font-size: 14px;
}

.info-row .value {
  color: var(--text-primary);
  font-weight: 500;
}

.info-row .value-path {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
  max-width: 350px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-row .value-path.clickable {
  cursor: pointer;
  transition: all var(--transition-normal);
}

.info-row .value-path.clickable:hover {
  color: var(--accent-primary);
  text-decoration: underline;
}

.text-success {
  color: #22c55e;
}

.text-warning {
  color: #f59e0b;
}

.text-error {
  color: #ef4444;
}

.text-accent {
  color: var(--accent-primary);
}

.update-hint {
  margin-top: var(--space-sm);
  padding: var(--space-sm);
  background: rgba(6, 182, 212, 0.1);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: var(--radius-sm);
  color: var(--accent-primary);
  font-size: 13px;
  text-align: center;
}

.latest-hint {
  margin-top: var(--space-sm);
  padding: var(--space-sm);
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: var(--radius-sm);
  color: #22c55e;
  font-size: 13px;
  text-align: center;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 14px;
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
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.3);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-elevated);
  border-color: var(--accent-primary);
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.btn-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
  border-color: #ef4444;
}

.error-message {
  padding: var(--space-md);
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: var(--radius-sm);
  font-size: 14px;
  text-align: center;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-md);
  color: white;
  border-radius: var(--radius-lg);
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
