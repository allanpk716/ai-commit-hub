<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="exclude-dialog">
      <div class="dialog-header">
        <h3>添加到排除列表</h3>
        <button @click="close" class="close-btn">×</button>
      </div>

      <div class="dialog-body">
        <label class="input-label">忽略文件名或模式:</label>
        <input v-model="pattern" class="pattern-input" />

        <div class="radio-group">
          <label class="radio-option">
            <input type="radio" value="exact" v-model="mode" />
            <span>忽略精确的文件名</span>
          </label>

          <label class="radio-option">
            <input type="radio" value="extension" v-model="mode" />
            <span>忽略所有文件的扩展名 ({{ extension }})</span>
          </label>

          <label class="radio-option">
            <input type="radio" value="directory" v-model="mode" :disabled="!hasDirectory" />
            <span>忽略下列所有:</span>
          </label>

          <select
            v-if="mode === 'directory'"
            v-model="selectedDirectory"
            class="directory-select"
            :disabled="!hasDirectory"
          >
            <option v-for="opt in directoryOptions" :key="opt.pattern" :value="opt.pattern">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="close" class="btn-secondary">取消</button>
        <button @click="confirm" class="btn-primary">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { DirectoryOption, ExcludeMode } from '../types'
import { GetDirectoryOptions } from '../../wailsjs/go/main/App'

const props = defineProps<{
  visible: boolean
  filePath: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm', mode: ExcludeMode, pattern: string): void
}>()

const pattern = ref(props.filePath)
const mode = ref<ExcludeMode>('exact')
const selectedDirectory = ref('')
const directoryOptions = ref<DirectoryOption[]>([])

const extension = computed(() => {
  const ext = props.filePath.split('.').pop()
  return ext ? `.${ext}` : ''
})

const hasDirectory = computed(() => {
  return props.filePath.includes('/') || props.filePath.includes('\\')
})

watch(() => props.filePath, async (newPath) => {
  pattern.value = newPath

  // 自动选择默认模式
  if (!hasDirectory.value) {
    mode.value = 'exact'
  } else {
    mode.value = 'directory'
  }

  // 加载目录选项
  if (hasDirectory.value) {
    try {
      const opts = await GetDirectoryOptions(newPath)
      directoryOptions.value = opts
      if (opts.length > 0 && opts[0]) {
        selectedDirectory.value = opts[0].pattern
      }
    } catch (e) {
      console.error('加载目录选项失败:', e)
    }
  } else {
    directoryOptions.value = []
  }
}, { immediate: true })

function close() {
  emit('close')
}

function confirm() {
  let finalPattern = pattern.value
  if (mode.value === 'directory' && selectedDirectory.value) {
    finalPattern = selectedDirectory.value
  }
  emit('confirm', mode.value, finalPattern)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}

.exclude-dialog {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg);
  border-bottom: 1px solid var(--border-default);
}

.dialog-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 24px;
  line-height: 1;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.close-btn:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.dialog-body {
  padding: var(--space-lg);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.input-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.pattern-input {
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.pattern-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.radio-option {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

.radio-option input[type="radio"] {
  cursor: pointer;
}

.radio-option input[type="radio"]:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.directory-select {
  width: 100%;
  margin-top: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
  padding: var(--space-lg);
  border-top: 1px solid var(--border-default);
}

.btn-secondary,
.btn-primary {
  padding: var(--space-sm) var(--space-lg);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-secondary {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

.btn-primary {
  background: var(--accent-success);
  color: white;
  border-color: var(--accent-success);
}

.btn-primary:hover {
  background: #059669;
}
</style>
