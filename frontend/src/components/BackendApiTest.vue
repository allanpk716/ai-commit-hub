<template>
  <div class="api-test-page">
    <div class="test-header">
      <h1>后端 Git API 测试</h1>
      <div class="test-controls">
        <button @click="testGetStagingStatus" class="btn-test">测试获取暂存状态</button>
        <button @click="clearResults" class="btn-test btn-secondary">清除结果</button>
      </div>
    </div>

    <div class="project-input">
      <label>项目路径:</label>
      <input v-model="projectPath" class="path-input" placeholder="输入 Git 项目路径..." />
      <button @click="loadCurrentProject" class="btn-load">使用当前项目</button>
    </div>

    <!-- 暂存状态 -->
    <div class="test-section" v-if="stagingStatus">
      <h3>暂存区状态</h3>
      <div class="staging-grid">
        <div class="staging-column">
          <h4>已暂存 ({{ stagingStatus.staged?.length ?? 0 }})</h4>
          <div class="file-list">
            <div v-for="file in stagingStatus.staged" :key="file.path" class="file-item staged">
              <span class="file-status">{{ file.status }}</span>
              <span class="file-path">{{ file.path }}</span>
              <button @click="testUnstageFile(file.path)" class="btn-mini">-</button>
            </div>
            <div v-if="!stagingStatus.staged?.length" class="empty">无已暂存文件</div>
          </div>
        </div>

        <div class="staging-column">
          <h4>未暂存 ({{ stagingStatus.unstaged?.length ?? 0 }})</h4>
          <div class="file-list">
            <div
              v-for="file in stagingStatus.unstaged"
              :key="file.path"
              :class="['file-item', 'unstaged', { 'ignored': file.ignored }]"
            >
              <span class="file-status">{{ file.status }}</span>
              <span class="ignored-badge" v-if="file.ignored">已忽略</span>
              <span class="file-path">{{ file.path }}</span>
              <button
                @click="testStageFile(file.path)"
                class="btn-mini"
                :disabled="file.ignored"
                :title="file.ignored ? '此文件被 .gitignore 忽略，需要强制添加' : '暂存'"
              >
                +
              </button>
            </div>
            <div v-if="!stagingStatus.unstaged?.length" class="empty">工作区干净</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Diff 预览 -->
    <div class="test-section" v-if="fileDiff">
      <h3>文件 Diff: {{ selectedFilePath }}</h3>
      <CodeDiff
        :old-string="fileDiff.old || ''"
        :new-string="fileDiff.new || ''"
        :output-format="'line-by-line'"
        :context="5"
      />
    </div>

    <!-- 操作日志 -->
    <div class="test-section log-section">
      <h3>操作日志</h3>
      <div class="log-container">
        <div v-for="(log, index) in logs" :key="index" :class="['log-item', log.type]">
          <span class="log-time">{{ log.time }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { CodeDiff } from 'v-code-diff'
import {
  GetStagingStatus,
  StageFile,
  UnstageFile
} from '../../wailsjs/go/main/App'

const projectPath = ref('')
const stagingStatus = ref<any>(null)
const selectedFilePath = ref('')
const fileDiff = ref<any>(null)
const logs = ref<Array<{ time: string; message: string; type: string }>>([])

function addLog(message: string, type = 'info') {
  const time = new Date().toLocaleTimeString()
  logs.value.unshift({ time, message, type })
}

async function loadCurrentProject() {
  // 使用 ai-commit-hub 项目路径
  projectPath.value = 'C:\\WorkSpace\\Go2Hell\\src\\github.com\\allanpk716\\ai-commit-hub'
  addLog(`加载项目路径: ${projectPath.value}`)
  await testGetStagingStatus()
}

async function testGetStagingStatus() {
  if (!projectPath.value) {
    addLog('请先输入项目路径', 'error')
    return
  }

  try {
    addLog('调用 GetStagingStatus...')
    const result: any = await GetStagingStatus(projectPath.value)
    console.log('GetStagingStatus 返回:', result)

    // StagingStatus 对象: {staged: [], unstaged: []}
    const staged = result?.staged || []
    const unstaged = result?.unstaged || []

    stagingStatus.value = { staged, unstaged }
    addLog(`成功获取暂存状态: 已暂存 ${staged.length} 个文件, 未暂存 ${unstaged.length} 个文件`, 'success')
  } catch (e: any) {
    addLog(`获取失败: ${e.message || e}`, 'error')
    console.error(e)
  }
}

async function testStageFile(filePath: string) {
  try {
    addLog(`暂存文件: ${filePath}`)
    await StageFile(projectPath.value, filePath)
    addLog(`暂存成功: ${filePath}`, 'success')
    await testGetStagingStatus() // 刷新状态
  } catch (e: any) {
    addLog(`暂存失败: ${e.message || e}`, 'error')
  }
}

async function testUnstageFile(filePath: string) {
  try {
    addLog(`取消暂存: ${filePath}`)
    await UnstageFile(projectPath.value, filePath)
    addLog(`取消暂存成功: ${filePath}`, 'success')
    await testGetStagingStatus() // 刷新状态
  } catch (e: any) {
    addLog(`取消暂存失败: ${e.message || e}`, 'error')
  }
}

function clearResults() {
  stagingStatus.value = null
  fileDiff.value = null
  selectedFilePath.value = ''
  logs.value = []
  addLog('已清除所有结果')
}

onMounted(() => {
  addLog('后端 API 测试页面已加载')
  addLog('请输入项目路径并点击"测试获取暂存状态"')
})
</script>

<style scoped>
.api-test-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 20px;
  gap: 20px;
  overflow: hidden;
  background: #1e1e1e;
  color: #e0e0e0;
}

