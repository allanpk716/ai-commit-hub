package git

import (
	"fmt"
)

// GetFileDiff 获取文件的 diff 内容
func GetFileDiff(repoPath, filePath string, staged bool) (string, error) {
	args := []string{"-C", repoPath, "diff"}

	if staged {
		args = append(args, "--cached")
	}

	args = append(args, "--", filePath)

	cmd := Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get file diff: %w", err)
	}

	return string(output), nil
}
