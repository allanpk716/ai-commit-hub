package pushover

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	fmt.Fprintf(os.Stderr, "[DEBUG] Executing: %s %v\n", pythonCmd, args)
	fmt.Fprintf(os.Stderr, "[DEBUG] Working dir: %s\n", in.extensionPath)

	// 执行安装脚本
	cmd := exec.Command(pythonCmd, args...)
	cmd.Dir = in.extensionPath // 设置工作目录为扩展目录

	// 打印实际的命令参数
	fmt.Fprintf(os.Stderr, "[DEBUG] cmd.Path: %s\n", cmd.Path)
	for i, arg := range cmd.Args {
		fmt.Fprintf(os.Stderr, "[DEBUG] cmd.Args[%d]: %s\n", i, arg)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("安装失败: %v\n输出: %s", err, string(output)),
		}, nil
	}

	// 解析输出的最后一行 JSON
	outputStr := string(output)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	lastLine := lines[len(lines)-1]

	var result InstallResult
	if err := json.Unmarshal([]byte(lastLine), &result); err != nil {
		// 如果不是 JSON 格式，可能是旧版本或输出格式错误
		if strings.Contains(outputStr, "success") || strings.Contains(outputStr, "complete") {
			return &InstallResult{
				Success:  true,
				Message:  "安装成功",
				HookPath: filepath.Join(projectPath, ".claude", "hooks", "pushover-hook"),
			}, nil
		}
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("无法解析安装结果: %v\n输出: %s", err, outputStr),
		}, nil
	}

	return &result, nil
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
	fmt.Fprintf(os.Stderr, "[DEBUG] Updating Hook: %s %v\n", pythonCmd, args)
	fmt.Fprintf(os.Stderr, "[DEBUG] Working dir: %s\n", in.extensionPath)

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

	// 解析输出的最后一行 JSON
	outputStr := string(output)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	lastLine := lines[len(lines)-1]

	var result InstallResult
	if err := json.Unmarshal([]byte(lastLine), &result); err != nil {
		// 如果不是 JSON 格式，可能是旧版本或输出格式错误
		if strings.Contains(outputStr, "success") || strings.Contains(outputStr, "complete") {
			return &InstallResult{
				Success:  true,
				Message:  "更新成功",
				HookPath: filepath.Join(projectPath, ".claude", "hooks", "pushover-hook"),
			}, nil
		}
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("无法解析更新结果: %v\n输出: %s", err, outputStr),
		}, nil
	}

	return &result, nil
}

// SetNotificationMode 设置通知模式
func (in *Installer) SetNotificationMode(projectPath string, mode NotificationMode) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	noPushoverPath := filepath.Join(claudeDir, ".no-pushover")
	noWindowsPath := filepath.Join(claudeDir, ".no-windows")

	switch mode {
	case ModeEnabled:
		// 删除两个标记文件
		os.Remove(noPushoverPath)
		os.Remove(noWindowsPath)

	case ModePushoverOnly:
		// 创建 .no-windows，删除 .no-pushover
		if err := os.WriteFile(noWindowsPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("创建 .no-windows 失败: %w", err)
		}
		os.Remove(noPushoverPath)

	case ModeWindowsOnly:
		// 创建 .no-pushover，删除 .no-windows
		if err := os.WriteFile(noPushoverPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("创建 .no-pushover 失败: %w", err)
		}
		os.Remove(noWindowsPath)

	case ModeDisabled:
		// 创建两个标记文件
		if err := os.WriteFile(noPushoverPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("创建 .no-pushover 失败: %w", err)
		}
		if err := os.WriteFile(noWindowsPath, []byte(""), 0644); err != nil {
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
