<template>
  <div class="commit-panel">
    <!-- Project Info Section -->
    <section class="panel-section staging-section" v-if="commitStore.projectStatus">
      <!-- Project Status Header -->
      <ProjectStatusHeader
        :branch="commitStore.projectStatus.branch"
        :project-path="currentProject?.path"
        :pushover-status="pushoverStatus"
        :pushover-loading="pushoverStore.loading"
        :available-terminals="availableTerminals"
        :preferred-terminal="preferredTerminal"
        @open-in-explorer="openInExplorer"
        @open-in-terminal="openInTerminal"
        @open-in-terminal-directly="openInTerminalDirectly"
        @refresh="handleRefresh"
        @install-pushover="handleInstallPushover"
        @update-pushover="handleUpdatePushover"
      />

      <!-- Staging Area -->
      <StagingArea v-if="commitStore.stagingStatus" />
    </section>

    <!-- Empty State -->
    <div class="empty-state-full" v-else>
      <div class="empty-illustration">📁</div>
      <h2>未选择项目</h2>
      <p>请从左侧列表选择一个项目</p>
    </div>

    <!-- Generated Message -->
    <section class="panel-section result-section" v-if="commitStore.projectStatus">
      <div class="result-header">
        <!-- 左侧：生成按钮（原"生成结果"标题位置） -->
        <div class="header-left">
          <button
            @click="handleGenerate"
            :disabled="!commitStore.hasStagedFiles || commitStore.isGenerating"
            class="btn-generate-main"
            :class="{ generating: commitStore.isGenerating }"
            title="生成 Commit 消息"
          >
            <span class="btn-icon">✨</span>
            <span class="btn-text" v-if="!commitStore.isGenerating">生成消息</span>
            <span class="btn-text" v-else>生成中...</span>
          </button>

          <!-- 提交和推送按钮（始终显示，禁用状态取决于是否有消息） -->
          <button
            @click="handleCommit"
            class="btn-action-inline btn-primary-inline"
            :disabled="!commitStore.hasStagedFiles || !(commitStore.streamingMessage || commitStore.generatedMessage)"
            title="提交到本地"
          >
            <span class="icon">✓</span>
            提交
          </button>
          <button
            @click="handlePush"
            class="btn-action-inline btn-push-inline"
            :disabled="isPushing || !pushStatus?.canPush"
            :title="pushStatus?.aheadCount ? `领先 ${pushStatus.aheadCount} 个提交` : pushStatus?.error || '无待推送内容'"
          >
            <span class="icon" :class="{ spin: isPushing }">↑</span>
            {{ isPushing ? '推送中' : '推送' }}
          </button>
        </div>

        <!-- 中间：配置控件 -->
        <div class="header-center">
          <div class="config-select-wrapper">
            <span class="config-label">🌐</span>
            <select v-model="commitStore.provider" class="config-select-inline" @change="handleConfigChange"
              :disabled="commitStore.isSavingConfig || commitStore.isGenerating">
              <option v-for="p in commitStore.availableProviders" :key="p.name" :value="p.name" :disabled="!p.configured">
                {{ getProviderDisplayName(p.name) }}
                <template v-if="!p.configured"> (未配置: {{ p.reason }})</template>
              </option>
            </select>
          </div>
          <div class="config-select-wrapper">
            <span class="config-label">🌍</span>
            <select v-model="commitStore.language" class="config-select-inline" @change="handleConfigChange"
              :disabled="commitStore.isSavingConfig || commitStore.isGenerating">
              <option value="zh">中文</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>

        <!-- 右侧：工具按钮（仅保留自定义标记和清除按钮） -->
        <div class="header-right">
          <span v-if="!commitStore.isDefaultConfig" class="config-badge-inline" @click="handleResetToDefault" title="重置为默认配置">自定义</span>
          <button v-if="commitStore.streamingMessage || commitStore.generatedMessage" @click="commitStore.clearMessage"
            class="icon-btn-small" title="清除" :disabled="commitStore.isGenerating">×</button>
        </div>
      </div>

      <!-- 配置不一致警告 -->
      <div v-if="commitStore.configValidation && !commitStore.configValidation.valid" class="config-warning-inline">
        <span class="icon">⚠️</span>
        <span>配置已过时：{{ formatResetFields(commitStore.configValidation.resetFields) }}</span>
        <button @click="handleConfirmReset" class="btn-confirm-reset">确认重置</button>
      </div>

      <div class="message-container">
        <!-- Streaming indicator (shown above content when generating) -->
        <div v-if="commitStore.isGenerating && !commitStore.streamingMessage" class="streaming-indicator">
          <span class="streaming-dot"></span>
          <span class="streaming-dot"></span>
          <span class="streaming-dot"></span>
        </div>

        <!-- Placeholder when no message -->
        <div v-if="!commitStore.streamingMessage && !commitStore.generatedMessage" class="message-hint-inline">
          <span class="hint-icon">⏳</span>
          <span>等待生成... 配置 AI 设置后点击下方按钮生成</span>
        </div>

        <!-- Message content (always shown when available) -->
        <pre v-if="commitStore.streamingMessage || commitStore.generatedMessage" class="message-content"
          :class="{ 'generating': commitStore.isGenerating }">{{ commitStore.streamingMessage || commitStore.generatedMessage }}
        </pre>
      </div>

      <div class="action-buttons-helper" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
        <button @click="handleCopy" class="btn-action btn-secondary">
          <span class="icon">📋</span>
          复制
        </button>
        <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-action btn-tertiary">
          <span class="icon">🔄</span>
          重新生成
        </button>
      </div>
    </section>

    <!-- Error Alert -->
    <transition name="slide-down">
      <div class="error-banner" v-if="commitStore.error">
        <span class="icon">⚠️</span>
        <span class="error-message">{{ commitStore.error }}</span>
        <button @click="commitStore.error = null" class="error-dismiss">×</button>
      </div>
    </transition>

    <!-- Toast Notification -->
    <transition name="toast-fade">
      <div v-if="toast.show" :class="['toast-notification', toast.type]">
        <span class="toast-icon">{{ toast.type === 'success' ? '✓' : '✕' }}</span>
        <span class="toast-message">{{ toast.message }}</span>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  CommitLocally,
  PushToRemote,
  GetAvailableTerminals,
  OpenInFileExplorer,
  OpenInTerminal,
  SaveCommitHistory
} from '../../wailsjs/go/main/App'
import { EventsEmit, EventsOff, EventsOn } from '../../wailsjs/runtime/runtime'
import { useCommitStore } from '../stores/commitStore'
import { useProjectStore } from '../stores/projectStore'
import { usePushoverStore } from '../stores/pushoverStore'
import { useStatusCache } from '../stores/statusCache'
import { useErrorStore } from '../stores/errorStore'
import ProjectStatusHeader from './ProjectStatusHeader.vue'
import StagingArea from './StagingArea.vue'

