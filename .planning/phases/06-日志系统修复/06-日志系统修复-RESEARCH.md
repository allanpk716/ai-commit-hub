# Phase 06: 日志系统修复 - Research

**Researched:** 2026-02-08
**Domain:** Go Logging with WQGroup/logger (Logrus wrapper)
**Confidence:** HIGH

## Summary

本阶段研究如何修复 AI Commit Hub 项目中的日志系统问题。项目使用 `github.com/WQGroup/logger v0.0.16`，这是一个基于 logrus 的日志库，支持日志轮转、自动清理和结构化日志。当前存在三个主要问题：1) 日志格式不符合 withField 标准；2) 日志输出路径未正确配置到可执行文件目录的 logs 文件夹；3) 大量 logger 调用使用了错误的方法签名（应使用 `WithField` 链式调用或 `*f` 格式化方法，而非带多个参数的普通方法）。

研究发现 WQGroup/logger 库虽然 API 与 logrus 兼容，但项目当前使用模式与库的标准用法不符。该库已内置日志轮转和清理功能，使用 lumberjack.v2 作为底层实现。日志系统配置应通过 `Settings` 结构体或 YAML 配置文件完成，支持多种格式器（withField、easy、json、text），其中 withField 格式器为默认和推荐选项。

**Primary recommendation:** 使用 WQGroup/logger 的 `Settings` 结构体配置日志系统，设置 `FormatterType="withField"`，通过 `SetLoggerSettings()` 应用配置；将所有 `logger.Info/Warn/Error` 的多参数调用改为 `logger.WithField().Info()` 或 `logger.Infof()`；使用可执行文件目录作为日志根目录。

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| WQGroup/logger | v0.0.16 | 统一日志库（基于 logrus + lumberjack） | 项目已依赖，支持轮转、清理、结构化日志 |
| sirupsen/logrus | v1.9.3 | WQGroup/logger 底层日志框架 | 业界标准结构化日志库 |
| natefinch/lumberjack.v2 | v2.2.1 | 日志轮转和清理 | 业界标准日志轮转库，WQGroup/logger 已内置 |
| lestrrat-go/file-rotatelogs | v2.4.0 | 时间轮转支持 | WQGroup/logger 已集成此依赖 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gopkg.in/yaml.v3 | v3.0.1 | YAML 配置解析 | 当需要从配置文件加载日志设置时 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| WQGroup/logger | uber-go/zap | Zap 性能更高但 API 不兼容当前代码，需要大量重构 |
| WQGroup/logger | rs/zerolog | Zerolog 零分配但同样需要大量重构 |
| Settings 代码配置 | YAML 配置文件 | YAML 更灵活但增加配置文件维护成本；代码配置更直接 |

**Installation:**
```bash
# 已安装，无需额外安装
go get github.com/WQGroup/logger@v0.0.16
```

## Architecture Patterns

### WQGroup/logger 初始化流程

```
1. 创建 Settings 结构体 (logger.NewSettings())
2. 配置日志参数 (Level, FormatterType, LogRootFPath, etc.)
3. 应用设置 (logger.SetLoggerSettings(settings))
4. 使用 logger 全局函数 (Info, Error, WithField, etc.)
```

### Pattern 1: 代码配置日志系统（推荐）

**What:** 在应用启动时通过 Settings 结构体配置日志系统
**When to use:** 所有 Wails 应用，避免配置文件管理复杂度
**Example:**
```go
// Source: WQGroup/logger Go documentation
import "github.com/WQGroup/logger"

func initLogger() error {
    settings := logger.NewSettings()

    // 格式器配置
    settings.FormatterType = "withField"  // 默认格式器
    settings.TimestampFormat = "2006-01-02 15:04:05.000"  // 毫秒级时间戳
    settings.DisableTimestamp = false
    settings.DisableLevel = false
    settings.DisableCaller = true  // 禁用调用者信息（默认）

    // 日志级别
    settings.Level = logrus.InfoLevel

    // 输出路径配置
    exeDir, err := os.Executable()
    if err != nil {
        return err
    }
    exePath := filepath.Dir(exeDir)
    settings.LogRootFPath = filepath.Join(exePath, "logs")

    // 轮转策略
    settings.RotationTime = 24 * time.Hour  // 24小时轮转
    settings.MaxSizeMB = 10  // 10MB 轮转
    settings.MaxAgeDays = 30  // 保留30天

    // 应用配置
    logger.SetLoggerSettings(settings)
    logger.SetLoggerName("ai-commit-hub")

    return nil
}
```

