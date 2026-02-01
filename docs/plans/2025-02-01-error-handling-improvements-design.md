# 错误处理和状态更新改进设计

**日期**: 2025-02-01
**状态**: 设计阶段
**优先级**: 高

## 需求概述

当前应用存在三个主要问题：

1. **错误提示自动消失**：错误提示在 3 秒后自动消失，用户来不及复制错误信息
2. **缺少日志记录**：前端错误没有记录到后端日志，无法追溯历史问题
3. **状态未同步更新**：本地提交成功后，项目列表中的"领先远程 X 个提交"计数未更新

## 设计目标

1. 错误提示常驻显示，除非用户手动关闭
2. 每个错误提供复制按钮，方便复制错误信息
3. 所有错误和警告信息记录到本地日志文件
4. 提交成功后自动刷新项目列表的推送状态

## 整体架构

### 分层设计

```
┌─────────────────────────────────────────┐
│         前端 UI 层         │
│  - ErrorToast 组件（右下角错误堆叠）    │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│      状态管理层 (Pinia Store)            │
│  - errorStore：错误列表管理              │
│  - 自动清理：error 最多 10 个            │
│  - 自动清理：warning 最多 5 个           │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│      后端服务层 (Go Service)             │
│  - ErrorService：日志记录                │
│  - Logger：本地文件日志                  │
└─────────────────────────────────────────┘
```

## 详细设计

### 1. ErrorToast 组件

**文件位置**: `frontend/src/components/ErrorToast.vue`

**功能特性**:
- 错误卡片包含：图标、消息、复制按钮、关闭按钮
- 固定在右下角显示，距离边缘 20px
- 支持多错误堆叠，新错误显示在底部（保持时间顺序）
- 错误卡片最大宽度 400px，自适应内容
- 进入动画：从底部滑入，淡入（0.3s ease-out）
- 离开动画：向右滑出，淡出（0.2s ease-in）

**UI 布局**:
```
┌─────────────────────────────────┐
│ ❌ 错误消息文本                 │
│    [复制] [×]                   │
└─────────────────────────────────┘
```

**交互功能**:
- 点击复制按钮：复制完整错误信息到剪贴板
- 点击关闭按钮：移除该错误
- 支持 Ctrl+C 快捷键复制选中的错误

### 2. errorStore 数据结构

**文件位置**: `frontend/src/stores/errorStore.ts`

**TypeScript 类型定义**:

```typescript
interface ErrorItem {
  id: string                    // 唯一标识（UUID）
  type: 'error' | 'warning'     // 错误类型
  message: string               // 主要错误消息（简短，1-2行）
  details?: string              // 详细错误信息（堆栈、调试信息）
  timestamp: number             // 时间戳
  source?: string               // 错误来源（如 'CommitPanel'）
}

interface ErrorState {
  errors: ErrorItem[]           // 错误列表
  maxErrors: number             // 最大错误数（默认10）
  maxWarnings: number           // 最大警告数（默认5）
}
```

**核心 API 方法**:

```typescript
// 添加错误（同时显示UI和记录日志）
addError(message: string, details?: string, type?: 'error' | 'warning'): void

// 移除指定错误
removeError(id: string): void

// 复制错误到剪贴板
copyError(id: string): Promise<void>

// 清除所有错误
clearAll(): void

// 发送到后端日志（内部方法）
sendToBackend(error: ErrorItem): Promise<void>
```

**自动管理策略**:
- 警告信息：最多保留 5 个，超过自动移除最早的
- 错误信息：最多保留 10 个，超过自动移除最早的
- 智能管理：根据错误类型独立计数

**使用示例**:

```typescript
import { useErrorStore } from '@/stores/errorStore'

const errorStore = useErrorStore()

// 添加错误（自动显示UI + 记录日志）
errorStore.addError('提交失败', 'Git command failed:...', 'error')

// 添加警告
errorStore.addError('配置未保存', 'Provider配置未找到', 'warning')
```

### 3. 后端日志集成

**文件位置**: `pkg/service/error_service.go`

**数据结构**:

```go
type FrontendError struct {
    Type      string    `json:"type"`      // "error" | "warning"
    Message   string    `json:"message"`   // 简短消息
    Details   string    `json:"details"`   // 详细信息
    Source    string    `json:"source"`    // 来源组件
    Timestamp time.Time `json:"timestamp"` // 时间戳
}

type ErrorService struct {
    logger *logger.Logger
}
```

**API 方法**:

```go
func (s *ErrorService) LogError(err FrontendError) error {
    // 使用 logger 库记录到本地文件
    if err.Type == "error" {
        s.logger.Errorf("[Frontend] %s: %s\nDetails: %s",
            err.Source, err.Message, err.Details)
    } else {
        s.logger.Warnf("[Frontend] %s: %s\nDetails: %s",
            err.Source, err.Message, err.Details)
    }
    return nil
}
```

**App.go 导出方法**:

```go
func (a *App) LogFrontendError(errJSON string) error {
    var fe service.FrontendError
    if err := json.Unmarshal([]byte(errJSON), &fe); err != nil {
        return err
    }
    return a.errorService.LogError(fe)
}
```

