import type { StagingStatus, HookStatus as PushoverHookStatus } from './index'

/**
 * Git 状态信息（来自后端）
 */
export interface GitStatus {
  branch: string
  // 可以根据需要添加其他 GitStatus 字段
}

/**
 * 推送状态
 */
export interface PushStatus {
  /** 是否可推送（本地领先远程） */
  canPush: boolean
  /** 本地领先远程的提交数量 */
  aheadCount: number
  /** 远程分支名（如 origin/main） */
  remoteBranch: string
  /** 错误信息（无远程仓库等） */
  error?: string
}

/**
 * 项目状态缓存条目
 */
export interface ProjectStatusCache {
  /** Git 状态（分支等） */
  gitStatus: GitStatus | null
  /** 暂存区状态 */
  stagingStatus: StagingStatus | null
  /** 未跟踪文件数量 */
  untrackedCount: number
  /** Pushover Hook 状态 */
  pushoverStatus: PushoverHookStatus | null
  /** 推送状态 */
  pushStatus: PushStatus | null
  /** 最后更新时间戳（毫秒） */
  lastUpdated: number
  /** 是否正在加载 */
  loading: boolean
  /** 错误信息 */
  error: string | null
  /** 缓存是否已过期 */
  stale: boolean
}

/**
 * 项目状态缓存映射表
 * 键：项目路径
 * 值：项目状态缓存
 */
export interface ProjectStatusCacheMap {
  [projectPath: string]: ProjectStatusCache
}

/**
 * 缓存配置选项
 */
export interface CacheOptions {
  /** 缓存过期时间（毫秒），默认 30 秒 */
  ttl?: number
  /** 是否在后台刷新过期缓存 */
  backgroundRefresh?: boolean
}
