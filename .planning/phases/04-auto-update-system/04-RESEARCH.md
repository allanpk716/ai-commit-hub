# Phase 04: Auto Update System - Research

**Researched:** 2026-02-07
**Domain:** Go Desktop Application Auto-Update (Wails v2 + GitHub Releases)
**Confidence:** HIGH

## Summary

Phase 04 需要实现一个完整的自动更新系统，包括版本检测、后台下载、外部更新器程序和自动重启。通过研究发现，Go 生态中有成熟的标准库和第三方库可以支持这些功能：

**核心发现：**
1. **版本比较** - 使用 `golang.org/x/mod/semver`（Go 官方维护）即可满足需求，无需引入额外依赖
2. **下载功能** - 标准库 `net/http` 配合 `io.Copy` 和 HTTP Range 请求可实现断点续传
3. **文件嵌入** - Go 1.16+ 内置的 `embed` 包可以将更新器可执行文件嵌入主程序
4. **事件通信** - Wails v2 的 `runtime.EventsEmit` 提供了流式进度推送机制
5. **更新器模式** - 外部更新器程序是 Windows 桌面应用的标准实践，避免文件锁定问题

**主要推荐：**
- 版本比较：`golang.org/x/mod/semver`（已在 Go 工具链中）
- 文件下载：标准库 `net/http` + 自定义进度跟踪
- ZIP 解压：标准库 `archive/zip`
- 文件嵌入：标准库 `embed`
- 事件通信：Wails 内置 Events 系统

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/mod/semver` | Go 1.24+ | 语义化版本比较（含预发布版本） | Go 官方维护，已包含在工具链中，完整支持 SemVer 2.0.0 规范 |
| `embed` (标准库) | Go 1.16+ | 嵌入更新器可执行文件到主程序 | Go 内置，无需额外依赖，跨平台支持 |
| `archive/zip` (标准库) | - | 解压 GitHub Release 的 ZIP 包 | Go 标准库，稳定可靠 |
| `net/http` (标准库) | - | HTTP 下载（支持 Range 请求实现断点续传） | Go 标准库，内置连接池和超时控制 |
| Wails Events | v2.11.0 | Go 后端到 Vue 前端的实时进度推送 | Wails 内置事件系统，已在使用中 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/cavaliercoder/grab` | v2.0.0 | 高级下载功能（需评估） | 如需复杂的断点续传、下载队列管理时考虑 |
| `github.com/mholt/archives` | v3.x | 统一压缩格式处理 | 如未来需要支持多种压缩格式（7z、rar 等） |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `golang.org/x/mod/semver` | `github.com/Masterminds/semver/v3` | Masterminds 提供更多特性（约束匹配），但本阶段只需要基础比较，官方库更轻量 |
| 标准库 `net/http` | `grab` 库 | grab 提供开箱即用的断点续传，但增加了依赖；标准库实现更透明，易于调试 |
| Wails Events | 轮询 API | Events 是推送模式，实时性更好；轮询实现简单但会增加不必要的网络请求 |

**Installation:**
```bash
# 核心库无需安装，已包含在 Go 1.24+ 中
# 如需使用 grab 库（可选）：
go get github.com/cavaliercoder/grab
```

## Architecture Patterns

### Recommended Project Structure
```
pkg/
├── update/
│   ├── checker.go       # 版本检测（调用 GitHub API）
│   ├── downloader.go    # 已存在，需增强断点续传
│   ├── installer.go     # 已存在，需改为嵌入更新器
│   ├── updater/         # 新增：外部更新器程序
│   │   └── main.go      # 独立的可执行文件
│   └── types.go         # 更新相关的类型定义
├── version/
│   └── version.go       # 已存在，需增强预发布版本支持
└── service/
    └── update_service.go # 已存在，需重构以使用新组件

frontend/src/
├── components/
│   ├── UpdateDialog.vue       # 已存在，需增强进度显示
│   ├── UpdateProgressDialog.vue # 新增：下载进度对话框
│   └── UpdateInstallerDialog.vue # 新增：安装确认对话框
└── stores/
    └── updateStore.ts         # 已存在，需增强状态管理
```

