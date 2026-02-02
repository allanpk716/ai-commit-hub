# AI Commit Hub 自动更新功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 为 AI Commit Hub 实现完整的自动更新功能，包括 GitHub Actions 自动构建、版本管理、更新检查、下载和安装。

**架构：** 分层架构，后端使用 Go 实现版本管理和更新逻辑，前端使用 Vue3 实现 UI 组件，通过 Wails Events 进行通信。

**技术栈：** Go 1.21+, Wails v2, Vue 3, TypeScript, GitHub Actions, GitHub API

---

## 阶段 1: 基础设施 - 版本管理和 CI/CD

本阶段创建版本管理模块和 GitHub Actions 工作流，为后续功能奠定基础。

### Task 1.1: 创建版本管理模块

**文件：**
- 创建: `pkg/version/version.go`
- 创建: `pkg/version/version_test.go`

**Step 1: 编写版本解析测试**

在 `pkg/version/version_test.go` 中创建测试：

```go
package version

import (
    "testing"
)

func TestParseVersion(t *testing.T) {
    tests := []struct {
        name     string
        version  string
        wantMajor int
        wantMinor int
        wantPatch int
        wantErr   bool
    }{
        {
            name:      "标准版本号带v前缀",
            version:   "v1.2.3",
            wantMajor: 1,
            wantMinor: 2,
            wantPatch: 3,
            wantErr:   false,
        },
        {
            name:      "标准版本号不带v前缀",
            version:   "1.2.3",
            wantMajor: 1,
            wantMinor: 2,
            wantPatch: 3,
            wantErr:   false,
        },
        {
            name:      "空字符串",
            version:   "",
            wantMajor: 0,
            wantMinor: 0,
            wantPatch: 0,
            wantErr:   true,
        },
        {
            name:      "格式错误",
            version:   "invalid",
            wantMajor: 0,
            wantMinor: 0,
            wantPatch: 0,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            major, minor, patch, err := ParseVersion(tt.version)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
                t.Errorf("ParseVersion() = %v.%v.%v, want %v.%v.%v",
                    major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
            }
        })
    }
}
```

**Step 2: 运行测试验证失败**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `FAIL: undefined: ParseVersion`

**Step 3: 实现版本解析函数**

在 `pkg/version/version.go` 中实现：

```go
package version

import (
    "fmt"
    "regexp"
    "strconv"
)

var (
    // Version 当前版本号，编译时通过 ldflags 注入
    Version = "dev"

    // CommitSHA Git commit hash，编译时通过 ldflags 注入
    CommitSHA = "unknown"

    // BuildTime 构建时间，编译时通过 ldflags 注入
    BuildTime = "unknown"
)

// ParseVersion 解析版本号，支持格式: "v1.2.3" 或 "1.2.3"
// 返回 major, minor, patch 版本号
func ParseVersion(version string) (major, minor, patch int, err error) {
    // 移除 v 前缀
    version = regexp.MustCompile(`^v`).ReplaceAllString(version, "")

    // 解析版本号
    re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
    matches := re.FindStringSubmatch(version)

    if matches == nil {
        return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
    }

    major, err = strconv.Atoi(matches[1])
    if err != nil {
        return 0, 0, 0, err
    }

    minor, err = strconv.Atoi(matches[2])
    if err != nil {
        return 0, 0, 0, err
    }

    patch, err = strconv.Atoi(matches[3])
    if err != nil {
        return 0, 0, 0, err
    }

    return major, minor, patch, nil
}
```

