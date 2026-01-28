import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { EventsOn } from '../wailsjs/runtime'
import { useStartupStore } from './stores/startupStore'
import type { StartupProgress } from './stores/startupStore'
import './style.css'
import App from './App.vue'

const pinia = createPinia()
const app = createApp(App)
app.use(pinia)

// 监听启动进度事件
EventsOn('startup-progress', (data: StartupProgress) => {
  const startupStore = useStartupStore()
  startupStore.updateProgress(data)
})

// 监听启动完成事件
EventsOn('startup-complete', () => {
  const startupStore = useStartupStore()
  startupStore.complete()
})

app.mount('#app')
