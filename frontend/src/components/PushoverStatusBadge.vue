<template>
  <div
    v-if="status"
    class="pushover-badge"
    :class="statusClass"
    :title="tooltipText"
  >
    <span class="badge-icon">{{ statusIcon }}</span>
    <span v-if="!compact" class="badge-text">{{ statusText }}</span>
  </div>
  <div v-else-if="loading" class="pushover-badge loading">
    <span class="badge-icon">⏳</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { HookStatus } from '../types/pushover'

interface Props {
  status?: HookStatus | null
  loading?: boolean
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  compact: true
})

// 状态对应的样式类
const statusClass = computed(() => {
  if (!props.status) return ''
  if (!props.status.installed) return 'not-installed'
  return modeClass.value
})

// 通知模式对应的样式类
const modeClass = computed<{ [key: string]: boolean }>(() => {
  const mode = props.status?.mode
  return {
    'mode-enabled': mode === 'enabled',
    'mode-pushover-only': mode === 'pushover_only',
    'mode-windows-only': mode === 'windows_only',
    'mode-disabled': mode === 'disabled'
  }
})

// 状态图标
const statusIcon = computed(() => {
  if (!props.status) return '🔔'
  if (!props.status.installed) return '🔕'

  const mode = props.status.mode
  switch (mode) {
    case 'enabled':
      return '🔔'
    case 'pushover_only':
      return '📱'
    case 'windows_only':
      return '💻'
    case 'disabled':
      return '🔕'
    default:
      return '🔔'
  }
})

// 状态文本
const statusText = computed(() => {
  if (!props.status) return ''
  if (!props.status.installed) return '未安装'

  const mode = props.status.mode
  switch (mode) {
    case 'enabled':
      return '已启用'
    case 'pushover_only':
      return '仅 Pushover'
    case 'windows_only':
      return '仅 Windows'
    case 'disabled':
      return '已禁用'
    default:
      return '未知'
  }
})

// 提示文本
const tooltipText = computed(() => {
  if (!props.status) return '加载中...'
  if (!props.status.installed) return 'Pushover Hook 未安装'

  const modeText = {
    enabled: '全部启用',
    pushover_only: '仅 Pushover',
    windows_only: '仅 Windows',
    disabled: '已禁用'
  }[props.status.mode]

  const version = props.status.version ? ` (v${props.status.version})` : ''
  return `Pushover Hook: ${modeText}${version}`
})
</script>

<style scoped>
.pushover-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs, 4px);
  padding: var(--space-xs, 4px) var(--space-sm, 8px);
  border-radius: var(--radius-md, 6px);
  font-size: 12px;
  font-weight: 500;
  transition: all var(--transition-normal, 0.2s);
}

.badge-icon {
  font-size: 14px;
  line-height: 1;
}

.badge-text {
  line-height: 1;
}

/* 未安装状态 */
.pushover-badge.not-installed {
  background: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-default);
}

/* 已启用状态 */
.pushover-badge.mode-enabled {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

/* 仅 Pushover */
.pushover-badge.mode-pushover-only {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

/* 仅 Windows */
.pushover-badge.mode-windows-only {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

/* 已禁用 */
.pushover-badge.mode-disabled {
  background: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-default);
}

/* 加载状态 */
.pushover-badge.loading {
  background: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-default);
  opacity: 0.7;
}

/* Hover 效果 */
.pushover-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
