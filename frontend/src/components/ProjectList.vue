<template>
  <div class="project-list">
    <div class="list-header">
      <h3>项目列表</h3>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="搜索..."
        class="search-input"
      />
    </div>

    <div v-if="projectStore.loading" class="loading">加载中...</div>
    <div v-else-if="projectStore.error" class="error">{{ projectStore.error }}</div>
    <div v-else-if="filteredProjects.length === 0" class="empty">
      {{ searchQuery ? '未找到匹配的项目' : '暂无项目，请添加项目' }}
    </div>
    <div v-else class="projects">
      <div
        v-for="(project, index) in filteredProjects"
        :key="project.id"
        class="project-item"
        :class="{ selected: selectedId === project.id }"
        draggable="true"
        @dragstart="handleDragStart(project, index, $event)"
        @dragover.prevent="handleDragOver"
        @drop="handleDrop(project, index)"
        @click="selectProject(project)"
      >
        <span class="drag-handle">⋮⋮</span>
        <span class="project-index">{{ index + 1 }}.</span>
        <span class="project-name">{{ project.name }}</span>
        <div class="project-actions">
          <button
            @click.stop="moveUp(project, index)"
            :disabled="index === 0"
            title="上移"
          >↑</button>
          <button
            @click.stop="moveDown(project, index)"
            :disabled="index === filteredProjects.length - 1"
            title="下移"
          >↓</button>
          <button
            @click.stop="handleDelete(project)"
            title="删除"
          >✕</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { GitProject } from '../types'
import { useProjectStore } from '../stores/projectStore'

const props = defineProps<{
  selectedId?: number
}>()

const emit = defineEmits<{
  select: [project: GitProject]
}>()

const projectStore = useProjectStore()
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

async function moveUp(project: GitProject, index: number) {
  if (index > 0) {
    try {
      await projectStore.moveProject(project.id, 'up')
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '移动失败'
      alert('移动失败: ' + message)
    }
  }
}

async function moveDown(project: GitProject, index: number) {
  if (index < filteredProjects.value.length - 1) {
    try {
      await projectStore.moveProject(project.id, 'down')
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '移动失败'
      alert('移动失败: ' + message)
    }
  }
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

  // Reorder projects
  const newProjects = [...filteredProjects.value]
  newProjects.splice(draggedIndex, 1)
  newProjects.splice(targetIndex, 0, draggedProject)

  // Update sort orders
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
</script>

<style scoped>
.project-list {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.list-header {
  padding: 15px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  gap: 10px;
  align-items: center;
}

.list-header h3 {
  margin: 0;
  white-space: nowrap;
}

.search-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.loading,
.error,
.empty {
  padding: 20px;
  text-align: center;
}

.error {
  color: #ff4444;
}

.empty {
  color: #999;
}

.projects {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.project-item {
  display: flex;
  align-items: center;
  padding: 10px;
  margin-bottom: 5px;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.project-item:hover {
  background-color: #f5f5f5;
}

.project-item.selected {
  background-color: #e3f2fd;
  border-color: #2196f3;
}

.drag-handle {
  cursor: grab;
  color: #999;
  margin-right: 8px;
  user-select: none;
}

.drag-handle:active {
  cursor: grabbing;
}

.project-index {
  color: #666;
  font-size: 12px;
  min-width: 30px;
}

.project-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.project-actions {
  display: none;
  gap: 4px;
}

.project-item:hover .project-actions {
  display: flex;
}

.project-actions button {
  padding: 4px 8px;
  font-size: 14px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 3px;
  cursor: pointer;
}

.project-actions button:hover:not(:disabled) {
  background-color: #f0f0f0;
}

.project-actions button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
</style>
