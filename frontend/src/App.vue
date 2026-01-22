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
      <ProjectList
        :selected-id="selectedProjectId"
        @select="handleSelectProject"
      />
      <CommitPanel v-if="selectedProjectId" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { useCommitStore } from './stores/commitStore'
import { OpenConfigFolder, SelectProjectFolder } from '../wailsjs/go/main/App'
import ProjectList from './components/ProjectList.vue'
import CommitPanel from './components/CommitPanel.vue'
import type { GitProject } from './types'

const projectStore = useProjectStore()
const commitStore = useCommitStore()
const selectedProjectId = ref<number>()

onMounted(() => {
  projectStore.loadProjects()
})

async function openAddProject() {
  try {
    const path = await SelectProjectFolder()
    if (path) {  // Empty string means user canceled
      await projectStore.addProject(path)
      alert('项目添加成功!')
    }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    alert('添加项目失败: ' + message)
  }
}

function handleSelectProject(project: GitProject) {
  selectedProjectId.value = project.id
  // Load project status for commit panel
  commitStore.loadProjectStatus(project.path)
}

async function openConfigFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    alert('打开配置文件夹失败: ' + message)
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
  display: flex;
  gap: 20px;
  height: calc(100vh - 70px);
}
</style>
