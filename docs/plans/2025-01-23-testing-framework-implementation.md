# AI Commit Hub 测试框架实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 AI Commit Hub 项目构建完整的自动化测试框架，包括单元测试、集成测试、E2E 测试和日志分析功能。

**架构:** 采用测试金字塔架构，从底层单元测试到顶层 E2E 测试，使用 Wails 内置测试支持、临时 Git 仓库、Mock AI Provider 和结构化日志捕获。

**Tech Stack:**
- Go testing + testify/assert (断言)
- gomock (Mock 外部依赖)
- SQLite 内存模式 (数据库隔离)
- Vitest + Vue Test Utils (前端测试)
- 自定义 LogCapture (日志分析)
- GitHub Actions (CI/CD)

---

## 阶段 1: 测试框架基础设施

### Task 1: 创建测试目录结构

**Files:**
- Create: `tests/e2e/helpers/`
- Create: `tests/integration/`
- Create: `tests/helpers/`
- Create: `pkg/git/` (已有，需添加测试)
- Create: `pkg/service/` (已有，需添加测试)
- Create: `frontend/tests/unit/`
- Create: `frontend/tests/integration/`

**Step 1: 创建 Go 测试目录**

```bash
mkdir -p tests/e2e/helpers
mkdir -p tests/integration
mkdir -p tests/helpers
```

运行: `mkdir -p tests/e2e/helpers tests/integration tests/helpers`
预期: 创建三个测试目录

**Step 2: 验证目录创建**

```bash
ls -la tests/
```

运行: `ls -la tests/`
预期: 显示 e2e, integration, helpers 三个子目录

**Step 3: 创建前端测试目录**

```bash
cd frontend
mkdir -p tests/unit/components tests/unit/stores tests/integration
```

运行: `mkdir -p frontend/tests/unit/components frontend/tests/unit/stores frontend/tests/integration`
预期: 创建前端测试目录结构

**Step 4: 验证前端目录创建**

```bash
ls -la frontend/tests/
```

运行: `ls -la frontend/tests/`
预期: 显示 unit 和 integration 两个子目录

**Step 5: 创建 .gitkeep 文件保持空目录**

```bash
touch tests/e2e/helpers/.gitkeep
touch tests/integration/.gitkeep
touch tests/helpers/.gitkeep
touch frontend/tests/unit/components/.gitkeep
touch frontend/tests/unit/stores/.gitkeep
touch frontend/tests/integration/.gitkeep
```

运行: `touch tests/e2e/helpers/.gitkeep tests/integration/.gitkeep tests/helpers/.gitkeep frontend/tests/unit/components/.gitkeep frontend/tests/unit/stores/.gitkeep frontend/tests/integration/.gitkeep`
预期: 创建 .gitkeep 文件

**Step 6: 提交**

```bash
git add tests/ frontend/tests/
git commit -m "test: 创建测试目录结构"
```

运行: `git add tests/ frontend/tests/ && git commit -m "test: 创建测试目录结构"`
预期: Git 提交成功

---

### Task 2: 创建测试辅助函数 - 日志捕获

**Files:**
- Create: `tests/helpers/log_capture.go`
- Test: `tests/helpers/log_capture_test.go`

**Step 1: 编写日志捕获器**