### Pattern 2: 使用 withField 格式器记录结构化日志

**What:** 使用 `WithField()` 或 `WithFields()` 添加结构化字段
**When to use:** 需要记录额外的上下文信息（如 requestID、userID、模块名等）
**Example:**
```go
// 正确：使用 WithField 添加模块字段
logger.WithField("module", "startup").Info("AI Commit Hub starting up...")

// 正确：使用 WithFields 添加多个字段
logger.WithFields(map[string]interface{}{
    "module": "update",
    "version": updateInfo.LatestVersion,
}).Info("发现新版本")

// 正确：使用格式化方法
logger.Infof("成功预加载 %d 个项目的状态", len(statuses))

// 错误：带多个参数的非格式化方法（当前代码中常见）
// logger.Info("发现新版本", "version", updateInfo.LatestVersion)  // ❌ 错误
```

### Pattern 3: 日志轮转和自动清理

**What:** 使用 lumberjack 内置的轮转和清理机制
**When to use:** 所有生产环境，防止日志文件无限增长
**Example:**
```go
// Settings 配置会自动应用轮转和清理策略
settings.MaxSizeMB = 10       // 10MB 时大小轮转
settings.MaxAgeDays = 30      // 30天后删除旧日志
settings.RotationTime = 24 * time.Hour  // 24小时时间轮转

// 轮转文件命名格式：
// 时间轮转：logger--YYYYMMDDHHMM--.log
// 大小轮转：当前文件名 + .001, .002 等后缀
```

### Pattern 4: 获取可执行文件目录（Wails 应用）

**What:** 获取可执行文件所在目录作为日志根目录
**When to use:** Wails 桌面应用，日志文件应与可执行文件放在同一目录
**Example:**
```go
// 获取可执行文件完整路径
exePath, err := os.Executable()
if err != nil {
    return fmt.Errorf("获取可执行文件路径失败: %w", err)
}

// 获取可执行文件所在目录
exeDir := filepath.Dir(exePath)

// 设置日志根目录
logDir := filepath.Join(exeDir, "logs")

// 确保 logs 目录存在
if err := os.MkdirAll(logDir, 0755); err != nil {
    return fmt.Errorf("创建日志目录失败: %w", err)
}

// 配置日志系统
settings.LogRootFPath = logDir
```

### Anti-Patterns to Avoid

- **Anti-pattern 1: 使用多参数非格式化日志方法**
  - 为什么错误：`logger.Info("msg", "key", value)` 这种写法在 logrus 中不支持，会输出奇怪的格式
  - 正确做法：使用 `logger.WithField("key", value).Info("msg")` 或 `logger.Infof("msg %v", value)`

- **Anti-pattern 2: 在热路径上频繁记录 Debug 日志**
  - 为什么错误：即使日志级别是 Info，字符串格式化仍会执行
  - 正确做法：使用 `if logger.IsLevelEnabled(logger.DebugLevel) { logger.Debugf(...) }`

- **Anti-pattern 3: 混合使用多种格式器**
  - 为什么错误：会导致日志格式不一致
  - 正确做法：统一使用 withField 格式器

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 日志轮转 | 自己实现文件大小检查和文件重命名 | WQGroup/logger 内置的 lumberjack.v2 | 边界情况多（并发、文件锁、压缩、清理），lumberjack 已充分测试 |
| 日志清理 | 自己写定时任务删除旧日志 | WQGroup/logger 的 MaxAgeDays 配置 | 需要处理文件名解析、时间戳比较、空目录清理等复杂逻辑 |
| 结构化字段 | 自己拼接字符串 | logrus.WithField() / WithFields() | 支持类型安全、JSON 序列化、字段过滤等高级功能 |

**Key insight:** WQGroup/logger 已经内置了日志轮转、清理和结构化日志功能，无需重复造轮子。正确配置 Settings 结构体即可获得所有功能。

## Common Pitfalls

### Pitfall 1: 使用错误的 logger 方法签名