**Step 4: 运行测试验证通过**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `PASS: TestParseVersion`

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add pkg/version/version.go pkg/version/version_test.go
git commit -m "feat(version): 添加版本号解析功能"
```

---

### Task 1.2: 实现版本比较功能

**文件：**
- 修改: `pkg/version/version.go`
- 修改: `pkg/version/version_test.go`

**Step 1: 编写版本比较测试**

在 `pkg/version/version_test.go` 中添加：

```go
func TestCompareVersions(t *testing.T) {
    tests := []struct {
        name     string
        v1       string
        v2       string
        expected int
    }{
        {"v1大于v2", "v1.2.3", "v1.2.2", 1},
        {"v1小于v2", "v1.2.2", "v1.2.3", -1},
        {"v1等于v2", "v1.2.3", "v1.2.3", 0},
        {"主版本不同", "v2.0.0", "v1.9.9", 1},
        {"次版本不同", "v1.3.0", "v1.2.9", 1},
        {"修订版本不同", "v1.2.4", "v1.2.3", 1},
        {"不带v前缀", "1.2.3", "1.2.2", 1},
        {"混合前缀", "v1.2.3", "1.2.2", 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CompareVersions(tt.v1, tt.v2)
            if result != tt.expected {
                t.Errorf("CompareVersions(%s, %s) = %d, want %d",
                    tt.v1, tt.v2, result, tt.expected)
            }
        })
    }
}
```

**Step 2: 运行测试验证失败**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `FAIL: undefined: CompareVersions`

**Step 3: 实现版本比较函数**

在 `pkg/version/version.go` 中添加：

```go
// CompareVersions 比较两个版本号
// 返回: 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
func CompareVersions(v1, v2 string) int {
    major1, minor1, patch1, err1 := ParseVersion(v1)
    major2, minor2, patch2, err2 := ParseVersion(v2)

    // 如果解析失败，视为相等
    if err1 != nil || err2 != nil {
        return 0
    }

    if major1 != major2 {
        if major1 > major2 {
            return 1
        }
        return -1
    }

    if minor1 != minor2 {
        if minor1 > minor2 {
            return 1
        }
        return -1
    }

    if patch1 != patch2 {
        if patch1 > patch2 {
            return 1
        }
        return -1
    }

    return 0
}
```

**Step 4: 运行测试验证通过**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `PASS: TestCompareVersions`

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add pkg/version/
git commit -m "feat(version): 添加版本号比较功能"
```

---

### Task 1.3: 实现版本获取功能

**文件：**
- 修改: `pkg/version/version.go`
- 修改: `pkg/version/version_test.go`

**Step 1: 编写版本获取测试**

在 `pkg/version/version_test.go` 中添加：

```go
func TestGetVersion(t *testing.T) {
    // 保存原始值
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name     string
        version  string
        expected string
    }{
        {"开发版本", "dev", "dev-uncommitted"},
        {"生产版本", "1.0.0", "v1.0.0"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            result := GetVersion()
            if result != tt.expected {
                t.Errorf("GetVersion() = %s, want %s", result, tt.expected)
            }
        })
    }
}
```

**Step 2: 运行测试验证失败**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `FAIL: undefined: GetVersion`

**Step 3: 实现版本获取函数**

在 `pkg/version/version.go` 中添加：

```go
// GetVersion 获取当前版本号
// 开发模式返回 "dev-uncommitted"
// 生产模式返回 "v{major}.{minor}.{patch}"
func GetVersion() string {
    if Version == "dev" {
        return "dev-uncommitted"
    }
    return "v" + Version
}

// GetFullVersion 获取完整版本信息
// 格式: "v1.0.0 (abc1234 2024-01-15)"
func GetFullVersion() string {
    if Version == "dev" {
        return "dev-uncommitted"
    }
    return fmt.Sprintf("v%s (%s %s)", Version, CommitSHA, BuildTime)
}

// IsDevVersion 判断是否为开发版本
func IsDevVersion() bool {
    return Version == "dev"
}
```

**Step 4: 运行测试验证通过**

运行: `cd .worktrees/auto-update && go test ./pkg/version -v`

