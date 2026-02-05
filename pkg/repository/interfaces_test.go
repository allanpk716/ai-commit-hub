package repository

import (
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

func TestGitProjectRepository_Interface(t *testing.T) {
	// 创建 mock 数据
	projects := []models.GitProject{
		{ID: 1, Name: "Test1", Path: "/path1", SortOrder: 1},
		{ID: 2, Name: "Test2", Path: "/path2", SortOrder: 2},
	}

	repo := NewMockGitProjectRepository(projects, nil)

	// 测试 GetAll
	result, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(result))
	}

	// 测试 GetByID
	project, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if project == nil || project.Name != "Test1" {
		t.Error("GetByID returned wrong project")
	}

	// 测试 GetByPath
	project, err = repo.GetByPath("/path2")
	if err != nil {
		t.Fatalf("GetByPath failed: %v", err)
	}

	if project == nil || project.Name != "Test2" {
		t.Error("GetByPath returned wrong project")
	}

	// 测试 Create
	newProject := &models.GitProject{ID: 3, Name: "Test3", Path: "/path3", SortOrder: 3}
	err = repo.Create(newProject)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, _ = repo.GetAll()
	if len(result) != 3 {
		t.Errorf("Expected 3 projects after Create, got %d", len(result))
	}

	// 测试 Update
	project, _ = repo.GetByID(1)
	project.Name = "Updated Test1"
	err = repo.Update(project)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := repo.GetByID(1)
	if updated.Name != "Updated Test1" {
		t.Errorf("Update did not modify the project: expected 'Updated Test1', got '%s'", updated.Name)
	}

	// 测试 Delete
	err = repo.Delete(1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	result, _ = repo.GetAll()
	if len(result) != 2 {
		t.Errorf("Expected 2 projects after Delete, got %d", len(result))
	}

	deleted, _ := repo.GetByID(1)
	if deleted != nil {
		t.Error("Deleted project still exists")
	}

	// 测试 GetMaxSortOrder
	maxOrder, err := repo.GetMaxSortOrder()
	if err != nil {
		t.Fatalf("GetMaxSortOrder failed: %v", err)
	}

	if maxOrder != 3 {
		t.Errorf("Expected max sort order 3, got %d", maxOrder)
	}
}

func TestCommitHistoryRepository_Interface(t *testing.T) {
	// 创建 mock 数据
	histories := []models.CommitHistory{
		{ID: 1, ProjectID: 1, Message: "Commit 1"},
		{ID: 2, ProjectID: 1, Message: "Commit 2"},
		{ID: 3, ProjectID: 2, Message: "Commit 3"},
	}

	repo := NewMockCommitHistoryRepository(histories, nil)

	// 测试 GetByProjectID
	result, err := repo.GetByProjectID(1, 0)
	if err != nil {
		t.Fatalf("GetByProjectID failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 histories for project 1, got %d", len(result))
	}

	// 测试 limit 参数
	result, err = repo.GetByProjectID(1, 1)
	if err != nil {
		t.Fatalf("GetByProjectID with limit failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 history with limit=1, got %d", len(result))
	}

	// 测试 Create
	newHistory := &models.CommitHistory{ID: 4, ProjectID: 2, Message: "Commit 4"}
	err = repo.Create(newHistory)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	result, _ = repo.GetByProjectID(2, 0)
	if len(result) != 2 {
		t.Errorf("Expected 2 histories after Create, got %d", len(result))
	}

	// 测试 GetRecent
	recent, err := repo.GetRecent(2)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(recent) != 2 {
		t.Errorf("Expected 2 recent histories, got %d", len(recent))
	}
}
