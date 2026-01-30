import { defineStore } from 'pinia'
import { ref, computed, onMounted } from 'vue'
import type {
  ProjectStatusCache,
  ProjectStatusCacheMap,
  CacheOptions,
  GitStatus
} from '../types/status'
import type { StagingStatus, HookStatus } from '../types'
import { GetStagingStatus } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { models } from '../../wailsjs/go/models'

/**
 * 默认缓存 TTL（30 秒）
 */
const DEFAULT_TTL = 30 * 1000

/**
 * StatusCache Store - 项目状态缓存管理
 *
 * 用于缓存项目状态信息，避免频繁调用后端 API 导致 UI 闪烁
 */
export const useStatusCache = defineStore('statusCache', () => {
  // ========== 状态 ==========

  /**
   * 状态缓存映射表
   * 键：项目路径
   * 值：项目状态缓存
   */
  const cache = ref<ProjectStatusCacheMap>({})

  /**
   * 正在进行的请求集合（防止重复请求）
   * 键：项目路径
   */
  const pendingRequests = ref<Set<string>>(new Set())

  /**
   * 缓存配置
   */
  const options = ref<CacheOptions>({
    ttl: DEFAULT_TTL,
    backgroundRefresh: true
  })

  // ========== 计算属性 ==========

  /**
   * 获取所有已缓存的项目路径
   */
  const cachedPaths = computed(() => Object.keys(cache.value))

  /**
   * 获取所有已过期的项目路径
   */
  const expiredPaths = computed(() =>
    cachedPaths.value.filter(path => isExpired(path))
  )

  // ========== 核心方法 ==========

  /**
   * 获取项目的缓存状态
   * @param path 项目路径
   * @returns 项目状态缓存，如果不存在则返回 null
   */
  function getStatus(path: string): ProjectStatusCache | null {
    return cache.value[path] || null
  }

  /**
   * 检查缓存是否过期
   * @param path 项目路径
   * @returns 如果缓存不存在或已过期返回 true
   */
  function isExpired(path: string): boolean {
    const cached = cache.value[path]
    if (!cached) {
      return true
    }

    const now = Date.now()
    const elapsed = now - cached.lastUpdated
    const ttl = options.value.ttl || DEFAULT_TTL

    return elapsed > ttl
  }

  /**
   * 检查是否正在加载
   * @param path 项目路径
   * @returns 如果正在加载返回 true
   */
  function isLoading(path: string): boolean {
    return pendingRequests.value.has(path)
  }

  /**
   * 创建一个新的空缓存条目
   * @param path 项目路径
   * @returns 新的缓存条目
   */
  function createCacheEntry(path: string): ProjectStatusCache {
    return {
      gitStatus: null,
      stagingStatus: null,
      untrackedCount: 0,
      pushoverStatus: null,
      lastUpdated: Date.now(),
      loading: false,
      error: null,
      stale: true
    }
  }

  /**
   * 初始化项目缓存（如果不存在）
   * @param path 项目路径
   */
  function initCache(path: string): void {
    if (!cache.value[path]) {
      cache.value[path] = createCacheEntry(path)
    }
  }

  /**
   * 更新缓存条目
   * @param path 项目路径
   * @param updates 要更新的字段
   */
  function updateCache(path: string, updates: Partial<ProjectStatusCache>): void {
    if (!cache.value[path]) {
      initCache(path)
    }

    cache.value[path] = {
      ...cache.value[path],
      ...updates,
      lastUpdated: Date.now()
    }
  }

  /**
   * 设置加载状态
   * @param path 项目路径
   * @param loading 是否正在加载
   */
  function setLoading(path: string, loading: boolean): void {
    if (loading) {
      pendingRequests.value.add(path)
    } else {
      pendingRequests.value.delete(path)
    }

    updateCache(path, { loading })
  }

  /**
   * 设置错误状态
   * @param path 项目路径
   * @param error 错误信息
   */
  function setError(path: string, error: string | null): void {
    updateCache(path, { error, loading: false })
    pendingRequests.value.delete(path)
  }

  /**
   * 使缓存失效
   * @param path 项目路径
   */
  function invalidate(path: string): void {
    if (cache.value[path]) {
      cache.value[path].stale = true
    }
  }

  /**
   * 使所有缓存失效
   */
  function invalidateAll(): void {
    Object.keys(cache.value).forEach(path => {
      cache.value[path].stale = true
    })
  }

  /**
   * 清除指定项目的缓存
   * @param path 项目路径
   */
  function clearCache(path: string): void {
    delete cache.value[path]
    pendingRequests.value.delete(path)
  }

  /**
   * 清除所有缓存
   */
  function clearAllCache(): void {
    cache.value = {}
    pendingRequests.value.clear()
  }

  /**
   * 刷新项目状态（从后端获取最新状态）
   * @param path 项目路径
   */
  async function refreshStatus(path: string): Promise<void> {
    // 防止重复请求
    if (isLoading(path)) {
      return
    }

    setLoading(path, true)
    setError(path, null)

    try {
      // 获取暂存区状态
      const stagingStatus = await GetStagingStatus(path) as models.StagingStatus

      // 转换为前端类型
      const frontendStagingStatus: StagingStatus = {
        staged: stagingStatus.staged.map(f => ({
          path: f.path,
          status: f.status,
          ignored: f.ignored
        })),
        unstaged: stagingStatus.unstaged.map(f => ({
          path: f.path,
          status: f.status,
          ignored: f.ignored
        })),
        untracked: stagingStatus.untracked.map(f => ({
          path: f.path
        }))
      }

      // 构造 GitStatus（从暂存区状态获取分支信息）
      const gitStatus: GitStatus = {
        branch: '' // TODO: 从后端 API 获取分支信息
      }

      // 计算未跟踪文件数量
      const untrackedCount = frontendStagingStatus.untracked.length

      // 更新缓存
      updateCache(path, {
        gitStatus,
        stagingStatus: frontendStagingStatus,
        untrackedCount,
        pushoverStatus: null, // TODO: 从后端 API 获取 Pushover 状态
        loading: false,
        error: null,
        stale: false
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '刷新状态失败'
      setError(path, message)
    } finally {
      setLoading(path, false)
    }
  }

  /**
   * 获取项目状态（优先从缓存获取）
   * @param path 项目路径
   * @param forceRefresh 是否强制刷新
   * @returns 项目状态缓存
   */
  async function getStatusOrRefresh(path: string, forceRefresh = false): Promise<ProjectStatusCache | null> {
    // 初始化缓存（如果不存在）
    initCache(path)

    // 如果缓存过期或强制刷新，则从后端获取最新状态
    if (forceRefresh || isExpired(path)) {
      await refreshStatus(path)
    }

    return getStatus(path)
  }

  /**
   * 批量预加载多个项目的状态
   * @param paths 项目路径数组
   */
  async function preloadStatuses(paths: string[]): Promise<void> {
    // 过滤出需要刷新的项目（不存在或已过期）
    const pathsToRefresh = paths.filter(path => isExpired(path))

    // 并发加载所有需要刷新的项目
    await Promise.all(
      pathsToRefresh.map(path => refreshStatus(path))
    )
  }

  /**
   * 更新缓存配置
   * @param newOptions 新的配置选项
   */
  function updateOptions(newOptions: Partial<CacheOptions>): void {
    options.value = {
      ...options.value,
      ...newOptions
    }
  }

  /**
   * 初始化事件监听器
   */
  function initEventListeners(): void {
    // 监听项目状态变化事件，自动使缓存失效
    EventsOn('project-status-changed', (data: { path?: string }) => {
      if (data.path) {
        invalidate(data.path)
      } else {
        invalidateAll()
      }
    })
  }

  // ========== 初始化 ==========

  // 在 store 创建时初始化事件监听器
  onMounted(() => {
    initEventListeners()
  })

  // ========== 返回 ==========

  return {
    // 状态
    cache,
    pendingRequests,
    options,

    // 计算属性
    cachedPaths,
    expiredPaths,

    // 方法
    getStatus,
    isExpired,
    isLoading,
    initCache,
    updateCache,
    setLoading,
    setError,
    invalidate,
    invalidateAll,
    clearCache,
    clearAllCache,
    refreshStatus,
    getStatusOrRefresh,
    preloadStatuses,
    updateOptions
  }
})