### Pattern 1: 语义化版本比较（含预发布版本）

**What:** 使用 `golang.org/x/mod/semver` 进行版本比较，完整支持 SemVer 2.0.0 规范，包括预发布标识符（alpha、beta、rc）。

**When to use:** 所有需要比较版本号的场景，包括：
- 检查是否有新版本
- 比较预发布版本（如 v1.0.0-beta.1 vs v1.0.0-alpha.2）
- 版本排序

**Example:**
```go
// Source: golang.org/x/mod/semver 官方文档
// https://pkg.go.dev/golang.org/x/mod/semver

import "golang.org/x/mod/semver"

// 比较版本（返回 -1, 0, 1）
func CompareVersions(v1, v2 string) int {
    // 确保版本号以 "v" 开头
    if !semver.IsValid(v1) || !semver.IsValid(v2) {
        return 0 // 无效版本视为相等
    }
    return semver.Compare(v1, v2)
}

// 检查是否为预发布版本
func IsPrerelease(version string) bool {
    return semver.Prerelease(version) != ""
}

// 解析预发布标识符
func GetPrerelease(version string) string {
    return semver.Prerelease(version) // 例如："-beta.1"
}

// 排序版本列表
func SortVersions(versions []string) {
    semver.Sort(versions)
}

// 实际使用示例
func main() {
    current := "v1.0.0"
    latest := "v1.0.1-beta.1"

    if semver.Compare(latest, current) > 0 {
        fmt.Println("发现新版本:", latest)
    }

    // 预发布版本比较
    beta1 := "v1.0.0-beta.1"
    beta2 := "v1.0.0-beta.2"
    alpha := "v1.0.0-alpha.1"
    stable := "v1.0.0"

    // alpha < beta.1 < beta.2 < stable
    fmt.Println(semver.Compare(alpha, beta1))  // -1
    fmt.Println(semver.Compare(beta1, beta2))  // -1
    fmt.Println(semver.Compare(beta2, stable)) // -1
}
```

**关键特性：**
- 自动处理 `v` 前缀（必须）
- 正确比较预发布版本：`1.0.0-alpha < 1.0.0-beta.1 < 1.0.0-rc.1 < 1.0.0`
- 支持构建元数据（如 `+build123`），但在比较时忽略

### Pattern 2: 嵌入更新器可执行文件

**What:** 使用 `embed` 包将独立的更新器程序（updater.exe）嵌入到主程序中，运行时释放到临时目录。

**When to use:**
- 需要打包外部工具或资源到单个二进制文件
- 更新器必须独立运行以避免文件锁定

**Example:**
```go
// Source: embed 包官方文档
// https://pkg.go.dev/embed

package main

import (
    "embed"
    "io"
    "os"
    "path/filepath"
)

//go:embed updater/updater.exe
var updaterBinary []byte

//go:embed updater/*
var updaterFS embed.FS

// ExtractUpdater 将嵌入的更新器释放到临时目录
func ExtractUpdater() (string, error) {
    // 创建临时文件
    tmpDir := os.TempDir()
    updaterPath := filepath.Join(tmpDir, "updater.exe")

    // 从嵌入的 []byte 写入文件
    if err := os.WriteFile(updaterPath, updaterBinary, 0755); err != nil {
        return "", err
    }

    return updaterPath, nil
}

// ExtractUpdaterFromFS 从 embed.FS 释放（适合多文件）
func ExtractUpdaterFromFS() (string, error) {
    tmpDir, err := os.MkdirTemp("", "updater-*")
    if err != nil {
        return "", err
    }

    // 读取所有文件并复制
    entries, err := updaterFS.ReadDir("updater")
    if err != nil {
        return "", err
    }

    for _, entry := range entries {
        src, err := updaterFS.Open("updater/" + entry.Name())
        if err != nil {
            return "", err
        }
        defer src.Close()

        dstPath := filepath.Join(tmpDir, entry.Name())
        dst, err := os.Create(dstPath)
        if err != nil {
            return "", err
        }
        defer dst.Close()

        if _, err := io.Copy(dst, src); err != nil {
            return "", err
        }
    }

    return tmpDir, nil
}
```