// 用户偏好存储键
const PREFERRED_TERMINAL_KEY = 'ai-commit-hub:preferred-terminal'

// 下拉菜单状态
const availableTerminals = ref<Array<{ id: string; name: string; icon: string }>>([])
const preferredTerminal = ref<string>('')

const commitStore = useCommitStore()
const projectStore = useProjectStore()
const pushoverStore = usePushoverStore()
const statusCache = useStatusCache()
const errorStore = useErrorStore()
const isPushing = ref(false) // 是否正在推送

// Toast 通知状态
const toast = ref<{
  show: boolean
  type: 'success' | 'error'
  message: string
}>({
  show: false,
  type: 'success',
  message: ''
})

// 当前选中项目的路径
const currentProjectPath = computed(() => commitStore.selectedProjectPath)
// 当前选中项目
const currentProject = computed(() =>
  projectStore.projects.find(p => p.path === currentProjectPath.value)
)
// Pushover Hook 状态 - 从 StatusCache 获取
const pushoverStatus = computed(() => {
  if (currentProjectPath.value) {
    const cached = statusCache.getStatus(currentProjectPath.value)
    return cached?.pushoverStatus || null
  }
  return null
})
// 推送状态 - 从 StatusCache 获取
const pushStatus = computed(() => {
  if (currentProjectPath.value) {
    return statusCache.getPushStatus(currentProjectPath.value)
  }
  return null
})

