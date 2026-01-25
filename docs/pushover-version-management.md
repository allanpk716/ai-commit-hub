# Pushover 版本管理

本文档描述 AI Commit Hub 中 cc-pushover-hook 扩展的版本管理和自动更新功能。

## 目录

- [概述](#概述)
- [自动下载功能](#自动下载功能)
- [VERSION 文件格式](#version-文件格式)
- [更新流程](#更新流程)
- [API 方法](#api-方法)
- [版本兼容性](#版本兼容性)
- [故障排除](#故障排除)

## 概述

cc-pushover-hook 是一个 Git Hook 扩展，用于在 Git 操作时发送 Pushover 和 Windows 通知。AI Commit Hub 提供了完整的版本管理功能，包括：

- **自动下载**: 首次使用时自动从 GitHub 克隆扩展
- **版本检测**: 读取 VERSION 文件获取当前安装的版本
- **更新检查**: 比较本地和远程版本，提示用户更新
- **自动更新**: 一键更新项目和扩展

### 架构组件

```
AI Commit Hub
    ↓
Pushover Service (pkg/pushover/service.go)
    ↓
├── Repository Manager (pkg/pushover/repository.go)
│   └── 管理扩展 Git 仓库（克隆、更新、版本获取）
├── Installer (pkg/pushover/installer.go)
│   └── 执行 install.py 安装/更新 Hook
└── Status Checker (pkg/pushover/status.go)
    └── 检测 Hook 状态、版本、通知模式
```

## 自动下载功能

### 触发时机

扩展会在以下情况下自动下载：

1. **应用启动时**: 检查扩展目录是否存在
2. **安装 Hook 前**: 确保扩展已下载
3. **用户手动触发**: 通过 UI 上的"下载扩展"按钮

### 下载位置

扩展被克隆到以下位置：

- **Windows**: `C:\Users\<username>\.ai-commit-hub\extensions\cc-pushover-hook`
- **macOS/Linux**: `~/.ai-commit-hub/extensions/cc-pushover-hook`

### 下载流程

```go
// pkg/pushover/service.go
func (s *Service) CloneExtension() error {
    return s.repoManager.Clone()
}
```

流程：
1. 检查扩展是否已存在（避免重复下载）
2. 创建 `extensions` 目录（如果不存在）
3. 执行 `git clone -b main --single-branch git@github.com:allanpk716/cc-pushover-hook.git`
4. 返回成功或错误

### 仓库配置

- **仓库 URL**: `git@github.com:allanpk716/cc-pushover-hook.git`
- **分支**: `main`
- **克隆模式**: `--single-branch`（仅克隆主分支，节省空间）

## VERSION 文件格式

### 文件位置

VERSION 文件位于项目内的 Hook 目录：

```
<project-path>/
  └── .claude/
      └── hooks/
          └── pushover-hook/
              ├── pushover-notify.py    # Hook 脚本
              └── VERSION               # 版本文件
```

### 文件格式

VERSION 文件使用简单的键值对格式：

```ini
version=1.0.0
```

**规则**：
- 每行一个键值对
- 格式：`key=value`
- 版本号遵循语义化版本规范（Semantic Versioning）
- 当前仅支持 `version` 字段

### 版本号规范

支持的版本号格式：

- **完整版本**: `1.2.3`（主版本.次版本.补丁版本）
- **带预发布标签**: `1.0.0-alpha`, `2.0.0-beta.1`
- **Git 描述**: `v1.0.0-5-gabcdef`（由 `git describe` 生成）
- **Commit Hash**: `abc1234`（当没有标签时）

### 读取版本

```go
// pkg/pushover/status.go
func (sc *StatusChecker) GetHookVersion() (string, error) {
    versionFilePath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "VERSION")
    data, err := os.ReadFile(versionFilePath)
    if err != nil {
        return "", fmt.Errorf("VERSION file not found: %w", err)
    }

    // 解析 version= 行
    content := strings.TrimSpace(string(data))
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "version=") {
            version := strings.TrimPrefix(line, "version=")
            version = strings.TrimSpace(version)
            if version != "" {
                return version, nil
            }
        }
    }

    return "", fmt.Errorf("no version found in VERSION file")
}
```

## 更新流程

### 完整更新流程

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 用户触发更新（点击"更新 Hook"按钮）                        │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. 前端调用 App.UpdatePushoverHook(projectPath)              │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Pushover Service 检查扩展是否已下载                        │
│    └── 如果未下载，返回错误提示用户先下载扩展                 │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Installer 调用 install.py --force                        │
│    └── --force 标志允许覆盖现有安装                          │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. install.py 执行以下操作：                                  │
│    a. 创建新的 Hook 目录结构                                 │
│    b. 安装 pushover-notify.py 脚本                          │
│    c. 从 git describe 获取版本号                             │
│    d. 写入 VERSION 文件                                      │
│    e. 输出 JSON 格式的安装结果                               │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. Installer 解析 JSON 结果，返回给前端                      │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. 前端显示更新结果，刷新 Hook 状态                          │
└─────────────────────────────────────────────────────────────┘
```

### install.py 的版本获取

install.py 脚本使用以下逻辑获取版本号：

```python
# 在 cc-pushover-hook 仓库中执行
import subprocess

try:
    # 获取 git 描述（优先使用标签）
    version = subprocess.check_output(
        ["git", "describe", "--tags", "--always"],
        cwd=extension_dir,
        text=True
    ).strip()
except subprocess.CalledProcessError:
    # 如果没有标签，使用 commit hash
    version = subprocess.check_output(
        ["git", "rev-parse", "--short", "HEAD"],
        cwd=extension_dir,
        text=True
    ).strip()

# 写入 VERSION 文件
version_file = os.path.join(hook_dir, "VERSION")
with open(version_file, "w") as f:
    f.write(f"version={version}\n")
```

### 扩展更新流程

扩展本身也需要定期更新以获取最新功能：

```go
// pkg/pushover/service.go
func (s *Service) UpdateExtension() error {
    return s.repoManager.Update()
}
```

流程：
1. 检查扩展是否已克隆
2. 执行 `git pull origin main`
3. 更新后可用版本号会变化

### 版本比较

```go
// pkg/pushover/version.go
func CompareVersions(v1, v2 string) int {
    // 返回值: -1 表示 v1 < v2, 0 表示 v1 == v2, 1 表示 v1 > v2
    // 支持语义化版本比较（主版本.次版本.补丁版本）
    // 支持预发布标签（alpha, beta, rc）
}
```

使用示例：

```go
needsUpdate, currentVersion, latestVersion, err := service.CheckForUpdates()
if needsUpdate {
    fmt.Printf("有可用更新: %s -> %s\n", currentVersion, latestVersion)
}
```

## API 方法

### Pushover Service 方法

#### CloneExtension

克隆 cc-pushover-hook 扩展。

```go
func (s *Service) CloneExtension() error
```

**返回**:
- `error`: 错误信息（成功时为 nil）

**示例**:
```go
service := pushover.NewService(appPath)
err := service.CloneExtension()
if err != nil {
    log.Error("克隆扩展失败", err)
}
```

#### UpdateExtension

更新扩展到最新版本。

```go
func (s *Service) UpdateExtension() error
```

**返回**:
- `error`: 错误信息（成功时为 nil）

#### IsExtensionDownloaded

检查扩展是否已下载。

```go
func (s *Service) IsExtensionDownloaded() bool
```

**返回**:
- `bool`: 扩展是否已下载

#### GetExtensionInfo

获取扩展信息。

```go
func (s *Service) GetExtensionInfo() (*ExtensionInfo, error)
```

**返回**:
- `*ExtensionInfo`: 扩展信息结构
- `error`: 错误信息

**ExtensionInfo 结构**:
```go
type ExtensionInfo struct {
    Downloaded      bool   `json:"downloaded"`       // 是否已下载
    Path            string `json:"path"`             // 扩展路径
    Version         string `json:"version"`          // 当前版本
    CurrentVersion  string `json:"current_version"`  // 当前版本（同上）
    LatestVersion   string `json:"latest_version"`   // 最新版本
    UpdateAvailable bool   `json:"update_available"` // 是否有可用更新
}
```

#### CheckForUpdates

检查是否有可用更新。

```go
func (s *Service) CheckForUpdates() (bool, string, string, error)
```

**返回**:
- `bool`: 是否需要更新
- `string`: 当前版本
- `string`: 最新版本
- `error`: 错误信息

#### InstallHook

为项目安装 Hook。

```go
func (s *Service) InstallHook(projectPath string, force bool) (*InstallResult, error)
```

**参数**:
- `projectPath`: 项目路径
- `force`: 是否强制安装（覆盖现有安装）

**返回**:
- `*InstallResult`: 安装结果
- `error`: 错误信息

**InstallResult 结构**:
```go
type InstallResult struct {
    Success  bool   `json:"success"`           // 是否成功
    Message  string `json:"message,omitempty"` // 消息
    HookPath string `json:"hook_path,omitempty"` // Hook 路径
    Version  string `json:"version,omitempty"` // 版本号
}
```

#### UpdateHook

更新项目的 Hook。

```go
func (s *Service) UpdateHook(projectPath string) (*InstallResult, error)
```

**参数**:
- `projectPath`: 项目路径

**返回**:
- `*InstallResult`: 更新结果
- `error`: 错误信息

**注意**: UpdateHook 内部使用 `--force` 标志调用安装脚本。

#### CheckHookInstalled

检查项目的 Hook 是否已安装。

```go
func (s *Service) CheckHookInstalled(projectPath string) bool
```

**参数**:
- `projectPath`: 项目路径

**返回**:
- `bool`: Hook 是否已安装

#### GetHookStatus

获取项目的 Hook 状态。

```go
func (s *Service) GetHookStatus(projectPath string) (*HookStatus, error)
```

**参数**:
- `projectPath`: 项目路径

**返回**:
- `*HookStatus`: Hook 状态信息
- `error`: 错误信息

**HookStatus 结构**:
```go
type HookStatus struct {
    Installed   bool            `json:"installed"`              // 是否已安装
    Mode        NotificationMode `json:"mode"`                  // 通知模式
    Version     string          `json:"version"`               // 版本号
    InstalledAt *time.Time      `json:"installed_at,omitempty"` // 安装时间
}
```

#### SetNotificationMode

设置项目的通知模式。

```go
func (s *Service) SetNotificationMode(projectPath string, mode NotificationMode) error
```

**参数**:
- `projectPath`: 项目路径
- `mode`: 通知模式

**通知模式**:
- `ModeEnabled`: 全部启用（Pushover + Windows）
- `ModePushoverOnly`: 仅 Pushover
- `ModeWindowsOnly`: 仅 Windows
- `ModeDisabled`: 全部禁用

#### UninstallHook

卸载项目的 Hook。

```go
func (s *Service) UninstallHook(projectPath string) error
```

**参数**:
- `projectPath`: 项目路径

**返回**:
- `error`: 错误信息（成功时为 nil）

### App 层 API 方法

这些方法通过 Wails 绑定暴露给前端：

#### GetPushoverExtensionInfo

获取扩展信息。

```go
func (a *App) GetPushoverExtensionInfo() (*pushover.ExtensionInfo, error)
```

#### CheckPushoverUpdates

检查 Pushover 更新。

```go
func (a *App) CheckPushoverUpdates() (bool, string, string, error)
```

#### ClonePushoverExtension

克隆 Pushover 扩展。

```go
func (a *App) ClonePushoverExtension() error
```

#### UpdatePushoverExtension

更新 Pushover 扩展。

```go
func (a *App) UpdatePushoverExtension() error
```

#### UpdatePushoverHook

更新项目的 Pushover Hook。

```go
func (a *App) UpdatePushoverHook(projectPath string) (*pushover.InstallResult, error)
```

## 版本兼容性

### 旧版本兼容

系统兼容两种旧版本的 Hook 安装：

1. **无 VERSION 文件的安装**（版本 1.0.0 之前）
   - 安装位置：`.claude/hooks/pushover-notify.py`
   - 版本检测：返回空字符串或 "unknown"

2. **新版本安装**（版本 1.0.0+）
   - 安装位置：`.claude/hooks/pushover-hook/pushover-notify.py`
   - 版本检测：从 VERSION 文件读取

### 状态检测逻辑

```go
// pkg/pushover/status.go
func (sc *StatusChecker) CheckInstalled() bool {
    // 优先检查新位置
    newHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "pushover-notify.py")
    if _, err := os.Stat(newHookPath); err == nil {
        return true
    }

    // 兼容旧位置
    oldHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-notify.py")
    if _, err := os.Stat(oldHookPath); err == nil {
        return true
    }

    return false
}
```

### 版本迁移

旧版本用户可以通过"更新 Hook"功能迁移到新版本：

1. 检测到旧版本安装（无 VERSION 文件）
2. 用户点击"更新 Hook"
3. install.py 创建新的目录结构
4. 创建 VERSION 文件
5. 保留用户配置（通知模式等）

## 故障排除

### 扩展下载失败

**问题**: 克隆扩展时失败

**可能原因**:
- 网络连接问题
- SSH 密钥未配置
- 磁盘空间不足

**解决方案**:
```bash
# 手动测试 SSH 连接
git clone git@github.com:allanpk716/cc-pushover-hook.git

# 或使用 HTTPS（需要修改代码中的仓库 URL）
git clone https://github.com/allanpk716/cc-pushover-hook.git
```

### VERSION 文件读取失败

**问题**: 无法读取版本号

**可能原因**:
- VERSION 文件不存在（旧版本安装）
- VERSION 文件格式错误
- 文件权限问题

**解决方案**:
```bash
# 检查 VERSION 文件
cat .claude/hooks/pushover-hook/VERSION

# 应该输出类似：
# version=1.0.0

# 如果格式错误，手动修复
echo "version=1.0.0" > .claude/hooks/pushover-hook/VERSION
```

### 更新 Hook 失败

**问题**: 更新 Hook 时报错

**可能原因**:
- Python 未安装或版本过低
- install.py 脚本不存在
- 扩展未下载

**解决方案**:
```bash
# 检查 Python 版本（需要 3.6+）
python --version
python3 --version

# 检查扩展是否存在
ls ~/.ai-commit-hub/extensions/cc-pushover-hook

# 手动运行安装脚本
cd ~/.ai-commit-hub/extensions/cc-pushover-hook
python install.py --target-dir /path/to/project --force
```

### 版本比较错误

**问题**: 版本比较结果不正确

**可能原因**:
- 版本号格式不符合语义化版本规范
- 特殊字符或前缀

**解决方案**:
```go
// 确保版本号格式正确
// 正确: "1.0.0", "2.1.3-beta"
// 错误: "v1.0.0" (v 前缀), "1.0" (缺少补丁版本)

// 测试版本比较
result := pushover.CompareVersions("1.0.0", "1.0.1")
// result 应该是 -1
```

### 通知模式不生效

**问题**: 设置通知模式后不生效

**可能原因**:
- 标记文件创建失败
- Hook 脚本未正确读取标记文件

**解决方案**:
```bash
# 检查标记文件
ls -la .claude/.no-pushover
ls -la .claude/.no-windows

# 手动设置模式
# 仅 Pushover:
touch .claude/.no-windows
rm -f .claude/.no-pushover

# 仅 Windows:
touch .claude/.no-pushover
rm -f .claude/.no-windows
```

### 调试技巧

启用调试输出：

```go
// 在 status.go 中已有调试输出
// 运行时会打印详细的路径检查信息
```

查看调试日志：

```bash
# Windows
$env:DEBUG="1"
wails dev

# Linux/macOS
DEBUG=1 wails dev
```

## 测试

### 运行集成测试

```bash
# 运行所有 Pushover 相关测试
go test ./tests/integration -v -run TestPushover

# 运行特定测试
go test ./tests/integration -v -run TestPushoverUpdateFlow
go test ./tests/integration -v -run TestPushoverVersionComparison

# 运行单元测试
go test ./pkg/pushover -v
```

### 测试覆盖的场景

1. **TestPushoverUpdateFlow**: 完整的更新流程
   - 旧版本安装 → 更新 → VERSION 文件创建

2. **TestPushoverVersionComparison**: 版本比较功能
   - 各种版本号格式的比较

3. **TestPushoverStatusChecker**: 状态检测器
   - 安装检测、版本读取、状态获取

4. **TestPushoverNotificationModes**: 通知模式
   - 四种通知模式的设置和验证

## 参考资源

- [cc-pushover-hook 仓库](https://github.com/allanpk716/cc-pushover-hook)
- [语义化版本规范](https://semver.org/lang/zh-CN/)
- [Git 描述文档](https://git-scm.com/docs/git-describe)
- [Wails 文档](https://wails.io/docs/next/introduction)

## 更新日志

### v1.0.0 (2024-01-25)

- 添加 VERSION 文件支持
- 实现版本检测和比较功能
- 添加自动更新功能
- 兼容旧版本安装
- 添加集成测试
- 完善文档
