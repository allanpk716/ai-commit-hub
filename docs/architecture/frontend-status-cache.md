# Frontend StatusCache Architecture

## 概述

StatusCache 是项目状态的缓存层，提供高性能的状态管理，采用模块化设计将职责分离到独立的模块中。

## 模块设计

### Core（核心模块）

**文件：** `frontend/src/stores/statusCache/core.ts`

**职责：**
- 缓存存储和检索
- 批量操作
- 缓存生命周期管理

**核心 API：**
```typescript
class StatusCacheCore {
  getStatus(path): ProjectStatusCache | null
  getStatuses(paths): Record<string, ProjectStatusCache>
  updateCache(path, status): void
  updateCacheBatch(statuses): void
  clearCache(path): void
  clearAll(): void
}
```

**使用示例：**
```typescript
const core = new StatusCacheCore()

// 获取单个项目状态
const status = core.getStatus('/path/to/project')

// 批量获取
const statuses = core.getStatuses(['/path1', '/path2'])

// 更新缓存
core.updateCache('/path/to/project', {
  gitStatus: newStatus,
  lastUpdated: Date.now()
})
```

### Validation（验证模块）

**文件：** `frontend/src/stores/statusCache/validation.ts`

**职责：**
- 数据完整性验证
- 自动修复损坏数据
- 确保默认值

**核心 API：**
```typescript
class StatusCacheValidation {
  validateProjectStatus(status): boolean
  validateStagingStatus(status): boolean
  validateHookStatus(status): boolean
  validatePushStatus(status): boolean
  validateCache(cache): { valid: boolean; errors: string[] }
  repairCache(cache): ProjectStatusCache
  ensureDefaults(cache): ProjectStatusCache
}
```

**验证规则：**
- ProjectStatus: 必须包含 `branch` (string) 和 `hasUncommittedChanges` (boolean)
- StagingStatus: 必须包含 `stagedFiles` 和 `unstagedFiles` (数组)
- HookStatus: 必须包含 `installed` 和 `isLatestVersion` (boolean)
- PushStatus: 必须包含 `canPush` 和 `pushed` (boolean)

**使用示例：**
```typescript
const validation = new StatusCacheValidation()

// 验证缓存
const result = validation.validateCache(cache)
if (!result.valid) {
  console.error('Invalid cache:', result.errors)
}

// 自动修复
const repaired = validation.repairCache(cache)

// 确保默认值
const withDefaults = validation.ensureDefaults(cache)
```

### Retry（重试模块）

**文件：** `frontend/src/stores/statusCache/retry.ts`

**职责：**
- 失败重试策略
- 指数退避算法
- 批量操作管理

**核心 API：**
```typescript
interface RetryOptions {
  maxAttempts?: number      // 最大重试次数
  initialDelay?: number     // 初始延迟
  backoffMultiplier?: number // 退避倍数
}

async function withRetry<T>(
  operation: () => Promise<T>,
  options?: RetryOptions
): Promise<RetryResult<T>>

class RetryManager {
  static execute<T>(operation, options?): Promise<RetryResult<T>>
  static executeBatch<T>(operations, options?): Promise<RetryResult<T>[]>
}
```

**重试策略：**
- 默认最大尝试次数：3
- 默认初始延迟：1000ms
- 默认退避倍数：2
- 指数退避：delay = initialDelay * (backoffMultiplier ^ (attempt - 1))

**使用示例：**
```typescript
import { withRetry, RetryManager } from '@/stores/statusCache/retry'

// 单次操作带重试
const result = await withRetry(
  () => fetchData(),
  { maxAttempts: 5, initialDelay: 2000 }
)

if (result.success) {
  console.log('Data:', result.data)
} else {
  console.error('Failed after', result.attempts, 'attempts')
}

// 使用 RetryManager
const result = await RetryManager.execute(
  () => fetchData(),
  { maxAttempts: 3 }
)

// 批量操作
const results = await RetryManager.executeBatch([
  () => fetchProject('/path1'),
  () => fetchProject('/path2')
])
```

