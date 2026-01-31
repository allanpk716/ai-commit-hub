package pushover

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/WQGroup/logger"
)

const (
	configFileMode = 0644
)

// Installer Hook 安装器
type Installer struct {
	extensionPath string
}

// NewInstaller 创建安装器
func NewInstaller(extensionPath string) *Installer {
	return &Installer{extensionPath: extensionPath}
}

// Install 安装 Hook 到项目
func (in *Installer) Install(projectPath string, force bool) (*InstallResult, error) {
	// 检查扩展目录是否存在
	if _, err := os.Stat(in.extensionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	// 检查 install.py 是否存在
	installScript := filepath.Join(in.extensionPath, "install.py")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return nil, fmt.Errorf("install.py 不存在，请确保 cc-pushover-hook 扩展完整")
	}

	// 检查 Python 是否可用
	pythonCmd, err := in.findPython()
	if err != nil {
		return nil, err
	}

	// 构建命令参数
	args := []string{
		installScript,
		"--target-dir", projectPath,
		"--non-interactive",
	}

	if force {
		args = append(args, "--force")
	}

	// 调试日志：打印执行的命令
	logger.Debugf("Executing: %s %v", pythonCmd, args)
	logger.Debugf("Working dir: %s", in.extensionPath)

	// 执行安装脚本
	cmd := exec.Command(pythonCmd, args...)
	cmd.Dir = in.extensionPath // 设置工作目录为扩展目录

	// 打印实际的命令参数
	logger.Debugf("cmd.Path: %s", cmd.Path)
	for i, arg := range cmd.Args {
		logger.Debugf("cmd.Args[%d]: %s", i, arg)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("安装失败: %v\n输出: %s", err, string(output)),
		}, nil
	}

	// 解析并返回结果
	hookPath := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	return in.parseInstallResult(output, hookPath)
}

// Uninstall 卸载 Hook
func (in *Installer) Uninstall(projectPath string) error {
	hookDir := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")

	// 删除 hook 目录
	if err := os.RemoveAll(hookDir); err != nil {
		return fmt.Errorf("删除 hook 目录失败: %w", err)
	}

	return nil
}

// Update 更新项目的 Hook
func (in *Installer) Update(projectPath string) (*InstallResult, error) {
	// 检查扩展目录是否存在
	if _, err := os.Stat(in.extensionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	// 检查 install.py 是否存在
	installScript := filepath.Join(in.extensionPath, "install.py")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return nil, fmt.Errorf("install.py 不存在，请确保 cc-pushover-hook 扩展完整")
	}

	// 检查 Python 是否可用
	pythonCmd, err := in.findPython()
	if err != nil {
		return nil, err
	}

	// 构建命令参数（更新时使用 force 标志）
	args := []string{
		installScript,
		"--target-dir", projectPath,
		"--non-interactive",
		"--force",
	}

	// 调试日志：打印执行的命令
	logger.Debugf("Updating Hook: %s %v", pythonCmd, args)
	logger.Debugf("Working dir: %s", in.extensionPath)

	// 执行安装脚本
	cmd := exec.Command(pythonCmd, args...)
	cmd.Dir = in.extensionPath // 设置工作目录为扩展目录

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("更新失败: %v\n输出: %s", err, string(output)),
		}, nil
	}

	// 解析并返回结果
	hookPath := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	return in.parseInstallResult(output, hookPath)
}

// Reinstall 重装 Hook（保留用户配置）
func (in *Installer) Reinstall(projectPath string) (*InstallResult, error) {
	// 检查扩展目录是否存在
	if _, err := os.Stat(in.extensionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	// 检查 install.py 是否存在
	installScript := filepath.Join(in.extensionPath, "install.py")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return nil, fmt.Errorf("install.py 不存在，请确保 cc-pushover-hook 扩展完整")
	}

	// 检查 Python 是否可用
	pythonCmd, err := in.findPython()
	if err != nil {
		return nil, err
	}

	// 读取当前通知配置
	config := in.readNotificationConfig(projectPath)

	// 构建命令参数（使用 --reinstall 标志）
	args := []string{
		installScript,
		"--target-dir", projectPath,
		"--non-interactive",
		"--reinstall",
	}

	// 调试日志
	logger.Debugf("Reinstalling Hook: %s %v", pythonCmd, args)
	logger.Debugf("Working dir: %s", in.extensionPath)
	logger.Debugf("Config: NoPushover=%v, NoWindows=%v", config.NoPushoverFile, config.NoWindowsFile)

	// 执行安装脚本
	cmd := exec.Command(pythonCmd, args...)
	cmd.Dir = in.extensionPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("重装失败: %v\n输出: %s", err, string(output)),
		}, nil
	}

	// 恢复通知配置
	if restoreErr := in.restoreNotificationConfig(projectPath, config); restoreErr != nil {
		// 配置恢复失败记录警告，但不影响重装结果
		logger.Warnf("恢复通知配置失败: %v", restoreErr)
	}

	// 解析并返回结果
	hookPath := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	return in.parseInstallResult(output, hookPath)
}

