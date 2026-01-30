# Status Cache Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现状态缓存层以消除项目切换时的双重刷新和状态闪烁问题

**Architecture:** 创建 StatusCache store 作为状态管理中间层，预加载所有项目状态，切换时优先返回缓存数据，后台静默刷新

**Tech Stack:** Vue 3 Pinia Store, Go Backend, TypeScript, Wails Bindings

---

## Task 1: 创建 StatusCache Store 基础结构

**Files:**
- Create: `frontend/src/stores/statusCache.ts`
- Create: `frontend/src/types/status.ts`

**Step 1: 定义状态类型**

创建 `frontend/src/types/status.ts`:

```typescript
export interface ProjectStatusCache {
  gitStatus: GitStatus | null;
  stagingStatus: StagingStatus | null;
  untrackedCount: number;
  pushoverStatus: PushoverHookStatus | null;
  lastUpdated: number;
  loading: boolean;
  error: string | null;
  stale: boolean;
}

export interface ProjectStatusCacheMap {
  [projectPath: string]: ProjectStatusCache;
}
```

**Step 2: 创建 StatusCache Store 基础框架**

创建 `frontend/src/stores/statusCache.ts`:

```typescript
import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { ProjectStatusCache, ProjectStatusCacheMap } from '@/types/status';
import { GetProjectStatus, GetStagingStatus, GetUntrackedFiles } from '@/wailsjs/go/main/App';
import { GetPushoverHookStatus } from '@/wailsjs/go/main/PushoverService';

const CACHE_TTL = 30000; // 30 seconds

export const useStatusCache = defineStore('statusCache', () => {
  const cache = ref<ProjectStatusCacheMap>({});
  const pendingRequests = ref<Set<string>>(new Set());

  function getStatus(path: string): ProjectStatusCache | undefined {
    return cache.value[path];
  }

  function isExpired(path: string): boolean {
    const entry = cache.value[path];
    if (!entry) return true;
    return Date.now() - entry.lastUpdated > CACHE_TTL;
  }

  return {
    cache,
    getStatus,
    isExpired
  };
});
```

**Step 3: 提交初始结构**

```bash
git add frontend/src/stores/statusCache.ts frontend/src/types/status.ts
git commit -m "feat(status-cache): 创建 StatusCache store 基础结构"
```

---

## Task 2: 实现单个项目状态刷新功能

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 实现 refresh 方法**

在 `statusCache.ts` 中添加 `refresh` 函数:

```typescript
async function refresh(path: string, options?: { force?: boolean }): Promise<void> {
  // 防止并发重复请求
  if (pendingRequests.value.has(path)) {
    return;
  }

  // 如果未强制刷新且未过期，跳过
  if (!options?.force && !isExpired(path)) {
    return;
  }

  pendingRequests.value.add(path);

  // 初始化或标记加载中
  if (!cache.value[path]) {
    cache.value[path] = {
      gitStatus: null,
      stagingStatus: null,
      untrackedCount: 0,
      pushoverStatus: null,
      lastUpdated: 0,
      loading: true,
      error: null,
      stale: false
    };
  } else {
    cache.value[path].loading = true;
  }

  try {
    const [gitStatus, stagingStatus, untrackedFiles, pushoverStatus] = await Promise.all([
      GetProjectStatus(path).catch(() => null),
      GetStagingStatus(path).catch(() => null),
      GetUntrackedFiles(path).catch(() => []),
      GetPushoverHookStatus(path).catch(() => null)
    ]);

    cache.value[path] = {
      gitStatus,
      stagingStatus,
      untrackedCount: untrackedFiles?.length || 0,
      pushoverStatus: pushoverStatus,
      lastUpdated: Date.now(),
      loading: false,
      error: null,
      stale: false
    };
  } catch (error) {
    cache.value[path].loading = false;
    cache.value[path].error = String(error);
  } finally {
    pendingRequests.value.delete(path);
  }
}

// 在 return 中导出
return {
  cache,
  getStatus,
  isExpired,
  refresh
};
```

**Step 2: 提交刷新功能**

```bash
git add frontend/src/stores/statusCache.ts
git commit -m "feat(status-cache): 实现单个项目状态刷新功能"
```

---

## Task 3: 实现批量预加载功能

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 实现 preload 和 init 方法**

