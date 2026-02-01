<template>
  <div class="error-toast-container">
    <transition-group name="error-slide" tag="div" class="error-list">
      <div
        v-for="error in errors"
        :key="error.id"
        class="error-card"
        :class="[`error-${error.type}`]"
      >
        <!-- 图标 -->
        <div class="error-icon">
          <span v-if="error.type === 'error'">❌</span>
          <span v-else>⚠️</span>
        </div>

        <!-- 消息内容 -->
        <div class="error-content">
          <div class="error-message">{{ error.message }}</div>
          <div v-if="error.details" class="error-details">{{ error.details }}</div>
          <div class="error-meta">
            {{ error.source }} · {{ formatTime(error.timestamp) }}
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="error-actions">
          <button
            @click="handleCopy(error)"
            class="action-btn"
            title="复制错误信息"
          >
            📋
          </button>
          <button
            @click="handleRemove(error.id)"
            class="action-btn close-btn"
            title="关闭"
          >
            ×
          </button>
        </div>
      </div>
    </transition-group>

    <!-- 清除所有按钮（如果有错误） -->
    <button
      v-if="errors.length > 0"
      @click="handleClearAll"
      class="clear-all-btn"
      title="清除所有错误"
    >
      清除全部 ({{ errors.length }})
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useErrorStore, type ErrorItem } from '../stores/errorStore'

const errorStore = useErrorStore()

// 获取所有错误，按时间排序（最早的在上面）
const errors = computed(() => {
  return [...errorStore.errors].sort((a, b) => a.timestamp - b.timestamp)
})

/**
 * 复制错误到剪贴板
 */
async function handleCopy(error: ErrorItem) {
  try {
    await errorStore.copyError(error.id)
    // 可以添加一个临时的视觉反馈
    console.log('已复制:', error.message)
  } catch (e) {
    console.error('复制失败:', e)
  }
}

/**
 * 移除单个错误
 */
function handleRemove(id: string) {
  errorStore.removeError(id)
}

/**
 * 清除所有错误
 */
function handleClearAll() {
  if (confirm(`确定要清除所有 ${errors.value.length} 条错误吗？`)) {
    errorStore.clearAll()
  }
}

/**
 * 格式化时间
 */
function formatTime(timestamp: number): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  // 小于 1 分钟
  if (diff < 60000) {
    return '刚刚'
  }

  // 小于 1 小时
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes} 分钟前`
  }

  // 小于 1 天
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours} 小时前`
  }

  // 显示完整时间
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.error-toast-container {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: var(--z-modal);
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  max-width: 420px;
  pointer-events: none; /* 让容器不阻挡点击，只有子元素可以点击 */
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
  pointer-events: auto;
}

/* 错误卡片 */
.error-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid;
  border-radius: var(--radius-md);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  min-width: 280px;
  max-width: 420px;
  pointer-events: auto;
  transition: all 0.2s ease;
}

.error-card:hover {
  transform: translateX(-2px);
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.4);
}

/* 错误类型样式 */
.error-card.error-error {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.3);
}

.error-card.error-error .error-icon {
  color: var(--accent-error);
}

.error-card.error-warning {
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.3);
}

.error-card.error-warning .error-icon {
  color: var(--accent-warning);
}

/* 图标 */
.error-icon {
  font-size: 20px;
  line-height: 1;
  flex-shrink: 0;
  padding-top: 2px;
}

/* 内容区域 */
.error-content {
  flex: 1;
  min-width: 0; /* 允许文本正确截断 */
}

.error-message {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  line-height: 1.4;
  word-break: break-word;
}

.error-details {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
  line-height: 1.4;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.error-meta {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 6px;
}

/* 操作按钮 */
.error-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 14px;
  color: var(--text-muted);
  transition: all 0.2s ease;
  padding: 0;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.action-btn.close-btn {
  font-size: 18px;
  font-weight: 300;
}

.action-btn.close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: var(--accent-error);
}

/* 清除全部按钮 */
.clear-all-btn {
  pointer-events: auto;
  padding: 6px 12px;
  background: var(--glass-bg);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

.clear-all-btn:hover {
  background: var(--bg-elevated);
  border-color: var(--border-hover);
  color: var(--text-primary);
}

/* 进入/离开动画 */
.error-slide-move,
.error-slide-enter-active,
.error-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.error-slide-enter-from {
  opacity: 0;
  transform: translateX(100%) scale(0.9);
}

.error-slide-leave-to {
  opacity: 0;
  transform: translateX(50px);
}

.error-slide-leave-active {
  position: absolute;
  right: 0;
  width: 100%;
}

/* 响应式调整 */
@media (max-width: 480px) {
  .error-toast-container {
    right: 10px;
    left: 10px;
    bottom: 10px;
    max-width: none;
  }

  .error-card {
    min-width: 0;
  }
}
</style>