```go
package helpers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/WQGroup/logger"
)

// LogCapture 捕获日志输出
type LogCapture struct {
	buffer *bytes.Buffer
	logger *logger.Logger
}

// NewLogCapture 创建日志捕获器
func NewLogCapture(t *testing.T) *LogCapture {
	t.Helper()

	buffer := &bytes.Buffer{}

	// 配置 logger 输出到 buffer
	log := logger.NewLogger(
		logger.WithOutput(buffer),
		logger.WithLevel(logger.DebugLevel),
		logger.WithFormat(logger.TextFormat),
	)

	return &LogCapture{
		buffer: buffer,
		logger: log,
	}
}

// GetLogs 获取所有日志
func (lc *LogCapture) GetLogs() string {
	return lc.buffer.String()
}

// GetLogsByLevel 按级别获取日志
func (lc *LogCapture) GetLogsByLevel(level string) []string {
	logs := strings.Split(lc.buffer.String(), "\n")
	var filtered []string

	for _, log := range logs {
		if strings.Contains(log, "["+level+"]") {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// Contains 验证日志包含特定内容
func (lc *LogCapture) Contains(substring string) bool {
	return strings.Contains(lc.buffer.String(), substring)
}

// ContainsError 验证有错误日志
func (lc *LogCapture) ContainsError() bool {
	logs := lc.buffer.String()
	return strings.Contains(logs, "[ERROR]") || strings.Contains(logs, "[WARN]")
}

// GetLogger 获取 logger 实例
func (lc *LogCapture) GetLogger() *logger.Logger {
	return lc.logger
}
```

运行: `cat > tests/helpers/log_capture.go` (然后粘贴代码)
预期: 创建文件

**Step 2: 编写日志捕获器测试**

```go
package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogCapture(t *testing.T) {
	capture := NewLogCapture(t)

	assert.NotNil(t, capture)
	assert.NotNil(t, capture.buffer)
	assert.NotNil(t, capture.logger)
}

func TestLogCapture_Contains(t *testing.T) {
	capture := NewLogCapture(t)

	capture.logger.Info("test message")

	assert.True(t, capture.Contains("test message"))
	assert.False(t, capture.Contains("not found"))
}

func TestLogCapture_ContainsError(t *testing.T) {
	capture := NewLogCapture(t)

	capture.logger.Debug("debug message")
	assert.False(t, capture.ContainsError())

	capture.logger.Error("error message")
	assert.True(t, capture.ContainsError())
}

func TestLogCapture_GetLogsByLevel(t *testing.T) {
	capture := NewLogCapture(t)

	capture.logger.Debug("debug msg")
	capture.logger.Info("info msg")
	capture.logger.Error("error msg")

	debugLogs := capture.GetLogsByLevel("DEBUG")
	assert.Len(t, debugLogs, 1)
	assert.Contains(t, debugLogs[0], "debug msg")
}
```

运行: `cat > tests/helpers/log_capture_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 3: 运行测试验证**

```bash
go test ./tests/helpers/... -v
```

运行: `go test ./tests/helpers/... -v`
预期: 测试通过

**Step 4: 提交**

```bash
git add tests/helpers/
git commit -m "test: 添加日志捕获器辅助函数"
```

运行: `git add tests/helpers/ && git commit -m "test: 添加日志捕获器辅助函数"`
预期: Git 提交成功

---

### Task 3: 创建测试辅助函数 - 日志断言

**Files:**
- Create: `tests/helpers/log_assertions.go`
- Test: `tests/helpers/log_assertions_test.go`

**Step 1: 编写日志断言辅助函数**

```go
package helpers

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// LogAssertions 日志断言
type LogAssertions struct {
	t       *testing.T
	capture *LogCapture
}

// NewLogAssertions 创建日志断言器
func NewLogAssertions(t *testing.T, capture *LogCapture) *LogAssertions {
	t.Helper()
	return &LogAssertions{
		t:       t,
		capture: capture,
	}
}

// AssertLogContains 断言日志包含文本
func (a *LogAssertions) AssertLogContains(substring string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	if !a.capture.Contains(substring) {
		assert.Fail(a.t, "日志不包含预期内容", substring, msgAndArgs)
		return false
	}
	return true
}

// AssertLogNotContains 断言日志不包含文本
func (a *LogAssertions) AssertLogNotContains(substring string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	if a.capture.Contains(substring) {
		assert.Fail(a.t, "日志包含不应出现的内容", substring, msgAndArgs)
		return false
	}
	return true
}

