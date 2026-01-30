<template>
  <div class="project-list">
    <!-- Header -->
    <div class="list-header">
      <div class="header-title">
        <span class="icon">📁</span>
        <h3>项目列表</h3>
      </div>
      <div class="project-count">{{ projectStore.projects.length }}</div>
    </div>

    <!-- Search -->
    <div class="search-container">
      <span class="search-icon">🔍</span>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="搜索项目..."
        class="search-input"
      />
      <button
        v-if="searchQuery"
        @click="searchQuery = ''"
        class="search-clear"
      >×</button>
    </div>

    <!-- Loading state -->
    <div v-if="projectStore.loading" class="state-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="projectStore.error" class="state-container error">
      <div class="state-icon">⚠️</div>
      <p>{{ projectStore.error }}</p>
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredProjects.length === 0" class="state-container empty">
      <div class="state-icon">📭</div>
      <p>{{ searchQuery ? '未找到匹配的项目' : '暂无项目' }}</p>
      <span v-if="!searchQuery" class="empty-hint">点击上方 "添加项目" 按钮开始</span>
    </div>

    <!-- Project list -->
    <transition-group v-else tag="div" name="list-item" class="projects">
      <div
        v-for="(project, index) in filteredProjects"
        :key="project.id"
        class="project-item"
        :class="{ selected: selectedId === project.id }"
        :draggable="!searchQuery"
        @dragstart="handleDragStart(project, index, $event)"
        @dragover.prevent="handleDragOver"
        @drop="handleDrop(project, index)"
        @click="selectProject(project)"
      >
        <span class="drag-handle" title="拖拽排序">⋮⋮</span>

        <div class="project-info">
          <span class="project-name">{{ project.name }}</span>
          <span class="project-path">{{ project.path }}</span>

          <!-- 状态指示器行（从 StatusCache 获取） -->
          <div class="project-status-row">
            <template v-if="!getProjectStatus(project).loading">
              <span
                v-if="getProjectStatus(project).untrackedCount > 0"
                class="status-indicator untracked"
                :title="`${getProjectStatus(project).untrackedCount} 个未跟踪文件`"
              >
                ➕ {{ getProjectStatus(project).untrackedCount }}
              </span>
              <span
                v-if="getProjectStatus(project).pushoverUpdateAvailable"
                class="status-indicator update"
                title="Pushover 插件可更新"
              >
                ⬆️
              </span>
            </template>
            <StatusSkeleton v-else />
          </div>
        </div>

        <div class="project-actions">
          <button
            @click.stop="handleDelete(project)"
            class="action-btn danger"
            title="删除"
          >×</button>
        </div>

        <div class="project-indicator" v-if="selectedId === project.id"></div>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { GitProject } from '../types'
import { useProjectStore } from '../stores/projectStore'
import { useStatusCache } from '../stores/statusCache'
import StatusSkeleton from './StatusSkeleton.vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'

const props = defineProps<{
  selectedId?: number
}>()

const emit = defineEmits<{
  select: [project: GitProject]
}>()

const projectStore = useProjectStore()
const statusCache = useStatusCache()
const searchQuery = ref('')
const draggedItem = ref<{ project: GitProject; index: number } | null>(null)

const filteredProjects = computed(() => {
  if (!searchQuery.value) {
    return projectStore.projects
  }
  const query = searchQuery.value.toLowerCase()
  return projectStore.projects.filter(p =>
    p.name.toLowerCase().includes(query) ||
    p.path.toLowerCase().includes(query)
  )
})

function selectProject(project: GitProject) {
  emit('select', project)
}

async function handleDelete(project: GitProject) {
  if (confirm(`确定要删除项目 "${project.name}" 吗?`)) {
    try {
      await projectStore.deleteProject(project.id)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '删除失败'
      alert('删除失败: ' + message)
    }
  }
}

