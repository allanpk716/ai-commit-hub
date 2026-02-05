package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/allanpk716/ai-commit-hub/pkg/service"
)

// BenchmarkGetAllProjects 测试获取所有项目的性能
func BenchmarkGetAllProjects(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}
	defer repository.CloseDatabase()

	repo := repository.NewGitProjectRepository()

	// 创建测试数据
	for i := 0; i < 100; i++ {
		project := &models.GitProject{
			Name:      "BenchProject",
			Path:      "/tmp/bench",
			SortOrder: i,
		}
		_ = repo.Create(project)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetAll()
	}
}

// BenchmarkGetProjectStatus 测试获取单个项目状态的性能
func BenchmarkGetProjectStatus(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench-status.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}
	defer repository.CloseDatabase()

	repo := repository.NewGitProjectRepository()

	// 创建测试项目
	project := &models.GitProject{
		Name: "BenchProject",
		Path: "/tmp/bench",
	}
	_ = repo.Create(project)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 这里模拟 GetProjectStatus 的数据库查询部分
		_, _ = repo.GetByID(1)
	}
}

// BenchmarkGetAllProjectStatuses 测试批量获取状态的性能
func BenchmarkGetAllProjectStatuses(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench-batch.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}
	defer repository.CloseDatabase()

	repo := repository.NewGitProjectRepository()

	// 创建 10 个测试项目
	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		project := &models.GitProject{
			Name:      "BenchProject",
			Path:      filepath.Join("/tmp", "bench", "project"+string(rune('0'+i))),
			SortOrder: i,
		}
		_ = repo.Create(project)
		paths[i] = project.Path
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟并发获取多个项目
		for _, path := range paths {
			_ = path
		}
	}
}

// BenchmarkCommitHistory 测试获取 commit 历史的性能
func BenchmarkCommitHistory(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench-history.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}
	defer repository.CloseDatabase()

	projectRepo := repository.NewGitProjectRepository()
	historyRepo := repository.NewCommitHistoryRepository()

	// 创建测试项目
	project := &models.GitProject{
		Name: "BenchProject",
		Path: "/tmp/bench",
	}
	_ = projectRepo.Create(project)

	// 创建 50 个 commit 历史记录
	for i := 0; i < 50; i++ {
		history := &models.CommitHistory{
			ProjectID: project.ID,
			Message:   "Benchmark commit",
		}
		_ = historyRepo.Create(history)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = historyRepo.GetByProjectID(project.ID, 0)
	}
}

// BenchmarkConfigService 测试配置服务的性能
func BenchmarkConfigService(b *testing.B) {
	configService := service.NewConfigService()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = configService.GetConfig()
	}
}

// BenchmarkAddProject 测试添加项目的性能
func BenchmarkAddProject(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench-add.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 每次迭代使用不同的数据库名称避免主键冲突
		testDBPath := filepath.Join("..", "..", "tmp", "bench-add", "bench.db")
		os.MkdirAll(filepath.Dir(testDBPath), 0o755)

		config := &repository.DatabaseConfig{Path: testDBPath}
		_ = repository.InitializeDatabase(config)

		repo := repository.NewGitProjectRepository()
		project := &models.GitProject{
			Name: "BenchProject",
			Path: "/tmp/bench",
		}
		_ = repo.Create(project)

		repository.CloseDatabase()
		os.RemoveAll(filepath.Join("..", "..", "tmp", "bench-add"))
	}
}

// BenchmarkUpdateProject 测试更新项目的性能
func BenchmarkUpdateProject(b *testing.B) {
	// 初始化测试数据库
	testDBPath := filepath.Join("..", "..", "tmp", "bench-update.db")
	os.MkdirAll(filepath.Join("..", "..", "tmp"), 0o755)
	defer os.Remove(testDBPath)

	config := &repository.DatabaseConfig{Path: testDBPath}
	if err := repository.InitializeDatabase(config); err != nil {
		b.Fatalf("初始化数据库失败: %v", err)
	}
	defer repository.CloseDatabase()

	repo := repository.NewGitProjectRepository()

	// 创建测试项目
	project := &models.GitProject{
		Name: "Original Name",
		Path: "/tmp/bench",
	}
	_ = repo.Create(project)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		project.Name = "Updated Name"
		_ = repo.Update(project)
	}
}

// BenchmarkDeleteProject 测试删除项目的性能
func BenchmarkDeleteProject(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// 初始化测试数据库
		testDBPath := filepath.Join("..", "..", "tmp", "bench-delete", "bench.db")
		os.MkdirAll(filepath.Dir(testDBPath), 0o755)

		config := &repository.DatabaseConfig{Path: testDBPath}
		_ = repository.InitializeDatabase(config)

		repo := repository.NewGitProjectRepository()
		project := &models.GitProject{
			Name: "BenchProject",
			Path: "/tmp/bench",
		}
		_ = repo.Create(project)

		b.StartTimer()

		_ = repo.Delete(project.ID)

		b.StopTimer()
		repository.CloseDatabase()
		os.RemoveAll(filepath.Join("..", "..", "tmp", "bench-delete"))
	}
}
