import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useErrorStore } from '../errorStore'

// Mock Wails API
vi.mock('../../../wailsjs/go/main/App', () => ({
  LogFrontendError: vi.fn(() => Promise.resolve())
}))

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(() => Promise.resolve()),
    readText: vi.fn(() => Promise.resolve(''))
  }
})

describe('ErrorStore', () => {
  beforeEach(() => {
    // 创建新的 Pinia 实例
    setActivePinia(createPinia())
  })

  describe('addError', () => {
    it('应该添加错误到列表', async () => {
      const store = useErrorStore()

      await store.addError('Test error', 'Test details', 'error', 'TestSource')

      expect(store.errors).toHaveLength(1)
      expect(store.errors[0].message).toBe('Test error')
      expect(store.errors[0].type).toBe('error')
      expect(store.errors[0].source).toBe('TestSource')
    })

    it('应该自动限制错误数量为10个', async () => {
      const store = useErrorStore()

      // 添加 15 个错误
      for (let i = 0; i < 15; i++) {
        await store.addError(`Error ${i}`, `details ${i}`, 'error')
      }

      // 应该只保留最新的 10 个
      expect(store.errors.length).toBe(10)
      expect(store.errors[0].message).toBe('Error 5')  // 最早保留的
      expect(store.errors[9].message).toBe('Error 14') // 最新的
    })

    it('应该自动限制警告数量为5个', async () => {
      const store = useErrorStore()

      // 添加 10 个警告
      for (let i = 0; i < 10; i++) {
        await store.addError(`Warning ${i}`, `details ${i}`, 'warning')
      }

      // 应该只保留最新的 5 个
      const warnings = store.getErrorsByType('warning')
      expect(warnings.length).toBe(5)
      expect(warnings[0].message).toBe('Warning 5')  // 最早保留的
      expect(warnings[4].message).toBe('Warning 9') // 最新的
    })

    it('应该生成唯一的 ID', async () => {
      const store = useErrorStore()

      await store.addError('Error 1', '', 'error')
      await store.addError('Error 2', '', 'error')

      expect(store.errors[0].id).not.toBe(store.errors[1].id)
    })

    it('应该记录时间戳', async () => {
      const store = useErrorStore()
      const before = Date.now()

      await store.addError('Test error', '', 'error')

      expect(store.errors[0].timestamp).toBeGreaterThanOrEqual(before)
      expect(store.errors[0].timestamp).toBeLessThanOrEqual(Date.now())
    })
  })

  describe('removeError', () => {
    it('应该移除指定的错误', async () => {
      const store = useErrorStore()

      await store.addError('Error 1', '', 'error')
      await store.addError('Error 2', '', 'error')

      const idToRemove = store.errors[0].id
      store.removeError(idToRemove)

      expect(store.errors).toHaveLength(1)
      expect(store.errors[0].message).toBe('Error 2')
    })

    it('移除不存在的ID不应该报错', () => {
      const store = useErrorStore()
      expect(() => store.removeError('non-existent-id')).not.toThrow()
    })
  })

  describe('copyError', () => {
    it('应该复制错误到剪贴板', async () => {
      const store = useErrorStore()

      await store.addError('Test error', 'Test details', 'error', 'TestSource')
      const errorId = store.errors[0].id

      await store.copyError(errorId)

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        expect.stringContaining('Test error')
      )
    })

    it('复制内容应该包含完整的错误信息', async () => {
      const store = useErrorStore()
      const mockWriteText = navigator.clipboard.writeText as ReturnType<typeof vi.fn>

      await store.addError('Test error', 'Test details', 'error', 'TestSource')
      const errorId = store.errors[0].id

      await store.copyError(errorId)

      const copiedText = mockWriteText.mock.calls[0][0] as string
      expect(copiedText).toContain('[ERROR]')
      expect(copiedText).toContain('Test error')
      expect(copiedText).toContain('Test details')
      expect(copiedText).toContain('TestSource')
    })

    it('复制不存在的错误应该记录警告', async () => {
      const store = useErrorStore()
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

      await store.copyError('non-existent-id')

      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('not found')
      )

      consoleSpy.mockRestore()
    })
  })

  describe('clearAll', () => {
    it('应该清除所有错误', async () => {
      const store = useErrorStore()

      await store.addError('Error 1', '', 'error')
      await store.addError('Error 2', '', 'warning')
      await store.addError('Error 3', '', 'error')

      store.clearAll()

      expect(store.errors).toHaveLength(0)
    })
  })

  describe('getErrorsByType', () => {
    it('应该返回指定类型的错误', async () => {
      const store = useErrorStore()

      await store.addError('Error 1', '', 'error')
      await store.addError('Warning 1', '', 'warning')
      await store.addError('Error 2', '', 'error')

      const errors = store.getErrorsByType('error')
      const warnings = store.getErrorsByType('warning')

      expect(errors).toHaveLength(2)
      expect(warnings).toHaveLength(1)
    })
  })

  describe('errorsByType', () => {
    it('应该计算并返回按类型分组的错误', async () => {
      const store = useErrorStore()

      await store.addError('Error 1', '', 'error')
      await store.addError('Warning 1', '', 'warning')
      await store.addError('Error 2', '', 'error')

      expect(store.errorsByType.error).toHaveLength(2)
      expect(store.errorsByType.warning).toHaveLength(1)
    })
  })

  describe('sendToBackend', () => {
    it('应该发送错误到后端', async () => {
      const { LogFrontendError } = await import('../../../wailsjs/go/main/App')
      const store = useErrorStore()

      await store.addError('Test error', 'Test details', 'error', 'TestSource')

      expect(LogFrontendError).toHaveBeenCalledWith(
        expect.stringContaining('Test error')
      )
    })

    it('后端发送失败不应该影响UI', async () => {
      const { LogFrontendError } = await import('../../../wailsjs/go/main/App')
      vi.mocked(LogFrontendError).mockRejectedValueOnce(new Error('Network error'))

      const store = useErrorStore()
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      await store.addError('Test error', '', 'error')

      // UI 应该仍然显示错误
      expect(store.errors).toHaveLength(1)
      // 控制台应该记录错误
      expect(consoleSpy).toHaveBeenCalled()

      consoleSpy.mockRestore()
    })
  })
})