// Provider 显示名称映射
const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  ollama: 'Ollama',
  google: 'Google',
  openrouter: 'OpenRouter',
  phind: 'Phind'
}

// 获取 provider 显示名称
function getProviderDisplayName(name: string): string {
  return PROVIDER_DISPLAY_NAMES[name] || name
}

// 从缓存更新 UI 状态
function updateUIFromCache(cached: any) {
  if (cached.gitStatus) {
    commitStore.projectStatus = cached.gitStatus
  }
  if (cached.stagingStatus) {
    commitStore.stagingStatus = cached.stagingStatus
  }
  // - 移除 Pushover 状态同步
}

// 监听选中的项目变化
watch(() => projectStore.selectedProject, async (project) => {
  if (project) {
    commitStore.clearMessage()
    await commitStore.loadProjectAIConfig(project.id)

    // 策略C：优先显示缓存，过期时等待刷新
    const cached = statusCache.getStatus(project.path)

    if (cached && !cached.loading && !statusCache.isExpired(project.path)) {
      updateUIFromCache(cached)
    } else {
      await statusCache.refresh(project.path, { force: true, silent: true })
      const fresh = statusCache.getStatus(project.path)
      if (fresh) {
        updateUIFromCache(fresh)
      }
    }
  } else {
    commitStore.clearStagingState()
  }
}, { immediate: true })

// 配置变更时立即保存
async function handleConfigChange() {
  if (commitStore.selectedProjectId) {
    commitStore.isDefaultConfig = false
    await commitStore.saveProjectConfig(commitStore.selectedProjectId)
  }
}

// 重置为默认配置
async function handleResetToDefault() {
  if (confirm('确定要重置为默认配置吗？')) {
    commitStore.isDefaultConfig = true
    await commitStore.saveProjectConfig(commitStore.selectedProjectId)
    // 重新加载配置
    await commitStore.loadProjectAIConfig(commitStore.selectedProjectId)
  }
}

// 确认重置过时的配置
async function handleConfirmReset() {
  if (commitStore.selectedProjectId) {
    await commitStore.confirmResetConfig(commitStore.selectedProjectId)
  }
}

function formatResetFields(fields: string[]): string {
  const fieldNames: Record<string, string> = {
    provider: '服务商',
    language: '语言'
  }
  return fields.map(f => fieldNames[f] || f).join('、')
}

// 显示 Toast 通知
function showToast(type: 'success' | 'error', message: string) {
  toast.value = { show: true, type, message }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

async function handleGenerate() {
  await commitStore.generateCommit()
}

async function handleCopy() {
  const text = commitStore.streamingMessage || commitStore.generatedMessage
  await navigator.clipboard.writeText(text)
  showToast('success', '已复制到剪贴板')
}

async function handleCommit() {
  if (!commitStore.selectedProjectPath) {
    errorStore.addError('请先选择项目', '', 'error', 'CommitPanel')
    return
  }

  const message = commitStore.streamingMessage || commitStore.generatedMessage
  if (!message) {
    errorStore.addError('请先生成 commit 消息', '需要先生成 commit 消息才能提交', 'error', 'CommitPanel')
    return
  }

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)

    // 保存历史记录到数据库（后台功能，不显示在UI中）
    const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
    if (project) {
      await SaveCommitHistory(project.id, message, commitStore.provider, commitStore.language)
    }

    showToast('success', '提交成功!')

    // 使用 StatusCache 刷新状态
    await statusCache.refresh(commitStore.selectedProjectPath, { force: true, silent: true })
    const fresh = statusCache.getStatus(commitStore.selectedProjectPath)
    if (fresh) {
      updateUIFromCache(fresh)
    }

    commitStore.clearMessage()

    // 通知项目列表状态已更新
    EventsEmit('project-status-changed', {
      projectPath: commitStore.selectedProjectPath,
      changeType: 'commit'
    })
  } catch (e: unknown) {
    let errMessage = '提交失败'
    if (e instanceof Error) {
      errMessage = e.message
    } else if (typeof e === 'string') {
      errMessage = e
    } else {
      errMessage = JSON.stringify(e)
    }
    console.error('提交失败详细错误:', e)
    errorStore.addError('提交失败: ' + errMessage, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  }
}

