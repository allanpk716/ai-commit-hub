# AI Commit Hub 自动更新功能设计文档

**日期**: 2026-02-02
**作者**: Allanpk716
**版本**: 1.0

---

## 1. 概述

本文档描述 AI Commit Hub 自动更新功能的完整设计方案。该功能允许应用在启动时自动检查 GitHub Release 上的最新版本，并在有更新时提供一键下载和安装功能。

### 1.1 核心目标

- **自动化 CI/CD**: Git tag 推送后自动通过 GitHub Actions 构建和发布
- **自动更新检测**: 程序启动时自动检查最新版本
- **无缝更新体验**: 下载、安装、重启全自动，用户无需手动操作
- **跨平台支持**: 同时支持 Windows 和 macOS 平台

### 1.2 设计原则

- **非侵入式**: 更新提示不打断用户正常使用
- **用户控制**: 用户可选择何时更新，支持跳过版本
- **容错设计**: 更新失败时自动回滚，不影响用户使用
- **轻量级**: 便携版发布，无需安装程序

---

## 2. 整体架构

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      GitHub Repository                       │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ Git Tag Push │────────>│ GitHub Actions│                 │
│  │  (v1.0.0)    │         │   Build &    │                 │
│  └──────────────┘         │   Release    │                 │
│                           └──────┬───────┘                 │
└──────────────────────────────────┼──────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    AI Commit Hub App                        │
│  ┌──────────┐    ┌─────────────┐    ┌──────────────┐      │
│  │  Version │───>│   Update    │───>│   Downloader │      │
│  │  Module  │    │   Service   │    │              │      │
│  └──────────┘    └──────┬──────┘    └──────┬───────┘      │
│                         │                  │              │
│                         ▼                  ▼              │
│  ┌──────────┐    ┌─────────────┐    ┌──────────────┐      │
│  │   UI     │<───│  Wails      │<───│   Installer  │      │
│  │Components│    │   Events    │    │  + Updater   │      │
│  └──────────┘    └─────────────┘    └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
应用启动
  ↓
版本模块获取当前版本
  ↓
UpdateService 检查 GitHub Release
  ↓
版本比较
  ↓ (有更新)
UI 显示内联提示条
  ↓ (用户点击下载)
下载器下载 zip
  ↓ (下载完成)
更新器进程启动
  ↓
主程序退出
  ↓
更新器解压并替换文件
  ↓
启动新版本程序
```

---

## 3. GitHub Actions 工作流

### 3.1 工作流配置

**文件**: `.github/workflows/release.yml`

**触发条件**:
- 推送匹配 `v*` 模式的 tag（如 `v1.0.0`）
- 手动触发（workflow_dispatch）

**构建矩阵**:
- Windows (windows-latest)
- macOS (macos-latest)

**构建步骤**:

1. **环境准备**
   - 检出代码
   - 安装 Go 1.21+
   - 安装 Wails CLI
   - 安装 Node.js 18+

2. **版本注入**
   - 从 GitHub tag 提取版本号
   - 设置 VERSION 环境变量

3. **Wails 构建**
   ```bash
   wails build -clean -ldflags "-X main.version=${VERSION}"
   ```

4. **打包**
   - Windows: `ai-commit-hub-{version}-windows.zip`
   - macOS: `ai-commit-hub-{version}-darwin.zip`

5. **发布**
   - 创建 GitHub Release
   - 上传 zip 文件
   - 自动生成 Release Notes

### 3.2 优化点

- 使用缓存加速构建（Go modules、node_modules）
- 并行构建多个平台
- 构建失败时可配置通知

---

## 4. 版本管理模块

### 4.1 模块结构

**文件**: `pkg/version/version.go`

**核心功能**:

1. **版本变量定义**
   ```go
   var (
       Version   = "dev"      // 编译时注入
       CommitSHA = "unknown"  // Git commit hash
       BuildTime = "unknown"  // 构建时间
   )
   ```

2. **版本获取方法**
   - `GetVersion() string`: 返回当前版本
     - 开发模式: `"dev-uncommitted"`
     - 生产模式: `"v1.0.0"`

   - `GetFullVersion() string`: 返回完整版本信息
     - 格式: `"v1.0.0 (abc1234 2024-01-15)"`

   - `IsDevVersion() bool`: 判断是否为开发版本

3. **版本比较方法**
   - `CompareVersions(v1, v2 string) int`: 比较两个版本号
     - 返回 1: v1 > v2
     - 返回 0: v1 == v2
     - 返回 -1: v1 < v2
     - 支持语义化版本比较

4. **版本号解析**
   - `ParseVersion(version string) (major, minor, patch int, err error)`
     - 支持格式: `"v1.2.3"`, `"1.2.3"`

### 4.2 构建配置

**wails.json 修改**:
```json
{
  "build:ldflags": "-X 'github.com/allanpk716/ai-commit-hub/pkg/version.Version={{.Version}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.CommitSHA={{.Commit}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.BuildTime={{.Date}}'"
}
```

---

## 5. 更新检查服务

### 5.1 服务定义

**文件**: `pkg/service/update_service.go`

**数据结构**:
```go
type UpdateService struct {
    repo           string        // GitHub 仓库
    httpClient     *http.Client  // HTTP 客户端
    currentVersion string        // 当前版本
}

