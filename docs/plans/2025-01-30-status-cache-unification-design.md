# 状态缓存统一架构设计方案

**日期**: 2025-01-30
**作者**: Claude
**状态**: 设计完成，待实现

## 问题概述

当前应用存在以下性能和体验问题：

1. **启动时状态更新延迟**：启动画面隐藏时，左侧项目列表的状态和 Pushover 版本更新状态不会立即显示
2. **切换项目时卡顿**：点击项目切换时，Git 和 Pushover 状态刷新有明显延迟，用户体验不流畅
3. **状态源分散**：项目状态在 `statusCache` 和 `pushoverStore` 中重复存储，可能导致不同步

## 核心设计原则

1. **StatusCache 作为唯一状态源**：所有项目状态（Git + Pushover）统一存储在 `statusCache` 中
2. **统一启动流程**：在 `App.vue` 中集中管理所有初始化任务，并行执行
3. **缓存优先策略**：UI 组件优先从缓存读取，仅在缓存过期时才触发后台刷新
4. **优雅降级**：初始化失败时记录错误并显示警告，但不阻塞应用启动

## 架构设计

### 数据流

**启动时**：
```
启动画面显示
    ↓
并行执行 Promise.all([
  - projectStore.loadProjects()
  - statusCache.init()           // 预加载所有项目状态
  - pushoverStore.checkExtensionStatus()  // 检查 Pushover 扩展状态
  - pushoverStore.checkPushoverConfig()   // 检查 Pushover 配置
])
    ↓
等待所有任务完成（捕获错误，不阻塞）
    ↓
隐藏启动画面
    ↓
显示主界面（数据已就绪）
```

**切换项目时**：
```
用户点击项目
    ↓
从 statusCache 读取缓存
    ↓
如果缓存未过期 → 立即显示
如果缓存已过期 → 显示加载中 → 刷新完成 → 显示
```

## 模块设计

### 1. StatusCache 改造

**目标**：成为唯一的项目状态数据源，统一管理 Git 状态和 Pushover 状态。

**类型定义**：
```typescript
interface ProjectStatusCache {
  gitStatus: GitStatus | null          // Git 状态
  stagingStatus: StagingStatus | null  // 暂存区状态
  untrackedCount: number               // 未跟踪文件数
  pushoverStatus: HookStatus | null    // Pushover 状态
  lastUpdated: number                  // 最后更新时间
  loading: boolean                     // 是否正在加载
  error: string | null                 // 错误信息
  stale: boolean                       // 是否过期
}
```

**新增方法**：
```typescript
function getPushoverStatus(path: string): HookStatus | null {
  return cache.value[path]?.pushoverStatus || null
}
```

### 2. PushoverStore 改造

**目标**：从 StatusCache 读取状态，移除重复存储。

**移除状态**：
```typescript
- const projectHookStatus = ref<Map<string, HookStatus>>(new Map())
- const statusVersion = ref(0)
```

**新增 Computed**：
```typescript
// 根据选中的项目获取 Pushover 状态
const currentProjectHookStatus = computed(() => {
  const selectedPath = projectStore.selectedProject
  if (!selectedPath) return null
  return statusCache.getPushoverStatus(selectedPath)
})

// 批量获取所有项目的 Pushover 状态（供 ProjectList 使用）
const allProjectHookStatuses = computed(() => {
  const statuses: Record<string, HookStatus> = {}
  for (const [path, cache] of Object.entries(statusCache.cache)) {
    if (cache.pushoverStatus) {
      statuses[path] = cache.pushoverStatus
    }
  }
  return statuses
})
```

**保持独立**：
- `extensionInfo`：扩展信息（非项目特定）
- `configValid`：配置有效性

### 3. App.vue 启动流程改造

**新增统一初始化函数**：
```typescript
async function initializeApp() {
  const tasks = [
    projectStore.loadProjects()
      .catch(err => ({ error: 'loadProjects', message: err.message })),
    statusCache.init()
      .catch(err => ({ error: 'statusCache', message: err.message })),
    pushoverStore.checkExtensionStatus()
      .catch(err => ({ error: 'extensionStatus', message: err.message })),
    pushoverStore.checkPushoverConfig()
      .catch(err => ({ error: 'pushoverConfig', message: err.message }))
  ]

  const results = await Promise.all(tasks)
  const errors = results.filter(r => r && r.error)

  if (errors.length > 0) {
    console.warn('[App] 部分初始化任务失败:', errors)
  }

  return errors
}
```

