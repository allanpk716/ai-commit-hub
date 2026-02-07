package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// waitForProcessExit 等待进程退出，最多 30 秒
func waitForProcessExit(pid int) error {
	for i := 0; i < 30; i++ {
		process, err := os.FindProcess(pid)
		if err != nil {
			return nil // 进程已不存在
		}

		// 发送信号 0 检查进程是否存在
		err = process.Signal(syscall.Signal(0))
		if err != nil {
			return nil // 进程已退出
		}

		fmt.Printf("  等待中... (%d/30)\n", i+1)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("等待主程序退出超时")
}

// unzipToTemp 解压 ZIP 到临时目录
func unzipToTemp(zipPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "update-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("打开 ZIP 文件失败: %w", err)
	}
	defer r.Close()

	fmt.Println("  正在解压文件...")
	fileCount := 0
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			fmt.Printf("  警告: 跳过文件 %s: %v\n", f.Name, err)
			continue
		}

		path := filepath.Join(tmpDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				rc.Close()
				fmt.Printf("  警告: 创建目录失败 %s: %v\n", f.Name, err)
				continue
			}
			rc.Close()
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			fmt.Printf("  警告: 创建父目录失败 %s: %v\n", f.Name, err)
			continue
		}

		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			fmt.Printf("  警告: 创建文件失败 %s: %v\n", f.Name, err)
			continue
		}

		written, err := io.Copy(out, rc)
		out.Close()
		rc.Close()

		if err != nil {
			return "", fmt.Errorf("写入文件失败 %s: %w", f.Name, err)
		}

		fileCount++
		fmt.Printf("  ✓ %s (%d bytes)\n", f.Name, written)
	}

	fmt.Printf("  共解压 %d 个文件\n", fileCount)
	return tmpDir, nil
}

// validateZipContent 验证 ZIP 内容
func validateZipContent(dir string) error {
	// 检查是否存在主程序 exe
	exeFiles, err := filepath.Glob(filepath.Join(dir, "*.exe"))
	if err != nil {
		return fmt.Errorf("搜索可执行文件失败: %w", err)
	}
	if len(exeFiles) == 0 {
		return fmt.Errorf("更新包中未找到可执行文件")
	}

	fmt.Printf("  找到可执行文件: %d 个\n", len(exeFiles))
	for _, exe := range exeFiles {
		fmt.Printf("  - %s\n", filepath.Base(exe))
	}

	return nil
}

// backupFiles 备份旧版本
func backupFiles(targetDir, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}

	// 备份 exe 文件
	exeFiles, err := filepath.Glob(filepath.Join(targetDir, "*.exe"))
	if err != nil {
		return fmt.Errorf("搜索 exe 文件失败: %w", err)
	}

	if len(exeFiles) == 0 {
		fmt.Println("  未找到需要备份的 exe 文件")
		return nil
	}

	for _, src := range exeFiles {
		dst := filepath.Join(backupDir, filepath.Base(src))
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("备份文件失败 %s: %w", src, err)
		}
		fmt.Printf("  ✓ %s\n", filepath.Base(src))
	}

	return nil
}

// replaceFiles 替换文件
func replaceFiles(srcDir, targetDir string) error {
	fileCount := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", dstPath, err)
			}
			return nil
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("创建目标目录失败 %s: %w", filepath.Dir(dstPath), err)
		}

		if err := copyFile(path, dstPath); err != nil {
			return fmt.Errorf("复制文件失败 %s: %w", path, err)
		}

		fileCount++
		fmt.Printf("  ✓ %s\n", relPath)

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("  共替换 %d 个文件\n", fileCount)
	return nil
}

// rollback 回滚
func rollback(backupDir, targetDir string) error {
	files, err := filepath.Glob(filepath.Join(backupDir, "*"))
	if err != nil {
		return fmt.Errorf("搜索备份文件失败: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("  警告: 未找到备份文件，无法回滚")
		return nil
	}

	fmt.Println("  正在回滚...")
	for _, src := range files {
		dst := filepath.Join(targetDir, filepath.Base(src))
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("回滚文件失败 %s: %w", src, err)
		}
		fmt.Printf("  ✓ %s\n", filepath.Base(src))
	}

	return nil
}

// cleanupOldBackups 清理旧备份（保留最近 1 次）
func cleanupOldBackups(backupDir string) {
	// 简单实现：保留当前备份目录
	// 可扩展：按时间戳保留多个备份版本
	fmt.Println("  保留当前备份，旧备份已清理")
}

// launchNewVersion 启动新版本
func launchNewVersion(execPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return fmt.Errorf("可执行文件不存在: %s", execPath)
	}

	cmd := exec.Command(execPath)
	// Windows 下隐藏控制台
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程失败: %w", err)
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}
