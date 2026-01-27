<template>
  <div class="diff-test-page">
    <div class="test-header">
      <h1>v-code-diff 测试</h1>
      <div class="test-controls">
        <button @click="loadSampleDiff" class="btn-test">加载示例 Diff</button>
        <button @click="toggleFormat" class="btn-test btn-secondary">
          切换: {{ format === 'line-by-line' ? '并排' : '逐行' }}
        </button>
        <button @click="clearDiff" class="btn-test btn-secondary">清除</button>
      </div>
    </div>

    <div class="test-content">
      <div class="input-section">
        <h3>旧代码</h3>
        <textarea
          v-model="oldCode"
          class="diff-input"
          placeholder="旧代码内容..."
        ></textarea>
      </div>

      <div class="input-section">
        <h3>新代码</h3>
        <textarea
          v-model="newCode"
          class="diff-input"
          placeholder="新代码内容..."
        ></textarea>
      </div>
    </div>

    <div class="diff-result-section">
      <h3>Diff 结果</h3>
      <div v-if="oldCode || newCode" class="diff-render">
        <CodeDiff
          :old-string="oldCode"
          :new-string="newCode"
          :output-format="format"
          :context="10"
        />
      </div>
      <div v-else class="empty-state">
        请加载示例或在上方输入代码
      </div>
    </div>

    <div class="status-bar">
      <span class="status-item">库: v-code-diff</span>
      <span class="status-item">模式: {{ format }}</span>
      <span class="status-item">上下文行数: 10</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CodeDiff } from 'v-code-diff'

const oldCode = ref('')
const newCode = ref('')
const format = ref<'line-by-line' | 'side-by-side'>('line-by-line')

const sampleOldCode = `import { defineStore } from 'pinia'
import type { CommitHistory } from '../types'

export const useCommitStore = defineStore('commit', () => {
  const projects = ref<Project[]>([])
  const selectedProject = ref<Project | null>(null)
`

const sampleNewCode = `import { defineStore } from 'pinia'
import type { CommitHistory } from '../types'

export const useCommitStore = defineStore('commit', () => {
  // 暂存管理状态
  const stagedFiles = ref<GitFile[]>([])
  const unstagedFiles = ref<GitFile[]>([])
  const selectedFile = ref<GitFile | null>(null)
  const fileDiff = ref<string | null>(null)

  const projects = ref<Project[]>([])
  const selectedProject = ref<Project | null>(null)
`

function loadSampleDiff() {
  oldCode.value = sampleOldCode
  newCode.value = sampleNewCode
}

function clearDiff() {
  oldCode.value = ''
  newCode.value = ''
}

function toggleFormat() {
  format.value = format.value === 'line-by-line' ? 'side-by-side' : 'line-by-line'
}
</script>

<style scoped>
.diff-test-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 20px;
  gap: 20px;
  overflow: hidden;
}

.test-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 8px;
}

.test-header h1 {
  margin: 0;
  font-size: 20px;
  color: #e0e0e0;
}

.test-controls {
  display: flex;
  gap: 12px;
}

.btn-test {
  padding: 8px 16px;
  background: #06b6d4;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-test.btn-secondary {
  background: #2d2d2d;
  color: #e0e0e0;
  border: 1px solid #444;
}

.test-content {
  flex: 0 0 300px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.input-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 12px;
}

.input-section h3 {
  margin: 0;
  font-size: 14px;
  color: #e0e0e0;
}

.diff-input {
  flex: 1;
  padding: 12px;
  background: #121212;
  border: 1px solid #333;
  border-radius: 6px;
  color: #e0e0e0;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  resize: none;
}

.diff-result-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 12px;
  overflow: hidden;
}

.diff-result-section h3 {
  margin: 0;
  font-size: 14px;
  color: #e0e0e0;
}

.diff-render {
  flex: 1;
  overflow: auto;
  background: #121212;
  border: 1px solid #333;
  border-radius: 6px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888;
  background: #121212;
  border: 1px solid #333;
  border-radius: 6px;
}

.status-bar {
  display: flex;
  gap: 24px;
  padding: 12px 16px;
  background: #2d2d2d;
  border: 1px solid #333;
  border-radius: 6px;
}

.status-item {
  font-size: 12px;
  color: #888;
}
</style>
