<template>
  <div class="commit-panel">
    <!-- Project Info Section -->
    <div class="section" v-if="commitStore.projectStatus">
      <h3>{{ commitStore.projectStatus.branch }}</h3>
      <div class="staged-files">
        <div v-if="!commitStore.projectStatus.has_staged" class="empty-state">
          暂存区为空，请先 git add 文件
        </div>
        <div v-else>
          <div
            v-for="file in commitStore.projectStatus.staged_files"
            :key="file.path"
            class="file-item"
          >
            <span class="file-status" :class="file.status.toLowerCase()">
              {{ file.status }}
            </span>
            <span class="file-path">{{ file.path }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div class="section empty" v-else>
      <p>请从左侧选择一个项目</p>
    </div>

    <!-- AI Settings -->
    <div class="section" v-if="commitStore.projectStatus">
      <h3>AI 设置</h3>
      <div class="settings">
        <div class="setting-row">
          <label>Provider:</label>
          <select v-model="commitStore.provider">
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="deepseek">DeepSeek</option>
            <option value="ollama">Ollama</option>
          </select>
        </div>
        <div class="setting-row">
          <label>语言:</label>
          <select v-model="commitStore.language">
            <option value="zh">中文</option>
            <option value="en">English</option>
          </select>
        </div>
      </div>
      <button
        @click="handleGenerate"
        :disabled="!commitStore.projectStatus.has_staged || commitStore.isGenerating"
        class="btn-primary"
      >
        {{ commitStore.isGenerating ? '生成中...' : '生成 Commit 消息' }}
      </button>
    </div>

    <!-- Generated Message -->
    <div class="section" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
      <h3>生成结果</h3>
      <div class="message-area">
        <!-- Loading state with spinner -->
        <div v-if="commitStore.isGenerating" class="loading-indicator">
          <span class="spinner"></span>
          AI 正在生成...
        </div>
        <pre class="message-content">{{ commitStore.streamingMessage || commitStore.generatedMessage }}</pre>
      </div>
      <div class="actions">
        <button @click="handleCopy" class="btn-secondary">复制</button>
        <button @click="handleCommit" class="btn-primary">提交到本地</button>
        <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-secondary">重新生成</button>
      </div>
    </div>

    <!-- History Section -->
    <div class="section" v-if="history.length > 0">
      <h3>历史记录</h3>
      <div class="history-list">
        <div
          v-for="item in history"
          :key="item.id"
          class="history-item"
          @click="loadHistory(item)"
        >
          <div class="history-meta">
            <span class="history-provider">{{ item.provider }}</span>
            <span class="history-time">{{ formatTime(item.created_at) }}</span>
          </div>
          <div class="history-message">{{ item.message.substring(0, 100) }}{{ item.message.length > 100 ? '...' : '' }}</div>
        </div>
      </div>
    </div>

    <!-- Error (dismissible) -->
    <div class="section error" v-if="commitStore.error">
      <div class="error-content">
        <span class="error-message">{{ commitStore.error }}</span>
        <button @click="commitStore.error = null" class="error-dismiss">×</button>
      </div>
    </div>
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

// Time constants for relative time formatting
const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// Load history when project changes
watch(() => commitStore.selectedProjectPath, async (path) => {
  if (path) {
    await loadHistoryForProject()
  }
}, { immediate: true })

async function loadHistoryForProject() {
  // Find current project by path to get ID
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

    // Save to history
    const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
    if (project) {
      await SaveCommitHistory(project.id, message, commitStore.provider, commitStore.language)
    }

    alert('提交成功!')

    // Reload project status and history
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)
    await loadHistoryForProject()

    // Clear message
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
  padding: 20px;
  overflow-y: auto;
}

.section {
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.section h3 {
  margin-top: 0;
  margin-bottom: 15px;
}

.empty {
  text-align: center;
  color: #999;
}

.staged-files {
  max-height: 200px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 6px 0;
  font-size: 14px;
}

.file-status {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  margin-right: 8px;
  font-weight: bold;
}

.file-status.modified { background: #fff3cd; color: #856404; }
.file-status.new { background: #d1e7dd; color: #0f5132; }
.file-status.deleted { background: #f8d7da; color: #842029; }
.file-status.renamed { background: #cff4fc; color: #055160; }

.file-path {
  flex: 1;
  font-family: monospace;
  word-break: break-all;
}

.settings {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 15px;
}

.setting-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.setting-row label {
  min-width: 80px;
}

.setting-row select {
  flex: 1;
  padding: 6px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.message-area {
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  max-height: 300px;
  overflow-y: auto;
}

.loading-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  color: #2196f3;
  font-weight: 500;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #f3f3f3;
  border-top: 2px solid #2196f3;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.message-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.btn-primary, .btn-secondary {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: #2196f3;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #1976d2;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #5a6268;
}

.history-list {
  max-height: 200px;
  overflow-y: auto;
}

.history-item {
  padding: 10px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.history-item:hover {
  background: #f5f5f5;
}

.history-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

.history-provider {
  padding: 2px 6px;
  background: #e3f2fd;
  border-radius: 3px;
  font-weight: 500;
}

.history-time {
  color: #888;
}

.history-message {
  font-size: 13px;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.error {
  background: #f8d7da;
  color: #842029;
  border: 1px solid #f5c2c7;
}

.error-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.error-message {
  flex: 1;
}

.error-dismiss {
  background: none;
  border: none;
  color: #842029;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-dismiss:hover {
  background: rgba(0,0,0,0.1);
  border-radius: 4px;
}
</style>