**实际应用示例（updater 嵌入）：**
```go
// 在主程序中
//go:embed build/bin/updater.exe
var embeddedUpdater []byte

func (i *Installer) extractUpdater() (string, error) {
    tmpDir := os.TempDir()
    updaterPath := filepath.Join(tmpDir, "ai-commit-hub-updater.exe")

    // 写入临时文件
    if err := os.WriteFile(updaterPath, embeddedUpdater, 0755); err != nil {
        return "", fmt.Errorf("写入更新器失败: %w", err)
    }

    return updaterPath, nil
}
```

### Pattern 3: HTTP 下载进度跟踪

**What:** 使用 `io.Reader` 包装器实时跟踪下载进度，并通过 Wails Events 推送到前端。

**When to use:**
- 需要显示下载进度、速度、剩余时间
- 大文件下载避免阻塞 UI

**Example:**
```go
// Source: 现有代码 + Wails Events 文档
// https://github.com/wailsapp/wails/blob/master/v2/pkg/assetserver/defaultindex.html

import (
    "io"
    "time"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProgressReader struct {
    reader      io.Reader
    total       int64
    downloaded  int64
    lastUpdate  time.Time
    ctx         context.Context
    url         string
}

func NewProgressReader(reader io.Reader, total int64, ctx context.Context, url string) *ProgressReader {
    return &ProgressReader{
        reader:     reader,
        total:      total,
        lastUpdate: time.Now(),
        ctx:        ctx,
        url:        url,
    }
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    if n > 0 {
        pr.downloaded += int64(n)
        pr.reportProgress()
    }
    return n, err
}

func (pr *ProgressReader) reportProgress() {
    now := time.Now()
    // 限制更新频率：每 100ms 最多一次
    if now.Sub(pr.lastUpdate) < 100*time.Millisecond {
        return
    }

    if pr.total > 0 {
        percentage := float64(pr.downloaded) / float64(pr.total) * 100

        // 通过 Wails Events 推送到前端
        runtime.EventsEmit(pr.ctx, "download-progress", map[string]interface{}{
            "percentage": percentage,
            "downloaded": pr.downloaded,
            "total":      pr.total,
            "url":        pr.url,
        })

        pr.lastUpdate = now
    }
}

// 使用示例
func DownloadWithProgress(ctx context.Context, url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 创建进度跟踪器
    progressReader := NewProgressReader(resp.Body, resp.ContentLength, ctx, url)

    // 复制到文件（自动触发进度更新）
    out, err := os.Create("output.zip")
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, progressReader)
    return err
}
```

### Pattern 4: 断点续传实现

**What:** 使用 HTTP Range 请求支持断点续传，记录已下载的字节位置。

**When to use:**
- 大文件下载（> 50MB）
- 网络不稳定环境

**Example:**
```go
// Source: HTTP Range Requests 规范
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests

import (
    "net/http"
    "os"
    "strconv"
)

type ResumableDownloader struct {
    client    *http.Client
    url       string
    destPath  string
}

func (rd *ResumableDownloader) Download() error {
    // 检查临时文件是否存在
    var downloaded int64 = 0
    if info, err := os.Stat(rd.destPath + ".tmp"); err == nil {
        downloaded = info.Size()
    }

    // 创建 Range 请求
    req, err := http.NewRequest("GET", rd.url, nil)
    if err != nil {
        return err
    }

    if downloaded > 0 {
        // 设置 Range 头：Range: bytes=1024-
        req.Header.Set("Range", "bytes="+strconv.FormatInt(downloaded, 10)+"-")
    }

    resp, err := rd.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 检查服务器是否支持 Range
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
        return fmt.Errorf("服务器不支持断点续传: %d", resp.StatusCode)
    }

    // 以追加模式打开文件
    flag := os.O_CREATE | os.O_WRONLY
    if downloaded > 0 {
        flag |= os.O_APPEND
    }
    file, err := os.OpenFile(rd.destPath+".tmp", flag, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // 继续下载
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return err
    }

    // 下载完成，重命名文件
    return os.Rename(rd.destPath+".tmp", rd.destPath)
}
```

