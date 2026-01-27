# 推送到远程仓库功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 AI Commit Hub 添加推送到远程仓库的功能，允许用户在本地提交后一键推送更改到远程 Git 仓库。

**Architecture:**
- 后端使用 go-git 库的 Push 方法实现推送操作
- 前端在 CommitPanel 中添加推送按钮和相关状态管理
- 通过 Wails 绑定连接前后端

**Tech Stack:**
- Go 1.21+ + go-git/v5
- Vue 3 + TypeScript + Pinia
- Wails v2

---

## Task 1: 添加 Git 推送功能（后端核心）

**Files:**
- Create: `pkg/git/push.go` (新文件)
- Test: `pkg/git/push_test.go` (新文件)

**Step 1: 编写推送功能的测试**

创建测试文件 `pkg/git/push_test.go`:

```go
package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPushToRemote(t *testing.T) {
	Convey("PushToRemote", t, func() {
		ctx := context.Background()

		Convey("should return error when no remote configured", func() {
			// 创建临时目录作为测试仓库
			tmpDir, err := os.MkdirTemp("", "ai-commit-hub-test-*")
			So(err, ShouldBeNil)
			defer os.RemoveAll(tmpDir)

			// 初始化本地仓库
			repo, err := gogit.PlainInit(tmpDir, false)
			So(err, ShouldBeNil)

			// 创建一个提交
			wt, err := repo.Worktree()
			So(err, ShouldBeNil)

			testFile := filepath.Join(tmpDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			So(err, ShouldBeNil)

			_, err = wt.Add("test.txt")
			So(err, ShouldBeNil)

			commit, err := wt.Commit("test commit", &gogit.CommitOptions{
				Author: &object.Signature{
					Name:  "Test User",
					Email: "test@example.com",
				},
			})
			So(err, ShouldBeNil)

			So(commit, ShouldNotBeZeroValue)

			// 切换到测试目录
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)
			err = os.Chdir(tmpDir)
			So(err, ShouldBeNil)

			// 调用 PushToRemote，应该返回错误
			err = PushToRemote(ctx)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "remote")
		})
	})
}
```

**Step 2: 运行测试验证失败**

运行: `cd pkg/git && go test -v -run TestPushToRemote`
预期: FAIL with "undefined: PushToRemote"

**Step 3: 实现推送功能**

创建文件 `pkg/git/push.go`:

```go
package git

import (
	"context"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

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

**Step 4: 运行测试验证通过**

运行: `cd pkg/git && go test -v -run TestPushToRemote`
预期: PASS

**Step 5: 提交**

```bash
git add pkg/git/push.go pkg/git/push_test.go
git commit -m "feat(git): 添加推送到远程仓库功能

- 新增 PushToRemote 函数推送当前分支到 origin 远程仓库
- 添加测试验证无远程仓库时的错误处理

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: 添加后端 API 方法

**Files:**
- Modify: `app.go:1124` (在文件末尾添加新方法)

**Step 1: 添加 PushToRemote API 方法**

在 `app.go` 文件末尾（`DebugHookStatus` 方法之后）添加:

```go
// PushToRemote 推送项目到远程仓库
func (a *App) PushToRemote(projectPath string) error {
	logger.Infof("PushToRemote 被调用 - projectPath: %s", projectPath)

	if a.initError != nil {
		logger.Errorf("数据库初始化错误: %v", a.initError)
		return a.initError
	}

	// 保存当前目录并切换到项目路径
	originalDir, err := os.Getwd()
	if err != nil {
		err := fmt.Errorf("failed to get current directory: %w", err)
		logger.Errorf("获取当前目录失败: %v", err)
		return err
	}

	if err := os.Chdir(projectPath); err != nil {
		err := fmt.Errorf("failed to change directory: %w", err)
		logger.Errorf("切换到项目目录失败: %v", err)
		return err
	}
	defer os.Chdir(originalDir)

	logger.Infof("准备推送 - 目录: %s", projectPath)

	// 调用 git 包执行推送
	if err := git.PushToRemote(context.Background()); err != nil {
		logger.Errorf("PushToRemote 失败: %v", err)
		return err
	}

	logger.Infof("推送成功 - 目录: %s", projectPath)
	return nil
}
```

**Step 2: 重新生成 Wails 绑定**

运行: `wails dev`
预期: 应用启动成功，在前端控制台无绑定错误

**Step 3: 验证 API 方法可用**

在前端代码中临时添加：
```typescript
import { PushToRemote } from '../../wailsjs/go/main/App'
console.log('PushToRemote available:', typeof PushToRemote)
```
预期: 控制台输出 "PushToRemote available: function"

