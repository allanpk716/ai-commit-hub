package git

import (
	"context"
	"os"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestPushToRemote_NoRemote(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	// 提交变更
	err := CommitChanges(context.Background(), "test: add file")
	assert.NoError(t, err)

	// 尝试推送到远程仓库（没有配置远程仓库）
	err = PushToRemote(context.Background())

	// 应该返回错误，因为没有配置远程仓库
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remote")
}

func TestPushToRemote_NoOriginRemote(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	// 提交变更
	err := CommitChanges(context.Background(), "test: add file")
	assert.NoError(t, err)

	// 添加一个非 origin 的远程仓库
	helpers.RunGitCmd(t, repo.Path, "remote", "add", "upstream", "https://github.com/test/test.git")

	// 尝试推送到 origin（不存在）
	err = PushToRemote(context.Background())

	// 应该返回错误，因为没有 origin 远程仓库
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "origin")
}