// AssertLogCount 断言日志出现次数
func (a *LogAssertions) AssertLogCount(substring string, expectedCount int, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	actualCount := strings.Count(logs, substring)

	if actualCount != expectedCount {
		assert.Fail(a.t,
			"日志出现次数不符",
			"期望 '%s' 出现 %d 次，实际 %d 次",
			substring, expectedCount, actualCount,
		)
		return false
	}
	return true
}

// AssertLogPattern 断言日志匹配正则表达式
func (a *LogAssertions) AssertLogPattern(pattern string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	matched, err := regexp.MatchString(pattern, logs)

	if err != nil || !matched {
		assert.Fail(a.t, "日志不匹配正则表达式", pattern, msgAndArgs)
		return false
	}
	return true
}

// AssertNoErrors 断言没有错误或警告日志
func (a *LogAssertions) AssertNoErrors(msgAndArgs ...interface{}) bool {
	a.t.Helper()
	if a.capture.ContainsError() {
		logs := a.capture.GetLogs()
		assert.Fail(a.t, "发现错误或警告日志", logs, msgAndArgs)
		return false
	}
	return true
}

// AssertLogSequence 断言日志按顺序出现
func (a *LogAssertions) AssertLogSequence(substrings ...string) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	lastIndex := -1

	for _, substr := range substrings {
		index := strings.Index(logs, substr)
		if index == -1 || index <= lastIndex {
			assert.Fail(a.t,
				"日志顺序不正确",
				"期望按顺序出现: %v",
				substrings,
			)
			return false
		}
		lastIndex = index
	}
	return true
}
```

运行: `cat > tests/helpers/log_assertions.go` (然后粘贴代码)
预期: 创建文件

**Step 2: 编写测试**

```go
package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogAssertions_AssertLogContains(t *testing.T) {
	capture := NewLogCapture(t)
	capture.logger.Info("test message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogContains("test message"))
}

func TestLogAssertions_AssertLogNotContains(t *testing.T) {
	capture := NewLogCapture(t)
	capture.logger.Info("test message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogNotContains("not found"))
}

func TestLogAssertions_AssertNoErrors(t *testing.T) {
	capture := NewLogCapture(t)
	capture.logger.Info("info message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertNoErrors())

	capture.logger.Error("error message")
	assert.False(t, assertions.AssertNoErrors())
}

func TestLogAssertions_AssertLogCount(t *testing.T) {
	capture := NewLogCapture(t)
	capture.logger.Info("test")
	capture.logger.Info("test")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogCount("test", 2))
}
```

运行: `cat > tests/helpers/log_assertions_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 3: 运行测试**

```bash
go test ./tests/helpers/... -v
```

运行: `go test ./tests/helpers/... -v`
预期: 所有测试通过

**Step 4: 提交**

```bash
git add tests/helpers/
git commit -m "test: 添加日志断言辅助函数"
```

运行: `git add tests/helpers/ && git commit -m "test: 添加日志断言辅助函数"`
预期: Git 提交成功

---

### Task 4: 创建测试固件 - Git 仓库

**Files:**
- Create: `tests/helpers/fixtures.go`
- Test: `tests/helpers/fixtures_test.go`

**Step 1: 编写测试固件**

```go
package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRepo 测试仓库结构
type TestRepo struct {
	Name string
	Path string
}

// SetupTestRepo 创建测试 Git 仓库
func SetupTestRepo(t *testing.T) *TestRepo {
	t.Helper()
	return SetupTestRepoNamed(t, "test-repo")
}

// SetupTestRepoNamed 创建指定名称的测试仓库
func SetupTestRepoNamed(t *testing.T, name string) *TestRepo {
	t.Helper()

	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, name)

	// 创建目录
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 初始化 Git 仓库
	runGitCmd(t, repoPath, "init")
	runGitCmd(t, repoPath, "config", "user.name", "Test User")
	runGitCmd(t, repoPath, "config", "user.email", "test@example.com")

	// 创建初始文件
	writeFile(t, repoPath, "README.md", "# Test Repository\n")
	runGitCmd(t, repoPath, "add", ".")
	runGitCmd(t, repoPath, "commit", "-m", "init")

	return &TestRepo{
		Name: name,
		Path: repoPath,
	}
}

// CreateStagedChange 在测试仓库中创建暂存变更
func (tr *TestRepo) CreateStagedChange(t *testing.T, filename, content string) {
	t.Helper()

	writeFile(t, tr.Path, filename, content)
	runGitCmd(t, tr.Path, "add", filename)
}

// CreateModifiedFile 在测试仓库中创建修改的文件
func (tr *TestRepo) CreateModifiedFile(t *testing.T, filename, content string) {
	t.Helper()

	writeFile(t, tr.Path, filename, content)
	// 不暂存，只修改工作区
}

// GetStatus 获取仓库状态
func (tr *TestRepo) GetStatus(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = tr.Path
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}
	return string(output)
}

// runGitCmd 运行 Git 命令
func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v 失败: %v\n输出: %s", args, err, string(output))
	}
}

// writeFile 写入文件
func writeFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入文件失败 %s: %v", path, err)
	}
}

// AssertRepoClean 断言仓库是干净的
func AssertRepoClean(t *testing.T, repo *TestRepo) {
	t.Helper()

	status := repo.GetStatus(t)
	assert.Empty(t, status, "仓库应该是干净的，但有状态: %s", status)
}

// AssertHasStagedChanges 断言仓库有暂存变更
func AssertHasStagedChanges(t *testing.T, repo *TestRepo) {
	t.Helper()

	status := repo.GetStatus(t)
	assert.NotEmpty(t, status, "仓库应该有暂存变更")
}
```

运行: `cat > tests/helpers/fixtures.go` (然后粘贴代码)
预期: 创建文件

**Step 2: 编写测试**

```go
package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupTestRepo(t *testing.T) {
	repo := SetupTestRepo(t)

	assert.NotNil(t, repo)
	assert.NotEmpty(t, repo.Name)
	assert.NotEmpty(t, repo.Path)
	assert.DirExists(t, repo.Path)
	assert.FileExists(t, repo.Path+"/README.md")
}

func TestTestRepo_CreateStagedChange(t *testing.T) {
	repo := SetupTestRepo(t)

	repo.CreateStagedChange(t, "test.txt", "content")

	status := repo.GetStatus(t)
	assert.Contains(t, status, "test.txt")
}

func TestTestRepo_CreateModifiedFile(t *testing.T) {
	repo := SetupTestRepo(t)

	repo.CreateModifiedFile(t, "README.md", "# Updated")

	// 文件存在但未暂存
	assert.FileExists(t, repo.Path+"/README.md")
}

func TestAssertRepoClean(t *testing.T) {
	repo := SetupTestRepo(t)

	AssertRepoClean(t, repo)
}

func TestAssertHasStagedChanges(t *testing.T) {
	repo := SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	AssertHasStagedChanges(t, repo)
}
```

运行: `cat > tests/helpers/fixtures_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 3: 运行测试**

```bash
go test ./tests/helpers/... -v
```

运行: `go test ./tests/helpers/... -v`
预期: 所有测试通过

**Step 4: 提交**

```bash
git add tests/helpers/
git commit -m "test: 添加 Git 测试固件"
```

运行: `git add tests/helpers/ && git commit -m "test: 添加 Git 测试固件"`
预期: Git 提交成功

---

## 阶段 2: Git 操作层测试

### Task 5: Git 状态获取测试

**Files:**
- Modify: `pkg/git/status.go:1-102` (参考)
- Create: `pkg/git/status_test.go`

**Step 1: 编写测试用例**

```go
package git

import (
	"context"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetProjectStatus_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	defer helpers.AssertRepoClean(t, repo)

	status, err := GetProjectStatus(context.Background(), repo.Path)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "master", status.Branch) // 或 "main" 取决于 git 版本
	assert.False(t, status.HasStaged)
	assert.Empty(t, status.StagedFiles)
}

func TestGetProjectStatus_WithStagedChanges(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "newfile.txt", "new content")

	status, err := GetProjectStatus(context.Background(), repo.Path)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.HasStaged)
	assert.Len(t, status.StagedFiles, 1)
	assert.Equal(t, "newfile.txt", status.StagedFiles[0].Path)
	assert.Equal(t, "New", status.StagedFiles[0].Status)
}