预期输出: `PASS: TestGetVersion`

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add pkg/version/
git commit -m "feat(version): 添加版本获取功能"
```

---

### Task 1.4: 创建 GitHub Actions 工作流

**文件：**
- 创建: `.github/workflows/release.yml`

**Step 1: 创建工作流配置文件**

创建 `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  release:
    strategy:
      matrix:
        os: [windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install Wails CLI
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Extract version from tag
        id: get_version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT
        shell: bash

      - name: Download Go dependencies
        run: go mod download

      - name: Download frontend dependencies
        run: |
          cd frontend
          npm install

      - name: Build with Wails
        run: |
          wails build -clean -ldflags "-X main.version=${{ steps.get_version.outputs.VERSION }}"
        env:
          CGO_ENABLED: '1'

      - name: Package (Windows)
        if: matrix.os == 'windows-latest'
        run: |
          cd build/bin
          7z a ../ai-commit-hub-${{ steps.get_version.outputs.VERSION }}-windows.zip ai-commit-hub.exe

      - name: Package (macOS)
        if: matrix.os == 'macos-latest'
        run: |
          cd build/bin
          zip -r ../ai-commit-hub-${{ steps.get_version.outputs.VERSION }}-darwin.zip "AI Commit Hub.app"

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            build/ai-commit-hub-${{ steps.get_version.outputs.VERSION }}-*.zip
          generate_release_notes: true
          name: v${{ steps.get_version.outputs.VERSION }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 2: 验证工作流文件语法**

运行: `cd .worktrees/auto-update && cat .github/workflows/release.yml`

预期输出: 工作流 YAML 文件内容

**Step 3: 提交**

```bash
cd .worktrees/auto-update
git add .github/workflows/release.yml
git commit -m "ci: 添加 GitHub Actions 自动发布工作流"
```

---

### Task 1.5: 修改构建脚本支持版本注入

**文件：**
- 修改: `wails.json`
- 修改: `main.go`

**Step 1: 修改 wails.json 添加 ldflags**

在 `wails.json` 中添加构建参数：

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "ai-commit-hub",
  "outputfilename": "ai-commit-hub",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "allanpk716",
    "email": "allanpk716@example.com"
  },
  "info": {
    "companyName": "AI Commit Hub",
    "productName": "AI Commit Hub",
    "productVersion": "1.0.0",
    "copyright": "Copyright........",
    "comments": "AI-powered Git commit message generator"
  },
  "build:ldflags": "-X 'github.com/allanpk716/ai-commit-hub/pkg/version.Version={{.Version}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.CommitSHA={{.Commit}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.BuildTime={{.Date}}'"
}
```

**Step 2: 修改 main.go 导入版本模块**

在 `main.go` 顶部添加导入（import）：

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    stdruntime "runtime"
    "time"

    "github.com/WQGroup/logger"
    "github.com/allanpk716/ai-commit-hub/pkg/git"
    "github.com/allanpk716/ai-commit-hub/pkg/models"
    "github.com/allanpk716/ai-commit-hub/pkg/pushover"
    "github.com/allanpk716/ai-commit-hub/pkg/repository"
    "github.com/allanpk716/ai-commit-hub/pkg/service"
    "github.com/allanpk716/ai-commit-hub/pkg/version"  // 添加这行
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "golang.org/x/sys/windows"
    "gorm.io/gorm"

    // Provider 注册 - 匿名导入以触发 init()
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/anthropic"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/deepseek"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/google"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/ollama"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/openai"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/openrouter"
    _ "github.com/allanpk716/ai-commit-hub/pkg/provider/phind"
)

// 其余代码保持不变...
```

**Step 3: 在 App 启动时打印版本信息**

在 `main.go` 的 `main()` 函数中添加版本日志：

```go
func main() {
    // 添加版本信息日志
    logger.Info("AI Commit Hub starting up...", "version", version.GetVersion())
    logger.Debug("Full version info", "info", version.GetFullVersion())

    // 原有代码...
    app := NewApp()

    err := wails.Run(&options.App{
        // ... 其他配置
    })

    if err != nil {
        logger.Fatalf("Error: %s", err.Error())
    }
}
```

**Step 4: 测试构建**

运行: `cd .worktrees/auto-update && wails build -clean -ldflags "-X main.version=1.0.0-test"`

预期输出: 构建成功，可执行文件在 `build/bin/` 目录

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add wails.json main.go
git commit -m "build: 支持版本号注入到可执行文件"
```

---

### Task 1.6: 验证端到端构建流程

**文件：**
- 无

**Step 1: 创建测试 tag**

```bash
cd .worktrees/auto-update
git tag v1.0.0-test
git push origin v1.0.0-test
```

**Step 2: 观察 GitHub Actions 运行**

访问: `https://github.com/allanpk716/ai-commit-hub/actions`

预期输出: 应该看到新的 Release workflow 在运行

**Step 3: 验证 Release 创建成功**

访问: `https://github.com/allanpk716/ai-commit-hub/releases`

预期输出: 应该看到 v1.0.0-test release，包含 zip 文件

**Step 4: 清理测试 tag（可选）**

```bash
cd .worktrees/auto-update
git tag -d v1.0.0-test
git push origin :refs/tags/v1.0.0-test
```

**Step 5: 提交阶段总结**

