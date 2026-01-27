package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// StagingStatus 暂存区状态
type StagingStatus struct {
	Staged   []StagedFile `json:"staged"`
	Unstaged []StagedFile `json:"unstaged"`
}

// GetStagingStatus 获取暂存区状态（已暂存和未暂存文件）
func GetStagingStatus(repoPath string) (*StagingStatus, error) {
	// 获取已暂存文件
	staged, err := getStagedFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("get staged files: %w", err)
	}

	// 获取未暂存文件
	unstaged, err := getUnstagedFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("get unstaged files: %w", err)
	}

	// 标记被忽略的文件
	staged = MarkIgnoredFiles(repoPath, staged)
	unstaged = MarkIgnoredFiles(repoPath, unstaged)

	return &StagingStatus{
		Staged:   staged,
		Unstaged: unstaged,
	}, nil
}

// getUnstagedFiles 获取未暂存文件列表
func getUnstagedFiles(repoPath string) ([]StagedFile, error) {
	cmd := exec.Command("git", "-C", repoPath,
		"diff", "--name-status", "--diff-filter=ADM")
	output, err := cmd.Output()
	if err != nil {
		return []StagedFile{}, nil // No unstaged files
	}

	return parseGitFilesToStaged(output), nil
}

// StageFile 暂存单个文件
func StageFile(repoPath, filePath string) error {
	cmd := exec.Command("git", "-C", repoPath, "add", "-f", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stage file: %s", string(output))
	}
	return nil
}

// StageAllFiles 暂存所有未暂存文件
func StageAllFiles(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "add", "-f", "-u")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stage all: %s", string(output))
	}
	return nil
}

// UnstageFile 取消暂存单个文件
func UnstageFile(repoPath, filePath string) error {
	cmd := exec.Command("git", "-C", repoPath, "reset", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unstage file: %s", string(output))
	}
	return nil
}

// UnstageAllFiles 取消暂存所有文件
func UnstageAllFiles(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "reset")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unstage all: %s", string(output))
	}
	return nil
}

// parseGitFilesToStaged 解析 git name-status 输出为 StagedFile 数组
func parseGitFilesToStaged(output []byte) []StagedFile {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]StagedFile, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
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
				Path:    filePath,
				Status:  status,
				Ignored: false, // 默认未被忽略，后面会检查
			})
		}
	}
	return files
}

// checkFileIgnored 检查文件是否被 .gitignore 忽略
func checkFileIgnored(repoPath, filePath string) bool {
	cmd := exec.Command("git", "-C", repoPath, "check-ignore", "-q", filePath)
	err := cmd.Run()
	// 如果命令返回 0，说明文件被忽略；返回 1，说明文件未被忽略
	return err == nil
}

// MarkIgnoredFiles 标记被忽略的文件
func MarkIgnoredFiles(repoPath string, files []StagedFile) []StagedFile {
	for i := range files {
		files[i].Ignored = checkFileIgnored(repoPath, files[i].Path)
	}
	return files
}
