<template>
  <!-- Loading screen -->
  <div v-if="initialLoading" class="startup-loading">
    <div class="spinner"></div>
    <p>加载项目状态...</p>
  </div>

  <!-- SplashScreen -->
  <SplashScreen v-else-if="showSplash" />

  <!-- Main App -->
  <div v-else class="app grid-pattern">
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

    <!-- Main content area -->
    <main class="content">
      <ProjectList
        :selected-id="selectedProjectId"
        @select="handleSelectProject"
      />
      <div class="commit-area">
        <transition name="fade-slide" mode="out-in">
          <CommitPanel v-if="selectedProjectId" :key="selectedProjectId" />
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
import { ref, onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { useCommitStore } from './stores/commitStore'
import { usePushoverStore } from './stores/pushoverStore'
import { SelectProjectFolder } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import ProjectList from './components/ProjectList.vue'
import CommitPanel from './components/CommitPanel.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import ExtensionStatusButton from './components/ExtensionStatusButton.vue'
import ExtensionInfoDialog from './components/ExtensionInfoDialog.vue'
import SplashScreen from './components/SplashScreen.vue'
import type { GitProject } from './types'

const projectStore = useProjectStore()
const commitStore = useCommitStore()
const pushoverStore = usePushoverStore()
const selectedProjectId = ref<number>()
const settingsOpen = ref(false)
const extensionDialogOpen = ref(false)
const initialLoading = ref(true)
const showSplash = ref(true)

onMounted(async () => {
  console.log('[App] onMounted 开始')
  try {
    // 初始化 StatusCache 并预加载
    console.log('[App] 开始初始化 StatusCache')
    const { useStatusCache } = await import('./stores/statusCache')
    const statusCache = useStatusCache()
    await statusCache.init()
    console.log('[App] StatusCache 初始化完成')
  } catch (error) {
    console.error('[App] StatusCache 初始化失败:', error)
  } finally {
    console.log('[App] 设置 initialLoading = false')
    initialLoading.value = false
  }

  // 延迟一点隐藏 SplashScreen，给用户更好的体验
  setTimeout(() => {
    console.log('[App] 隐藏 SplashScreen')
    showSplash.value = false
  }, 1000)

  // 监听启动完成事件（备用，防止其他模块依赖）
  EventsOn("startup-complete", () => {
    console.log('[App] startup-complete 事件触发')
    showSplash.value = false
  })

  // 检查 Pushover 配置
  await pushoverStore.checkPushoverConfig()
  if (!pushoverStore.configValid) {
    console.warn('Pushover 环境变量未配置，通知功能可能不可用')
  }

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
</script>

<style scoped>
/* Loading screen */
.startup-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: var(--bg-primary);
  position: relative;
  z-index: var(--z-elevated);
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid var(--border-default);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: var(--space-lg);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.startup-loading p {
  margin: 0;
  font-size: 16px;
  color: var(--text-secondary);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

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
</style>
