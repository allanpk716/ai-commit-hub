# 单项目状态更新优化设计

**日期**: 2026-01-29
**状态**: 设计完成，待实现
**优先级**: 高（性能优化）

## 问题描述

当前实现中，每次项目状态变化都会刷新整个项目列表，这对于大量项目（20+个）的场景性能低下。

## 解决方案

采用**单项目状态更新 + 防抖**的混合方案：

- 前端维护项目状态缓存
- 新增 `GetSingleProjectStatus` API
- 使用 300ms 防抖合并频繁操作
- 增量更新受影响的项目状态

## 架构设计

### 数据流

```
用户操作（暂存/提交等）
     ↓
commitStore 操作
     ↓
EventsEmit('project-status-changed', { path, operation })
     ↓
ProjectList 监听事件
     ↓
防抖处理（300ms）
     ↓
调用增量更新 API
     ↓
只更新受影响项目的状态
```

## 后端实现

### 新增 API

**方法**: `GetSingleProjectStatus(projectPath string)`

**返回结构**:
```go
type SingleProjectStatus struct {
    Path                 string
    HasUncommittedChanges bool
    UntrackedCount       int
    PushoverNeedsUpdate  bool
}
```

**实现逻辑**:
1. 从数据库查询项目基本信息
2. 检查 Pushover 更新状态（如果已安装）
3. 检查 Git 状态（暂存区、未跟踪文件）
4. 返回单个项目的运行时状态

## 前端实现

### 1. ProjectList.vue

添加增量更新方法：

```typescript
async function updateSingleProjectStatus(projectPath: string) {
  const status = await GetSingleProjectStatus(projectPath)
  const project = projectStore.projects.find(p => p.path === projectPath)
  if (project) {
    project.has_uncommitted_changes = status.has_uncommitted_changes
    project.untracked_count = status.untracked_count
    project.pushover_needs_update = status.pushover_needs_update
  }
}
```

### 2. commitStore.ts

修改事件通知，携带项目路径：

```typescript
function notifyProjectStatusChanged() {
  if (!selectedProjectPath.value) return
  EventsEmit('project-status-changed', {
    path: selectedProjectPath.value
  })
}
```

### 3. 防抖处理

使用 300ms 防抖：

```typescript
const debouncedUpdate = debounce((projectPath: string) => {
  updateSingleProjectStatus(projectPath)
}, 300)
```

## 性能对比

| 场景 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| 20 个项目，操作 1 个 | O(20) | O(1) | 20x |
| 50 个项目，操作 1 个 | O(50) | O(1) | 50x |
| 频繁操作（10次/秒） | 10 次全量刷新 | 1 次单项目更新 | ~200x |

## 实现步骤

1. 后端：添加 `GetSingleProjectStatus` API
2. 前端：ProjectList 添加增量更新逻辑
3. 前端：commitStore 事件携带路径
4. 前端：实现防抖机制
5. 测试：验证性能提升和正确性

## 兼容性

- 保留原有的 `loadProjectsWithStatus()` 方法
- 启动时仍使用全量加载
- 运行时操作使用增量更新
- 向后兼容，无破坏性变更
