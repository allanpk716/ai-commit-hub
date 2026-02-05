/**
 * Wails 事件名称常量
 * 所有事件名称统一管理，避免硬编码和拼写错误
 */
export const APP_EVENTS = {
  // 应用生命周期
  STARTUP_COMPLETE: "startup:complete",
  WINDOW_SHOWN: "window:shown",
  WINDOW_HIDDEN: "window:hidden",

  // Commit 生成
  COMMIT_DELTA: "commit:delta",
  COMMIT_COMPLETE: "commit:complete",
  COMMIT_ERROR: "commit:error",

  // 项目状态
  PROJECT_STATUS_CHANGED: "project:status-changed",
  PROJECT_HOOK_UPDATED: "project:hook-updated",

  // Pushover
  PUSHOVER_STATUS_CHANGED: "pushover:status-changed",
} as const;

export type AppEvent = (typeof APP_EVENTS)[keyof typeof APP_EVENTS];
