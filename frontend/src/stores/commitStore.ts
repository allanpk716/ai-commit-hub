import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProjectStatus, ProjectAIConfig, ProviderInfo, StagingStatus, StagedFile, UntrackedFile } from '../types'
import {
  GetProjectStatus,
  GenerateCommit,
  GetProjectAIConfig,
  UpdateProjectAIConfig,
  ValidateProjectConfig,
  ConfirmResetProjectConfig,
  GetConfiguredProviders,
  GetStagingStatus,
  GetFileDiff,
  GetUntrackedFileContent,
  StageFile,
  StageAllFiles,
  UnstageFile,
  UnstageAllFiles,
  GetUntrackedFiles,
  StageFiles,
  AddToGitIgnore,
  DiscardFileChanges
} from '../../wailsjs/go/main/App'

export const useCommitStore = defineStore('commit', () => {
  const selectedProjectPath = ref<string>('')
  const selectedProjectId = ref<number>(0)
  const projectStatus = ref<ProjectStatus | null>(null)
  const isGenerating = ref(false)
  const streamingMessage = ref('')
  const generatedMessage = ref('')
  const error = ref<string | null>(null)

  // Provider 列表
  const availableProviders = ref<ProviderInfo[]>([])

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

  // 暂存区状态
  const stagingStatus = ref<StagingStatus | null>(null)
  const isLoadingStaging = ref(false)
  const selectedFileDiff = ref<{
    filePath: string
    diff: string
  } | null>(null)

  // 文件选择状态
  const selectedStagedFiles = ref<Set<string>>(new Set())
  const selectedUnstagedFiles = ref<Set<string>>(new Set())
  const selectedFile = ref<StagedFile | null>(null)
  const fileDiff = ref<string | null>(null)
  const isLoadingDiff = ref(false)

  // 未跟踪文件状态
  const untrackedFiles = ref<UntrackedFile[]>([])
  const untrackedFilesLoading = ref(false)

  async function loadProjectStatus(path: string) {
    selectedProjectPath.value = path
    error.value = null

    try {
      const result = await GetProjectStatus(path)
      console.log('[loadProjectStatus] 后端返回:', result)
      console.log('[loadProjectStatus] has_staged 值:', (result as any).has_staged)
      projectStatus.value = result as ProjectStatus
      console.log('[loadProjectStatus] projectStatus.value.has_staged:', projectStatus.value?.has_staged)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载项目状态失败'
      error.value = message
    }
  }

  // 添加一个计算属性来判断是否有暂存文件
  const hasStagedFiles = computed(() => {
    console.log('[hasStagedFiles] 计算中...', {
      stagingStatus: stagingStatus.value,
      stagedLength: stagingStatus.value?.staged?.length ?? 0,
      projectStatus: projectStatus.value
    })
    // 优先使用 stagingStatus.staged 的长度
    if (stagingStatus.value && stagingStatus.value.staged && stagingStatus.value.staged.length > 0) {
      return true
    }
    // 降级使用 projectStatus.has_staged
    return projectStatus.value?.has_staged ?? false
  })

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

  async function loadAvailableProviders() {
    try {
      const result = await GetConfiguredProviders()
      availableProviders.value = result as ProviderInfo[]
    } catch (e) {
      console.error('加载 provider 列表失败:', e)
      // 失败时使用空数组，避免界面崩溃
      availableProviders.value = []
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

  // 事件处理函数（供组件调用）
  function handleDelta(delta: string) {
    console.log('[commit-delta] 收到 delta:', delta.substring(0, 50) + '...')
    streamingMessage.value += delta
    console.log('[commit-delta] 当前 streamingMessage 长度:', streamingMessage.value.length)
  }

  function handleComplete(message: string) {
    console.log('[commit-complete] 收到完整消息:', message.substring(0, 50) + '...')
    generatedMessage.value = message
    streamingMessage.value = message
    isGenerating.value = false
  }

  function handleError(err: string) {
    console.log('[commit-error] 收到错误:', err)
    error.value = err
    isGenerating.value = false
  }

  // 暂存区管理方法
  async function loadStagingStatus(path: string) {
    if (!path) {
      stagingStatus.value = null
      return
    }

    isLoadingStaging.value = true
    error.value = null

    try {
      console.log('[loadStagingStatus] 开始加载暂存状态, path:', path)
      const result = await GetStagingStatus(path)
      console.log('[loadStagingStatus] 后端返回:', result)
      console.log('[loadStagingStatus] staged 数量:', (result as any)?.staged?.length)
      console.log('[loadStagingStatus] unstaged 数量:', (result as any)?.unstaged?.length)
      stagingStatus.value = result as StagingStatus
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载暂存区状态失败'
      error.value = message
      stagingStatus.value = null
    } finally {
      isLoadingStaging.value = false
    }
  }

  async function loadFileDiff(filePath: string, staged: boolean) {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    error.value = null

    try {
      const diff = await GetFileDiff(selectedProjectPath.value, filePath, staged)
      selectedFileDiff.value = {
        filePath,
        diff: diff as string
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载文件差异失败'
      error.value = message
      selectedFileDiff.value = null
    }
  }

  async function stageFile(filePath: string) {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      console.log('[stageFile] 开始暂存文件:', filePath)
      await StageFile(selectedProjectPath.value, filePath)
      console.log('[stageFile] 暂存文件成功，开始重新加载状态')
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      console.log('[stageFile] loadStagingStatus 完成')
      await loadProjectStatus(selectedProjectPath.value)
      console.log('[stageFile] loadProjectStatus 完成')
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '暂存文件失败'
      error.value = message
      throw e
    }
  }

  async function stageAllFiles() {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      await StageAllFiles(selectedProjectPath.value)
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '暂存所有文件失败'
      error.value = message
      throw e
    }
  }

  async function unstageFile(filePath: string) {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      await UnstageFile(selectedProjectPath.value, filePath)
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
      // 刷新未跟踪文件列表
      await loadUntrackedFiles(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '取消暂存文件失败'
      error.value = message
      throw e
    }
  }

  async function discardFileChanges(filePath: string) {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      await DiscardFileChanges(selectedProjectPath.value, filePath)
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
      // 刷新未跟踪文件列表
      await loadUntrackedFiles(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '还原文件失败'
      error.value = message
      throw e
    }
  }

  async function unstageAllFiles() {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      await UnstageAllFiles(selectedProjectPath.value)
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '取消暂存所有文件失败'
      error.value = message
      throw e
    }
  }

  function clearFileDiff() {
    selectedFileDiff.value = null
  }

  // 选择文件
  function selectFile(file: StagedFile) {
    selectedFile.value = file

    // 如果文件路径为空，清空 diff 而不是尝试加载
    // 修复：避免在关闭 diff 按钮时加载空路径导致的错误
    if (!file.path) {
      selectedFileDiff.value = null
      return
    }

    // 判断文件是已暂存还是未暂存
    const isStaged = stagingStatus.value?.staged?.some((f: StagedFile) => f.path === file.path) ?? false
    // 加载文件差异
    loadFileDiff(file.path, isStaged)
  }

  // 选择未跟踪文件（专门用于未跟踪文件，避免调用 loadFileDiff）
  async function selectUntrackedFile(file: StagedFile) {
    console.log('[selectUntrackedFile] 选择未跟踪文件:', file.path)

    if (!file.path) {
      selectedFile.value = null
      selectedFileDiff.value = null
      return
    }

    // 先设置文件和清空旧的 diff，显示加载状态
    selectedFile.value = file
    selectedFileDiff.value = null
    console.log('[selectUntrackedFile] 已清空旧的 diff，准备加载新内容')

    // 加载文件内容（不是 diff）
    await loadUntrackedFileContent(file.path)
    console.log('[selectUntrackedFile] 文件内容加载完成')
  }

  // 加载未跟踪文件内容
  async function loadUntrackedFileContent(filePath: string) {
    if (!selectedProjectPath.value) {
      error.value = '请先选择项目'
      return
    }

    try {
      console.log('[loadUntrackedFileContent] 开始加载未跟踪文件:', filePath)
      console.log('[loadUntrackedFileContent] 项目路径:', selectedProjectPath.value)
      const result = await GetUntrackedFileContent(selectedProjectPath.value, filePath)
      console.log('[loadUntrackedFileContent] API 返回结果:', result)
      console.log('[loadUntrackedFileContent] IsBinary:', result.IsBinary)
      console.log('[loadUntrackedFileContent] Content 长度:', result.Content?.length)

      if (result.IsBinary) {
        // 二进制文件：显示占位提示
        console.log('[loadUntrackedFileContent] 检测到二进制文件')
        selectedFileDiff.value = {
          filePath,
          diff: '[二进制文件，无法预览内容]'
        }
      } else {
        // 文本文件：设置 diff
        console.log('[loadUntrackedFileContent] 设置文件内容到 selectedFileDiff')
        selectedFileDiff.value = {
          filePath,
          diff: result.Content
        }
        console.log('[loadUntrackedFileContent] selectedFileDiff 已设置:', selectedFileDiff.value)
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '读取文件内容失败'
      console.error('[loadUntrackedFileContent] 加载失败:', e)
      error.value = message
      selectedFileDiff.value = null
    }
  }

  // 批量暂存选中的文件
  async function stageSelectedFiles() {
    if (!selectedProjectPath.value || selectedUnstagedFiles.value.size === 0) {
      return
    }

    try {
      for (const filePath of selectedUnstagedFiles.value) {
        await StageFile(selectedProjectPath.value, filePath)
      }
      // 清空选择
      selectedUnstagedFiles.value.clear()
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '批量暂存文件失败'
      error.value = message
      throw e
    }
  }

  // 批量取消暂存选中的文件
  async function unstageSelectedFiles() {
    if (!selectedProjectPath.value || selectedStagedFiles.value.size === 0) {
      return
    }

    try {
      for (const filePath of selectedStagedFiles.value) {
        await UnstageFile(selectedProjectPath.value, filePath)
      }
      // 清空选择
      selectedStagedFiles.value.clear()
      // 重新加载暂存区状态和项目状态
      await loadStagingStatus(selectedProjectPath.value)
      await loadProjectStatus(selectedProjectPath.value)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '批量取消暂存文件失败'
      error.value = message
      throw e
    }
  }

  // 切换文件选择状态
  function toggleFileSelection(filePath: string, type: 'staged' | 'unstaged') {
    if (type === 'staged') {
      if (selectedStagedFiles.value.has(filePath)) {
        selectedStagedFiles.value.delete(filePath)
      } else {
        selectedStagedFiles.value.add(filePath)
      }
    } else {
      if (selectedUnstagedFiles.value.has(filePath)) {
        selectedUnstagedFiles.value.delete(filePath)
      } else {
        selectedUnstagedFiles.value.add(filePath)
      }
    }
  }

  // 清空暂存区选择状态
  function clearStagingState() {
    selectedStagedFiles.value.clear()
    selectedUnstagedFiles.value.clear()
    selectedFile.value = null
    fileDiff.value = null
    isLoadingDiff.value = false
  }

  // 加载未跟踪文件列表
  async function loadUntrackedFiles(projectPath: string) {
    untrackedFilesLoading.value = true
    try {
      const files = await GetUntrackedFiles(projectPath)
      untrackedFiles.value = files
    } catch (e) {
      console.error('加载未跟踪文件失败:', e)
      untrackedFiles.value = []
    } finally {
      untrackedFilesLoading.value = false
    }
  }

  // 批量暂存文件
  async function stageFiles(files: string[]) {
    if (!selectedProjectPath.value) return

    try {
      await StageFiles(selectedProjectPath.value, files)
      // 刷新暂存区和未跟踪文件
      await Promise.all([
        loadStagingStatus(selectedProjectPath.value),
        loadUntrackedFiles(selectedProjectPath.value)
      ])
    } catch (e) {
      console.error('添加到暂存区失败:', e)
      throw e
    }
  }

  // 添加到 .gitignore
  async function addToGitIgnore(file: string, mode: 'exact' | 'extension' | 'directory') {
    if (!selectedProjectPath.value) return

    try {
      await AddToGitIgnore(selectedProjectPath.value, file, mode)
      // 刷新暂存区和未跟踪文件
      await Promise.all([
        loadStagingStatus(selectedProjectPath.value),
        loadUntrackedFiles(selectedProjectPath.value)
      ])
    } catch (e) {
      console.error('添加到排除列表失败:', e)
      throw e
    }
  }

  return {
    selectedProjectPath,
    selectedProjectId,
    projectStatus,
    isGenerating,
    streamingMessage,
    generatedMessage,
    error,
    availableProviders,
    provider,
    language,
    isDefaultConfig,
    isSavingConfig,
    configValidation,
    stagingStatus,
    isLoadingStaging,
    selectedFileDiff,
    selectedStagedFiles,
    selectedUnstagedFiles,
    selectedFile,
    fileDiff,
    isLoadingDiff,
    hasStagedFiles,
    untrackedFiles,
    untrackedFilesLoading,
    loadProjectStatus,
    loadProjectAIConfig,
    loadAvailableProviders,
    saveProjectConfig,
    confirmResetConfig,
    generateCommit,
    clearMessage,
    handleDelta,
    handleComplete,
    handleError,
    loadStagingStatus,
    loadFileDiff,
    stageFile,
    stageAllFiles,
    unstageFile,
    unstageAllFiles,
    discardFileChanges,
    clearFileDiff,
    selectFile,
    selectUntrackedFile,
    loadUntrackedFileContent,
    stageSelectedFiles,
    unstageSelectedFiles,
    toggleFileSelection,
    clearStagingState,
    loadUntrackedFiles,
    stageFiles,
    addToGitIgnore
  }
})
