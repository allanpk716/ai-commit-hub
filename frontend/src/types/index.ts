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
