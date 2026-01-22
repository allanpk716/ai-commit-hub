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

	var createdID uint

	t.Run("GetByID existing project", func(t *testing.T) {
		// First, get the project we created earlier
		projects, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed to get projects: %v", err)
		}
		if len(projects) == 0 {
			t.Fatal("No projects found")
		}
		createdID = projects[0].ID

		project, err := repo.GetByID(createdID)
		if err != nil {
			t.Errorf("Failed to get project by ID: %v", err)
		}
		if project == nil {
			t.Error("Expected project, got nil")
		}
		if project.ID != createdID {
			t.Errorf("Expected ID %d, got %d", createdID, project.ID)
		}
		if project.Name != "test-project" {
			t.Errorf("Expected name 'test-project', got '%s'", project.Name)
		}
	})

	t.Run("GetByID non-existent project", func(t *testing.T) {
		_, err := repo.GetByID(99999)
		if err == nil {
			t.Error("Expected error for non-existent project ID, got nil")
		}
	})

	t.Run("Update project", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No valid project ID to update")
		}

		project, err := repo.GetByID(createdID)
		if err != nil {
			t.Fatalf("Failed to get project for update: %v", err)
		}

		// Update project fields
		project.Name = "updated-project"
		project.SortOrder = 100

		err = repo.Update(project)
		if err != nil {
			t.Errorf("Failed to update project: %v", err)
		}

		// Verify the update
		updatedProject, err := repo.GetByID(createdID)
		if err != nil {
			t.Fatalf("Failed to get updated project: %v", err)
		}

		if updatedProject.Name != "updated-project" {
			t.Errorf("Expected name 'updated-project', got '%s'", updatedProject.Name)
		}
		if updatedProject.SortOrder != 100 {
			t.Errorf("Expected sort order 100, got %d", updatedProject.SortOrder)
		}
	})

	t.Run("Delete project", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No valid project ID to delete")
		}

		// Create a new project specifically for deletion test
		deleteTestProject := &models.GitProject{
			Path:      "/test/path/to-delete",
			Name:      "delete-test-project",
			SortOrder: 200,
		}
		if err := repo.Create(deleteTestProject); err != nil {
			t.Fatalf("Failed to create project for deletion: %v", err)
		}

		deleteID := deleteTestProject.ID

		// Verify project exists before deletion
		_, err := repo.GetByID(deleteID)
		if err != nil {
			t.Fatalf("Failed to get project before deletion: %v", err)
		}

		// Delete the project
		err = repo.Delete(deleteID)
		if err != nil {
			t.Errorf("Failed to delete project: %v", err)
		}

		// Verify project is deleted
		_, err = repo.GetByID(deleteID)
		if err == nil {
			t.Error("Expected error when getting deleted project, got nil")
		}
	})

	t.Run("Delete non-existent project", func(t *testing.T) {
		err := repo.Delete(88888)
		if err != nil {
			// Delete should not error even if project doesn't exist (GORM behavior)
			t.Errorf("Delete should not error for non-existent ID, got: %v", err)
		}
	})

	// Cleanup: close database connection to allow Windows to delete temp directory
	t.Cleanup(func() {
		CloseDatabase()
	})
}
