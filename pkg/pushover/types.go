package pushover

import "time"

// NotificationMode 通知模式枚举
type NotificationMode string

const (
	ModeEnabled      NotificationMode = "enabled"       // 全部启用
	ModePushoverOnly NotificationMode = "pushover_only" // 仅 Pushover
	ModeWindowsOnly  NotificationMode = "windows_only"  // 仅 Windows
	ModeDisabled     NotificationMode = "disabled"      // 全部禁用
)

// HookStatus Hook 状态信息
type HookStatus struct {
	Installed       bool             `json:"installed"`
	Mode            NotificationMode `json:"mode"`
	Version         string           `json:"version"`
	InstalledAt     *time.Time       `json:"installed_at,omitempty"`
	UpdateAvailable bool             `json:"update_available,omitempty"` // 是否有可用更新
}

// ExtensionInfo 扩展信息
type ExtensionInfo struct {
	Downloaded      bool   `json:"downloaded"`
	Path            string `json:"path"`
	Version         string `json:"version"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
}

// InstallResult 安装结果
type InstallResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	HookPath string `json:"hook_path,omitempty"`
	Version  string `json:"version,omitempty"`
}

// PythonInstallResult Python 脚本输出的 JSON 格式
type PythonInstallResult struct {
	Status   string `json:"status"` // "success", "error", "cancelled"
	Message  string `json:"message,omitempty"`
	HookPath string `json:"hook_path,omitempty"`
	Version  string `json:"version,omitempty"`
}

// ToInstallResult 将 Python 格式的结果转换为 InstallResult
func (p *PythonInstallResult) ToInstallResult() InstallResult {
	return InstallResult{
		Success:  p.Status == "success",
		Message:  p.Message,
		HookPath: p.HookPath,
		Version:  p.Version,
	}
}
