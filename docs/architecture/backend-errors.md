# Backend Error Handling System

## 错误类型层次

```
error (interface)
  ├── AppInitError
  ├── ValidationError
  ├── GitOperationError
  └── AIProviderError
```

## 错误类型详解

### AppInitError（应用初始化错误）

**文件：** `pkg/errors/app_errors.go`

**用途：** 表示应用初始化阶段的错误

**结构：**
```go
type AppInitError struct {
    OriginalErr error
}

func (e *AppInitError) Error() string {
    return fmt.Sprintf("app not initialized: %v", e.OriginalErr)
}
```

**使用场景：**
- 数据库连接失败
- 配置文件加载失败
- 必要的服务初始化失败

**辅助函数：**
```go
func CheckInit(initErr error) error {
    if initErr != nil {
        return &AppInitError{OriginalErr: initErr}
    }
    return nil
}
```

**使用示例：**
```go
func (a *App) SomeMethod() error {
    if err := a.initError; err != nil {
        return apperrors.CheckInit(err)
    }
    // ... 方法逻辑
}
```

### ValidationError（验证错误）

**文件：** `pkg/errors/domain_errors.go`

**用途：** 表示数据验证失败

**结构：**
```go
type ValidationError struct {
    Field   string  // 验证失败的字段名
    Message string  // 错误信息
    Err     error   // 原始错误
}

func (e *ValidationError) Error() string {
    if e.Field != "" {
        return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
    }
    return fmt.Sprintf("validation failed: %s", e.Message)
}
```

**使用场景：**
- 配置参数验证
- 用户输入验证
- 数据完整性检查

**创建函数：**
```go
func NewValidationError(field, message string, err error) *ValidationError {
    return &ValidationError{
        Field:   field,
        Message: message,
        Err:     err,
    }
}
```

**使用示例：**
```go
func (s *ConfigService) ValidateConfig(config *models.Config) error {
    if config.Provider == "" {
        return apperrors.NewValidationError("provider", "cannot be empty", nil)
    }

    if config.ApiKey == "" && config.Provider != "ollama" {
        return apperrors.NewValidationError("api_key", "required for this provider", nil)
    }

    return nil
}
```

### GitOperationError（Git 操作错误）

**文件：** `pkg/errors/domain_errors.go`

**用途：** 表示 Git 命令执行失败

**结构：**
```go
type GitOperationError struct {
    Operation string  // Git 操作类型（如 "commit", "status", "push"）
    Path      string  // 项目路径
    Err       error   // 原始错误
}

func (e *GitOperationError) Error() string {
    return fmt.Sprintf("git operation '%s' failed at %s: %v", e.Operation, e.Path, e.Err)
}
```

**使用场景：**
- Git 命令执行失败
- 仓库状态检查失败
- 提交/推送操作失败

**创建函数：**
```go
func NewGitOperationError(operation, path string, err error) *GitOperationError {
    return &GitOperationError{
        Operation: operation,
        Path:      path,
        Err:       err,
    }
}
```

**使用示例：**
```go
func (s *CommitService) GenerateCommit(projectPath string) error {
    // 获取 diff
    diff, err := git.GetStagedDiff(context.Background())
    if err != nil {
        return apperrors.NewGitOperationError("diff --cached", projectPath, err)
    }

    // ... 其他逻辑
}
```

### AIProviderError（AI Provider 错误）

**文件：** `pkg/errors/domain_errors.go`

**用途：** 表示 AI Provider 调用失败

**结构：**
```go
type AIProviderError struct {
    Provider string  // Provider 名称（如 "openai", "anthropic"）
    Message  string  // 错误信息
    Err      error   // 原始错误
}

func (e *AIProviderError) Error() string {
    return fmt.Sprintf("AI provider '%s' error: %s", e.Provider, e.Message)
}
```

**使用场景：**
- AI client 创建失败
- API 调用失败
- 模型生成失败

**创建函数：**
```go
func NewAIProviderError(provider, message string, err error) *AIProviderError {
    return &AIProviderError{
        Provider: provider,
        Message:  message,
        Err:      err,
    }
}
```

**使用示例：**
```go
func (s *CommitService) GenerateCommit(projectPath string) error {
    client, err := factory(ctx, cfg.Provider, ps)
    if err != nil {
        return apperrors.NewAIProviderError(cfg.Provider, "failed to create AI client", err)
    }

    msg, err := client.GetCommitMessage(ctx, promptText)
    if err != nil {
        return apperrors.NewAIProviderError(cfg.Provider, "failed to generate commit message", err)
    }

    // ... 其他逻辑
}
```