async function handlePush() {
  if (!commitStore.selectedProjectPath) {
    errorStore.addError('请先选择项目', '', 'error', 'CommitPanel')
    return
  }

  isPushing.value = true
  try {
    await PushToRemote(commitStore.selectedProjectPath)
    showToast('success', '推送成功!')

    // 使用 StatusCache 刷新状态
    await statusCache.refresh(commitStore.selectedProjectPath, { force: true, silent: true })
    const fresh = statusCache.getStatus(commitStore.selectedProjectPath)
    if (fresh) {
      updateUIFromCache(fresh)
    }

    // 通知项目列表状态已更新
    EventsEmit('project-status-changed', {
      projectPath: commitStore.selectedProjectPath,
      changeType: 'push'
    })
  } catch (e) {
    let errMessage = '推送失败'
    if (e instanceof Error) {
      errMessage = e.message
    } else if (typeof e === 'string') {
      errMessage = e
    } else {
      errMessage = JSON.stringify(e)
    }
    console.error('推送失败详细错误:', e)
    errorStore.addError('推送失败: ' + errMessage, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  } finally {
    isPushing.value = false
  }
}

async function handleRegenerate() {
  commitStore.clearMessage()
  await commitStore.generateCommit()
}

// 处理安装 Pushover Hook
async function handleInstallPushover() {
  if (!currentProject.value) return
  const result = await pushoverStore.installHook(currentProject.value.path, false)
  if (result.success) {
    // 使用 StatusCache 刷新状态
    await statusCache.refresh(currentProject.value.path, { force: true, silent: true })
    const fresh = statusCache.getStatus(currentProject.value.path)
    if (fresh) {
      updateUIFromCache(fresh)
    }
    // 通知项目列表状态已更新（Hook 状态变化会影响 pushover_needs_update）
    EventsEmit('project-status-changed', {
      projectPath: currentProject.value.path,
      changeType: 'pushover'
    })
  } else {
    alert('安装失败: ' + (result.message || '未知错误'))
  }
}

// 处理更新 Pushover Hook
async function handleUpdatePushover() {
  if (!currentProject.value) return
  const result = await pushoverStore.updateHook(currentProject.value.path)
  if (result.success) {
    // 使用 StatusCache 刷新状态
    await statusCache.refresh(currentProject.value.path, { force: true, silent: true })
    const fresh = statusCache.getStatus(currentProject.value.path)
    if (fresh) {
      updateUIFromCache(fresh)
    }
    // 通知项目列表状态已更新（Hook 状态变化会影响 pushover_needs_update）
    EventsEmit('project-status-changed', {
      projectPath: currentProject.value.path,
      changeType: 'pushover'
    })
  } else {
    alert('更新失败: ' + (result.message || '未知错误'))
  }
}

// 加载用户偏好的终端类型
function loadPreferredTerminal(): string {
  const stored = localStorage.getItem(PREFERRED_TERMINAL_KEY)
  return stored || 'powershell' // 默认 PowerShell
}

// 保存用户偏好的终端类型
function savePreferredTerminal(terminalId: string) {
  localStorage.setItem(PREFERRED_TERMINAL_KEY, terminalId)
  preferredTerminal.value = terminalId
}

// 在文件管理器中打开
async function openInExplorer() {
  if (!currentProjectPath.value) return

  try {
    await OpenInFileExplorer(currentProjectPath.value)
    showToast('success', '已在文件管理器中打开')
  } catch (e) {
    const message = e instanceof Error ? e.message : '打开失败'
    errorStore.addError('打开终端失败: ' + message, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  }
}

// 直接打开用户偏好的终端
async function openInTerminalDirectly() {
  if (!currentProjectPath.value) return

  const terminalId = preferredTerminal.value || 'powershell'

  try {
    await OpenInTerminal(currentProjectPath.value, terminalId)
    showToast('success', '已在终端中打开')
  } catch (e) {
    const message = e instanceof Error ? e.message : '打开失败'
    errorStore.addError('打开终端失败: ' + message, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  }
}

// 在终端中打开（从菜单选择）
async function openInTerminal(terminalId: string) {
  if (!currentProjectPath.value) return

  try {
    await OpenInTerminal(currentProjectPath.value, terminalId)
    // 保存用户偏好
    savePreferredTerminal(terminalId)
    showToast('success', '已在终端中打开')
  } catch (e) {
    const message = e instanceof Error ? e.message : '打开失败'
    errorStore.addError('操作失败: ' + message, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  }
}

// 手动刷新项目状态
async function handleRefresh() {
  if (!currentProjectPath.value) return

  try {
    // 使用 StatusCache 强制刷新
    await statusCache.refresh(currentProjectPath.value, { force: true })

    // 从缓存更新 UI
    const fresh = statusCache.getStatus(currentProjectPath.value)
    if (fresh) {
      updateUIFromCache(fresh)
    }

    showToast('success', '已刷新')
  } catch (e) {
    const message = e instanceof Error ? e.message : '刷新失败'
    errorStore.addError('刷新项目状态失败: ' + message, e instanceof Error ? e.stack : '', 'error', 'CommitPanel')
  }
}

// 组件挂载时加载 provider 列表并注册事件监听
onMounted(async () => {
  commitStore.loadAvailableProviders()

  // 加载可用终端列表
  try {
    const terminals = await GetAvailableTerminals()
    availableTerminals.value = terminals
    preferredTerminal.value = loadPreferredTerminal()
  } catch (e) {
    console.error('Failed to load terminals:', e)
  }

  // 注册 Wails 事件监听器
  EventsOn('commit-delta', (delta: string) => {
    commitStore.handleDelta(delta)
  })

  EventsOn('commit-complete', (message: string) => {
    commitStore.handleComplete(message)
  })

  EventsOn('commit-error', (err: string) => {
    commitStore.handleError(err)
  })
})

// 组件卸载时清理事件监听器
onUnmounted(() => {
  // 清理 Wails 事件监听器
  EventsOff('commit-delta')
  EventsOff('commit-complete')
  EventsOff('commit-error')
})
</script>

<style scoped>
.commit-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: var(--space-lg);
  gap: var(--space-md);
}

/* Panel sections */
.panel-section {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  animation: fade-in 0.3s ease-out;
}

/* Staging section: 占据剩余空间 */
.staging-section {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  gap: var(--space-md); /* 添加间距 */
}

/* 确保 StagingArea 占据剩余空间 */
:deep(.staging-section .staging-area) {
  flex: 1;
  min-height: 0;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid var(--border-default);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.section-title .icon {
  font-size: 16px;
  line-height: 1;
}

.section-title h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.icon-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  line-height: 1;
}

.icon-btn:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

/* Branch badge */
.branch-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: rgba(139, 92, 246, 0.15);
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: 12px;
  font-size: 12px;
  font-family: var(--font-mono);
  font-weight: 500;
  color: var(--accent-secondary);
}

.branch-badge .icon {
  font-size: 10px;
  line-height: 1;
}

.header-badges {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.branch-badge {
  flex-shrink: 0;
}

/* 操作按钮组 */
.action-buttons-inline {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: var(--space-md);
}

/* 终端按钮组合 */
.terminal-button-wrapper {
  display: flex;
  position: relative;
}

.terminal-btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  padding-right: 6px;
}

.terminal-btn-main:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.terminal-btn-dropdown {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 6px;
  padding-right: 6px;
  font-size: 12px;
}

.terminal-btn-dropdown:hover {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
}

.dropdown-arrow {
  font-size: 10px;
  line-height: 1;
}

/* 终端菜单 */
.terminal-menu {
  right: 0;
  top: calc(100% + 4px);
  min-width: 180px;
}

/* 下拉菜单 */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  z-index: 1000;
  min-width: 200px;
  max-width: 280px;
  padding: var(--space-xs) 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
}

.menu-item:hover {
  background: var(--bg-elevated);
}

.menu-icon {
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}

.menu-divider {
  height: 1px;
  background: var(--border-default);
  margin: var(--space-xs) 0;
}

.menu-header {
  padding: var(--space-xs) var(--space-md);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.check-mark {
  margin-left: auto;
  color: var(--accent-primary);
  font-size: 12px;
  font-weight: 600;
}

/* 内联提示 */
.message-hint-inline {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-muted);
}

.hint-icon {
  font-size: 14px;
  opacity: 0.7;
  flex-shrink: 0;
}

/* Settings */
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.setting-label {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.setting-label .icon {
  font-size: 14px;
  line-height: 1;
}

.setting-select {
  padding: var(--space-sm);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.setting-select:hover {
  border-color: var(--border-hover);
}

.setting-select:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.setting-select option {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.setting-select option:disabled {
  color: var(--text-muted);
  background: var(--bg-tertiary);
  opacity: 0.6;
}

/* Generate button */
.btn-generate {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  padding: var(--space-md);
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-normal);
  box-shadow: var(--glow-primary);
}

.btn-generate .icon {
  font-size: 18px;
  line-height: 1;
}

.btn-generate:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 0 30px rgba(6, 182, 212, 0.5);
}

.btn-generate:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-generate.generating {
  background: linear-gradient(135deg, var(--accent-secondary), var(--accent-primary));
}

.btn-generate .icon.spin {
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

/* Result section */
.result-section {
  border-color: rgba(6, 182, 212, 0.3);
  background: rgba(6, 182, 212, 0.05);
}

.message-container {
  position: relative;
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-sm) var(--space-xs);
  max-height: 250px;
  overflow-y: auto;
  margin-bottom: var(--space-md);
}

.streaming-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: var(--space-sm) 0;
  margin-bottom: var(--space-sm);
}

.streaming-dot {
  width: 6px;
  height: 6px;
  background: var(--accent-primary);
  border-radius: 50%;
  animation: pulse 1.4s ease-in-out infinite;
}

.streaming-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.streaming-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes pulse {

  0%,
  60%,
  100% {
    transform: scale(0.8);
    opacity: 0.4;
  }

  30% {
    transform: scale(1);
    opacity: 1;
  }
}

.message-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
}

