# 状态缓存统一架构实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 统一项目状态管理，消除启动时状态更新延迟和切换项目时的卡顿感

**Architecture:** StatusCache 作为唯一状态数据源，PushoverStore 改从 StatusCache 读取，App.vue 统一启动流程并行执行所有初始化任务

**Tech Stack:** Vue 3 Pinia, TypeScript, Wails Events, Go Backend

---

## Task 1: StatusCache 添加 getPushoverStatus 方法

**Files:**
- Modify: `frontend/src/stores/statusCache.ts`

**Step 1: 添加 getPushoverStatus 方法**

在 `statusCache.ts` 中的返回对象前添加:

```typescript
/**
 * 获取项目的 Pushover 状态
 * @param path 项目路径
 * @returns Pushover Hook 状态，如果不存在则返回 null
 */
function getPushoverStatus(path: string): HookStatus | null {
  const cached = cache.value[path]
  return cached?.pushoverStatus || null
}
```

**Step 2: 导出新方法**

在 `return` 语句中添加:

```typescript
return {
  // ... 现有导出
  getPushoverStatus
}
```

**Step 3: 提交**

```bash
git add frontend/src/stores/statusCache.ts
git commit -m "feat(status-cache): 添加 getPushoverStatus 方法"
```

---

## Task 2: PushoverStore 添加 computed 属性

**Files:**
- Modify: `frontend/src/stores/pushoverStore.ts`

**Step 1: 导入 StatusCache**

在文件顶部添加导入:

```typescript
import { useStatusCache } from './statusCache'
import type { ProjectStatusCache } from '../types/status'
```

**Step 2: 添加 computed 属性**

在 `isUpdateAvailable` computed 后添加:

```typescript
/**
 * 当前选中项目的 Pushover 状态（从 StatusCache 获取）
 */
const currentProjectHookStatus = computed(() => {
  const { useProjectStore } = require('./projectStore')
  const projectStore = useProjectStore()
  const selectedPath = projectStore.selectedProject
  if (!selectedPath) return null

  const statusCache = useStatusCache()
  const cached = statusCache.getStatus(selectedPath)
  return cached?.pushoverStatus || null
})

/**
 * 批量获取所有项目的 Pushover 状态（从 StatusCache 获取）
 * 供 ProjectList 使用
 */
const allProjectHookStatuses = computed(() => {
  const statusCache = useStatusCache()
  const statuses: Record<string, HookStatus> = {}
  for (const [path, cache] of Object.entries(statusCache.cache)) {
    if (cache.pushoverStatus) {
      statuses[path] = cache.pushoverStatus
    }
  }
  return statuses
})
```

**Step 3: 导出 computed 属性**

在 `return` 语句中添加:

```typescript
return {
  // State
  extensionInfo,
  // - projectHookStatus,  // 移除
  // - statusVersion,       // 移除
  loading,
  error,
  configValid,
  updateCheckError,
  isCheckingUpdate,

  // Computed
  isExtensionDownloaded,
  isUpdateAvailable,
  currentProjectHookStatus,    // 新增
  allProjectHookStatuses,      // 新增

  // Actions
  // ...
}
```

**Step 4: 提交**

```bash
git add frontend/src/stores/pushoverStore.ts
git commit -m "refactor(pushover-store): 添加从 StatusCache 读取状态的 computed"
```

---

## Task 3: PushoverStore 移除重复状态

**Files:**
- Modify: `frontend/src/stores/pushoverStore.ts`

**Step 1: 删除重复的状态声明**

删除以下代码:

```typescript
- const projectHookStatus = ref<Map<string, HookStatus>>(new Map())
- const statusVersion = ref(0) // 版本号，用于触发响应式更新
```

**Step 2: 修改 getProjectHookStatus 方法**

更新方法实现，不再存储到 Map:

```typescript
async function getProjectHookStatus(projectPath: string): Promise<HookStatus | null> {
  console.log('[DEBUG pushoverStore] getProjectHookStatus called for:', projectPath)
  loading.value = true
  error.value = null

  try {
    // 不再存储到 Map，直接刷新 StatusCache
    const { useStatusCache } = await import('./statusCache')
    const statusCache = useStatusCache()
    await statusCache.refresh(projectPath, { force: true })

    const status = statusCache.getPushoverStatus(projectPath)
    console.log('[DEBUG pushoverStore] Status from cache:', status)
    return status
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    console.error('[DEBUG pushoverStore] Error:', e)
    error.value = `获取 Hook 状态失败: ${message}`
    return null
  } finally {
    loading.value = false
  }
}
```