func TestGetProjectStatus_NotGitRepo(t *testing.T) {
	tempDir := t.TempDir()

	status, err := GetProjectStatus(context.Background(), tempDir)

	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "不是 git 仓库")
}

func TestGetProjectStatus_ModifiedFile(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	// 修改现有文件
	repo.CreateStagedChange(t, "README.md", "# Updated")

	status, err := GetProjectStatus(context.Background(), repo.Path)

	assert.NoError(t, err)
	assert.True(t, status.HasStaged)

	// 查找 README.md
	found := false
	for _, f := range status.StagedFiles {
		if f.Path == "README.md" {
			found = true
			assert.Equal(t, "Modified", f.Status)
		}
	}
	assert.True(t, found, "应该找到 README.md")
}
```

运行: `cat > pkg/git/status_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 2: 运行测试**

```bash
go test ./pkg/git/... -v -run TestGetProjectStatus
```

运行: `go test ./pkg/git/... -v -run TestGetProjectStatus`
预期: 所有测试通过

**Step 3: 提交**

```bash
git add pkg/git/status_test.go
git commit -m "test: 添加 GetProjectStatus 测试"
```

运行: `git add pkg/git/status_test.go && git commit -m "test: 添加 GetProjectStatus 测试"`
预期: Git 提交成功

