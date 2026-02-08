package update

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/WQGroup/logger"
	"golang.org/x/sys/windows"
)

//go:embed updater/updater.exe
var embeddedUpdater []byte

// 注意：更新器需要在嵌入前单独构建
// 构建命令：go build -o pkg/update/updater/updater.exe ./pkg/update/updater
// 或使用 Makefile/Wails 构建脚本自动构建

// Installer 安装器
type Installer struct{}

// NewInstaller 创建安装器
func NewInstaller() *Installer {
	return &Installer{}
}

// ExtractUpdater 释放嵌入的更新器到临时目录
func (i *Installer) ExtractUpdater() (string, error) {
	tmpDir := os.TempDir()
	updaterPath := filepath.Join(tmpDir, "ai-commit-hub-updater.exe")

	// 写入嵌入的更新器
	if err := os.WriteFile(updaterPath, embeddedUpdater, 0755); err != nil {
		return "", fmt.Errorf("写入更新器失败: %w", err)
	}

	logger.Infof("已释放更新器到: %s", updaterPath)
	return updaterPath, nil
}

// Install 安装更新
func (i *Installer) Install(updateZipPath string) error {
	logger.WithField("zip", updateZipPath).Info("开始安装更新...")

	// 释放更新器
	updaterPath, err := i.ExtractUpdater()
	if err != nil {
		return fmt.Errorf("释放更新器失败: %w", err)
	}

	// 获取主程序信息
	pid := os.Getpid()
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	targetDir := filepath.Dir(execPath)

	logger.Infof("启动更新器: %s", updaterPath)
	logger.Infof("参数: source=%s, target=%s, pid=%d, exec=%s", updateZipPath, targetDir, pid, execPath)

	// 启动更新器进程
	cmd := exec.Command(updaterPath,
		"--source", updateZipPath,
		"--target", targetDir,
		"--pid", strconv.Itoa(pid),
		"--exec", execPath,
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