**Step 3: 移除 clearCache 方法中的 Map 操作**

更新 `clearCache` 方法:

```typescript
function clearCache() {
  // - projectHookStatus.value.clear()  // 移除
  // StatusCache 会自动管理缓存
}
```

**Step 4: 提交**

```bash
git add frontend/src/stores/pushoverStore.ts
git commit -m "refactor(pushover-store): 移除重复的状态存储，改从 StatusCache 读取"
```

---

## Task 4: CommitPanel 适配新的状态源

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 修改 pushoverStatus computed**

找到 `pushoverStatus` computed 并替换:

```typescript
// Pushover Hook 状态 - 从 StatusCache 获取
const pushoverStatus = computed(() => {
  if (currentProjectPath.value) {
    const cached = statusCache.getStatus(currentProjectPath.value)
    return cached?.pushoverStatus || null
  }
  return null
})
```

**Step 2: 移除 pushoverStore 依赖更新**

修改 `updateUIFromCache` 函数，移除 Pushover 状态同步:

```typescript
// 从缓存更新 UI 状态
function updateUIFromCache(cached: any) {
  if (cached.gitStatus) {
    commitStore.projectStatus = cached.gitStatus
  }
  if (cached.stagingStatus) {
    commitStore.stagingStatus = cached.stagingStatus
  }
  // - 移除 Pushover 状态同步，已由组件直接从 StatusCache 读取
}
```

**Step 3: 修改切换项目时的刷新逻辑**

更新 `watch(projectStore.selectedProject)`:

```typescript
// 监听选中的项目变化
watch(() => projectStore.selectedProject, async (project) => {
  if (project) {
    commitStore.clearMessage()
    canPush.value = false

    // 加载 AI 配置（不使用缓存）
    await commitStore.loadProjectAIConfig(project.id)

    // 策略C：优先显示缓存，过期时等待刷新
    const cached = statusCache.getStatus(project.path)

    if (cached && !cached.loading && !statusCache.isExpired(project.path)) {
      // 缓存未过期，立即显示
      updateUIFromCache(cached)
    } else {
      // 缓存过期或不存在，等待刷新
      await statusCache.refresh(project.path, { force: true })
      const fresh = statusCache.getStatus(project.path)
      if (fresh) {
        updateUIFromCache(fresh)
      }
    }
  } else {
    commitStore.clearStagingState()
  }
}, { immediate: true })
```

