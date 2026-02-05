<template>
  <div class="commit-controls">
    <button
      class="btn btn-primary"
      :disabled="!canGenerate"
      @click="$emit('generate')"
      :title="generateTooltip"
    >
      {{ isGenerating ? "生成中..." : "生成消息" }}
    </button>

    <button
      class="btn btn-success"
      :disabled="!canCommit"
      @click="$emit('commit')"
      title="提交到本地仓库"
    >
      提交
    </button>

    <button
      class="btn btn-push"
      :disabled="!canPush"
      @click="$emit('push')"
      :title="pushTooltip"
    >
      {{ isPushing ? "推送中" : "推送" }}
    </button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  canGenerate: boolean;
  isGenerating: boolean;
  canCommit: boolean;
  canPush: boolean;
  isPushing?: boolean;
  generateTooltip?: string;
  pushTooltip?: string;
}

withDefaults(defineProps<Props>(), {
  isPushing: false,
  generateTooltip: "生成 Commit 消息",
  pushTooltip: "推送到远程仓库",
});

defineEmits<{
  generate: [];
  commit: [];
  push: [];
}>();
</script>

<style scoped>
.commit-controls {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  align-items: center;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn:not(:disabled):hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-primary:not(:disabled):hover {
  background-color: #0056b3;
}

.btn-success {
  background-color: #28a745;
  color: white;
}

.btn-success:not(:disabled):hover {
  background-color: #218838;
}

.btn-push {
  background-color: #6c757d;
  color: white;
}

.btn-push:not(:disabled):hover {
  background-color: #5a6268;
}

.btn-push:not(:disabled) {
  background-color: #ffc107;
  color: #000;
}

.btn-push:not(:disabled):hover {
  background-color: #e0a800;
}
</style>
