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
