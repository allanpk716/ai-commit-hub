# Frontend Events Documentation

## 概述

本文档描述了 AI Commit Hub 前端使用的所有 Wails 事件。

## 事件常量

所有事件名称定义在 `frontend/src/constants/events.ts` 中：

```typescript
export const APP_EVENTS = {
  STARTUP_COMPLETE: 'startup:complete',
  WINDOW_SHOWN: 'window:shown',
  WINDOW_HIDDEN: 'window:hidden',
  COMMIT_DELTA: 'commit:delta',
  COMMIT_COMPLETE: 'commit:complete',
  COMMIT_ERROR: 'commit:error',
  PROJECT_STATUS_CHANGED: 'project:status-changed',
  PROJECT_HOOK_UPDATED: 'project:hook-updated',
  PUSHOVER_STATUS_CHANGED: 'pushover:status-changed',
} as const
```

**使用建议**: 始终使用 `APP_EVENTS` 常量而非硬编码字符串，以避免拼写错误并便于重构。

---

## 应用生命周期事件

### startup:complete

应用启动完成。

**数据类型:**
```typescript
interface StartupCompleteData {
  success?: boolean
  statuses?: Record<string, ProjectStatusCache>
}
```

**用途:**
- 隐藏启动画面
- 填充项目状态缓存
- 标记应用就绪

**监听示例:**
```typescript
import { APP_EVENTS } from '@/constants/events'
import { EventsOn } from '@/wailsjs/runtime'

onMounted(() => {
  EventsOn(APP_EVENTS.STARTUP_COMPLETE, (data: StartupCompleteData) => {
    if (data?.success && data?.statuses) {
      // 填充 StatusCache
      const statusCache = useStatusCache()
      for (const [path, status] of Object.entries(data.statuses)) {
        statusCache.updateCache(path, status)
      }
    }
    // 隐藏启动画面
    showSplash.value = false
  })
})
```

---

### window:shown

窗口已显示。

**数据类型:**
```typescript
// 无数据
```

**用途:**
- 更新 UI 状态
- 恢复窗口位置

**监听示例:**
```typescript
EventsOn(APP_EVENTS.WINDOW_SHOWN, () => {
  isWindowVisible.value = true
})
```

---

### window:hidden

窗口已隐藏（最小化到托盘）。

**数据类型:**
```typescript
// 无数据
```

**用途:**
- 更新 UI 状态
- 暂停非必要操作

**监听示例:**
```typescript
EventsOn(APP_EVENTS.WINDOW_HIDDEN, () => {
  isWindowVisible.value = false
})
```

---

## Commit 生成事件

### commit:delta

Commit 消息流式输出。

**数据类型:**
```typescript
type CommitDeltaData = string  // commit 消息片段
```

**用途:**
- 实时显示 AI 生成的 commit 消息
- 提供即时反馈

**监听示例:**
```typescript
EventsOn(APP_EVENTS.COMMIT_DELTA, (delta: string) => {
  commitMessage.value += delta
  // 自动滚动到底部
  nextTick(() => {
    outputContainer.value?.scrollIntoView({ behavior: 'smooth' })
  })
})
```

**特点:**
- 多次触发，每次传递一部分文本
- 最终组合成完整的 commit 消息
- 流式输出提供更好的用户体验

---

### commit:complete

Commit 消息生成完成。

**数据类型:**
```typescript
interface CommitCompleteData {
  success: boolean
  error?: string
}
```

**用途:**
- 停止加载状态
- 显示错误信息
- 启用提交按钮

**监听示例:**
```typescript
EventsOn(APP_EVENTS.COMMIT_COMPLETE, (data: CommitCompleteData) => {
  isGenerating.value = false

  if (!data.success) {
    commitError.value = data.error || '生成失败'
    showErrorNotification(data.error)
  } else {
    showSuccessNotification('Commit 消息生成完成')
  }
})
```

---

### commit:error

Commit 生成错误（可选事件，用于错误通知）。

**数据类型:**
```typescript
interface CommitErrorData {
  error: string
}
```

**用途:**
- 显示错误通知
- 记录错误日志

**监听示例:**
```typescript
EventsOn(APP_EVENTS.COMMIT_ERROR, (data: CommitErrorData) => {
  console.error('Commit generation error:', data.error)
  showErrorNotification(data.error)
})
```

**注意:** 此事件可能不总是触发，错误通常通过 `commit:complete` 传递。

---

## 项目状态事件

### project:status-changed

项目状态已变更。

**数据类型:**
```typescript
interface ProjectStatusChangedData {
  projectPath: string
}
```

**用途:**
- 刷新项目状态
- 使缓存失效
- 触发 UI 更新

**监听示例:**
```typescript
EventsOn(APP_EVENTS.PROJECT_STATUS_CHANGED, async (data: ProjectStatusChangedData) => {
  const statusCache = useStatusCache()

  // 强制刷新项目状态
  await statusCache.refresh(data.projectPath, { force: true })

  console.log(`Project status updated: ${data.projectPath}`)
})
```

**触发场景:**
- Git 操作完成后
- 外部 Git 修改
- 手动刷新请求

---

### project:hook-updated

项目 Hook 已更新。

**数据类型:**
```typescript
interface ProjectHookUpdatedData {
  projectPath: string
  hookStatus: HookStatus
}

interface HookStatus {
  installed: boolean
  isLatestVersion: boolean
  version?: string
  updateAvailable?: boolean
}
```

**用途:**
- 更新 Pushover Hook 状态显示
- 显示更新可用提示

