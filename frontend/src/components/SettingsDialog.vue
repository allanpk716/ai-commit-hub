<template>
  <div v-if="open" class="dialog-overlay" @click.self="close">
    <div class="dialog">
      <div class="dialog-header">
        <h2>设置</h2>
        <button class="close-btn" @click="close">×</button>
      </div>

      <div class="dialog-body">
        <!-- 配置管理 -->
        <section class="config-section">
          <h3>配置管理</h3>
          <button class="btn btn-secondary" @click="openConfigFolder">
            <span>📁</span>
            <span>打开配置文件夹</span>
          </button>
        </section>
      </div>

      <div class="dialog-footer">
        <button class="btn btn-secondary" @click="close">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { OpenConfigFolder } from '../../wailsjs/go/main/App'

interface Props {
  modelValue: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const open = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

function close() {
  open.value = false
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    console.error('打开配置文件夹失败:', message)
  }
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  width: 600px;
  max-width: 90vw;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg) var(--space-xl);
  border-bottom: 1px solid var(--border-default);
}

.dialog-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-normal);
}

.close-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.dialog-body {
  padding: var(--space-xl);
  overflow-y: auto;
  flex: 1;
}

section {
  margin-bottom: var(--space-xl);
}

section h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.config-section button {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-normal);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.dialog-footer {
  padding: var(--space-lg) var(--space-xl);
  border-top: 1px solid var(--border-default);
  display: flex;
  justify-content: flex-end;
}
</style>
