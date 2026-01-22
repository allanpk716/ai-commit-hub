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
            <option value="english">English</option>
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
        <pre class="message-content">{{ commitStore.streamingMessage || commitStore.generatedMessage }}</pre>
      </div>
      <div class="actions">
        <button @click="handleCopy" class="btn-secondary">复制</button>
        <button @click="handleCommit" class="btn-primary">提交到本地</button>
        <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-secondary">重新生成</button>
      </div>
    </div>

    <!-- Error -->
    <div class="section error" v-if="commitStore.error">
      {{ commitStore.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCommitStore } from '../stores/commitStore'

const commitStore = useCommitStore()

async function handleGenerate() {
  await commitStore.generateCommit()
}

async function handleCopy() {
  const text = commitStore.streamingMessage || commitStore.generatedMessage
  await navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

async function handleCommit() {
  // TODO: Implement in next task
  alert('提交功能将在下一步实现')
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

.error {
  background: #f8d7da;
  color: #842029;
  border: 1px solid #f5c2c7;
}
</style>
