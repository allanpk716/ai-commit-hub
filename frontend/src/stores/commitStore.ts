import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProjectStatus, ProjectAIConfig } from '../types'
import {
  GetProjectStatus,
  GenerateCommit,
  GetProjectAIConfig,
  UpdateProjectAIConfig,
  ValidateProjectConfig,
  ConfirmResetProjectConfig
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useCommitStore = defineStore('commit', () => {
  const selectedProjectPath = ref<string>('')
  const selectedProjectId = ref<number>(0)
  const projectStatus = ref<ProjectStatus | null>(null)
  const isGenerating = ref(false)
  const streamingMessage = ref('')
  const generatedMessage = ref('')
  const error = ref<string | null>(null)

  // Provider settings
  const provider = ref('openai')
  const language = ref('zh')
  const isDefaultConfig = ref(true)  // 标记是否使用默认配置
  const isSavingConfig = ref(false)  // 保存状态

  // 配置验证状态
  const configValidation = ref<{
    valid: boolean
    resetFields: string[]
    suggestedConfig?: ProjectAIConfig
  } | null>(null)

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

  async function loadProjectAIConfig(projectId: number) {
    selectedProjectId.value = projectId

    try {
      const config = await GetProjectAIConfig(projectId)
      provider.value = config.Provider
      language.value = config.Language
      isDefaultConfig.value = config.IsDefault

      // 验证配置
      const result = await ValidateProjectConfig(projectId) as any
      if (result && result.length === 3) {
        const [valid, resetFields, suggestedConfig] = result
        if (!valid && resetFields.length > 0) {
          configValidation.value = {
            valid: false,
            resetFields,
            suggestedConfig: {
              provider: suggestedConfig.Provider,
              language: suggestedConfig.Language,
              isDefault: suggestedConfig.IsDefault
            }
          }
        } else {
          configValidation.value = null
        }
      }
    } catch (e: unknown) {
      console.error('加载项目配置失败:', e)
      // 失败时使用默认配置
      provider.value = 'openai'
      language.value = 'zh'
      isDefaultConfig.value = true
    }
  }

  async function saveProjectConfig(projectId: number) {
    if (isSavingConfig.value) {
      return
    }

    isSavingConfig.value = true

    try {
      await UpdateProjectAIConfig(
        projectId,
        isDefaultConfig.value ? '' : provider.value,
        isDefaultConfig.value ? '' : language.value,
        '',
        isDefaultConfig.value
      )
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '保存配置失败'
      error.value = message
      throw e
    } finally {
      isSavingConfig.value = false
    }
  }

  async function confirmResetConfig(projectId: number) {
    try {
      await ConfirmResetProjectConfig(projectId)

      // 重新加载配置
      await loadProjectAIConfig(projectId)

      configValidation.value = null
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '重置配置失败'
      error.value = message
      throw e
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
    selectedProjectId,
    projectStatus,
    isGenerating,
    streamingMessage,
    generatedMessage,
    error,
    provider,
    language,
    isDefaultConfig,
    isSavingConfig,
    configValidation,
    loadProjectStatus,
    loadProjectAIConfig,
    saveProjectConfig,
    confirmResetConfig,
    generateCommit,
    clearMessage
  }
})
