package integration

import (
	"context"
	"os"
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

	// 使用项目根目录下的测试数据库，避免 Windows 上的 SQLite 文件锁定问题
	testDBPath := filepath.Join("..", "..", "tmp", "test_integration.db")

	// 确保 tmp 目录存在
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0755)

	config := &repository.DatabaseConfig{Path: testDBPath}

	if err := repository.InitializeDatabase(config); err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	// 清理数据：删除所有测试数据（不关闭数据库连接）
	t.Cleanup(func() {
		db := repository.GetDB()
		db.Exec("DELETE FROM commit_histories")
		db.Exec("DELETE FROM git_projects")
	})

	return config, context.Background()
}

func TestAppAPI_AddProject(t *testing.T) {
	_, _ = setupTestApp(t)
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
	_, _ = setupTestApp(t)
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

func TestAppAPI_GetAllProjects(t *testing.T) {
	_, _ = setupTestApp(t)
	repo := repository.NewGitProjectRepository()

	// 创建多个项目
	project1 := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "project-1",
		SortOrder: 0,
	}
	project2 := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "project-2",
		SortOrder: 1,
	}

	repo.Create(project1)
	repo.Create(project2)

	// 获取所有项目
	projects, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestAppAPI_MoveProject(t *testing.T) {
	_, _ = setupTestApp(t)
	repo := repository.NewGitProjectRepository()

	// 创建3个项目
	project1 := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "project-1",
		SortOrder: 0,
	}
	project2 := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "project-2",
		SortOrder: 1,
	}
	project3 := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "project-3",
		SortOrder: 2,
	}

	repo.Create(project1)
	repo.Create(project2)
	repo.Create(project3)

	// 获取所有项目（按 sort_order 排序）
	projects, _ := repo.GetAll()

	// 交换 project2 和 project3 的排序
	// 交换前: [project1(0), project2(1), project3(2)]
	// 交换后: [project1(0), project2(2), project3(1)]
	projects[1].SortOrder, projects[2].SortOrder =
		projects[2].SortOrder, projects[1].SortOrder

	repo.Update(&projects[1])
	repo.Update(&projects[2])

	// 验证排序已交换
	// GetAll() 按sort_order排序，所以返回: [project1(0), project3(1), project2(2)]
	updatedProjects, _ := repo.GetAll()
	assert.Equal(t, "project-1", updatedProjects[0].Name)
	assert.Equal(t, "project-3", updatedProjects[1].Name) // project3 移到了第二个位置
	assert.Equal(t, "project-2", updatedProjects[2].Name) // project2 移到了第三个位置
	assert.Equal(t, 1, updatedProjects[1].SortOrder)
	assert.Equal(t, 2, updatedProjects[2].SortOrder)
}

func TestAppAPI_CommitHistory(t *testing.T) {
	_, _ = setupTestApp(t)
	projectRepo := repository.NewGitProjectRepository()
	historyRepo := repository.NewCommitHistoryRepository()

	// 创建测试项目
	project := &models.GitProject{
		Path:      helpers.SetupTestRepo(t).Path,
		Name:      "test-project",
		SortOrder: 0,
	}
	projectRepo.Create(project)

	// 创建历史记录
	history1 := &models.CommitHistory{
		ProjectID: project.ID,
		Message:   "feat: first commit",
		Provider:  "openai",
		Language:  "zh",
	}
	history2 := &models.CommitHistory{
		ProjectID: project.ID,
		Message:   "fix: bug fix",
		Provider:  "openai",
		Language:  "zh",
	}

	historyRepo.Create(history1)
	historyRepo.Create(history2)

	// 获取历史记录（按 created_at DESC 排序，最新的在前）
	histories, err := historyRepo.GetByProjectID(project.ID, 10)

	assert.NoError(t, err)
	assert.Len(t, histories, 2)
	// history2 后创建，所以在前面
	assert.Equal(t, "fix: bug fix", histories[0].Message)
	assert.Equal(t, "feat: first commit", histories[1].Message)
}
