package repository

import (
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

func TestGitProjectRepository(t *testing.T) {
	// Use temp database
	tempDir := t.TempDir()
	testDBPath := filepath.Join(tempDir, "test.db")

	// Initialize test database
	config := &DatabaseConfig{Path: testDBPath}
	if err := InitializeDatabase(config); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	repo := NewGitProjectRepository()

	t.Run("Create project", func(t *testing.T) {
		project := &models.GitProject{
			Path:      "/test/path",
			Name:      "test-project",
			SortOrder: 0,
		}
		if err := repo.Create(project); err != nil {
			t.Errorf("Failed to create project: %v", err)
		}
		if project.ID == 0 {
			t.Error("Project ID should be set after creation")
		}
	})

	t.Run("GetAll projects", func(t *testing.T) {
		projects, err := repo.GetAll()
		if err != nil {
			t.Errorf("Failed to get all projects: %v", err)
		}
		if len(projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(projects))
		}
	})

	t.Run("GetMaxSortOrder", func(t *testing.T) {
		maxOrder, err := repo.GetMaxSortOrder()
		if err != nil {
			t.Errorf("Failed to get max sort order: %v", err)
		}
		if maxOrder != 0 {
			t.Errorf("Expected max sort order 0, got %d", maxOrder)
		}
	})

	// Cleanup: close database connection to allow Windows to delete temp directory
	t.Cleanup(func() {
		CloseDatabase()
	})
}
