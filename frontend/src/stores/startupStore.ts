import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface StartupProgress {
  stage: string
  percent: number
  message: string
}

export const useStartupStore = defineStore('startup', () => {
  const isVisible = ref(true)
  const progress = ref<StartupProgress>({
    stage: 'initializing',
    percent: 0,
    message: '正在初始化...'
  })

  function updateProgress(data: StartupProgress) {
    progress.value = data
  }

  function complete() {
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