// SetNotificationMode 设置通知模式
func (in *Installer) SetNotificationMode(projectPath string, mode NotificationMode) error {
	// 文件直接放在项目根目录下，与 Python hook 的路径一致
	noPushoverPath := filepath.Join(projectPath, ".no-pushover")
	noWindowsPath := filepath.Join(projectPath, ".no-windows")

	switch mode {
	case ModeEnabled:
		// 删除两个标记文件
		if err := os.Remove(noPushoverPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除 .no-pushover 失败: %w", err)
		}
		if err := os.Remove(noWindowsPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除 .no-windows 失败: %w", err)
		}

	case ModePushoverOnly:
		// 创建 .no-windows，删除 .no-pushover
		if err := os.WriteFile(noWindowsPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("创建 .no-windows 失败: %w", err)
		}
		os.Remove(noPushoverPath)

	case ModeWindowsOnly:
		// 创建 .no-pushover，删除 .no-windows
		if err := os.WriteFile(noPushoverPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("创建 .no-pushover 失败: %w", err)
		}
		os.Remove(noWindowsPath)

	case ModeDisabled:
		// 创建两个标记文件
		if err := os.WriteFile(noPushoverPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("创建 .no-pushover 失败: %w", err)
		}
		if err := os.WriteFile(noWindowsPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("创建 .no-windows 失败: %w", err)
		}
	}

	return nil
}

// findPython 查找 Python 可执行文件
func (in *Installer) findPython() (string, error) {
	// 尝试 python3
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("python3"); err == nil {
			return "python3", nil
		}
	}

	// 尝试 python
	if _, err := exec.LookPath("python"); err == nil {
		return "python", nil
	}

	// Windows 上尝试 py
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("py"); err == nil {
			return "py", nil
		}
	}

	return "", fmt.Errorf("未找到 Python，请确保已安装 Python 3.6+")
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	NoPushoverFile bool
	NoWindowsFile  bool
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readNotificationConfig 读取通知配置
func (in *Installer) readNotificationConfig(projectPath string) NotificationConfig {
	noPushoverPath := filepath.Join(projectPath, ".no-pushover")
	noWindowsPath := filepath.Join(projectPath, ".no-windows")

	return NotificationConfig{
		NoPushoverFile: fileExists(noPushoverPath),
		NoWindowsFile:  fileExists(noWindowsPath),
	}
}

// restoreNotificationConfig 恢复通知配置
func (in *Installer) restoreNotificationConfig(projectPath string, config NotificationConfig) error {
	noPushoverPath := filepath.Join(projectPath, ".no-pushover")
	noWindowsPath := filepath.Join(projectPath, ".no-windows")

	// 恢复 .no-pushover
	if config.NoPushoverFile {
		if err := os.WriteFile(noPushoverPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("恢复 .no-pushover 失败: %w", err)
		}
	} else {
		if err := os.Remove(noPushoverPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除 .no-pushover 失败: %w", err)
		}
	}

	// 恢复 .no-windows
	if config.NoWindowsFile {
		if err := os.WriteFile(noWindowsPath, []byte(""), configFileMode); err != nil {
			return fmt.Errorf("恢复 .no-windows 失败: %w", err)
		}
	} else {
		if err := os.Remove(noWindowsPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除 .no-windows 失败: %w", err)
		}
	}

	return nil
}

// parseInstallResult 解析安装结果
func (in *Installer) parseInstallResult(output []byte, hookPath string) (*InstallResult, error) {
	outputStr := string(output)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	if len(lines) == 0 {
		return &InstallResult{
			Success: false,
			Message: "安装脚本无输出",
		}, nil
	}

	lastLine := lines[len(lines)-1]

	// 尝试解析 Python 格式的输出 (status: "success")
	var pythonResult PythonInstallResult
	if err := json.Unmarshal([]byte(lastLine), &pythonResult); err == nil {
		result := pythonResult.ToInstallResult()
		return &result, nil
	}

	// 尝试解析标准格式的输出 (success: true)
	var standardResult InstallResult
	if err := json.Unmarshal([]byte(lastLine), &standardResult); err == nil {
		return &standardResult, nil
	}

	// 如果两种格式都无法解析，检查输出中是否包含成功关键字
	if strings.Contains(outputStr, "success") || strings.Contains(outputStr, "complete") {
		return &InstallResult{
			Success:  true,
			Message:  "操作成功",
			HookPath: hookPath,
		}, nil
	}

	return &InstallResult{
		Success: false,
		Message:  fmt.Sprintf("无法解析安装结果: %s", outputStr),
	}, nil
}