```bash
cd .worktrees/auto-update
git add .
git commit -m "docs: 完成阶段1 - 基础设施和 CI/CD"
```

---

## 阶段 2: 后端更新逻辑

本阶段实现更新检查、下载和安装逻辑。

### Task 2.1: 创建更新数据模型

**文件：**
- 创建: `pkg/models/update_info.go`
- 创建: `pkg/models/update_preferences.go`

**Step 1: 定义 UpdateInfo 结构**

创建 `pkg/models/update_info.go`:

```go
package models

import "time"

// UpdateInfo 更新信息
type UpdateInfo struct {
    HasUpdate      bool      `json:"hasUpdate"`      // 是否有更新
    LatestVersion  string    `json:"latestVersion"`  // 最新版本号
    CurrentVersion string    `json:"currentVersion"` // 当前版本号
    ReleaseNotes   string    `json:"releaseNotes"`   // Release notes
    PublishedAt    time.Time `json:"publishedAt"`    // 发布时间
    DownloadURL    string    `json:"downloadURL"`    // 下载链接
    AssetName      string    `json:"assetName"`      // 资源文件名
    Size           int64     `json:"size"`           // 文件大小
}
```

**Step 2: 定义 UpdatePreferences 结构**

创建 `pkg/models/update_preferences.go`:

```go
package models

import "time"

// UpdatePreferences 用户更新偏好设置
type UpdatePreferences struct {
    ID             uint      `gorm:"primaryKey" json:"id"`
    SkippedVersion string    `gorm:"index" json:"skippedVersion"` // 用户跳过的版本号
    SkipReason     string    `json:"skipReason"`                  // 跳过原因 (not_now/this_version)
    CreatedAt      time.Time `json:"createdAt"`                   // 跳过时间
    LastCheckTime  time.Time `json:"lastCheckTime"`               // 最后检查更新的时间
    AutoCheck      bool      `json:"autoCheck"`                   // 是否自动检查（默认 true）
}
```

**Step 3: 在数据库初始化时自动迁移**

修改 `pkg/repository/db.go`，在 `AutoMigrate` 中添加 `UpdatePreferences`:

```go
func InitDB(dbPath string) (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    err = db.AutoMigrate(
        &models.GitProject{},
        &models.CommitHistory{},
        &models.UpdatePreferences{},  // 添加这行
    )
    // ...
}
```

**Step 4: 运行测试验证数据库迁移**

运行: `cd .worktrees/auto-update && go test ./pkg/repository -v -run TestInitDB`

预期输出: 数据库初始化成功

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add pkg/models/update_info.go pkg/models/update_preferences.go pkg/repository/db.go
git commit -m "feat(models): 添加更新信息数据模型"
```

---

### Task 2.2: 创建更新检查服务

**文件：**
- 创建: `pkg/service/update_service.go`
- 创建: `pkg/service/update_service_test.go`

**Step 1: 编写更新检查测试**

创建 `pkg/service/update_service_test.go`:

```go
package service

import (
    "testing"
    "github.com/allanpk716/ai-commit-hub/pkg/models"
)

func TestCheckForUpdates(t *testing.T) {
    service := NewUpdateService("allanpk716/ai-commit-hub")

    info, err := service.CheckForUpdates()
    if err != nil {
        t.Logf("CheckForUpdates failed (expected in CI): %v", err)
        return
    }

    t.Logf("Update info: %+v", info)
    if info == nil {
        t.Error("Expected non-nil update info")
    }

    if info.CurrentVersion == "" {
        t.Error("Expected current version to be set")
    }
}
```

**Step 2: 运行测试验证失败**

运行: `cd .worktrees/auto-update && go test ./pkg/service -v -run TestCheckForUpdates`

预期输出: `FAIL: undefined: NewUpdateService`

**Step 3: 实现更新检查服务**

创建 `pkg/service/update_service.go`:

```go
package service

import (
    "encoding/json"
    "fmt"
    "net/http"
    "stdruntime"
    "strings"
    "time"

    "github.com/WQGroup/logger"
    "github.com/allanpk716/ai-commit-hub/pkg/models"
    "github.com/allanpk716/ai-commit-hub/pkg/version"
)

