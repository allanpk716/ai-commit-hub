export interface GitProject {
  id: number
  path: string
  name: string
  sort_order: number
  created_at?: string
  updated_at?: string

  // 项目 AI 配置（可选）
  provider?: string | null      // null 表示使用默认
  language?: string | null      // null 表示使用默认
  model?: string | null         // null 表示使用默认
  use_default?: boolean         // true 表示使用默认配置

  // Pushover Hook 配置
  hook_installed?: boolean
  notification_mode?: string
  hook_version?: string
  hook_installed_at?: string

  // 运行时状态字段（由 GetProjectsWithStatus 填充）
  has_uncommitted_changes?: boolean   // 是否有未提交更改
  untracked_count?: number            // 未跟踪文件数量
  pushover_needs_update?: boolean     // Pushover 插件是否需要更新
}

export interface ProjectAIConfig {
  provider: string
  language: string
  model?: string
  isDefault: boolean
}

export interface ProjectInfo {
  branch: string
  files_changed: number
  has_staged: boolean
  path: string
  name: string
}

export interface StagedFile {
  path: string
  status: string // 'Modified' | 'New' | 'Deleted' | 'Renamed'
  ignored: boolean // 是否被 .gitignore 忽略
}

export interface UntrackedFile {
  path: string
}

export interface ProjectStatus {
  branch: string
  staged_files: StagedFile[]
  has_staged: boolean
}

export interface StagingStatus {
  staged: StagedFile[]
  unstaged: StagedFile[]
  untracked: UntrackedFile[]
}

export interface CommitHistory {
  id: number
  project_id: number
  message: string
  provider: string
  language: string
  created_at: string
  project?: GitProject
}

// Provider 配置信息
export interface ProviderInfo {
  name: string           // provider 名称，如 'openai'
  configured: boolean    // 是否已配置
  reason?: string        // 未配置的原因
}

// 排除模式
export type ExcludeMode = 'exact' | 'extension' | 'directory'

// 重新导出 Pushover 相关类型（从 types/pushover.ts）
export type {
  HookStatus,
  ExtensionInfo,
  InstallResult,
  NotificationMode,
  PushoverConfigStatus
} from './pushover'

// 目录选项
export interface DirectoryOption {
  pattern: string
  label: string
}