---

### Task 6: Git Diff 测试

**Files:**
- Create: `pkg/git/diff_test.go`

**Step 1: 编写测试**

```go
package git

import (
	"context"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetGitDiffIgnoringMoves_NoChanges(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	diff, err := GetGitDiffIgnoringMoves(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, diff)
}

func TestGetGitDiffIgnoringMoves_WithNewFile(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "new content")

	// 切换到测试目录执行
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	diff, err := GetGitDiffIgnoringMoves(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "test.txt")
}

func TestGetStagedDiff_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	diff, err := GetStagedDiff(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "diff --git")
}
```

运行: `cat > pkg/git/diff_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 2: 运行测试**

```bash
go test ./pkg/git/... -v -run TestGetGitDiff
```

运行: `go test ./pkg/git/... -v -run TestGetGitDiff`
预期: 测试通过

**Step 3: 提交**

```bash
git add pkg/git/diff_test.go
git commit -m "test: 添加 Git diff 测试"
```

运行: `git add pkg/git/diff_test.go && git commit -m "test: 添加 Git diff 测试"`
预期: Git 提交成功

---

### Task 7: Git Commit 测试

**Files:**
- Create: `pkg/git/commit_test.go`

**Step 1: 编写测试**

```go
package git

import (
	"context"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestCommitChanges_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	err := CommitChanges(context.Background(), "test: add file")

	assert.NoError(t, err)
	helpers.AssertRepoClean(t, repo)
}

func TestGetHeadCommitMessage_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	CommitChanges(context.Background(), "feat: my commit message")

	msg, err := GetHeadCommitMessage(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "feat: my commit message", msg)
}

func TestGetCurrentBranch_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	branch, err := GetCurrentBranch(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, branch)
	assert.Equal(t, "master", branch) // 或 "main"
}
```

运行: `cat > pkg/git/commit_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 2: 运行测试**

```bash
go test ./pkg/git/... -v -run TestCommit
```

运行: `go test ./pkg/git/... -v -run TestCommit`
预期: 测试通过

**Step 3: 提交**