**前端 Wails 绑定调用**:

```typescript
// stores/errorStore.ts
import { LogFrontendError } from '../../wailsjs/go/main/App'

async sendToBackend(error: ErrorItem) {
  try {
    const errorJSON = JSON.stringify({
      type: error.type,
      message: error.message,
      details: error.details,
      source: error.source || 'Unknown',
      timestamp: new Date(error.timestamp).toISOString()
    })
    await LogFrontendError(errorJSON)
  } catch (e) {
    console.error('Failed to log error to backend:', e)
  }
}
```

**日志文件位置**:
- **Windows**: `C:\Users\<username>\.ai-commit-hub\logs\errors.log`
- **macOS/Linux**: `~/.ai-commit-hub/logs/errors.log`

### 4. 提交后状态自动更新

**问题分析**:
当前流程在 `CommitPanel.vue:323-371` 中，调用 `CommitLocally()` 后虽然调用了 `statusCache.refresh()`，但只更新了缓存，没有触发项目列表重新渲染。

**解决方案：统一的状态刷新事件流**

#### 4.1 后端：提交成功后发送事件

修改 `App.go` 中的 `CommitLocally` 方法：

```go
func (a *App) CommitLocally(projectPath, message string) error {
    // ... 执行 git commit ...

    // 提交成功后发送全局事件
    runtime.EventsEmit(a.ctx, "project-status-changed", map[string]interface{}{
        "projectPath": projectPath,
        "changeType": "commit",
        "timestamp": time.Now(),
    })

    return nil
}
```

#### 4.2 前端：监听全局事件并刷新

修改 `stores/projectStore.ts`：

```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useProjectStore = defineStore('project', () => {
  // ... 现有代码 ...

  // 监听项目状态变更事件
  EventsOn('project-status-changed', async (data) => {
    const { projectPath, changeType } = data

    // 更新该项目在列表中的显示状态
    const project = projects.value.find(p => p.path === projectPath)
    if (project) {
      // 刷新该项目的状态（包括推送状态、分支信息等）
      const statusCache = useStatusCache()
      await statusCache.refresh(projectPath, { force: true })

      // 触发项目列表重新渲染（通过更新 project 的某个响应式属性）
      project.lastModified = Date.now()  // 添加时间戳触发更新
    }
  })
})
```

#### 4.3 类型定义更新

修改 `types/index.ts`：

```typescript
export interface GitProject {
  // ... 现有字段 ...
  lastModified?: number  // 添加：最后修改时间（用于触发更新）
}
```

#### 4.4 ProjectList 组件响应变化

修改 `components/ProjectList.vue`：

```vue
<template>
  <div v-for="project in sortedProjects" :key="project.id || project.path">
    <!-- 项目卡片 -->
    <ProjectCard
      :project="project"
      :status="getStatus(project.path)"
      :key="project.lastModified"  <!-- 时间戳变化触发重新渲染 -->
    />
  </div>
</template>
```

### 5. 完整的错误处理数据流

**场景：用户执行 Git 提交失败**

```
1. 用户点击"提交"按钮
   ↓
2. CommitPanel.handleCommit() 调用 CommitLocally()
   ↓
3. Git 命令执行失败，返回错误
   ↓
4. 前端 catch 错误
   ↓
5. errorStore.addError(
      '提交失败',
      'Git error: conflicts...',
      'error',
      'CommitPanel'
    )
   ↓
6. errorStore 内部处理：
   a. 生成唯一 ID
   b. 添加到错误列表（触发 UI 更新）
   c. 调用 sendToBackend()
   d. 自动管理数量（超过限制移除旧错误）
   ↓
7. sendToBackend() 调用后端 LogFrontendError()
   ↓
8. 后端 errorService.LogError() 记录到日志文件
   ↓
9. 用户看到：
   - 右下角弹出错误提示（常驻）
   - 日志文件中记录了完整错误
```

### 6. 错误级别管理

| 类型 | 最大数量 | 自动清理 | 日志级别 |
|------|---------|---------|---------|
| error | 10 | ✅ | Errorf() |
| warning | 5 | ✅ | Warnf() |

## 实施计划

### 第 1 步：创建 ErrorStore（前端）
- 创建 `frontend/src/stores/errorStore.ts`
- 实现错误列表管理、自动清理、复制功能
- 编写单元测试

**预计时间**: 2-3 小时

### 第 2 步：创建 ErrorToast 组件（前端）
- 创建 `frontend/src/components/ErrorToast.vue`
- 实现错误卡片 UI、堆叠动画、复制/关闭按钮
- 编写组件测试

**预计时间**: 2-3 小时

### 第 3 步：集成到 App.vue（前端）
- 在 App.vue 中引入 ErrorToast 组件
- 全局样式配置

**预计时间**: 0.5-1 小时

### 第 4 步：创建后端 ErrorService（后端）
- 创建 `pkg/service/error_service.go`
- 实现 LogError 方法

**预计时间**: 1 小时

### 第 5 步：修改现有错误处理（前端）
- 替换 CommitPanel.vue 中的 `showToast` 为 `errorStore.addError`
- 替换其他组件中的错误处理（如果有的话）

