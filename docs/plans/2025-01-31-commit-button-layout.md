# Commit 按钮布局与分支状态显示实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 将提交/推送按钮移至标题栏，并在项目状态头部显示分支同步状态（领先/落后数量）

**架构:** 后端扩展 PushStatus 添加 behindCount 字段，前端重构 CommitPanel 标题栏布局，ProjectStatusHeader 添加同步状态徽章

**技术栈:** Go 1.21+, Vue 3, TypeScript, Wails v2, Git

---

## Task 1: 后端 - 扩展 PushStatus 结构体

**Files:**
- Modify: `pkg/git/git.go:640-646`

**Step 1: 添加 BehindCount 字段**

在 `PushStatus` 结构体中添加 `BehindCount` 字段：

```go
// PushStatus represents the push status of a Git repository.
type PushStatus struct {
    CanPush       bool   `json:"canPush"`
    AheadCount    int    `json:"ahead_count"`
    BehindCount   int    `json:"behind_count"`   // 新增字段
    RemoteBranch  string `json:"remote_branch"`
    Error         string `json:"error,omitempty"`
}
```

**Step 2: 运行测试验证编译通过**

Run: `go build -o tmp/test-build.exe .`
Expected: 编译成功，无错误

**Step 3: 提交**

```bash
git add pkg/git/git.go
git commit -m "feat(git): 扩展 PushStatus 添加 BehindCount 字段"
```

---

## Task 2: 后端 - 更新 GetPushStatus 函数

**Files:**
- Modify: `pkg/git/git.go:648-695`

**Step 1: 在 GetPushStatus 函数中添加统计落后数量的逻辑**

在函数返回前添加统计落后数量的代码：

```go
// GetPushStatus detects whether the local branch is ahead of the remote branch.
func GetPushStatus(projectPath string) (*PushStatus, error) {
    // Check if there's a remote tracking branch
    cmd := Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
    cmd.Dir = projectPath

    var remoteBranch bytes.Buffer
    cmd.Stdout = &remoteBranch
    err := cmd.Run()

    // No remote tracking branch
    if err != nil {
        return &PushStatus{
            CanPush:      false,
            AheadCount:   0,
            BehindCount:  0,  // 新增
            RemoteBranch: "",
            Error:        "未配置远程仓库",
        }, nil
    }

    remoteBranchName := strings.TrimSpace(remoteBranch.String())

    // Count local commits ahead of remote
    cmd2 := Command("git", "rev-list", "--count", "@{u}..HEAD")
    cmd2.Dir = projectPath
    var aheadCount bytes.Buffer
    cmd2.Stdout = &aheadCount
    if err := cmd2.Run(); err != nil {
        return &PushStatus{
            CanPush:      false,
            AheadCount:   0,
            BehindCount:  0,  // 新增
            RemoteBranch: remoteBranchName,
            Error:        "获取推送状态失败",
        }, nil
    }

    ahead := strings.TrimSpace(aheadCount.String())
    count := 0
    if ahead != "" {
        count, _ = strconv.Atoi(ahead)
    }

    // ===== 新增开始：统计落后数量 =====
    // Count remote commits ahead of local
    cmd3 := Command("git", "rev-list", "--count", "HEAD..@{u}")
    cmd3.Dir = projectPath
    var behindCount bytes.Buffer
    cmd3.Stdout = &behindCount
    if err := cmd3.Run(); err != nil {
        // 失败时不阻塞主流程，返回 0
        behindCount.WriteString("0")
    }

    behind := strings.TrimSpace(behindCount.String())
    behindCountInt := 0
    if behind != "" {
        behindCountInt, _ = strconv.Atoi(behind)
    }
    // ===== 新增结束 =====

    return &PushStatus{
        CanPush:      count > 0,
        AheadCount:   count,
        BehindCount:  behindCountInt,  // 新增
        RemoteBranch: remoteBranchName,
    }, nil
}
```

**Step 2: 运行测试验证编译通过**

Run: `go build -o tmp/test-build.exe .`
Expected: 编译成功，无错误

**Step 3: 提交**

```bash
git add pkg/git/git.go
git commit -m "feat(git): 更新 GetPushStatus 函数添加落后数量统计"
```

---

## Task 3: 前端 - 扩展 TypeScript PushStatus 接口

**Files:**
- Modify: `frontend/src/types/status.ts:14-23`

**Step 1: 添加 behindCount 字段**

