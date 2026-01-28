import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { useStartupStore } from './stores/startupStore'
import './style.css'
import App from './App.vue'

const pinia = createPinia()
const app = createApp(App)
app.use(pinia)

app.mount('#app')

// 添加 30 秒超时保护，防止启动画面卡住
const startupStore = useStartupStore()
const startupTimeout = setTimeout(() => {
  if (startupStore.isVisible) {
    console.warn('启动超时，强制隐藏启动画面')
    startupStore.complete()
  }
}, 30000)

// 启动成功后清除超时定时器
const checkComplete = setInterval(() => {
  if (!startupStore.isVisible) {
    clearTimeout(startupTimeout)
    clearInterval(checkComplete)
  }
}, 1000)