function handleDragStart(project: GitProject, index: number, event: DragEvent) {
  if (searchQuery.value) {
    event.preventDefault()
    alert('搜索时无法拖拽排序')
    return
  }
  draggedItem.value = { project, index }
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

function handleDragOver(event: DragEvent) {
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

async function handleDrop(targetProject: GitProject, targetIndex: number) {
  if (!draggedItem.value) return

  const { project: draggedProject, index: draggedIndex } = draggedItem.value

  if (draggedProject.id === targetProject.id) {
    draggedItem.value = null
    return
  }

  const newProjects = [...filteredProjects.value]
  newProjects.splice(draggedIndex, 1)
  newProjects.splice(targetIndex, 0, draggedProject)

  const reorderedProjects = newProjects.map((p, i) => ({
    ...p,
    sort_order: i
  }))

  try {
    await projectStore.reorderProjects(reorderedProjects as GitProject[])
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '排序失败'
    alert('排序失败: ' + message)
  }

  draggedItem.value = null
}

/**
 * 获取项目状态（优先从缓存获取）
 * @param project Git 项目
 * @returns 项目状态对象
 */
const getProjectStatus = (project: GitProject) => {
  const cached = statusCache.getStatus(project.path)

  if (cached?.loading) {
    return { loading: true }
  }

  if (cached?.error) {
    return { error: true, message: cached.error }
  }

  return {
    loading: false,
    error: false,
    untrackedCount: cached?.untrackedCount ?? 0,
    pushoverUpdateAvailable: cached?.pushoverStatus?.updateAvailable ?? false,
    stale: cached?.stale ?? false
  }
}

// 监听启动完成事件，刷新项目列表
onMounted(() => {
  console.log('[ProjectList] 组件已挂载，注册事件监听器')

  // 监听启动完成事件
  EventsOn('startup-complete', async () => {
    console.log('[ProjectList] startup-complete 事件触发，加载项目列表')
    try {
      // 仅加载项目基本信息，状态由 StatusCache 管理
      await projectStore.loadProjects()
      console.log('[ProjectList] 项目列表加载完成', {
        projectCount: projectStore.projects.length
      })
    } catch (error) {
      console.error('[ProjectList] 加载项目列表失败:', error)
    }
  })
})

// 清理事件监听器，防止内存泄漏
onUnmounted(() => {
  console.log('[ProjectList] 组件卸载，移除事件监听器')
  EventsOff('startup-complete')
})
</script>

<style scoped>
.project-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 300px;
  min-width: 300px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

/* Header */
.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg) var(--space-lg) var(--space-md);
  border-bottom: 1px solid var(--glass-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.header-title .icon {
  font-size: 16px;
  line-height: 1;
}

.header-title h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.project-count {
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-display);
  color: var(--accent-primary);
  background: rgba(6, 182, 212, 0.15);
  padding: 2px 8px;
  border-radius: 10px;
}

/* Search */
.search-container {
  position: relative;
  margin: 0 var(--space-lg) var(--space-md);
}

.search-icon {
  position: absolute;
  left: var(--space-md);
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  font-size: 14px;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: var(--space-sm) var(--space-xl) var(--space-sm) 36px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 13px;
  transition: all var(--transition-fast);
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.search-clear {
  position: absolute;
  right: var(--space-sm);
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  font-size: 18px;
  line-height: 1;
}

.search-clear:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

/* State containers */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-2xl);
  text-align: center;
  flex: 1;
  animation: fade-in 0.3s ease-out;
}

.state-icon {
  font-size: 48px;
  margin-bottom: var(--space-md);
  opacity: 0.6;
}

.state-container p {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
}

.state-container.error .state-icon {
  filter: hue-rotate(30deg);
}

.state-container.error p {
  color: var(--accent-error);
}

.empty-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: var(--space-sm) !important;
}

/* Loading spinner */
.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--bg-tertiary);
  border-top: 2px solid var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: var(--space-md);
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* Project list */
.projects {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-sm);
}

.project-item {
  position: relative;
  display: flex;
  align-items: center;
  padding: var(--space-md);
  margin-bottom: var(--space-xs);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  animation: slide-in 0.3s ease-out;
}

.project-item:hover {
  background: var(--bg-elevated);
  border-color: var(--border-hover);
  transform: translateX(2px);
}

.project-item.selected {
  background: rgba(6, 182, 212, 0.15);
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 1px var(--accent-primary);
}

.drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-right: var(--space-sm);
  color: var(--text-muted);
  cursor: grab;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity var(--transition-fast);
  font-size: 12px;
  line-height: 1;
  letter-spacing: -2px;
}

.project-item:hover .drag-handle {
  opacity: 0.5;
}

.drag-handle:hover {
  opacity: 1 !important;
}

.drag-handle:active {
  cursor: grabbing;
}

.project-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.project-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.project-path {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.project-status-row {
  display: flex;
  gap: 6px;
  margin-top: 6px;
  flex-wrap: wrap;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 500;
}

.status-indicator.uncommitted {
  color: #f97316;
  background: rgba(249, 115, 22, 0.15);
}

.status-indicator.untracked {
  color: #eab308;
  background: rgba(234, 179, 8, 0.15);
}

.status-indicator.update {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.15);
}

.status-indicator.loading {
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.15);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.project-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.project-item:hover .project-actions {
  opacity: 1;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: 16px;
  line-height: 1;
}

.action-btn:hover:not(:disabled) {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: white;
}

.action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.action-btn.danger:hover:not(:disabled) {
  background: var(--accent-error);
  border-color: var(--accent-error);
}

.project-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: linear-gradient(180deg, var(--accent-primary), var(--accent-secondary));
  border-radius: 0 2px 2px 0;
}

/* List transitions */
.list-item-enter-active {
  transition: all 0.3s ease;
}

.list-item-leave-active {
  transition: all 0.2s ease;
}

.list-item-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.list-item-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.list-item-move {
  transition: transform 0.3s ease;
}
</style>