### Pattern 5: 外部更新器通信

**What:** 主程序通过命令行参数传递信息给更新器，更新器完成文件替换后重启主程序。

**When to use:**
- Windows 桌面应用（避免文件锁定）
- 需要替换正在运行的可执行文件

**Example:**
```go
// 主程序 (app.go)
func (a *App) InstallUpdate(zipPath string) error {
    // 释放更新器
    updaterPath, err := a.installer.ExtractUpdater()
    if err != nil {
        return err
    }

    // 获取当前进程信息
    execPath, _ := os.Executable()
    pid := os.Getpid()

    // 启动更新器
    cmd := exec.Command(updaterPath,
        "--source", zipPath,
        "--target", filepath.Dir(execPath),
        "--pid", strconv.Itoa(pid),
        "--exec", execPath,
    )

    // Windows 下隐藏控制台
    if runtime.GOOS == "windows" {
        cmd.SysProcAttr = &windows.SysProcAttr{
            CreationFlags: 0x08000000, // CREATE_NO_WINDOW
        }
    }

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("启动更新器失败: %w", err)
    }

    // 主程序退出（释放文件锁）
    a.Quit()
    return nil
}
```

```go
// 更新器 (updater/main.go)
package main

import (
    "flag"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "time"
    "archive/zip"
)

var (
    source = flag.String("source", "", "更新包路径")
    target = flag.String("target", "", "安装目录")
    pid    = flag.Int("pid", 0, "主进程 PID")
    execPath = flag.String("exec", "", "主程序路径")
)

func main() {
    flag.Parse()

    // 1. 等待主程序退出
    if *pid > 0 {
        waitForProcessExit(*pid)
    }

    // 2. 解压 ZIP 到临时目录
    tmpDir, err := os.MkdirTemp("", "update-*")
    if err != nil {
        showError("创建临时目录失败", err)
        return
    }
    defer os.RemoveAll(tmpDir)

    if err := unzip(*source, tmpDir); err != nil {
        showError("解压失败", err)
        rollback()
        return
    }

    // 3. 备份旧版本
    backupDir := filepath.Join(*target, "backup")
    if err := os.MkdirAll(backupDir, 0755); err != nil {
        showError("创建备份目录失败", err)
        return
    }

    // 4. 替换文件
    if err := replaceFiles(tmpDir, *target, backupDir); err != nil {
        showError("文件替换失败", err)
        rollback()
        return
    }

    // 5. 启动新版本
    time.Sleep(2 * time.Second)
    exec.Command(*execPath).Start()
}

func waitForProcessExit(pid int) {
    for i := 0; i < 30; i++ { // 最多等待 30 秒
        process, err := os.FindProcess(pid)
        if err != nil {
            return // 进程已不存在
        }

        if err := process.Signal(os.Signal(nil)); err != nil {
            return // 进程已退出
        }

        time.Sleep(1 * time.Second)
    }
}

func unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        // 解压文件
        rc, err := f.Open()
        if err != nil {
            return err
        }

        path := filepath.Join(dest, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
            continue
        }

        out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }

        _, err = io.Copy(out, rc)
        out.Close()
        rc.Close()

        if err != nil {
            return err
        }
    }

    return nil
}
```

### Anti-Patterns to Avoid

