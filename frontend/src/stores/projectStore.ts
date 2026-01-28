import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { GitProject } from '../types'
import { GetAllProjects, AddProject, DeleteProject, MoveProject, ReorderProjects, DebugHookStatus } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'

// 扩展接口，包含运行时状态字段
interface GitProjectWithStatus extends models.GitProject {
  has_uncommitted_changes?: boolean
  untracked_count?: number
  pushover_needs_update?: boolean
}

// Try to import GetProjectsWithStatus, will be undefined if not yet implemented
let GetProjectsWithStatus: (() => Promise<models.GitProject[]>) | undefined
try {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore - This will be implemented in Task 5
  const appModule = require('../../wailsjs/go/main/App')
  GetProjectsWithStatus = appModule.GetProjectsWithStatus
} catch {
  // GetProjectsWithStatus not available yet, will use GetAllProjects
}

export const useProjectStore = defineStore('project', () => {
  const projects = ref<GitProject[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const selectedPath = ref<string>('')

  // 计算属性：当前选中的项目
  const selectedProject = computed(() => {
    return projects.value.find(p => p.path === selectedPath.value)
  })

  async function loadProjects() {
    loading.value = true
    error.value = null
    try {
      const result = await GetAllProjects() as models.GitProject[]
      projects.value = result.map(p => ({
        id: p.id,
        path: p.path,
        name: p.name,
        sort_order: p.sort_order,
        provider: p.provider ?? null,
        language: p.language ?? null,
        model: p.model ?? null,
        use_default: p.use_default,
        hook_installed: p.hook_installed,
        notification_mode: p.notification_mode,
        hook_version: p.hook_version,
        hook_installed_at: p.hook_installed_at
      }))
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载项目失败'
      error.value = message
      console.error('Failed to load projects:', e)
    } finally {
      loading.value = false
    }
  }

  async function loadProjectsWithStatus() {
    loading.value = true
    error.value = null
    try {
      // Try to use GetProjectsWithStatus if available (Task 5+)
      const result = GetProjectsWithStatus
        ? await GetProjectsWithStatus() as GitProjectWithStatus[]
        : await GetAllProjects() as GitProjectWithStatus[]

      projects.value = result.map(p => ({
        id: p.id,
        path: p.path,
        name: p.name,
        sort_order: p.sort_order,
        provider: p.provider ?? null,
        language: p.language ?? null,
        model: p.model ?? null,
        use_default: p.use_default,
        hook_installed: p.hook_installed,
        notification_mode: p.notification_mode,
        hook_version: p.hook_version,
        hook_installed_at: p.hook_installed_at,
        // Runtime status fields (will be populated by GetProjectsWithStatus)
        has_uncommitted_changes: p.has_uncommitted_changes ?? false,
        untracked_count: p.untracked_count ?? 0,
        pushover_needs_update: p.pushover_needs_update ?? false
      }))
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '加载失败'
      error.value = message
      console.error('Failed to load projects with status:', e)
    } finally {
      loading.value = false
    }
  }

  async function addProject(path: string) {
    loading.value = true
    error.value = null
    try {
      const result = await AddProject(path) as models.GitProject
      const newProject: GitProject = {
        id: result.id,
        path: result.path,
        name: result.name,
        sort_order: result.sort_order,
        provider: result.provider ?? null,
        language: result.language ?? null,
        model: result.model ?? null,
        use_default: result.use_default,
        hook_installed: result.hook_installed,
        notification_mode: result.notification_mode,
        hook_version: result.hook_version,
        hook_installed_at: result.hook_installed_at
      }
      projects.value.push(newProject)
      return newProject
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
      // Convert to models.GitProject format
      const modelsProjects = updatedProjects.map(p => {
        const mp = new models.GitProject()
        mp.id = p.id
        mp.path = p.path
        mp.name = p.name
        mp.sort_order = p.sort_order
        mp.provider = p.provider ?? undefined
        mp.language = p.language ?? undefined
        mp.model = p.model ?? undefined
        mp.use_default = p.use_default ?? true
        return mp
      })
      await ReorderProjects(modelsProjects)
      await loadProjects()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '重新排序失败'
      error.value = message
      throw e
    } finally {
      loading.value = false
    }
  }

  function selectProject(path: string) {
    selectedPath.value = path
  }

  async function debugHookStatus() {
    try {
      const debug = await DebugHookStatus()
      console.log('=== Hook Status Debug ===')
      console.table(debug.projects || [])
      return debug
    } catch (e) {
      console.error('Debug failed:', e)
      return null
    }
  }

  return {
    projects,
    loading,
    error,
    selectedPath,
    selectedProject,
    loadProjects,
    loadProjectsWithStatus,
    addProject,
    deleteProject,
    moveProject,
    reorderProjects,
    selectProject,
    debugHookStatus
  }
})
