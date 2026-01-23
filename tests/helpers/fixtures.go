package helpers

import (
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
	RunGitCmd(t, repoPath, "init")
	RunGitCmd(t, repoPath, "config", "user.name", "Test User")
	RunGitCmd(t, repoPath, "config", "user.email", "test@example.com")

	// 创建初始文件
	WriteFile(t, repoPath, "README.md", "# Test Repository\n")
	RunGitCmd(t, repoPath, "add", ".")
	RunGitCmd(t, repoPath, "commit", "-m", "init")

	return &TestRepo{
		Name: name,
		Path: repoPath,
	}
}

// CreateStagedChange 在测试仓库中创建暂存变更
func (tr *TestRepo) CreateStagedChange(t *testing.T, filename, content string) {
	t.Helper()

	WriteFile(t, tr.Path, filename, content)
	RunGitCmd(t, tr.Path, "add", filename)
}

// CreateModifiedFile 在测试仓库中创建修改的文件
func (tr *TestRepo) CreateModifiedFile(t *testing.T, filename, content string) {
	t.Helper()

	WriteFile(t, tr.Path, filename, content)
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

// WriteFile 写入文件（导出供其他测试使用）
func WriteFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入文件失败 %s: %v", path, err)
	}
}

// RunGitCmd 运行 Git 命令（导出供其他测试使用）
func RunGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v 失败: %v\n输出: %s", args, err, string(output))
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