.message-content.generating {
  position: relative;
}

.message-content.generating::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 8px;
  height: 16px;
  background: var(--accent-primary);
  animation: blink 1s step-end infinite;
}

@keyframes blink {

  0%,
  50% {
    opacity: 1;
  }

  51%,
  100% {
    opacity: 0;
  }
}

.action-buttons {
  display: flex;
  gap: var(--space-sm);
  flex-wrap: wrap;
}

.btn-action {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-action .icon {
  font-size: 14px;
  line-height: 1;
}

.icon.spin {
  animation: spin 1s linear infinite;
}

.btn-primary {
  background: var(--accent-success);
  color: white;
  border-color: var(--accent-success);
}

.btn-primary:hover:not(:disabled) {
  background: #059669;
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.4);
}

.btn-secondary {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

.btn-tertiary {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.btn-tertiary:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-primary);
}

.btn-primary-push {
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  color: white;
  border-color: #8b5cf6;
}

.btn-primary-push:hover:not(:disabled) {
  background: #7c3aed;
  box-shadow: 0 0 15px rgba(139, 92, 246, 0.4);
}

.btn-primary-push:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* History section */
.history-section {
  max-height: 300px;
}

.history-count {
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-display);
  color: var(--text-muted);
  background: var(--bg-elevated);
  padding: 2px 8px;
  border-radius: 10px;
}