**What goes wrong:** 代码中大量使用 `logger.Info("message", "key", value)` 这种多参数调用方式，这是不正确的 logrus 用法。

**Why it happens:** 开发者可能混淆了其他语言（如 Python logging.info(msg, key=value)）或其他日志库（如 Zap）的 API。logrus 的非格式化方法（Info/Error/Warn/Debug）只接受单个参数或可变参数 `...interface{}`，但会直接用空格拼接所有参数，不会处理键值对。

**How to avoid:**
1. **需要格式化时使用 *f 方法：**
   ```go
   logger.Infof("成功预加载 %d 个项目的状态", len(statuses))
   logger.Errorf("数据库连接失败: %v", err)
   ```

2. **需要添加结构化字段时使用 WithField：**
   ```go
   logger.WithField("module", "startup").Info("AI Commit Hub starting up...")
   logger.WithFields(map[string]interface{}{
       "module": "update",
       "version": updateInfo.LatestVersion,
   }).Info("发现新版本")
   ```

3. **简单消息使用普通方法：**
   ```go
   logger.Info("AI Commit Hub initialized successfully")
   logger.Warn("Pushover service 未初始化，跳过 Hook 状态同步")
   ```

**Warning signs:**
- 日志输出中出现奇怪的空格分隔格式
- 编译器警告 "too many arguments to function"
- 静态分析工具报错

### Pitfall 2: 日志路径配置错误

**What goes wrong:** 日志文件未输出到可执行文件目录的 logs 文件夹，而是输出到当前工作目录或用户目录。

**Why it happens:** Wails 应用可能从不同路径启动（快捷方式、命令行等），依赖当前工作目录不可靠。

**How to avoid:**
```go
// 始终使用可执行文件所在目录
exePath, err := os.Executable()
if err != nil {
    return fmt.Errorf("获取可执行文件路径失败: %w", err)
}
exeDir := filepath.Dir(exePath)
logDir := filepath.Join(exeDir, "logs")

// 确保 logs 目录存在
if err := os.MkdirAll(logDir, 0755); err != nil {
    return fmt.Errorf("创建日志目录失败: %w", err)
}

settings.LogRootFPath = logDir
```

**Warning signs:**
- 日志文件出现在项目根目录而非可执行文件目录
- 不同启动方式下日志文件位置不一致
- 日志文件写入失败（权限问题）

### Pitfall 3: 忽略日志清理

**What goes wrong:** 日志文件无限增长，最终占满磁盘空间。

**Why it happens:** 开发阶段未配置轮转和清理策略，或者使用默认配置（MaxAgeDays 默认为 0，表示不清理）。

**How to avoid:**
```go
settings.MaxSizeMB = 10        // 10MB 时大小轮转
settings.MaxAgeDays = 30       // 30天后删除旧日志
settings.RotationTime = 24 * time.Hour  // 24小时时间轮转
```

**Warning signs:**
- logs 目录占用空间持续增长
- 出现单个数 GB 的日志文件
- 磁盘空间不足警告

### Pitfall 4: 混合使用中英文日志消息

**What goes wrong:** 日志消息中混用中英文，降低日志可读性。

**Why it happens:** 不同开发者习惯不同，或复制粘贴代码。

**How to avoid:**
- 统一使用中文日志消息（项目已决定）
- 仅在技术术语、错误消息（来自库函数）中使用英文

**Warning signs:**
- 日志查询时需要同时搜索中英文关键词
- 日志分析工具无法正确解析

## Code Examples

Verified patterns from official sources:

### 配置日志系统（完整示例）

```go
// Source: WQGroup/logger package documentation
import (
    "github.com/WQGroup/logger"
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
)

func setupLogger() error {
    // 1. 创建 Settings
    settings := logger.NewSettings()

    // 2. 配置格式器（withField 格式器）
    settings.FormatterType = "withField"
    settings.TimestampFormat = "2006-01-02 15:04:05.000"
    settings.DisableTimestamp = false
    settings.DisableLevel = false
    settings.DisableCaller = true

    // 3. 配置日志级别
    settings.Level = logrus.InfoLevel

    // 4. 配置输出路径（使用可执行文件目录）
    exePath, err := os.Executable()
    if err != nil {
        return err
    }
    exeDir := filepath.Dir(exePath)
    logDir := filepath.Join(exeDir, "logs")
    settings.LogRootFPath = logDir

    // 5. 配置轮转和清理策略
    settings.RotationTime = 24 * time.Hour  // 24小时轮转
    settings.MaxSizeMB = 10                 // 10MB 大小轮转
    settings.MaxAgeDays = 30                // 保留30天

    // 6. 应用配置
    logger.SetLoggerSettings(settings)
    logger.SetLoggerName("ai-commit-hub")

    return nil
}
```

