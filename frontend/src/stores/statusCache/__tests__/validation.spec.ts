import { describe, it, expect } from 'vitest'
import { StatusCacheValidation } from '../validation'
import type { ProjectStatus, StagingStatus, HookStatus, PushStatus } from '@/types/status'

describe('StatusCacheValidation', () => {
  const validator = new StatusCacheValidation()

  describe('validateProjectStatus', () => {
    it('should validate correct project status', () => {
      const status: ProjectStatus = {
        branch: 'main',
        hasUncommittedChanges: true,
        lastCommitHash: 'abc123',
        lastCommitTime: 1234567890,
      }

      expect(validator.validateProjectStatus(status)).toBe(true)
    })

    it('should reject invalid project status', () => {
      const status = {
        branch: 'main',
        // missing hasUncommittedChanges
      } as unknown as ProjectStatus

      expect(validator.validateProjectStatus(status)).toBe(false)
    })

    it('should reject null status', () => {
      expect(validator.validateProjectStatus(null)).toBe(false)
    })
  })

  describe('validateStagingStatus', () => {
    it('should validate correct staging status', () => {
      const status: StagingStatus = {
        stagedFiles: [{ path: 'file1.txt', status: 'M' }],
        unstagedFiles: [{ path: 'file2.txt', status: 'M' }],
      }

      expect(validator.validateStagingStatus(status)).toBe(true)
    })

    it('should accept null staging status', () => {
      expect(validator.validateStagingStatus(null)).toBe(true)
    })

    it('should reject invalid staging status', () => {
      const status = {
        stagedFiles: 'not an array' as unknown as [],
        unstagedFiles: [],
      } as unknown as StagingStatus

      expect(validator.validateStagingStatus(status)).toBe(false)
    })
  })

  describe('validateHookStatus', () => {
    it('should validate correct hook status', () => {
      const status: HookStatus = {
        installed: true,
        isLatestVersion: true,
        version: '1.0.0',
      }

      expect(validator.validateHookStatus(status)).toBe(true)
    })

    it('should accept null hook status', () => {
      expect(validator.validateHookStatus(null)).toBe(true)
    })
  })

  describe('validatePushStatus', () => {
    it('should validate correct push status', () => {
      const status: PushStatus = {
        canPush: true,
        pushed: false,
      }

      expect(validator.validatePushStatus(status)).toBe(true)
    })

    it('should accept null push status', () => {
      expect(validator.validatePushStatus(null)).toBe(true)
    })
  })

  describe('validateCache', () => {
    it('should validate complete cache', () => {
      const cache = {
        gitStatus: {
          branch: 'main',
          hasUncommittedChanges: false,
        } as ProjectStatus,
        stagingStatus: {
          stagedFiles: [],
          unstagedFiles: [],
        } as StagingStatus,
        pushoverStatus: {
          installed: false,
          isLatestVersion: true,
        } as HookStatus,
        pushStatus: {
          canPush: false,
          pushed: false,
        } as PushStatus,
        untrackedCount: 0,
        lastUpdated: Date.now(),
        loading: false,
        error: null,
        stale: false,
      }

      const result = validator.validateCache(cache)

      expect(result.valid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should detect invalid cache', () => {
      const cache = {
        gitStatus: null,
        stagingStatus: null,
        pushoverStatus: null,
        pushStatus: null,
        untrackedCount: 'invalid' as unknown as number,
        lastUpdated: 0,
        loading: false,
        error: null,
        stale: false,
      }

      const result = validator.validateCache(cache)

      expect(result.valid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)
    })
  })

  describe('ensureDefaults', () => {
    it('should add default values', () => {
      const cache = {
        gitStatus: null,
        stagingStatus: null,
        pushoverStatus: null,
        pushStatus: null,
        untrackedCount: undefined as unknown as number,
        lastUpdated: undefined as unknown as number,
        loading: undefined as unknown as boolean,
        error: undefined,
        stale: undefined as unknown as boolean,
      }

      const result = validator.ensureDefaults(cache)

      expect(result.untrackedCount).toBe(0)
      expect(result.lastUpdated).toBe(0)
      expect(result.loading).toBe(false)
      expect(result.stale).toBe(false)
    })
  })
})
