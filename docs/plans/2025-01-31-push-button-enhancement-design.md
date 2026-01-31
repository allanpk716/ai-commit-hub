# 推送按钮优化设计文档

## 概述

改进推送按钮的行为，使其能够实时检测本地分支与远程分支的差异状态。推送按钮将始终显示，但只有当本地分支领先于远程分支时才可点击。

## 背景

当前实现中，推送按钮只有在本地提交成功后才会显示。这导致用户无法提前知道是否有内容需要推送。改进后，推送按钮将始终可见，并根据实际的 Git 状态动态启用/禁用。

## 设计目标

1. **实时检测**: 使用 Git 命令精确计算本地领先远程的提交数量
2. **统一管理**: 推送状态集成到现有的 StatusCache 中，与其他状态同步刷新
3. **用户友好**: 按钮始终显示，清晰指示推送状态和领先提交数

## 数据结构

### TypeScript 类型定义

```typescript
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
```

### Go 结构体

```go
// PushStatus 表示推送状态
type PushStatus struct {
    CanPush      bool   `json:"canPush"`
    AheadCount   int    `json:"ahead_count"`
    RemoteBranch string `json:"remote_branch"`
    Error        string `json:"error,omitempty"`
}
```

## 后端实现

### 推送状态检测函数

在 `pkg/git/git.go` 中添加 `GetPushStatus()` 函数：

```go
// GetPushStatus 检测本地分支是否领先于远程分支
func GetPushStatus(projectPath string) (*PushStatus, error)
```

**检测逻辑**:

1. 使用 `git rev-parse --abbrev-ref --symbolic-full-name @{u}` 检查是否有远程跟踪分支
2. 如果无远程分支，返回 `canPush: false`，错误信息为"未配置远程仓库"
3. 使用 `git rev-list --count @{u}..HEAD` 计算本地领先提交数
4. 返回 `canPush: aheadCount > 0`

## Wails API 绑定

### 导出方法

在 `app.go` 中添加：

```go
// GetPushStatus 获取项目的推送状态
func (a *App) GetPushStatus(projectPath string) *git.PushStatus
```

### 批量加载更新

更新 `GetAllProjectStatuses()` 以包含推送状态：

```go
type ProjectFullStatus struct {
    // ... 现有字段 ...
    PushStatus *git.PushStatus `json:"pushStatus"`
}
```

## 前端集成

### StatusCache 更新

在 `frontend/src/stores/statusCache.ts` 中：

1. 更新 `ProjectStatusCache` 接口，添加 `pushStatus` 字段
2. 在 `refresh()` 方法的 `Promise.all` 中添加 `GetPushStatus()` 调用
3. 添加 `getPushStatus(path)` 辅助方法用于获取推送状态

### CommitPanel UI 修改

**布局变更**:

- 移除独立的推送按钮区域
- 将推送按钮放在"提交到本地"按钮右侧，与其他操作按钮在同一行

**状态控制**:

```vue
<button
  @click="handlePush"
  class="btn-action btn-primary-push"
  :disabled="isPushing || !pushStatus?.canPush"
  :title="pushStatus?.aheadCount ? `领先 ${pushStatus.aheadCount} 个提交` : '无待推送内容'"
>
  <span class="icon" :class="{ spin: isPushing }">↑</span>
  {{ isPushing ? '推送中...' : `推送${pushStatus?.aheadCount ? ` (${pushStatus.aheadCount})` : ''}` }}
</button>
```

**状态来源**:

```typescript
const pushStatus = computed(() => {
  if (currentProjectPath.value) {
    return statusCache.getPushStatus(currentProjectPath.value)
  }
  return null
})
```

## 变更文件清单

1. `frontend/src/types/status.ts` - 添加 `PushStatus` 类型
2. `pkg/git/git.go` - 添加 `GetPushStatus()` 函数
3. `app.go` - 添加 `GetPushStatus()` API 方法
4. `frontend/src/stores/statusCache.ts` - 集成推送状态到缓存
5. `frontend/src/components/CommitPanel.vue` - UI 修改

## 测试要点

1. **无远程仓库**: 确认按钮禁用，tooltip 显示错误信息
2. **领先远程**: 确认按钮启用，显示领先提交数
3. **与远程同步**: 确认按钮禁用，无领先提示
4. **提交后刷新**: 确认本地提交后推送状态正确更新
5. **推送后刷新**: 确认推送后按钮禁用状态正确更新

## 后续优化

1. 考虑添加推送进度指示
2. 支持推送到不同远程分支
3. 添加推送历史记录
