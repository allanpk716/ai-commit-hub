<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="about-dialog">
      <div class="dialog-header">
        <h2>关于 AI Commit Hub</h2>
        <button @click="close" class="close-btn">×</button>
      </div>

      <div class="dialog-body">
        <!-- Logo 和标题 -->
        <div class="app-info">
          <img src="../assets/app-icon.png" alt="AI Commit Hub" class="app-logo" />
          <h1>AI Commit Hub</h1>
          <p class="tagline">简化 Git 工作流 - AI 自动生成 commit 消息</p>
        </div>

        <!-- 版本信息 -->
        <div class="version-info">
          <div class="version-row">
            <span class="label">当前版本:</span>
            <span class="value">{{ version }}</span>
          </div>

          <div v-if="fullVersion !== version" class="version-row details">
            <span class="label">完整信息:</span>
            <span class="value monospace">{{ fullVersion }}</span>
          </div>
        </div>

        <!-- 功能特性 -->
        <div class="features">
          <h3>核心功能</h3>
          <ul>
            <li>✨ AI 驱动的 commit 消息生成</li>
            <li>🔄 支持多个 Git 仓库管理</li>
            <li>🏷️ 自动生成语义化版本号</li>
            <li>🔔 Pushover 和系统通知集成</li>
            <li>🔄 自动更新支持</li>
          </ul>
        </div>

        <!-- 链接 -->
        <div class="links">
          <a href="https://github.com/allanpk716/ai-commit-hub" target="_blank" class="link">
            <span class="icon">🔗</span>
            <span>GitHub 仓库</span>
          </a>
          <a href="https://github.com/allanpk716/ai-commit-hub/releases" target="_blank" class="link">
            <span class="icon">📦</span>
            <span>更新日志</span>
          </a>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="checkForUpdates" class="btn btn-secondary">
          <span class="icon">🔄</span>
          <span>检查更新</span>
        </button>
        <button @click="close" class="btn btn-primary">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { GetVersion, GetFullVersion } from '../../wailsjs/go/main/App'
import { useUpdateStore } from '../stores/updateStore'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits(['close'])

const updateStore = useUpdateStore()
const version = ref('加载中...')
const fullVersion = ref('')

// 加载版本信息
async function loadVersion() {
  try {
    version.value = await GetVersion()
    fullVersion.value = await GetFullVersion()
  } catch (error) {
    console.error('获取版本信息失败:', error)
    version.value = '未知版本'
    fullVersion.value = ''
  }
}

// 检查更新
async function checkForUpdates() {
  try {
    await updateStore.checkForUpdates()
  } catch (error) {
    console.error('检查更新失败:', error)
  }
}

// 关闭对话框
function close() {
  emit('close')
}

// 组件挂载时加载版本
onMounted(() => {
  if (props.visible) {
    loadVersion()
  }
})
</script>

<style scoped>
.about-dialog {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 500px;
  max-width: 90vw;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  line-height: 1;
  color: #9ca3af;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #4b5563;
}

.dialog-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.app-info {
  text-align: center;
  margin-bottom: 24px;
}

.app-logo {
  width: 80px;
  height: 80px;
  margin-bottom: 16px;
}

.app-info h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
}

.tagline {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.version-info {
  background: #f9fafb;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.version-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.version-row:last-child {
  margin-bottom: 0;
}

.version-row .label {
  font-weight: 500;
  color: #6b7280;
}

.version-row .value {
  font-weight: 600;
  color: #1f2937;
}

.version-row.details .value {
  font-size: 12px;
  color: #4b5563;
}

.monospace {
  font-family: 'Consolas', 'Monaco', monospace;
}

.features {
  margin-bottom: 24px;
}

.features h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.features ul {
  margin: 0;
  padding-left: 20px;
  list-style: none;
}

.features li {
  margin-bottom: 8px;
  color: #4b5563;
  font-size: 14px;
}

.links {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: #f3f4f6;
  border-radius: 8px;
  color: #1f2937;
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.link:hover {
  background: #e5e7eb;
  color: #111827;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-secondary {
  background: white;
  color: #1f2937;
  border: 1px solid #d1d5db;
}

.btn-secondary:hover {
  background: #f9fafb;
  border-color: #9ca3af;
}

.modal-overlay {
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
</style>
