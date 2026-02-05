import type {
  ProjectStatusCache,
  ProjectStatus,
  StagingStatus,
  HookStatus,
  PushStatus,
} from "@/types/status";

/**
 * StatusCacheValidation - 数据验证和修复
 * 负责验证缓存数据的完整性并修复问题
 */
export class StatusCacheValidation {
  /**
   * 验证项目状态完整性
   */
  validateProjectStatus(status: ProjectStatus | null): boolean {
    if (!status) return false;

    // 验证必需字段
    return (
      typeof status.branch === "string" &&
      typeof status.hasUncommittedChanges === "boolean"
    );
  }

  /**
   * 验证暂存区状态完整性
   */
  validateStagingStatus(status: StagingStatus | null): boolean {
    if (!status) return true; // 空 staging 状态是有效的

    return (
      Array.isArray(status.stagedFiles) && Array.isArray(status.unstagedFiles)
    );
  }

  /**
   * 验证 Hook 状态完整性
   */
  validateHookStatus(status: HookStatus | null): boolean {
    if (!status) return true;

    return (
      typeof status.installed === "boolean" &&
      typeof status.isLatestVersion === "boolean"
    );
  }

  /**
   * 验证推送状态完整性
   */
  validatePushStatus(status: PushStatus | null): boolean {
    if (!status) return true;

    return (
      typeof status.canPush === "boolean" && typeof status.pushed === "boolean"
    );
  }

  /**
   * 验证完整缓存对象
   */
  validateCache(cache: ProjectStatusCache): {
    valid: boolean;
    errors: string[];
  } {
    const errors: string[] = [];

    if (!this.validateProjectStatus(cache.gitStatus)) {
      errors.push("Invalid git status");
    }

    if (!this.validateStagingStatus(cache.stagingStatus)) {
      errors.push("Invalid staging status");
    }

    if (!this.validateHookStatus(cache.pushoverStatus)) {
      errors.push("Invalid pushover status");
    }

    if (!this.validatePushStatus(cache.pushStatus)) {
      errors.push("Invalid push status");
    }

    return {
      valid: errors.length === 0,
      errors,
    };
  }

  /**
   * 修复缓存数据
   */
  repairCache(cache: ProjectStatusCache): ProjectStatusCache {
    const repaired = { ...cache };

    // 修复 gitStatus
    if (!this.validateProjectStatus(repaired.gitStatus)) {
      repaired.gitStatus = null;
    }

    // 修复 stagingStatus
    if (!this.validateStagingStatus(repaired.stagingStatus)) {
      repaired.stagingStatus = null;
    }

    // 修复 pushoverStatus
    if (!this.validateHookStatus(repaired.pushoverStatus)) {
      repaired.pushoverStatus = null;
    }

    // 修复 pushStatus
    if (!this.validatePushStatus(repaired.pushStatus)) {
      repaired.pushStatus = null;
    }

    return repaired;
  }

  /**
   * 修复缺失的默认值
   */
  ensureDefaults(cache: ProjectStatusCache): ProjectStatusCache {
    return {
      gitStatus: cache.gitStatus || null,
      stagingStatus: cache.stagingStatus || null,
      pushoverStatus: cache.pushoverStatus || null,
      pushStatus: cache.pushStatus || null,
      untrackedCount: cache.untrackedCount ?? 0,
      lastUpdated: cache.lastUpdated || 0,
      loading: cache.loading || false,
      error: cache.error || null,
      stale: cache.stale || false,
    };
  }

  /**
   * 验证并修复缓存
   */
  validateAndRepair(cache: ProjectStatusCache): ProjectStatusCache {
    const validated = this.validateCache(cache);

    if (!validated.valid) {
      // 先尝试修复
      let repaired = this.repairCache(cache);

      // 如果修复后仍有问题，应用默认值
      const validation = this.validateCache(repaired);
      if (!validation.valid) {
        repaired = this.ensureDefaults(repaired);
      }

      return repaired;
    }

    return this.ensureDefaults(cache);
  }
}
