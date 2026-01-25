<template>
  <div class="pushover-management-panel">
    <div class="panel-header">
      <h3>cc-pushover-hook 管理</h3>
      <p class="panel-description">
        管理所有项目的 Pushover 通知 Hook 扩展
      </p>
    </div>

    <!-- 扩展状态卡片 -->
    <div class="extension-card">
      <div v-if="pushoverStore.loading" class="loading-state">
        <span>⏳</span>
        <span>加载中...</span>
      </div>

      <div v-else>
        <div class="extension-status">
          <div class="status-icon">
            {{ pushoverStore.isExtensionDownloaded ? '✅' : '❌' }}
          </div>
          <div class="status-info">
            <div class="status-title">
              {{ pushoverStore.isExtensionDownloaded ? '扩展已下载' : '扩展未下载' }}
            </div>
            <div v-if="pushoverStore.isExtensionDownloaded" class="status-details">
              <div class="version-info">
                <span v-if="pushoverStore.extensionInfo.current_version" class="version-current">
                  当前版本: v{{ pushoverStore.extensionInfo.current_version }}
                </span>
                <span v-if="pushoverStore.extensionInfo.latest_version" class="version-latest">
                  最新版本: v{{ pushoverStore.extensionInfo.latest_version }}
                </span>
              </div>
              <span v-if="pushoverStore.isUpdateAvailable" class="update-badge">
                有新版本
              </span>
            </div>
          </div>
        </div>

        <div class="extension-actions">
          <button
            v-if="!pushoverStore.isExtensionDownloaded"
            class="btn btn-primary"
            :disabled="pushoverStore.loading"
            @click="handleClone"
          >
            ⬇️ 下载扩展
          </button>

          <template v-else>
            <button
              v-if="pushoverStore.isUpdateAvailable"
              class="btn btn-primary"
              :disabled="pushoverStore.loading"
              @click="handleUpdateExtension"
            >
              🔄 更新扩展
            </button>

            <button
              class="btn btn-secondary"
              :disabled="pushoverStore.loading"
              @click="handleCheckUpdate"
            >
              🔍 检查更新
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 项目列表 -->
    <div class="projects-section">
      <div class="section-header">
        <h4>已安装的项目 ({{ installedCount }})</h4>
        <button
          v-if="installedCount > 0"
          class="btn-link"
          @click="showAll = !showAll"
        >
          {{ showAll ? '收起' : '展开全部' }}
        </button>
      </div>

      <div v-if="installedCount === 0" class="empty-state">
        <span>📭</span>
        <p>暂无项目安装了 Pushover Hook</p>
      </div>

      <div v-else class="project-list">
        <div
          v-for="project in displayedProjects"
          :key="project.id"
          class="project-card"
        >
          <div class="project-info">
            <span class="project-name">{{ project.name }}</span>
            <span class="project-mode" :class="`mode-${project.notification_mode}`">
              {{ getModeLabel(project.notification_mode) }}
            </span>
          </div>
          <div class="project-actions">
            <button
              class="btn-icon"
              title="更改通知模式"
              @click="openModeSelector(project)"
            >
              ⚙️
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="pushoverStore.error" class="error-alert">
      <span>⚠️</span>
      <span>{{ pushoverStore.error }}</span>
      <button class="alert-close" @click="pushoverStore.error = null">×</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { useProjectStore } from '../stores/projectStore'
import { NOTIFICATION_MODES } from '../types/pushover'
import type { GitProject } from '../types'

const pushoverStore = usePushoverStore()
const projectStore = useProjectStore()

const showAll = ref(false)

// 已安装的项目列表
const installedProjects = computed(() => {
  return projectStore.projects.filter(p => p.hook_installed)
})

// 显示的项目数量
const installedCount = computed(() => installedProjects.value.length)

// 显示的项目列表
const displayedProjects = computed(() => {
  if (showAll.value) {
    return installedProjects.value
  }
  return installedProjects.value.slice(0, 3)
})

function getModeLabel(mode: string | undefined): string {
  if (!mode) return '未知'
  const modeConfig = NOTIFICATION_MODES.find(m => m.value === mode)
  return modeConfig?.label || '未知'
}

async function handleClone() {
  try {
    await pushoverStore.cloneExtension()
  } catch (e) {
    // Error handled in store
  }
}

async function handleUpdateExtension() {
  try {
    await pushoverStore.updateExtension()
    // 更新后重新检查版本信息
    await handleCheckUpdate()
  } catch (e) {
    // Error handled in store
  }
}

async function handleCheckUpdate() {
  try {
    await pushoverStore.checkExtensionStatus()
  } catch (e) {
    // Error handled in store
  }
}

function openModeSelector(project: GitProject) {
  // TODO: 打开模式选择器
  console.log('Open mode selector for project:', project.name)
}

onMounted(() => {
  pushoverStore.checkExtensionStatus()
})
</script>

<style scoped>
.pushover-management-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.panel-header {
  margin-bottom: var(--space-sm);
}

.panel-header h3 {
  margin: 0 0 var(--space-xs) 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.panel-description {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
}

.extension-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-md);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  color: var(--text-muted);
  padding: var(--space-lg);
}

.extension-status {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-md);
}

.status-icon {
  font-size: 32px;
}

.status-info {
  flex: 1;
}

.status-title {
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-xs);
}

.status-details {
  font-size: 14px;
  color: var(--text-secondary);
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.version-current {
  color: var(--text-primary);
  font-weight: 500;
}

.version-latest {
  color: var(--text-secondary);
  font-size: 13px;
}

.update-badge {
  display: inline-block;
  margin-left: var(--space-sm);
  padding: 2px var(--space-sm);
  background: var(--accent-primary);
  color: white;
  font-size: 11px;
  border-radius: 10px;
  font-weight: 600;
}

.extension-actions {
  display: flex;
  gap: var(--space-sm);
}

.projects-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-md);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
}

.section-header h4 {
  margin: 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.btn-link {
  background: none;
  border: none;
  color: var(--accent-primary);
  cursor: pointer;
  font-size: 13px;
  padding: 0;
}

.btn-link:hover {
  text-decoration: underline;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xl);
  color: var(--text-muted);
  text-align: center;
}

.empty-state span {
  font-size: 32px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
}

.project-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.project-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm);
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
}

.project-info {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex: 1;
}

.project-name {
  color: var(--text-primary);
  font-weight: 500;
}

.project-mode {
  font-size: 12px;
  padding: 2px var(--space-sm);
  border-radius: 10px;
  font-weight: 600;
}

.project-mode.mode-enabled {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.project-mode.mode-pushover_only {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.project-mode.mode-windows_only {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}

.project-mode.mode-disabled {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.project-actions {
  display: flex;
  gap: var(--space-xs);
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: var(--space-xs);
  opacity: 0.7;
  transition: opacity var(--transition-normal);
}

.btn-icon:hover {
  opacity: 1;
}

.btn {
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

.error-alert {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: var(--radius-sm);
  font-size: 14px;
}

.alert-close {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  margin-left: auto;
  font-size: 18px;
}
</style>
