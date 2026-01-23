<template>
  <div class="commit-panel">
    <!-- Project Info Section -->
    <section class="panel-section" v-if="commitStore.projectStatus">
      <div class="section-header">
        <div class="section-title">
          <span class="icon">📊</span>
          <h3>当前状态</h3>
        </div>
        <div class="branch-badge">
          <span class="icon">⑂</span>
          {{ commitStore.projectStatus.branch }}
        </div>
      </div>

      <div class="staged-files-container">
        <div v-if="!commitStore.projectStatus.has_staged" class="empty-state-compact">
          <div class="icon">📄</div>
          <p>暂存区为空</p>
          <span class="hint">请先使用 git add 添加文件</span>
        </div>
        <div v-else class="files-list">
          <div
            v-for="file in commitStore.projectStatus.staged_files"
            :key="file.path"
            class="file-item"
          >
            <span class="file-status" :class="file.status.toLowerCase()">
              {{ file.status }}
            </span>
            <span class="file-icon">📄</span>
            <span class="file-path">{{ file.path }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Empty State -->
    <div class="empty-state-full" v-else>
      <div class="empty-illustration">📁</div>
      <h2>未选择项目</h2>
      <p>请从左侧列表选择一个项目</p>
    </div>

    <!-- AI Settings -->
    <section class="panel-section" v-if="commitStore.projectStatus">
      <div class="section-header">
        <div class="section-title">
          <span class="icon">🤖</span>
          <h3>AI 配置</h3>
        </div>
      </div>

      <div class="settings-grid">
        <div class="setting-group">
          <label class="setting-label">
            <span class="icon">🌐</span>
            Provider
          </label>
          <select v-model="commitStore.provider" class="setting-select">
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="deepseek">DeepSeek</option>
            <option value="ollama">Ollama</option>
            <option value="google">Google</option>
            <option value="phind">Phind</option>
          </select>
        </div>

        <div class="setting-group">
          <label class="setting-label">
            <span class="icon">🌍</span>
            语言
          </label>
          <select v-model="commitStore.language" class="setting-select">
            <option value="zh">中文</option>
            <option value="en">English</option>
          </select>
        </div>
      </div>

      <button
        @click="handleGenerate"
        :disabled="!commitStore.projectStatus.has_staged || commitStore.isGenerating"
        class="btn-generate"
        :class="{ generating: commitStore.isGenerating }"
      >
        <span class="icon" v-if="!commitStore.isGenerating">⚡</span>
        <span class="icon spin" v-else>⏳</span>
        {{ commitStore.isGenerating ? '生成中...' : '生成 Commit 消息' }}
      </button>
    </section>

    <!-- Generated Message -->
    <section class="panel-section result-section" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
      <div class="section-header">
        <div class="section-title">
          <span class="icon">✅</span>
          <h3>生成结果</h3>
        </div>
        <button
          @click="commitStore.clearMessage"
          class="icon-btn"
          title="清除"
        >×</button>
      </div>

      <div class="message-container">
        <!-- Streaming indicator -->
        <div v-if="commitStore.isGenerating" class="streaming-indicator">
          <span class="streaming-dot"></span>
          <span class="streaming-dot"></span>
          <span class="streaming-dot"></span>
        </div>

        <pre class="message-content">{{ commitStore.streamingMessage || commitStore.generatedMessage }}</pre>
      </div>

      <div class="action-buttons">
        <button @click="handleCopy" class="btn-action btn-secondary">
          <span class="icon">📋</span>
          复制
        </button>
        <button @click="handleCommit" class="btn-action btn-primary">
          <span class="icon">✓</span>
          提交到本地
        </button>
        <button
          @click="handleRegenerate"
          :disabled="commitStore.isGenerating"
          class="btn-action btn-tertiary"
        >
          <span class="icon">🔄</span>
          重新生成
        </button>
      </div>
    </section>

    <!-- History Section -->
    <section class="panel-section history-section" v-if="history.length > 0">
      <div class="section-header">
        <div class="section-title">
          <span class="icon">📜</span>
          <h3>历史记录</h3>
        </div>
        <span class="history-count">{{ history.length }}</span>
      </div>

      <div class="history-list">
        <div
          v-for="item in history"
          :key="item.id"
          class="history-item"
          @click="loadHistory(item)"
        >
          <div class="history-header">
            <span class="history-provider" :class="'provider-' + item.provider">
              {{ item.provider }}
            </span>
            <span class="history-time">{{ formatTime(item.created_at) }}</span>
          </div>
          <div class="history-message">{{ item.message.substring(0, 80) }}{{ item.message.length > 80 ? '...' : '' }}</div>
        </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import { useProjectStore } from '../stores/projectStore'
import { GetProjectHistory, SaveCommitHistory, CommitLocally } from '../../wailsjs/go/main/App'
import type { CommitHistory } from '../types'

const commitStore = useCommitStore()
const projectStore = useProjectStore()
const history = ref<CommitHistory[]>([])

const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

watch(() => commitStore.selectedProjectPath, async (path) => {
  if (path) {
    await loadHistoryForProject()
  }
}, { immediate: true })

async function loadHistoryForProject() {
  const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
  if (!project) return

  try {
    const result = await GetProjectHistory(project.id)
    history.value = result || []
  } catch (e) {
    console.error('Failed to load history:', e)
  }
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < MINUTE) return '刚刚'
  if (diff < HOUR) return `${Math.floor(diff / MINUTE)} 分钟前`
  if (diff < DAY) return `${Math.floor(diff / HOUR)} 小时前`
  return date.toLocaleDateString()
}

