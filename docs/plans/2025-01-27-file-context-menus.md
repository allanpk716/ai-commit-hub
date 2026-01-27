# 通用文件右键菜单设计

## 概述

为未暂存、已暂存、未跟踪三种文件列表统一实现右键菜单功能，创建可复用的通用组件。

## 目标

1. 统一三种文件列表的右键菜单体验
2. 支持未暂存文件的"还原文件更改"功能
3. 提供配置化的菜单项显示
4. 确保 Windows 平台无控制台弹窗

## 架构设计

### 组件结构

```
FileContextMenu.vue (通用组件)
├── 通过 menuItems prop 控制显示项
├── 使用 Teleport 渲染到 body
└── 通过 emit 通知父组件操作

使用方:
├── UnstagedList.vue → 未暂存文件
├── StagedList.vue → 已暂存文件
└── UnstagedList.vue (未跟踪部分)
```

### 菜单配置

| 文件类型 | 菜单项 |
|---------|--------|
| 未跟踪 | 复制路径、暂存、排除、打开 |
| 未暂存 | 复制路径、暂存、还原、打开 |
| 已暂存 | 复制路径、取消暂存、打开 |

## 组件实现

### FileContextMenu.vue

**Props**
```typescript
interface Props {
  visible: boolean
  x: number
  y: number
  menuItems: MenuItemType[]
}

type MenuItemType =
  | 'copy-path'
  | 'stage'
  | 'unstage'
  | 'discard'
  | 'exclude'
  | 'open-explorer'
```

**Emits**
```typescript
const emit = defineEmits<{
  (e: 'copy-path'): void
  (e: 'stage'): void
  (e: 'unstage'): void
  (e: 'discard'): void
  (e: 'exclude'): void
  (e: 'open-explorer'): void
  (e: 'close'): void
}>()
```

### DiscardConfirmDialog.vue

还原文件确认对话框，防止误操作丢失数据。

**Props**
```typescript
interface Props {
  visible: boolean
  fileName: string
}
```

**交互流程**
1. 点击"还原文件更改"
2. 关闭菜单，显示确认对话框
3. 用户确认后执行 `git checkout -- file`
4. 刷新文件列表

## 后端实现

### 新增 API

`app.go` 中添加 `DiscardFileChanges` 方法：

```go
func (a *App) DiscardFileChanges(projectPath string, filePath string) error {
    gitCmd := exec.Command("git", "checkout", "--", filePath)
    gitCmd.Dir = projectPath

    // Windows: 避免控制台弹窗
    if runtime.GOOS == "windows" {
        gitCmd.SysProcAttr = &syscall.SysProcAttr{
            HideWindow:    true,
            CreationFlags: 0x08000000, // CREATE_NO_WINDOW
        }
    }

    output, err := gitCmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("还原文件失败: %s\n%s", err.Error(), string(output))
    }

    return nil
}
```

### 避免控制台弹窗

所有 Git 命令必须使用 `CREATE_NO_WINDOW` 标志：

```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    HideWindow:    true,
    CreationFlags: 0x08000000,
}
```

## 前端集成

### commitStore.ts

添加还原文件方法：

```typescript
async discardFileChanges(filePath: string): Promise<void> {
  if (!this.selectedProjectPath) {
    throw new Error('未选择项目')
  }

  try {
    await DiscardFileChanges(this.selectedProjectPath, filePath)
    await this.loadProjectStatus(this.selectedProjectPath)
    await this.loadStagingStatus(this.selectedProjectPath)
  } catch (error) {
    const message = error instanceof Error ? error.message : '还原文件失败'
    throw new Error(message)
  }
}
```

### 组件集成示例

**UnstagedList.vue (未暂存部分)**
```vue
<FileContextMenu
  :visible="contextMenuVisible"
  :x="contextMenuX"
  :y="contextMenuY"
  :menuItems="['copy-path', 'stage', 'discard', 'open-explorer']"
  @copy-path="handleCopyPath"
  @stage="handleStage"
  @discard="handleDiscard"
  @open-explorer="handleOpenExplorer"
  @close="closeContextMenu"
/>

<DiscardConfirmDialog
  :visible="discardDialogVisible"
  :fileName="selectedFile?.path || ''"
  @confirm="handleDiscardConfirm"
  @cancel="discardDialogVisible = false"
/>
```

## 文件清单

| 文件 | 操作 | 说明 |
|-----|------|-----|
| `FileContextMenu.vue` | 新建 | 通用右键菜单组件 |
| `DiscardConfirmDialog.vue` | 新建 | 还原确认对话框 |
| `StagedList.vue` | 修改 | 集成右键菜单 |
| `UnstagedList.vue` | 修改 | 集成右键菜单 |
| `ContextMenu.vue` | 删除 | 被通用组件替代 |
| `ExcludeDialog.vue` | 保留 | 继续用于未跟踪文件 |
| `app.go` | 修改 | 添加 DiscardFileChanges |

## 实施顺序

1. 后端：添加 DiscardFileChanges API（确保无控制台弹窗）
2. 创建 FileContextMenu.vue 通用组件
3. 创建 DiscardConfirmDialog.vue 确认对话框
4. 修改 StagedList.vue 集成右键菜单
5. 修改 UnstagedList.vue 集成右键菜单
6. 删除旧的 ContextMenu.vue
7. 测试三种文件类型的所有菜单功能

## 测试要点

- 未暂存文件：复制、暂存、还原、打开
- 已暂存文件：复制、取消暂存、打开
- 未跟踪文件：复制、暂存、排除、打开
- 还原操作确认对话框正常工作
- 所有操作后列表正确刷新
- Windows 平台无控制台弹窗
