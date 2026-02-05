package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestGitProject_Validate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("Empty path should fail", func(t *testing.T) {
		project := &GitProject{
			Path: "",
		}
		err := project.Validate()
		if err == nil {
			t.Error("Expected error for empty path, got nil")
		}
		expectedErr := "项目路径不能为空"
		if err.Error() != expectedErr {
			t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
		}
	})

	t.Run("Non-existent path should fail", func(t *testing.T) {
		project := &GitProject{
			Path: filepath.Join(tempDir, "non-existent-path"),
		}
		err := project.Validate()
		if err == nil {
			t.Error("Expected error for non-existent path, got nil")
		}
		expectedErr := "路径不存在:"
		if err.Error()[:len(expectedErr)] != expectedErr {
			t.Errorf("Expected error to start with '%s', got '%s'", expectedErr, err.Error())
		}
	})

	t.Run("Invalid git repo should fail", func(t *testing.T) {
		// Create a directory that is not a git repo
		nonGitDir := filepath.Join(tempDir, "not-a-git-repo")
		if err := os.Mkdir(nonGitDir, 0o755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		project := &GitProject{
			Path: nonGitDir,
		}
		err := project.Validate()
		if err == nil {
			t.Error("Expected error for invalid git repo, got nil")
		}
		expectedErr := "不是有效的 git 仓库:"
		if err.Error()[:len(expectedErr)] != expectedErr {
			t.Errorf("Expected error to start with '%s', got '%s'", expectedErr, err.Error())
		}
	})

	t.Run("Valid git repo should pass", func(t *testing.T) {
		// Create a valid git repository
		validGitDir := filepath.Join(tempDir, "valid-git-repo")
		if err := os.Mkdir(validGitDir, 0o755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Initialize git repository
		_, err := git.PlainInit(validGitDir, false)
		if err != nil {
			t.Fatalf("Failed to initialize git repository: %v", err)
		}

		project := &GitProject{
			Path: validGitDir,
		}
		err = project.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid git repo, got: %v", err)
		}
	})
}

func TestGitProject_DetectName(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("Normal folder path should return folder name", func(t *testing.T) {
		projectPath := filepath.Join(tempDir, "my-project")
		if err := os.Mkdir(projectPath, 0o755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Initialize git repository
		_, err := git.PlainInit(projectPath, false)
		if err != nil {
			t.Fatalf("Failed to initialize git repository: %v", err)
		}

		project := &GitProject{
			Path: projectPath,
		}
		name, err := project.DetectName()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if name != "my-project" {
			t.Errorf("Expected name 'my-project', got '%s'", name)
		}
	})

	t.Run("Nested folder path should return leaf folder name", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "parent", "child", "project")
		if err := os.MkdirAll(nestedPath, 0o755); err != nil {
			t.Fatalf("Failed to create nested test directory: %v", err)
		}

		// Initialize git repository
		_, err := git.PlainInit(nestedPath, false)
		if err != nil {
			t.Fatalf("Failed to initialize git repository: %v", err)
		}

		project := &GitProject{
			Path: nestedPath,
		}
		name, err := project.DetectName()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if name != "project" {
			t.Errorf("Expected name 'project', got '%s'", name)
		}
	})

	t.Run("Path with special characters", func(t *testing.T) {
		specialPath := filepath.Join(tempDir, "my-project-123")
		if err := os.Mkdir(specialPath, 0o755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Initialize git repository
		_, err := git.PlainInit(specialPath, false)
		if err != nil {
			t.Fatalf("Failed to initialize git repository: %v", err)
		}

		project := &GitProject{
			Path: specialPath,
		}
		name, err := project.DetectName()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if name != "my-project-123" {
			t.Errorf("Expected name 'my-project-123', got '%s'", name)
		}
	})
}
