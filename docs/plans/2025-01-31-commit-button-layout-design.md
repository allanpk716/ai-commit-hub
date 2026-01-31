# Commit 按钮布局与分支状态显示设计

**日期:** 2025-01-31
**状态:** 设计完成

## 概述

将"提交到本地"和"推送"按钮移至标题栏与"生成消息"按钮同行显示，并在项目状态头部显示本地分支与远程分支的同步状态（领先/落后数量）。

## 设计目标

1. 提升操作效率：主要操作按钮集中在标题栏，无需滚动
2. 状态可视化：清晰显示分支同步状态，辅助用户决策
3. 布局整洁：所有按钮在一行内，保持界面简洁

## 布局设计

### 标题栏布局

标题栏分为三个区域，从左到右依次排列：

```
┌────────────────────────────────────────────────────────────────────┐
│ ┌──────────┐ ┌──────────────────┐ ┌──────────────────────────┐    │
│ │ 主要操作 │ │   AI 配置        │ │      工具按钮            │    │
│ │ ✨生成   │ │ 🌐Provider 🌍语言 │ │ 自定义 清除×            │    │
│ │ ✓提交    │ │                  │ │                          │    │
│ │ ↑推送    │ │                  │ │                          │    │
│ └──────────┘ └──────────────────┘ └──────────────────────────┘    │
└────────────────────────────────────────────────────────────────────┘
```

**左侧 - 主要操作按钮组：**
- `✨ 生成消息` - 始终显示
- `✓ 提交到本地` - 仅在生成消息后显示
- `↑ 推送` - 仅在生成消息后显示

**中间 - AI 配置控件：**
- 🌐 Provider 选择器
- 🌍 语言选择器

**右侧 - 工具按钮：**
- 自定义配置标记（如适用）
- 清除按钮（有消息时）

### 分支状态显示

在项目状态头部（`ProjectStatusHeader` 组件）的分支徽章旁边显示同步状态：

```
┌─────────────────────────────────────────────────────┐
│ ai-commit-hub  [⑂ main ↑3 ↓2]  [📁] [⌘] [↻]       │
└─────────────────────────────────────────────────────┘
```

**状态徽章样式：**
- `↑3` - 绿色，本地领先 3 个提交（可推送）
- `↓2` - 橙色，本地落后 2 个提交（需要拉取）
- `↑3 ↓2` - 红色，分支分叉（领先 3 个，落后 2 个）
- 无徽章 - 已同步（ahead=0, behind=0）

## 数据结构扩展

### PushStatus 扩展

**Go 后端 (`pkg/git/git.go`)：**
```go
type PushStatus struct {
    CanPush       bool   `json:"canPush"`
    AheadCount    int    `json:"ahead_count"`
    BehindCount   int    `json:"behind_count"`   // 新增
    RemoteBranch  string `json:"remote_branch"`
    Error         string `json:"error,omitempty"`
}
```

**TypeScript 前端 (`frontend/src/types/status.ts`)：**
```typescript
export interface PushStatus {
    canPush: boolean
    aheadCount: number
    behindCount: number   // 新增
    remoteBranch: string
    error?: string
}
```

## 实现细节

### 后端实现

**`pkg/git/git.go` - `GetPushStatus` 函数更新：**

1. 检测远程跟踪分支（现有逻辑）
2. 统计领先数量（现有逻辑）
3. **新增：** 统计落后数量

```go
// Count remote commits ahead of local
cmd3 := Command("git", "rev-list", "--count", "HEAD..@{u}")
cmd3.Dir = projectPath
var behindCount bytes.Buffer
cmd3.Stdout = &behindCount
if err := cmd3.Run(); err != nil {
    behindCount.WriteString("0")  // 失败时默认为 0
}

behind := strings.TrimSpace(behindCount.String())
behindCountInt := 0
if behind != "" {
    behindCountInt, _ = strconv.Atoi(behind)
}

return &PushStatus{
    CanPush:      count > 0,
    AheadCount:   count,
    BehindCount:  behindCountInt,  // 新增
    RemoteBranch: remoteBranchName,
}, nil
```