function loadHistory(item: CommitHistory) {
  commitStore.generatedMessage = item.message
}

async function handleGenerate() {
  await commitStore.generateCommit()
}

async function handleCopy() {
  const text = commitStore.streamingMessage || commitStore.generatedMessage
  await navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

async function handleCommit() {
  if (!commitStore.selectedProjectPath) {
    alert('请先选择项目')
    return
  }

  const message = commitStore.streamingMessage || commitStore.generatedMessage
  if (!message) {
    alert('请先生成 commit 消息')
    return
  }

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)

    const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
    if (project) {
      await SaveCommitHistory(project.id, message, commitStore.provider, commitStore.language)
    }

    alert('提交成功!')
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)
    await loadHistoryForProject()
    commitStore.clearMessage()
  } catch (e: unknown) {
    const errMessage = e instanceof Error ? e.message : '提交失败'
    alert('提交失败: ' + errMessage)
  }
}

async function handleRegenerate() {
  commitStore.clearMessage()
  await commitStore.generateCommit()
}
</script>

<style scoped>
.commit-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: var(--space-lg);
  overflow-y: auto;
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

/* Staged files */
.staged-files-container {
  max-height: 200px;
  overflow-y: auto;
}

.empty-state-compact {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl) var(--space-lg);
  text-align: center;
}

.empty-state-compact .icon {
  font-size: 32px;
  margin-bottom: var(--space-sm);
  opacity: 0.4;
}

.empty-state-compact p {
  margin: 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.empty-state-compact .hint {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: var(--space-xs);
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.file-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.file-item:hover {
  border-color: var(--border-hover);
}

.file-status {
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-display);
  text-transform: uppercase;
  flex-shrink: 0;
}

.file-status.modified {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.file-status.new {
  background: rgba(16, 185, 129, 0.2);
  color: var(--accent-success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.file-status.deleted {
  background: rgba(239, 68, 68, 0.2);
  color: var(--accent-error);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.file-status.renamed {
  background: rgba(59, 130, 246, 0.2);
  color: var(--accent-info);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.file-icon {
  color: var(--text-muted);
  flex-shrink: 0;
  font-size: 14px;
  line-height: 1;
}

.file-path {
  flex: 1;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
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
  padding: var(--space-md);
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
  0%, 60%, 100% {
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

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-xs);
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
</style>
