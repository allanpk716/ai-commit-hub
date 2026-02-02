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
  const downloadSpeed = ref(0)
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

  function skipVersion(version: string) {
    skippedVersion.value = version
    hasUpdate.value = false
  }

  function resetUpdateState() {
    hasUpdate.value = false
    updateInfo.value = null
    isDownloading.value = false
    downloadProgress.value = 0
    isReadyToInstall.value = false
  }

  // 监听后端事件
  EventsOn('update-available', (info: models.UpdateInfo) => {
    console.log('收到更新可用事件:', info)
    updateInfo.value = info
    hasUpdate.value = info.hasUpdate
  })

  EventsOn('download-progress', (data: { percentage: number; speed: number }) => {
    downloadProgress.value = data.percentage
    downloadSpeed.value = data.speed
    isDownloading.value = true
  })

  EventsOn('download-complete', () => {
    isDownloading.value = false
    isReadyToInstall.value = true
  })

  return {
    hasUpdate,
    updateInfo,
    isChecking,
    isDownloading,
    downloadProgress,
    downloadSpeed,
    isReadyToInstall,
    skippedVersion,
    displayVersion,
    releaseNotes,
    formattedSize,
    checkForUpdates,
    skipVersion,
    resetUpdateState
  }
})