### 前端实现

**1. CommitPanel.vue - 标题栏重构**

移除原有的 `.action-buttons` 区域，将提交/推送按钮移入标题栏：

```vue
<div class="result-header">
  <div class="header-left">
    <button class="btn-generate-main" @click="handleGenerate">
      <span class="btn-icon">✨</span>
      <span class="btn-text">生成消息</span>
    </button>

    <template v-if="commitStore.streamingMessage || commitStore.generatedMessage">
      <button @click="handleCommit" class="btn-action-inline btn-primary-inline">
        <span class="icon">✓</span>
        提交
      </button>
      <button @click="handlePush" class="btn-action-inline btn-push-inline">
        <span class="icon">↑</span>
        推送
      </button>
    </template>
  </div>

  <div class="header-center"><!-- 配置控件 --></div>
  <div class="header-right"><!-- 工具按钮 --></div>
</div>
```

保留辅助操作按钮（复制、重新生成）在消息区域下方，使用更紧凑的样式。

**2. ProjectStatusHeader.vue - 分支状态徽章**

```vue
<template>
  <div class="branch-badge-wrapper">
    <span class="branch-badge">
      <span class="icon">⑂</span>
      {{ branch }}
    </span>

    <span v-if="syncStatus" class="sync-status-badge" :class="syncStatusClass">
      {{ syncStatusText }}
    </span>
  </div>
</template>

<script setup lang="ts">
const syncStatus = computed(() => {
  const pushStatus = statusCache.getPushStatus(props.projectPath)
  if (!pushStatus) return null

  const ahead = pushStatus.aheadCount || 0
  const behind = pushStatus.behindCount || 0

  if (ahead === 0 && behind === 0) return null

  return { ahead, behind }
})

const syncStatusText = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  let text = ''
  if (ahead > 0) text += `↑${ahead}`
  if (behind > 0) text += (text ? ' ' : '') + `↓${behind}`
  return text
})

const syncStatusClass = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  if (ahead > 0 && behind === 0) return 'status-ahead'
  if (behind > 0 && ahead === 0) return 'status-behind'
  return 'status-diverged'
})
</script>
```

**样式定义：**
```css
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

**3. statusCache.ts - 确保字段正确传递**

无需修改，`refresh` 方法中的 `Promise.all` 会自动获取新增的 `behindCount` 字段。

## 状态更新流程

### 本地提交后
1. 调用 `CommitLocally` 成功
2. 调用 `statusCache.refresh(path, { force: true })`
3. `aheadCount` 自动 +1
4. UI 自动更新

### 推送后
1. 调用 `PushToRemote` 成功
2. 调用 `statusCache.refresh(path, { force: true })`
3. `aheadCount` 归零
4. UI 自动更新，推送按钮禁用

## 边缘情况处理

| 场景 | 行为 |
|------|------|
| 无远程仓库 | 显示 "未配置远程仓库"，不显示同步徽章 |
| 分支分叉 | 显示 `↑3 ↓2` 红色徽章 |
| 已同步 | 不显示同步徽章 |
| 推送中 | 按钮显示 "推送中..." 并禁用，图标旋转 |
| 获取失败 | 使用缓存旧状态并标记 `stale` |

## 文件修改清单

1. `pkg/git/git.go` - 扩展 `PushStatus` 和 `GetPushStatus`
2. `frontend/src/types/status.ts` - 扩展 `PushStatus` 接口
3. `frontend/src/components/CommitPanel.vue` - 重构标题栏布局
4. `frontend/src/components/ProjectStatusHeader.vue` - 添加同步状态徽章
5. `frontend/src/stores/statusCache.ts` - 无需修改（自动兼容）
