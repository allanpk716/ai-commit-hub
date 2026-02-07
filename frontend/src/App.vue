<template>
  <!-- SplashScreen (优先显示) -->
  <SplashScreen v-if="showSplash" />

  <!-- Main App (初始化完成后显示) -->
  <div v-show="!showSplash" class="app grid-pattern">
    <!-- 错误横幅 -->
    <transition name="slide-down">
      <div v-if="initErrors.length > 0" class="init-error-banner">
        <span class="icon">⚠️</span>
        <span class="message">部分功能加载失败，请稍后手动刷新</span>
        <button @click="initErrors = []" class="dismiss">×</button>
      </div>
    </transition>

    <!-- Animated background gradient -->
    <div class="bg-gradient"></div>

    <!-- Main toolbar -->
    <header class="toolbar">
      <div class="toolbar-left">
        <div class="logo">
          <img src="./assets/app-icon.png" alt="AI Commit Hub" class="logo-icon" />
          <h1>AI Commit Hub</h1>
        </div>
        <div class="toolbar-divider"></div>
      </div>

      <div class="toolbar-actions">
        <button @click="openAddProject" class="btn btn-primary">
          <span class="icon">＋</span>
          <span>添加项目</span>
        </button>
        <!-- 扩展状态按钮 -->
        <ExtensionStatusButton @open="extensionDialogOpen = true" />
        <button @click="openSettings" class="btn btn-secondary">
          <span class="icon">⚙</span>
          <span>设置</span>
        </button>
      </div>
    </header>

    <!-- Settings Dialog -->
    <SettingsDialog v-model="settingsOpen" />

    <!-- Extension Info Dialog -->
    <ExtensionInfoDialog :open="extensionDialogOpen" @close="extensionDialogOpen = false" />

    <!-- Error Toast (全局错误提示) -->
    <ErrorToast />

    <!-- Update Dialog (更新对话框) -->
    <UpdateDialog :visible="showUpdateDialog" @close="showUpdateDialog = false" />

    <!-- Update Progress Dialog (下载进度对话框) -->
    <UpdateProgressDialog :visible="updateStore.isDownloading" @close="updateStore.cancelDownload" />

    <!-- Update Installer Dialog (安装确认对话框) -->
    <UpdateInstallerDialog :visible="updateStore.showInstallConfirm" @close="updateStore.cancelInstall" />

    <!-- 删除确认对话框 -->
    <ConfirmDialog
      :visible="showDeleteDialog"
      :title="deleteDialogTitle"
      :message="deleteDialogMessage"
      :details="deleteDialogDetails"
      :note="deleteDialogNote"
      :confirm-text="deleteDialogConfirmText"
      :cancel-text="deleteDialogCancelText"
      :type="deleteDialogType"
      @confirm="handleDeleteConfirm"
      @cancel="showDeleteDialog = false"
    />

    <!-- Main content area -->
    <main class="content">
      <ProjectList
        :selected-id="selectedProjectId"
        @select="handleSelectProject"
        @show-delete-dialog="handleShowDeleteDialog"
      />
      <div class="commit-area">
        <transition name="fade-slide" mode="out-in">
          <CommitPanel
            v-if="selectedProjectId"
            :key="selectedProjectId"
            @show-delete-dialog="handleShowDeleteDialog"
          />
          <div v-else class="empty-state">
            <div class="empty-icon">📝</div>
            <h2>选择一个项目开始</h2>
            <p>从左侧列表选择一个 Git 项目来生成 AI 驱动的 commit 消息</p>
          </div>
        </transition>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { useCommitStore } from './stores/commitStore'
import { usePushoverStore } from './stores/pushoverStore'
import { useUpdateStore } from './stores/updateStore'
import { SelectProjectFolder } from '../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import ProjectList from './components/ProjectList.vue'
import CommitPanel from './components/CommitPanel.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import ExtensionStatusButton from './components/ExtensionStatusButton.vue'
import ExtensionInfoDialog from './components/ExtensionInfoDialog.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import SplashScreen from './components/SplashScreen.vue'
import ErrorToast from './components/ErrorToast.vue'
import UpdateDialog from './components/UpdateDialog.vue'
import UpdateProgressDialog from './components/UpdateProgressDialog.vue'
import UpdateInstallerDialog from './components/UpdateInstallerDialog.vue'
import type { GitProject } from './types'

const projectStore = useProjectStore()
const commitStore = useCommitStore()
const pushoverStore = usePushoverStore()
const updateStore = useUpdateStore()
const selectedProjectId = ref<number>()
const settingsOpen = ref(false)
const extensionDialogOpen = ref(false)
const showSplash = ref(true)
const initErrors = ref<Array<{ error: string; message: string }>>([])
const showUpdateDialog = ref(false)

