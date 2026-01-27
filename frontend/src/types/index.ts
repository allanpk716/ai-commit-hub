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
  ignored: boolean
}

export interface ProjectStatus {
  branch: string
  staged_files: StagedFile[]
  has_staged: boolean
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

// Pushover Hook 扩展信息
export interface ExtensionInfo {
  downloaded: boolean      // 是否已下载
  path: string            // 扩展路径
  version: string         // 当前版本
  current_version: string // 当前版本（同 version）
  latest_version: string  // 最新版本
  update_available: boolean // 是否有可用更新
}

// Pushover Hook 状态
export interface HookStatus {
  installed: boolean
  mode: string          // 'silent' | 'normal' | 'verbose'
  version: string       // Hook 版本
  installed_at: string  // 安装时间（ISO 8601）
}

// Hook 安装结果
export interface InstallResult {
  success: boolean
  message: string
}

// 通知模式
export type NotificationMode = 'silent' | 'normal' | 'verbose'
