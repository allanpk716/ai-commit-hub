package git

import (
	"context"
	"os"
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

func TestCommitChanges_EmptyMessage(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "content")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	// 有变更但消息为空
	err := CommitChanges(context.Background(), "")

	// go-git 允许空消息
	assert.NoError(t, err)
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

func TestGetHeadCommitMessage_InitialCommit(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	msg, err := GetHeadCommitMessage(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "init", msg)
}

func TestGetCurrentBranch_Success(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	branch, err := GetCurrentBranch(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, branch)
	// Git 2.47+ 默认是 "main"，旧版本是 "master"
	assert.Contains(t, []string{"master", "main"}, branch)
}

func TestCommitChanges_MultipleCommits(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

	// 第一次提交
	repo.CreateStagedChange(t, "file1.txt", "content1")
	err := CommitChanges(context.Background(), "feat: first commit")
	assert.NoError(t, err)

	// 第二次提交
	repo.CreateStagedChange(t, "file2.txt", "content2")
	err = CommitChanges(context.Background(), "feat: second commit")
	assert.NoError(t, err)

	// 验证最新的提交消息
	msg, err := GetHeadCommitMessage(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "feat: second commit", msg)
}