在 `statusCache.ts` 中添加批量加载函数:

```typescript
async function preload(projectPaths: string[]): Promise<void> {
  const MAX_CONCURRENT = 10;
  const chunks: string[][] = [];

  // 分块处理，避免同时发起太多请求
  for (let i = 0; i < projectPaths.length; i += MAX_CONCURRENT) {
    chunks.push(projectPaths.slice(i, i + MAX_CONCURRENT));
  }

  for (const chunk of chunks) {
    await Promise.all(chunk.map(path => refresh(path, { force: true })));
  }
}

async function init(): Promise<void> {
  // 从 projectStore 获取所有项目路径
  const { useProjectStore } = await import('./project');
  const projectStore = useProjectStore();
  const paths = projectStore.projects.map(p => p.path);
  await preload(paths);
}

// 在 return 中导出
return {
  cache,
  getStatus,
  isExpired,
  refresh,
  preload,
  init
};
```

**Step 2: 提交预加载功能**

```bash
git add frontend/src/stores/statusCache.ts
git commit -m "feat(status-cache): 实现批量预加载功能"
```

---

## Task 4: 后端添加批量查询接口

**Files:**
- Modify: `app.go`

**Step 1: 定义 ProjectFullStatus 结构体**

在 `app.go` 中添加:

```go
type ProjectFullStatus struct {
    GitStatus      *git.GitStatus      `json:"gitStatus"`
    StagingStatus  *git.StagingStatus  `json:"stagingStatus"`
    UntrackedCount int                 `json:"untrackedCount"`
    PushoverStatus *PushoverHookStatus `json:"pushoverStatus"`
    LastUpdated    time.Time           `json:"lastUpdated"`
}
```

**Step 2: 实现 GetAllProjectStatuses 方法**

在 `app.go` 中添加:

```go
const maxConcurrent = 10

func (a *App) GetAllProjectStatuses(projectPaths []string) (map[string]*ProjectFullStatus, error) {
    if a.initError != nil {
        return nil, a.initError
    }

    type result struct {
        path   string
        status *ProjectFullStatus
    }

    sem := make(chan struct{}, maxConcurrent)
    results := make(chan result, len(projectPaths))

    for _, path := range projectPaths {
        sem <- struct{}{}
        go func(p string) {
            defer func() { <-sem }()

            gitStatus, _ := a.gitService.GetStatus(p)
            staging, _ := a.gitService.GetStagingStatus(p)
            untracked, _ := a.gitService.GetUntrackedFiles(p)
            pushover, _ := a.pushoverService.GetHookStatus(p)

            results <- result{
                path: p,
                status: &ProjectFullStatus{
                    GitStatus:      gitStatus,
                    StagingStatus:  staging,
                    UntrackedCount: len(untracked),
                    PushoverStatus: pushover,
                    LastUpdated:    time.Now(),
                },
            }
        }(path)
    }

    statuses := make(map[string]*ProjectFullStatus)
    for i := 0; i < len(projectPaths); i++ {
        r := <-results
        statuses[r.path] = r.status
    }

    return statuses, nil
}
```

**Step 3: 提交后端批量接口**

```bash
git add app.go
git commit -m "feat(backend): 添加 GetAllProjectStatuses 批量查询接口"
```

---

## Task 5: 前端使用后端批量接口优化预加载

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`
- Modify: `wailsjs/go/main/App.js` (由 wails generate 生成)

**Step 1: 重新生成 Wails 绑定**

```bash
wails generate module
```

**Step 2: 更新 preload 方法使用批量接口**

在 `statusCache.ts` 中修改 `preload`:

```typescript
async function preload(projectPaths: string[]): Promise<void> {
  if (projectPaths.length === 0) return;

  try {
    const statuses = await GetAllProjectStatuses(projectPaths);

    for (const [path, status] of Object.entries(statuses)) {
      cache.value[path] = {
        gitStatus: status.gitStatus,
        stagingStatus: status.stagingStatus,
        untrackedCount: status.untrackedCount,
        pushoverStatus: status.pushoverStatus,
        lastUpdated: new Date(status.lastUpdated).getTime(),
        loading: false,
        error: null,
        stale: false
      };
    }
  } catch (error) {
    console.error('Preload failed, falling back to individual loads:', error);
    // 降级到逐个加载
    await Promise.all(projectPaths.map(path => refresh(path, { force: true })));
  }
}
```

**Step 3: 提交优化**

```bash
git add frontend/src/stores/statusCache.ts wailsjs/
git commit -m "feat(status-cache): 使用批量接口优化预加载"
```

---

## Task 6: 更新 ProjectList 使用 StatusCache

**Files:**
- Modify: `frontend/src/components/ProjectList.vue`

**Step 1: 导入 StatusCache**

在 `<script setup>` 中添加:

```typescript
import { useStatusCache } from '@/stores/statusCache';

