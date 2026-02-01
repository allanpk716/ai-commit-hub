import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { GitProject } from '../types'
import { GetAllProjects, GetProjectsWithStatus, AddProject, DeleteProject, MoveProject, ReorderProjects, DebugHookStatus } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useStatusCache } from './statusCache'

// 扩展接口，将运行时状态字段设为可选
type GitProjectWithStatus = Omit<models.GitProject, 'has_uncommitted_changes' | 'untracked_count' | 'pushover_needs_update'> & {
  has_uncommitted_changes?: boolean
  untracked_count?: number
  pushover_needs_update?: boolean
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
      // 直接使用 GetProjectsWithStatus，获取带运行时状态的项目列表
      const result = await GetProjectsWithStatus() as GitProjectWithStatus[]

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
        // 运行时状态字段（由 GetProjectsWithStatus 填充）
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

  // 监听项目状态变更事件（来自后端的 Git 操作）
  EventsOn('project-status-changed', async (data: { projectPath: string; changeType: string; timestamp: string }) => {
    console.log('[projectStore] 收到 project-status-changed 事件:', data)

    const { projectPath } = data
    const project = projects.value.find(p => p.path === projectPath)

    if (project) {
      console.log('[projectStore] 刷新项目状态:', projectPath)

      // 刷新该项目的状态（包括推送状态、分支信息等）
      const statusCache = useStatusCache()
      await statusCache.refresh(projectPath, { force: true })

      // 触发项目列表重新渲染（通过更新 project 的某个响应式属性）
      // 使用时间戳触发更新，这会强制 Vue 重新渲染该项目的组件
      project.lastModified = Date.now()

      console.log('[projectStore] 项目状态已刷新，lastModified 更新为:', project.lastModified)
    }
  })

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
