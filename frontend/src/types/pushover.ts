/**
 * Pushover Hook 相关类型定义
 */

/** 通知模式 */
export type NotificationMode = 'enabled' | 'pushover_only' | 'windows_only' | 'disabled'

/** Hook 状态信息 */
export interface HookStatus {
  installed: boolean
  mode: NotificationMode
  version: string
  installed_at?: string
}

/** 扩展信息 */
export interface ExtensionInfo {
  downloaded: boolean
  path: string
  version: string
  latest_version: string
  update_available: boolean
}

/** 安装结果 */
export interface InstallResult {
  success: boolean
  message?: string
  hook_path?: string
  version?: string
}

/** 通知模式配置 */
export interface NotificationModeConfig {
  value: NotificationMode
  label: string
  description: string
  icon: string
}

/** 预设通知模式列表 */
export const NOTIFICATION_MODES: NotificationModeConfig[] = [
  {
    value: 'enabled',
    label: '全部启用',
    description: 'Pushover 和 Windows 桌面通知都启用',
    icon: '🔔'
  },
  {
    value: 'pushover_only',
    label: '仅 Pushover',
    description: '仅使用 Pushover 推送通知',
    icon: '📱'
  },
  {
    value: 'windows_only',
    label: '仅 Windows',
    description: '仅使用 Windows 桌面通知',
    icon: '💻'
  },
  {
    value: 'disabled',
    label: '全部禁用',
    description: '不发送任何通知',
    icon: '🔕'
  }
]
