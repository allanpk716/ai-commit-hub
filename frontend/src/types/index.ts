export interface GitProject {
  id: number
  path: string
  name: string
  sort_order: number
  created_at?: string
  updated_at?: string
}

export interface ProjectInfo {
  branch: string
  files_changed: number
  has_staged: boolean
  path: string
  name: string
}