**监听示例:**
```typescript
EventsOn(APP_EVENTS.PROJECT_HOOK_UPDATED, (data: ProjectHookUpdatedData) => {
  const statusCache = useStatusCache()

  // 更新缓存中的 Hook 状态
  statusCache.updateCache(data.projectPath, {
    pushoverStatus: data.hookStatus
  })

  // 显示更新提示
  if (data.hookStatus.updateAvailable) {
    showUpdateNotification(data.projectPath)
  }
})
```

---

## Pushover 事件

### pushover:status-changed

Pushover 状态已变更。

**数据类型:**
```typescript
interface PushoverStatusChangedData {
  projectPath: string
  status: PushoverStatus
}

interface PushoverStatus {
  canPush: boolean
  pushed: boolean
  aheadCount: number
  behindCount: number
}
```

**用途:**
- 更新推送按钮状态
- 显示推送状态指示器

**监听示例:**
```typescript
EventsOn(APP_EVENTS.PUSHOVER_STATUS_CHANGED, (data: PushoverStatusChangedData) => {
  const statusCache = useStatusCache()

  // 更新缓存中的推送状态
  statusCache.updateCache(data.projectPath, {
    pushStatus: data.status
  })

  // 更新 UI
  updatePushButtonState(data.projectPath, data.status.canPush)
})
```

---

## 事件使用最佳实践

### 1. 使用事件常量

始终使用 `APP_EVENTS` 常量而非硬编码字符串：

```typescript
// ✅ 正确
EventsOn(APP_EVENTS.COMMIT_DELTA, handler)

// ❌ 错误
EventsOn('commit:delta', handler)
```

### 2. 及时清理监听器

组件销毁时使用 `EventsOff` 清理监听器，避免内存泄漏：

```typescript
import { EventsOn, EventsOff } from '@/wailsjs/runtime'
import { APP_EVENTS } from '@/constants/events'

onMounted(() => {
  EventsOn(APP_EVENTS.COMMIT_DELTA, handleCommitDelta)
})

onUnmounted(() => {
  EventsOff(APP_EVENTS.COMMIT_DELTA)
})
```

### 3. 避免重复监听

检查是否已经监听过某个事件：

```typescript
const isListenerSetup = ref(false)

onMounted(() => {
  if (!isListenerSetup.value) {
    EventsOn(APP_EVENTS.COMMIT_DELTA, handleCommitDelta)
    isListenerSetup.value = true
  }
})
```

### 4. 错误处理

始终处理事件数据可能为空的情况：

```typescript
EventsOn(APP_EVENTS.COMMIT_COMPLETE, (data?: CommitCompleteData) => {
  if (!data) {
    console.warn('Commit complete event received without data')
    return
  }

  if (!data.success) {
    const errorMsg = data.error || 'Unknown error'
    handleError(errorMsg)
  }
})
```

### 5. 类型安全

使用 TypeScript 类型定义确保数据类型正确：

```typescript
interface StartupCompleteData {
  success?: boolean
  statuses?: Record<string, ProjectStatusCache>
}

EventsOn(APP_EVENTS.STARTUP_COMPLETE, (data: StartupCompleteData) => {
  // TypeScript 会检查 data 的类型
  if (data?.success) {
    // 类型安全的访问
    console.log(data.statuses)
  }
})
```

---

## 事件流程示例

### Commit 生成完整流程

```
用户点击"生成 Commit"
    ↓
前端调用 GenerateCommit(projectPath)
    ↓
后端开始生成 commit 消息
    ↓
[commit:delta] 事件（流式输出）
    ↓
[commit:delta] 事件（重复多次）
    ↓
[commit:delta] 事件（最后一次）
    ↓
[commit:complete] 事件（成功或失败）
    ↓
前端停止加载状态
```

**代码示例:**
```typescript
// 1. 开始生成
const startGeneration = async () => {
  isGenerating.value = true
  commitError.value = ''
  commitMessage.value = ''

  try {
    await app.GenerateCommit(projectPath.value)
  } catch (e) {
    isGenerating.value = false
    commitError.value = e as string
  }
}

// 2. 监听流式输出
onMounted(() => {
  EventsOn(APP_EVENTS.COMMIT_DELTA, (delta: string) => {
    commitMessage.value += delta
  })

  EventsOn(APP_EVENTS.COMMIT_COMPLETE, (data: CommitCompleteData) => {
    isGenerating.value = false
    if (!data.success) {
      commitError.value = data.error || '生成失败'
    }
  })
})

// 3. 清理
onUnmounted(() => {
  EventsOff(APP_EVENTS.COMMIT_DELTA)
  EventsOff(APP_EVENTS.COMMIT_COMPLETE)
})
```

---

## 调试技巧

### 监听所有事件

在开发环境中监听所有事件以便调试：

```typescript
if (import.meta.env.DEV) {
  const allEvents = Object.values(APP_EVENTS)

  allEvents.forEach(event => {
    EventsOn(event, (data) => {
      console.log(`[Event] ${event}:`, data)
    })
  })
}
```

### 事件计数

统计事件触发次数：

```typescript
const eventCounts = new Map<string, number>()

EventsOn(APP_EVENTS.COMMIT_DELTA, (data) => {
  eventCounts.set(APP_EVENTS.COMMIT_DELTA, (eventCounts.get(APP_EVENTS.COMMIT_DELTA) || 0) + 1)
  console.log(`commit:delta triggered ${eventCounts.get(APP_EVENTS.COMMIT_DELTA)} times`)
})
```

---

## 相关文档

- [后端 API 文档](./backend-api.md)
- [StatusCache 架构](../architecture/frontend-status-cache.md)
- [错误处理系统](../architecture/backend-errors.md)
