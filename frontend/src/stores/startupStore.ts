import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface StartupProgress {
  stage: string
  percent: number
  message: string
}

export const useStartupStore = defineStore('startup', () => {
  const isVisible = ref(true)
  const isCompleted = ref(false)
  const progress = ref<StartupProgress>({
    stage: 'initializing',
    percent: 0,
    message: '正在初始化...'
  })

  function updateProgress(data: StartupProgress) {
    progress.value = data
  }

  function complete() {
    // 防止重复调用
    if (isCompleted.value) {
      console.warn('启动已完成，忽略重复的 complete() 调用')
      return
    }

    isCompleted.value = true
    progress.value.percent = 100
    progress.value.message = '完成'
    setTimeout(() => {
      isVisible.value = false
    }, 500)
  }

  return {
    isVisible,
    progress,
    updateProgress,
    complete
  }
})
