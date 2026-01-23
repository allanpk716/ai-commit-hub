<template>
  <div v-if="open" class="dialog-overlay" @click.self="close">
    <div class="dialog">
      <div class="dialog-header">
        <h2>设置</h2>
        <button class="close-btn" @click="close">×</button>
      </div>

      <div class="dialog-body">
        <!-- 配置管理 -->
        <section class="config-section">
          <h3>配置管理</h3>
          <button class="btn btn-secondary" @click="openConfigFolder">
            <span>📁</span>
            <span>打开配置文件夹</span>
          </button>
        </section>

        <!-- cc-pushover-hook 管理 -->
        <section class="pushover-section">
          <h3>cc-pushover-hook 管理</h3>

          <div v-if="pushoverStore.loading" class="loading-state">
            <span>⏳</span>
            <span>加载中...</span>
          </div>

          <div v-else class="extension-info">
            <div class="info-row">
              <span class="label">状态:</span>
              <span class="value" :class="{ 'text-success': pushoverStore.isExtensionDownloaded }">
                {{ pushoverStore.isExtensionDownloaded ? '✅ 已下载' : '❌ 未下载' }}
              </span>
            </div>

            <div v-if="pushoverStore.isExtensionDownloaded" class="info-row">
              <span class="label">版本:</span>
              <span class="value">v{{ pushoverStore.extensionInfo.version }}</span>
            </div>

            <div v-if="pushoverStore.isExtensionDownloaded && pushoverStore.isUpdateAvailable" class="info-row">
              <span class="label">最新版本:</span>
              <span class="value text-accent">v{{ pushoverStore.extensionInfo.latest_version }}</span>
            </div>

            <div class="info-row">
              <span class="label">位置:</span>
              <span class="value value-path">{{ pushoverStore.extensionInfo.path || '-' }}</span>
            </div>

            <div class="extension-actions">
              <button
                v-if="!pushoverStore.isExtensionDownloaded"
                class="btn btn-primary"
                :disabled="pushoverStore.loading"
                @click="handleClone"
              >
                <span>⬇️</span>
                <span>下载扩展</span>
              </button>

              <template v-else>
                <button
                  v-if="pushoverStore.isUpdateAvailable"
                  class="btn btn-primary"
                  :disabled="pushoverStore.loading"
                  @click="handleUpdate"
                >
                  <span>🔄</span>
                  <span>更新到最新版本</span>
                </button>

                <button
                  class="btn btn-secondary"
                  :disabled="pushoverStore.loading"
                  @click="handleCheckUpdate"
                >
                  <span>🔍</span>
                  <span>检查更新</span>
                </button>

                <button
                  class="btn btn-danger"
                  :disabled="pushoverStore.loading"
                  @click="handleReclone"
                >
                  <span>🔃</span>
                  <span>重新下载</span>
                </button>
              </template>
            </div>
          </div>

          <!-- 已安装的项目列表 -->
          <div v-if="pushoverStore.isExtensionDownloaded" class="installed-projects">
            <h4>已安装的项目</h4>
            <div v-if="installedProjects.length === 0" class="empty-state">
              暂无项目安装了 Pushover Hook
            </div>
            <div v-else class="project-list">
              <div
                v-for="project in installedProjects"
                :key="project.id"
                class="project-item"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="project-mode">{{ getModeLabel(project) }}</span>
              </div>
            </div>
          </div>
        </section>

        <!-- 错误信息 -->
        <div v-if="pushoverStore.error" class="error-message">
          {{ pushoverStore.error }}
        </div>
      </div>

      <div class="dialog-footer">
        <button class="btn btn-secondary" @click="close">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { useProjectStore } from '../stores/projectStore'
import { OpenConfigFolder } from '../../wailsjs/go/main/App'
import type { NotificationMode } from '../types/pushover'

interface Props {
  modelValue: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const pushoverStore = usePushoverStore()
const projectStore = useProjectStore()

const open = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 已安装 Hook 的项目列表
const installedProjects = computed(() => {
  return projectStore.projects.filter(p => p.hook_installed)
})

function getModeLabel(project: any): string {
  const modeLabels: Record<NotificationMode, string> = {
    enabled: '全部启用',
    pushover_only: '仅 Pushover',
    windows_only: '仅 Windows',
    disabled: '已禁用'
  }
  return modeLabels[project.notification_mode as NotificationMode] || '未知'
}

function close() {
  open.value = false
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    console.error('打开配置文件夹失败:', message)
  }
}

async function handleClone() {
  try {
    await pushoverStore.cloneExtension()
  } catch (e) {
    // Error handled in store
  }
}

async function handleUpdate() {
  try {
    await pushoverStore.updateExtension()
  } catch (e) {
    // Error handled in store
  }
}

async function handleCheckUpdate() {
  await pushoverStore.checkExtensionStatus()
}

async function handleReclone() {
  if (!confirm('确定要重新下载扩展吗？这将覆盖现有文件。')) return
  try {
    await pushoverStore.cloneExtension()
  } catch (e) {
    // Error handled in store
  }
}

onMounted(() => {
  if (open.value) {
    pushoverStore.checkExtensionStatus()
  }
})
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
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg) var(--space-xl);
  border-bottom: 1px solid var(--border-default);
}

.dialog-header h2 {
  margin: 0;
  font-size: 20px;
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
}

section {
  margin-bottom: var(--space-xl);
}

section h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.config-section button {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.pushover-section {
  border-top: 1px solid var(--border-default);
  padding-top: var(--space-lg);
}

.loading-state {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  color: var(--text-muted);
  padding: var(--space-lg);
  justify-content: center;
}

.extension-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
}

.text-success {
  color: #22c55e;
}

.text-accent {
  color: var(--accent-primary);
}

.extension-actions {
  display: flex;
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.installed-projects {
  margin-top: var(--space-lg);
}

.installed-projects h4 {
  margin: 0 0 var(--space-md) 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.empty-state {
  padding: var(--space-md);
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.project-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.project-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm);
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.project-name {
  color: var(--text-primary);
  font-weight: 500;
}

.project-mode {
  color: var(--text-muted);
  font-size: 13px;
}

.btn {
  display: flex;
  align-items: center;
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
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.btn-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.error-message {
  padding: var(--space-md);
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: var(--radius-sm);
  font-size: 14px;
}

.dialog-footer {
  padding: var(--space-lg) var(--space-xl);
  border-top: 1px solid var(--border-default);
  display: flex;
  justify-content: flex-end;
}
</style>