const statusCache = useStatusCache();
```

**Step 2: 修改状态获取逻辑**

替换原有的状态计算逻辑:

```typescript
const getProjectStatus = (project: GitProject) => {
  const cached = statusCache.getStatus(project.path);

  if (cached?.loading) {
    return { loading: true };
  }

  if (cached?.error) {
    return { error: true, message: cached.error };
  }

  return {
    loading: false,
    error: false,
    untrackedCount: cached?.untrackedCount ?? 0,
    pushoverUpdateAvailable: cached?.pushoverStatus?.updateAvailable ?? false,
    stale: cached?.stale ?? false
  };
};
```

**Step 3: 移除启动时的重复加载**

删除 `onMounted` 中的 `loadProjectsWithStatus` 调用，改为在 `App.vue` 中初始化 StatusCache。

**Step 4: 提交**

```bash
git add frontend/src/components/ProjectList.vue
git commit -m "refactor(project-list): 使用 StatusCache 替代直接调用"
```

---

## Task 7: 更新 CommitPanel 使用 StatusCache

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 导入 StatusCache**

```typescript
import { useStatusCache } from '@/stores/statusCache';

const statusCache = useStatusCache();
```

**Step 2: 修改项目切换 watch**

替换原有的 watch:

```typescript
watch(selectedProject, async (newProject) => {
  if (!newProject) return;

  // 清空之前的状态
  commitMessage.value = '';
  stagedFiles.value = [];
  untrackedFiles.value = [];

  // 立即显示缓存状态
  const cached = statusCache.getStatus(newProject.path);
  if (cached && !cached.loading) {
    updateUIFromCache(cached);
  } else {
    loading.value = true;
  }

  // 后台刷新
  await statusCache.refresh(newProject.path);

  // 刷新完成后更新 UI
  const fresh = statusCache.getStatus(newProject.path);
  if (fresh) {
    updateUIFromCache(fresh);
    loading.value = false;
  }
});

function updateUIFromCache(cached: ProjectStatusCache) {
  if (cached.gitStatus) {
    currentBranch.value = cached.gitStatus.branch;
  }
  if (cached.stagingStatus) {
    stagedFiles.value = cached.stagingStatus.stagedFiles || [];
  }
  // ... 其他 UI 更新
}
```

**Step 3: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(commit-panel): 使用 StatusCache 实现无延迟切换"
```

---

## Task 8: 在 App.vue 中初始化 StatusCache

**Files:**
- Modify: `frontend/src/App.vue`

**Step 1: 添加初始化逻辑**

在 `onMounted` 中添加:

```typescript
onMounted(async () => {
  // 初始化 StatusCache 并预加载
  const { useStatusCache } = await import('./stores/statusCache');
  const statusCache = useStatusCache();
  await statusCache.init();

  // 其他初始化逻辑...
});
```

**Step 2: 添加加载状态**

在模板中添加启动加载指示器:

```vue
<template>
  <div v-if="initialLoading" class="startup-loading">
    <div class="spinner"></div>
    <p>加载项目状态...</p>
  </div>
  <div v-else class="app-container">
    <!-- 原有内容 -->
  </div>
</template>

<script setup lang="ts">
const initialLoading = ref(true);

onMounted(async () => {
  const { useStatusCache } = await import('./stores/statusCache');
  const statusCache = useStatusCache();
  await statusCache.init();

  initialLoading.value = false;

  EventsOn("startup-complete", () => {
    // ...
  });
});
</script>
```

**Step 3: 提交**

```bash
git add frontend/src/App.vue
git commit -m "feat(app): 应用启动时初始化 StatusCache"
```

---

## Task 9: 添加骨架屏组件