**Step 4: 删除验证代码**

删除临时添加的验证代码

**Step 5: 提交**

```bash
git add app.go
git commit -m "feat(api): 添加 PushToRemote API 方法

- 新增后端方法用于推送项目到远程仓库
- 自动切换到项目目录执行推送操作
- 添加完整的错误处理和日志记录

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: 前端添加推送按钮（UI）

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue:103-116` (按钮区域)

**Step 1: 添加推送按钮**

在"提交到本地"按钮后面添加"推送到远程"按钮：

找到第 108-111 行的"提交到本地"按钮，在其后添加：

```vue
<button @click="handlePush" class="btn-action btn-primary-push" :disabled="!canPush || isPushing">
  <span class="icon" :class="{ spin: isPushing }">↑</span>
  {{ isPushing ? '推送中...' : '推送到远程' }}
</button>
```

修改后的按钮组应该看起来像：

```vue
<div class="action-buttons" v-if="commitStore.streamingMessage || commitStore.generatedMessage">
  <button @click="handleCopy" class="btn-action btn-secondary">
    <span class="icon">📋</span>
    复制
  </button>
  <button @click="handleCommit" class="btn-action btn-primary" :disabled="!commitStore.projectStatus?.has_staged">
    <span class="icon">✓</span>
    提交到本地
  </button>
  <button @click="handlePush" class="btn-action btn-primary-push" :disabled="!canPush || isPushing">
    <span class="icon" :class="{ spin: isPushing }">↑</span>
    {{ isPushing ? '推送中...' : '推送到远程' }}
  </button>
  <button @click="handleRegenerate" :disabled="commitStore.isGenerating" class="btn-action btn-tertiary">
    <span class="icon">🔄</span>
    重新生成
  </button>
</div>
```

**Step 2: 添加 CSS 样式**

在 `<style scoped>` 部分添加推送按钮样式（在第 1117 行 `.btn-primary` 样式之后）：

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
```

**Step 3: 验证 UI 显示**

运行: `wails dev`
预期: 按钮显示在"重新生成"按钮旁边，初始状态为禁用（灰色）

**Step 4: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "feat(ui): 添加推送到远程按钮

- 在提交到本地按钮旁边添加推送到远程按钮
- 使用紫色渐变样式区分于提交按钮
- 添加加载状态和禁用状态样式

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: 前端状态管理

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue:236-275` (script 部分)

**Step 1: 添加响应式状态**

在 `<script setup lang="ts">` 中，找到其他 ref 声明（约第 264 行 `aiSettingsExpanded`），在其后添加：

```typescript
const canPush = ref(false)      // 推送按钮是否可用
const isPushing = ref(false)    // 是否正在推送
```

**Step 2: 导入 PushToRemote API**

在 import 语句中（约第 237-244 行），添加 `PushToRemote`:

修改前：
```typescript
import {
  CommitLocally,
  GetAvailableTerminals,
  GetProjectHistory,
  OpenInFileExplorer,
  OpenInTerminal,
  SaveCommitHistory
} from '../../wailsjs/go/main/App'
```

修改后：
```typescript
import {
  CommitLocally,
  PushToRemote,
  GetAvailableTerminals,
  GetProjectHistory,
  OpenInFileExplorer,
  OpenInTerminal,
  SaveCommitHistory
} from '../../wailsjs/go/main/App'
```

**Step 3: 修改 handleCommit 函数**

找到 `handleCommit` 函数（约第 411-447 行），在成功处理部分添加 `canPush.value = true`:

```typescript
async function handleCommit() {
  if (!commitStore.selectedProjectPath) {
    showToast('error', '请先选择项目')
    return
  }

  const message = commitStore.streamingMessage || commitStore.generatedMessage
  if (!message) {
    showToast('error', '请先生成 commit 消息')
    return
  }

  try {
    await CommitLocally(commitStore.selectedProjectPath, message)

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
  } catch (e: unknown) {
    let errMessage = '提交失败'
    if (e instanceof Error) {
      errMessage = e.message
    } else if (typeof e === 'string') {
      errMessage = e
    } else {
      errMessage = JSON.stringify(e)
    }
    console.error('提交失败详细错误:', e)
    showToast('error', '提交失败: ' + errMessage)
    canPush.value = false
  }
}
```

**Step 4: 添加 handlePush 函数**

