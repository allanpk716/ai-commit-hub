package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/WQGroup/logger"
	"golang.org/x/sys/windows"
)

// Installer 安装器
type Installer struct {
	updaterPath string
}

// NewInstaller 创建安装器
func NewInstaller() *Installer {
	// 更新器路径：与主程序同目录
	execPath, err := os.Executable()
	if err != nil {
		logger.Errorf("获取可执行文件路径失败: %v", err)
		return &Installer{}
	}

	updaterDir := filepath.Dir(execPath)
	updaterName := "updater.exe"

	return &Installer{
		updaterPath: filepath.Join(updaterDir, updaterName),
	}
}

// Install 安装更新
func (i *Installer) Install(updateZipPath string) error {
	logger.Info("开始安装更新...", "zip", updateZipPath)

	// 获取主程序 PID
	pid := os.Getpid()

	// 获取可执行文件目录
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	targetDir := filepath.Dir(execPath)

	// 验证更新器存在
	if _, err := os.Stat(i.updaterPath); os.IsNotExist(err) {
		return fmt.Errorf("更新器程序不存在: %s", i.updaterPath)
	}

	logger.Infof("启动更新器: %s", i.updaterPath)
	logger.Infof("参数: source=%s, target=%s, pid=%d", updateZipPath, targetDir, pid)

	// 启动更新器进程
	cmd := exec.Command(i.updaterPath,
		"--source", updateZipPath,
		"--target", targetDir,
		"--pid", strconv.Itoa(pid),
	)

	// Windows 下隐藏控制台窗口
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新器失败: %w", err)
	}

	logger.Info("更新器已启动，主程序即将退出")

	// 退出主程序，释放文件锁
	return nil
}