- **同步阻塞下载**：不要在主 goroutine 中执行下载，会阻塞 UI。使用 goroutine + Events 推送进度。
- **直接替换可执行文件**：不要在程序运行时尝试替换自己的可执行文件（Windows 会锁定文件）。必须使用外部更新器。
- **硬编码版本比较逻辑**：不要自己实现版本比较逻辑，使用 `golang.org/x/mod/semver`。
- **忽略预发布版本**：不要只比较数字部分（如 1.0.0），忽略 `-beta` 等标识符会导致版本判断错误。
- **下载时不验证**：不要跳过文件大小和哈希验证，可能下载到损坏的文件。
- **不清理临时文件**：不要在临时目录留下大量下载残留，使用 `defer os.Remove` 清理。

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 版本比较 | 自己解析字符串和数字 | `golang.org/x/mod/semver` | 预发布版本规则复杂（alpha < beta < rc < stable），容易出错 |
| ZIP 解压 | 手动解析 ZIP 格式 | `archive/zip` | ZIP 格式有压缩、加密、多种编码等复杂情况 |
| 临时文件 | 使用固定路径（如 `/tmp/update.zip`） | `os.MkdirTemp` + `os.CreateTemp` | 避免并发冲突，跨平台路径问题 |
| HTTP 下载 | 手动管理连接池、超时 | `net/http` Client | 内置连接池、Keep-Alive、超时控制 |
| 进程等待 | 轮询 `ps` 命令 | `os.FindProcess` + `process.Signal` | 跨平台兼容，API 统一 |

**Key insight:** 虽然很多功能看起来简单（如下载文件、比较版本），但边缘情况（网络中断、预发布版本、文件权限）会消耗大量时间。使用标准库或成熟库可以避免这些问题。

## Common Pitfalls

### Pitfall 1: 版本号格式不一致

**What goes wrong:** 当前代码假设版本号格式为 `v1.2.3`，但 GitHub Release 可能返回 `1.2.3`（无 v 前缀）或 `v1.2.3-beta.1`（含预发布标识符），导致解析失败。

**Why it happens:** 不同工具和平台对版本号格式要求不同：
- `golang.org/x/mod/semver` **要求** `v` 前缀
- GitHub Release 的 `tag_name` 可能有或没有 `v` 前缀
- 当前代码的正则 `^(\d+)\.(\d+)\.(\d+)$` 不接受预发布标识符

**How to avoid:**
1. 统一版本号格式化：在比较前确保版本号以 `v` 开头
2. 使用 `semver.IsValid()` 验证版本号
3. 支持预发布版本：修改正则为 `^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`

**代码示例：**
```go
import "golang.org/x/mod/semver"

// 规范化版本号
func NormalizeVersion(version string) string {
    // 移除空格
    version = strings.TrimSpace(version)
    // 确保有 v 前缀
    if !strings.HasPrefix(version, "v") {
        version = "v" + version
    }
    return version
}

// 安全的版本比较
func SafeCompareVersions(v1, v2 string) int {
    v1 = NormalizeVersion(v1)
    v2 = NormalizeVersion(v2)

    if !semver.IsValid(v1) || !semver.IsValid(v2) {
        return 0 // 无效版本视为相等
    }

    return semver.Compare(v1, v2)
}
```

**Warning signs:**
- `ParseVersion` 返回 "invalid version format" 错误
- 版本比较结果不符合预期（如认为 beta 版本比稳定版本新）

### Pitfall 2: 文件下载不完整或损坏

**What goes wrong:** 下载过程中网络中断，导致文件不完整；或者下载了错误的文件（如 HTML 错误页面而非二进制文件）。

**Why it happens:**
- 没有验证 `Content-Length`
- 没有检查 HTTP 状态码（可能返回 404）
- 没有验证文件哈希

**How to avoid:**
1. 检查 HTTP 状态码必须是 200
2. 验证下载文件大小与 `Content-Length` 一致
3. 计算并验证 SHA256/MD5 哈希值
4. 在临时目录下载，完成后重命名

**代码示例：**
```go
func DownloadWithValidation(url, destPath string, expectedHash string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 1. 检查状态码
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
    }

    // 2. 创建临时文件
    tmpPath := destPath + ".tmp"
    file, err := os.Create(tmpPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // 3. 下载并计算哈希
    hasher := sha256.New()
    writer := io.MultiWriter(file, hasher)

    downloaded, err := io.Copy(writer, resp.Body)
    if err != nil {
        os.Remove(tmpPath)
        return err
    }

    // 4. 验证文件大小
    if resp.ContentLength > 0 && downloaded != resp.ContentLength {
        os.Remove(tmpPath)
        return fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", resp.ContentLength, downloaded)
    }

    // 5. 验证哈希
    actualHash := hex.EncodeToString(hasher.Sum(nil))
    if expectedHash != "" && actualHash != expectedHash {
        os.Remove(tmpPath)
        return fmt.Errorf("哈希验证失败: 期望 %s, 实际 %s", expectedHash, actualHash)
    }

    // 6. 重命名为正式文件
    return os.Rename(tmpPath, destPath)
}
```