在 `handleRegenerate` 函数之后（约第 453 行后）添加：

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
    let errMessage = '推送失败'
    if (e instanceof Error) {
      errMessage = e.message
    } else if (typeof e === 'string') {
      errMessage = e
    } else {
      errMessage = JSON.stringify(e)
    }
    console.error('推送失败详细错误:', e)
    showToast('error', '推送失败: ' + errMessage)
  } finally {
    isPushing.value = false
  }
}
```

**Step 5: 添加状态重置**

在项目切换 watch 中重置 `canPush`（约第 318-328 行）:

```typescript
watch(() => projectStore.selectedProject, async (project) => {
  if (project) {
    // 立即清除上一次的生成结果，避免项目切换时显示错误的内容
    commitStore.clearMessage()
    canPush.value = false  // 重置推送按钮状态
    await commitStore.loadProjectAIConfig(project.id)
    await commitStore.loadProjectStatus(project.path)
    await loadHistoryForProject()
    // 加载 Pushover Hook 状态
    await pushoverStore.getProjectHookStatus(project.path)
  }
}, { immediate: true })
```

在 `handleRefresh` 函数中也重置（约第 526-536 行）:

```typescript
async function handleRefresh() {
  if (!currentProjectPath.value) return

  try {
    await commitStore.loadProjectStatus(currentProjectPath.value)
    canPush.value = false  // 重置推送按钮状态
    showToast('success', '已刷新')
  } catch (e) {
    const message = e instanceof Error ? e.message : '刷新失败'
    showToast('error', message)
  }
}
```

**Step 6: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "feat(commit): 添加推送状态管理和处理函数

- 添加 canPush 和 isPushing 响应式状态
- 实现 handlePush 函数处理推送操作
- 修改 handleCommit 在成功后启用推送按钮
- 在项目切换和刷新时重置推送状态

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: 端到端测试

**Files:**
- 无修改

**Step 1: 准备测试环境**

确保有一个测试用的 Git 仓库：
- 本地有提交但未推送到远程
- 配置了远程仓库（可以是真实的或本地模拟的）

**Step 2: 测试正常推送流程**

1. 启动应用: `wails dev`
2. 选择一个有未提交更改的项目
3. 生成 commit 消息
4. 点击"提交到本地"
5. 验证"推送到远程"按钮变为可用
6. 点击"推送到远程"
7. 验证显示"推送中..."状态
8. 验证推送成功后显示 Toast 成功通知
9. 验证按钮重新变为禁用

**Step 3: 测试错误场景**

1. 测试无远程仓库的项目：
   - 推送应失败并显示错误信息
   - 按钮保持可用状态

2. 测试网络错误（可拔网线或使用无效 URL）：
   - 推送应失败并显示网络错误信息

3. 测试冲突场景（远程有新提交）：
   - 推送应失败并显示冲突提示

**Step 4: 测试状态重置**

1. 提交成功后切换项目
2. 验证推送按钮变为禁用
3. 切回原项目
4. 验证推送按钮仍为禁用

**Step 5: 提交测试结果**

如果没有问题，创建文档记录测试通过：

```bash
echo "✅ 推送到远程功能测试通过" >> tmp/test-results.txt
git add tmp/test-results.txt
git commit -m "test: 推送到远程功能端到端测试通过

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: 文档更新

**Files:**
- Modify: `CLAUDE.md`

**Step 1: 更新项目文档**

在 `CLAUDE.md` 的"常用命令"部分之后，添加新功能的说明：

```markdown
## 功能特性

### Commit 生成和提交
1. **生成 Commit 消息**: AI 根据暂存区更改生成 commit 消息
2. **提交到本地**: 将生成的消息提交到本地 Git 仓库
3. **推送到远程**: 在本地提交后，可一键推送到远程仓库

### 推送功能说明
- 推送按钮只在本地提交成功后可用
- 推送到当前分支的同名远程分支
- 推送成功后自动禁用，避免重复推送
- 切换项目或刷新状态时重置推送按钮
```

**Step 2: 提交文档更新**

```bash
git add CLAUDE.md
git commit -m "docs: 更新项目文档说明推送功能

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 验收标准

- [ ] 后端 `PushToRemote` 函数实现并测试通过
- [ ] 前端推送按钮正确显示和响应状态
- [ ] 本地提交成功后推送按钮可用
- [ ] 推送成功后显示成功通知并禁用按钮
- [ ] 推送失败后显示错误信息并保持按钮可用
- [ ] 项目切换时正确重置推送状态
- [ ] 刷新状态时正确重置推送状态
- [ ] 端到端测试通过所有场景

## 注意事项

1. **测试环境**: 建议使用测试仓库进行推送测试，避免影响实际项目
2. **错误处理**: 所有错误都应该以用户友好的方式显示
3. **状态同步**: 确保 canPush 状态与实际操作流程同步
4. **Wails 绑定**: 修改后端 API 后必须重新生成前端绑定