**预计时间**: 2-3 小时

### 第 6 步：实现提交后状态更新（前后端）
- 修改 App.go 的 CommitLocally 方法，添加事件发送
- 修改 projectStore.ts，监听事件并刷新状态
- 更新 GitProject 类型定义，添加 lastModified 字段

**预计时间**: 2-3 小时

### 第 7 步：配置日志输出
- 更新 logger 配置，确保错误日志写入文件
- 配置日志轮转和清理策略

**预计时间**: 0.5-1 小时

### 第 8 步：集成测试
- 测试完整流程：错误显示、复制、日志记录、状态更新
- 修复发现的问题

**预计时间**: 2-3 小时

**总预计时间**: 约 12-16 小时

## 测试策略

### 前端单元测试

**ErrorStore 测试** (`stores/__tests__/errorStore.spec.ts`):

```typescript
describe('ErrorStore', () => {
  test('添加错误后自动限制数量', async () => {
    const store = useErrorStore()

    // 添加 15 个错误
    for (let i = 0; i < 15; i++) {
      await store.addError(`Error ${i}`, 'details', 'error')
    }

    // 应该只保留最新的 10 个
    expect(store.errors.length).toBe(10)
    expect(store.errors[0].message).toBe('Error 5')
  })

  test('复制错误到剪贴板', async () => {
    const store = useErrorStore()
    await store.addError('Test error', 'details', 'error')

    const errorId = store.errors[0].id
    await store.copyError(errorId)

    const clipboardText = await navigator.clipboard.readText()
    expect(clipboardText).toContain('Test error')
  })
})
```

**ErrorToast 组件测试** (`components/__tests__/ErrorToast.spec.ts`):

```typescript
describe('ErrorToast', () => {
  test('显示错误列表', () => {
    const wrapper = mount(ErrorToast)
    const store = useErrorStore()
    store.addError('Test', 'details', 'error')

    expect(wrapper.text()).toContain('Test')
  })

  test('点击关闭按钮移除错误', async () => {
    const wrapper = mount(ErrorToast)
    const store = useErrorStore()
    await store.addError('Test', 'details', 'error')

    await wrapper.find('.close-btn').trigger('click')
    expect(store.errors.length).toBe(0)
  })
})
```

### 后端单元测试

**ErrorService 测试** (`pkg/service/error_service_test.go`):

```go
func TestErrorService_LogError(t *testing.T) {
    service := NewErrorService(logger)

    err := FrontendError{
        Type:      "error",
        Message:   "Test error",
        Details:   "Test details",
        Source:    "Test",
        Timestamp: time.Now(),
    }

    // 不应该返回错误
    assert.NoError(t, service.LogError(err))
}
```

### 集成测试场景

**场景 1：提交失败 → 显示错误 + 记录日志**
1. 模拟 Git 提交失败
2. 验证右下角显示错误提示
3. 验证日志文件中有错误记录

**场景 2：多个错误堆叠显示**
1. 连续触发 3 个错误
2. 验证显示 3 个错误卡片
3. 验证新错误在底部

**场景 3：自动清理旧错误**
1. 添加 15 个错误
2. 验证只保留最新 10 个

**场景 4：提交成功 → 自动刷新状态**
1. 执行提交操作
2. 验证项目列表的"领先 X 个提交"数字更新

## 技术栈

- **前端**: Vue 3, TypeScript, Pinia
- **后端**: Go 1.21+, Wails v2
- **日志库**: github.com/WQGroup/logger
- **测试**: Vitest (前端), Go testing (后端)

## 依赖项

- 需要安装 `@vueuse/core`（用于 clipboard API）
- 需要配置 logger 库的文件输出
- Wails Events 系统用于前后端通信

## 风险和注意事项

1. **性能考虑**:
   - 错误列表不要过长，通过自动清理限制数量
   - 日志记录不应阻塞 UI，使用异步处理

2. **用户体验**:
   - 成功提示保持 3 秒自动消失（不改变）
   - 只有错误和警告需要常驻显示
   - 错误堆叠不要遮挡太多屏幕空间

3. **日志文件管理**:
   - 需要配置日志轮转，防止日志文件过大
   - 定期清理旧日志文件
   - 用户应该能够找到日志文件位置

4. **兼容性**:
   - 确保 Clipboard API 在所有支持的平台上工作
   - Wails Events 在不同平台的行为一致性

## 未来改进

1. **错误中心**：提供专门的界面查看历史错误
2. **错误过滤**：支持按类型、来源、时间过滤错误
3. **错误搜索**：在历史错误中搜索关键字
4. **错误导出**：支持导出错误报告
5. **错误统计**：统计错误发生频率，帮助发现常见问题

## 相关文档

- [CLAUDE.md](../../CLAUDE.md) - 项目总体说明
- [启动流程与状态管理](../CLAUDE.md) - StatusCache 机制说明
- [日志库使用说明](https://github.com/WQGroup/logger) - logger 库文档

## 变更历史

| 日期 | 版本 | 变更说明 | 作者 |
|------|------|---------|------|
| 2025-02-01 | 1.0 | 初始设计文档 | Claude Sonnet |
