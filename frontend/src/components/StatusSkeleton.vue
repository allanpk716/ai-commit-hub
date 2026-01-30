<template>
  <div class="status-skeleton">
    <div class="skeleton-badge"></div>
    <div class="skeleton-text"></div>
  </div>
</template>

<style scoped>
.status-skeleton {
  display: flex;
  align-items: center;
  gap: 8px;
}

.skeleton-badge {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-text {
  width: 60px;
  height: 16px;
  border-radius: 4px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
