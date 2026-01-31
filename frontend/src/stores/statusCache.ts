import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  ProjectStatusCache,
  ProjectStatusCacheMap,
  CacheOptions
} from '../types/status'
import type { HookStatus } from '../types/index'
import { GetStagingStatus, GetProjectStatus, GetUntrackedFiles, GetPushoverHookStatus, GetAllProjectStatuses, GetPushStatus } from '../../wailsjs/go/main/App'
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
  function createCacheEntry(_path: string): ProjectStatusCache {
    return {
      gitStatus: null,
      stagingStatus: null,
      untrackedCount: 0,
      pushoverStatus: null,
      pushStatus: null,
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

    const current = cache.value[path]!
    cache.value[path] = {
      ...current,
      ...updates,
      lastUpdated: Date.now()
    }

    // 验证并修正数据一致性
    validateAndFix(path)
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

    // 验证并修正数据一致性
    validateAndFix(path)

    // 返回回滚函数
    return () => {
      cache.value[path] = previous
      // 回滚后也需要验证一致性
      validateAndFix(path)
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
      const cached = cache.value[path]
      if (cached) {
        cached.stale = true
      }
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

  // ========== 数据验证层 ==========

  /**
   * 验证缓存条目的数据一致性
   * @param cacheEntry 要验证的缓存条目
   * @returns 是否一致
   */
  function isConsistent(cacheEntry: ProjectStatusCache | null): boolean {
    if (!cacheEntry) return true

    const expectedUntrackedCount = cacheEntry.stagingStatus?.untracked?.length ?? 0
    return cacheEntry.untrackedCount === expectedUntrackedCount
  }

  /**
   * 规范化缓存条目（自动修正不一致数据）
   * @param cacheEntry 要规范化的缓存条目
   * @returns 规范化后的缓存条目
   */
  function normalizeCache(cacheEntry: ProjectStatusCache): ProjectStatusCache {
    if (!cacheEntry) return cacheEntry

    const normalized = { ...cacheEntry }

    // 修正 untrackedCount
    if (normalized.stagingStatus?.untracked) {
      normalized.untrackedCount = normalized.stagingStatus.untracked.length
    }

    return normalized
  }

  /**
   * 验证并修复指定路径的缓存
   * @param path 项目路径
   */
  function validateAndFix(path: string): void {
    const cached = cache.value[path]
    if (cached && !isConsistent(cached)) {
      // 仅在开发环境输出调试信息
      if (import.meta.env.DEV) {
        console.debug('[StatusCache] 检测到不一致，已自动修正', {
          path,
          before: { untrackedCount: cached.untrackedCount },
          after: { untrackedCount: normalizeCache(cached).untrackedCount }
        })
      }
      cache.value[path] = normalizeCache(cached)
    }
  }

  /**
   * 批量验证并修复所有缓存
   */
  function validateAndFixAll(): void {
    cachedPaths.value.forEach(path => validateAndFix(path))
  }

  /**
   * 获取缓存健康状态（用于调试）
   * @returns 缓存健康状态信息
   */
  function getHealthStatus(): {
    total: number
    consistent: number
    inconsistent: string[]
  } {
    const paths = cachedPaths.value
    const inconsistent: string[] = []
    let consistent = 0

    for (const path of paths) {
      const cached = cache.value[path]
      if (cached) {
        if (isConsistent(cached)) {
          consistent++
        } else {
          inconsistent.push(path)
        }
      }
    }

    return {
      total: paths.length,
      consistent,
      inconsistent
    }
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
      const [gitStatus, stagingStatus, untrackedFiles, pushoverStatus, pushStatus] = await Promise.all([
        GetProjectStatus(path).catch(() => null),
        GetStagingStatus(path).catch(() => null),
        GetUntrackedFiles(path).catch(() => []),
        GetPushoverHookStatus(path).catch(() => null),
        GetPushStatus(path).catch(() => null)
      ])

      const cached = cache.value[path]
      if (cached) {
        cached.gitStatus = gitStatus as any
        cached.stagingStatus = stagingStatus
        cached.untrackedCount = untrackedFiles?.length || 0
        cached.pushoverStatus = pushoverStatus as any
        cached.pushStatus = pushStatus as any
        cached.lastUpdated = Date.now()
        cached.loading = false
        cached.error = null
        cached.stale = false
      }

      // 验证并修正数据一致性
      validateAndFix(path)
    } catch (error) {
      const cached = cache.value[path]
      if (cached) {
        cached.loading = false
        cached.error = String(error)
      }
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
        // 使用 normalizeCache 规范化数据，确保一致性
        cache.value[path] = normalizeCache({
          gitStatus: status.gitStatus,
          stagingStatus: status.stagingStatus,
          untrackedCount: status.untrackedCount,
          pushoverStatus: status.pushoverStatus,
          pushStatus: status.pushStatus,
          lastUpdated: new Date(status.lastUpdated).getTime(),
          loading: false,
          error: null,
          stale: false
        });
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
    try {
      // 添加超时保护，防止预加载卡住
      const timeoutPromise = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('StatusCache init timeout')), 10000)
      )

      // 从 projectStore 获取所有项目路径
      const { useProjectStore } = await import('./projectStore')
      const projectStore = useProjectStore()
      const paths = projectStore.projects.map(p => p.path)

      // 竞速：预加载 vs 超时
      await Promise.race([preload(paths), timeoutPromise])
    } catch (error) {
      console.warn('StatusCache 初始化失败或超时，将在使用时懒加载:', error)
      // 不抛出错误，允许应用继续运行
    }
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

  /**
   * 获取项目的 Pushover 状态
   * @param path 项目路径
   * @returns Pushover Hook 状态，如果不存在则返回 null
   */
  function getPushoverStatus(path: string): HookStatus | null {
    const cached = cache.value[path]
    return cached?.pushoverStatus || null
  }

  /**
   * 获取项目的推送状态
   * @param path 项目路径
   * @returns 推送状态，如果不存在则返回 null
   */
  function getPushStatus(path: string): any | null {
    const cached = cache.value[path]
    return cached?.pushStatus || null
  }

  // ========== 初始化 ==========

  // 直接初始化事件监听器（Pinia store 是单例，不需要 onMounted）
  initEventListeners()

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
    getPushoverStatus,
    getPushStatus,
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
    manualRefresh,

    // 数据验证层
    isConsistent,
    normalizeCache,
    validateAndFix,
    validateAndFixAll,
    getHealthStatus
  }
})
