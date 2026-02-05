# 日志输出规范

本文档详细说明了 AI Commit Hub 项目中的日志输出规范和最佳实践。

## 日志库

项目统一使用 `github.com/WQGroup/logger` 日志库。

## 基本用法

### 日志级别

```go
import "github.com/WQGroup/logger"

// 基本日志级别
logger.Debug("调试信息")
logger.Info("普通信息")
logger.Warn("警告信息")
logger.Error("错误信息")
```

### 格式化日志

```go
// 格式化版本（支持 printf 风格的格式化）
logger.Debugf("用户 ID: %s, 操作: %s", userID, action)
logger.Infof("应用启动，版本: %s", version)
logger.Warnf("配置文件未找到，使用默认值: %v", defaults)
logger.Errorf("数据库连接失败: %v", err)
```

## 日志配置

### YAML 配置文件

日志库支持通过 YAML 配置文件配置日志行为：

```yaml
logger:
  level: debug          # 日志级别：debug, info, warn, error
  format: json          # 输出格式：json, text
  output:
    - type: file        # 输出到文件
      path: ./logs/app.log
      maxSize: 100      # 单个文件最大大小（MB）
      maxBackups: 3     # 保留的旧日志文件数量
      maxAge: 7         # 日志文件保留天数
      compress: true    # 压缩旧日志文件
```

### 代码配置

```go
import "github.com/WQGroup/logger"

// 配置日志级别
logger.SetLevel(logger.DebugLevel)

// 配置输出格式
logger.SetFormatter(logger.JSONFormatter)

// 配置输出到文件
logger.SetOutput(logger.FileOutput(
    "./logs/app.log",
    100,  // 100MB
    3,    // 保留 3 个备份
    7,    // 保留 7 天
    true, // 压缩
))
```

## 使用场景

### 1. 应用启动和关闭

```go
func startup() {
    logger.Info("AI Commit Hub starting up...")
    logger.Infof("Version: %s", version)
    logger.Infof("Working directory: %s", wd)
    // ...
    logger.Info("AI Commit Hub started successfully")
}

func shutdown() {
    logger.Info("Shutting down AI Commit Hub...")
    // 清理资源
    logger.Info("AI Commit Hub stopped")
}
```

### 2. 错误处理

```go
project, err := r.GetByID(id)
if err != nil {
    logger.Errorf("Failed to get project by ID %s: %v", id, err)
    return nil, err
}
```

### 3. 调试信息

```go
logger.Debugf("Executing git command: %s %v", cmdName, cmdArgs)
logger.Debugf("Git status output: %s", output)
```

### 4. 用户操作

```go
logger.Infof("User added project: %s", projectPath)
logger.Infof("User deleted project: %s", projectName)
```

### 5. AI 请求

```go
logger.Infof("Sending request to AI provider: %s", providerName)
logger.Debugf("Request prompt: %s", prompt)
logger.Debugf("Request options: %+v", options)
```

## 最佳实践

### 1. 选择合适的日志级别

- **Debug**: 详细的调试信息，仅在开发环境使用
- **Info**: 一般信息，记录应用的正常流程
- **Warn**: 警告信息，不影响应用运行但需要注意
- **Error**: 错误信息，需要关注和处理的错误

### 2. 避免日志泄露敏感信息

```go
// ❌ 错误：记录敏感信息
logger.Infof("User login: username=%s, password=%s", username, password)

// ✅ 正确：只记录必要的非敏感信息
logger.Infof("User login attempt: username=%s", username)
```

### 3. 使用结构化日志（JSON 格式）

```go
// JSON 格式更适合日志分析和查询
logger.SetFormatter(logger.JSONFormatter)

// 日志输出示例：
// {"level":"info","time":"2024-01-01T12:00:00Z","msg":"User added project","project_path":"C:\\Projects\\my-project"}
```

### 4. 日志轮转和清理

```go
// 配置日志轮转，防止单个日志文件过大
logger.SetOutput(logger.FileOutput(
    "./logs/app.log",
    100,  // 100MB 后轮转
    3,    // 保留 3 个备份
    7,    // 7 天后清理
    true, // 压缩旧日志
))
```

### 5. 线程安全

`github.com/WQGroup/logger` 是线程安全的，可以在 goroutine 中安全使用：

```go
go func() {
    logger.Info("Background task started")
    defer logger.Info("Background task completed")
    // ...
}()
```

### 6. 性能考虑

```go
// ❌ 错误：在热路径上频繁记录 Debug 日志
for i := 0; i < 1000000; i++ {
    logger.Debugf("Processing item %d", i)  // 即使日志级别是 Info，也会执行字符串格式化
}

// ✅ 正确：使用条件判断或日志级别的惰性求值
for i := 0; i < 1000000; i++ {
    if logger.IsLevelEnabled(logger.DebugLevel) {
        logger.Debugf("Processing item %d", i)
    }
}
```

## 常见问题

### Q: 为什么不使用 `fmt.Printf` 或 `log.Println`？

A: `github.com/WQGroup/logger` 提供了以下优势：
- 统一的日志级别管理
- 支持多种输出格式（JSON、文本）
- 日志轮转和自动清理
- 结构化日志，便于日志分析
- 线程安全
- 可配置的输出目标（文件、控制台、远程日志服务）

### Q: 生产环境应该使用什么日志级别？

A: 生产环境建议使用 `Info` 或 `Warn` 级别，避免产生过多日志。可以在配置文件中设置：

```yaml
logger:
  level: info  # 生产环境使用 info
  format: json
```

### Q: 如何在生产环境中禁用控制台日志输出？

A: 配置日志只输出到文件：

```go
// 只输出到文件，不输出到控制台
logger.SetOutput(logger.FileOutput(
    "./logs/app.log",
    100, 3, 7, true,
))
```

## 相关文档

- Wails 开发规范：`docs/development/wails-development-standards.md`
- CLAUDE.md：项目根目录