.history-list {
  max-height: 200px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.history-item {
  padding: var(--space-md);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.history-item:hover {
  border-color: var(--border-hover);
  background: var(--bg-tertiary);
}

.history-header-inline {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-sm);
}

.history-provider {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-display);
  text-transform: uppercase;
}

.provider-openai {
  background: rgba(16, 185, 129, 0.2);
  color: var(--accent-success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.provider-anthropic {
  background: rgba(139, 92, 246, 0.2);
  color: var(--accent-secondary);
  border: 1px solid rgba(139, 92, 246, 0.3);
}

.provider-deepseek {
  background: rgba(6, 182, 212, 0.2);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.provider-ollama {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.history-time {
  font-size: 11px;
  color: var(--text-muted);
}

.history-message {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* Empty state */
.empty-state-full {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: var(--space-2xl);
  text-align: center;
  animation: fade-in 0.5s ease-out;
}

.empty-illustration {
  margin-bottom: var(--space-lg);
  font-size: 64px;
  opacity: 0.3;
}

.empty-state-full h2 {
  margin: 0 0 var(--space-sm) 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-state-full p {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
}

/* Error banner */
.error-banner {
  position: fixed;
  bottom: var(--space-lg);
  right: var(--space-lg);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-md);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  z-index: var(--z-modal);
  animation: slide-up 0.3s ease-out;
}

.error-banner .icon {
  font-size: 20px;
  line-height: 1;
}

.error-message {
  flex: 1;
  font-size: 13px;
  color: var(--accent-error);
}

.error-dismiss {
  background: none;
  border: none;
  color: var(--accent-error);
  cursor: pointer;
  padding: 4px 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  font-size: 18px;
  line-height: 1;
}

.error-dismiss:hover {
  background: rgba(239, 68, 68, 0.2);
}

/* Toast Notification */
.toast-notification {
  position: fixed;
  bottom: var(--space-lg);
  right: var(--space-lg);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  border-radius: var(--radius-md);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  z-index: var(--z-modal);
  min-width: 280px;
}

.toast-notification.success {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.toast-notification.success .toast-icon {
  color: var(--accent-success);
}

.toast-notification.success .toast-message {
  color: var(--accent-success);
}

.toast-notification.error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.toast-notification.error .toast-icon {
  color: var(--accent-error);
}

.toast-notification.error .toast-message {
  color: var(--accent-error);
}

.toast-icon {
  font-size: 20px;
  font-weight: 600;
  line-height: 1;
}

.toast-message {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
}

/* Toast transition */
.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.3s ease-out;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* Transitions */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all var(--transition-normal);
}

.slide-down-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.slide-down-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

@keyframes slide-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slide-in {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }

  to {
    opacity: 1;
    transform: translateX(0);
  }
}

/* 新增样式 */
.config-badge {
  padding: 2px 8px;
  background: rgba(6, 182, 212, 0.2);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.btn-reset {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 11px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-reset:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.config-warning-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md);
  margin-bottom: var(--space-md);
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-md);
}

.warning-content {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  flex: 1;
}

.warning-content .icon {
  font-size: 18px;
  line-height: 1;
  flex-shrink: 0;
}

.warning-text strong {
  display: block;
  font-size: 13px;
  color: var(--accent-warning);
  margin-bottom: 2px;
}

.warning-text p {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.btn-confirm-reset {
  padding: var(--space-sm) var(--space-md);
  background: var(--accent-warning);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
}

.btn-confirm-reset:hover {
  filter: brightness(1.1);
}

.saving-indicator {
  margin-left: auto;
  font-size: 10px;
  color: var(--accent-primary);
  animation: pulse 1s infinite;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.5;
  }
}

.setting-select:disabled {
  opacity: 0.6;
  cursor: wait;
}

/* 新增样式：折叠/展开功能 */
.header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.collapse-icon {
  font-size: 12px;
  transition: transform var(--transition-fast);
}

.collapse-icon.expanded {
  transform: rotate(180deg);
}

.collapsed-info {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: var(--space-sm);
  padding: 2px 8px;
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
}

/* 历史记录内容区域 */
.history-content {
  /* 不需要额外的 padding，直接贴边 */
}

.history-message-full {
  margin: 0;
  padding: var(--space-sm) var(--space-xs) var(--space-sm) var(--space-xs);
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.history-lang {
  padding: 2px 8px;
  background: rgba(6, 182, 212, 0.15);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--accent-primary);
}

/* 整合式标题栏 */
.result-header {
  display: flex;
  align-items: center;
  /* 移除 space-between，改用 flex-start 让元素自然排列 */
  justify-content: flex-start;
  gap: var(--space-md);
  padding-bottom: var(--space-md);
  margin-bottom: var(--space-md);
  border-bottom: 1px solid var(--border-default);
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex-shrink: 0;
  flex-wrap: wrap;  /* 允许在小屏幕上换行 */
}

/* 新的左侧主按钮样式 */
.btn-generate-main {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  padding: 8px 16px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-normal);
  box-shadow: 0 2px 8px rgba(6, 182, 212, 0.3);
  white-space: nowrap;
  min-width: 100px;
}

.btn-generate-main .btn-icon {
  font-size: 16px;
  line-height: 1;
}

.btn-generate-main .btn-text {
  line-height: 1;
}

.btn-generate-main:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(6, 182, 212, 0.5);
}

.btn-generate-main:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-generate-main.generating {
  background: linear-gradient(135deg, var(--accent-secondary), var(--accent-primary));
  animation: pulse-glow 1.5s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% {
    box-shadow: 0 2px 8px rgba(6, 182, 212, 0.3);
  }
  50% {
    box-shadow: 0 4px 16px rgba(139, 92, 246, 0.5);
  }
}

.header-center {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex: 0 1 auto; /* 不强制占据所有空间 */
  justify-content: flex-start;
  flex-wrap: wrap;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex-shrink: 0;
  margin-left: auto; /* 将工具按钮推到右侧 */
}

/* 配置选择框包装器 */
.config-select-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.config-label {
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}

.config-select-inline {
  padding: 4px 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
  min-width: 100px;
}

.config-select-inline:hover {
  border-color: var(--border-hover);
}

.config-select-inline:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 2px rgba(6, 182, 212, 0.1);
}

