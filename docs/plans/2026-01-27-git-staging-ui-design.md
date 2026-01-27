# Git 暂存管理 UI 设计方案

**日期**: 2026-01-27
**状态**: 设计完成，待实施

---

## 1. 概述

为 AI Commit Hub 添加完整的 Git 暂存管理功能，支持文件级别的暂存/取消暂存操作，并提供直观的 diff 预览。

**核心目标**:
- 文件级别的暂存控制
- 双面板布局（已暂存 / 未暂存）
- 内联 diff 预览
- 批量操作支持

---

## 2. 整体架构

### 2.1 组件结构

```
CommitPanel.vue (重构)
├── ProjectStatusHeader.vue (新增 - 顶部状态栏)
│   ├── 分支信息
│   ├── 操作按钮组 (文件夹/终端/刷新)
│   └── Pushover 状态条
│
├── StagingArea.vue (新增 - 暂存管理区)
│   ├── StagedList.vue (已暂存文件列表)
│   └── UnstagedList.vue (未暂存文件列表)
│
├── DiffViewer.vue (新增 - Diff 预览面板)
│   └── @git-diff-view/vue 组件封装
│
└── CommitActions.vue (重构 - 提交操作区)
    ├── AI 配置
    ├── 生成按钮
    ├── 结果显示
    └── 提交/推送按钮
```

### 2.2 布局结构

```
┌─────────────────────────────────────────────────┐
│  分支: main  📁  _>_  🔄  Pushover状态          │  <- ProjectStatusHeader
├─────────────────────┬───────────────────────────┤
│  已暂存 (3)         │   Diff 预览                │  <- StagingArea + DiffViewer
│  ✓ src/app.go  [-]  │   ┌─────────────────────┐ │
│  ✓ main.go     [-]  │   │  + old line         │ │
│  ✓ utils.go    [-]  │   │  - new line         │ │
│                     │   └─────────────────────┘ │
│  未暂存 (5)         │                           │
│  ○ test.go     [+]  │                           │
│  ○ README.md   [+]  │                           │
│  ○ config.yaml [+]  │                           │
├─────────────────────┴───────────────────────────┤
│  🤖 AI 配置 | ⚡ 生成 | ✅ 提交 | ↑ 推送          │  <- CommitActions
└─────────────────────────────────────────────────┘
```

---

## 3. 后端实现

### 3.1 新增文件

**`pkg/git/staging.go`** - 暂存操作
- `GetStagingStatus(repoPath)` → `*StagingStatus` - 获取暂存区状态
- `StageFile(repoPath, filePath)` - 暂存单个文件
- `StageAllFiles(repoPath)` - 暂存所有文件
- `UnstageFile(repoPath, filePath)` - 取消暂存单个文件
- `UnstageAllFiles(repoPath)` - 取消暂存所有文件
- `MarkIgnoredFiles(repoPath, files)` - 标记被 .gitignore 忽略的文件
- `checkFileIgnored(repoPath, filePath)` → `bool` - 检查单个文件是否被忽略

**`pkg/git/diff.go`** - Diff 获取
- `GetFileDiff(repoPath, filePath, staged)` → `string` - 获取文件 diff 内容

### 3.2 数据结构

**`StagingStatus` 结构**（新增）:
```go
type StagingStatus struct {
    Staged   []StagedFile `json:"staged"`
    Unstaged []StagedFile `json:"unstaged"`
}
```

**`StagedFile` 结构**（扩展）:
```go
type StagedFile struct {
    Path    string `json:"path"`
    Status  string `json:"status"`   // Modified, New, Deleted, Renamed
    Ignored bool   `json:"ignored"` // 是否被 .gitignore 忽略（新增）
}
```

### 3.3 Git 忽略文件处理

**检测机制**:
- 使用 `git check-ignore -q <file>` 检测文件是否被忽略
- 返回 0 = 被忽略，返回 1 = 未被忽略

**UI 展示规则**:
- **显示**: 被忽略的文件仍然显示在列表中（不隐藏）
- **标记**: 显示"已忽略"徽章
- **样式**:
  - 半透明背景（opacity: 0.6）
  - 文件路径添加删除线
  - 灰色徽章 "已忽略"
- **交互**:
  - 暂存按钮禁用（disabled）
  - 鼠标悬停提示："此文件被 .gitignore 忽略，需要强制添加"

**暂存行为**:
- 正常文件：`git add <file>`
- 被忽略文件：`git add -f <file>`（强制添加，但在 UI 中禁用）

### 3.4 App 层导出方法

在 `app.go` 中添加导出方法：
```go
func (a *App) GetStagingStatus(projectPath string) (*git.StagingStatus, error)
func (a *App) GetFileDiff(projectPath, filePath string, staged bool) (string, error)
func (a *App) StageFile(projectPath, filePath string) error
func (a *App) StageAllFiles(projectPath string) error
func (a *App) UnstageFile(projectPath, filePath string) error
func (a *App) UnstageAllFiles(projectPath string) error
```

---

## 4. 前端实现

### 4.1 类型定义

```typescript
// types/index.ts
export interface StagedFile {
  path: string     // 文件路径
  status: string   // M/A/D/R 状态
  ignored: boolean  // 是否被 .gitignore 忽略（新增）
}
```

### 4.2 Store 扩展

**`commitStore.ts`** 新增状态和方法：

**状态**:
- `stagedFiles: StagedFile[]` - 已暂存文件
- `unstagedFiles: StagedFile[]` - 未暂存文件
- `selectedStagedFiles: Set<string>` - 已暂存文件选中状态
- `selectedUnstagedFiles: Set<string>` - 未暂存文件选中状态
- `selectedFile: StagedFile | null` - 当前选中查看 diff 的文件
- `fileDiff: string | null` - 当前文件的 diff 内容
- `isLoadingDiff: boolean` - diff 加载状态

