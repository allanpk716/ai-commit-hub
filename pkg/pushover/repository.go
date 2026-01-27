package pushover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	pushoverRepoURL = "git@github.com:allanpk716/cc-pushover-hook.git"
	pushoverBranch  = "main"
)

// RepositoryManager Git 仓库管理器
type RepositoryManager struct {
	basePath string
}

// NewRepositoryManager 创建仓库管理器
func NewRepositoryManager(basePath string) *RepositoryManager {
	return &RepositoryManager{basePath: basePath}
}

// GetExtensionPath 获取扩展目录路径
func (rm *RepositoryManager) GetExtensionPath() string {
	return filepath.Join(rm.basePath, "cc-pushover-hook")
}

// IsCloned 检查扩展是否已克隆
func (rm *RepositoryManager) IsCloned() bool {
	extensionPath := rm.GetExtensionPath()
	_, err := os.Stat(extensionPath)
	return err == nil
}

// Clone 克隆 cc-pushover-hook 仓库
func (rm *RepositoryManager) Clone() error {
	if rm.IsCloned() {
		return fmt.Errorf("扩展已经存在，请使用 Update 方法更新")
	}

	// 确保 extensions 目录存在
	if err := os.MkdirAll(rm.basePath, 0755); err != nil {
		return fmt.Errorf("创建 extensions 目录失败: %w", err)
	}

	// 克隆仓库
	extensionPath := rm.GetExtensionPath()
	cmd := Command("git", "clone", "-b", pushoverBranch, "--single-branch", pushoverRepoURL, extensionPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("克隆失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// Update 更新 cc-pushover-hook 到最新版本
func (rm *RepositoryManager) Update() error {
	if !rm.IsCloned() {
		return fmt.Errorf("扩展不存在，请先使用 Clone 方法克隆")
	}

	extensionPath := rm.GetExtensionPath()

	// 执行 git pull
	cmd := Command("git", "pull", "origin", pushoverBranch)
	cmd.Dir = extensionPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("更新失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// fetchRemoteTags 获取远程 tags 和分支更新
func (rm *RepositoryManager) fetchRemoteTags() error {
	if !rm.IsCloned() {
		return fmt.Errorf("扩展不存在")
	}

	extensionPath := rm.GetExtensionPath()

	// 只获取远程的 tags 和 main 分支，不拉取完整代码
	cmd := Command("git", "fetch", "origin", "main", "--tags")
	cmd.Dir = extensionPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fetch 远程更新失败: %v\n输出: %s", err, string(output))
	}

	return nil
}

// GetVersion 获取当前扩展版本
func (rm *RepositoryManager) GetVersion() (string, error) {
	if !rm.IsCloned() {
		return "", fmt.Errorf("扩展不存在")
	}

	extensionPath := rm.GetExtensionPath()

	// 获取 git describe 输出作为版本
	cmd := Command("git", "describe", "--tags", "--always")
	cmd.Dir = extensionPath
	output, err := cmd.Output()
	if err != nil {
		// 如果没有 tag，使用 commit hash
		cmd = Command("git", "rev-parse", "--short", "HEAD")
		cmd.Dir = extensionPath
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("获取版本失败: %w", err)
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// GetLatestVersion 获取远程最新版本
func (rm *RepositoryManager) GetLatestVersion() (string, error) {
	if !rm.IsCloned() {
		return "", fmt.Errorf("扩展不存在")
	}

	// 先获取远程更新
	if err := rm.fetchRemoteTags(); err != nil {
		return "", err
	}

	extensionPath := rm.GetExtensionPath()

	// 获取远程最新 tag
	cmd := Command("git", "describe", "--tags", "--abbrev=0", "origin/main")
	cmd.Dir = extensionPath
	output, err := cmd.Output()
	if err != nil {
		// 如果没有 tag，返回 commit hash
		cmd = Command("git", "rev-parse", "--short", "origin/main")
		cmd.Dir = extensionPath
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("获取最新版本失败: %w", err)
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// GetExtensionInfo 获取扩展信息
func (rm *RepositoryManager) GetExtensionInfo() (*ExtensionInfo, error) {
	downloaded := rm.IsCloned()

	info := &ExtensionInfo{
		Downloaded: downloaded,
		Path:       rm.GetExtensionPath(),
	}

	if downloaded {
		version, err := rm.GetVersion()
		if err == nil {
			info.Version = version
			info.CurrentVersion = version
		}

		latestVersion, err := rm.GetLatestVersion()
		if err == nil {
			info.LatestVersion = latestVersion
			info.UpdateAvailable = version != latestVersion
		}
	}

	return info, nil
}

// Reclone 删除并重新克隆扩展
func (rm *RepositoryManager) Reclone() error {
	extensionPath := rm.GetExtensionPath()

	// 删除现有扩展目录
	if rm.IsCloned() {
		if err := os.RemoveAll(extensionPath); err != nil {
			return fmt.Errorf("删除扩展目录失败: %w", err)
		}
	}

	// 重新克隆
	return rm.Clone()
}
