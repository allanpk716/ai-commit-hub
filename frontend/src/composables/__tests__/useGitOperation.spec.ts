import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useGitOperation } from '../useGitOperation'
import { useStatusCache } from '@/stores/statusCache'
import type { ProjectStatusCache } from '@/types/status'

// Mock useStatusCache
vi.mock('@/stores/statusCache', () => ({
  useStatusCache: vi.fn(),
}))

describe('useGitOperation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should execute operation successfully', async () => {
    const mockUpdateCache = vi.fn()
    const mockRefresh = vi.fn().mockResolvedValue(undefined)
    const mockUpdateOptimistic = vi.fn().mockReturnValue(vi.fn())

    vi.mocked(useStatusCache).mockReturnValue({
      updateCache: mockUpdateCache,
      refresh: mockRefresh,
      updateOptimistic: mockUpdateOptimistic,
    } as unknown as ReturnType<typeof useStatusCache>)

    const { executeGitOperation } = useGitOperation()
    const mockOperation = vi.fn().mockResolvedValue('success')

    const result = await executeGitOperation(
      '/test/path',
      mockOperation,
      { optimisticUpdate: undefined, refreshOnSuccess: false },
    )

    expect(result.success).toBe(true)
    expect(result.data).toBe('success')
    expect(mockOperation).toHaveBeenCalledTimes(1)
  })

  it('should handle operation failure', async () => {
    const mockRefresh = vi.fn().mockResolvedValue(undefined)
    const mockUpdateOptimistic = vi.fn().mockReturnValue(vi.fn())

    vi.mocked(useStatusCache).mockReturnValue({
      refresh: mockRefresh,
      updateOptimistic: mockUpdateOptimistic,
    } as unknown as ReturnType<typeof useStatusCache>)

    const { executeGitOperation } = useGitOperation()
    const mockOperation = vi.fn().mockRejectedValue(new Error('test error'))

    const result = await executeGitOperation(
      '/test/path',
      mockOperation,
      { optimisticUpdate: undefined, refreshOnSuccess: false },
    )

    expect(result.success).toBe(false)
    expect(result.error).toBeDefined()
  })

  it('should perform optimistic update and rollback on error', async () => {
    const rollback = vi.fn()
    const mockRefresh = vi.fn().mockResolvedValue(undefined)
    const mockUpdateOptimistic = vi.fn().mockReturnValue(rollback)

    vi.mocked(useStatusCache).mockReturnValue({
      refresh: mockRefresh,
      updateOptimistic: mockUpdateOptimistic,
    } as unknown as ReturnType<typeof useStatusCache>)

    const { executeGitOperation } = useGitOperation()
    const mockOperation = vi.fn().mockRejectedValue(new Error('test error'))

    const optimisticUpdate: Partial<ProjectStatusCache> = {
      hasUncommittedChanges: false,
      untrackedCount: 0,
    }

    const result = await executeGitOperation(
      '/test/path',
      mockOperation,
      { optimisticUpdate, refreshOnSuccess: false },
    )

    expect(result.success).toBe(false)
    expect(mockUpdateOptimistic).toHaveBeenCalledWith('/test/path', optimisticUpdate)
    expect(rollback).toHaveBeenCalled()
  })
})