// UpdateService 更新检查服务
type UpdateService struct {
    repo       string
    httpClient *http.Client
}

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
    TagName   string `json:"tag_name"`
    Name      string `json:"name"`
    Body      string `json:"body"`
    Draft     bool   `json:"draft"`
    Prerelease bool  `json:"prerelease"`
    PublishedAt string `json:"published_at"`
    Assets    []Asset `json:"assets"`
}

// Asset Release 资源
type Asset struct {
    Name  string `json:"name"`
    Size  int64  `json:"size"`
    URL   string `json:"browser_download_url"`
}

// NewUpdateService 创建更新检查服务
func NewUpdateService(repo string) *UpdateService {
    return &UpdateService{
        repo: repo,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// CheckForUpdates 检查更新
func (s *UpdateService) CheckForUpdates() (*models.UpdateInfo, error) {
    logger.Info("检查更新", "repo", s.repo)

    // 获取最新 Release
    release, err := s.fetchLatestRelease()
    if err != nil {
        logger.Warnf("获取 Release 失败: %v", err)
        return nil, err
    }

    currentVersion := version.GetVersion()
    latestVersion := release.TagName

    logger.Info("版本信息", "current", currentVersion, "latest", latestVersion)

    // 比较版本
    hasUpdate := s.compareVersions(latestVersion, currentVersion)

    // 找到对应平台的资源
    assetName, downloadURL := s.findPlatformAsset(release.Assets)

    info := &models.UpdateInfo{
        HasUpdate:      hasUpdate,
        LatestVersion:  latestVersion,
        CurrentVersion: currentVersion,
        ReleaseNotes:   release.Body,
        PublishedAt:    s.parseTime(release.PublishedAt),
        DownloadURL:    downloadURL,
        AssetName:      assetName,
        Size:           s.getAssetSize(release.Assets, assetName),
    }

    logger.Infof("更新检查完成: hasUpdate=%v", info.HasUpdate)
    return info, nil
}

// fetchLatestRelease 获取最新 Release
func (s *UpdateService) fetchLatestRelease() (*GitHubRelease, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", s.repo)

    resp, err := s.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("GitHub API 返回错误: %d", resp.StatusCode)
    }

    var release GitHubRelease
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return nil, fmt.Errorf("解析 JSON 失败: %w", err)
    }

    return &release, nil
}

// compareVersions 比较版本号
func (s *UpdateService) compareVersions(latest, current string) bool {
    result := version.CompareVersions(latest, current)
    return result > 0
}

// findPlatformAsset 找到对应平台的资源
func (s *UpdateService) findPlatformAsset(assets []Asset) (name, url string) {
    os := stdruntime.GOOS
    arch := stdruntime.GOARCH

    var targetOS string
    if os == "windows" {
        targetOS = "windows"
    } else if os == "darwin" {
        targetOS = "darwin"
    } else {
        targetOS = os
    }

    for _, asset := range assets {
        if strings.Contains(asset.Name, targetOS) {
            return asset.Name, asset.URL
        }
    }

    return "", ""
}

// getAssetSize 获取资源大小
func (s *UpdateService) getAssetSize(assets []Asset, assetName string) int64 {
    for _, asset := range assets {
        if asset.Name == assetName {
            return asset.Size
        }
    }
    return 0
}

