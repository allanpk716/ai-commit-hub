<template>
  <div class="app">
    <div class="toolbar">
      <h1>AI Commit Hub</h1>
      <div class="toolbar-actions">
        <button @click="openAddProject">+ 添加项目</button>
        <button @click="openConfigFolder">⚙ 设置</button>
      </div>
    </div>

    <div class="content">
      <div class="project-list">
        <h2>项目列表</h2>
        <div v-if="projectStore.loading">加载中...</div>
        <div v-else-if="projectStore.error" class="error">
          {{ projectStore.error }}
        </div>
        <div v-else-if="projectStore.projects.length === 0" class="empty">
          暂无项目，请添加项目
        </div>
        <div v-else class="projects">
          <div
            v-for="project in projectStore.projects"
            :key="project.id"
            class="project-item"
          >
            <span class="project-name">{{ project.name }}</span>
            <span class="project-path">{{ project.path }}</span>
            <button @click="handleDelete(project)" class="delete-btn">✕</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { OpenConfigFolder } from '../wailsjs/go/main/App'

const projectStore = useProjectStore()

onMounted(() => {
  projectStore.loadProjects()
})

async function openAddProject() {
  // TODO: Open file dialog to select project path
  const path = prompt('请输入 Git 仓库路径:')
  if (path) {
    try {
      await projectStore.addProject(path)
      alert('项目添加成功!')
    } catch (e: any) {
      alert('添加失败: ' + e.message)
    }
  }
}

async function handleDelete(project: any) {
  if (confirm(`确定要删除项目 "${project.name}" 吗?`)) {
    try {
      await projectStore.deleteProject(project.id)
    } catch (e: any) {
      alert('删除失败: ' + e.message)
    }
  }
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: any) {
    alert('打开配置文件夹失败: ' + e.message)
  }
}
</script>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #e0e0e0;
}

.toolbar h1 {
  margin: 0;
  font-size: 20px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.toolbar-actions button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.toolbar-actions button:hover {
  background: #f5f5f5;
}

.content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.project-list h2 {
  margin-top: 0;
}

.project-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  margin-bottom: 8px;
}

.project-name {
  font-weight: bold;
  margin-right: 10px;
}

.project-path {
  flex: 1;
  color: #666;
  font-size: 14px;
}

.delete-btn {
  padding: 4px 8px;
  border: 1px solid #ff4444;
  color: #ff4444;
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.delete-btn:hover {
  background: #fff5f5;
}

.error {
  color: #ff4444;
}

.empty {
  color: #999;
  text-align: center;
  padding: 40px;
}
</style>
