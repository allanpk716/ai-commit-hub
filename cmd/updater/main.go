package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var (
	sourceFile = flag.String("source", "", "更新包 zip 文件路径")
	targetDir  = flag.String("target", "", "应用程序目录")
	pid        = flag.Int("pid", 0, "主进程 PID")
)

func main() {
	flag.Parse()

	fmt.Println("AI Commit Hub 更新器启动")
	fmt.Printf("源文件: %s\n", *sourceFile)
	fmt.Printf("目标目录: %s\n", *targetDir)
	fmt.Printf("主进程 PID: %d\n", *pid)

	// 验证参数
	if *sourceFile == "" || *targetDir == "" || *pid == 0 {
		fmt.Println("用法: updater.exe --source=<zip文件> --target=<应用目录> --pid=<主进程PID>")
		os.Exit(1)
	}

	// 等待主程序退出
	fmt.Println("等待主程序退出...")
	waitForProcess(*pid)

	// 解压文件
	if err := extractUpdate(*sourceFile, *targetDir); err != nil {
		fmt.Printf("解压失败: %v\n", err)
		os.Exit(1)
	}

	// 启动新版本
	if err := launchNewVersion(*targetDir); err != nil {
		fmt.Printf("启动新版本失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("更新完成！")
}

// waitForProcess 等待进程退出
func waitForProcess(pid int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("查找进程失败: %v\n", err)
		return
	}

	// 等待进程退出（最多等待 30 秒）
	done := make(chan bool, 1)
	go func() {
		process.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("进程已退出")
	case <-time.After(30 * time.Second):
		fmt.Println("等待超时，继续执行更新")
	}
}

// extractUpdate 解压更新文件
func extractUpdate(zipPath, targetDir string) error {
	fmt.Println("开始解压更新文件...")

	// 打开 zip 文件
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开 zip 文件失败: %w", err)
	}
	defer r.Close()

	// 创建备份目录
	backupDir := filepath.Join(targetDir, ".backup")
	os.MkdirAll(backupDir, 0755)

	// 备份当前可执行文件
	exeName := "ai-commit-hub.exe"
	if runtime.GOOS == "darwin" {
		exeName = "AI Commit Hub"
	}

	oldExePath := filepath.Join(targetDir, exeName)
	backupPath := filepath.Join(backupDir, exeName)

	if _, err := os.Stat(oldExePath); err == nil {
		fmt.Printf("备份当前程序到: %s\n", backupPath)
		os.Rename(oldExePath, backupPath)
	}

	// 解压文件
	for _, f := range r.File {
		// 只替换可执行文件
		if filepath.Base(f.Name) != exeName {
			continue
		}

		fmt.Printf("解压文件: %s\n", f.Name)

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("打开 zip 文件中的文件失败: %w", err)
		}

		// 创建目标文件
		destPath := filepath.Join(targetDir, exeName)
		destFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("创建目标文件失败: %w", err)
		}

		// 复制数据
		_, err = io.Copy(destFile, rc)
		rc.Close()
		destFile.Close()

		if err != nil {
			// 如果失败，恢复备份
			os.Remove(destPath)
			os.Rename(backupPath, oldExePath)
			return fmt.Errorf("解压文件失败: %w", err)
		}

		fmt.Printf("文件已更新: %s\n", destPath)
		break
	}

	r.Close()

	fmt.Println("解压完成")
	return nil
}

// launchNewVersion 启动新版本
func launchNewVersion(targetDir string) error {
	exeName := "ai-commit-hub.exe"
	if runtime.GOOS == "darwin" {
		exeName = "AI Commit Hub"
	}

	exePath := filepath.Join(targetDir, exeName)

	fmt.Printf("启动新版本: %s\n", exePath)

	cmd := exec.Command(exePath)
	cmd.Dir = targetDir

	// 启动新进程（不等待）
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动新版本失败: %w", err)
	}

	fmt.Println("新版本已启动")
	return nil
}
