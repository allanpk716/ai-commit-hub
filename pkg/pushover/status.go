package pushover

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/WQGroup/logger"
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
	logger.Debugf("StatusChecker: 检查新位置: %s", newHookPath)
	if _, err := os.Stat(newHookPath); err == nil {
		logger.Debug("StatusChecker: 找到文件！")
		return true
	}
	logger.Debug("StatusChecker: 新位置不存在")

	// 兼容旧位置
	oldHookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-notify.py")
	logger.Debugf("StatusChecker: 检查旧位置: %s", oldHookPath)
	if _, err := os.Stat(oldHookPath); err == nil {
		logger.Debug("StatusChecker: 找到文件！")
		return true
	}
	logger.Debug("StatusChecker: 旧位置不存在")

	return false
}

// GetNotificationMode 获取当前通知模式
func (sc *StatusChecker) GetNotificationMode() NotificationMode {
	// 文件直接放在项目根目录下，与 Python hook 的路径一致
	noPushoverPath := filepath.Join(sc.projectPath, ".no-pushover")
	noWindowsPath := filepath.Join(sc.projectPath, ".no-windows")

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

// cleanVersion 清理版本号，从 Git describe 输出中提取纯版本号
// 例如: "v1.6.0-1-g3871faa" -> "v1.6.0"
// 如果不是 Git describe 格式（如 "v1.6.0-alpha"），则保持不变
func cleanVersion(version string) string {
	// 匹配 Git describe 输出格式: v{major}.{minor}[.{patch}]-{num}-g{hash}
	// 例如: v1.6.0-1-g3871faa, v2.0.0-15-gabc123, v1.6-5-gabc123
	re := regexp.MustCompile(`^v?(\d+\.\d+(?:\.\d+)?)-\d+-g[0-9a-f]+$`)
	if re.MatchString(version) {
		// 提取纯版本号部分
		parts := strings.Split(version, "-")
		if len(parts) >= 1 {
			return parts[0] // 返回版本号部分
		}
	}
	return version
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
				// 清理版本号，移除 Git 提交信息
				version = cleanVersion(version)
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
func (sc *StatusChecker) GetStatus(latestExtensionVersion string) (*HookStatus, error) {
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

	// 检查是否有可用更新
	updateAvailable := false
	if installed && version != "" && version != "unknown" && latestExtensionVersion != "" {
		updateAvailable = CompareVersions(version, latestExtensionVersion) < 0
	}

	return &HookStatus{
		Installed:       true,
		Mode:            mode,
		Version:         version,
		InstalledAt:     installedAt,
		UpdateAvailable: updateAvailable,
	}, nil
}