// parseTime 解析时间
func (s *UpdateService) parseTime(timeStr string) time.Time {
    t, err := time.Parse(time.RFC3339, timeStr)
    if err != nil {
        logger.Warnf("解析时间失败: %v", err)
        return time.Time{}
    }
    return t
}
```

**Step 4: 运行测试验证通过**

运行: `cd .worktrees/auto-update && go test ./pkg/service -v -run TestCheckForUpdates`

预期输出: `PASS: TestCheckForUpdates`

**Step 5: 提交**

```bash
cd .worktrees/auto-update
git add pkg/service/update_service.go pkg/service/update_service_test.go
git commit -m "feat(service): 添加更新检查服务"
```

---

### Task 2.3: 在 App.go 中集成更新检查

**文件：**
- 修改: `app.go`

**Step 1: 添加 UpdateService 到 App 结构**

在 `app.go` 的 `App` 结构体中添加：

```go
type App struct {
    ctx                  context.Context
    dbPath               string
    db                   *gorm.DB
    gitProjectRepo       *repository.GitProjectRepository
    commitHistoryRepo    *repository.CommitHistoryRepository
    pushoverService      *pushover.Service
    startupService       *service.StartupService
    updateService        *service.UpdateService  // 添加这行
}
```

**Step 2: 在 NewApp 中初始化 UpdateService**

在 `app.go` 的 `NewApp()` 函数中添加：

```go
func NewApp() *App {
    // ... 现有代码 ...

    app := &App{
        db:                  db,
        gitProjectRepo:      gitProjectRepo,
        commitHistoryRepo:   commitHistoryRepo,
        pushoverService:     pushoverService,
        startupService:      startupService,
        updateService:       service.NewUpdateService("allanpk716/ai-commit-hub"),  // 添加这行
    }

    return app
}
```

**Step 3: 在 startup 中检查更新**

在 `app.go` 的 `startup()` 函数中添加更新检查：

```go
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    logger.Info("AI Commit Hub starting up...")

    // ... 现有的初始化代码 ...

    // 异步检查更新
    go func() {
        updateInfo, err := a.updateService.CheckForUpdates()
        if err != nil {
            logger.Warnf("检查更新失败: %v", err)
            return
        }

        if updateInfo.HasUpdate {
            logger.Info("发现新版本", "version", updateInfo.LatestVersion)
            runtime.EventsEmit(ctx, "update-available", updateInfo)
        } else {
            logger.Info("已是最新版本")
        }
    }()
}
```

**Step 4: 添加导出的 API 方法**

在 `app.go` 中添加导出的方法：

```go
// CheckForUpdates 手动检查更新
func (a *App) CheckForUpdates() (*models.UpdateInfo, error) {
    return a.updateService.CheckForUpdates()
}
```

**Step 5: 测试更新检查**

运行: `cd .worktrees/auto-update && wails dev`

预期输出: 在控制台日志中看到"检查更新"和"已是最新版本"或"发现新版本"

**Step 6: 提交**

```bash
cd .worktrees/auto-update
git add app.go
git commit -m "feat(app): 集成更新检查功能"
```

---

## 阶段 3: 前端 UI 组件

本阶段创建前端 UI 组件，用于显示更新通知和下载进度。

### Task 3.1: 创建 UpdateStore

**文件：**
- 创建: `frontend/src/stores/updateStore.ts`

**Step 1: 创建 UpdateStore**

创建 `frontend/src/stores/updateStore.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { EventsOn } from '../wailsjs/runtime/runtime'
import type { models } from '../wailsjs/go/models'

export const useUpdateStore = defineStore('update', () => {
  // State
  const hasUpdate = ref(false)
  const updateInfo = ref<models.UpdateInfo | null>(null)
  const isChecking = ref(false)
  const isDownloading = ref(false)
  const downloadProgress = ref(0)
  const downloadSpeed = ref(0)
  const isReadyToInstall = ref(false)
  const skippedVersion = ref<string | null>(null)

  // Computed
  const displayVersion = computed(() => {
    return updateInfo.value?.latestVersion || ''
  })

  const releaseNotes = computed(() => {
    return updateInfo.value?.releaseNotes || ''
  })

  // Actions
  async function checkForUpdates() {
    isChecking.value = true
    try {
      // 这里调用后端 API
      const info = await window.go.main.App.CheckForUpdates()
      updateInfo.value = info
      hasUpdate.value = info.hasUpdate
      return info
    } catch (error) {
      console.error('检查更新失败:', error)
      throw error
    } finally {
      isChecking.value = false
    }
  }

  function skipVersion(version: string) {
    skippedVersion.value = version
    hasUpdate.value = false
  }

  function resetUpdateState() {
    hasUpdate.value = false
    updateInfo.value = null
    isDownloading.value = false
    downloadProgress.value = 0
    isReadyToInstall.value = false
  }

  // 监听后端事件
  EventsOn('update-available', (info: models.UpdateInfo) => {
    console.log('收到更新可用事件:', info)
    updateInfo.value = info
    hasUpdate.value = info.hasUpdate
  })

  EventsOn('download-progress', (progress: { percentage: number; speed: number }) => {
    downloadProgress.value = progress.percentage
    downloadSpeed.value = progress.speed
    isDownloading.value = true
  })

  EventsOn('download-complete', () => {
    isDownloading.value = false
    isReadyToInstall.value = true
  })

  return {
    hasUpdate,
    updateInfo,
    isChecking,
    isDownloading,
    downloadProgress,
    downloadSpeed,
    isReadyToInstall,
    skippedVersion,
    displayVersion,
    releaseNotes,
    checkForUpdates,
    skipVersion,
    resetUpdateState
  }
})
```

**Step 2: 验证 TypeScript 编译**

运行: `cd .worktrees/auto-update/frontend && npm run build`

预期输出: 编译成功，无类型错误

**Step 3: 提交**

```bash
cd .worktrees/auto-update
git add frontend/src/stores/updateStore.ts
git commit -m "feat(frontend): 创建 UpdateStore"
```

---

### Task 3.2: 创建更新通知组件

**文件：**
- 创建: `frontend/src/components/UpdateNotification.vue`

**Step 1: 创建更新通知组件**

创建 `frontend/src/components/UpdateNotification.vue`:

```vue
<template>
  <div v-if="updateStore.hasUpdate" class="update-notification">
    <div class="notification-content">
      <div class="notification-icon">🔄</div>
      <div class="notification-text">
        发现新版本 {{ updateStore.displayVersion }}
      </div>
      <div class="notification-actions">
        <button @click="showDetails" class="btn-primary">查看详情</button>
        <button @click="ignore" class="btn-secondary">忽略</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUpdateStore } from '../stores/updateStore'

