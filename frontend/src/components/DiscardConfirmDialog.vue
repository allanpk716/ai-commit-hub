<template>
  <div v-if="visible" class="dialog-overlay" @click.self="$emit('cancel')">
    <div class="dialog">
      <div class="dialog-header">
        <h3>⚠️ 确认还原文件</h3>
      </div>
      <div class="dialog-body">
        <p>此操作将<strong>永久丢失</strong>文件 <code>{{ fileName }}</code> 的所有未提交修改。</p>
        <p class="warning-text">此操作无法撤销！</p>
      </div>
      <div class="dialog-footer">
        <button @click="$emit('cancel')" class="btn-cancel">取消</button>
        <button @click="$emit('confirm')" class="btn-confirm">确认还原</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
  fileName: string
}>()

defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}

.dialog {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  min-width: 400px;
  max-width: 500px;
}

.dialog-header {
  padding: var(--space-lg) var(--space-lg) var(--space-md);
  border-bottom: 1px solid var(--border-default);
}

.dialog-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-warning);
}

.dialog-body {
  padding: var(--space-lg);
}

.dialog-body p {
  margin: 0 0 var(--space-sm) 0;
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.dialog-body code {
  background: var(--bg-elevated);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-primary);
}

.warning-text {
  color: var(--color-error);
  font-weight: 500;
  margin: 0 !important;
}

.dialog-footer {
  padding: var(--space-md) var(--space-lg);
  border-top: 1px solid var(--border-default);
  display: flex;
  gap: var(--space-sm);
  justify-content: flex-end;
}

.btn-cancel,
.btn-confirm {
  padding: var(--space-sm) var(--space-lg);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: none;
}

.btn-cancel {
  background: var(--bg-elevated);
  color: var(--text-secondary);
  border: 1px solid var(--border-default);
}

.btn-cancel:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-hover);
}

.btn-confirm {
  background: var(--color-error);
  color: white;
}

.btn-confirm:hover {
  background: #dc2626;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}
</style>
