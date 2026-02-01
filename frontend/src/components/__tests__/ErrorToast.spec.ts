import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ErrorToast from '../ErrorToast.vue'
import { useErrorStore } from '../../stores/errorStore'

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

// Mock confirm
global.confirm = vi.fn(() => true)

describe('ErrorToast.vue', () => {
  let wrapper: VueWrapper
  let errorStore: ReturnType<typeof useErrorStore>

  beforeEach(() => {
    // 创建新的 Pinia 实例
    const pinia = createPinia()
    setActivePinia(pinia)

    // 挂载组件
    wrapper = mount(ErrorToast, {
      global: {
        plugins: [pinia]
      }
    })

    // 获取 store
    errorStore = useErrorStore()

    // 清空错误列表
    errorStore.clearAll()
  })

  describe('渲染', () => {
    it('当没有错误时不应该显示任何内容', () => {
      expect(wrapper.find('.error-card').exists()).toBe(false)
      expect(wrapper.find('.clear-all-btn').exists()).toBe(false)
    })

    it('应该显示错误列表', async () => {
      await errorStore.addError('Test error', 'Test details', 'error', 'TestSource')

      await wrapper.vm.$nextTick()

      const cards = wrapper.findAll('.error-card')
      expect(cards).toHaveLength(1)
      expect(cards[0].text()).toContain('Test error')
    })

    it('应该显示正确的错误图标', async () => {
      await errorStore.addError('Error', '', 'error')
      await errorStore.addError('Warning', '', 'warning')

      await wrapper.vm.$nextTick()

      const cards = wrapper.findAll('.error-card')
      expect(cards[0].find('.error-icon').text()).toBe('❌')
      expect(cards[1].find('.error-icon').text()).toBe('⚠️')
    })

    it('应该显示错误详情', async () => {
      await errorStore.addError('Test error', 'Test details', 'error')

      await wrapper.vm.$nextTick()

      const card = wrapper.find('.error-card')
      expect(card.text()).toContain('Test details')
    })

    it('应该显示错误来源和时间', async () => {
      await errorStore.addError('Test error', '', 'error', 'CommitPanel')

      await wrapper.vm.$nextTick()

      const card = wrapper.find('.error-card')
      expect(card.text()).toContain('CommitPanel')
    })

    it('多个错误应该按时间顺序堆叠（最早的在上）', async () => {
      await errorStore.addError('Error 1', '', 'error')
      await new Promise(resolve => setTimeout(resolve, 10)) // 确保时间戳不同
      await errorStore.addError('Error 2', '', 'error')
      await new Promise(resolve => setTimeout(resolve, 10))
      await errorStore.addError('Error 3', '', 'error')

      await wrapper.vm.$nextTick()

      const cards = wrapper.findAll('.error-card')
      expect(cards).toHaveLength(3)
      expect(cards[0].text()).toContain('Error 1') // 最早的在上面
      expect(cards[2].text()).toContain('Error 3') // 最新的在下面
    })

    it('应该显示清除全部按钮', async () => {
      await errorStore.addError('Error 1', '', 'error')
      await errorStore.addError('Error 2', '', 'error')

      await wrapper.vm.$nextTick()

      const clearBtn = wrapper.find('.clear-all-btn')
      expect(clearBtn.exists()).toBe(true)
      expect(clearBtn.text()).toContain('清除全部 (2)')
    })
  })

  describe('交互', () => {
    it('点击关闭按钮应该移除错误', async () => {
      await errorStore.addError('Test error', '', 'error')

      await wrapper.vm.$nextTick()

      const closeBtn = wrapper.find('.close-btn')
      await closeBtn.trigger('click')

      await wrapper.vm.$nextTick()

      expect(errorStore.errors).toHaveLength(0)
    })

    it('点击复制按钮应该调用 copyError', async () => {
      await errorStore.addError('Test error', 'Test details', 'error')

      await wrapper.vm.$nextTick()

      const copySpy = vi.spyOn(errorStore, 'copyError')
      const copyBtn = wrapper.findAll('.action-btn')[0]
      await copyBtn.trigger('click')

      expect(copySpy).toHaveBeenCalledWith(errorStore.errors[0].id)
    })

    it('点击清除全部按钮应该清空所有错误', async () => {
      await errorStore.addError('Error 1', '', 'error')
      await errorStore.addError('Error 2', '', 'error')

      await wrapper.vm.$nextTick()

      const clearBtn = wrapper.find('.clear-all-btn')
      await clearBtn.trigger('click')

      await wrapper.vm.$nextTick()

      expect(errorStore.errors).toHaveLength(0)
    })

    it('取消清除全部不应该删除错误', async () => {
      vi.mocked(global.confirm).mockReturnValueOnce(false)

      await errorStore.addError('Error 1', '', 'error')
      await errorStore.addError('Error 2', '', 'error')

      await wrapper.vm.$nextTick()

      const clearBtn = wrapper.find('.clear-all-btn')
      await clearBtn.trigger('click')

      await wrapper.vm.$nextTick()

      expect(errorStore.errors).toHaveLength(2)
    })
  })

  describe('样式类', () => {
    it('错误应该有 error-error 类', async () => {
      await errorStore.addError('Test error', '', 'error')

      await wrapper.vm.$nextTick()

      const card = wrapper.find('.error-card')
      expect(card.classes()).toContain('error-error')
    })

    it('警告应该有 error-warning 类', async () => {
      await errorStore.addError('Test warning', '', 'warning')

      await wrapper.vm.$nextTick()

      const card = wrapper.find('.error-card')
      expect(card.classes()).toContain('error-warning')
    })
  })

  describe('响应式', () => {
    it('添加新错误后应该自动更新显示', async () => {
      expect(wrapper.findAll('.error-card')).toHaveLength(0)

      await errorStore.addError('Error 1', '', 'error')
      await wrapper.vm.$nextTick()

      expect(wrapper.findAll('.error-card')).toHaveLength(1)

      await errorStore.addError('Error 2', '', 'error')
      await wrapper.vm.$nextTick()

      expect(wrapper.findAll('.error-card')).toHaveLength(2)
    })

    it('移除错误后应该自动更新显示', async () => {
      await errorStore.addError('Error 1', '', 'error')
      await errorStore.addError('Error 2', '', 'error')

      await wrapper.vm.$nextTick()
      expect(wrapper.findAll('.error-card')).toHaveLength(2)

      errorStore.removeError(errorStore.errors[0].id)
      await wrapper.vm.$nextTick()

      expect(wrapper.findAll('.error-card')).toHaveLength(1)
    })
  })

  describe('时间格式化', () => {
    it('应该正确格式化时间', async () => {
      const now = Date.now()
      await errorStore.addError('Test error', '', 'error')

      await wrapper.vm.$nextTick()

      const card = wrapper.find('.error-card')
      expect(card.text()).toContain('刚刚')
    })
  })
})
