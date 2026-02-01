import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { LogFrontendError } from '../../wailsjs/go/main/App'

/**
 * 错误项数据结构
 */
export interface ErrorItem {
  id: string                    // 唯一标识（UUID）
  type: 'error' | 'warning'     // 错误类型
  message: string               // 主要错误消息（简短，1-2行）
  details?: string              // 详细错误信息（堆栈、调试信息）
  timestamp: number             // 时间戳
  source?: string               // 错误来源（如 'CommitPanel'）
}

/**
 * 生成 UUID
 */
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

export const useErrorStore = defineStore('error', () => {
  // 状态
  const errors = ref<ErrorItem[]>([])
  const maxErrors = ref(10)
  const maxWarnings = ref(5)

  // 计算属性：按类型分组错误
  const errorsByType = computed(() => {
    const grouped = {
      error: [] as ErrorItem[],
      warning: [] as ErrorItem[]
    }

    errors.value.forEach(err => {
      grouped[err.type].push(err)
    })

    return grouped
  })

  /**
   * 添加错误（同时显示UI和记录日志）
   */
  async function addError(
    message: string,
    details?: string,
    type: 'error' | 'warning' = 'error',
    source?: string
  ) {
    const error: ErrorItem = {
      id: generateUUID(),
      type,
      message,
      details,
      timestamp: Date.now(),
      source: source || 'Unknown'
    }

    // 添加到列表
    errors.value.push(error)

    // 自动清理旧错误
    await autoCleanup(type)

    // 发送到后端日志
    await sendToBackend(error)
  }

  /**
   * 自动清理旧错误
   */
  async function autoCleanup(type: 'error' | 'warning') {
    const limit = type === 'error' ? maxErrors.value : maxWarnings.value
    const typeErrors = errors.value.filter(e => e.type === type)

    if (typeErrors.length > limit) {
      // 移除最早的同类型错误
      const toRemove = typeErrors.slice(0, typeErrors.length - limit)
      toRemove.forEach(err => removeError(err.id))
    }
  }

  /**
   * 移除指定错误
   */
  function removeError(id: string) {
    const index = errors.value.findIndex(e => e.id === id)
    if (index !== -1) {
      errors.value.splice(index, 1)
    }
  }

  /**
   * 复制错误到剪贴板
   */
  async function copyError(id: string) {
    const error = errors.value.find(e => e.id === id)
    if (!error) {
      console.warn(`Error with id ${id} not found`)
      return
    }

    // 构建复制内容
    const text = [
      `[${error.type.toUpperCase()}] ${error.message}`,
      error.details ? `详情: ${error.details}` : '',
      `来源: ${error.source}`,
      `时间: ${new Date(error.timestamp).toLocaleString('zh-CN')}`
    ].filter(Boolean).join('\n')

    try {
      await navigator.clipboard.writeText(text)
      console.log('错误已复制到剪贴板')
    } catch (e) {
      console.error('复制失败:', e)
      throw new Error('复制失败，请手动选择文本复制')
    }
  }

  /**
   * 清除所有错误
   */
  function clearAll() {
    errors.value = []
  }

  /**
   * 发送到后端日志
   */
  async function sendToBackend(error: ErrorItem) {
    try {
      const errorJSON = JSON.stringify({
        type: error.type,
        message: error.message,
        details: error.details || '',
        source: error.source,
        timestamp: new Date(error.timestamp).toISOString()
      })

      await LogFrontendError(errorJSON)
      console.log('[ErrorStore] 已发送到后端日志:', error.message)
    } catch (e) {
      console.error('[ErrorStore] 发送到后端日志失败:', e)
      // 失败不影响 UI 显示，只记录错误
    }
  }

  /**
   * 获取指定类型的错误列表
   */
  function getErrorsByType(type: 'error' | 'warning'): ErrorItem[] {
    return errors.value.filter(e => e.type === type)
  }

  return {
    // 状态
    errors,
    maxErrors,
    maxWarnings,
    errorsByType,

    // 方法
    addError,
    removeError,
    copyError,
    clearAll,
    sendToBackend,
    getErrorsByType
  }
})
