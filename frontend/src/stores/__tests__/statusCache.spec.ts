import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStatusCache } from '../statusCache'
import type { ProjectStatusCache } from '../../types/status'

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn()
}))

// Mock App API
const mockGetProjectStatus = vi.fn()
const mockGetStagingStatus = vi.fn()
const mockGetUntrackedFiles = vi.fn()
const mockGetPushoverHookStatus = vi.fn()
const mockGetAllProjectStatuses = vi.fn()

vi.mock('../../../wailsjs/go/main/App', () => ({
  GetStagingStatus: () => mockGetStagingStatus(),
  GetProjectStatus: () => mockGetProjectStatus(),
  GetUntrackedFiles: () => mockGetUntrackedFiles(),
  GetPushoverHookStatus: () => mockGetPushoverHookStatus(),
  GetAllProjectStatuses: () => mockGetAllProjectStatuses()
}))

describe('StatusCache Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  describe('缓存项目状态', () => {
    it('应该正确缓存项目状态', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'
      const mockStatus: ProjectStatusCache = {
        gitStatus: { branch: 'main' },
        stagingStatus: { hasChanges: true, stagedCount: 2 },
        untrackedCount: 3,
        pushoverStatus: { enabled: true, version: '1.0.0' },
        lastUpdated: Date.now(),
        loading: false,
        error: null,
        stale: false
      }

      mockGetProjectStatus.mockResolvedValue(mockStatus.gitStatus)
      mockGetStagingStatus.mockResolvedValue(mockStatus.stagingStatus)
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(mockStatus.pushoverStatus)

      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()

      const cached = store.getStatus(testPath)
      expect(cached).toBeTruthy()
      expect(cached?.gitStatus?.branch).toBe('main')
      expect(cached?.stagingStatus?.hasChanges).toBe(true)
      expect(cached?.pushoverStatus?.enabled).toBe(true)
      expect(cached?.error).toBeNull()
    })

    it('应该返回 null 对于不存在的项目', () => {
      const store = useStatusCache()
      const status = store.getStatus('/nonexistent')
      expect(status).toBeNull()
    })

    it('应该正确初始化缓存条目', () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      store.initCache(testPath)
      const cached = store.getStatus(testPath)

      expect(cached).toBeTruthy()
      expect(cached?.gitStatus).toBeNull()
      expect(cached?.stagingStatus).toBeNull()
      expect(cached?.untrackedCount).toBe(0)
      expect(cached?.loading).toBe(false)
      expect(cached?.stale).toBe(true)
    })
  })

  describe('防止并发重复请求', () => {
    it('应该防止同时发起相同项目的多个请求', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      // Mock API 延迟
      mockGetProjectStatus.mockImplementation(() =>
        new Promise(resolve => setTimeout(() => resolve({ branch: 'main' }), 100))
      )
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      // 同时发起三个请求
      const promise1 = store.refresh(testPath)
      const promise2 = store.refresh(testPath)
      const promise3 = store.refresh(testPath)

      await Promise.all([promise1, promise2, promise3])
      await vi.runAllTimersAsync()

      // 应该只调用一次 API
      expect(mockGetProjectStatus).toHaveBeenCalledTimes(1)
      expect(mockGetStagingStatus).toHaveBeenCalledTimes(1)
    })

    it('应该在第一个请求完成后允许新的请求', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      let callCount = 0
      mockGetProjectStatus.mockImplementation(() => {
        callCount++
        return Promise.resolve({ branch: 'main' })
      })
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      // 第一个请求
      await store.refresh(testPath)
      await vi.runAllTimersAsync()

      // 等待一段时间
      vi.advanceTimersByTime(1000)

      // 第二个请求（应该执行）
      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()

      expect(callCount).toBe(2)
    })

    it('应该正确跟踪加载状态', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      mockGetProjectStatus.mockImplementation(() =>
        new Promise(resolve => setTimeout(() => resolve({ branch: 'main' }), 100))
      )
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      // 开始加载
      const promise = store.refresh(testPath)

      expect(store.isLoading(testPath)).toBe(true)
      expect(store.getStatus(testPath)?.loading).toBe(true)

      await promise
      await vi.runAllTimersAsync()

      expect(store.isLoading(testPath)).toBe(false)
      expect(store.getStatus(testPath)?.loading).toBe(false)
    })
  })

  describe('处理 TTL 过期', () => {
    it('应该正确判断缓存是否过期', () => {
      const store = useStatusCache()
      const testPath = '/test/project'
      const customTTL = 5000 // 5 秒

      store.updateOptions({ ttl: customTTL })
      store.initCache(testPath)

      // 刚创建的缓存不应该过期
      expect(store.isExpired(testPath)).toBe(false)

      // 前进时间但未过期
      vi.advanceTimersByTime(customTTL - 100)
      expect(store.isExpired(testPath)).toBe(false)

      // 超过 TTL
      vi.advanceTimersByTime(200)
      expect(store.isExpired(testPath)).toBe(true)
    })

    it('应该在过期时自动刷新', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'
      const customTTL = 2000 // 2 秒

      store.updateOptions({ ttl: customTTL })

      let callCount = 0
      mockGetProjectStatus.mockImplementation(() => {
        callCount++
        return Promise.resolve({ branch: 'main' })
      })
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      // 首次加载
      await store.getStatusOrRefresh(testPath)
      expect(callCount).toBe(1)

      // 未过期，不应该刷新
      await store.getStatusOrRefresh(testPath)
      expect(callCount).toBe(1)

      // 时间流逝导致缓存过期
      vi.advanceTimersByTime(customTTL + 100)

      // 应该自动刷新
      await store.getStatusOrRefresh(testPath)
      await vi.runAllTimersAsync()
      expect(callCount).toBe(2)
    })

    it('应该支持强制刷新忽略 TTL', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      let callCount = 0
      mockGetProjectStatus.mockImplementation(() => {
        callCount++
        return Promise.resolve({ branch: 'main' })
      })
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      // 首次加载
      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()
      expect(callCount).toBe(1)

      // 立即强制刷新（忽略 TTL）
      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()
      expect(callCount).toBe(2)
    })

    it('应该提供过期路径列表', () => {
      const store = useStatusCache()
      const customTTL = 2000

      store.updateOptions({ ttl: customTTL })

      const path1 = '/test/project1'
      const path2 = '/test/project2'
      const path3 = '/test/project3'

      store.initCache(path1)
      store.initCache(path2)
      store.initCache(path3)

      // 都未过期
      expect(store.expiredPaths).toHaveLength(0)

      // path1 过期
      vi.advanceTimersByTime(customTTL + 100)
      store.updateCache(path1, { lastUpdated: Date.now() - customTTL - 200 })

      const expired = store.expiredPaths
      expect(expired).toContain(path1)
      expect(expired).not.toContain(path2)
      expect(expired).not.toContain(path3)
    })
  })

  describe('批量预加载', () => {
    it('应该批量预加载多个项目', async () => {
      const store = useStatusCache()
      const paths = ['/project1', '/project2', '/project3']

      const mockStatuses = {
        '/project1': {
          gitStatus: { branch: 'main' },
          stagingStatus: { hasChanges: true, stagedCount: 1 },
          untrackedCount: 0,
          pushoverStatus: null,
          lastUpdated: new Date().toISOString()
        },
        '/project2': {
          gitStatus: { branch: 'develop' },
          stagingStatus: { hasChanges: false, stagedCount: 0 },
          untrackedCount: 2,
          pushoverStatus: { enabled: true, version: '1.0.0' },
          lastUpdated: new Date().toISOString()
        },
        '/project3': {
          gitStatus: { branch: 'feature' },
          stagingStatus: { hasChanges: false, stagedCount: 0 },
          untrackedCount: 0,
          pushoverStatus: null,
          lastUpdated: new Date().toISOString()
        }
      }

      mockGetAllProjectStatuses.mockResolvedValue(mockStatuses)

      await store.preload(paths)
      await vi.runAllTimersAsync()

      expect(mockGetAllProjectStatuses).toHaveBeenCalledWith(paths)
      expect(store.getStatus('/project1')?.gitStatus?.branch).toBe('main')
      expect(store.getStatus('/project2')?.gitStatus?.branch).toBe('develop')
      expect(store.getStatus('/project3')?.gitStatus?.branch).toBe('feature')
    })

    it('应该在批量接口失败时降级到逐个加载', async () => {
      const store = useStatusCache()
      const paths = ['/project1', '/project2']

      mockGetAllProjectStatuses.mockRejectedValue(new Error('Batch API failed'))

      mockGetProjectStatus.mockImplementation((path: string) =>
        Promise.resolve({ branch: path.split('/').pop() })
      )
      mockGetStagingStatus.mockResolvedValue({ hasChanges: false, stagedCount: 0 })
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      await store.preload(paths)
      await vi.runAllTimersAsync()

      // 应该降级到逐个加载
      expect(mockGetProjectStatus).toHaveBeenCalledTimes(2)
      expect(store.getStatus('/project1')?.gitStatus?.branch).toBe('project1')
      expect(store.getStatus('/project2')?.gitStatus?.branch).toBe('project2')
    })
  })

  describe('乐观更新', () => {
    it('应该支持乐观更新并可以回滚', () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      // 初始状态
      const initialStatus: ProjectStatusCache = {
        gitStatus: { branch: 'main' },
        stagingStatus: { hasChanges: true, stagedCount: 2 },
        untrackedCount: 1,
        pushoverStatus: null,
        lastUpdated: Date.now(),
        loading: false,
        error: null,
        stale: false
      }

      store.cache[testPath] = initialStatus

      // 乐观更新
      const rollback = store.updateOptimistic(testPath, {
        stagingStatus: { hasChanges: false, stagedCount: 0 }
      })

      expect(store.getStatus(testPath)?.stagingStatus?.hasChanges).toBe(false)

      // 回滚
      rollback?.()
      expect(store.getStatus(testPath)?.stagingStatus?.hasChanges).toBe(true)
    })

    it('应该在缓存不存在时返回 undefined', () => {
      const store = useStatusCache()
      const rollback = store.updateOptimistic('/nonexistent', {
        stagingStatus: { hasChanges: false, stagedCount: 0 }
      })

      expect(rollback).toBeUndefined()
    })
  })

  describe('缓存失效和清理', () => {
    it('应该支持单个项目失效', () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      store.initCache(testPath)
      store.updateCache(testPath, { stale: false })

      expect(store.getStatus(testPath)?.stale).toBe(false)

      store.invalidate(testPath)
      expect(store.getStatus(testPath)?.stale).toBe(true)
    })

    it('应该支持所有项目失效', () => {
      const store = useStatusCache()

      store.initCache('/project1')
      store.initCache('/project2')
      store.updateCache('/project1', { stale: false })
      store.updateCache('/project2', { stale: false })

      store.invalidateAll()

      expect(store.getStatus('/project1')?.stale).toBe(true)
      expect(store.getStatus('/project2')?.stale).toBe(true)
    })

    it('应该支持清除单个缓存', () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      store.initCache(testPath)
      expect(store.getStatus(testPath)).toBeTruthy()

      store.clearCache(testPath)
      expect(store.getStatus(testPath)).toBeNull()
    })

    it('应该支持清除所有缓存', () => {
      const store = useStatusCache()

      store.initCache('/project1')
      store.initCache('/project2')

      expect(store.cachedPaths).toHaveLength(2)

      store.clearAllCache()
      expect(store.cachedPaths).toHaveLength(0)
    })
  })

  describe('错误处理', () => {
    it('应该正确处理 API 错误', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      mockGetProjectStatus.mockRejectedValue(new Error('Network error'))
      mockGetStagingStatus.mockResolvedValue(null)
      mockGetUntrackedFiles.mockResolvedValue([])
      mockGetPushoverHookStatus.mockResolvedValue(null)

      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()

      const status = store.getStatus(testPath)
      expect(status?.error).toContain('Network error')
      expect(status?.loading).toBe(false)
    })

    it('应该部分处理失败的 API 调用', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      mockGetProjectStatus.mockResolvedValue({ branch: 'main' })
      mockGetStagingStatus.mockRejectedValue(new Error('Staging error'))
      mockGetUntrackedFiles.mockResolvedValue(['file1.ts', 'file2.ts'])
      mockGetPushoverHookStatus.mockResolvedValue({ enabled: true, version: '1.0.0' })

      await store.refresh(testPath, { force: true })
      await vi.runAllTimersAsync()

      const status = store.getStatus(testPath)
      expect(status?.gitStatus?.branch).toBe('main')
      expect(status?.untrackedCount).toBe(2)
      expect(status?.pushoverStatus?.enabled).toBe(true)
      // stagingStatus 应该为 null 因为失败了
      expect(status?.stagingStatus).toBeNull()
    })
  })

  describe('缓存配置', () => {
    it('应该支持更新 TTL 配置', () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      store.updateOptions({ ttl: 5000 })
      store.initCache(testPath)

      expect(store.isExpired(testPath)).toBe(false)

      vi.advanceTimersByTime(6000)
      expect(store.isExpired(testPath)).toBe(true)

      // 更新配置
      store.updateOptions({ ttl: 10000 })
      store.updateCache(testPath, { lastUpdated: Date.now() })

      expect(store.isExpired(testPath)).toBe(false)

      vi.advanceTimersByTime(11000)
      expect(store.isExpired(testPath)).toBe(true)
    })
  })

  describe('计算属性', () => {
    it('应该正确计算已缓存路径', () => {
      const store = useStatusCache()

      store.initCache('/project1')
      store.initCache('/project2')
      store.initCache('/project3')

      const paths = store.cachedPaths
      expect(paths).toHaveLength(3)
      expect(paths).toContain('/project1')
      expect(paths).toContain('/project2')
      expect(paths).toContain('/project3')
    })
  })
})
