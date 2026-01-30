import { defineStore } from 'pinia'
import { ref, computed, onMounted } from 'vue'
import type {
  ProjectStatusCache,
  ProjectStatusCacheMap,
  CacheOptions
} from '../types/status'
import { GetStagingStatus, GetProjectStatus, GetUntrackedFiles, GetPushoverHookStatus, GetAllProjectStatuses } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

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
   * 乐观更新缓存（立即应用更新，支持回滚）
   * @param path 项目路径
   * @param updates 要更新的字段
   * @returns 回滚函数，如果缓存不存在返回 undefined
   */
  function updateOptimistic(path: string, updates: Partial<ProjectStatusCache>): (() => void) | undefined {
    const current = cache.value[path]
    if (!current) {
      return undefined
    }

    // 保存当前状态用于可能的回滚
    const previous = { ...current }

    // 应用更新
    cache.value[path] = {
      ...current,
      ...updates,
      lastUpdated: Date.now()
    }

    // 返回回滚函数
    return () => {
      cache.value[path] = previous
    }
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
   * @param options 刷新选项
   */
  async function refresh(path: string, options?: { force?: boolean }): Promise<void> {
    // 防止并发重复请求
    if (pendingRequests.value.has(path)) {
      return
    }

    // 如果未强制刷新且未过期，跳过
    if (!options?.force && !isExpired(path)) {
      return
    }

    pendingRequests.value.add(path)

    // 初始化或标记加载中
    if (!cache.value[path]) {
      cache.value[path] = {
        gitStatus: null,
        stagingStatus: null,
        untrackedCount: 0,
        pushoverStatus: null,
        lastUpdated: 0,
        loading: true,
        error: null,
        stale: false
      }
    } else {
      cache.value[path].loading = true
    }

    try {
      const [gitStatus, stagingStatus, untrackedFiles, pushoverStatus] = await Promise.all([
        GetProjectStatus(path).catch(() => null),
        GetStagingStatus(path).catch(() => null),
        GetUntrackedFiles(path).catch(() => []),
        GetPushoverHookStatus(path).catch(() => null)
      ])

      cache.value[path] = {
        gitStatus,
        stagingStatus,
        untrackedCount: untrackedFiles?.length || 0,
        pushoverStatus: pushoverStatus,
        lastUpdated: Date.now(),
        loading: false,
        error: null,
        stale: false
      }
    } catch (error) {
      cache.value[path].loading = false
      cache.value[path].error = String(error)
    } finally {
      pendingRequests.value.delete(path)
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
      await refresh(path, { force: forceRefresh })
    }

    return getStatus(path)
  }

  /**
   * 批量预加载多个项目的状态
   * @param projectPaths 项目路径数组
   */
  async function preload(projectPaths: string[]): Promise<void> {
    if (projectPaths.length === 0) return;

    try {
      // 使用批量接口获取所有项目状态
      const statuses = await GetAllProjectStatuses(projectPaths);

      // 将批量获取的状态填充到缓存中
      for (const [path, status] of Object.entries(statuses)) {
        cache.value[path] = {
          gitStatus: status.gitStatus,
          stagingStatus: status.stagingStatus,
          untrackedCount: status.untrackedCount,
          pushoverStatus: status.pushoverStatus,
          lastUpdated: new Date(status.lastUpdated).getTime(),
          loading: false,
          error: null,
          stale: false
        };
      }
    } catch (error) {
      console.error('Preload failed, falling back to individual loads:', error);
      // 降级到逐个加载
      await Promise.all(projectPaths.map(path => refresh(path, { force: true })));
    }
  }

  /**
   * 初始化状态缓存，预加载所有项目状态
   */
  async function init(): Promise<void> {
    // 从 projectStore 获取所有项目路径
    const { useProjectStore } = await import('./project')
    const projectStore = useProjectStore()
    const paths = projectStore.projects.map(p => p.path)
    await preload(paths)
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
   * 判断错误是否可重试
   * @param error 错误对象
   * @returns 如果错误可重试返回 true
   */
  function isRetryable(error: unknown): boolean {
    if (error instanceof Error) {
      return error.message.includes('network') ||
             error.message.includes('timeout') ||
             error.message.includes('ECONN')
    }
    return false
  }

  /**
   * 带重试的刷新方法
   * @param path 项目路径
   * @param maxRetries 最大重试次数（默认 2 次）
   */
  async function refreshWithRetry(path: string, maxRetries = 2): Promise<void> {
    for (let i = 0; i <= maxRetries; i++) {
      try {
        await refresh(path, { force: true })
        return
      } catch (error) {
        // 如果是最后一次尝试或错误不可重试，则抛出错误
        if (i === maxRetries || !isRetryable(error)) {
          throw error
        }
        // 等待一段时间后重试（指数退避）
        await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)))
      }
    }
  }

  /**
   * 手动刷新方法（带错误处理）
   * @param path 项目路径
   */
  async function manualRefresh(path: string): Promise<void> {
    try {
      await refreshWithRetry(path)
    } catch (error) {
      console.error('Manual refresh failed:', error)
      throw error
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
    updateOptimistic,
    invalidate,
    invalidateAll,
    clearCache,
    clearAllCache,
    refresh,
    getStatusOrRefresh,
    preload,
    init,
    updateOptions,
    isRetryable,
    refreshWithRetry,
    manualRefresh
  }
})