```bash
git add pkg/git/commit_test.go
git commit -m "test: 添加 Git commit 测试"
```

运行: `git add pkg/git/commit_test.go && git commit -m "test: 添加 Git commit 测试"`
预期: Git 提交成功

---

## 阶段 3: Service 层测试

### Task 8: ConfigService 测试

**Files:**
- Create: `pkg/service/config_service_test.go`

**Step 1: 编写测试**

```go
package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigService_LoadConfig(t *testing.T) {
	service := NewConfigService()

	ctx := context.Background()
	cfg, err := service.LoadConfig(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "zh", cfg.Language)
}

func TestConfigService_GetAvailableProviders(t *testing.T) {
	service := NewConfigService()

	providers := service.GetAvailableProviders()

	assert.NotEmpty(t, providers)
	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
}

func TestConfigService_ResolvePromptTemplate_Default(t *testing.T) {
	service := NewConfigService()

	template, err := service.ResolvePromptTemplate("", "")

	assert.NoError(t, err)
	assert.NotEmpty(t, template)
}

func TestConfigService_ResolvePromptTemplate_Custom(t *testing.T) {
	service := NewConfigService()

	// 创建临时目录和文件
	tempDir := t.TempDir()
	promptsDir := filepath.Join(tempDir, "prompts")
	os.Mkdir(promptsDir, 0755)

	customPrompt := "Custom prompt template"
	promptPath := filepath.Join(promptsDir, "custom.txt")
	os.WriteFile(promptPath, []byte(customPrompt), 0644)

	template, err := service.ResolvePromptTemplate(tempDir, "custom.txt")

	assert.NoError(t, err)
	assert.Equal(t, customPrompt, template)
}
```

运行: `cat > pkg/service/config_service_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 2: 运行测试**

```bash
go test ./pkg/service/... -v -run TestConfigService
```

运行: `go test ./pkg/service/... -v -run TestConfigService`
预期: 测试通过

**Step 3: 提交**

```bash
git add pkg/service/config_service_test.go
git commit -m "test: 添加 ConfigService 测试"
```

运行: `git add pkg/service/config_service_test.go && git commit -m "test: 添加 ConfigService 测试"`
预期: Git 提交成功

---

### Task 9: CommitService 测试 (使用 Mock)

**Files:**
- Create: `pkg/service/commit_service_test.go`
- Create: `pkg/service/mock_ai_client.go`

**Step 1: 创建 Mock AI Client**

```go
package service

import (
	"context"
	"fmt"
)

// MockAIClient 是用于测试的 Mock AI Client
type MockAIClient struct {
	Response string
	Error    error
	Deltas   []string
}

func (m *MockAIClient) GetCommitMessage(ctx context.Context, prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

func (m *MockAIClient) StreamCommitMessage(ctx context.Context, prompt string, deltaFunc func(string)) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}

	full := ""
	for _, delta := range m.Deltas {
		deltaFunc(delta)
		full += delta
	}
	return full, nil
}

// NewMockAIClient 创建新的 Mock AI Client
func NewMockAIClient(response string, deltas []string) *MockAIClient {
	return &MockAIClient{
		Response: response,
		Deltas:   deltas,
	}
}

// NewMockAIClientWithError 创建返回错误的 Mock AI Client
func NewMockAIClientWithError(err error) *MockAIClient {
	return &MockAIClient{Error: err}
}
```

运行: `cat > pkg/service/mock_ai_client.go` (然后粘贴代码)
预期: 创建文件

**Step 2: 编写 CommitService 测试**

```go
package service