**修改 onMounted**：
```typescript
onMounted(async () => {
  await initializeApp()
  initialLoading.value = false

  setTimeout(() => {
    showSplash.value = false
  }, 500)
})
```

### 4. 组件使用缓存改造

**ProjectList.vue**：
- 已正确实现，无需修改
- 从 `statusCache.getStatus()` 读取项目状态

**CommitPanel.vue**：
```typescript
// 从 statusCache 读取 Pushover 状态
const pushoverStatus = computed(() => {
  if (!currentProjectPath.value) return null
  const cached = statusCache.getStatus(currentProjectPath.value)
  return cached?.pushoverStatus || null
})

// 切换项目时的逻辑
watch(() => projectStore.selectedProject, async (project) => {
  if (project) {
    commitStore.clearMessage()
    await commitStore.loadProjectAIConfig(project.id)

    const cached = statusCache.getStatus(project.path)

    if (cached && !statusCache.isExpired(project.path)) {
      // 缓存未过期，立即显示
      updateUIFromCache(cached)
    } else {
      // 缓存过期，等待刷新
      await statusCache.refresh(project.path, { force: true })
      const fresh = statusCache.getStatus(project.path)
      if (fresh) {
        updateUIFromCache(fresh)
      }
    }
  }
})
```

**ExtensionStatusButton.vue**：
- 移除 `onMounted` 中的 `checkExtensionStatus()` 调用
- 扩展状态已在 App.vue 启动时检查

## 错误处理

### 启动时错误
- 收集所有任务错误，不阻塞启动
- 可选：在 UI 上显示警告横幅

### 缓存刷新失败
- 标记错误状态，允许使用过期缓存
- 组件显示错误提示和重试按钮

### 边界情况
- **首次启动（无项目）**：正常显示空状态 UI
- **项目路径变更**：自动同步缓存
- **网络断开**：显示缓存数据并标记为 stale
- **外部修改**：通过 Wails Events 自动失效缓存

## 预期效果

- ✅ 启动画面隐藏时，所有数据已加载完成
- ✅ Pushover 扩展状态在启动画面期间检查完成（< 1 秒）
- ✅ 切换项目时无延迟感（缓存优先）
- ✅ 项目列表和详情页状态完全同步
- ✅ 优雅的错误处理，不阻塞应用启动

## 实现计划

### Phase 1: StatusCache 和 PushoverStore 改造
1. 在 StatusCache 中添加 `getPushoverStatus()` 方法
2. 重构 PushoverStore，移除 `projectHookStatus` 和 `statusVersion`
3. 添加 `currentProjectHookStatus` 和 `allProjectHookStatuses` computed
4. 更新相关方法的实现

### Phase 2: App.vue 启动流程
1. 创建 `initializeApp()` 函数
2. 修改 `onMounted` 逻辑
3. 移除 ExtensionStatusButton 的 `onMounted`

### Phase 3: 组件适配
1. 修改 CommitPanel.vue 的 `pushoverStatus` computed
2. 修改切换项目时的刷新逻辑
3. 移除 ExtensionStatusButton 的重复加载

### Phase 4: 测试和验证
1. 启动流程测试（首次启动、正常启动、部分失败）
2. 切换项目测试（缓存未过期、缓存过期）
3. 状态同步测试（项目列表 vs 详情页）
4. 错误处理测试（网络断开、后端失败）

## 文件变更清单

- `frontend/src/stores/statusCache.ts` - 添加 `getPushoverStatus()` 方法
- `frontend/src/stores/pushoverStore.ts` - 重构状态管理
- `frontend/src/App.vue` - 统一启动流程
- `frontend/src/components/CommitPanel.vue` - 适配新的状态源
- `frontend/src/components/ExtensionStatusButton.vue` - 移除重复加载
- `frontend/src/components/ProjectList.vue` - 无需修改

## 后续优化建议

1. **性能监控**：添加启动时间统计，确保在合理范围内
2. **增量刷新**：考虑只刷新变化的项目状态，而非全部刷新
3. **后台同步**：定期静默刷新所有项目状态，保持数据新鲜度
4. **离线支持**：更好的网络断开处理和恢复策略
