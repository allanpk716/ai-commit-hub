<template>
  <div v-if="visible" class="dialog-overlay" @click.self="$emit('cancel')">
    <div class="dialog">
      <div class="dialog-header">
        <h3 :class="type">{{ title }}</h3>
      </div>

      <div class="dialog-body">
        <p class="message">{{ message }}</p>

        <!-- 详细信息列表 -->
        <div v-if="details && details.length > 0" class="details-list">
          <div v-for="item in details" :key="item.label" class="detail-item">
            <span class="detail-label">{{ item.label }}:</span>
            <span class="detail-value">{{ item.value }}</span>
          </div>
        </div>

        <!-- 附加提示信息 -->
        <p v-if="note" class="note-text">{{ note }}</p>
      </div>

      <div class="dialog-footer">
        <button @click="$emit('cancel')" class="btn-cancel">{{ cancelText }}</button>
        <button @click="$emit('confirm')" :class="['btn-confirm', type]">{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface DetailItem {
  label: string
  value: string
}

defineProps<{
  visible: boolean
  title: string
  message: string
  details?: DetailItem[]
  note?: string
  confirmText: string
  cancelText: string
  type?: 'warning' | 'danger'
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
}

.dialog-header h3.warning {
  color: var(--color-warning);
}

.dialog-header h3.danger {
  color: var(--color-error);
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

/* 详细信息列表 */
.details-list {
  margin: var(--space-md) 0;
  padding: var(--space-md);
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
}

.detail-item {
  display: flex;
  gap: var(--space-sm);
  margin-bottom: var(--space-sm);
  font-size: 13px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.detail-label {
  font-weight: 500;
  color: var(--text-secondary);
  min-width: 80px;
}

.detail-value {
  flex: 1;
  font-family: var(--font-mono);
  color: var(--text-primary);
  word-break: break-all;
}

/* 提示信息 */
.note-text {
  margin: var(--space-sm) 0 0 0 !important;
  font-size: 13px;
  color: var(--accent-primary);
  line-height: 1.6;
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
  color: white;
}

.btn-confirm.warning {
  background: var(--color-warning);
}

.btn-confirm.warning:hover {
  background: #d97706;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.btn-confirm.danger {
  background: var(--color-error);
}

.btn-confirm.danger:hover {
  background: #dc2626;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}
</style>
