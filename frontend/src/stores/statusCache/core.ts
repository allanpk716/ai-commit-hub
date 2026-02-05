import { ref, Ref } from "vue";
import type { ProjectStatusCache } from "@/types/status";

/**
 * StatusCacheCore - 核心缓存功能
 * 负责基本的缓存存储和检索
 */
export class StatusCacheCore {
  private cache: Ref<Record<string, ProjectStatusCache>>;

  constructor() {
    this.cache = ref<Record<string, ProjectStatusCache>>({});
  }

  /**
   * 获取项目状态缓存
   */
  getStatus(path: string): ProjectStatusCache | null {
    return this.cache.value[path] || null;
  }

  /**
   * 获取多个项目的状态缓存
   */
  getStatuses(paths: string[]): Record<string, ProjectStatusCache> {
    const result: Record<string, ProjectStatusCache> = {};

    for (const path of paths) {
      const status = this.getStatus(path);
      if (status) {
        result[path] = status;
      }
    }

    return result;
  }

  /**
   * 更新项目状态缓存
   */
  updateCache(path: string, status: Partial<ProjectStatusCache>): void {
    const existing = this.getStatus(path) || this.createEmptyCache();
    this.cache.value[path] = {
      ...existing,
      ...status,
      lastUpdated: Date.now(),
    };
  }

  /**
   * 批量更新缓存
   */
  updateCacheBatch(
    statuses: Record<string, Partial<ProjectStatusCache>>,
  ): void {
    for (const [path, status] of Object.entries(statuses)) {
      this.updateCache(path, status);
    }
  }

  /**
   * 清除项目缓存
   */
  clearCache(path: string): void {
    delete this.cache.value[path];
  }

  /**
   * 清除所有缓存
   */
  clearAll(): void {
    this.cache.value = {};
  }

  /**
   * 创建空缓存对象
   */
  private createEmptyCache(): ProjectStatusCache {
    return {
      gitStatus: null,
      stagingStatus: null,
      pushoverStatus: null,
      pushStatus: null,
      untrackedCount: 0,
      lastUpdated: 0,
      loading: false,
      error: null,
      stale: false,
    };
  }

  /**
   * 获取所有缓存（用于调试）
   */
  getAllCache(): Record<string, ProjectStatusCache> {
    return { ...this.cache.value };
  }

  /**
   * 检查缓存是否存在
   */
  has(path: string): boolean {
    return path in this.cache.value;
  }

  /**
   * 获取缓存数量
   */
  get size(): number {
    return Object.keys(this.cache.value).length;
  }
}