**Warning signs:**
- 下载的文件无法运行或解压失败
- 文件大小明显小于预期
- 用户报告更新后程序崩溃

### Pitfall 3: Windows 文件锁定导致更新失败

**What goes wrong:** 更新器尝试替换正在运行的主程序可执行文件时失败，提示"文件正在使用"。

**Why it happens:** Windows 会锁定正在运行的可执行文件，无法直接覆盖或删除。

**How to avoid:**
1. 主程序启动更新器后立即退出
2. 更新器等待主程序完全退出（检查进程是否存在）
3. 使用独立进程而非共享内存

**代码示例：**
```go
// 主程序
func (a *App) StartUpdater() {
    cmd := exec.Command(updaterPath, "--pid", strconv.Itoa(os.Getpid()))
    cmd.Start()

    // 立即退出，释放文件锁
    os.Exit(0)
}

// 更新器
func waitForProcessExit(pid int) error {
    process, err := os.FindProcess(pid)
    if err != nil {
        return nil // 进程已不存在
    }

    // 检查进程是否还在运行
    for i := 0; i < 30; i++ {
        err = process.Signal(syscall.Signal(0)) // 不发送信号，仅检查进程是否存在
        if err != nil {
            return nil // 进程已退出
        }
        time.Sleep(1 * time.Second)
    }

    return fmt.Errorf("等待主程序退出超时")
}
```

**Warning signs:**
- 更新器报告"Access denied"或"文件正在使用"
- 更新后程序无法启动（文件损坏）

### Pitfall 4: GitHub API 速率限制

**What goes wrong:** 频繁检查更新导致 GitHub API 返回 403 Forbidden（速率限制）。

**Why it happens:** GitHub API 对未认证请求限制为每小时 60 次。

**How to avoid:**
1. 匿名访问时缓存检查结果，限制频率（如每 24 小时一次）
2. 使用 `/releases` 端点而非 `/releases/latest`，一次请求获取所有版本
3. 处理 403 错误，提示用户稍后重试

**代码示例：**
```go
const (
    checkInterval = 24 * time.Hour
    apiTimeout    = 10 * time.Second
)

type UpdateChecker struct {
    lastCheck     time.Time
    cachedResult  *models.UpdateInfo
    httpClient    *http.Client
}

func (c *UpdateChecker) CheckForUpdates() (*models.UpdateInfo, error) {
    // 检查缓存
    if time.Since(c.lastCheck) < checkInterval && c.cachedResult != nil {
        logger.Info("使用缓存的更新检查结果")
        return c.cachedResult, nil
    }

    // 调用 GitHub API
    result, err := c.fetchFromGitHub()
    if err != nil {
        // 如果是速率限制错误，返回缓存
        if isRateLimitError(err) && c.cachedResult != nil {
            logger.Warn("GitHub API 速率限制，使用缓存")
            return c.cachedResult, nil
        }
        return nil, err
    }

    c.lastCheck = time.Now()
    c.cachedResult = result
    return result, nil
}

func isRateLimitError(err error) bool {
    if urlErr, ok := err.(*url.Error); ok {
        if urlErr.Timeout() {
            return true
        }
    }
    // 检查 HTTP 403
    return strings.Contains(err.Error(), "403")
}
```

**Warning signs:**
- GitHub API 返回 403 错误
- 更新检查偶尔失败

### Pitfall 5: 更新器界面显示问题

**What goes wrong:** 更新器是控制台程序，用户看不到进度信息，以为程序卡死。

**Why it happens:** 更新器使用 `CREATE_NO_WINDOW` 标志隐藏控制台，但没有提供 GUI 反馈。

**How to avoid:**
1. 更新器显示独立进度窗口（Wails 或 Fyne）
2. 或者使用系统托盘通知
3. 记录详细日志到文件，供用户查看