type UpdateInfo struct {
    HasUpdate      bool          // 是否有更新
    LatestVersion  string        // 最新版本号
    CurrentVersion string        // 当前版本号
    ReleaseNotes   string        // Release notes
    PublishedAt    time.Time     // 发布时间
    DownloadURL    string        // 下载链接
    AssetName      string        // 资源文件名
    Size           int64         // 文件大小
}
```

**核心方法**:

1. **CheckForUpdates(ctx context.Context) (*UpdateInfo, error)**
   - 调用 GitHub API: `https://api.github.com/repos/{repo}/releases/latest`
   - 无需认证（使用公共 API）
   - 解析响应获取版本信息
   - 比较版本号判断是否需要更新

2. **平台识别**
   - 自动识别当前平台（windows/darwin）
   - 从 Release Assets 筛选对应平台的 zip 文件

3. **错误处理**
   - 网络错误: 返回友好提示
   - API 限流: 记录日志，建议稍后重试
   - 解析失败: 降级为检查 Git tags

### 5.2 App.go 集成

```go
// 在 startup(ctx) 中调用
func (a *App) startup(ctx context.Context) {
    // ... 现有代码

    go func() {
        updateInfo, err := a.updateService.CheckForUpdates(ctx)
        if err != nil {
            logger.Warnf("检查更新失败: %v", err)
            return
        }

        if updateInfo.HasUpdate {
            runtime.EventsEmit(ctx, "update-available", updateInfo)
        }
    }()
}
```

---

## 6. 下载器模块

### 6.1 模块结构

**文件**: `pkg/update/downloader.go`

**数据结构**:
```go
type Downloader struct {
    client      *http.Client
    downloadDir string           // 临时下载目录
    onProgress  ProgressFunc     // 进度回调
}

type ProgressFunc func(downloaded, total int64)

type DownloadProgress struct {
    Downloaded int64   // 已下载字节数
    Total      int64   // 总字节数
    Percentage float64 // 百分比
    Speed      int64   // 下载速度
}
```

**核心功能**:

1. **Download(url, filename string) (string, error)**
   - 创建临时文件
   - 流式下载
   - 定期调用进度回调
   - 验证文件大小
   - 返回本地文件路径

2. **进度报告**
   - 通过 Wails Events 发送 `download-progress` 事件
   - 前端实时更新进度条

3. **断点续传（可选）**
   - 支持 HTTP Range 请求
   - 下载中断后可恢复

4. **下载验证**
   - 检查 HTTP 状态码
   - 验证 Content-Length
   - 计算 SHA256 校验和

### 6.2 临时目录管理

- 使用 `os.TempDir()` 创建临时目录
- 文件名: `ai-commit-hub-update-{version}-{timestamp}.zip`
- 下载完成后保留文件，供更新器使用
- 更新完成后清理临时文件

---

## 7. 更新安装器

### 7.1 更新器架构

由于系统限制，程序无法直接替换正在运行的自身文件，因此使用**独立更新器进程**：

**文件**:
- `cmd/updater/main.go`: 独立更新器程序
- `pkg/update/installer.go`: 安装器接口

### 7.2 更新器程序

**功能**:
1. 接收命令行参数:
   ```bash
   updater.exe --source="path\to\update.zip" --target="path\to\app" --pid=1234
   ```

2. 执行步骤:
   - 等待主程序退出（通过 pid）
   - 解压 zip 文件到临时目录
   - 备份当前程序文件
   - 替换程序文件
   - 启动新版本程序
   - 清理临时文件和备份

### 7.3 更新流程

```
主程序下载完成
  ↓
调用 installer.Install()
  ↓
启动更新器进程（传递 pid）
  ↓
主程序调用 runtime.Quit() 退出
  ↓
更新器等待主程序退出
  ↓
更新器解压 zip
  ↓
更新器备份并替换文件
  ↓
更新器启动新版本主程序
  ↓
更新器退出
```

### 7.4 回滚机制

- 替换前备份原程序为 `.bak` 文件
- 新版本启动失败时，用户可手动恢复备份
- 下次启动时检测备份文件，自动回滚

### 7.5 Windows 特殊处理

