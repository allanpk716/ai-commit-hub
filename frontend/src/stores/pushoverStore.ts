import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  GetPushoverHookStatus,
  InstallPushoverHook,
  SetPushoverNotificationMode,
  GetPushoverExtensionInfo,
  ClonePushoverExtension,
  UpdatePushoverExtension,
  CheckPushoverUpdates,
  UpdatePushoverHook,
  ToggleNotification,
  CheckPushoverConfig
} from '../../wailsjs/go/main/App'
import type {
  HookStatus,
  ExtensionInfo,
  InstallResult,
  NotificationMode,
  PushoverConfigStatus
} from '../types/pushover'
import { useStatusCache } from './statusCache'
import { useProjectStore } from './projectStore'
import type { ProjectStatusCache } from '../types/status'

export const usePushoverStore = defineStore('pushover', () => {
  // State
  const extensionInfo = ref<ExtensionInfo>({
    downloaded: false,
    path: '',
    version: '',
    current_version: '',
    latest_version: '',
    update_available: false
  })

  const loading = ref(false)
  const error = ref<string | null>(null)
  const configValid = ref(false)
  const updateCheckError = ref<string | null>(null) // 更新检查错误
  const isCheckingUpdate = ref(false) // 是否正在检查更新

  // Computed
  const isExtensionDownloaded = computed(() => extensionInfo.value.downloaded)
  const isUpdateAvailable = computed(() => extensionInfo.value.update_available)

  /**
   * 当前选中项目的 Pushover 状态（从 StatusCache 获取）
   */
  const currentProjectHookStatus = computed(() => {
    const projectStore = useProjectStore()
    const selectedPath = projectStore.selectedProject
    if (!selectedPath) return null

    const statusCache = useStatusCache()
    const cached = statusCache.getStatus(selectedPath)
    return cached?.pushoverStatus || null
  })

  /**
   * 批量获取所有项目的 Pushover 状态（从 StatusCache 获取）
   * 供 ProjectList 使用
   */
  const allProjectHookStatuses = computed(() => {
    const statusCache = useStatusCache()
    const statuses: Record<string, HookStatus> = {}
    for (const [path, cache] of Object.entries(statusCache.cache)) {
      if (cache.pushoverStatus) {
        statuses[path] = cache.pushoverStatus
      }
    }
    return statuses
  })

  // Actions

  /**
   * 克隆扩展仓库
   */
  async function cloneExtension() {
    loading.value = true
    error.value = null

    try {
      await ClonePushoverExtension()
      await checkExtensionStatus()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `克隆扩展失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新扩展
   */
  async function updateExtension() {
    loading.value = true
    error.value = null

    try {
      await UpdatePushoverExtension()
      await checkExtensionStatus()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `更新扩展失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取项目 Hook 状态
   */
  async function getProjectHookStatus(projectPath: string): Promise<HookStatus | null> {
    console.log('[DEBUG pushoverStore] getProjectHookStatus called for:', projectPath)
    loading.value = true
    error.value = null

    try {
      // 不再存储到 Map，直接刷新 StatusCache
      const { useStatusCache } = await import('./statusCache')
      const statusCache = useStatusCache()
      await statusCache.refresh(projectPath, { force: true })

      const status = statusCache.getPushoverStatus(projectPath)
      console.log('[DEBUG pushoverStore] Status from cache:', status)
      return status
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      console.error('[DEBUG pushoverStore] Error:', e)
      error.value = `获取 Hook 状态失败: ${message}`
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 安装 Hook
   */
  async function installHook(projectPath: string, force = false): Promise<InstallResult> {
    loading.value = true
    error.value = null

    try {
      const result = await InstallPushoverHook(projectPath, force)
      if (result && result.success) {
        // 刷新项目状态
        await getProjectHookStatus(projectPath)
      }
      return result || { success: false, message: '安装失败' }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `安装 Hook 失败: ${message}`
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  /**
   * 设置通知模式
   */
  async function setNotificationMode(projectPath: string, mode: NotificationMode) {
    loading.value = true
    error.value = null

    try {
      await SetPushoverNotificationMode(projectPath, mode)
      // 刷新项目状态
      await getProjectHookStatus(projectPath)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `设置通知模式失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 切换通知状态（Pushover 或 Windows 通知）
   * @param projectPath 项目路径
   * @param notificationType 通知类型：'pushover' 或 'windows'
   */
  async function toggleNotification(projectPath: string, notificationType: 'pushover' | 'windows') {
    loading.value = true
    error.value = null

    try {
      await ToggleNotification(projectPath, notificationType)
      // 刷新项目状态
      await getProjectHookStatus(projectPath)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `切换通知状态失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新项目的 Hook
   */
  async function updateHook(projectPath: string): Promise<InstallResult> {
    loading.value = true
    error.value = null

    try {
      const result = await UpdatePushoverHook(projectPath)
      if (result && result.success) {
        // 刷新项目状态
        await getProjectHookStatus(projectPath)
      }
      return result || { success: false, message: '更新失败' }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `更新 Hook 失败: ${message}`
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  /**
   * 检查扩展更新
   */
  async function checkExtensionStatus() {
    loading.value = true
    error.value = null

    try {
      const info = await GetPushoverExtensionInfo()
      if (info) {
        extensionInfo.value = info
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `获取扩展信息失败: ${message}`
    } finally {
      loading.value = false
    }
  }

  /**
   * 检查项目的 Hook 更新
   */
  async function checkForUpdates(projectPath: string) {
    loading.value = true
    error.value = null

    try {
      const result = await CheckPushoverUpdates(projectPath)
      return {
        updateAvailable: result.update_available as boolean,
        currentVersion: result.current_version as string,
        latestVersion: result.latest_version as string
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `检查更新失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 检查扩展自身更新（而非项目 Hook）
   */
  async function checkForExtensionUpdates() {
    updateCheckError.value = null
    isCheckingUpdate.value = true

    try {
      const { CheckPushoverExtensionUpdates } = await import('../../wailsjs/go/main/App')
      const result = await CheckPushoverExtensionUpdates()
      extensionInfo.value.update_available = result.needs_update as boolean
      extensionInfo.value.current_version = result.current_version as string
      extensionInfo.value.latest_version = result.latest_version as string
      return {
        updateAvailable: result.needs_update as boolean,
        currentVersion: result.current_version as string,
        latestVersion: result.latest_version as string
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      updateCheckError.value = `检查更新失败: ${message}`
      throw e
    } finally {
      isCheckingUpdate.value = false
    }
  }

  /**
   * 重新下载扩展（删除并克隆）
   */
  async function recloneExtension() {
    loading.value = true
    error.value = null

    try {
      const { ReclonePushoverExtension } = await import('../../wailsjs/go/main/App')
      await ReclonePushoverExtension()
      await checkExtensionStatus()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `重新下载扩展失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 清除缓存
   */
  function clearCache() {
    // StatusCache 会自动管理缓存
  }

  /**
   * 检查 Pushover 配置状态
   * 检查环境变量 PUSHOVER_TOKEN 和 PUSHOVER_USER 是否已设置
   * @returns 配置状态对象
   */
  async function checkPushoverConfig(): Promise<PushoverConfigStatus> {
    try {
      const result = await CheckPushoverConfig()
      configValid.value = result.valid as boolean
      return {
        valid: result.valid as boolean,
        tokenSet: result.token_set as boolean,
        userSet: result.user_set as boolean
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `检查 Pushover 配置失败: ${message}`
      configValid.value = false
      throw e
    }
  }

  return {
    // State
    extensionInfo,
    loading,
    error,
    configValid,
    updateCheckError,
    isCheckingUpdate,

    // Computed
    isExtensionDownloaded,
    isUpdateAvailable,
    currentProjectHookStatus,
    allProjectHookStatuses,

    // Actions
    checkExtensionStatus,
    cloneExtension,
    updateExtension,
    getProjectHookStatus,
    installHook,
    setNotificationMode,
    toggleNotification,
    clearCache,
    checkForUpdates,
    updateHook,
    checkForExtensionUpdates,
    recloneExtension,
    checkPushoverConfig
  }
})