**推荐方案：**
- 更新器使用简单的 GUI 库（如 Fyne 或 Walk）显示进度
- 或者使用 Windows 任务栏进度指示器（通过 COM 接口）

## Code Examples

### GitHub Releases API 调用

```go
// Source: GitHub 官方文档
// https://docs.github.com/en/rest/releases/releases

type GitHubRelease struct {
    TagName     string   `json:"tag_name"`
    Name        string   `json:"name"`
    Body        string   `json:"body"`
    Draft       bool     `json:"draft"`
    Prerelease  bool     `json:"prerelease"`
    PublishedAt string   `json:"published_at"`
    Assets      []Asset  `json:"assets"`
}

type Asset struct {
    Name               string `json:"name"`
    Size               int64  `json:"size"`
    BrowserDownloadURL string `json:"browser_download_url"`
}

// 获取所有 releases（包括预发布版本）
func (s *UpdateService) fetchAllReleases() ([]GitHubRelease, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/releases", s.repo)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // 匿名访问（不使用 Token）
    req.Header.Set("Accept", "application/vnd.github+json")

    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("GitHub API 错误: %d", resp.StatusCode)
    }

    var releases []GitHubRelease
    if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
        return nil, err
    }

    return releases, nil
}

// 查找最新版本（包括预发布版本）
func (s *UpdateService) findLatestVersion(releases []GitHubRelease) (*GitHubRelease, error) {
    if len(releases) == 0 {
        return nil, fmt.Errorf("没有找到任何 release")
    }

    // 过滤掉 draft releases
    var validReleases []GitHubRelease
    for _, r := range releases {
        if !r.Draft {
            validReleases = append(validReleases, r)
        }
    }

    if len(validReleases) == 0 {
        return nil, fmt.Errorf("没有找到非 draft 的 release")
    }

    // 按版本号排序（最新的在前）
    sort.Slice(validReleases, func(i, j int) bool {
        vi := NormalizeVersion(validReleases[i].TagName)
        vj := NormalizeVersion(validReleases[j].TagName)
        return semver.Compare(vi, vj) > 0
    })

    return &validReleases[0], nil
}

// 查找对应平台的资源
func (s *UpdateService) findPlatformAsset(assets []Asset) (Asset, error) {
    const targetPattern = "windows-amd64"

    for _, asset := range assets {
        // 精确匹配 windows-amd64
        if strings.Contains(asset.Name, targetPattern) {
            return asset, nil
        }
    }

    return Asset{}, fmt.Errorf("未找到 %s 平台的资源", targetPattern)
}
```

### 完整的下载流程（含进度推送）

```go
// Source: 整合现有代码 + Wails Events
func (a *App) DownloadUpdate(url string) error {
    // 发起下载请求
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
    }

    // 创建目标文件
    updatesDir := filepath.Join(getConfigDir(), "updates")
    os.MkdirAll(updatesDir, 0755)

    filename := filepath.Base(url)
    destPath := filepath.Join(updatesDir, filename)

    file, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // 创建进度跟踪器
    progressReader := &ProgressReader{
        reader:     resp.Body,
        total:      resp.ContentLength,
        ctx:        a.ctx,
        url:        url,
        lastUpdate: time.Now(),
    }

    // 开始下载
    runtime.EventsEmit(a.ctx, "download-started", map[string]interface{}{
        "url": url,
        "size": resp.ContentLength,
    })

    downloaded, err := io.Copy(file, progressReader)
    if err != nil {
        return err
    }

    // 验证文件大小
    if resp.ContentLength > 0 && downloaded != resp.ContentLength {
        return fmt.Errorf("文件大小不匹配")
    }

    // 下载完成
    runtime.EventsEmit(a.ctx, "download-complete", map[string]interface{}{
        "path": destPath,
        "size": downloaded,
    })

    return nil
}
```

### 前端进度显示

