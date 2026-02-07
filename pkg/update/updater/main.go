package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	source   = flag.String("source", "", "更新包路径 (ZIP 文件)")
	target   = flag.String("target", "", "安装目录")
	pid      = flag.Int("pid", 0, "主进程 PID")
	execPath = flag.String("exec", "", "主程序完整路径")
)

func main() {
	flag.Parse()

	// 验证参数
	if *source == "" || *target == "" {
		fmt.Println("错误: 参数不完整")
		fmt.Println("用法: updater.exe --source <更新包路径> --target <安装目录> [--pid <进程PID>] [--exec <主程序路径>]")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("       AI Commit Hub 更新器")
	fmt.Println("========================================")
	fmt.Printf("更新包: %s\n", *source)
	fmt.Printf("安装目录: %s\n", *target)
	if *pid > 0 {
		fmt.Printf("等待进程退出: PID=%d\n", *pid)
	}
	if *execPath != "" {
		fmt.Printf("主程序: %s\n", *execPath)
	}
	fmt.Println("========================================")

	// 1. 等待主程序退出
	if *pid > 0 {
		fmt.Println("\n[1/6] 等待主程序退出...")
		if err := waitForProcessExit(*pid); err != nil {
			fmt.Printf("错误: 等待主程序退出失败: %v\n", err)
			pauseAndExit(1)
		}
		fmt.Println("✓ 主程序已退出")
	} else {
		fmt.Println("\n[1/6] 跳过等待主程序（未指定 PID）")
	}

	// 2. 解压 ZIP 到临时目录
	fmt.Println("\n[2/6] 解压更新包...")
	tmpDir, err := unzipToTemp(*source)
	if err != nil {
		fmt.Printf("错误: 解压失败: %v\n", err)
		pauseAndExit(1)
	}
	defer os.RemoveAll(tmpDir)
	fmt.Printf("✓ 解压到临时目录: %s\n", tmpDir)

	// 3. 验证 ZIP 内容
	fmt.Println("\n[3/6] 验证更新包内容...")
	if err := validateZipContent(tmpDir); err != nil {
		fmt.Printf("错误: 验证失败: %v\n", err)
		pauseAndExit(1)
	}
	fmt.Println("✓ 更新包验证通过")

	// 4. 备份旧版本
	fmt.Println("\n[4/6] 备份当前版本...")
	backupDir := filepath.Join(*target, "backup")
	if err := backupFiles(*target, backupDir); err != nil {
		fmt.Printf("错误: 备份失败: %v\n", err)
		pauseAndExit(1)
	}
	fmt.Printf("✓ 备份完成: %s\n", backupDir)

	// 5. 替换文件
	fmt.Println("\n[5/6] 替换文件...")
	if err := replaceFiles(tmpDir, *target); err != nil {
		fmt.Printf("错误: 替换失败，正在回滚: %v\n", err)
		if rollbackErr := rollback(backupDir, *target); rollbackErr != nil {
			fmt.Printf("错误: 回滚失败: %v\n", rollbackErr)
		} else {
			fmt.Println("✓ 已回滚到旧版本")
		}
		pauseAndExit(1)
	}
	fmt.Println("✓ 文件替换完成")

	// 6. 清理旧备份（保留最近 1 次）
	fmt.Println("\n[6/6] 清理旧备份...")
	cleanupOldBackups(backupDir)
	fmt.Println("✓ 清理完成")

	// 7. 启动新版本
	if *execPath != "" {
		fmt.Println("\n正在启动新版本...")
		time.Sleep(2 * time.Second)
		if err := launchNewVersion(*execPath); err != nil {
			fmt.Printf("警告: 启动新版本失败: %v\n", err)
			fmt.Println("请手动启动应用程序")
		} else {
			fmt.Println("✓ 新版本已启动")
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("       更新完成!")
	fmt.Println("========================================")

	// 自动退出（不等待用户按键）
	time.Sleep(2 * time.Second)
}

// pauseAndExit 暂停并退出（仅用于错误情况）
func pauseAndExit(code int) {
	fmt.Println("\n按任意键退出...")
	time.Sleep(5 * time.Second)
	os.Exit(code)
}
