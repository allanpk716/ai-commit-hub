package pushover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadNotificationConfig(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pushover-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	installer := NewInstaller("")

	// 测试：没有配置文件
	config := installer.readNotificationConfig(tempDir)
	if config.NoPushoverFile || config.NoWindowsFile {
		t.Errorf("空目录应该没有任何配置文件")
	}

	// 测试：只有 .no-pushover
	noPushoverPath := filepath.Join(tempDir, ".no-pushover")
	if err := os.WriteFile(noPushoverPath, []byte(""), 0o644); err != nil {
		t.Fatalf("创建 .no-pushover 失败: %v", err)
	}

	config = installer.readNotificationConfig(tempDir)
	if !config.NoPushoverFile {
		t.Errorf("应该检测到 .no-pushover 文件")
	}
	if config.NoWindowsFile {
		t.Errorf("不应该检测到 .no-windows 文件")
	}

	// 测试：两个文件都存在
	noWindowsPath := filepath.Join(tempDir, ".no-windows")
	if err := os.WriteFile(noWindowsPath, []byte(""), 0o644); err != nil {
		t.Fatalf("创建 .no-windows 失败: %v", err)
	}

	config = installer.readNotificationConfig(tempDir)
	if !config.NoPushoverFile || !config.NoWindowsFile {
		t.Errorf("应该检测到两个配置文件")
	}
}

func TestRestoreNotificationConfig(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pushover-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	installer := NewInstaller("")

	// 测试：恢复到无配置状态
	config := NotificationConfig{NoPushoverFile: false, NoWindowsFile: false}
	if err := installer.restoreNotificationConfig(tempDir, config); err != nil {
		t.Fatalf("恢复配置失败: %v", err)
	}

	noPushoverPath := filepath.Join(tempDir, ".no-pushover")
	noWindowsPath := filepath.Join(tempDir, ".no-windows")

	if fileExists(noPushoverPath) || fileExists(noWindowsPath) {
		t.Errorf("不应该创建任何配置文件")
	}

	// 测试：恢复到全部启用状态
	config = NotificationConfig{NoPushoverFile: true, NoWindowsFile: true}
	if err := installer.restoreNotificationConfig(tempDir, config); err != nil {
		t.Fatalf("恢复配置失败: %v", err)
	}

	if !fileExists(noPushoverPath) || !fileExists(noWindowsPath) {
		t.Errorf("应该创建两个配置文件")
	}

	// 测试：从全部启用切换到只有 Pushover
	// 先创建两个文件
	config = NotificationConfig{NoPushoverFile: true, NoWindowsFile: true}
	installer.restoreNotificationConfig(tempDir, config)

	// 切换到只有 .no-pushover
	config = NotificationConfig{NoPushoverFile: true, NoWindowsFile: false}
	if err := installer.restoreNotificationConfig(tempDir, config); err != nil {
		t.Fatalf("恢复配置失败: %v", err)
	}

	if !fileExists(noPushoverPath) {
		t.Errorf(".no-pushover 应该存在")
	}
	if fileExists(noWindowsPath) {
		t.Errorf(".no-windows 不应该存在")
	}
}

func TestFileExists(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "pushover-test-*")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if !fileExists(tempFile.Name()) {
		t.Errorf("文件应该存在")
	}

	if fileExists("/nonexistent/path/12345") {
		t.Errorf("不存在的路径应该返回 false")
	}
}