## 错误处理最佳实践

### 1. 创建错误

```go
// 验证错误
return apperrors.NewValidationError("path", "cannot be empty", nil)

// Git 错误
return apperrors.NewGitOperationError("commit", projectPath, err)

// AI Provider 错误
return apperrors.NewAIProviderError("openai", "rate limit exceeded", err)
```

### 2. 检查错误类型

```go
import apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"

if apperrors.IsValidationError(err) {
    // 处理验证错误
    fmt.Println("Validation error:", err)
}

if apperrors.IsGitError(err) {
    // 处理 Git 错误
    fmt.Println("Git error:", err)
}

if apperrors.IsAIProviderError(err) {
    // 处理 AI Provider 错误
    fmt.Println("AI Provider error:", err)
}

if apperrors.IsNotFoundError(err) {
    // 处理未找到错误
    fmt.Println("Not found:", err)
}
```

### 3. 错误包装

```go
// 包装错误，保留上下文
if err != nil {
    return fmt.Errorf("failed to save project: %w", err)
}

// 使用领域错误类型包装
if err != nil {
    return apperrors.NewGitOperationError("commit", path, err)
}
```

### 4. 错误传播规则

**原则：**
1. **不要丢弃错误**: 所有错误都必须处理或传播
2. **添加上下文**: 包装错误时添加有意义的上下文信息
3. **使用正确的错误类型**: 根据错误性质选择合适的领域错误类型
4. **日志记录**: 在适当的层级记录错误日志

**示例：**
```go
// Service 层：使用领域错误
func (s *CommitService) GenerateCommit(path string) error {
    diff, err := git.GetStagedDiff(ctx)
    if err != nil {
        // 使用领域错误包装，提供上下文
        return apperrors.NewGitOperationError("diff --cached", path, err)
    }
    // ...
}

// App 层：记录日志并传播
func (a *App) GenerateCommit(path string) error {
    logger.Infof("Generating commit for %s", path)
    err := a.commitService.GenerateCommit(path)
    if err != nil {
        logger.Errorf("Failed to generate commit: %v", err)
        return err // 错误已经包含了足够的上下文
    }
    // ...
}
```

## 错误类型检查函数

```go
// IsNotFoundError 检查是否是"未找到"错误
func IsNotFoundError(err error) bool

// IsValidationError 检查是否是验证错误
func IsValidationError(err error) bool

// IsGitError 检查是否是 Git 操作错误
func IsGitError(err error) bool

// IsAIProviderError 检查是否是 AI Provider 错误
func IsAIProviderError(err error) bool
```

## 错误处理流程图

```
应用启动
    ↓
初始化检查
    ↓ (失败)
AppInitError
    ↓
用户操作
    ↓
输入验证
    ↓ (失败)
ValidationError
    ↓
Git 操作
    ↓ (失败)
GitOperationError
    ↓
AI 调用
    ↓ (失败)
AIProviderError
    ↓
错误处理
    ↓
用户友好的错误消息
```

## 前端错误传递

**后端 -> 前端错误传递：**

1. **Wails Events**: 通过事件系统传递错误
```typescript
// Go 后端
runtime.EventsEmit(ctx, "commit-error", errMsg)

// TypeScript 前端
EventsOn('commit-error', (errMsg: string) => {
  error.value = errMsg
})
```

2. **API 返回值**: 通过方法返回值传递错误
```go
// Go 后端
func (a *App) SomeMethod() error {
    return apperrors.NewValidationError("field", "message", err)
}

// TypeScript 前端
try {
  await SomeMethod()
} catch (e) {
  error.value = e.message
}
```

## 错误日志记录

**使用 logger 包记录错误：**

```go
import "github.com/WQGroup/logger"

// Info 级别：正常流程信息
logger.Info("Starting commit generation")

// Warn 级别：警告信息
logger.Warn("Provider not configured, using fallback")

// Error 级别：错误信息
logger.Errorf("Failed to generate commit: %v", err)

// Debug 级别：调试信息
logger.Debugf("Current status: %+v", status)
```

## 相关文件

- `pkg/errors/app_errors.go`: 应用初始化错误
- `pkg/errors/domain_errors.go`: 领域错误类型
- `pkg/service/commit_service.go`: CommitService 错误处理示例
- `pkg/service/config_service.go`: ConfigService 验证示例
