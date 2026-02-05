<template>
  <div class="commit-message">
    <h3>Commit 消息</h3>

    <div v-if="isGenerating" class="loading-indicator">
      <div class="spinner"></div>
      <span>正在生成...</span>
    </div>

    <div v-else-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      <span>{{ error }}</span>
    </div>

    <textarea
      v-else
      v-model="messageValue"
      class="message-textarea"
      placeholder="生成的 commit 消息将显示在这里"
      rows="8"
      @input="handleInput"
    />

    <div class="message-info">
      <span class="char-count">{{ messageValue.length }} 字符</span>
      <button
        v-if="messageValue.length > 0"
        @click="handleClear"
        class="clear-btn"
        title="清空消息"
        :disabled="isGenerating"
      >
        ×
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  message: string;
  isGenerating: boolean;
  error?: string;
}

const props = withDefaults(defineProps<Props>(), {
  error: "",
});

const emit = defineEmits<{
  "update:message": [value: string];
}>();

const messageValue = computed({
  get: () => props.message,
  set: (value) => emit("update:message", value),
});

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement;
  emit("update:message", target.value);
}

function handleClear() {
  emit("update:message", "");
}
</script>

<style scoped>
.commit-message {
  margin-bottom: 20px;
}

.commit-message h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

.loading-indicator {
  padding: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #666;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 3px solid #e0e0e0;
  border-top-color: #007bff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  padding: 12px 16px;
  background-color: #f8d7da;
  color: #721c24;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-icon {
  font-size: 18px;
}

.message-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: "Courier New", "Monaco", "Menlo", monospace;
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  min-height: 120px;
  transition: border-color 0.2s;
}

.message-textarea:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.1);
}

.message-textarea::placeholder {
  color: #999;
}

.message-info {
  margin-top: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.char-count {
  font-size: 12px;
  color: #666;
}

.clear-btn {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  color: #999;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
}

.clear-btn:hover:not(:disabled) {
  background-color: #f0f0f0;
  color: #333;
}

.clear-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