.config-select-inline option {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.config-select-inline option:disabled {
  color: var(--text-muted);
  opacity: 0.6;
}

.config-select-inline:disabled {
  opacity: 0.6;
  cursor: wait;
}

/* 内联配置标记 */
.config-badge-inline {
  padding: 2px 8px;
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.config-badge-inline:hover {
  background: rgba(6, 182, 212, 0.25);
  border-color: rgba(6, 182, 212, 0.5);
}

/* 紧凑型图标按钮 */
.icon-btn-small {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  line-height: 1;
  flex-shrink: 0;
}

.icon-btn-small:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

/* 紧凑型操作按钮（标题栏内） */
.btn-action-inline {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 8px 14px;
  border: none;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.btn-action-inline .icon {
  font-size: 14px;
  line-height: 1;
}

.btn-primary-inline {
  background: var(--accent-success);
  color: white;
}

.btn-primary-inline:hover:not(:disabled) {
  background: #059669;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.4);
}

.btn-primary-inline:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-push-inline {
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  color: white;
}

.btn-push-inline:hover:not(:disabled) {
  background: #7c3aed;
  box-shadow: 0 0 12px rgba(139, 92, 246, 0.4);
}

.btn-push-inline:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* 辅助操作按钮区域（消息下方） */
.action-buttons-helper {
  display: flex;
  gap: var(--space-sm);
  justify-content: flex-start;
}

/* 内联警告提示 */
.config-warning-inline {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  margin-bottom: var(--space-md);
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
}

.config-warning-inline .icon {
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}

.config-warning-inline .btn-confirm-reset {
  margin-left: auto;
  padding: 2px 8px;
  background: var(--accent-warning);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
}

.config-warning-inline .btn-confirm-reset:hover {
  filter: brightness(1.1);
}

/* 响应式处理 */
@media (max-width: 768px) {
  .result-header {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-sm);
  }

  .header-left,
  .header-center,
  .header-right {
    width: 100%;
    justify-content: space-between;
  }

  .btn-generate-main {
    width: 100%;
  }
}
</style>