**Expected output:**
```
2025-12-18 18:32:07.379 - [INFO]: AI Commit Hub starting up...
2025-12-18 18:32:07.380 - [INFO]: 发现新版本 version=v1.0.1
2025-12-18 18:32:08.123 - [ERROR]: 数据库连接失败 connection refused
```

### 正确的 logger 调用方式

```go
// 场景 1: 简单消息
logger.Info("AI Commit Hub initialized successfully")
logger.Warn("Pushover service 未初始化，跳过 Hook 状态同步")

// 场景 2: 格式化消息
logger.Infof("成功预加载 %d 个项目的状态", len(statuses))
logger.Errorf("数据库连接失败: %v", err)

// 场景 3: 带结构化字段（模块名）
logger.WithField("module", "startup").Info("AI Commit Hub starting up...")
logger.WithField("module", "update").Infof("发现新版本: %s", updateInfo.LatestVersion)

// 场景 4: 带多个结构化字段
logger.WithFields(map[string]interface{}{
    "module": "updater",
    "url": downloadURL,
    "asset": assetName,
}).Info("开始安装更新")

// 场景 5: 错误日志（带错误对象）
if err != nil {
    logger.WithFields(map[string]interface{}{
        "module": "database",
        "error": err.Error(),
    }).Error("Failed to initialize database")
}
```

### 静态分析检查错误的 logger 调用

```go
// 使用 grep 查找可能错误的调用
// grep -rn 'logger\.\(Info\|Error\|Warn\|Debug\)([^,)]*,[^)]*)' *.go

// 错误示例（需要修复）：
// logger.Info("发现新版本", "version", updateInfo.LatestVersion)
// logger.Warn("退出逻辑已被调用过,这是不应该的状态")
// logger.Info("托盘图标双击,显示窗口")

// 修复后：
// logger.WithField("version", updateInfo.LatestVersion).Info("发现新版本")
// logger.Warn("退出逻辑已被调用过，这是不应该的状态")
// logger.WithField("module", "systray").Info("托盘图标双击，显示窗口")
```

### YAML 配置方式（可选）

```yaml
# config/logger.yaml
logger:
  log_name_base: "ai-commit-hub"
  level: "info"
  days_to_keep: 30
  max_size_mb: 10
  use_hierarchical_path: false

  # 格式器配置
  formatter_type: "withField"
  timestamp_format: "2006-01-02 15:04:05.000"
  disable_timestamp: false
  disable_level: false
  disable_caller: true
  full_timestamp: false
```