**Step 4: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(commit-panel): 适配从 StatusCache 读取 Pushover 状态"
```

---

## Task 5: App.vue 创建统一初始化函数

**Files:**
- Modify: `frontend/src/App.vue`

**Step 1: 添加 initializeApp 函数**

在 `<script setup>` 中添加:

```typescript
async function initializeApp() {
  console.log('[App] 开始初始化应用')

  const tasks = [
    // 任务1：加载项目列表
    projectStore.loadProjects()
      .catch(err => ({ error: 'loadProjects', message: err?.message || '未知错误' })),

    // 任务2：预加载所有项目状态
    (async () => {
      const { useStatusCache } = await import('./stores/statusCache')
      const statusCache = useStatusCache()
      return statusCache.init()
        .catch(err => ({ error: 'statusCache', message: err?.message || '未知错误' }))
    })(),

    // 任务3：检查 Pushover 扩展状态
    pushoverStore.checkExtensionStatus()
      .catch(err => ({ error: 'extensionStatus', message: err?.message || '未知错误' })),

    // 任务4：检查 Pushover 配置
    pushoverStore.checkPushoverConfig()
      .catch(err => ({ error: 'pushoverConfig', message: err?.message || '未知错误' }))
  ]

  // 并行执行所有任务
  const results = await Promise.all(tasks)

  // 收集错误（不阻塞启动）
  const errors = results.filter(r => r && r.error)
  if (errors.length > 0) {
    console.warn('[App] 部分初始化任务失败:', errors)
    // TODO: 可选：在 UI 上显示警告提示
  }

  console.log('[App] 应用初始化完成')
  return errors
}
```

**Step 2: 修改 onMounted 使用新函数**

替换现有的 `onMounted`:

```typescript
onMounted(async () => {
  console.log('[App] onMounted 开始')

  // 执行统一初始化
  await initializeApp()

  // 隐藏启动画面
  console.log('[App] 设置 initialLoading = false')
  initialLoading.value = false

  // 延迟隐藏 SplashScreen
  setTimeout(() => {
    console.log('[App] 隐藏 SplashScreen')
    showSplash.value = false
  }, 500)

  // 监听启动完成事件（备用）
  EventsOn("startup-complete", () => {
    console.log('[App] startup-complete 事件触发')
    showSplash.value = false
  })

  console.log('[App] onMounted 完成')
})
```

**Step 3: 移除冗余的 Pushover 配置检查**

删除 onMounted 中单独的 `checkPushoverConfig` 调用（已在 `initializeApp` 中处理）:

```typescript
- // 检查 Pushover 配置
- await pushoverStore.checkPushoverConfig()
- if (!pushoverStore.configValid) {
-   console.warn('Pushover 环境变量未配置，通知功能可能不可用')
- }
```

**Step 4: 提交**

```bash
git add frontend/src/App.vue
git commit -m "refactor(app): 创建统一初始化函数，并行执行所有启动任务"
```

---

## Task 6: ExtensionStatusButton 移除重复加载

**Files:**
- Modify: `frontend/src/components/ExtensionStatusButton.vue`

**Step 1: 删除 onMounted 中的状态检查**

移除以下代码:

```typescript
- onMounted(async () => {
-   await pushoverStore.checkExtensionStatus()
- })
```

**Step 2: 移除 onMounted 导入**

如果 `onMounted` 不再使用，删除导入:

```typescript
- import { computed, onMounted } from 'vue'
+ import { computed } from 'vue'
```

**Step 3: 提交**

```bash
git add frontend/src/components/ExtensionStatusButton.vue
git commit -m "refactor(extension-button): 移除重复的扩展状态检查"
```

---

## Task 7: ProjectList 适配（验证）

**Files:**
- Modify: `frontend/src/components/ProjectList.vue`

**Step 1: 验证现有实现**

确认 `getProjectStatus` 函数从 `statusCache` 读取:

```typescript
const getProjectStatus = (project: GitProject) => {
  const cached = statusCache.getStatus(project.path)
  // ...
}
```

**Step 2: 验证 Pushover 状态读取**

确认 Pushover 更新标记从缓存读取:

```typescript
pushoverUpdateAvailable: cached?.pushoverStatus?.updateAvailable ?? false
```

**注意**: ProjectList 已经正确使用 StatusCache，此任务仅作验证。

**Step 3: 如需修改则提交**

```bash
git add frontend/src/components/ProjectList.vue
git commit -m "refactor(project-list): 验证 StatusCache 集成"
```

---

## Task 8: 添加错误处理 UI

**Files:**
- Modify: `frontend/src/App.vue`

**Step 1: 添加错误状态**

在 `<script setup>` 中添加:

```typescript
const initErrors = ref<Array<{ error: string; message: string }>>([])
```

**Step 2: 修改 initializeApp 收集错误**

更新 `initializeApp` 函数:

```typescript
async function initializeApp() {
  // ... 现有代码

  // 收集错误
  const errors = results.filter(r => r && r.error)
  if (errors.length > 0) {
    console.warn('[App] 部分初始化任务失败:', errors)
    initErrors.value = errors
  }

  return errors
}
```

**Step 3: 添加错误提示 UI**

在模板中添加错误横幅:

```vue
<template>
  <!-- 主应用 -->
  <div v-else class="app grid-pattern">
    <!-- 错误横幅 -->
    <transition name="slide-down">
      <div v-if="initErrors.length > 0" class="init-error-banner">
        <span class="icon">⚠️</span>
        <span class="message">部分功能加载失败，请稍后手动刷新</span>
        <button @click="initErrors = []" class="dismiss">×</button>
      </div>
    </transition>

    <!-- 原有内容 -->
  </div>
</template>

<style scoped>
.init-error-banner {
  position: fixed;
  top: var(--space-lg);
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-md);
  z-index: var(--z-modal);
  animation: slide-down 0.3s ease-out;
}

.init-error-banner .icon {
  font-size: 18px;
}

.init-error-banner .message {
  font-size: 13px;
  color: var(--accent-warning);
}

.init-error-banner .dismiss {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 18px;
  padding: 0 4px;
}

.init-error-banner .dismiss:hover {
  color: var(--text-primary);
}

