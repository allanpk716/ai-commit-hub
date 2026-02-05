import { describe, bench } from 'vitest'
import { StatusCacheCore } from '@/stores/statusCache/core'
import type { ProjectStatusCache } from '@/types/status'

describe('StatusCache Performance', () => {
  const core = new StatusCacheCore()

  // 准备测试数据
  const testProjects = Array.from({ length: 100 }, (_, i) => ({
    path: `/path/to/project${i}`,
    status: {
      gitStatus: {
        branch: 'main',
        hasUncommittedChanges: i % 2 === 0,
        lastCommitHash: `abc${i}`,
        lastCommitTime: Date.now() - i * 1000,
      },
      stagingStatus: {
        staged: [
          { path: `file${i}.txt`, status: 'M' }
        ],
        unstaged: []
      },
      untrackedCount: i % 3,
      pushoverStatus: {
        installed: i % 2 === 0,
        isLatestVersion: true
      },
      pushStatus: {
        canPush: i % 2 === 0,
        pushed: i % 3 === 0,
        aheadCount: i,
        behindCount: 0
      },
      lastUpdated: Date.now(),
      loading: false,
      error: null,
      stale: false,
    } as ProjectStatusCache
  }))

  // 预填充缓存
  beforeAll(() => {
    const cacheMap: Record<string, ProjectStatusCache> = {}
    testProjects.forEach(p => {
      cacheMap[p.path] = p.status
    })
    core.updateCacheBatch(cacheMap)
  })

  bench('getStatus - single lookup', () => {
    core.getStatus('/path/to/project50')
  })

  bench('updateCache - single update', () => {
    core.updateCache('/path/to/project0', testProjects[0].status)
  })

  bench('getStatuses - batch lookup (10 items)', () => {
    const paths = testProjects.slice(0, 10).map(p => p.path)
    core.getStatuses(paths)
  })

  bench('getStatuses - batch lookup (50 items)', () => {
    const paths = testProjects.slice(0, 50).map(p => p.path)
    core.getStatuses(paths)
  })

  bench('getStatuses - batch lookup (100 items)', () => {
    const paths = testProjects.map(p => p.path)
    core.getStatuses(paths)
  })

  bench('updateCacheBatch - batch update (10 items)', () => {
    const statuses: Record<string, ProjectStatusCache> = {}
    testProjects.slice(0, 10).forEach(p => {
      statuses[p.path] = p.status
    })
    core.updateCacheBatch(statuses)
  })

  bench('updateCacheBatch - batch update (50 items)', () => {
    const statuses: Record<string, ProjectStatusCache> = {}
    testProjects.slice(0, 50).forEach(p => {
      statuses[p.path] = p.status
    })
    core.updateCacheBatch(statuses)
  })

  bench('updateCacheBatch - batch update (100 items)', () => {
    const statuses: Record<string, ProjectStatusCache> = {}
    testProjects.forEach(p => {
      statuses[p.path] = p.status
    })
    core.updateCacheBatch(statuses)
  })

  bench('clearCache - clear all', () => {
    core.clearAll()
  })
})
