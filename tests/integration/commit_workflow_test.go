package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommitWorkflow 测试完整的 Commit 工作流
func TestCommitWorkflow(t *testing.T) {
	// 创建临时 Git 仓库
	tempDir, err := os.MkdirTemp("", "git-test-")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// 初始化 Git 仓库
	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "Test User")
	runGitCommand(t, tempDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tempDir, "config", "commit.gpgsign", "false")

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// 初始化应用
	testApp := app.NewApp()
	defer testApp.Quit()

	// 添加项目
	project, err := testApp.AddProject(models.GitProject{
		Name: "Test Project",
		Path: tempDir,
	})
	require.NoError(t, err, "Failed to add project")
	require.NotNil(t, project, "Project should not be nil")

	// 获取项目状态
	status, err := testApp.GetProjectStatus(tempDir)
	require.NoError(t, err, "Failed to get project status")
	require.NotNil(t, status, "Status should not be nil")

	// 验证有未提交的更改
	assert.True(t, status.GitStatus.HasUncommittedChanges, "Expected uncommitted changes")

	// 暂存文件
	err = testApp.StageFile(tempDir, "test.txt")
	require.NoError(t, err, "Failed to stage file")

	// 验证文件已暂存
	status, err = testApp.GetProjectStatus(tempDir)
	require.NoError(t, err, "Failed to get status after staging")
	require.NotNil(t, status.StagingStatus, "StagingStatus should not be nil")
	assert.Greater(t, len(status.StagingStatus.Staged), 0, "Expected staged files")

	// 提交更改
	commitMessage := "test: add test file"
	err = testApp.CommitProject(tempDir, commitMessage)
	require.NoError(t, err, "Failed to commit")

	// 验证提交成功
	status, err = testApp.GetProjectStatus(tempDir)
	require.NoError(t, err, "Failed to get status after commit")

	// 提交后应该没有未提交的更改（如果没有新修改）
	assert.False(t, status.GitStatus.HasUncommittedChanges, "Expected no uncommitted changes after commit")

	// 验证提交消息
	assert.Contains(t, status.GitStatus.LastCommitMessage, "test", "Commit message should contain 'test'")
}

// runGitCommand 辅助函数：在指定目录执行 Git 命令
func runGitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Git command failed: %v\nOutput: %s", err, string(output))
	}

	return string(output)
}

// TestProjectCRUD 测试项目的 CRUD 操作
func TestProjectCRUD(t *testing.T) {
	// 创建临时目录用于测试数据库
	tempDir, err := os.MkdirTemp("", "ai-commit-hub-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 初始化应用
	testApp := app.NewApp()
	defer testApp.Quit()

	// 创建测试项目路径
	projectPath1 := filepath.Join(tempDir, "project1")
	projectPath2 := filepath.Join(tempDir, "project2")

	// 初始化 Git 仓库
	os.MkdirAll(projectPath1, 0755)
	os.MkdirAll(projectPath2, 0755)
	runGitCommand(t, projectPath1, "init")
	runGitCommand(t, projectPath2, "init")

	// 测试 Create
	project1, err := testApp.AddProject(models.GitProject{
		Name:     "Project 1",
		Path:     projectPath1,
		SortOrder: 0,
	})
	require.NoError(t, err, "Failed to create project 1")
	assert.NotNil(t, project1)
	assert.Greater(t, project1.ID, 0, "Project should have valid ID")

	project2, err := testApp.AddProject(models.GitProject{
		Name:     "Project 2",
		Path:     projectPath2,
		SortOrder: 1,
	})
	require.NoError(t, err, "Failed to create project 2")

	// 测试 ReadAll
	projects, err := testApp.GetAllProjects()
	require.NoError(t, err, "Failed to get all projects")
	assert.Len(t, projects, 2, "Should have 2 projects")

	// 测试 Update
	project1.Name = "Updated Project 1"
	err = testApp.UpdateProject(*project1)
	require.NoError(t, err, "Failed to update project")

	// 验证更新
	projects, err = testApp.GetAllProjects()
	require.NoError(t, err)
	found := false
	for _, p := range projects {
		if p.ID == project1.ID {
			assert.Equal(t, "Updated Project 1", p.Name)
			found = true
			break
		}
	}
	assert.True(t, found, "Updated project should be found")

	// 测试 Delete
	err = testApp.DeleteProject(project1.ID)
	require.NoError(t, err, "Failed to delete project")

	// 验证删除
	projects, err = testApp.GetAllProjects()
	require.NoError(t, err)
	assert.Len(t, projects, 1, "Should have 1 project after deletion")
	assert.Equal(t, project2.ID, projects[0].ID, "Remaining project should be project 2")
}

// TestGetCommitHistory 测试获取 commit 历史
func TestGetCommitHistory(t *testing.T) {
	// 创建临时 Git 仓库
	tempDir, err := os.MkdirTemp("", "git-history-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 初始化 Git 仓库
	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "Test User")
	runGitCommand(t, tempDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tempDir, "config", "commit.gpgsign", "false")

	// 创建测试文件并提交
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(tempDir, filepath.Join("file", i))
		err = os.MkdirAll(filepath.Join(tempDir, "file"), 0755)
		require.NoError(t, err)

		err = os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		runGitCommand(t, tempDir, "add", ".")
		runGitCommand(t, tempDir, "commit", "-m", "Commit "+string(rune('0'+i)))
	}

	// 初始化应用
	testApp := app.NewApp()
	defer testApp.Quit()

	// 添加项目
	project, err := testApp.AddProject(models.GitProject{
		Name: "History Test",
		Path: tempDir,
	})
	require.NoError(t, err)

	// 获取 commit 历史
	history, err := testApp.GetCommitHistory(project.ID, 0)
	require.NoError(t, err, "Failed to get commit history")

	// 应该有至少 3 个提交（加上可能的初始提交）
	assert.GreaterOrEqual(t, len(history), 3, "Should have at least 3 commits")

	// 测试 limit 参数
	limitedHistory, err := testApp.GetCommitHistory(project.ID, 2)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(limitedHistory), 2, "Should have at most 2 commits with limit")
}