@keyframes slide-down {
  from {
    opacity: 0;
    transform: translate(-50%, -20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}
</style>
```

**Step 4: 提交**

```bash
git add frontend/src/App.vue
git commit -m "feat(app): 添加初始化错误提示横幅"
```

---

## Task 9: 更新文档

**Files:**
- Modify: `CLAUDE.md`

**Step 1: 更新 StatusCache 层文档**

在 StatusCache 层说明中添加 Pushover 状态管理:

```markdown
### StatusCache 层

`frontend/src/stores/statusCache.ts` 是状态缓存层，用于优化项目状态的加载和更新性能。

**核心功能：**

1. **预加载（Preload）**: 应用启动时批量加载所有项目状态，避免 UI 闪烁
2. **缓存优先（Cache-First）**: 切换项目时立即返回缓存数据，提供快速响应
3. **后台刷新（Background Refresh）**: 静默更新过期缓存以保持数据新鲜度
4. **乐观更新（Optimistic Updates）**: 用户操作后立即更新 UI，异步验证结果
5. **错误恢复（Error Recovery）**: 失败时使用过期缓存或显示友好错误提示

**统一状态管理：**

StatusCache 是项目状态的唯一数据源，管理以下内容：
- Git 状态（分支、提交信息等）
- 暂存区状态（已暂存文件）
- 未跟踪文件数量
- **Pushover Hook 状态**（是否安装、版本信息等）

**使用方法：**

\`\`\`typescript
import { useStatusCache } from '@/stores/statusCache'

const statusCache = useStatusCache()

// 获取缓存状态（立即返回，无等待）
const status = statusCache.getStatus(projectPath)

// 获取 Pushover 状态
const pushoverStatus = statusCache.getPushoverStatus(projectPath)

// 刷新状态（如果缓存未过期可能跳过）
await statusCache.refresh(projectPath)

// 强制刷新（忽略 TTL）
await statusCache.refresh(projectPath, { force: true })

// 批量预加载
await statusCache.preload(projectPaths)
\`\`\`
```

**Step 2: 更新 PushoverStore 说明**

更新 PushoverStore 的职责描述:

```markdown
**PushoverStore (`pushoverStore.ts`)**: Pushover 扩展和配置管理
- 扩展信息（下载状态、版本信息、更新检查）
- Pushover 环境变量配置验证
- 项目 Hook 状态从 StatusCache 读取（不重复存储）
```

**Step 3: 提交文档更新**

```bash
git add CLAUDE.md
git commit -m "docs: 更新 StatusCache 和 PushoverStore 文档说明"
```

---

## Task 10: 手动测试和验证

**Files:**
- Create: `tmp/test_unification.md`

**Step 1: 创建测试清单**

创建测试文档:

```markdown
# 状态缓存统一架构测试清单

## 启动体验
- [ ] 启动画面隐藏时，左侧项目列表所有状态已显示
- [ ] Pushover 扩展状态在启动画面期间检查完成（< 1 秒）
- [ ] 无"先显示空，后更新状态"的闪烁

## 切换项目体验
- [ ] 切换到已缓存项目，状态立即显示（< 100ms）
- [ ] 缓存过期时，显示加载状态，刷新完成后显示
- [ ] 项目列表和详情页的 Pushover 状态完全同步

## 错误处理
- [ ] 初始化失败时显示警告横幅，应用继续运行
- [ ] 单个项目状态加载失败不影响其他项目
- [ ] 手动刷新按钮可恢复失败的状态

## 数据一致性
- [ ] ProjectList 和 CommitPanel 显示相同的 Pushover 状态
- [ ] 安装/更新 Pushover Hook 后，两处状态同步更新
- [ ] 切换项目后返回，状态保持一致
```

**Step 2: 执行测试**

运行 `wails dev`，按照清单逐项测试并记录结果。

**Step 3: 提交测试文档**

```bash
git add tmp/test_unification.md
git commit -m "test: 添加状态缓存统一架构测试清单"
```

---

## Task 11: 最终验证和清理

**Step 1: 运行所有测试**

```bash
# 前端测试
cd frontend && npm run test:run

# Go 测试
go test ./... -v
```

**Step 2: 类型检查**

```bash
cd frontend && npm run type-check
```

**Step 3: 构建验证**

```bash
wails build
```

**Step 4: 清理测试文件**

```bash
rm tmp/test_unification.md
```

**Step 5: 最终提交**

```bash
git add -A
git commit -m "feat(status-cache): 完成状态缓存统一架构实现

- StatusCache 作为唯一项目状态数据源
- PushoverStore 改从 StatusCache 读取状态
- App.vue 统一启动流程，并行执行初始化任务
- 切换项目时缓存优先，过期时等待刷新
- 添加初始化错误处理 UI

Co-Authored-By: Claude (glm-4.7) <noreply@anthropic.com>"
```

---

## 实施顺序

按照任务编号依次执行：

1. **Task 1**: StatusCache 添加方法
2. **Task 2-3**: PushoverStore 改造
3. **Task 4-6**: 组件适配
4. **Task 7**: 验证 ProjectList
5. **Task 8-9**: 错误处理和文档
6. **Task 10-11**: 测试和清理

每个任务完成后提交，确保可回滚。
