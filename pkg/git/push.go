package git

import (
	"context"
	"fmt"
	"strings"
)

// PushToRemote pushes the current branch to the origin remote repository.
// It uses the system git command to ensure compatibility with SSH keys and credentials.
func PushToRemote(ctx context.Context) error {
	// 获取当前分支名
	branchCmd := Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branchName := strings.TrimSpace(string(branchOutput))

	if branchName == "" || branchName == "HEAD" {
		return fmt.Errorf("not on any branch")
	}

	// 执行推送命令
	pushCmd := Command("git", "push", "origin", branchName)
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		// 提供详细的错误信息
		errorMsg := string(output)
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("push failed: %s", errorMsg)
	}

	return nil
}