- 使用 `MoveFileEx` with `MOVEFILE_DELAY_UNTIL_REBOOT` 作为最后手段
- 文件被占用时，标记为下次重启时替换

---

## 8. 前端 UI 设计

### 8.1 组件结构

1. **更新提示条** (`UpdateNotification.vue`)
   - 位置: CommitPanel 上方内联显示
   - 样式: 类似 Pushover 状态提示
   - 内容:
     - 图标 + 文字: "发现新版本 v1.0.0"
     - "查看详情"按钮
     - "忽略"按钮

2. **更新对话框** (`UpdateDialog.vue`)
   - 显示更新信息:
     - 当前版本 vs 最新版本
     - 发布时间
     - Release Notes（Markdown 渲染）
     - 文件大小
   - 操作按钮:
     - "立即更新"
     - "稍后提醒"
     - "跳过此版本"

3. **下载进度对话框** (`DownloadProgressDialog.vue`)
   - 实时显示下载进度:
     - 进度条（0-100%）
     - 已下载 / 总大小
     - 下载速度
     - 预计剩余时间
   - 按钮:
     - "后台下载"
     - "取消下载"

4. **重启确认对话框** (`RestartDialog.vue`)
   - 下载完成后显示:
     - "更新已下载完成"
     - "程序将关闭并自动安装更新"
   - 按钮:
     - "立即重启"
     - "稍后重启"

### 8.2 Store 集成

**文件**: `stores/updateStore.ts`

**状态管理**:
```typescript
interface UpdateState {
    hasUpdate: boolean
    updateInfo: UpdateInfo | null
    isDownloading: boolean
    downloadProgress: DownloadProgress
    isReadyToInstall: boolean
    skippedVersion: string | null
}
```

**方法**:
- `checkForUpdates()`: 检查更新
- `downloadUpdate()`: 开始下载
- `installUpdate()`: 安装更新
- `skipVersion(version)`: 跳过版本

### 8.3 事件监听

**App.vue 或 main.ts**:
```typescript
EventsOn('update-available', (data: UpdateInfo) => {
    // 显示更新提示条
})

EventsOn('download-progress', (data: DownloadProgress) => {
    // 更新进度条
})

EventsOn('download-complete', () => {
    // 显示重启对话框
})

EventsOn('download-error', (error: string) => {
    // 显示错误信息
})
```

### 8.4 API 绑定

**App.go 方法**:
```go
// CheckForUpdates 检查更新
func (a *App) CheckForUpdates() (*UpdateInfo, error)

// DownloadUpdate 下载更新
func (a *App) DownloadUpdate(downloadURL string) error

// InstallUpdate 安装更新
func (a *App) InstallUpdate() error
```

---

## 9. 用户偏好存储

### 9.1 数据模型

**文件**: `pkg/models/update_preferences.go`

```go
type UpdatePreferences struct {
    ID             uint      `gorm:"primaryKey"`
    SkippedVersion string    `gorm:"index"` // 用户跳过的版本号
    SkipReason     string    // 跳过原因
    CreatedAt      time.Time // 跳过时间
    LastCheckTime  time.Time // 最后检查更新的时间
    AutoCheck      bool      // 是否自动检查（默认 true）
}
```

### 9.2 存储逻辑

- 用户选择"稍后提醒": 不保存（本次会话不再提示）
- 用户选择"跳过此版本": 保存到数据库，下次检查时过滤
- 用户成功更新后: 清空 `SkippedVersion`

---

## 10. 错误处理

### 10.1 错误处理策略

1. **网络错误**
   - 检查更新失败: 静默失败，记录日志
   - 下载失败: 显示错误对话框，提供"重试"按钮
   - API 限流（403）: 提示用户稍后再试

2. **文件操作错误**
   - 磁盘空间不足: 下载前检查，提前提示用户
   - 解压失败: 保留下载的 zip 文件，提示用户手动安装
   - 文件替换失败: 使用 `MOVEFILE_DELAY_UNTIL_REBOOT` 延迟替换

3. **安装错误**
   - 更新器启动失败: 提示用户手动运行下载的安装包
   - 程序启动失败: 自动回滚到备份版本
   - 验证失败: 显示错误，保留旧版本

### 10.2 降级策略

- GitHub API 失败 → 检查 Git Tags
- Release 不可用 → 提示访问 GitHub Releases 页面手动下载
- 自动更新失败 → 降级为手动更新流程

### 10.3 日志记录

使用 `github.com/WQGroup/logger` 记录关键操作:
```go
logger.Info("检查更新", "current", currentVersion, "latest", latestVersion)
logger.Info("开始下载更新", "url", downloadURL, "size", fileSize)
logger.Infof("下载完成: %s (%d bytes)", filename, size)
logger.Warn("更新跳过", "version", skippedVersion)
logger.Errorf("更新失败: %v", err)
```