```typescript
/**
 * 推送状态
 */
export interface PushStatus {
  /** 是否可推送（本地领先远程） */
  canPush: boolean
  /** 本地领先远程的提交数量 */
  aheadCount: number
  /** 本地落后远程的提交数量 */
  behindCount: number
  /** 远程分支名（如 origin/main） */
  remoteBranch: string
  /** 错误信息（无远程仓库等） */
  error?: string
}
```

**Step 2: 运行 TypeScript 类型检查**

Run: `cd frontend && npm run type-check`
Expected: 类型检查通过，无错误

**Step 3: 提交**

```bash
git add frontend/src/types/status.ts
git commit -m "feat(types): 扩展 PushStatus 接口添加 behindCount 字段"
```

---

## Task 4: 前端 - ProjectStatusHeader 添加分支同步状态徽章

**Files:**
- Modify: `frontend/src/components/ProjectStatusHeader.vue`

**Step 1: 在 template 中添加同步状态徽章**

找到分支徽章的位置，在其后添加同步状态徽章：

```vue
<template>
  <div class="project-status-header">
    <div class="header-left">
      <span class="project-name">{{ projectName }}</span>

      <!-- 分支徽章 + 同步状态 -->
      <div class="branch-badge-wrapper">
        <span class="branch-badge">
          <span class="icon">⑂</span>
          {{ branch }}
        </span>

        <!-- 同步状态徽章 -->
        <span v-if="syncStatus" class="sync-status-badge" :class="syncStatusClass">
          {{ syncStatusText }}
        </span>
      </div>
    </div>

    <div class="header-right">
      <!-- 现有的操作按钮保持不变 -->
    </div>
  </div>
</template>
```

**Step 2: 在 script 中添加同步状态计算属性**

```typescript
import { computed } from 'vue'
import { useStatusCache } from '@/stores/statusCache'

// ... 现有代码 ...

const statusCache = useStatusCache()

// 分支同步状态
const syncStatus = computed(() => {
  if (!props.projectPath) return null
  const pushStatus = statusCache.getPushStatus(props.projectPath)
  if (!pushStatus) return null

  const ahead = pushStatus.aheadCount || 0
  const behind = pushStatus.behindCount || 0

  // 如果同步了，不显示徽章
  if (ahead === 0 && behind === 0) return null

  return { ahead, behind }
})

// 同步状态文本
const syncStatusText = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  let text = ''
  if (ahead > 0) text += `↑${ahead}`
  if (behind > 0) text += (text ? ' ' : '') + `↓${behind}`
  return text
})

// 同步状态样式类
const syncStatusClass = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  if (ahead > 0 && behind === 0) return 'status-ahead'
  if (behind > 0 && ahead === 0) return 'status-behind'
  return 'status-diverged'
})
```

**Step 3: 添加同步状态徽章样式**

```css
.branch-badge-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.sync-status-badge {
  padding: 2px 6px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.sync-status-badge.status-ahead {
  background: rgba(16, 185, 129, 0.2);
  color: var(--accent-success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.sync-status-badge.status-behind {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.sync-status-badge.status-diverged {
  background: rgba(239, 68, 68, 0.2);
  color: var(--accent-error);
  border: 1px solid rgba(239, 68, 68, 0.3);
}
```

**Step 4: 运行开发服务器验证**

Run: `wails dev`
Expected: 开发服务器启动，分支徽章旁显示同步状态

**Step 5: 提交**

```bash
git add frontend/src/components/ProjectStatusHeader.vue
git commit -m "feat(ui): 在 ProjectStatusHeader 添加分支同步状态徽章"
```

---

## Task 5: 前端 - CommitPanel 标题栏重构（添加按钮到标题栏）

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 修改 result-header 布局，在 header-left 添加提交和推送按钮**

找到 `result-header` 区域的 `header-left`，在"生成消息"按钮后添加提交和推送按钮：