**Files:**
- Create: `frontend/src/components/StatusSkeleton.vue`
- Modify: `frontend/src/components/ProjectList.vue`
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 创建骨架屏组件**

创建 `StatusSkeleton.vue`:

```vue
<template>
  <div class="status-skeleton">
    <div class="skeleton-badge"></div>
    <div class="skeleton-text"></div>
  </div>
</template>

<style scoped>
.status-skeleton {
  display: flex;
  align-items: center;
  gap: 8px;
}

.skeleton-badge {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-text {
  width: 60px;
  height: 16px;
  border-radius: 4px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
```

**Step 2: 在 ProjectList 中使用骨架屏**

```vue
<template>
  <div class="project-item">
    <template v-if="getProjectStatus(project).loading">
      <StatusSkeleton />
    </template>
    <template v-else>
      <!-- 正常状态显示 -->
    </template>
  </div>
</template>
```

**Step 3: 提交**

```bash
git add frontend/src/components/StatusSkeleton.vue frontend/src/components/ProjectList.vue
git commit -m "feat(ui): 添加状态加载骨架屏"
```

---

## Task 10: 实现乐观更新功能

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`
- Modify: `frontend/src/stores/commit.ts`

**Step 1: 在 StatusCache 中添加乐观更新方法**

```typescript
function updateOptimistic(path: string, updates: Partial<ProjectStatusCache>): void {
  const current = cache.value[path];
  if (!current) return;

  // 保存当前状态用于可能的回滚
  const previous = { ...current };

  // 应用更新
  cache.value[path] = {
    ...current,
    ...updates,
    lastUpdated: Date.now()
  };

  // 返回回滚函数
  return () => {
    cache.value[path] = previous;
  };
}
```

**Step 2: 在 commitStore 中使用乐观更新**

修改 commit 成功后的处理:

```typescript
async function commitChanges() {
  const { useStatusCache } = await import('./statusCache');
  const statusCache = useStatusCache();

  // 保存回滚函数
  const rollback = statusCache.updateOptimistic(projectPath.value, {
    stagingStatus: { stagedFiles: [] },
    untrackedCount: 0
  });

  try {
    await DoCommit(projectPath.value, commitMessage.value);
    // 验证真实状态
    await statusCache.refresh(projectPath.value);
  } catch (error) {
    rollback?.();
    throw error;
  }
}
```

**Step 3: 提交**

```bash
git add frontend/src/stores/statusCache.ts frontend/src/stores/commit.ts
git commit -m "feat(status-cache): 实现乐观更新和回滚机制"
```

---

## Task 11: 添加错误处理和重试逻辑

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 添加重试辅助函数**

```typescript
async function refreshWithRetry(path: string, maxRetries = 2): Promise<void> {
  for (let i = 0; i <= maxRetries; i++) {
    try {
      await refresh(path, { force: true });
      return;
    } catch (error) {
      if (i === maxRetries) {
        throw error;
      }
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}

function isRetryable(error: unknown): boolean {
  if (error instanceof Error) {
    return error.message.includes('network') ||
           error.message.includes('timeout') ||
           error.message.includes('ECONN');
  }
  return false;
}
```

**Step 2: 添加手动刷新方法**

```typescript
async function manualRefresh(path: string): Promise<void> {
  try {
    await refreshWithRetry(path);
  } catch (error) {
    console.error('Manual refresh failed:', error);
    throw error;
  }
}
```

**Step 3: 导出新方法**

```typescript
return {
  // ...
  refreshWithRetry,
  manualRefresh
};
```

**Step 4: 提交**

```bash
git add frontend/src/stores/statusCache.ts
git commit -m "feat(status-cache): 添加错误处理和重试逻辑"
```

---

## Task 12: 编写单元测试

**Files:**
- Create: `frontend/src/stores/__tests__/statusCache.spec.ts`

**Step 1: 编写 StatusCache 测试**

创建测试文件:

```typescript
import { setActivePinia, createPinia } from 'pinia';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useStatusCache } from '../statusCache';

describe('StatusCache', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it('should cache project status', async () => {
    const cache = useStatusCache();
    // Mock API calls...
    await cache.refresh('/test/path');

    const status = cache.getStatus('/test/path');
    expect(status).toBeDefined();
    expect(status?.loading).toBe(false);
  });

  it('should prevent concurrent duplicate requests', async () => {
    const cache = useStatusCache();
    const refreshSpy = vi.spyOn(cache, 'refresh');

    await Promise.all([
      cache.refresh('/test/path'),
      cache.refresh('/test/path'),
      cache.refresh('/test/path')
    ]);

    expect(refreshSpy).toHaveBeenCalledTimes(1);
  });

  it('should handle TTL expiration', () => {
    const cache = useStatusCache();
    cache.cache['/test'] = {
      gitStatus: null,
      stagingStatus: null,
      untrackedCount: 0,
      pushoverStatus: null,
      lastUpdated: Date.now() - 40000,
      loading: false,
      error: null,
      stale: false
    };

    expect(cache.isExpired('/test')).toBe(true);
  });
});
```

**Step 2: 提交测试**

```bash
git add frontend/src/stores/__tests__/statusCache.spec.ts
git commit -m "test(status-cache): 添加单元测试"
```

---

## Task 13: 手动测试和验证

**Files:**
- Create: `tmp/test_status_cache.md`

**Step 1: 创建测试清单**

创建测试文档:

```markdown
# Status Cache 手动测试清单

