# 状态同步修复方案

## 问题描述

**左侧项目列表**和**右侧项目详情**显示的状态不一致，特别是：
- Pushover 更新状态
- Git 分支状态
- 暂存区状态
- Push 状态

## 根本原因

1. **左侧**：依赖应用启动时的预加载数据，**不会自动刷新**
2. **右侧**：切换项目时主动刷新，获取**实时数据**
3. **数据源不统一**，导致状态不一致

## 解决方案

采用**方案 A（定期刷新）+ 方案 C（事件驱动）**组合：

### 1. 定期刷新机制（ProjectList.vue）

**修改文件**：`frontend/src/components/ProjectList.vue`

**改动内容**：
- 添加每 30 秒定期刷新所有项目状态
- 使用静默模式（`silent: true`）避免 UI 闪烁
- 组件卸载时清理定时器

**代码片段**：
```typescript
// 定期刷新所有项目状态（每 30 秒）
const REFRESH_INTERVAL = 30000 // 30 秒，与 StatusCache TTL 一致

const refreshInterval = setInterval(async () => {
  // 静默刷新所有项目状态，避免 UI 闪烁
  for (const project of projectStore.projects) {
    try {
      await statusCache.refresh(project.path, { silent: true })
    } catch (error) {
      console.warn(`[ProjectList] 刷新项目状态失败: ${project.name}`, error)
    }
  }
  console.log('[ProjectList] 定期刷新完成，已刷新所有项目状态')
}, REFRESH_INTERVAL)
```

### 2. 优化事件监听器（statusCache.ts）

**修改文件**：`frontend/src/stores/statusCache.ts`

**改动内容**：
- 收到 `project-status-changed` 事件后**立即刷新**（而不仅仅是使缓存失效）
- 使用静默模式避免 UI 闪烁
- 兼容新旧事件格式（`projectPath` 和 `path`）

**代码片段**：
```typescript
EventsOn('project-status-changed', async (data: { projectPath?: string; path?: string; changeType?: string }) => {
  const path = data.projectPath || data.path

  if (path) {
    console.log('[StatusCache] 收到状态变化事件，立即刷新项目:', path, '类型:', data.changeType)

    // 立即刷新该项目状态（静默模式，避免 UI 闪烁）
    try {
      await refresh(path, { force: true, silent: true })
      console.log('[StatusCache] 项目状态刷新完成:', path)
    } catch (error) {
      console.warn('[StatusCache] 刷新项目状态失败:', path, error)
      // 失败时也使缓存失效，确保下次读取时会重新获取
      invalidate(path)
    }
  } else {
    console.log('[StatusCache] 收到全局状态变化事件，使所有缓存失效')
    invalidateAll()
  }
})
```

### 3. 补充推送操作事件（CommitPanel.vue）

**修改文件**：`frontend/src/components/CommitPanel.vue`

**改动内容**：
- 在推送成功后发送 `project-status-changed` 事件
- 确保项目列表能及时更新推送状态

**代码片段**：
```typescript
async function handlePush() {
  // ...
  try {
    await PushToRemote(commitStore.selectedProjectPath)
    showToast('success', '推送成功!')

    // 使用 StatusCache 刷新状态
    await statusCache.refresh(commitStore.selectedProjectPath, { force: true, silent: true })
    const fresh = statusCache.getStatus(commitStore.selectedProjectPath)
    if (fresh) {
      updateUIFromCache(fresh)
    }

    // 通知项目列表状态已更新（新增）
    EventsEmit('project-status-changed', {
      projectPath: commitStore.selectedProjectPath,
      changeType: 'push'
    })
  } catch (e) {
    // ...
  }
}
```

## 已存在的事件发送点

以下操作已经正确发送了 `project-status-changed` 事件：

1. **提交操作**（CommitPanel.vue + commitStore.ts）
2. **暂存操作**（commitStore.ts 的 `stageFile`, `unstageFile`, `stageAllFiles` 等）
3. **Pushover 安装/更新**（CommitPanel.vue）

## 工作原理

### 定期刷新流程

```
应用启动
  ↓
ProjectList 挂载
  ↓
启动定时器（每 30 秒）
  ↓
后台静默刷新所有项目状态
  ↓
项目列表自动更新显示
```

### 事件驱动流程

```
用户操作（提交/推送/暂存等）
  ↓
发送 project-status-changed 事件
  ↓
StatusCache 收到事件
  ↓
立即刷新受影响的项目状态
  ↓
项目列表自动更新显示
```

## 优点

1. ✅ **状态一致性**：左侧和右侧始终显示相同的状态
2. ✅ **用户体验好**：静默刷新，无 UI 闪烁
3. ✅ **实时响应**：关键操作后立即更新状态
4. ✅ **资源可控**：30 秒间隔 + 事件驱动，避免频繁请求
5. ✅ **错误处理**：失败时降级为使缓存失效

## 注意事项

1. **定时器清理**：组件卸载时必须清理定时器，避免内存泄漏
2. **错误处理**：每个刷新操作都包装在 try-catch 中，避免单个项目失败影响整体
3. **静默模式**：使用 `silent: true` 避免加载状态影响用户体验
4. **事件格式**：兼容新旧两种事件格式，确保向后兼容

## 测试建议

1. **功能测试**：
   - 启动应用，观察左侧列表状态是否正确显示
   - 执行提交操作，观察左侧列表是否自动更新
   - 执行推送操作，观察左侧列表是否自动更新
   - 等待 30 秒，观察控制台是否有刷新日志

2. **性能测试**：
   - 观察定时器是否正常清理
   - 检查是否有重复刷新的问题

3. **边界测试**：
   - 网络断开时的错误处理
   - 大量项目时的刷新性能

## 修改文件列表

1. `frontend/src/components/ProjectList.vue` - 添加定期刷新
2. `frontend/src/stores/statusCache.ts` - 优化事件监听器
3. `frontend/src/components/CommitPanel.vue` - 补充推送事件

## 相关 Issue

修复了左侧项目列表和右侧项目详情状态不一致的问题，特别是 Pushover 更新状态的显示差异。
