<template>
  <div v-if="startupStore.isVisible" class="splash-screen">
    <div class="splash-content">
      <!-- Logo -->
      <div class="app-logo">
        <span class="logo-icon">🚀</span>
      </div>

      <!-- Title -->
      <h1 class="app-title">AI Commit Hub</h1>
      <p class="app-version">v1.0.0</p>

      <!-- Progress Bar -->
      <div class="progress-container">
        <div class="progress-bar">
          <div
            class="progress-fill"
            :style="{ width: startupStore.progress.percent + '%' }"
          ></div>
        </div>
        <span class="progress-text">{{ startupStore.progress.percent }}%</span>
      </div>

      <!-- Status Message -->
      <p class="status-message">{{ startupStore.progress.message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'
import { useStartupStore } from '../stores/startupStore'
import type { StartupProgress } from '../stores/startupStore'

const startupStore = useStartupStore()

// 组件挂载时设置事件监听器
onMounted(() => {
  // 监听启动进度事件
  EventsOn('startup-progress', (data: StartupProgress) => {
    startupStore.updateProgress(data)
  })

  // 监听启动完成事件
  EventsOn('startup-complete', () => {
    startupStore.complete()
  })
})

// 组件卸载时清理事件监听器
onUnmounted(() => {
  // 清理 Wails 事件监听器
  EventsOff('startup-progress')
  EventsOff('startup-complete')
})
</script>

<style scoped>
.splash-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #1b263b 0%, #0d1b2a 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fade-in 0.3s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.splash-content {
  text-align: center;
  color: white;
}

.app-logo {
  margin-bottom: 2rem;
}

.logo-icon {
  font-size: 80px;
  display: inline-block;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.app-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
  background: linear-gradient(135deg, #06b6d4, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.app-version {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin: 0 0 3rem 0;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  max-width: 300px;
  margin: 0 auto 1.5rem;
}

.progress-bar {
  flex: 1;
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #06b6d4, #8b5cf6);
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.8);
  min-width: 40px;
  text-align: right;
}

.status-message {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  margin: 0;
}
</style>