.test-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: #2d2d2d;
  border: 1px solid #444;
  border-radius: 8px;
}

.test-header h1 {
  margin: 0;
  font-size: 20px;
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
  background: #444;
}

.project-input {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #2d2d2d;
  border: 1px solid #444;
  border-radius: 8px;
}

.path-input {
  flex: 1;
  padding: 8px 12px;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 6px;
  color: #e0e0e0;
}

.btn-load {
  padding: 8px 16px;
  background: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.test-section {
  background: #2d2d2d;
  border: 1px solid #444;
  border-radius: 8px;
  padding: 16px;
  overflow: hidden;
}

.test-section h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.staging-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.staging-column h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #888;
}

.file-list {
  max-height: 200px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 4px;
  margin-bottom: 4px;
}

.file-item.staged {
  border-color: #10b981;
}

.file-item.unstaged {
  border-color: #f59e0b;
}

.file-status {
  padding: 2px 6px;
  font-size: 10px;
  border-radius: 4px;
  background: #444;
}

.file-path {
  flex: 1;
  font-family: monospace;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ignored-badge {
  padding: 2px 6px;
  font-size: 9px;
  border-radius: 4px;
  background: #666;
  color: #aaa;
  white-space: nowrap;
}

.file-item.ignored {
  opacity: 0.6;
  background: #2a2a2a !important;
  border-color: #666 !important;
}

.file-item.ignored .file-path {
  color: #888;
  text-decoration: line-through;
}

.file-item.ignored .btn-mini {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-mini {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
}

.file-item.staged .btn-mini {
  background: #ef4444;
  color: white;
}

.file-item.unstaged .btn-mini {
  background: #10b981;
  color: white;
}

.empty {
  padding: 16px;
  text-align: center;
  color: #666;
  font-size: 13px;
}

.log-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.log-container {
  flex: 1;
  overflow-y: auto;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 6px;
  padding: 8px;
}

.log-item {
  display: flex;
  gap: 12px;
  padding: 6px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.log-item.info {
  color: #e0e0e0;
}

.log-item.success {
  color: #10b981;
}

.log-item.error {
  color: #ef4444;
}

.log-time {
  color: #666;
}
</style>