```typescript
// Source: 现有代码 + Wails Events
// frontend/src/stores/updateStore.ts

import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useUpdateStore = defineStore('update', () => {
  const downloadProgress = ref(0)
  const downloadSpeed = ref(0)
  const downloadedSize = ref(0)
  const totalSize = ref(0)

  // 监听下载进度
  EventsOn('download-progress', (data) => {
    downloadProgress.value = data.percentage
    downloadedSize.value = data.downloaded
    totalSize.value = data.total

    // 计算速度（需要保存上次更新时间和大小）
    updateSpeed(data.downloaded)
  })

  // 监听下载开始
  EventsOn('download-started', (data) => {
    totalSize.value = data.size
    downloadedSize.value = 0
    downloadProgress.value = 0
  })

  // 监听下载完成
  EventsOn('download-complete', (data) => {
    downloadProgress.value = 100
    isReadyToInstall.value = true
  })

  // 格式化文件大小
  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  // 计算剩余时间
  function calculateETA(downloaded: number, total: number, speed: number): string {
    if (speed === 0) return '--:--'
    const remaining = total - downloaded
    const seconds = remaining / speed
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 使用 `/releases/latest` 端点 | 使用 `/releases` 端点 | Phase 04 | 支持预发布版本检测 |
| 自定义版本解析 | `golang.org/x/mod/semver` | Phase 04 | 标准化版本比较，支持预发布标识符 |
| 简单 HTTP GET 下载 | HTTP Range 请求 + 进度跟踪 | Phase 04 | 支持断点续传，提升用户体验 |
| 内置更新逻辑 | 外部更新器程序 | Phase 04 | 避免 Windows 文件锁定问题 |
| 轮询检查进度 | Wails Events 推送 | 现有代码已使用 | 实时性好，减少不必要的请求 |

**Deprecated/outdated:**
- **自定义版本比较**：现有代码的手动解析容易出错，应使用 `semver` 库
- **简单包含匹配资源**：现有代码 `strings.Contains(asset.Name, "windows")` 不够精确，应精确匹配 `windows-amd64`

## Open Questions

1. **更新器界面实现方式**
   - What we know: 需要显示解压、替换等步骤的进度
   - What's unclear: 使用 Wails 还是其他 GUI 库（Fyne、Walk）
   - Recommendation: 使用 Fyne 或简单的控制台输出 + 日志文件，保持更新器轻量级

2. **校验和来源**
   - What we know: 需要验证下载文件的完整性
   - What's unclear: GitHub Release 页面如何提供 SHA256/MD5 校验和
   - Recommendation: 在 Release 中包含 `checksums.txt` 文件，或者在 Asset 的 `name` 中包含哈希值

3. **代理配置存储**
   - What we know: 用户需要在设置页面配置代理
   - What's unclear: 存储在数据库还是配置文件（YAML）
   - Recommendation: 存储在配置文件（`config.yaml`），与其他设置保持一致

## Sources

### Primary (HIGH confidence)
- `golang.org/x/mod/semver` - Go 官方语义化版本库
- `embed` (标准库) - Go 官方文件嵌入文档
- `archive/zip` (标准库) - Go 官方 ZIP 处理文档
- Wails Events - Wails 官方事件系统文档
- GitHub Releases API - GitHub 官方 REST API 文档

### Secondary (MEDIUM confidence)
- `github.com/cavaliercoder/grab` - Go 高级下载库（通过 pkg.go.dev）
- `github.com/mholt/archives` - 统一压缩格式库（通过 pkg.go.dev）
- Go by Example - 临时文件处理教程
- Transloadit - Go 断点续传实现指南

### Tertiary (LOW confidence)
- Go Samples - ZIP 解压示例（需验证）
- 各类博客文章 - 嵌入文件、断点续传实现

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - 所有核心库都是 Go 标准库或官方维护的包
- Architecture: HIGH - 基于 Wails v2 官方文档和现有代码结构
- Pitfalls: HIGH - 大部分基于现有代码的问题和 GitHub API 官方文档
- 版本比较: HIGH - `golang.org/x/mod/semver` 是 Go 官方推荐
- 下载实现: MEDIUM - 标准库足够，但断点续传需要验证
- 更新器模式: HIGH - 外部更新器是 Windows 应用的标准实践

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (30 days - Go 标准库和 Wails v2 稳定，但 GitHub API 可能变化)