const updateStore = useUpdateStore()

function showDetails() {
  // 发射事件到父组件，显示更新对话框
  emit('show-update-dialog')
}

function ignore() {
  updateStore.skipVersion(updateStore.updateInfo?.latestVersion || '')
}

const emit = defineEmits(['show-update-dialog'])
</script>

<style scoped>
.update-notification {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.notification-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-icon {
  font-size: 24px;
}

.notification-text {
  flex: 1;
  font-weight: 500;
}

.notification-actions {
  display: flex;
  gap: 8px;
}

.btn-primary, .btn-secondary {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-primary {
  background: white;
  color: #667eea;
  font-weight: 600;
}

.btn-primary:hover {
  background: #f0f0f0;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
```

**Step 2: 在 CommitPanel 中使用更新通知**

修改 `frontend/src/components/CommitPanel.vue`，在顶部添加更新通知：

```vue
<template>
  <div class="commit-panel">
    <UpdateNotification @show-update-dialog="showUpdateDialog = true" />

    <!-- 现有的其他内容 -->
  </div>
</template>

<script setup lang="ts">
import UpdateNotification from './UpdateNotification.vue'
import { ref } from 'vue'

const showUpdateDialog = ref(false)
</script>
```

**Step 3: 验证 UI 显示**

运行: `cd .worktrees/auto-update && wails dev`

预期输出: 如果有更新，应该能在 CommitPanel 顶部看到更新通知

**Step 4: 提交**

```bash
cd .worktrees/auto-update
git add frontend/src/components/UpdateNotification.vue frontend/src/components/CommitPanel.vue
git commit -m "feat(frontend): 添加更新通知组件"
```

---

### Task 3.3: 创建更新详情对话框

**文件：**
- 创建: `frontend/src/components/UpdateDialog.vue`

**Step 1: 创建更新详情对话框**

创建 `frontend/src/components/UpdateDialog.vue`:

```vue
<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="update-dialog">
      <div class="dialog-header">
        <h2>发现新版本</h2>
        <button @click="close" class="close-btn">&times;</button>
      </div>

      <div class="dialog-body">
        <div class="version-comparison">
          <div class="version-item">
            <span class="label">当前版本:</span>
            <span class="value">{{ updateInfo?.currentVersion }}</span>
          </div>
          <div class="version-item">
            <span class="label">最新版本:</span>
            <span class="value highlight">{{ updateInfo?.latestVersion }}</span>
          </div>
        </div>

        <div class="release-notes">
          <h3>更新内容</h3>
          <div class="notes-content" v-html="formattedReleaseNotes"></div>
        </div>

        <div class="file-info">
          <span>文件大小: {{ formatSize(updateInfo?.size || 0) }}</span>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="download" class="btn-download" :disabled="isDownloading">
          {{ isDownloading ? '下载中...' : '立即更新' }}
        </button>
        <button @click="skip" class="btn-skip">跳过此版本</button>
        <button @click="close" class="btn-cancel">稍后提醒</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useUpdateStore } from '../stores/updateStore'
import { marked } from 'marked'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits(['close'])

const updateStore = useUpdateStore()
const isDownloading = ref(false)

const updateInfo = computed(() => updateStore.updateInfo)

const formattedReleaseNotes = computed(() => {
  if (!updateInfo.value?.releaseNotes) return ''
  return marked(updateInfo.value.releaseNotes)
})

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

function close() {
  emit('close')
}

function skip() {
  updateStore.skipVersion(updateInfo.value?.latestVersion || '')
  close()
}

async function download() {
  isDownloading.value = true
  try {
    // TODO: 调用下载 API
    console.log('开始下载更新')
  } catch (error) {
    console.error('下载失败:', error)
    isDownloading.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.update-dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 32px;
  cursor: pointer;
  color: #6b7280;
  line-height: 1;
}

.close-btn:hover {
  color: #1f2937;
}

.dialog-body {
  padding: 24px;
}

.version-comparison {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version-item .label {
  color: #6b7280;
  font-weight: 500;
}

.version-item .value {
  font-weight: 600;
  color: #1f2937;
}

.version-item .value.highlight {
  color: #667eea;
  font-size: 18px;
}

.release-notes {
  margin-bottom: 24px;
}

.release-notes h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
}

.notes-content {
  color: #374151;
  line-height: 1.6;
}

.notes-content :deep(h1),
.notes-content :deep(h2),
.notes-content :deep(h3) {
  margin-top: 16px;
  margin-bottom: 8px;
}

.notes-content :deep(ul),
.notes-content :deep(ol) {
  margin: 8px 0;
  padding-left: 24px;
}

.notes-content :deep(li) {
  margin: 4px 0;
}

.notes-content :deep(code) {
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 14px;
}

.file-info {
  color: #6b7280;
  font-size: 14px;
  text-align: center;
}

.dialog-footer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn-download,
.btn-skip,
.btn-cancel {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-download {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-download:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-download:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-skip,
.btn-cancel {
  background: white;
  color: #6b7280;
  border: 1px solid #d1d5db;
}

.btn-skip:hover,
.btn-cancel:hover {
  background: #f9fafb;
}
</style>
```

**Step 2: 安装 Markdown 解析库**

运行: `cd .worktrees/auto-update/frontend && npm install marked`

预期输出: marked 库安装成功

**Step 3: 在 CommitPanel 中集成更新对话框**

修改 `frontend/src/components/CommitPanel.vue`:

```vue
<template>
  <div class="commit-panel">
    <UpdateNotification @show-update-dialog="showUpdateDialog = true" />
    <UpdateDialog :visible="showUpdateDialog" @close="showUpdateDialog = false" />

    <!-- 现有内容 -->
  </div>
</template>

<script setup lang="ts">
import UpdateNotification from './UpdateNotification.vue'
import UpdateDialog from './UpdateDialog.vue'
import { ref } from 'vue'

const showUpdateDialog = ref(false)
</script>
```

**Step 4: 提交**

```bash
cd .worktrees/auto-update
git add frontend/src/components/UpdateDialog.vue frontend/src/components/CommitPanel.vue frontend/package.json frontend/package-lock.json
git commit -m "feat(frontend): 添加更新详情对话框"
```

---

## 后续任务（待完成）

- 下载器实现
- 更新器实现
- 安装器实现
- 下载进度对话框
- 重启确认对话框
- 用户偏好存储
- 测试和优化

---

## 测试指南

### 手动测试流程

1. **启动应用**：运行 `wails dev`
2. **检查更新**：查看控制台日志，确认检查更新成功
3. **显示通知**：如果有更新，应该看到更新通知条
4. **查看详情**：点击"查看详情"按钮，打开更新对话框
5. **跳过版本**：点击"跳过此版本"，通知应该消失

### 自动化测试

```bash
# 运行所有测试
cd .worktrees/auto-update
go test ./... -v

# 运行特定包测试
go test ./pkg/version -v
go test ./pkg/service -v
```

---

## 提交规范

每次提交都应该遵循 Conventional Commits 格式：

- `feat:` 新功能
- `fix:` 修复 bug
- `ci:` CI/CD 相关
- `docs:` 文档
- `test:` 测试
- `refactor:` 重构

示例：
```bash
git commit -m "feat(version): 添加版本号解析功能"
git commit -m "fix(download): 修复下载进度显示问题"
git commit -m "ci: 添加 GitHub Actions 工作流"
```