import (
	"context"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestCommitService_GenerateCommit_EmptyDiff(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	service := NewCommitService(context.Background())

	// 没有暂存变更
	err := service.GenerateCommit(repo.Path, "mock", "zh")

	// 应该返回 nil（空 diff 不算错误）
	assert.NoError(t, err)
}

func TestCommitService_GenerateCommit_WithChanges(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	service := NewCommitService(context.Background())

	// 注意：这个测试需要实际的配置或 mock
	// 暂时跳过，等集成测试完善
	t.Skip("需要 mock AI provider registry")
}
```

运行: `cat > pkg/service/commit_service_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 3: 运行测试**

```bash
go test ./pkg/service/... -v -run TestCommitService
```

运行: `go test ./pkg/service/... -v -run TestCommitService`
预期: 部分测试通过（一个被跳过）

**Step 4: 提交**

```bash
git add pkg/service/commit_service_test.go pkg/service/mock_ai_client.go
git commit -m "test: 添加 CommitService 测试和 Mock AI Client"
```

运行: `git add pkg/service/commit_service_test.go pkg/service/mock_ai_client.go && git commit -m "test: 添加 CommitService 测试和 Mock AI Client"`
预期: Git 提交成功

---

## 阶段 4: 集成测试

### Task 10: Wails API 集成测试

**Files:**
- Create: `tests/integration/api_test.go`

**Step 1: 编写 API 测试**

```go
package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// setupTestApp 创建测试用的 App 实例
func setupTestApp(t *testing.T) (*repository.DatabaseConfig, context.Context) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	config := &repository.DatabaseConfig{Path: dbPath}

	if err := repository.InitializeDatabase(config); err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	t.Cleanup(func() {
		repository.CloseDatabase()
	})

	return config, context.Background()
}

func TestAppAPI_AddProject(t *testing.T) {
	_, ctx := setupTestApp(t)
	repo := repository.NewGitProjectRepository()

	repoPath := helpers.SetupTestRepo(t).Path

	project := &models.GitProject{
		Path:      repoPath,
		Name:      "test-project",
		SortOrder: 0,
	}

	err := repo.Create(project)

	assert.NoError(t, err)
	assert.NotZero(t, project.ID)
	assert.Equal(t, "test-project", project.Name)
}

func TestAppAPI_DeleteProject(t *testing.T) {
	_, ctx := setupTestApp(t)
	repo := repository.NewGitProjectRepository()

	project := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "to-delete",
		SortOrder: 0,
	}
	repo.Create(project)

	err := repo.Delete(project.ID)

	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(project.ID)
	assert.Error(t, err)
}
```

运行: `cat > tests/integration/api_test.go` (然后粘贴代码)
预期: 创建测试文件

**Step 2: 运行测试**

```bash
go test ./tests/integration/... -v
```

运行: `go test ./tests/integration/... -v`
预期: 测试通过

**Step 3: 提交**

```bash
git add tests/integration/
git commit -m "test: 添加 API 集成测试"
```

运行: `git add tests/integration/ && git commit -m "test: 添加 API 集成测试"`
预期: Git 提交成功

---

## 阶段 5: 测试脚本和报告

### Task 11: 创建本地测试脚本

**Files:**
- Create: `scripts/run-tests.bat`

**Step 1: 编写 Windows 测试脚本**

```batch
@echo off
setlocal enabledelayedexpansion

echo ===================================
echo AI Commit Hub - Test Runner
echo ===================================

set RUN_UNIT=1
set RUN_INTEGRATION=1
set RUN_FRONTEND=0
set COVERAGE=0

:parse_args
if "%~1"=="--no-unit" set RUN_UNIT=0
if "%~1"=="--no-integration" set RUN_INTEGRATION=0
if "%~1"=="--frontend" set RUN_FRONTEND=1
if "%~1"=="--coverage" set COVERAGE=1
shift
if not "%~1"=="" goto parse_args

if not exist "tmp\test-results" mkdir tmp\test-results

set TOTAL_TESTS=0
set PASSED_TESTS=0
set FAILED_TESTS=0

if %RUN_UNIT%==1 (
    echo.
    echo [1/2] Running Backend Unit Tests...
    echo -----------------------------------

    go test ./pkg/git/... -v > tmp\test-results\unit-git.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Git tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Git tests - see tmp\test-results\unit-git.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1

    go test ./pkg/service/... -v > tmp\test-results\unit-service.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Service tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Service tests - see tmp\test-results\unit-service.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1
)

if %RUN_INTEGRATION%==1 (
    echo.
    echo [2/2] Running Integration Tests...
    echo -----------------------------------

    go test ./tests/integration/... -v > tmp\test-results\integration.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Integration tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Integration tests - see tmp\test-results\integration.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1
)

echo.
echo ===================================
echo Test Summary
echo ===================================
echo Total Suites: %TOTAL_TESTS%
echo Passed: %PASSED_TESTS%
echo Failed: %FAILED_TESTS%
echo ===================================

if %FAILED_TESTS%==0 (
    echo All tests passed!
    exit /b 0
) else (
    echo Some tests failed!
    exit /b 1
)
```

运行: `cat > scripts/run-tests.bat` (然后粘贴代码)
预期: 创建脚本文件

**Step 2: 创建 scripts 目录（如果不存在）**

```bash
mkdir -p scripts
```

运行: `mkdir scripts`
预期: 创建目录

**Step 3: 测试脚本**

```bash
scripts\run-tests.bat --no-integration
```

运行: `scripts\run-tests.bat --no-integration`
预期: 运行单元测试

**Step 4: 提交**

```bash
git add scripts/
git commit -m "test: 添加本地测试运行脚本"
```

运行: `git add scripts/ && git commit -m "test: 添加本地测试运行脚本"`
预期: Git 提交成功

---

### Task 12: 创建 Makefile

**Files:**
- Create: `Makefile`

**Step 1: 编写 Makefile**

```makefile
.PHONY: test test-unit test-integration test-e2e test-frontend test-all coverage clean help

# 默认目标
help:
	@echo "AI Commit Hub - 测试命令"
	@echo ""
	@echo "可用命令:"
	@echo "  make test-all       - 运行所有测试"
	@echo "  make test-unit      - 运行后端单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make coverage       - 生成覆盖率报告"
	@echo "  make clean          - 清理测试文件"

# 运行所有测试
test-all:
	@echo "运行所有测试..."
	@scripts/run-tests.bat

# 后端单元测试
test-unit:
	@echo "运行后端单元测试..."
	@go test ./pkg/git/... -v
	@go test ./pkg/service/... -v
	@go test ./pkg/repository/... -v
	@go test ./tests/helpers/... -v

# 集成测试
test-integration:
	@echo "运行集成测试..."
	@go test ./tests/integration/... -v

# 覆盖率报告
coverage:
	@echo "生成覆盖率报告..."
	@go test ./pkg/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

# 清理测试文件
clean:
	@echo "清理测试文件..."
	@rm -rf tmp/test-results
	@rm -f coverage.out coverage.html
```

运行: `cat > Makefile` (然后粘贴代码)
预期: 创建 Makefile

**Step 2: 测试 Makefile**

```bash
make test-unit
```

运行: `make test-unit`
预期: 运行单元测试

**Step 3: 提交**

```bash
git add Makefile
git commit -m "test: 添加 Makefile 测试目标"
```

运行: `git add Makefile && git commit -m "test: 添加 Makefile 测试目标"`
预期: Git 提交成功

---

## 完成检查清单

执行完所有任务后，验证以下内容：

- [ ] 所有测试目录创建完成
- [ ] 日志捕获和断言辅助函数正常工作
- [ ] Git 测试固件可以创建临时仓库
- [ ] Git 操作层测试通过
- [ ] Service 层测试通过
- [ ] 集成测试通过
- [ ] 本地测试脚本可以运行
- [ ] Makefile 可以使用

运行最终验证:

```bash
make test-unit
make test-integration
```

预期: 所有测试通过

---

**计划完成！保存到 `docs/plans/2025-01-23-testing-framework-implementation.md`**
