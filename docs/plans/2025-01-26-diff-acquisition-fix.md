# 修复 Diff 获取逻辑以匹配 ai-commit 项目

## 问题分析

### 核心问题
对于相同的暂存区变更，`ai-commit-hub` 生成的 commit 消息质量比 `ai-commit` 差。

### 根本原因
两个项目在获取 Git diff 时使用了不同的方法：

| 项目 | 使用方法 | 实现方式 | 读取内容 |
|------|----------|----------|----------|
| **ai-commit** | `GetStagedDiff()` | `git diff --cached` | **暂存区** (staged changes) |
| **ai-commit-hub** | `GetGitDiffIgnoringMoves()` | go-git 库 + 文件系统读取 | **工作树** (working tree) |

### 为什么 GetGitDiffIgnoringMoves 有问题

该函数代码注释已经明确说明了限制（`git.go:38-40`）：

```go
// NOTE: New content is read from the working tree, not the index. This is a known limitation
// if the user stages partial changes and then edits further. To make it *exactly* reflect the
// index, you'd need to read blobs from the index (or shell-out to `git show :path`).
```

**问题场景**：
1. 用户修改文件 `A.txt`
2. 用户 `git add A.txt`（暂存）
3. 用户继续修改 `A.txt`（工作区有新的未暂存变更）
4. 调用 `GetGitDiffIgnoringMoves()` → 读取的是**工作区的最新内容**，而不是**暂存区的内容**

**正确行为**：
- 使用 `GetStagedDiff()` → 通过 `git diff --cached` → 始终读取**暂存区的内容**

## 改进计划

### 阶段 1：修复核心 Diff 获取逻辑

**文件**: `pkg/service/commit_service.go`

**当前代码** (第 122 行):
```go
diff, err := git.GetGitDiffIgnoringMoves(context.Background())
```

**修改为**:
```go
diff, err := git.GetStagedDiff(context.Background())
```

**理由**:
- 与 ai-commit 项目保持一致（`ai-commit/cmd/ai-commit/ai-commit.go:267-272`）
- 确保读取的是暂存区内容，而非工作树内容
- 符合 Git commit 的语义（commit 暂存区的内容）

---

### 阶段 2：代码审查 - 确保无遗漏

检查其他可能使用 `GetGitDiffIgnoringMoves` 的地方，确保都应该使用 `GetStagedDiff`：

**需要审查的调用位置**:
1. ✅ `commit_service.go:122` - **主要问题**
2. ⚠️ `pkg/git/git.go:559` - `partialCommit` 函数内部使用
   - 这个函数用于交互式 splitter，需要确认是否也应该改

**参考 ai-commit 项目的使用模式**:
- `ui.go:865`: 使用 `GetStagedDiff()`
- `ai-commit.go:267`: 使用 `GetStagedDiff()`
- `ai-commit.go:382`: 使用 `GetStagedDiff()` (code review)
- `splitter.go:181`: 使用 `GetGitDiffIgnoringMoves()` - 但这是用于 partial commit，场景特殊

---

### 阶段 3：保留 GetGitDiffIgnoringMoves 的正确使用场景

`GetGitDiffIgnoringMoves` 仍然有存在的价值，用于：
1. 交互式 commit splitter（需要读取工作树来生成 patch）
2. 需要 diff 清理功能（移除 moved blocks、注释-only 变更等）

**不应该删除** `GetGitDiffIgnoringMoves`，而是：
- **Commit 生成流程**：使用 `GetStagedDiff()`
- **交互式 Splitter**：继续使用 `GetGitDiffIgnoringMoves()`（这个场景下读取工作树是正确的）

---

### 阶段 4：验证修复

**测试场景**:
1. 修改一个文件
2. `git add` 暂存
3. 继续修改同一文件（不暂存）
4. 生成 commit 消息
5. 验证：生成的 commit 消息应该基于**暂存区的内容**，而非工作树的最新内容

**对比测试**:
```bash
# 在测试仓库中
echo "v1" > test.txt
git add test.txt
echo "v2" > test.txt  # 工作区修改，不暂存

# ai-commit 应该基于 "v1" 生成
# ai-commit-hub 修复前可能基于 "v2" 生成（错误）
# ai-commit-hub 修复后应该基于 "v1" 生成（正确）
```

---

## 实施步骤

### Step 1: 修改 commit_service.go
将第 122 行的 `GetGitDiffIgnoringMoves` 改为 `GetStagedDiff`

### Step 2: 添加日志
在获取 diff 后添加日志，确认使用的是正确的方法：
```go
logger.Info("使用 GetStagedDiff 获取暂存区变更")
```

### Step 3: 更新测试
确保 `pkg/service/commit_service_test.go` 中的测试也使用正确的方法

### Step 4: 手动测试
1. 使用上述测试场景验证
2. 对比 ai-commit 和 ai-commit-hub 的输出
3. 确认两者生成的 commit 消息一致

---

## 预期结果

修复后，`ai-commit-hub` 的行为将与 `ai-commit` 完全一致：

| 场景 | ai-commit | ai-commit-hub (修复前) | ai-commit-hub (修复后) |
|------|-----------|----------------------|----------------------|
| 暂存后不编辑 | ✅ 正确 | ✅ 正确 | ✅ 正确 |
| 暂存后继续编辑 | ✅ 正确（读暂存区） | ❌ 错误（读工作树） | ✅ 正确（读暂存区） |

---

## 风险评估

**低风险**:
- 修改仅影响 diff 获取方式
- `GetStagedDiff()` 已经存在于代码库中
- 逻辑与成熟的 `ai-commit` 项目一致

**需要注意**:
- 确保 `GetStagedDiff()` 错误处理正确
- 确保暂存区为空时返回空字符串（原有逻辑）
- 确保 context 超时设置合理

---

## 参考资料

- ai-commit 项目: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit`
- GetStagedDiff 实现: `pkg/git/git.go:593-601`
- GetGitDiffIgnoringMoves 注释: `pkg/git/git.go:38-40`