## Git 操作封装

**文件：** `frontend/src/composables/useGitOperation.ts`

**职责：**
- 统一 Git 操作处理
- 乐观更新管理
- 错误回滚

**核心 API：**
```typescript
interface GitOperationOptions {
  optimisticUpdate?: Partial<ProjectStatusCache>
  refreshOnSuccess?: boolean
  silent?: boolean
}

interface GitOperationResult<T> {
  success: boolean
  data?: T
  error?: Error
  rollback?: () => void
}

function useGitOperation() {
  return {
    executeGitOperation<T>(
      projectPath: string,
      operation: () => Promise<T>,
      options?: GitOperationOptions
    ): Promise<GitOperationResult<T>>

    executeBatch<T>(
      operations: Array<{
        fn: () => Promise<T>
        projectPath: string
        options?: GitOperationOptions
      }>
    ): Promise<GitOperationResult<T>[]>

    // ... 其他辅助方法
  }
}
```

**使用示例：**
```typescript
const { executeGitOperation } = useGitOperation()

// 单个操作
const result = await executeGitOperation(
  '/project/path',
  () => StageFile('/project/path', 'file.txt'),
  {
    optimisticUpdate: {
      stagingStatus: { /* 更新后的暂存区状态 */ }
    },
    refreshOnSuccess: true
  }
)

if (result.success) {
  console.log('操作成功')
} else {
  console.error('操作失败:', result.error)
}
```

## 最佳实践

### 1. 缓存优先策略

```typescript
// 优先使用缓存数据
const cached = statusCache.getStatus(projectPath)
if (cached && !isExpired(cached)) {
  return cached  // 立即返回
}

// 后台刷新过期缓存
if (isExpired(cached)) {
  statusCache.refresh(projectPath, { silent: true })
}
```

### 2. 乐观更新模式

```typescript
// 1. 立即更新 UI
const rollback = statusCache.updateOptimistic(projectPath, {
  hasUncommittedChanges: false
})

try {
  // 2. 执行 Git 操作
  await CommitProject(projectPath, message)

  // 3. 强制刷新确保一致性
  await statusCache.refresh(projectPath, { force: true })
} catch (error) {
  // 4. 失败时回滚
  rollback?.()
  throw error
}
```

### 3. 错误恢复

```typescript
try {
  await statusCache.refresh(projectPath)
} catch (error) {
  // 使用过期缓存作为降级
  const stale = statusCache.getStatus(projectPath)
  if (stale && isStaleAcceptable(stale)) {
    return stale
  }
  throw error
}
```

### 4. 批量操作优化

```typescript
// 使用批量接口提高性能
const statuses = await statusCache.refreshBatch(projectPaths)

// 并行执行多个操作
const results = await executeBatch(
  projectPaths.map(path => ({
    fn: () => refreshProject(path),
    projectPath: path
  }))
)
```

## 性能优化建议

1. **预加载**: 应用启动时批量加载所有项目状态
2. **后台刷新**: 缓存过期后在后台静默刷新
3. **智能 TTL**: 根据用户操作频率动态调整缓存过期时间
4. **批量操作**: 使用批量接口减少网络请求
5. **防抖**: 对频繁的状态更新进行防抖处理

## 调试技巧

```typescript
// 获取所有缓存（调试用）
const allCache = statusCache.getAllCache()
console.log('All cache:', allCache)

// 检查缓存是否存在
if (statusCache.has('/project/path')) {
  console.log('Cache exists')
}

// 检查缓存数量
console.log('Cache count:', statusCache.size)
```

## 相关文件

- `frontend/src/stores/statusCache.ts`: StatusCache 主入口
- `frontend/src/stores/statusCache/core.ts`: 核心缓存功能
- `frontend/src/stores/statusCache/validation.ts`: 数据验证
- `frontend/src/stores/statusCache/retry.ts`: 重试逻辑
- `frontend/src/composables/useGitOperation.ts`: Git 操作封装
