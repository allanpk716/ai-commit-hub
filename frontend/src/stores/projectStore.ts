import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GitProject } from '../types'
import { GetAllProjects, AddProject, DeleteProject, MoveProject, ReorderProjects } from '../../wailsjs/go/main/App'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<GitProject[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadProjects() {
    loading.value = true
    error.value = null
    try {
      const result = await GetAllProjects()
      projects.value = result
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载项目失败'
      error.value = message
      console.error('Failed to load projects:', e)
    } finally {
      loading.value = false
    }
  }

  async function addProject(path: string) {
    loading.value = true
    error.value = null
    try {
      const result = await AddProject(path)
      projects.value.push(result)
      return result
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '添加项目失败'
      error.value = message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteProject(id: number) {
    loading.value = true
    error.value = null
    try {
      await DeleteProject(id)
      projects.value = projects.value.filter(p => p.id !== id)
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '删除项目失败'
      error.value = message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function moveProject(id: number, direction: 'up' | 'down') {
    loading.value = true
    error.value = null
    try {
      await MoveProject(id, direction)
      await loadProjects()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '移动项目失败'
      error.value = message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function reorderProjects(updatedProjects: GitProject[]) {
    loading.value = true
    error.value = null
    try {
      await ReorderProjects(updatedProjects)
      await loadProjects()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '重新排序失败'
      error.value = message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    projects,
    loading,
    error,
    loadProjects,
    addProject,
    deleteProject,
    moveProject,
    reorderProjects
  }
})
