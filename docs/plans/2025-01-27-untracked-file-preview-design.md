# 未跟踪文件预览功能设计

**日期**: 2025-01-27
**状态**: 设计完成

## 需求概述

当用户点击未跟踪文件（untracked files）时，希望能够预览文件内容。目前未暂存的文件（修改、删除等）点击后会显示 diff 预览，但未跟踪文件点击时没有响应。

## 设计目标

- 复用现有的 `DiffViewer` 组件
- 最小化代码修改
- 保持用户体验一致性
- 正确处理二进制文件

## 架构设计

### 数据流

```
用户点击未跟踪文件
    ↓
handleUntrackedFileClick()
    ↓
commitStore.loadUntrackedFileContent(filePath)
    ↓
后端: GetUntrackedFileContent(projectPath, filePath)
    ↓
返回: { content: string, isBinary: bool }
    ↓
设置 selectedFileDiff = { filePath, diff: content }
    ↓
DiffViewer 显示: oldCode="" (空), newCode=文件内容
```

## 实现方案

### 后端实现

#### 1. 新增文件：`pkg/git/filecontent.go`

创建新文件来读取未跟踪文件内容：

```go
package git

import (
    "fmt"
    "os"
    "path/filepath"
)

// FileContentResult 文件内容读取结果
type FileContentResult struct {
    Content  string // 文件内容（文本文件）
    IsBinary bool   // 是否为二进制文件
}

// ReadFileContent 读取文件内容并判断是否为二进制
func ReadFileContent(repoPath, filePath string) (FileContentResult, error) {
    fullPath := filepath.Join(repoPath, filePath)

    // 读取文件内容
    data, err := os.ReadFile(fullPath)
    if err != nil {
        return FileContentResult{}, fmt.Errorf("读取文件失败: %w", err)
    }

    // 判断是否为二进制文件（复用现有的 isBinary 函数）
    if isBinary(data) {
        return FileContentResult{
            Content:  "",
            IsBinary: true,
        }, nil
    }

    return FileContentResult{
        Content:  string(data),
        IsBinary: false,
    }, nil
}
```

#### 2. 修改：`app.go`

在 App 结构体中添加导出方法：

```go
// GetUntrackedFileContent 获取未跟踪文件内容
func (a *App) GetUntrackedFileContent(projectPath, filePath string) (git.FileContentResult, error) {
    if a.initError != nil {
        return git.FileContentResult{}, a.initError
    }
    return git.ReadFileContent(projectPath, filePath)
}
```

### 前端实现

#### 1. 修改：`frontend/src/stores/commitStore.ts`

**添加导入：**
```typescript
import {
  // ... 现有导入
  GetUntrackedFileContent
} from '../../wailsjs/go/main/App'
```

**添加加载方法：**
```typescript
async function loadUntrackedFileContent(filePath: string) {
  if (!selectedProjectPath.value) {
    error.value = '请先选择项目'
    return
  }

  try {
    const result = await GetUntrackedFileContent(selectedProjectPath.value, filePath)

    if (result.isBinary) {
      // 二进制文件：显示占位提示
      selectedFileDiff.value = {
        filePath,
        diff: '[二进制文件，无法预览内容]'
      }
    } else {
      // 文本文件：设置 diff
      selectedFileDiff.value = {
        filePath,
        diff: result.content
      }
    }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '读取文件内容失败'
    error.value = message
    selectedFileDiff.value = null
  }
}
```

**导出新方法：**
```typescript
return {
  // ... 现有导出
  loadUntrackedFileContent
}
```

#### 2. 修改：`frontend/src/components/UnstagedList.vue`

修改 `handleUntrackedFileClick` 函数（第 329-332 行）：

```typescript
function handleUntrackedFileClick(file: UntrackedFile) {
  commitStore.selectFile(file as StagedFile)
  commitStore.loadUntrackedFileContent(file.path)
}
```

#### 3. 检查：`frontend/src/components/DiffViewer.vue`

确认 `getOldCode` 和 `getNewCode` 方法能正确处理：
- 未跟踪文件的 `oldCode` 返回空字符串
- 未跟踪文件的 `newCode` 返回文件完整内容

## 边界情况处理

### 1. 文件读取失败
- **场景**: 文件不存在、权限问题
- **处理**: 显示错误提示，清空 diff 预览

### 2. 大文件处理
- **场景**: 文件超过 1MB
- **处理**: 显示"文件过大，无法预览"提示（可选优化）

### 3. 特殊文件类型
- **二进制文件**: 显示占位提示 `[二进制文件，无法预览内容]`
- **空文件**: 正常显示空内容
- **符号链接**: 当前实现会读取链接目标内容

### 4. 类型兼容性
- `UntrackedFile` 转换为 `StagedFile` 类型以复用 `selectFile`
- 确保 DiffViewer 能正确处理空 oldCode 的情况

## 测试计划

### 测试用例

1. **文本文件预览**
   - 点击文本类型的未跟踪文件
   - 预期：显示文件完整内容

2. **二进制文件处理**
   - 点击图片、exe 等二进制文件
   - 预期：显示 `[二进制文件，无法预览内容]`

3. **空文件**
   - 点击空文件
   - 预期：正常显示（无报错）

4. **大文件**（如果实现限制）
   - 点击超过大小限制的文件
   - 预期：显示"文件过大，无法预览"

5. **错误场景**
   - 文件在点击后被删除
   - 预期：显示错误提示

6. **项目切换**
   - 在不同项目间切换
   - 预期：状态正确清理，不显示错误内容

## 相关文件

### 后端
- `pkg/git/filecontent.go`（新建）
- `app.go`（修改）
- `pkg/git/git.go`（复用 `isBinary` 函数）

### 前端
- `frontend/src/stores/commitStore.ts`（修改）
- `frontend/src/components/UnstagedList.vue`（修改）
- `frontend/src/components/DiffViewer.vue`（可能需要调整）

## 后续优化

- [ ] 添加文件大小限制
- [ ] 支持大文件的分页预览
- [ ] 添加文件编码检测
- [ ] 支持图片缩略图预览
