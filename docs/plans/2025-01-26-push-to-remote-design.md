# 推送到远程仓库功能设计

## 概述

为 AI Commit Hub 添加推送到远程仓库的功能，允许用户在本地提交后一键推送更改到远程 Git 仓库。

## 功能需求

### 核心功能
- 在"提交到本地"按钮旁边添加"推送到远程"按钮
- 只有在成功提交到本地后，推送按钮才可用
- 推送到当前分支的同名远程分支
- 推送失败时显示错误信息，用户手动解决

### 用户交互流程
1. 用户生成 Commit 消息
2. 点击"提交到本地" → 成功后"推送到远程"按钮变为可用
3. 点击"推送到远程" → 按钮显示加载状态
4. 推送成功 → 显示成功通知，按钮禁用
5. 推送失败 → 显示错误通知，按钮保持可用

## 技术设计

### 后端实现（Go）

#### 新增 API 方法 (`app.go`)

```go
// PushToRemote 推送到远程仓库
func (a *App) PushToRemote(projectPath string) error {
    if a.initError != nil {
        return a.initError
    }

    // 切换到项目目录
    originalDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %w", err)
    }

    if err := os.Chdir(projectPath); err != nil {
        return fmt.Errorf("failed to change directory: %w", err)
    }
    defer os.Chdir(originalDir)

    // 调用 git 包执行推送
    if err := git.PushToRemote(context.Background()); err != nil {
        return err
    }

    return nil
}
```

#### Git 操作 (`pkg/git/git.go`)

新增推送函数：

```go
// PushToRemote 推送当前分支到远程仓库
func PushToRemote(ctx context.Context) error {
    repo, err := gogit.PlainOpen(".")
    if err != nil {
        return fmt.Errorf("failed to open repository: %w", err)
    }

    // 获取当前分支
    headRef, err := repo.Head()
    if err != nil {
        return fmt.Errorf("failed to get HEAD reference: %w", err)
    }

    branchName := headRef.Name().Short()

    // 执行推送
    if err := repo.Push(&gogit.PushOptions{
        RemoteName: "origin",
        RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName))},
    }); err != nil {
        return fmt.Errorf("push failed: %w", err)
    }

    return nil
}
```

### 前端实现（Vue 3）

#### UI 改动 (`CommitPanel.vue`)

**新增按钮**：
```vue
<button
  @click="handlePush"
  class="btn-action btn-primary-push"
  :disabled="!canPush || isPushing"
>
  <span class="icon" :class="{ spin: isPushing }">↑</span>
  {{ isPushing ? '推送中...' : '推送到远程' }}
</button>
```

**新增状态**：
```typescript
const canPush = ref(false)      // 推送按钮是否可用
const isPushing = ref(false)    // 是否正在推送
```

**修改提交成功处理**：
```typescript
async function handleCommit() {
  // ... 现有代码 ...

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)

    // 保存历史记录
    const project = projectStore.projects.find(p => p.path === commitStore.selectedProjectPath)
    if (project) {
      await SaveCommitHistory(project.id, message, commitStore.provider, commitStore.language)
    }

    showToast('success', '提交成功!')
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)
    await loadHistoryForProject()
    commitStore.clearMessage()

    // 启用推送按钮
    canPush.value = true
  } catch (e) {
    // 错误处理
    canPush.value = false
  }
}
```

**新增推送处理函数**：
```typescript
async function handlePush() {
  if (!commitStore.selectedProjectPath) {
    showToast('error', '请先选择项目')
    return
  }

  isPushing.value = true
  try {
    await PushToRemote(commitStore.selectedProjectPath)
    showToast('success', '推送成功!')
    canPush.value = false  // 推送成功后禁用按钮
    await commitStore.loadProjectStatus(commitStore.selectedProjectPath)
  } catch (e) {
    const message = e instanceof Error ? e.message : '推送失败'
    showToast('error', '推送失败: ' + message)
  } finally {
    isPushing.value = false
  }
}
```

**状态重置**：
```typescript
// 在项目切换时重置
watch(() => projectStore.selectedProject, async (project) => {
  canPush.value = false
  // ... 现有代码 ...
})

// 在手动刷新时重置
async function handleRefresh() {
  // ... 现有代码 ...
  canPush.value = false
}
```

**导入 API**：
```typescript
import {
  CommitLocally,
  PushToRemote,  // 新增
  GetAvailableTerminals,
  GetProjectHistory,
  OpenInFileExplorer,
  OpenInTerminal,
  SaveCommitHistory
} from '../../wailsjs/go/main/App'
```

#### 样式 (`CommitPanel.vue`)

```css
.btn-primary-push {
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  color: white;
  border-color: #8b5cf6;
}

.btn-primary-push:hover:not(:disabled) {
  background: #7c3aed;
  box-shadow: 0 0 15px rgba(139, 92, 246, 0.4);
}

.btn-primary-push:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
  display: inline-block;
}
```

## 错误处理

### 后端错误类型

| 错误类型 | 错误信息 | 前端处理 |
|---------|---------|---------|
| 无远程仓库 | "no remote repository configured" | 显示"未配置远程仓库" |
| 网络错误 | 连接超时、DNS 失败等 | 显示原始错误信息 |
| 冲突 | "failed to push some refs" | 显示"推送失败，请先拉取远程更新" |
| 认证失败 | "authentication failed" | 显示"认证失败，请检查凭据" |

### 前端错误处理
- 所有错误通过 Toast 通知显示
- 推送失败时保持 `canPush = true`，允许重试
- 推送成功时设置 `canPush = false`，避免重复推送

## 实现步骤

1. **后端实现**
   - 在 `pkg/git/git.go` 中添加 `PushToRemote()` 函数
   - 在 `app.go` 中添加 `PushToRemote()` API 方法
   - 编写单元测试

2. **前端实现**
   - 修改 `CommitPanel.vue` 添加推送按钮和状态
   - 添加 `handlePush()` 函数
   - 修改 `handleCommit()` 函数启用推送按钮
   - 添加样式

3. **测试**
   - 测试正常推送流程
   - 测试各种错误场景
   - 测试按钮状态切换

4. **构建**
   - 运行 `wails dev` 重新生成绑定
   - 验证功能正常工作

## 注意事项

- 推送按钮只在本地提交成功后可用
- 推送成功后自动禁用，避免重复推送
- 切换项目或刷新状态时重置推送按钮状态
- 所有错误都以用户友好的方式显示
