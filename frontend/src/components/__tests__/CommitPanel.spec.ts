import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import CommitPanel from '../CommitPanel.vue'
import CommitControls from '../CommitControls.vue'
import CommitMessage from '../CommitMessage.vue'

// Mock Wails runtime
vi.mock('@/wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn()
}))

// Mock Wails app methods
vi.mock('@/wailsjs/go/main/App', () => ({
  GetAllProjects: vi.fn(() => Promise.resolve([])),
  GetProjectStatus: vi.fn(() => Promise.resolve({
    branch: 'main',
    has_uncommitted_changes: false
  }))
}))

describe('CommitPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should render commit controls', () => {
    const wrapper = mount(CommitPanel, {
      global: {
        stubs: {
          CommitControls: true,
          CommitMessage: true,
          UpdateNotification: true,
          UpdateDialog: true,
          ProjectStatusHeader: true,
          StagingArea: true
        }
      }
    })

    expect(wrapper.findComponent(CommitControls).exists()).toBe(true)
  })

  it('should render commit message component', () => {
    const wrapper = mount(CommitPanel, {
      global: {
        stubs: {
          CommitControls: true,
          CommitMessage: true,
          UpdateNotification: true,
          UpdateDialog: true,
          ProjectStatusHeader: true,
          StagingArea: true
        }
      }
    })

    expect(wrapper.findComponent(CommitMessage).exists()).toBe(true)
  })

  it('should disable generate button when no project selected', () => {
    const wrapper = mount(CommitPanel, {
      global: {
        stubs: {
          CommitControls: true,
          CommitMessage: true,
          UpdateNotification: true,
          UpdateDialog: true,
          ProjectStatusHeader: true,
          StagingArea: true
        }
      }
    })

    const controls = wrapper.findComponent(CommitControls)
    expect(controls.props('canGenerate')).toBe(false)
  })
})

describe('CommitControls', () => {
  it('should emit generate event when generate button clicked', async () => {
    const wrapper = mount(CommitControls, {
      props: {
        canGenerate: true,
        isGenerating: false,
        canCommit: false,
        canPush: false
      }
    })

    await wrapper.find('.btn-primary').trigger('click')
    expect(wrapper.emitted('generate')).toBeTruthy()
  })

  it('should emit commit event when commit button clicked', async () => {
    const wrapper = mount(CommitControls, {
      props: {
        canGenerate: false,
        isGenerating: false,
        canCommit: true,
        canPush: false
      }
    })

    await wrapper.find('.btn-success').trigger('click')
    expect(wrapper.emitted('commit')).toBeTruthy()
  })

  it('should emit push event when push button clicked', async () => {
    const wrapper = mount(CommitControls, {
      props: {
        canGenerate: false,
        isGenerating: false,
        canCommit: false,
        canPush: true
      }
    })

    await wrapper.find('.btn-push').trigger('click')
    expect(wrapper.emitted('push')).toBeTruthy()
  })

  it('should disable buttons when props indicate', () => {
    const wrapper = mount(CommitControls, {
      props: {
        canGenerate: false,
        isGenerating: false,
        canCommit: false,
        canPush: false
      }
    })

    const buttons = wrapper.findAll('.btn')
    buttons.forEach(button => {
      expect(button.attributes('disabled')).toBeDefined()
    })
  })
})

describe('CommitMessage', () => {
  it('should display loading indicator when generating', () => {
    const wrapper = mount(CommitMessage, {
      props: {
        message: '',
        isGenerating: true
      }
    })

    expect(wrapper.find('.loading-indicator').exists()).toBe(true)
    expect(wrapper.text()).toContain('正在生成')
  })

  it('should display error message when error exists', () => {
    const errorMessage = '生成失败'
    const wrapper = mount(CommitMessage, {
      props: {
        message: '',
        isGenerating: false,
        error: errorMessage
      }
    })

    expect(wrapper.find('.error-message').exists()).toBe(true)
    expect(wrapper.text()).toContain(errorMessage)
  })

  it('should display textarea when not generating and no error', () => {
    const message = 'feat: add new feature'
    const wrapper = mount(CommitMessage, {
      props: {
        message,
        isGenerating: false,
        error: ''
      }
    })

    expect(wrapper.find('.message-textarea').exists()).toBe(true)
    const textarea = wrapper.find('.message-textarea') as any
    expect(textarea.element.value).toBe(message)
  })

  it('should display character count', () => {
    const message = 'test commit message'
    const wrapper = mount(CommitMessage, {
      props: {
        message,
        isGenerating: false
      }
    })

    expect(wrapper.find('.char-count').text()).toContain(`${message.length} 字符`)
  })

  it('should emit update:message when input changes', async () => {
    const wrapper = mount(CommitMessage, {
      props: {
        message: '',
        isGenerating: false
      }
    })

    const textarea = wrapper.find('.message-textarea')
    await textarea.setValue('new message')
    await textarea.trigger('input')

    expect(wrapper.emitted('update:message')).toBeTruthy()
    expect(wrapper.emitted('update:message')![0]).toEqual(['new message'])
  })

  it('should emit clear message when clear button clicked', async () => {
    const wrapper = mount(CommitMessage, {
      props: {
        message: 'some message',
        isGenerating: false
      }
    })

    const clearBtn = wrapper.find('.clear-btn')
    await clearBtn.trigger('click')

    expect(wrapper.emitted('update:message')).toBeTruthy()
    expect(wrapper.emitted('update:message')![0]).toEqual([''])
  })
})
