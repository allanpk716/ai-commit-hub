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
func (sc *StatusChecker) CheckInstalled() bool {
	hookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "pushover-notify.py")
	_, err := os.Stat(hookPath)
	return err == nil
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
func (sc *StatusChecker) GetInstalledAt() (*time.Time, error) {
	hookPath := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "pushover-notify.py")
	info, err := os.Stat(hookPath)
	if err != nil {
		return nil, err
	}

	modTime := info.ModTime()
	return &modTime, nil
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
