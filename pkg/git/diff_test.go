package git

import (
	"context"
	"os"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetGitDiffIgnoringMoves_NoChanges(t *testing.T) {
	repo := helpers.SetupTestRepo(t)

	// 切换到测试目录执行
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repo.Path)

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
