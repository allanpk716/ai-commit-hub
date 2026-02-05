# Backend API Documentation

## 概述

本文档描述了 AI Commit Hub 后端暴露的所有 API 方法。

## 初始化

### startup(ctx context.Context)

应用启动时调用，初始化数据库和服务。

**参数:**
- `ctx context.Context`: Wails 上下文

**返回:** 无

**事件发送:**
- `startup-complete`: 启动完成事件，包含预加载的项目状态

---

## 项目管理

### GetAllProjects() ([]models.GitProject, error)

获取所有 Git 项目列表。

**返回:**
- `[]models.GitProject`: 项目列表
- `error`: 错误信息

**示例:**
```go
projects, err := app.GetAllProjects()
if err != nil {
    return err
}
```

---

### AddProject(project models.GitProject) (*models.GitProject, error)

添加新项目。

**参数:**
- `project models.GitProject`: 项目信息

**返回:**
- `*models.GitProject`: 创建的项目（包含 ID）
- `error`: 错误信息

**验证:**
- `project.Name`: 必填
- `project.Path`: 必填，必须为有效的 Git 仓库路径

**示例:**
```go
project, err := app.AddProject(models.GitProject{
    Name: "My Project",
    Path: "/path/to/repo",
})
```

---

### UpdateProject(project models.GitProject) error

更新项目信息。

**参数:**
- `project models.GitProject`: 项目信息（包含 ID）

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.UpdateProject(models.GitProject{
    ID:   1,
    Name: "Updated Name",
    Path: "/new/path",
})
```

---

### DeleteProject(id int) error

删除项目。

**参数:**
- `id int`: 项目 ID

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.DeleteProject(1)
```

---

### ReorderProjects(projects []models.GitProject) error

重新排序项目。

**参数:**
- `projects []models.GitProject`: 项目列表（包含新的 sort_order）

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.ReorderProjects([]models.GitProject{
    {ID: 1, SortOrder: 0},
    {ID: 2, SortOrder: 1},
})
```

---

## Commit 生成

### GenerateCommit(projectPath string) error

为指定项目生成 commit 消息（流式输出）。

**参数:**
- `projectPath string`: 项目路径

**返回:**
- `error`: 错误信息

**事件发送:**
- `commit-delta`: 流式输出 commit 消息片段
- `commit-complete`: 生成完成
  - `success`: boolean - 是否成功
  - `error`?: string - 错误信息

**示例:**
```typescript
// 前端监听事件
EventsOn('commit-delta', (delta: string) => {
  commitMessage.value += delta
})

EventsOn('commit-complete', (data) => {
  isGenerating.value = false
  if (!data.success) {
    error.value = data.error
  }
})

// 调用 API
await app.GenerateCommit(projectPath)
```

---

## Git 操作

### StageFile(projectPath string, filePath string) error

暂存文件。

**参数:**
- `projectPath string`: 项目路径
- `filePath string`: 文件路径（相对于项目根目录）

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.StageFile("/path/to/repo", "src/main.go")
```

---

### UnstageFile(projectPath string, filePath string) error

取消暂存文件。

**参数:**
- `projectPath string`: 项目路径
- `filePath string`: 文件路径（相对于项目根目录）

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.UnstageFile("/path/to/repo", "src/main.go")
```

---

### DiscardChanges(projectPath string, filePath string) error

丢弃文件的未暂存更改。

**参数:**
- `projectPath string`: 项目路径
- `filePath string`: 文件路径（相对于项目根目录）

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.DiscardChanges("/path/to/repo", "src/main.go")
```

---

### CommitProject(projectPath string, message string) error

提交更改。

**参数:**
- `projectPath string`: 项目路径
- `message string`: commit 消息

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.CommitProject("/path/to/repo", "feat: add new feature")
```

---

### PushProject(projectPath string) error

推送更改到远程仓库。

**参数:**
- `projectPath string`: 项目路径

**返回:**
- `error`: 错误信息

**错误类型:**
- 无远程仓库
- 认证失败
- 推送冲突
- 网络错误

**示例:**
```go
err := app.PushProject("/path/to/repo")
```

---

## 状态查询

### GetProjectStatus(projectPath string) (*ProjectFullStatus, error)

获取单个项目的完整状态。

**参数:**
- `projectPath string`: 项目路径

**返回:**
- `*ProjectFullStatus`: 项目状态信息
  - `GitStatus`: Git 状态（分支、提交信息等）
  - `StagingStatus`: 暂存区状态
  - `UntrackedCount`: 未跟踪文件数量
  - `PushoverStatus`: Pushover Hook 状态
  - `PushStatus`: 推送状态（ahead、behind、canPush）
  - `LastUpdated`: 最后更新时间

**示例:**
```go
status, err := app.GetProjectStatus("/path/to/repo")
if err != nil {
    return err
}
fmt.Printf("Branch: %s\n", status.GitStatus.Branch)
fmt.Printf("Uncommitted changes: %v\n", status.GitStatus.HasUncommittedChanges)
```

---

### GetAllProjectStatuses(projectPaths []string) (map[string]*ProjectFullStatus, error)

批量获取多个项目的完整状态。

**参数:**
- `projectPaths []string`: 项目路径列表

**返回:**
- `map[string]*ProjectFullStatus`: 项目路径到状态的映射

**性能:**
- 使用并发加载，自动根据项目数量和 CPU 核心数调整并发度
- 超时时间: 30 秒

**示例:**
```go
statuses, err := app.GetAllProjectStatuses([]string{
    "/path/to/repo1",
    "/path/to/repo2",
})
for path, status := range statuses {
    fmt.Printf("%s: %s\n", path, status.GitStatus.Branch)
}
```

---

## Commit 历史

### GetCommitHistory(projectID int, limit int) ([]models.CommitHistory, error)

获取项目的 commit 历史记录。

**参数:**
- `projectID int`: 项目 ID
- `limit int`: 返回记录数量限制（0 表示不限制）

**返回:**
- `[]models.CommitHistory`: commit 历史记录
- `error`: 错误信息

**示例:**
```go
history, err := app.GetCommitHistory(1, 10)
if err != nil {
    return err
}
for _, commit := range history {
    fmt.Printf("%s: %s\n", commit.CreatedAt, commit.Message)
}
```

---

### GetRecentCommits(limit int) ([]models.CommitHistory, error)

获取最近的 commit 记录（跨所有项目）。

**参数:**
- `limit int`: 返回记录数量限制

**返回:**
- `[]models.CommitHistory`: commit 历史记录
- `error`: 错误信息

**示例:**
```go
recent, err := app.GetRecentCommits(20)
if err != nil {
    return err
}
```

---

## Pushover Hook

### GetPushoverStatus(projectPath string) (*pushover.HookStatus, error)

获取项目的 Pushover Hook 状态。

**参数:**
- `projectPath string`: 项目路径

**返回:**
- `*pushover.HookStatus`: Hook 状态信息
  - `Installed`: 是否已安装
  - `IsLatestVersion`: 是否为最新版本
  - `Version`: 当前版本
  - `UpdateAvailable`: 是否有更新可用

**示例:**
```go
status, err := app.GetPushoverStatus("/path/to/repo")
if err != nil {
    return err
}
fmt.Printf("Installed: %v\n", status.Installed)
fmt.Printf("Version: %s\n", status.Version)
```

---

### ReinstallPushoverHook(projectPath string) error

重装 Pushover Hook。

**参数:**
- `projectPath string`: 项目路径

**行为:**
1. 保存当前通知配置
2. 重新安装 Hook
3. 恢复通知配置

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.ReinstallPushoverHook("/path/to/repo")
```

