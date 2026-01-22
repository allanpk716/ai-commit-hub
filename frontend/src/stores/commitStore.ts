import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProjectStatus } from '../types'
import { GetProjectStatus, GenerateCommit } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useCommitStore = defineStore('commit', () => {
  const selectedProjectPath = ref<string>('')
  const projectStatus = ref<ProjectStatus | null>(null)
  const isGenerating = ref(false)
  const streamingMessage = ref('')
  const generatedMessage = ref('')
  const error = ref<string | null>(null)

  // Provider settings
  const provider = ref('openai')
  const language = ref('zh')

  async function loadProjectStatus(path: string) {
    selectedProjectPath.value = path
    error.value = null

    try {
      const result = await GetProjectStatus(path)
      projectStatus.value = result as ProjectStatus
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载项目状态失败'
      error.value = message
    }
  }

  async function generateCommit() {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    isGenerating.value = true
    streamingMessage.value = ''
    generatedMessage.value = ''
    error.value = null

    try {
      await GenerateCommit(
        selectedProjectPath.value,
        provider.value,
        language.value
      )
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '生成失败'
      error.value = message
      isGenerating.value = false
    }
  }

  function clearMessage() {
    streamingMessage.value = ''
    generatedMessage.value = ''
  }

  // Setup event listeners
  EventsOn('commit-delta', (delta: string) => {
    streamingMessage.value += delta
  })

  EventsOn('commit-complete', (message: string) => {
    generatedMessage.value = message
    streamingMessage.value = message
    isGenerating.value = false
  })

  EventsOn('commit-error', (err: string) => {
    error.value = err
    isGenerating.value = false
  })

  return {
    selectedProjectPath,
    projectStatus,
    isGenerating,
    streamingMessage,
    generatedMessage,
    error,
    provider,
    language,
    loadProjectStatus,
    generateCommit,
    clearMessage
  }
})