```go
// 从 YAML 加载配置
settings, err := logger.LoadSettingsFromYAML("config/logger.yaml")
if err != nil {
    return err
}
logger.SetLoggerSettings(settings)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| log 标准库 | 结构化日志库（logrus/zap/zerolog） | 2018-2020 | 结构化日志成为主流，支持 JSON 格式、字段过滤 |
| 手动轮转脚本 | lumberjack 自动轮转 | 2016+ | 简化日志管理，防止磁盘满 |
| 文本日志格式 | JSON/结构化格式 | 2019+ | 便于日志聚合和分析（ELK、Loki 等） |
| 单一日志文件 | 按日期/大小轮转 | 长期最佳实践 | 提高查询性能，降低单文件大小 |

**Deprecated/outdated:**
- **log 标准库**：功能过于简单，不支持结构化日志和日志级别
- **手动日志轮转**：容易出错，不建议自己实现
- **fmt.Printf 用于日志**：无法控制日志级别和输出目标
- **混合使用多种日志库**：导致配置不一致，应统一使用 WQGroup/logger

**Current best practices (2026):**
- 使用结构化日志（logrus.WithField）
- 自动日志轮转和清理（lumberjack）
- 统一配置管理（Settings 或 YAML）
- 可观测性集成（JSON 格式便于导入日志系统）

## Open Questions

1. **清理策略执行时机**
   - What we know: lumberjack 在每次轮转时执行清理（millRunOnce）
   - What's unclear: 是否需要应用启动时额外执行一次清理？
   - Recommendation: 使用 lumberjack 默认行为即可，无需额外清理任务。lumberjack 会在创建新日志文件时自动清理旧日志。如果需要确保应用启动时清理，可以在 setupLogger 中手动调用 `logger.CleanupExpiredLogs(logDir, 30)`。

2. **模块名字段命名规范**
   - What we know: 需要使用 `WithField("module", "模块名")` 添加模块标识
   - What's unclear: 模块名应使用中文还是英文？是否需要标准化命名列表？
   - Recommendation:
     - 优先使用中文模块名（与项目日志消息语言一致）
     - 建议的模块名列表：startup, database, config, git, ai, update, systray, pushover, project
     - 在 docs/development/logging-standards.md 中维护模块名命名规范

3. **日志轮转文件命名格式**
   - What we know: WQGroup/logger 使用 lumberjack，默认轮转文件名为 `logger--YYYYMMDDHHMM--.log`
   - What's unclear: 用户期望的文件名格式是 `YYYY-MM-DD.log` 还是 `YYYY-MM-DD.001.log`？
   - Recommendation: 根据用户需求，时间轮转使用 `YYYY-MM-DD.log`，大小轮转使用 `YYYY-MM-DD.001.log`。这需要自定义 lumberjack 的 Filename 格式，或者接受 WQGroup/logger 的默认命名格式。建议接受默认格式以减少配置复杂度。

## Sources

### Primary (HIGH confidence)

- **github.com/WQGroup/logger** - WQGroup/logger 源代码和 README
  - API 文档：Settings 结构体、WithField 方法、轮转配置
  - 验证日期：2026-02-08
  - URL: https://github.com/WQGroup/logger

- **go doc github.com/WQGroup/logger** - Go 包文档
  - 函数签名、类型定义、常量
  - 验证日期：2026-02-08

- **pkg.go.dev/github.com/natefinch/lumberjack.v2** - Lumberjack 官方文档
  - 轮转策略、清理机制、配置选项
  - 验证日期：2026-02-08
  - URL: https://pkg.go.dev/github.com/natefinch/lumberjack.v2

- **docs/development/logging-standards.md** - 项目日志规范文档
  - 当前日志使用模式、配置示例
  - 验证日期：2026-02-08

### Secondary (MEDIUM confidence)

- **go doc github.com/WQGroup/logger.Settings** - Settings 结构体字段说明
  - 验证日期：2026-02-08

- **项目代码分析** - 实际 logger 调用模式
  - app.go 中的 logger 使用示例
  - 错误的 logger 调用示例（多参数非格式化方法）
  - 验证日期：2026-02-08

### Tertiary (LOW confidence)

- **WebSearch 结果** - 未直接使用，仅用于验证
  - 大多数结果关于 logrus、zap、zerolog 等，而非 WQGroup/logger
  - 未提供 WQGroup/logger 的具体 API 使用方法

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - 项目已使用 WQGroup/logger v0.0.16，Go 文档和源代码可验证
- Architecture: **HIGH** - WQGroup/logger API 和 Settings 配置已通过 go doc 验证
- Pitfalls: **HIGH** - 通过代码分析发现大量错误的 logger 调用模式
- Code examples: **HIGH** - 所有示例基于 Go 文档和项目实际代码

**Research date:** 2026-02-08
**Valid until:** 2026-03-10 (30 days - WQGroup/logger 是稳定库，但需验证是否有新版本)

**Researcher notes:**
- WQGroup/logger 库的公开文档较少，主要通过 Go doc 和源代码分析 API
- 当前代码中存在大量 logger 调用错误（约 50+ 处），需要系统性修复
- 建议分 3 个独立 plan 修复：1) 格式配置 2) 路径配置 3) 方法签名修复
- WQGroup/logger 的 withField 格式器默认输出格式符合用户需求，无需自定义格式器
- 日志轮转和清理功能已内置，正确配置 Settings 即可