---

### CheckPushoverUpdate() (map[string]bool, error)

检查所有项目的 Pushover Hook 更新。

**返回:**
- `map[string]bool`: 项目路径到是否有更新的映射
- `error`: 错误信息

**示例:**
```go
updates, err := app.CheckPushoverUpdate()
for path, hasUpdate := range updates {
    if hasUpdate {
        fmt.Printf("%s has update available\n", path)
    }
}
```

---

## 配置管理

### GetConfig() (*models.Config, error)

获取 AI Provider 配置。

**返回:**
- `*models.Config`: 配置信息
  - `Provider`: AI Provider 名称
  - `ApiKey`: API Key
  - `Model`: 模型名称
  - `Language`: commit 消息语言
  - `CustomPrompt`: 自定义 Prompt

**示例:**
```go
config, err := app.GetConfig()
if err != nil {
    return err
}
fmt.Printf("Provider: %s\n", config.Provider)
fmt.Printf("Model: %s\n", config.Model)
```

---

### SaveConfig(config models.Config) error

保存 AI Provider 配置。

**参数:**
- `config models.Config`: 配置信息

**验证:**
- `config.Provider`: 必须为支持的 Provider
- `config.ApiKey`: 除了 Ollama 外必填

**返回:**
- `error`: 错误信息

**示例:**
```go
err := app.SaveConfig(models.Config{
    Provider: "openai",
    ApiKey:   "sk-...",
    Model:     "gpt-3.5-turbo",
    Language:  "zh",
})
```

---

### GetAvailableProviders() []string

获取所有可用的 AI Provider 列表。

**返回:**
- `[]string`: Provider 名称列表

**示例:**
```go
providers := app.GetAvailableProviders()
// ["openai", "anthropic", "google", "deepseek", "ollama", "phind"]
```

---

### GetProviderModels(provider string) []string

获取指定 Provider 的可用模型列表。

**参数:**
- `provider string`: Provider 名称

**返回:**
- `[]string`: 模型名称列表

**示例:**
```go
models := app.GetProviderModels("openai")
// ["gpt-3.5-turbo", "gpt-4", "gpt-4-turbo", ...]
```

---

## 应用控制

### Quit()

退出应用。

**行为:**
- 清理所有资源
- 关闭数据库连接
- 停止系统托盘
- 退出应用

**示例:**
```go
app.Quit()
```

---

## 错误处理

所有 API 方法在失败时返回错误。错误类型包括：

### AppInitError

应用初始化错误（数据库连接失败等）。

**处理:**
```go
if err != nil {
    if apperrors.IsInitError(err) {
        // 应用未正确初始化
    }
}
```

### ValidationError

数据验证错误（缺少必填字段等）。

**处理:**
```go
if err != nil {
    if apperrors.IsValidationError(err) {
        // 数据验证失败
    }
}
```

### GitOperationError

Git 操作失败。

**处理:**
```go
if err != nil {
    if apperrors.IsGitError(err) {
        // Git 操作失败
    }
}
```

### AIProviderError

AI Provider 调用失败。

**处理:**
```go
if err != nil {
    if apperrors.IsAIProviderError(err) {
        // AI Provider 错误
    }
}
```

---

## 性能考虑

### 批量操作

- 优先使用 `GetAllProjectStatuses` 而非多次调用 `GetProjectStatus`
- 批量操作使用并发，性能更优

### 缓存

- 项目状态使用 StatusCache 缓存（TTL: 30 秒）
- 配置信息在内存中缓存

### 超时

- 批量状态查询超时: 30 秒
- Git 操作超时: 取决于 Git 配置
- AI API 调用超时: 取决于 Provider 配置

---

## 相关文档

- [前端事件文档](./frontend-events.md)
- [错误处理系统](../architecture/backend-errors.md)
- [StatusCache 架构](../architecture/frontend-status-cache.md)