// 删除对话框状态
const showDeleteDialog = ref(false)
const deleteDialogTitle = ref('')
const deleteDialogMessage = ref('')
const deleteDialogDetails = ref<Array<{label: string; value: string}>>([])
const deleteDialogNote = ref('')
const deleteDialogConfirmText = ref('删除')
const deleteDialogCancelText = ref('取消')
const deleteDialogType = ref<'warning' | 'danger'>('danger')
const deleteDialogCallback = ref<(() => Promise<void>) | null>(null)

async function initializeApp() {
  console.log('[App] 开始初始化前端应用')

  // 只执行不阻塞启动的基础初始化
  const tasks = [
    projectStore.loadProjects()
      .catch(err => ({ error: 'loadProjects', message: err?.message || '未知错误' })),
    pushoverStore.checkExtensionStatus()
      .catch(err => ({ error: 'extensionStatus', message: err?.message || '未知错误' })),
    pushoverStore.checkPushoverConfig()
      .catch(err => ({ error: 'pushoverConfig', message: err?.message || '未知错误' }))
  ]

  const results = await Promise.all(tasks)
  const errors = results.filter((r): r is { error: string; message: string } => r !== null && typeof r === 'object' && 'error' in r && 'message' in r)
  if (errors.length > 0) {
    console.warn('[App] 部分初始化任务失败:', errors)
    initErrors.value = errors
  }

  console.log('[App] 前端应用初始化完成')
}

onMounted(async () => {
  console.log('[App] onMounted 开始')

  // 1. 立即执行前端基础初始化
  await initializeApp()

  // 2. 监听后端启动完成事件
  EventsOn('startup-complete', async (data: { success?: boolean; statuses?: Record<string, any> } | null) => {
    console.log('[App] 收到 startup-complete 事件', { data })

    // 如果后端发送了预加载的状态数据，填充到 StatusCache
    if (data?.success && data?.statuses) {
      try {
        const { useStatusCache } = await import('./stores/statusCache')
        const statusCache = useStatusCache()

        // 将后端预加载的状态数据填充到缓存
        for (const [path, status] of Object.entries(data.statuses)) {
          statusCache.updateCache(path, {
            gitStatus: status.gitStatus,
            stagingStatus: status.stagingStatus,
            untrackedCount: status.untrackedCount,
            pushoverStatus: status.pushoverStatus,
            pushStatus: status.pushStatus,
            lastUpdated: new Date(status.lastUpdated).getTime(),
            loading: false,
            error: null,
            stale: false
          })
        }

        console.log('[App] StatusCache 已填充预加载数据', {
          count: Object.keys(data.statuses).length
        })
      } catch (error) {
        console.error('[App] 填充 StatusCache 失败:', error)
        // 失败不影响进入主界面，StatusCache 会按需加载
      }
    } else {
      console.log('[App] 后端未发送预加载数据，StatusCache 将按需加载')
    }

    // 隐藏 SplashScreen
    showSplash.value = false
  })

  // 监听窗口可见性事件 (系统托盘相关)
  EventsOn('window-shown', (data: { timestamp: string }) => {
    console.log('[App] 窗口已从托盘恢复', data.timestamp)
  })

  EventsOn('window-hidden', (data: { timestamp: string }) => {
    console.log('[App] 窗口已隐藏到托盘', data.timestamp)
  })

  // 监听更新可用事件
  EventsOn('update-available', (data: { hasUpdate: boolean; info: any }) => {
    console.log('[App] 检测到更新', data)
    if (data.hasUpdate) {
      showUpdateDialog.value = true
    }
  })

  // 监听托盘菜单的"检查更新"事件
  EventsOn('check-update-from-tray', async () => {
    console.log('[App] 从托盘触发检查更新')
    try {
      const info = await updateStore.checkForUpdates()
      if (info.hasUpdate) {
        showUpdateDialog.value = true
      } else {
        // 可以显示"已是最新版本"提示
        console.log('[App] 已是最新版本')
      }
    } catch (error) {
      console.error('[App] 检查更新失败:', error)
    }
  })

  // 3. 超时保护（30秒后强制进入主界面）
  const timeoutId = setTimeout(() => {
    if (showSplash.value) {
      console.warn('[App] 启动超时（30秒），强制进入主界面')
      showSplash.value = false
    }
  }, 30000)

  // 4. 组件卸载时清理
  onUnmounted(() => {
    EventsOff('startup-complete')
    EventsOff('window-shown')
    EventsOff('window-hidden')
    EventsOff('update-available')
    EventsOff('check-update-from-tray')
    clearTimeout(timeoutId)
  })

  console.log('[App] onMounted 完成')
})

