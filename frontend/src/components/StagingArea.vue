<template>
  <div class="staging-area">
    <div class="staging-panels">
      <div class="file-lists-panel">
        <StagedList />
        <div class="list-divider"></div>
        <UnstagedList />
      </div>

      <div class="diff-panel">
        <DiffViewer />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import StagedList from './StagedList.vue'
import UnstagedList from './UnstagedList.vue'
import DiffViewer from './DiffViewer.vue'
</script>

<style scoped>
.staging-area {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.staging-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.file-lists-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  overflow: hidden;
  min-height: 0;
}

.list-divider {
  height: 1px;
  background: var(--border-default);
  flex-shrink: 0;
}

.diff-panel {
  overflow: hidden;
  min-height: 0;
}

@media (max-width: 1024px) {
  .staging-panels {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }
}
</style>