**方法**:
- `loadStagingStatus(projectPath)` - 加载暂存区状态（包含 ignored 状态）
- `selectFile(file)` - 选择文件并加载 diff
- `stageFile(filePath)` - 暂存单个文件（检查 ignored 状态）
- `unstageFile(filePath)` - 取消暂存单个文件
- `stageAllFiles()` - 暂存所有文件（跳过 ignored 文件）
- `unstageAllFiles()` - 取消暂存所有文件
- `stageSelectedFiles()` - 暂存选中的文件（跳过 ignored 文件）
- `unstageSelectedFiles()` - 取消暂存选中的文件

### 4.3 组件说明

**`StagingArea.vue`**
- 横向布局容器
- 左侧：双列表（已暂存 + 未暂存）
- 右侧：Diff 预览

**`StagedList.vue` / `UnstagedList.vue`**
- 批量操作栏（横向）:
  - 全选复选框
  - 暂存/取消暂存所选
  - 暂存/取消暂存所有
- 文件列表:
  - 多选复选框
  - 文件状态图标
  - 文件路径
  - 已忽略徽章（`ignored: true` 时显示）
  - 单文件操作按钮 (+/-)
  - 已忽略文件的按钮禁用状态

**已忽略文件 UI 样式**:
- 整体透明度: `opacity: 0.6`
- 背景色: `#2a2a2a`（更暗的灰色）
- 文件路径:
  - 颜色: `#888`（灰色）
  - 删除线: `text-decoration: line-through`
- 徽章样式:
  - 背景色: `#666`
  - 文字颜色: `#aaa`
  - 内边距: `2px 6px`
  - 圆角: `4px`
  - 字号: `9px`
  - 内容: "已忽略"
- 交互:
  - 暂存按钮: `disabled`（禁用）
  - 鼠标悬停提示: "此文件被 .gitignore 忽略，需要强制添加"

**`DiffViewer.vue`**
- 使用 `v-code-diff` 的 `CodeDiff` 组件
- 显示文件信息和状态
- 支持语法高亮
- 暗色主题
- 已暂存文件: 显示 `--cached` diff
- 未暂存文件: 显示工作区 diff

---

## 5. 交互设计

### 5.1 单文件操作

| 操作 | 按钮 | 位置 |
|------|------|------|
| 暂存单个文件 | `+` (绿色圆形) | 未暂存文件行尾 |
| 取消暂存单个文件 | `-` (红色圆形) | 已暂存文件行尾 |

### 5.2 批量操作

**已暂存列表**:
- 全选复选框
- `[-] 取消选定` - 取消暂存选中的文件
- `[═] 取消所有` - 取消暂存所有文件

**未暂存列表**:
- 全选复选框
- `[+] 暂存所选` - 暂存选中的文件（跳过被忽略的文件）
- `[║] 暂存所有` - 暂存所有未忽略文件

**已忽略文件处理**:
- 不参与批量暂存操作
- 单文件暂存按钮禁用
- 可选中但操作时自动跳过

### 5.3 Diff 预览

- 点击文件列表中的文件 → 右侧 Diff 面板显示该文件的差异
- 支持已暂存和未暂存文件的 diff 预览
- 加载状态提示

---

## 6. 第三方依赖

### 6.1 前端库

```bash
npm install v-code-diff
```

**库地址**: https://github.com/Shimada666/v-code-diff

**已测试**: ✅ 渲染正常

**特性**:
- 支持 Vue 3
- API 简单，只需传入旧/新代码字符串
- 语法高亮
- 支持逐行(line-by-line)和并排(side-by-side)两种模式
- 可配置上下文行数

**API 示例**:
```vue
<script setup>
import { CodeDiff } from 'v-code-diff'
</script>

<template>
  <CodeDiff
  :old-string="oldCode"
  :new-string="newCode"
  :output-format="'line-by-line'"
  :context="10"
/>
```

---

## 7. 实施步骤

### Phase 1: 后端实现
1. 创建 `pkg/git/staging.go`
2. 创建 `pkg/git/diff.go`
3. 在 `app.go` 中添加导出方法
4. 运行 `wails generate` 生成绑定

### Phase 2: 前端基础
1. 安装 `v-code-diff`
2. 更新 `types/index.ts`（添加 `ignored` 字段）
3. 扩展 `commitStore.ts`

### Phase 3: 组件开发
1. 创建 `ProjectStatusHeader.vue`
2. 创建 `StagedList.vue`
3. 创建 `UnstagedList.vue`
4. 创建 `DiffViewer.vue`
5. 创建 `StagingArea.vue`
6. 重构 `CommitActions.vue`

### Phase 4: 集成
1. 重构 `CommitPanel.vue`
2. 测试完整流程

---

## 8. 测试要点

- [ ] 暂存单个文件
- [ ] 取消暂存单个文件
- [ ] 暂存所有文件
- [ ] 取消所有暂存
- [ ] 多选文件批量操作
- [ ] Diff 预览正确显示
- [ ] 已暂存文件 diff
- [ ] 未暂存文件 diff
- [ ] 提交后状态刷新
- [ ] 切换项目状态更新
- [ ] 空状态显示
- [ ] 错误处理和提示
- [ ] **已忽略文件检测**
- [ ] **已忽略文件样式显示**（半透明、删除线、徽章）
- [ ] **已忽略文件暂存按钮禁用**
- [ ] **批量暂存时跳过已忽略文件**

---

## 9. 参考资料

- **v-code-diff**: https://github.com/Shimada666/v-code-diff
- **Wails 文档**: https://wails.io/docs/next/introduction
