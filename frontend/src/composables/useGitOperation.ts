import { ref } from "vue";
import { useStatusCache } from "@/stores/statusCache";
import type { GitStatus, StagingStatus, PushStatus } from "@/types/status";

/**
 * Git 操作结果接口
 */
export interface GitOperationResult<T = void> {
  success: boolean;
  data?: T;
  error?: Error;
  rollback?: () => void;
}

/**
 * Git 操作选项接口
 */
export interface GitOperationOptions {
  optimisticUpdate?: Partial<{
    hasUncommittedChanges: boolean;
    untrackedCount: number;
    lastCommitTime: number;
    branch: string;
  }>;
  refreshOnSuccess?: boolean;
  silent?: boolean;
}

/**
 * Git 操作封装 composable
 * 统一处理 Git 操作、乐观更新和错误恢复
 */
export function useGitOperation() {
  const statusCache = useStatusCache();
  const isOperating = ref(false);

  /**
   * 执行单个 Git 操作
   */
  async function executeGitOperation<T = void>(
    operation: () => Promise<T>,
    projectPath: string,
    options: GitOperationOptions = {},
  ): Promise<GitOperationResult<T>> {
    const {
      optimisticUpdate,
      refreshOnSuccess = true,
      silent = false,
    } = options;

    isOperating.value = true;
    let rollback: (() => void) | undefined;

    try {
      // 1. 乐观更新（如果指定）
      if (optimisticUpdate) {
        rollback = statusCache.updateOptimistic(projectPath, optimisticUpdate);
      }

      // 2. 执行操作
      const data = await operation();

      // 3. 成功后刷新状态（如果需要）
      if (refreshOnSuccess) {
        await statusCache.refresh(projectPath, { force: true, silent });
      }

      return {
        success: true,
        data,
        rollback,
      };
    } catch (error) {
      // 4. 失败时回滚
      rollback?.();

      return {
        success: false,
        error: error instanceof Error ? error : new Error(String(error)),
        rollback,
      };
    } finally {
      isOperating.value = false;
    }
  }

  /**
   * 批量执行 Git 操作（并行）
   */
  async function executeBatch<T = void>(
    operations: Array<{
      fn: () => Promise<T>;
      projectPath: string;
      options?: GitOperationOptions;
    }>,
  ): Promise<GitOperationResult<T>[]> {
    isOperating.value = true;

    try {
      const results = await Promise.all(
        operations.map((op) =>
          executeGitOperation(op.fn, op.projectPath, op.options),
        ),
      );

      return results;
    } finally {
      isOperating.value = false;
    }
  }

  /**
   * 获取项目状态（带缓存）
   */
  function getStatus(projectPath: string) {
    return statusCache.getStatus(projectPath);
  }

  /**
   * 获取 Git 状态
   */
  function getGitStatus(projectPath: string): GitStatus | null {
    return statusCache.getGitStatus(projectPath);
  }

  /**
   * 获取暂存区状态
   */
  function getStagingStatus(projectPath: string): StagingStatus | null {
    return statusCache.getStagingStatus(projectPath);
  }

  /**
   * 获取推送状态
   */
  function getPushStatus(projectPath: string): PushStatus | null {
    return statusCache.getPushStatus(projectPath);
  }

  /**
   * 刷新项目状态
   */
  async function refreshStatus(
    projectPath: string,
    options?: { force?: boolean; silent?: boolean },
  ) {
    return statusCache.refresh(projectPath, options);
  }

  return {
    isOperating,
    executeGitOperation,
    executeBatch,
    getStatus,
    getGitStatus,
    getStagingStatus,
    getPushStatus,
    refreshStatus,
  };
}
