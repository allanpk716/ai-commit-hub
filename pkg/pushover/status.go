package pushover

import (
	"encoding/json"
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
func (sc *StatusChecker) GetHookVersion() (string, error) {
	settingsPath := filepath.Join(sc.projectPath, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", fmt.Errorf("无法读取 settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", fmt.Errorf("无法解析 settings.json: %w", err)
	}

	// 尝试从命令中提取版本信息（如果包含版本标记）
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	for _, eventHooks := range hooks {
		if eventArray, ok := eventHooks.([]interface{}); ok {
			for _, eventConfig := range eventArray {
				if configMap, ok := eventConfig.(map[string]interface{}); ok {
					if hooksList, ok := configMap["hooks"].([]interface{}); ok {
						for _, hook := range hooksList {
							if hookMap, ok := hook.(map[string]interface{}); ok {
								if command, ok := hookMap["command"].(string); ok {
									if strings.Contains(command, "pushover-notify.py") {
										// 目前版本信息存储在其他地方，返回默认值
										return "1.0.0", nil
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return "1.0.0", nil
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
	version, _ := sc.GetHookVersion()
	installedAt, _ := sc.GetInstalledAt()

	return &HookStatus{
		Installed:   true,
		Mode:        mode,
		Version:     version,
		InstalledAt: installedAt,
	}, nil
}