```vue
<div class="result-header">
  <!-- 左侧：操作按钮组 -->
  <div class="header-left">
    <button
      @click="handleGenerate"
      :disabled="!commitStore.hasStagedFiles || commitStore.isGenerating"
      class="btn-generate-main"
      :class="{ generating: commitStore.isGenerating }"
      title="生成 Commit 消息"
    >
      <span class="btn-icon">✨</span>
      <span class="btn-text" v-if="!commitStore.isGenerating">生成消息</span>
      <span class="btn-text" v-else>生成中...</span>
    </button>

    <!-- 新增：提交和推送按钮（仅在有消息时显示） -->
    <template v-if="commitStore.streamingMessage || commitStore.generatedMessage">
      <button
        @click="handleCommit"
        class="btn-action-inline btn-primary-inline"
        :disabled="!commitStore.hasStagedFiles"
        title="提交到本地"
      >
        <span class="icon">✓</span>
        提交
      </button>
      <button
        @click="handlePush"
        class="btn-action-inline btn-push-inline"
        :disabled="isPushing || !pushStatus?.canPush"
        :title="pushStatus?.aheadCount ? `领先 ${pushStatus.aheadCount} 个提交` : pushStatus?.error || '无待推送内容'"
      >
        <span class="icon" :class="{ spin: isPushing }">↑</span>
        {{ isPushing ? '推送中' : '推送' }}
      </button>
    </template>
  </div>

  <!-- 中间：配置控件（保持不变） -->
  <div class="header-center">...</div>

  <!-- 右侧：工具按钮（保持不变） -->
  <div class="header-right">...</div>
</div>
```

**Step 2: 移除原有的 action-buttons 区域中的提交和推送按钮**

找到原有的 `action-buttons` 区域（约在第 107-129 行），移除"提交到本地"和"推送"按钮，保留"复制"和"重新生成"按钮：

```vue
<!-- 修改后的 action-buttons：只保留辅助操作 -->
<div class="action-buttons-helper" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
  <button @click="handleCopy" class="btn-action btn-secondary">
    <span class="icon">📋</span>
    复制
  </button>
  <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-action btn-tertiary">
    <span class="icon">🔄</span>
    重新生成
  </button>
</div>
```

**Step 3: 添加新的按钮样式**

在 style 区域添加：

```css
/* 紧凑型操作按钮（标题栏内） */
.btn-action-inline {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 8px 14px;
  border: none;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.btn-action-inline .icon {
  font-size: 14px;
  line-height: 1;
}

.btn-primary-inline {
  background: var(--accent-success);
  color: white;
}

.btn-primary-inline:hover:not(:disabled) {
  background: #059669;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.4);
}

.btn-primary-inline:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-push-inline {
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  color: white;
}

.btn-push-inline:hover:not(:disabled) {
  background: #7c3aed;
  box-shadow: 0 0 12px rgba(139, 92, 246, 0.4);
}

.btn-push-inline:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* 辅助操作按钮区域（消息下方） */
.action-buttons-helper {
  display: flex;
  gap: var(--space-sm);
  justify-content: flex-start;
}
```

**Step 4: 更新 header-left 样式确保按钮对齐**

```css
.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex-shrink: 0;
  flex-wrap: wrap;  /* 允许在小屏幕上换行 */
}
```

**Step 5: 运行开发服务器验证**

Run: `wails dev`
Expected: 标题栏显示"生成消息"、"提交"、"推送"按钮，消息下方只保留"复制"和"重新生成"

**Step 6: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "feat(ui): 将提交和推送按钮移至标题栏"
```

---

## Task 6: 测试和验证

**Files:**
- Manual testing

**Step 1: 启动开发服务器**

Run: `wails dev`

**Step 2: 验证分支同步状态显示**

1. 选择一个有远程仓库的项目
2. 检查项目状态头部是否显示同步状态徽章（如 `↑3 ↓2`）
3. 创建新提交，验证 `aheadCount` 增加
4. 推送到远程，验证 `aheadCount` 归零

**Step 3: 验证按钮布局**

1. 点击"生成消息"按钮
2. 验证标题栏显示"提交"和"推送"按钮
3. 点击"提交"，验证提交成功
4. 验证"推送"按钮状态正确更新

**Step 4: 测试边缘情况**

1. 无远程仓库的项目：验证不显示同步徽章，推送按钮显示错误提示
2. 分支分叉：验证红色徽章显示
3. 已同步：验证不显示同步徽章

**Step 5: 最终提交**

```bash
git add -A
git commit -m "test: 验证分支同步状态和按钮布局功能"
```

---

## 任务完成检查清单

- [ ] 后端 `PushStatus` 结构体包含 `BehindCount` 字段
- [ ] 后端 `GetPushStatus` 函数正确统计落后数量
- [ ] 前端 `PushStatus` 接口包含 `behindCount` 字段
- [ ] `ProjectStatusHeader` 显示分支同步状态徽章
- [ ] `CommitPanel` 标题栏包含所有操作按钮
- [ ] 按钮状态正确响应项目状态变化
- [ ] 所有边缘情况正确处理
