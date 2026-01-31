package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
)

func TestReinstallHook(t *testing.T) {
	// 获取应用数据目录
	appDataDir := filepath.Join(os.Getenv("APPDATA"), "ai-commit-hub")

	// 创建测试服务
	service := pushover.NewService(appDataDir)

	// 确保扩展已下载
	if !service.IsExtensionDownloaded() {
		t.Skip("扩展未下载，跳过集成测试")
	}

	// 创建临时项目目录
	tempProject, err := os.MkdirTemp("", "pushover-test-project-*")
	if err != nil {
		t.Fatalf("创建临时项目失败: %v", err)
	}
	defer os.RemoveAll(tempProject)

	// 1. 先安装 Hook
	installResult, err := service.InstallHook(tempProject, false)
	if err != nil {
		t.Fatalf("安装 Hook 失败: %v", err)
	}
	if !installResult.Success {
		t.Fatalf("安装 Hook 失败: %s", installResult.Message)
	}

	// 2. 修改通知配置（禁用 Pushover）
	noPushoverPath := filepath.Join(tempProject, ".no-pushover")
	if err := os.WriteFile(noPushoverPath, []byte(""), 0644); err != nil {
		t.Fatalf("创建 .no-pushover 失败: %v", err)
	}

	// 3. 验证配置已创建
	if _, err := os.Stat(noPushoverPath); os.IsNotExist(err) {
		t.Fatal(".no-pushover 文件应该存在")
	}

	// 4. 重装 Hook
	reinstallResult, err := service.ReinstallHook(tempProject)
	if err != nil {
		t.Fatalf("重装 Hook 失败: %v", err)
	}
	if !reinstallResult.Success {
		t.Fatalf("重装 Hook 失败: %s", reinstallResult.Message)
	}

	// 5. 验证配置已保留
	if _, err := os.Stat(noPushoverPath); os.IsNotExist(err) {
		t.Fatal("重装后 .no-pushover 文件应该仍然存在")
	}

	t.Log("重装测试通过，配置已正确保留")
}

func TestReinstallHook_NotInstalled(t *testing.T) {
	appDataDir := filepath.Join(os.Getenv("APPDATA"), "ai-commit-hub")
	service := pushover.NewService(appDataDir)

	// 确保扩展已下载，否则跳过测试
	if !service.IsExtensionDownloaded() {
		t.Skip("扩展未下载，跳过集成测试")
	}

	// 创建临时项目目录（未安装 Hook）
	tempProject, err := os.MkdirTemp("", "pushover-test-project-*")
	if err != nil {
		t.Fatalf("创建临时项目失败: %v", err)
	}
	defer os.RemoveAll(tempProject)

	// 尝试重装未安装的 Hook
	_, err = service.ReinstallHook(tempProject)
	if err == nil {
		t.Fatal("应该返回错误：项目未安装 Hook")
	}

	expectedError := "项目未安装 Pushover Hook"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配，期望: %s, 实际: %s", expectedError, err.Error())
	}
}
