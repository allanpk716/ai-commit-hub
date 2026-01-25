package pushover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StatusChecker Hook 状态检测器
type StatusChecker struct {
	projectPath string
}

// NewStatusChecker 创建状态检测器
func NewStatusChecker(projectPath string) *StatusChecker {
	return &StatusChecker{projectPath: projectPath}
}

// CheckInstalled 检查 Hook 是否已安装
// 兼容新旧两种安装位置：
// - 新版本（1.0.0+）: .claude/hooks/pushover-hook/pushover-notify.py
// - 旧版本: .claude/hooks/pushover-notify.py
func (sc *StatusChecker) CheckInstalled() bool {
	// 优先检查新位置
	newHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "pushover-notify.py")
	fmt.Printf("[DEBUG] StatusChecker: 检查新位置: %s\n", newHookPath)
	if _, err := os.Stat(newHookPath); err == nil {
		fmt.Printf("[DEBUG] StatusChecker: 找到文件！\n")
		return true
	}
	fmt.Printf("[DEBUG] StatusChecker: 新位置不存在\n")

	// 兼容旧位置
	oldHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-notify.py")
	fmt.Printf("[DEBUG] StatusChecker: 检查旧位置: %s\n", oldHookPath)
	if _, err := os.Stat(oldHookPath); err == nil {
		fmt.Printf("[DEBUG] StatusChecker: 找到文件！\n")
		return true
	}
	fmt.Printf("[DEBUG] StatusChecker: 旧位置不存在\n")

	return false
}

// GetNotificationMode 获取当前通知模式
func (sc *StatusChecker) GetNotificationMode() NotificationMode {
	claudeDir := filepath.Join(sc.projectPath, ".claude")

	noPushoverPath := filepath.Join(claudeDir, ".no-pushover")
	noWindowsPath := filepath.Join(claudeDir, ".no-windows")

	_, hasNoPushover := os.Stat(noPushoverPath)
	_, hasNoWindows := os.Stat(noWindowsPath)

	// 文件存在时 isExists 为 true，不存在时为 false
	noPushoverExists := !os.IsNotExist(hasNoPushover)
	noWindowsExists := !os.IsNotExist(hasNoWindows)

	switch {
	case !noPushoverExists && !noWindowsExists:
		// 两个文件都不存在 → 全部启用
		return ModeEnabled
	case !noPushoverExists && noWindowsExists:
		// 只有 .no-windows 存在 → 仅 Pushover
		return ModePushoverOnly
	case noPushoverExists && !noWindowsExists:
		// 只有 .no-pushover 存在 → 仅 Windows
		return ModeWindowsOnly
	default:
		// 两个文件都存在 → 全部禁用
		return ModeDisabled
	}
}

// GetHookVersion 获取 Hook 版本
// 优先从 VERSION 文件读取，如果不存在则返回空字符串（表示旧版本）
func (sc *StatusChecker) GetHookVersion() (string, error) {
	// 检查新版本的 VERSION 文件
	versionFilePath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "VERSION")
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		// 文件不存在，可能是旧版本安装
		return "", fmt.Errorf("VERSION file not found: %w", err)
	}

	// 解析 version= 行
	content := strings.TrimSpace(string(data))
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "version=") {
			version := strings.TrimPrefix(line, "version=")
			version = strings.TrimSpace(version)
			if version != "" {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("no version found in VERSION file")
}

// GetInstalledAt 获取安装时间
// 兼容新旧两种安装位置
func (sc *StatusChecker) GetInstalledAt() (*time.Time, error) {
	// 优先检查新位置
	newHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "pushover-notify.py")
	if info, err := os.Stat(newHookPath); err == nil {
		modTime := info.ModTime()
		return &modTime, nil
	}

	// 兼容旧位置
	oldHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-notify.py")
	if info, err := os.Stat(oldHookPath); err == nil {
		modTime := info.ModTime()
		return &modTime, nil
	}

	return nil, fmt.Errorf("hook 文件不存在")
}

// GetStatus 获取完整的 Hook 状态
func (sc *StatusChecker) GetStatus() (*HookStatus, error) {
	installed := sc.CheckInstalled()
	if !installed {
		return &HookStatus{
			Installed: false,
			Mode:      ModeDisabled,
		}, nil
	}

	mode := sc.GetNotificationMode()
	version, err := sc.GetHookVersion()
	if err != nil {
		// VERSION 文件不存在，可能是旧版本安装
		version = "unknown"
	}
	installedAt, _ := sc.GetInstalledAt()

	return &HookStatus{
		Installed:   true,
		Mode:        mode,
		Version:     version,
		InstalledAt: installedAt,
	}, nil
}
