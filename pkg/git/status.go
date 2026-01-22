package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type StagedFile struct {
	Path   string `json:"path"`
	Status string `json:"status"` // Modified, New, Deleted, Renamed
}

type ProjectStatus struct {
	Branch      string      `json:"branch"`
	StagedFiles []StagedFile `json:"staged_files"`
	HasStaged   bool        `json:"has_staged"`
}

func GetProjectStatus(ctx context.Context, projectPath string) (*ProjectStatus, error) {
	// Check if it's a git repo
	_, err := os.Stat(filepath.Join(projectPath, ".git"))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("不是 git 仓库: %s", projectPath)
	}

	// Get current branch
	branch, _ := getCurrentBranch(projectPath)

	// Get staged files
	stagedFiles, err := getStagedFiles(projectPath)
	if err != nil {
		return nil, err
	}

	return &ProjectStatus{
		Branch:      branch,
		StagedFiles: stagedFiles,
		HasStaged:   len(stagedFiles) > 0,
	}, nil
}

func getCurrentBranch(projectPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getStagedFiles(projectPath string) ([]StagedFile, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return []StagedFile{}, nil // No staged files
	}

	var files []StagedFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		statusCode := parts[0]
		filePath := parts[1]

		var status string
		switch statusCode {
		case "M":
			status = "Modified"
		case "A":
			status = "New"
		case "D":
			status = "Deleted"
		case "R":
			status = "Renamed"
		default:
			status = "Modified"
		}

		files = append(files, StagedFile{
			Path:   filePath,
			Status: status,
		})
	}

	return files, nil
}
