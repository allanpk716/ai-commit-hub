import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { models } from '../../wailsjs/go/models'

export const useUpdateStore = defineStore('update', () => {
  // State
  const hasUpdate = ref(false)
  const updateInfo = ref<models.UpdateInfo | null>(null)
  const isChecking = ref(false)
  const isDownloading = ref(false)
  const downloadProgress = ref(0)
  const downloadedSize = ref(0)
  const totalSize = ref(0)
  const downloadSpeed = ref(0)
  const downloadETA = ref('')
  const canCancel = ref(false)
  const isReadyToInstall = ref(false)
  const skippedVersion = ref<string | null>(null)

  // Computed
  const displayVersion = computed(() => {
    return updateInfo.value?.latestVersion || ''
  })

  const releaseNotes = computed(() => {
    return updateInfo.value?.releaseNotes || ''
  })

  const formattedSize = computed(() => {
    if (!updateInfo.value?.size) return '未知大小'
    const bytes = updateInfo.value.size
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  })

  // Actions
  async function checkForUpdates() {
    isChecking.value = true
    try {
      // 这里调用后端 API
      const { CheckForUpdates } = await import('../../wailsjs/go/main/App')
      const info = await CheckForUpdates()
      updateInfo.value = info
      hasUpdate.value = info.hasUpdate
      return info
    } catch (error) {
      console.error('检查更新失败:', error)
      throw error
    } finally {
      isChecking.value = false
    }
  }

  async function downloadUpdate(url: string, filename: string, proxyURL: string = '') {
    if (!updateInfo.value) {
      throw new Error('没有可用的更新信息')
    }

    isDownloading.value = true
    downloadProgress.value = 0
    canCancel.value = true

    try {
      const { DownloadUpdate } = await import('../../wailsjs/go/main/App')
      await DownloadUpdate(url, filename, proxyURL)
    } catch (error) {
      console.error('下载更新失败:', error)
      isDownloading.value = false
      canCancel.value = false
      throw error
    }
  }

  async function cancelDownload() {
    if (!updateInfo.value) {
      return
    }

    try {
      const { CancelDownload } = await import('../../wailsjs/go/main/App')
      await CancelDownload(updateInfo.value.assetName)
    } catch (error) {
      console.error('取消下载失败:', error)
      throw error
    } finally {
      isDownloading.value = false
      canCancel.value = false
      downloadProgress.value = 0
    }
  }

  async function installUpdate() {
    if (!updateInfo.value) {
      throw new Error('没有可用的更新信息')
    }

    isDownloading.value = true
    downloadProgress.value = 0

    try {
      const { InstallUpdate } = await import('../../wailsjs/go/main/App')
      await InstallUpdate(updateInfo.value.downloadURL, updateInfo.value.assetName)
    } catch (error) {
      console.error('安装更新失败:', error)
      isDownloading.value = false
      throw error
    }
    // 注意：成功后程序会退出，不需要重置状态
  }

  function skipVersion(version: string) {
    skippedVersion.value = version
    hasUpdate.value = false
  }

  function resetUpdateState() {
    hasUpdate.value = false
    updateInfo.value = null
    isDownloading.value = false
    downloadProgress.value = 0
    downloadedSize.value = 0
    totalSize.value = 0
    downloadSpeed.value = 0
    downloadETA.value = ''
    canCancel.value = false
    isReadyToInstall.value = false
  }

  // 工具方法：格式化字节大小
  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  // 工具方法：格式化速度
  function formatSpeed(bytesPerSecond: number): string {
    return formatBytes(bytesPerSecond) + '/s'
  }

  // 工具方法：计算剩余时间
  function calculateETA(downloaded: number, total: number, speed: number): string {
    if (speed <= 0) return '计算中...'
    const remaining = total - downloaded
    const seconds = Math.floor(remaining / speed)
    if (seconds < 60) return `${seconds}秒`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
    return `${Math.floor(seconds / 3600)}小时${Math.floor((seconds % 3600) / 60)}分`
  }

  // 监听后端事件
  EventsOn('update-available', (data: { hasUpdate: boolean; info: models.UpdateInfo }) => {
    console.log('收到更新可用事件:', data)
    updateInfo.value = data.info
    hasUpdate.value = data.hasUpdate
  })

  EventsOn('download-progress', (data: {
    percentage: number
    downloaded: number
    total: number
    speed: number
    eta: string
    url: string
  }) => {
    console.log('收到下载进度事件:', data)
    downloadProgress.value = data.percentage
    downloadedSize.value = data.downloaded
    totalSize.value = data.total
    downloadSpeed.value = data.speed
    downloadETA.value = data.eta
    isDownloading.value = true
  })

  EventsOn('download-complete', () => {
    console.log('收到下载完成事件')
    isDownloading.value = false
    isReadyToInstall.value = true
    canCancel.value = false
  })

  return {
    hasUpdate,
    updateInfo,
    isChecking,
    isDownloading,
    downloadProgress,
    downloadedSize,
    totalSize,
    downloadSpeed,
    downloadETA,
    canCancel,
    isReadyToInstall,
    skippedVersion,
    displayVersion,
    releaseNotes,
    formattedSize,
    checkForUpdates,
    downloadUpdate,
    cancelDownload,
    installUpdate,
    skipVersion,
    resetUpdateState,
    formatBytes,
    formatSpeed,
    calculateETA
  }
})
