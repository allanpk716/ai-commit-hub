# Code Optimization Phase 2: Advanced Refactoring Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 继续优化代码质量，拆分 statusCache、创建统一的 Git 操作包装器、定义事件常量、添加接口抽象、改善错误类型系统。

**Architecture:** 将 statusCache.ts 按职责拆分为多个模块（缓存、验证、重试），创建通用的 Git 操作包装器减少重复代码，定义统一的事件名称常量，为 Service 层添加接口抽象提升可测试性，建立领域特定的错误类型系统。

**Tech Stack:** Go 1.21+、Wails v2、Vue 3、TypeScript、Vite、Pinia

---

## Phase 1: Frontend Status Cache Refactoring

### Task 1: 创建事件名称常量文件

**Files:**
- Create: `frontend/src/constants/events.ts`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 创建事件常量定义**

```typescript
// frontend/src/constants/events.ts
/**
 * Wails 事件名称常量
 * 所有事件名称统一管理，避免硬编码和拼写错误
 */
export const APP_EVENTS = {
  // 应用生命周期
  STARTUP_COMPLETE: 'startup:complete',
  WINDOW_SHOWN: 'window:shown',
  WINDOW_HIDDEN: 'window:hidden',

  // Commit 生成
  COMMIT_DELTA: 'commit:delta',
  COMMIT_COMPLETE: 'commit:complete',
  COMMIT_ERROR: 'commit:error',

  // 项目状态
  PROJECT_STATUS_CHANGED: 'project:status-changed',
  PROJECT_HOOK_UPDATED: 'project:hook-updated',

  // Pushover
  PUSHOVER_STATUS_CHANGED: 'pushover:status-changed',
} as const

export type AppEvent = typeof APP_EVENTS[keyof typeof APP_EVENTS]
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 编译检查**

Run: `cd frontend && npm run build`
Expected: Build succeeds

**Step 4: 提交**

```bash
git add frontend/src/constants/events.ts
git commit -m "feat(frontend): 添加事件名称常量定义"
```

---

### Task 2: 在 main.ts 中使用事件常量

**Files:**
- Modify: `frontend/src/main.ts:13-28`, `frontend/src/main.ts:37-70`

**Step 1: 添加常量导入**

在文件顶部添加：
```typescript
import { APP_EVENTS } from './constants/events'
```

**Step 2: 替换 startup-complete 事件 (main.ts:37)**

查找：
```typescript
EventsOn('startup-complete', async (data: { success?: boolean; statuses?: Record<string, any> } | null) => {
```

替换为：
```typescript
EventsOn(APP_EVENTS.STARTUP_COMPLETE, async (data: { success?: boolean; statuses?: Record<string, any> } | null) => {
```

**Step 3: 替换 project-status-changed 事件 (main.ts:61)**

查找：
```typescript
EventsOn('project-status-changed', async (data: { projectPath: string }) => {
```

替换为：
```typescript
EventsOn(APP_EVENTS.PROJECT_STATUS_CHANGED, async (data: { projectPath: string }) => {
```

**Step 4: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 5: 手动测试**

Run: `wails dev`
Expected: 应用正常启动，事件监听正常

**Step 6: 提交**

```bash
git add frontend/src/main.ts
git commit -m "refactor(frontend): main.ts 使用事件常量"
```

---

### Task 3: 在 App.vue 中使用事件常量

**Files:**
- Modify: `frontend/src/App.vue`

**Step 1: 添加常量导入**

在 script setup 部分添加：
```typescript
import { APP_EVENTS } from '@/constants/events'
```

**Step 2: 替换所有事件字符串**

使用编辑器查找替换：
- `'startup-complete'` → `APP_EVENTS.STARTUP_COMPLETE`
- `'window-shown'` → `APP_EVENTS.WINDOW_SHOWN`
- `'window-hidden'` → `APP_EVENTS.WINDOW_HIDDEN`
- `'project-status-changed'` → `APP_EVENTS.PROJECT_STATUS_CHANGED`

**Step 3: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 4: 提交**

```bash
git add frontend/src/App.vue
git commit -m "refactor(frontend): App.vue 使用事件常量"
```

---

### Task 4: 在 commitStore.ts 中使用事件常量

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 添加常量导入**

在文件顶部添加：
```typescript
import { APP_EVENTS } from '@/constants/events'
```

**Step 2: 替换 commit-delta 事件**

查找所有：
```typescript
EventsOn('commit-delta', ...)
```

替换为：
```typescript
EventsOn(APP_EVENTS.COMMIT_DELTA, ...)
```

**Step 3: 替换 commit-complete 事件**

查找所有：
```typescript
EventsOn('commit-complete', ...)
```

替换为：
```typescript
EventsOn(APP_EVENTS.COMMIT_COMPLETE, ...)
```

**Step 4: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 5: 提交**

```bash
git add frontend/src/stores/commitStore.ts
git commit -m "refactor(frontend): commitStore 使用事件常量"
```

---

### Task 5: 拆分 statusCache.ts - 创建核心缓存模块

**Files:**
- Create: `frontend/src/stores/statusCache/core.ts`
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 创建核心缓存模块**

```typescript
// frontend/src/stores/statusCache/core.ts
import { ref, Ref } from 'vue'
import type { ProjectStatusCache } from '@/types/status'

/**
 * StatusCacheCore - 核心缓存功能
 * 负责基本的缓存存储和检索
 */
export class StatusCacheCore {
  private cache: Ref<Record<string, ProjectStatusCache>>

  constructor() {
    this.cache = ref<Record<string, ProjectStatusCache>>({})
  }

  /**
   * 获取项目状态缓存
   */
  getStatus(path: string): ProjectStatusCache | null {
    return this.cache.value[path] || null
  }

  /**
   * 获取多个项目的状态缓存
   */
  getStatuses(paths: string[]): Record<string, ProjectStatusCache> {
    const result: Record<string, ProjectStatusCache> = {}

    for (const path of paths) {
      const status = this.getStatus(path)
      if (status) {
        result[path] = status
      }
    }

    return result
  }

  /**
   * 更新项目状态缓存
   */
  updateCache(path: string, status: Partial<ProjectStatusCache>): void {
    const existing = this.getStatus(path) || this.createEmptyCache()
    this.cache.value[path] = {
      ...existing,
      ...status,
      lastUpdated: Date.now(),
    }
  }

  /**
   * 批量更新缓存
   */
  updateCacheBatch(statuses: Record<string, Partial<ProjectStatusCache>>): void {
    for (const [path, status] of Object.entries(statuses)) {
      this.updateCache(path, status)
    }
  }

  /**
   * 清除项目缓存
   */
  clearCache(path: string): void {
    delete this.cache.value[path]
  }

  /**
   * 清除所有缓存
   */
  clearAll(): void {
    this.cache.value = {}
  }

  /**
   * 创建空缓存对象
   */
  private createEmptyCache(): ProjectStatusCache {
    return {
      gitStatus: null,
      stagingStatus: null,
      pushoverStatus: null,
      pushStatus: null,
      untrackedCount: 0,
      lastUpdated: 0,
      loading: false,
      error: null,
      stale: false,
    }
  }

  /**
   * 获取所有缓存（用于调试）
   */
  getAllCache(): Record<string, ProjectStatusCache> {
    return { ...this.cache.value }
  }
}
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors（可能有未使用的警告，这是预期的）

**Step 3: 暂存文件**

```bash
git add frontend/src/stores/statusCache/core.ts
git commit -m "feat(statusCache): 创建核心缓存模块"
```

---

### Task 6: 拆分 statusCache.ts - 创建验证模块

**Files:**
- Create: `frontend/src/stores/statusCache/validation.ts`
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 创建验证模块**

```typescript
// frontend/src/stores/statusCache/validation.ts
import type { ProjectStatusCache, ProjectStatus, StagingStatus, HookStatus, PushStatus } from '@/types/status'

/**
 * StatusCacheValidation - 数据验证和修复
 * 负责验证缓存数据的完整性并修复问题
 */
export class StatusCacheValidation {
  /**
   * 验证项目状态完整性
   */
  validateProjectStatus(status: ProjectStatus | null): boolean {
    if (!status) return false

    // 验证必需字段
    return typeof status.branch === 'string' &&
           typeof status.hasUncommittedChanges === 'boolean'
  }

  /**
   * 验证暂存区状态完整性
   */
  validateStagingStatus(status: StagingStatus | null): boolean {
    if (!status) return true // 空 staging 状态是有效的

    return Array.isArray(status.stagedFiles) &&
           Array.isArray(status.unstagedFiles)
  }

  /**
   * 验证 Hook 状态完整性
   */
  validateHookStatus(status: HookStatus | null): boolean {
    if (!status) return true

    return typeof status.installed === 'boolean' &&
           typeof status.isLatestVersion === 'boolean'
  }

  /**
   * 验证推送状态完整性
   */
  validatePushStatus(status: PushStatus | null): boolean {
    if (!status) return true

    return typeof status.canPush === 'boolean' &&
           typeof status.pushed === 'boolean'
  }

  /**
   * 验证完整缓存对象
   */
  validateCache(cache: ProjectStatusCache): { valid: boolean; errors: string[] } {
    const errors: string[] = []

    if (!this.validateProjectStatus(cache.gitStatus)) {
      errors.push('Invalid git status')
    }

    if (!this.validateStagingStatus(cache.stagingStatus)) {
      errors.push('Invalid staging status')
    }

    if (!this.validateHookStatus(cache.pushoverStatus)) {
      errors.push('Invalid pushover status')
    }

    if (!this.validatePushStatus(cache.pushStatus)) {
      errors.push('Invalid push status')
    }

    return {
      valid: errors.length === 0,
      errors,
    }
  }

  /**
   * 修复缓存数据
   */
  repairCache(cache: ProjectStatusCache): ProjectStatusCache {
    const repaired = { ...cache }

    // 修复 gitStatus
    if (!this.validateProjectStatus(repaired.gitStatus)) {
      repaired.gitStatus = null
    }

    // 修复 stagingStatus
    if (!this.validateStagingStatus(repaired.stagingStatus)) {
      repaired.stagingStatus = null
    }

    // 修复 pushoverStatus
    if (!this.validateHookStatus(repaired.pushoverStatus)) {
      repaired.pushoverStatus = null
    }

    // 修复 pushStatus
    if (!this.validatePushStatus(repaired.pushStatus)) {
      repaired.pushStatus = null
    }

    return repaired
  }

  /**
   * 修复缺失的默认值
   */
  ensureDefaults(cache: ProjectStatusCache): ProjectStatusCache {
    return {
      gitStatus: cache.gitStatus || null,
      stagingStatus: cache.stagingStatus || null,
      pushoverStatus: cache.pushoverStatus || null,
      pushStatus: cache.pushStatus || null,
      untrackedCount: cache.untrackedCount ?? 0,
      lastUpdated: cache.lastUpdated || 0,
      loading: cache.loading || false,
      error: cache.error || null,
      stale: cache.stale || false,
    }
  }
}
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 提交**

```bash
git add frontend/src/stores/statusCache/validation.ts
git commit -m "feat(statusCache): 创建数据验证模块"
```

---

### Task 7: 拆分 statusCache.ts - 创建重试模块

**Files:**
- Create: `frontend/src/stores/statusCache/retry.ts`
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 创建重试模块**

```typescript
// frontend/src/stores/statusCache/retry.ts
/**
 * StatusCacheRetry - 重试逻辑
 * 负责处理失败操作的重试策略
 */

export interface RetryOptions {
  maxAttempts?: number
  initialDelay?: number // ms
  backoffMultiplier?: number
}

export interface RetryResult<T> {
  success: boolean
  data?: T
  error?: Error
  attempts: number
}

/**
 * 带指数退避的重试函数
 */
export async function withRetry<T>(
  operation: () => Promise<T>,
  options: RetryOptions = {}
): Promise<RetryResult<T>> {
  const {
    maxAttempts = 3,
    initialDelay = 1000,
    backoffMultiplier = 2,
  } = options

  let lastError: Error | undefined
  let delay = initialDelay

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const data = await operation()

      return {
        success: true,
        data,
        attempts: attempt,
      }
    } catch (error) {
      lastError = error as Error

      // 如果是最后一次尝试，不再等待
      if (attempt < maxAttempts) {
        await sleep(delay)
        delay *= backoffMultiplier
      }
    }
  }

  return {
    success: false,
    error: lastError,
    attempts: maxAttempts,
  }
}

/**
 * 延迟函数
 */
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * RetryManager - 重试管理器
 */
export class RetryManager {
  private static defaultOptions: RetryOptions = {
    maxAttempts: 3,
    initialDelay: 1000,
    backoffMultiplier: 2,
  }

  /**
   * 设置默认重试选项
   */
  static setDefaultOptions(options: RetryOptions): void {
    this.defaultOptions = { ...this.defaultOptions, ...options }
  }

  /**
   * 使用默认选项执行重试
   */
  static async execute<T>(
    operation: () => Promise<T>,
    options?: RetryOptions
  ): Promise<RetryResult<T>> {
    return withRetry(operation, {
      ...this.defaultOptions,
      ...options,
    })
  }

  /**
   * 批量重试多个操作
   */
  static async executeBatch<T>(
    operations: Array<() => Promise<T>>,
    options?: RetryOptions
  ): Promise<RetryResult<T>[]> {
    const results = await Promise.all(
      operations.map(op => this.execute(op, options))
    )

    return results
  }
}
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 提交**

```bash
git add frontend/src/stores/statusCache/retry.ts
git commit -m "feat(statusCache): 创建重试逻辑模块"
```

---

### Task 8: 创建 Git 操作包装器

**Files:**
- Create: `frontend/src/composables/useGitOperation.ts`
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 创建通用 Git 操作包装器**

```typescript
// frontend/src/composables/useGitOperation.ts
import { useStatusCache } from '@/stores/statusCache'
import type { ProjectStatusCache } from '@/types/status'

/**
 * Git 操作结果
 */
export interface GitOperationResult<T = void> {
  success: boolean
  data?: T
  error?: Error
}

/**
 * Git 操作选项
 */
export interface GitOperationOptions {
  optimisticUpdate?: Partial<ProjectStatusCache>
  refreshAfter?: boolean
  notifyAfter?: boolean
}

/**
 * 通用 Git 操作包装器
 * 处理乐观更新、错误回滚、状态刷新
 */
export function useGitOperation() {
  const statusCache = useStatusCache()

  /**
   * 执行 Git 操作
   */
  async function executeGitOperation<T>(
    projectPath: string,
    operation: () => Promise<T>,
    options: GitOperationOptions = {}
  ): Promise<GitOperationResult<T>> {
    const {
      optimisticUpdate,
      refreshAfter = true,
      notifyAfter = true,
    } = options

    // 1. 乐观更新
    let rollback: (() => void) | undefined
    if (optimisticUpdate) {
      rollback = statusCache.updateOptimistic(projectPath, optimisticUpdate)
    }

    try {
      // 2. 执行操作
      const data = await operation()

      // 3. 刷新状态
      if (refreshAfter) {
        await statusCache.refresh(projectPath, { force: true })
      }

      // 4. 通知状态变更
      if (notifyAfter) {
        notifyProjectStatusChanged(projectPath)
      }

      return {
        success: true,
        data,
      }
    } catch (error) {
      // 5. 错误回滚
      rollback?.()

      return {
        success: false,
        error: error as Error,
      }
    }
  }

  /**
   * 批量执行 Git 操作
   */
  async function executeBatch<T>(
    operations: Array<{
      path: string
      operation: () => Promise<T>
      options?: GitOperationOptions
    }>
  ): Promise<GitOperationResult<T>[]> {
    const results = await Promise.all(
      operations.map(({ path, operation, options }) =>
        executeGitOperation(path, operation, options)
      )
    )

    return results
  }

  return {
    executeGitOperation,
    executeBatch,
  }
}

/**
 * 通知项目状态变更
 */
function notifyProjectStatusChanged(projectPath: string) {
  // 这里可以触发全局事件或调用 store 方法
  // 具体实现根据项目的状态管理方式而定
}
```

**Step 2: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 3: 提交**

```bash
git add frontend/src/composables/useGitOperation.ts
git commit -m "feat(composable): 创建通用 Git 操作包装器"
```

---

### Task 9: 使用 Git 操作包装器重构 commitStore

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 添加 Git 操作包装器导入**

在文件顶部添加：
```typescript
import { useGitOperation } from '@/composables/useGitOperation'
```

**Step 2: 在 setup 中初始化**

在 store 的 setup 函数中添加：
```typescript
const { executeGitOperation } = useGitOperation()
```

**Step 3: 重构 stageFile 函数**

查找 stageFile 函数，替换为：

```typescript
async function stageFile(filePath: string) {
  if (!selectedProjectPath.value) {
    throw new Error('No project selected')
  }

  const result = await executeGitOperation(
    selectedProjectPath.value,
    () => StageFile(selectedProjectPath.value!, filePath),
    {
      optimisticUpdate: {
        // 根据实际情况设置乐观更新
      },
    }
  )

  if (!result.success) {
    throw result.error
  }

  await refreshAll()
}
```

**Step 4: 重构 unstageFile 函数**

类似地重构 unstageFile

**Step 5: 重构 commitProject 函数**

类似地重构 commitProject

**Step 6: 重构 discardChanges 函数**

类似地重构 discardChanges

**Step 7: 运行测试**

Run: `cd frontend && npm run test:run`
Expected: All tests pass

**Step 8: 运行类型检查**

Run: `cd frontend && npm run type-check`
Expected: No errors

**Step 9: 提交**

```bash
git add frontend/src/stores/commitStore.ts
git commit -m "refactor(commitStore): 使用 Git 操作包装器简化代码"
```

---

## Phase 2: Backend Refactoring

### Task 10: 创建领域错误类型系统

**Files:**
- Create: `pkg/errors/domain_errors.go`
- Modify: `pkg/service/*.go` (多个 service 文件)

**Step 1: 创建领域错误类型**

```go
// pkg/errors/domain_errors.go
package errors

import "fmt"

// ValidationError 表示数据验证错误
type ValidationError struct {
    Field   string
    Message string
    Err     error
}

func (e *ValidationError) Error() string {
    if e.Field != "" {
        return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
    }
    return fmt.Sprintf("validation failed: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
    return e.Err
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string, err error) *ValidationError {
    return &ValidationError{
        Field:   field,
        Message: message,
        Err:     err,
    }
}

// GitOperationError 表示 Git 操作错误
type GitOperationError struct {
    Operation string
    Path      string
    Err       error
}

func (e *GitOperationError) Error() string {
    return fmt.Sprintf("git operation '%s' failed at %s: %v", e.Operation, e.Path, e.Err)
}

func (e *GitOperationError) Unwrap() error {
    return e.Err
}

// NewGitOperationError 创建 Git 操作错误
func NewGitOperationError(operation, path string, err error) *GitOperationError {
    return &GitOperationError{
        Operation: operation,
        Path:      path,
        Err:       err,
    }
}

// AIProviderError 表示 AI Provider 错误
type AIProviderError struct {
    Provider string
    Message  string
    Err      error
}

func (e *AIProviderError) Error() string {
    return fmt.Sprintf("AI provider '%s' error: %s", e.Provider, e.Message)
}

func (e *AIProviderError) Unwrap() error {
    return e.Err
}

// NewAIProviderError 创建 AI Provider 错误
func NewAIProviderError(provider, message string, err error) *AIProviderError {
    return &AIProviderError{
        Provider: provider,
        Message:  message,
        Err:      err,
    }
}

// IsNotFoundError 检查是否是"未找到"错误
func IsNotFoundError(err error) bool {
    return err != nil && (err.Error() == "record not found" || err.Error() == "not found")
}

// IsValidationError 检查是否是验证错误
func IsValidationError(err error) bool {
    _, ok := err.(*ValidationError)
    return ok
}

// IsGitError 检查是否是 Git 操作错误
func IsGitError(err error) bool {
    _, ok := err.(*GitOperationError)
    return ok
}
```

**Step 2: 运行格式化**

Run: `go fmt ./pkg/errors/domain_errors.go`
Expected: No errors

**Step 3: 编译检查**

Run: `go build ./pkg/errors`
Expected: Success

**Step 4: 提交**

```bash
git add pkg/errors/domain_errors.go
git commit -m "feat(errors): 添加领域错误类型系统"
```

---

### Task 11: 在 Service 层使用领域错误

**Files:**
- Modify: `pkg/service/config_service.go`
- Modify: `pkg/service/commit_service.go`

**Step 1: 在 ConfigService 中使用错误类型**

查找验证逻辑，替换为：
```go
import apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"

func (s *ConfigService) ValidateConfig(config *models.Config) error {
    if config.Provider == "" {
        return apperrors.NewValidationError("provider", "cannot be empty", nil)
    }

    if config.ApiKey == "" && config.Provider != "ollama" {
        return apperrors.NewValidationError("api_key", "required for this provider", nil)
    }

    return nil
}
```

**Step 2: 在 CommitService 中使用 Git 错误**

查找 Git 操作错误处理，替换为：
```go
import apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"

func (s *CommitService) GenerateCommit(projectPath string) error {
    // ... 执行 git 操作

    if err != nil {
        return apperrors.NewGitOperationError("status", projectPath, err)
    }

    // ...
}
```

**Step 3: 运行测试**

Run: `go test ./pkg/service/... -v`
Expected: All tests pass

**Step 4: 提交**

```bash
git add pkg/service/config_service.go pkg/service/commit_service.go
git commit -m "refactor(service): 使用领域错误类型"
```

---

### Task 12: 创建 Repository 接口抽象

**Files:**
- Create: `pkg/repository/interfaces.go`
- Modify: `pkg/repository/git_project_repository.go`
- Modify: `app.go`

**Step 1: 创建 Repository 接口**

```go
// pkg/repository/interfaces.go
package repository

import (
    "github.com/allanpk716/ai-commit-hub/pkg/models"
)

// GitProjectRepository 定义 Git 项目仓库接口
type GitProjectRepository interface {
    // GetAll 获取所有项目
    GetAll() ([]models.GitProject, error)

    // GetByID 根据 ID 获取项目
    GetByID(id uint) (*models.GitProject, error)

    // GetByPath 根据路径获取项目
    GetByPath(path string) (*models.GitProject, error)

    // Create 创建新项目
    Create(project *models.GitProject) error

    // Update 更新项目
    Update(project *models.GitProject) error

    // Delete 删除项目
    Delete(id uint) error

    // UpdateLastCommitTime 更新最后提交时间
    UpdateLastCommitTime(id uint, commitTime int64) error

    // UpdateHookStatus 更新 Hook 状态
    UpdateHookStatus(id uint, needsUpdate bool) error
}

// CommitHistoryRepository 定义提交历史仓库接口
type CommitHistoryRepository interface {
    // Create 创建提交历史记录
    Create(history *models.CommitHistory) error

    // GetByProjectID 获取项目的提交历史
    GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error)

    // Delete 删除历史记录
    Delete(id uint) error

    // DeleteByProjectID 删除项目的所有历史
    DeleteByProjectID(projectID uint) error
}
```

**Step 2: 更新 GitProjectRepository 实现**

在 `git_project_repository.go` 中，确保实现接口：

```go
// 确保类型实现了接口
var _ GitProjectRepository = (*GitProjectRepository)(nil)

type GitProjectRepository struct {
    db *gorm.DB
}

// ... 方法实现保持不变
```

**Step 3: 运行测试**

Run: `go test ./pkg/repository/... -v`
Expected: All tests pass

**Step 4: 提交**

```bash
git add pkg/repository/interfaces.go pkg/repository/git_project_repository.go
git commit -m "feat(repository): 添加 Repository 接口抽象"
```

---

### Task 13: 更新 Service 使用接口

**Files:**
- Modify: `pkg/service/startup_service.go`
- Modify: `app.go`

**Step 1: 更新 StartupService 使用接口**

```go
// pkg/service/startup_service.go
type StartupService struct {
    gitProjectRepo  repository.GitProjectRepository  // 使用接口而非具体类型
    pushoverService *pushover.Service
}

func NewStartupService(
    repo repository.GitProjectRepository,
    pushover *pushover.Service,
) *StartupService {
    return &StartupService{
        gitProjectRepo:  repo,
        pushoverService: pushover,
    }
}
```

**Step 2: 更新 app.go 中的初始化代码**

```go
// app.go
func NewApp() *App {
    // ...
    startupService := service.NewStartupService(a.gitProjectRepo, a.pushoverService)
    // ...
}
```

**Step 3: 运行测试**

Run: `go test ./... -v`
Expected: All tests pass

**Step 4: 编译检查**

Run: `wails build`
Expected: Build succeeds

**Step 5: 提交**

```bash
git add pkg/service/startup_service.go app.go
git commit -m "refactor(service): StartupService 使用 Repository 接口"
```

---

## Phase 3: Testing & Documentation

### Task 14: 编写接口单元测试

**Files:**
- Create: `pkg/repository/interfaces_test.go`

**Step 1: 创建 Mock Repository（用于测试）**

```go
// pkg/repository/mock_test.go
package repository

import (
    "github.com/allanpk716/ai-commit-hub/pkg/models"
)

// MockGitProjectRepository 是用于测试的 Mock 实现
type MockGitProjectRepository struct {
    projects []models.GitProject
    err      error
}

func NewMockGitProjectRepository(projects []models.GitProject, err error) *MockGitProjectRepository {
    return &MockGitProjectRepository{
        projects: projects,
        err:      err,
    }
}

func (m *MockGitProjectRepository) GetAll() ([]models.GitProject, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.projects, nil
}

func (m *MockGitProjectRepository) GetByID(id uint) (*models.GitProject, error) {
    if m.err != nil {
        return nil, m.err
    }
    for _, p := range m.projects {
        if p.ID == id {
            return &p, nil
        }
    }
    return nil, nil
}

// ... 实现其他接口方法
```

**Step 2: 创建接口测试**

```go
// pkg/repository/interfaces_test.go
package repository

import (
    "testing"

    "github.com/allanpk716/ai-commit-hub/pkg/models"
)

func TestGitProjectRepository_Interface(t *testing.T) {
    // 创建 mock 数据
    projects := []models.GitProject{
        {ID: 1, Name: "Test1", Path: "/path1"},
        {ID: 2, Name: "Test2", Path: "/path2"},
    }

    repo := NewMockGitProjectRepository(projects, nil)

    // 测试 GetAll
    result, err := repo.GetAll()
    if err != nil {
        t.Fatalf("GetAll failed: %v", err)
    }

    if len(result) != 2 {
        t.Errorf("Expected 2 projects, got %d", len(result))
    }

    // 测试 GetByID
    project, err := repo.GetByID(1)
    if err != nil {
        t.Fatalf("GetByID failed: %v", err)
    }

    if project == nil || project.Name != "Test1" {
        t.Error("GetByID returned wrong project")
    }
}
```

**Step 3: 运行测试**

Run: `go test ./pkg/repository/... -v`
Expected: All tests pass

**Step 4: 提交**

```bash
git add pkg/repository/mock_test.go pkg/repository/interfaces_test.go
git commit -m "test(repository): 添加 Repository 接口单元测试"
```

---

### Task 15: 创建前端测试用例

**Files:**
- Create: `frontend/src/composables/__tests__/useGitOperation.spec.ts`
- Create: `frontend/src/stores/statusCache/__tests__/validation.spec.ts`

**Step 1: 测试 Git 操作包装器**

```typescript
// frontend/src/composables/__tests__/useGitOperation.spec.ts
import { describe, it, expect, vi } from 'vitest'
import { useGitOperation } from '../useGitOperation'

describe('useGitOperation', () => {
  it('should execute operation successfully', async () => {
    const { executeGitOperation } = useGitOperation()

    const mockOperation = vi.fn().mockResolvedValue('success')

    const result = await executeGitOperation(
      '/test/path',
      mockOperation,
      { refreshAfter: false, notifyAfter: false }
    )

    expect(result.success).toBe(true)
    expect(result.data).toBe('success')
    expect(mockOperation).toHaveBeenCalledTimes(1)
  })

  it('should handle operation failure', async () => {
    const { executeGitOperation } = useGitOperation()

    const mockOperation = vi.fn().mockRejectedValue(new Error('test error'))

    const result = await executeGitOperation(
      '/test/path',
      mockOperation,
      { refreshAfter: false, notifyAfter: false }
    )

    expect(result.success).toBe(false)
    expect(result.error).toBeDefined()
  })
})
```

**Step 2: 测试验证模块**

```typescript
// frontend/src/stores/statusCache/__tests__/validation.spec.ts
import { describe, it, expect } from 'vitest'
import { StatusCacheValidation } from '../validation'

describe('StatusCacheValidation', () => {
  const validator = new StatusCacheValidation()

  it('should validate correct project status', () => {
    const status = {
      branch: 'main',
      hasUncommittedChanges: true,
      lastCommitHash: 'abc123',
    }

    expect(validator.validateProjectStatus(status)).toBe(true)
  })

  it('should reject invalid project status', () => {
    const status = {
      branch: 'main',
      // missing hasUncommittedChanges
    }

    expect(validator.validateProjectStatus(status as any)).toBe(false)
  })

  it('should repair invalid cache', () => {
    const invalidCache = {
      gitStatus: null,
      stagingStatus: null,
      pushoverStatus: null,
      pushStatus: null,
      untrackedCount: 'invalid' as any,
    }

    const repaired = validator.ensureDefaults(invalidCache)

    expect(repaired.untrackedCount).toBe(0)
  })
})
```

**Step 3: 运行前端测试**

Run: `cd frontend && npm run test:run`
Expected: All tests pass

**Step 4: 提交**

```bash
git add frontend/src/composables/__tests__/useGitOperation.spec.ts
git add frontend/src/stores/statusCache/__tests__/validation.spec.ts
git commit -m "test(frontend): 添加 Git 操作和验证模块测试"
```

---

### Task 16: 端到端测试

**Files:**
- Test: 手动测试清单
- Create: `tmp/test-report-phase2.md`

**Step 1: 启动应用测试**

Run: `wails dev`
Expected: 应用正常启动

**Step 2: 测试事件常量**

验证所有事件正常触发：
- 启动完成事件
- Commit 生成事件
- 状态变更事件

**Step 3: 测试 Git 操作包装器**

验证以下操作：
- 文件暂存
- 文件取消暂存
- 提交
- 丢弃更改

**Step 4: 测试错误处理**

验证错误提示：
- 验证错误显示
- Git 操作错误显示
- AI Provider 错误显示

**Step 5: 创建测试报告**

```markdown
# Phase 2 测试报告

## 测试日期
2026-02-04

## 事件系统
- [x] 所有事件使用常量定义
- [x] 事件监听正常工作
- [x] 事件参数正确传递

## StatusCache 重构
- [x] 核心缓存功能正常
- [x] 数据验证工作正常
- [x] 重试机制有效

## Git 操作包装器
- [x] 乐观更新正常
- [x] 错误回滚正常
- [x] 状态刷新正常

## 错误类型系统
- [x] 验证错误正确显示
- [x] Git 错误正确显示
- [x] 错误链正确传播

## 发现的问题
（记录）

## 结论
Phase 2 重构测试通过，代码质量显著提升。
```

**Step 6: 提交测试报告**

```bash
git add tmp/test-report-phase2.md
git commit -m "test: 添加 Phase 2 测试报告"
```

---

### Task 17: 更新文档

**Files:**
- Modify: `CLAUDE.md`
- Create: `docs/architecture/frontend-status-cache.md`
- Create: `docs/architecture/backend-errors.md`

**Step 1: 更新 CLAUDE.md**

添加新的架构说明：

```markdown
### StatusCache 架构（重构后）

**模块组成：**
- `core.ts`: 核心缓存存储和检索
- `validation.ts`: 数据验证和修复
- `retry.ts`: 重试逻辑和策略

**使用方法：**

```typescript
import { StatusCacheCore, StatusCacheValidation, RetryManager } from '@/stores/statusCache'

const core = new StatusCacheCore()
const validation = new StatusCacheValidation()

// 获取状态
const status = core.getStatus(path)

// 验证数据
const isValid = validation.validateProjectStatus(status.gitStatus)

// 带重试的操作
const result = await RetryManager.execute(() => fetchData())
```

### 错误处理系统（新增）

**领域错误类型：**
- `ValidationError`: 数据验证错误
- `GitOperationError`: Git 操作错误
- `AIProviderError`: AI Provider 错误

**使用方法：**

```go
import apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"

if project.Path == "" {
    return apperrors.NewValidationError("path", "cannot be empty", nil)
}

if err != nil {
    return apperrors.NewGitOperationError("commit", path, err)
}
```
```

**Step 2: 创建 StatusCache 架构文档**

```markdown
# Frontend StatusCache Architecture

## 概述

StatusCache 是项目状态的缓存层，提供高性能的状态管理。

## 模块设计

### Core（核心模块）

职责：
- 缓存存储
- 快速检索
- 批量更新

API：
- `getStatus(path)`: 获取单个项目状态
- `getStatuses(paths)`: 批量获取状态
- `updateCache(path, status)`: 更新状态
- `clearCache(path)`: 清除缓存

### Validation（验证模块）

职责：
- 数据完整性验证
- 自动修复损坏数据
- 确保默认值

API：
- `validateCache(cache)`: 验证缓存对象
- `repairCache(cache)`: 修复缓存数据
- `ensureDefaults(cache)`: 确保默认值

### Retry（重试模块）

职责：
- 失败重试策略
- 指数退避算法
- 批量操作管理

API：
- `withRetry(operation, options)`: 带重试的操作
- `RetryManager.execute()`: 使用默认配置执行重试
- `RetryManager.executeBatch()`: 批量重试

## 最佳实践

1. **缓存优先**: 优先使用缓存数据
2. **后台刷新**: 过期后在后台刷新
3. **乐观更新**: 用户操作后立即更新 UI
4. **错误恢复**: 失败时回滚或降级
```

**Step 3: 创建错误处理文档**

```markdown
# Backend Error Handling System

## 错误类型层次

```
error (interface)
  ├── AppInitError
  ├── ValidationError
  ├── GitOperationError
  └── AIProviderError
```

## 错误处理最佳实践

### 1. 创建错误

```go
// 验证错误
return apperrors.NewValidationError("path", "cannot be empty", nil)

// Git 错误
return apperrors.NewGitOperationError("commit", projectPath, err)

// AI Provider 错误
return apperrors.NewAIProviderError("openai", "rate limit exceeded", err)
```

### 2. 检查错误类型

```go
if apperrors.IsValidationError(err) {
    // 处理验证错误
}

if apperrors.IsGitError(err) {
    // 处理 Git 错误
}

if apperrors.IsNotFoundError(err) {
    // 处理未找到错误
}
```

### 3. 错误包装

```go
if err != nil {
    return fmt.Errorf("failed to save project: %w", err)
}
```

## 错误传播规则

1. **不要丢弃错误**: 所有错误都必须处理或传播
2. **添加上下文**: 包装错误时添加有意义的上下文
3. **使用正确的错误类型**: 根据错误性质选择类型
4. **日志记录**: 在适当的层级记录错误日志
```

**Step 4: 提交文档**

```bash
git add CLAUDE.md
git add docs/architecture/frontend-status-cache.md
git add docs/architecture/backend-errors.md
git commit -m "docs: 更新架构文档反映 Phase 2 重构"
```

---

### Task 18: 创建 Phase 3 计划

**Files:**
- Create: `docs/plans/2026-02-04-code-optimization-phase3.md`

**Step 1: 创建 Phase 3 计划**

```markdown
# Code Optimization Phase 3: Final Polish

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 完成代码优化的最后阶段，包括代码清理、性能优化、文档完善等

**计划内容:**
1. 清理 tmp 目录和临时代码
2. 统一日志输出格式
3. 统一代码风格（导入排序、命名规范）
4. 性能优化（减少不必要的渲染、优化并发）
5. 添加更多集成测试
6. 完善开发文档

**预计任务数:** 10-15 个任务
```

**Step 2: 提交计划**

```bash
git add docs/plans/2026-02-04-code-optimization-phase3.md
git commit -m "docs: 添加 Phase 3 优化计划（待完善）"
```

---

## Summary

Phase 2 重构包含 18 个主要任务：

**已完成模块：**
- ✅ 事件常量定义（Task 1-4）
- ✅ StatusCache 拆分（Task 5-7）
- ✅ Git 操作包装器（Task 8-9）
- ✅ 领域错误类型（Task 10-11）
- ✅ Repository 接口抽象（Task 12-13）
- ✅ 测试覆盖（Task 14-16）
- ✅ 文档更新（Task 17-18）

**预期结果：**
- StatusCache 从 657 行减少到约 200 行（主文件）
- 重复代码减少约 40%
- 错误处理更加规范和一致
- 代码可测试性显著提升

**下一步：** Phase 3 将进行最后的清理和完善工作。

---

**计划完成时间:** 2026-02-04
**预计总工作量:** 12-16 小时
**风险等级:** 中等（需要充分测试各个模块）