---

## 11. 测试策略

### 11.1 单元测试

1. **版本管理测试** (`pkg/version/version_test.go`)
   - 测试版本号解析
   - 测试版本比较逻辑
   - 测试边界情况

2. **更新服务测试** (`pkg/service/update_service_test.go`)
   - Mock GitHub API 响应
   - 测试版本比较逻辑
   - 测试平台识别

3. **下载器测试** (`pkg/update/downloader_test.go`)
   - Mock HTTP 响应
   - 测试进度回调
   - 测试下载中断恢复

### 11.2 集成测试

**文件**: `tests/integration/update_test.go`

1. **模拟 Release 测试**
   - 创建测试 GitHub 仓库
   - 发布测试 tag
   - 运行完整的更新流程

2. **端到端测试**
   - 启动应用
   - 检查更新
   - 下载更新
   - 验证更新器启动

### 11.3 前端测试

**文件**: `frontend/src/components/__tests__/UpdateDialog.spec.ts`

- 测试更新对话框渲染
- 测试下载进度显示
- 测试用户交互

### 11.4 手动测试流程

1. **首次发布测试**
   ```bash
   # 1. 创建测试 tag
   git tag v1.0.0
   git push origin v1.0.0

   # 2. 观察 GitHub Actions 构建
   # 3. 验证 Release 创建成功
   # 4. 下载并测试构建产物
   ```

2. **更新流程测试**
   - 运行旧版本程序
   - 创建新版本 tag
   - 测试下载和安装
   - 验证新版本正常运行

---

## 12. 实现阶段划分

### 阶段 1: 基础设施（优先级最高）

1. 创建版本管理模块（`pkg/version/`）
2. 配置 GitHub Actions 工作流（`.github/workflows/release.yml`）
3. 修改构建脚本，支持 ldflags 注入版本号
4. 测试第一个 Release 构建

### 阶段 2: 后端更新逻辑

1. 创建 UpdateService（`pkg/service/update_service.go`）
2. 实现下载器（`pkg/update/downloader.go`）
3. 创建更新器程序（`cmd/updater/main.go`）
4. 实现安装器接口（`pkg/update/installer.go`）
5. 在 App.go 中集成更新 API

### 阶段 3: 前端 UI

1. 创建 UpdateStore（`stores/updateStore.ts`）
2. 实现更新提示条组件（`UpdateNotification.vue`）
3. 创建更新对话框（`UpdateDialog.vue`）
4. 实现下载进度对话框（`DownloadProgressDialog.vue`）
5. 创建重启确认对话框（`RestartDialog.vue`）
6. 在 App.vue 中集成事件监听

### 阶段 4: 数据持久化

1. 创建 UpdatePreferences 模型
2. 实现数据库迁移
3. 添加跳过版本逻辑

### 阶段 5: 测试和优化

1. 编写单元测试
2. 编写集成测试
3. 手动端到端测试
4. 性能优化和错误处理完善

---

## 13. 技术要点

### 13.1 关键依赖

- `github.com/Masterminds/semver/v3`: 语义化版本比较
- 标准库 `archive/zip`: 解压更新文件
- Wails Events: 前后端通信

### 13.2 安全考虑

- 验证下载文件的完整性
- 更新器进程权限控制
- 防止中间人攻击（HTTPS）

### 13.3 性能优化

- 更新检查使用缓存（15 分钟内不重复检查）
- 下载使用流式传输，避免内存占用
- 使用并发下载（可选）

---

## 14. 用户体验设计

### 14.1 更新检查方式

- **被动检查**: 程序启动时在后台静默检查
- **非侵入式通知**: 内联提示条，不阻塞用户操作
- **用户控制**: 用户可选择何时更新

### 14.2 下载体验

- **后台下载**: 用户可继续使用程序
- **进度显示**: 实时显示下载进度和速度
- **可中断**: 支持取消下载

### 14.3 安装体验

- **自动安装**: 下载完成后自动安装
- **无缝重启**: 自动启动新版本
- **失败回滚**: 安装失败时自动回滚

---

## 15. 总结

本设计文档描述了一个完整的自动更新系统，包含以下特性：

✅ **完整的 CI/CD 流程**: 从 tag 发布到自动构建和上传
✅ **版本管理**: 灵活的版本注入和比较机制
✅ **自动更新**: 检查、下载、安装全流程自动化
✅ **用户体验**: 非侵入式通知、进度显示、平滑升级
✅ **容错机制**: 回滚、降级、错误恢复
✅ **跨平台支持**: Windows 和 macOS
✅ **可扩展性**: 支持未来添加更多平台和功能

该设计遵循 YAGNI 原则，避免过度设计，同时为未来的功能扩展预留了空间。
