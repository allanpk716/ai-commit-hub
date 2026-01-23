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
	// Git 2.47+ 默认分支是 "main"，旧版本是 "master"
	assert.Contains(t, []string{"master", "main"}, status.Branch)
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

func TestGetProjectStatus_DeletedFile(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	// 创建并提交文件
	helpers.WriteFile(t, repo.Path, "temp.txt", "temp content")
	helpers.RunGitCmd(t, repo.Path, "add", "temp.txt")
	helpers.RunGitCmd(t, repo.Path, "commit", "-m", "add temp")

	// 删除文件并暂存删除
	helpers.RunGitCmd(t, repo.Path, "rm", "temp.txt")
	helpers.RunGitCmd(t, repo.Path, "add", "-u")  // 暂存删除操作

	status, err := GetProjectStatus(context.Background(), repo.Path)

	assert.NoError(t, err)
	assert.True(t, status.HasStaged)

	// 查找被删除的文件
	found := false
	for _, f := range status.StagedFiles {
		if f.Path == "temp.txt" {
			found = true
			assert.Equal(t, "Deleted", f.Status)
		}
	}
	assert.True(t, found, "应该找到被删除的 temp.txt")
}