## 启动体验
- [ ] 应用启动后所有项目状态同时显示
- [ ] 首次显示时有骨架屏，无"未跟踪"闪烁
- [ ] 约 1 秒内完成所有项目状态加载

## 项目切换
- [ ] 切换已加载项目状态立即显示（<100ms）
- [ ] 切换时显示"更新中..."但不阻塞 UI
- [ ] 后台刷新完成后状态平滑过渡
- [ ] 不出现"未跟踪→正确"的闪烁

## 错误处理
- [ ] 网络断开时显示友好错误提示
- [ ] 过期缓存显示"可能已过期"警告
- [ ] 手动刷新按钮工作正常

## Git 操作
- [ ] 提交后暂存区立即清空（乐观更新）
- [ ] Pushover 状态更新后图标立即变化
```

**Step 2: 执行手动测试**

运行 `wails dev`，按照清单逐项测试。

**Step 3: 提交测试文档**

```bash
git add tmp/test_status_cache.md
git commit -m "test(status-cache): 添加手动测试清单"
```

---

## Task 14: 更新项目文档

**Files:**
- Modify: `CLAUDE.md`

**Step 1: 添加 StatusCache 说明**

在 `CLAUDE.md` 的架构部分添加:

```markdown
### StatusCache 层

`frontend/src/stores/statusCache.ts` 是状态缓存层，位于 UI 组件和后端 API 之间：

- **预加载**: 启动时批量加载所有项目状态
- **缓存优先**: 切换项目时立即返回缓存数据
- **后台刷新**: 静默更新状态以保持新鲜度
- **乐观更新**: 用户操作后立即更新 UI，异步验证
- **错误恢复**: 失败时使用过期缓存或显示友好错误

**使用方法**:
\`\`\`typescript
import { useStatusCache } from '@/stores/statusCache';

const statusCache = useStatusCache();

// 获取缓存状态
const status = statusCache.getStatus(projectPath);

// 刷新状态
await statusCache.refresh(projectPath);

// 批量预加载
await statusCache.preload(projectPaths);
\`\`\`
```

**Step 2: 提交文档更新**

```bash
git add CLAUDE.md
git commit -m "docs(status-cache): 更新项目文档"
```

---

## Task 15: 最终验证和清理

**Step 1: 运行所有测试**

```bash
# 前端测试
cd frontend && npm test

# Go 测试
go test ./... -v
```

**Step 2: 检查 TypeScript 类型**

```bash
cd frontend && npm run type-check
```

**Step 3: 构建验证**

```bash
wails build
```

**Step 4: 清理临时文件**

```bash
rm tmp/test_status_cache.md
```

**Step 5: 最终提交**

```bash
git add -A
git commit -m "feat(status-cache): 完成状态缓存功能实现"
```

---

## 实施顺序

按照任务编号依次执行：

1. Task 1-3: 创建 StatusCache 基础功能
2. Task 4-5: 后端批量接口
3. Task 6-8: 集成到现有组件
4. Task 9-11: UI 优化和错误处理
5. Task 12-15: 测试和文档

每个任务完成后提交，确保可回滚。