async function openAddProject() {
  try {
    const path = await SelectProjectFolder()
    if (path) {
      await projectStore.addProject(path)
      alert('项目添加成功!')
    }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    alert('添加项目失败: ' + message)
  }
}

function handleSelectProject(project: GitProject) {
  selectedProjectId.value = project.id
  projectStore.selectProject(project.path)  // 同步到 projectStore
  commitStore.loadProjectStatus(project.path)
}

function openSettings() {
  settingsOpen.value = true
}

// 处理删除对话框显示请求
function handleShowDeleteDialog(config: {
  title: string
  message: string
  details: Array<{label: string; value: string}>
  note?: string
  confirmText: string
  cancelText: string
  type: 'warning' | 'danger'
  onConfirm: () => Promise<void>
}) {
  openDeleteDialog(config)
}

function openDeleteDialog(config: {
  title: string
  message: string
  details: Array<{label: string; value: string}>
  note?: string
  confirmText: string
  cancelText: string
  type: 'warning' | 'danger'
  onConfirm: () => Promise<void>
}) {
  deleteDialogTitle.value = config.title
  deleteDialogMessage.value = config.message
  deleteDialogDetails.value = config.details
  deleteDialogNote.value = config.note || ''
  deleteDialogConfirmText.value = config.confirmText
  deleteDialogCancelText.value = config.cancelText
  deleteDialogType.value = config.type
  deleteDialogCallback.value = config.onConfirm
  showDeleteDialog.value = true
}

async function handleDeleteConfirm() {
  if (deleteDialogCallback.value) {
    try {
      await deleteDialogCallback.value()
      showDeleteDialog.value = false
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '操作失败'
      console.error('操作失败:', message)
    }
  }
}
</script>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  position: relative;
  background: var(--bg-primary);
  color: var(--text-primary);
}

/* Animated background gradient */
.bg-gradient {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%, rgba(6, 182, 212, 0.15), transparent),
    radial-gradient(ellipse 60% 40% at 100% 100%, rgba(139, 92, 246, 0.1), transparent);
  pointer-events: none;
  z-index: 0;
}

/* Toolbar */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg) var(--space-xl);
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--glass-border);
  position: relative;
  z-index: var(--z-elevated);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.logo-icon {
  width: 36px;
  height: 36px;
  object-fit: contain;
  border-radius: var(--radius-md);
  box-shadow: var(--glow-primary);
  animation: pulse-glow 3s ease-in-out infinite;
}

.logo h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--text-primary), var(--accent-primary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: -0.5px;
}

.toolbar-divider {
  width: 1px;
  height: 24px;
  background: var(--border-default);
}

.toolbar-actions {
  display: flex;
  gap: var(--space-md);
}

/* Buttons */
.btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-lg);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.btn .icon {
  font-size: 16px;
  line-height: 1;
}

.btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, transparent, rgba(255,255,255,0.05), transparent);
  transform: translateX(-100%);
  transition: transform var(--transition-slow);
}

.btn:hover::before {
  transform: translateX(100%);
}

.btn-primary {
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  color: white;
  border-color: transparent;
  box-shadow: var(--glow-primary);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 0 30px rgba(6, 182, 212, 0.5);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.btn-secondary:hover {
  background: var(--bg-elevated);
  border-color: var(--border-hover);
}

/* Content area */
.content {
  display: flex;
  gap: var(--space-md);
  padding: var(--space-md);
  height: calc(100vh - 73px);
  position: relative;
  z-index: var(--z-base);
}

.commit-area {
  flex: 1;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  overflow-y: auto;
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: var(--space-2xl);
  text-align: center;
  animation: fade-in 0.5s ease-out;
}

.empty-icon {
  width: 80px;
  height: 80px;
  margin-bottom: var(--space-lg);
  font-size: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.5;
}

.empty-state h2 {
  margin: 0 0 var(--space-sm) 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-state p {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
  max-width: 400px;
  line-height: 1.6;
}

/* Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all var(--transition-normal);
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* Error banner */
.init-error-banner {
  position: fixed;
  top: var(--space-lg);
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-md);
  z-index: var(--z-modal);
  animation: slide-down 0.3s ease-out;
}

.init-error-banner .icon {
  font-size: 18px;
}

.init-error-banner .message {
  font-size: 13px;
  color: var(--accent-warning);
}

.init-error-banner .dismiss {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 18px;
  padding: 0 4px;
}

.init-error-banner .dismiss:hover {
  color: var(--text-primary);
}

@keyframes slide-down {
  from {
    opacity: 0;
    transform: translate(-50%, -20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}
</style>
