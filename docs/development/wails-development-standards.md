# Wails 开发规范

本文档详细说明了 AI Commit Hub 项目中 Wails 框架的开发规范和最佳实践。

## 调试流程

### 启动开发服务器

```bash
# 启动开发服务器（支持前端热更新）
wails dev
```

### 测试前后端通信

使用浏览器技能（dev-browser）进行前后端通信和交互测试：

1. 启动 `wails dev` 后，应用会自动打开浏览器窗口
2. 使用浏览器开发者工具查看网络请求和控制台输出
3. 测试前后端 API 调用和事件监听

### 代码修改后

- **Go 代码修改**：需要重启 `wails dev` 服务器
- **前端代码修改**：支持热更新，自动刷新浏览器

## API 方法命名

### 导出方法规范

在 `app.go` 中定义的导出方法会自动生成驼峰命名的 JavaScript 绑定：

```go
// Go 代码（app.go）
func (a *App) AddProject(path string) error {
    // ...
}

// 自动生成的 JavaScript 绑定
// window.goMain.App.AddProject(path)
```

### 命名规则

- Go 导出方法使用大写开头（Go 导出规则）
- JavaScript 绑定会自动转换为驼峰命名
- 方法名应该清晰表达其功能

```go
// ✅ 好的方法名
func (a *App) GetProjects() ([]models.GitProject, error)
func (a *App) GenerateCommitMessage(projectID uint) error

// ❌ 不好的方法名
func (a *App) Do() error
func (a *App) Process(id uint) error
```

## 错误处理

### 初始化错误检查

所有 API 方法应检查 `a.initError`，如果数据库初始化失败应返回错误：

```go
func (a *App) GetProjects() ([]models.GitProject, error) {
    // 检查初始化错误
    if a.initError != nil {
        return nil, fmt.Errorf("app not initialized: %w", a.initError)
    }

    // 业务逻辑
    projects, err := a.projectRepo.GetAll()
    if err != nil {
        return nil, err
    }

    return projects, nil
}
```

### 错误传递

使用 `fmt.Errorf` 和 `%w` 包装错误，保留错误链：

```go
if err != nil {
    return nil, fmt.Errorf("failed to add project: %w", err)
}
```

### 日志记录

使用 `github.com/WQGroup/logger` 记录错误：

```go
import "github.com/WQGroup/logger"

if err != nil {
    logger.Errorf("Failed to add project %s: %v", path, err)
    return err
}
```

详见：`docs/development/logging-standards.md`

## Wails Events

### 流式输出实现

使用 Wails Events 实现流式输出（例如 AI 生成的 commit 消息）：

#### 后端发送事件

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// 发送事件到前端
func (a *App) GenerateCommitMessage(projectID uint) error {
    // 流式输出 AI 生成的内容
    for chunk := range aiStream {
        runtime.EventsEmit(a.ctx, "commit-stream", chunk)
    }

    // 发送完成事件
    runtime.EventsEmit(a.ctx, "commit-complete", nil)

    return nil
}
```

#### 前端监听事件

```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime'

// 监听流式事件
EventsOn("commit-stream", (chunk: string) => {
  commitStore.appendMessage(chunk)
})

// 监听完成事件
EventsOn("commit-complete", () => {
  commitStore.setGenerating(false)
})
```

### 事件命名规范

- 使用 kebab-case 命名事件（例如：`commit-stream`）
- 事件名应该清晰表达其用途
- 相关事件使用前缀分组（例如：`commit-*`）

## 控制台窗口隐藏（Windows）

### 问题描述

在 Windows 上执行外部命令（如 git）时，会出现控制台窗口闪烁，影响用户体验。

### 解决方案

使用自定义 `Command` 函数封装 `exec.Cmd`，在 Windows 平台下设置 `CREATE_NO_WINDOW` 标志。

### 代码实现

```go
// app.go:36-49
import (
    "os/exec"
    stdruntime "runtime"
    "golang.org/x/sys/windows"
)

// Command creates a new exec.Cmd with hidden window on Windows
// This prevents console windows from popping up when running external commands
func Command(name string, args ...string) *exec.Cmd {
    cmd := exec.Command(name, args...)

    // On Windows, hide the console window to prevent popups
    if stdruntime.GOOS == "windows" {
        cmd.SysProcAttr = &windows.SysProcAttr{
            CreationFlags: 0x08000000, // CREATE_NO_WINDOW
        }
    }

    return cmd
}
```

### 使用示例

```go
// ✅ 正确：使用自定义 Command 函数
cmd := Command("git", "status", "--porcelain")
cmd.Dir = projectPath
output, err := cmd.CombinedOutput()

// ❌ 错误：直接使用 exec.Command
cmd := exec.Command("git", "status", "--porcelain")  // 会导致控制台窗口闪烁
```

### 注意事项

- **所有外部命令**（git、python 等）都必须使用 `Command` 函数
- 不能直接使用 `exec.Command`，否则会导致控制台窗口闪烁
- Unix/Linux 平台不受影响，`CREATE_NO_WINDOW` 标志仅在 Windows 下生效

详细说明：`docs/lessons-learned/windows-console-hidden-fix.md`

## 前后端类型同步

### Go 结构体定义

```go
// pkg/models/gitproject.go
type GitProject struct {
    ID        uint      `gorm:"primarykey"`
    Name      string    `json:"name"`
    Path      string    `json:"path"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### TypeScript 类型定义

Go 结构体修改后需要重新生成 Wails 绑定。确保 `frontend/src/types/index.ts` 中的 TypeScript 类型与 Go 结构体保持一致：

```typescript
// frontend/src/types/index.ts
export interface GitProject {
  id: number
  name: string
  path: string
  created_at: string
  updated_at: string
}
```

### 类型同步流程

1. 修改 Go 结构体
2. 运行 `wails dev` 或 `wails build` 重新生成绑定
3. 检查 `frontend/wailsjs/go/models/Go.js` 中的类型定义
4. 更新 `frontend/src/types/index.ts` 中的 TypeScript 类型

## 常见问题

### Wails 绑定生成错误

**错误信息**：`wailsbindings.exe: %1 is not a valid Win32 application`

**解决方案**：

1. 删除临时目录下的 wbindings 文件
   ```bash
   # Windows
   rm -rf %TEMP%\wbindings
   ```

2. 重新运行 `wails dev`

3. 或使用已有的绑定文件，直接 `go build`：
   ```bash
   go build -o build/bin/ai-commit-hub.exe .
   ```

### 前端无法调用后端方法

**检查清单**：

1. Go 方法是否导出（大写开头）
2. 方法是否在 `App` 结构体上定义
3. 是否重新生成了 Wails 绑定
4. 浏览器控制台是否有错误信息

### Events 无法接收

**检查清单**：

1. 后端是否使用 `runtime.EventsEmit` 发送事件
2. 前端是否使用 `EventsOn` 监听事件
3. 事件名称是否一致（区分大小写）
4. 是否在组件卸载时取消事件监听

```typescript
// 组件卸载时取消事件监听
import { EventsOff } from '../../wailsjs/runtime/runtime'

onUnmounted(() => {
  EventsOff("commit-stream")
  EventsOff("commit-complete")
})
```

## 相关文档

- 日志输出规范：`docs/development/logging-standards.md`
- Windows 控制台窗口隐藏：`docs/lessons-learned/windows-console-hidden-fix.md`
- CLAUDE.md：项目根目录
