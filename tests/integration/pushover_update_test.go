package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPushoverUpdateFlow 测试完整的 Pushover 更新流程
// 流程：旧版本安装 → 检测版本 → 更新 → 验证 VERSION 文件
func TestPushoverUpdateFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 1. 设置测试环境
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	// 创建测试项目
	setupTestProject(t, projectPath)

	// 2. 模拟旧版本安装（没有 VERSION 文件）
	installLegacyHook(t, projectPath)

	// 3. 验证旧版本安装成功
	checker := pushover.NewStatusChecker(projectPath)
	installed := checker.CheckInstalled()
	require.True(t, installed, "Hook 应该已安装")

	// 4. 验证 VERSION 文件不存在（旧版本）
	versionFilePath := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook", "VERSION")
	_, err := os.Stat(versionFilePath)
	require.True(t, os.IsNotExist(err), "旧版本不应该有 VERSION 文件")

	// 5. 获取旧版本号（应该返回 unknown 或空）
	oldVersion, err := checker.GetHookVersion()
	require.Error(t, err, "旧版本应该返回错误（VERSION 文件不存在）")
	assert.Equal(t, "", oldVersion, "旧版本号应该为空")

	// 6. 模拟更新流程：创建新版本的 Hook 目录和 VERSION 文件
	// (直接模拟 install.py 的结果，避免 Python 依赖)
	newHookDir := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	err = os.MkdirAll(newHookDir, 0755)
	require.NoError(t, err, "创建新 Hook 目录不应该失败")

	// 创建新版本的 hook 脚本
	newHookPath := filepath.Join(newHookDir, "pushover-notify.py")
	newHookContent := `#!/usr/bin/env python3
# cc-pushover-hook version 1.0.0
print("Hook version 1.0.0 installed")
`
	err = os.WriteFile(newHookPath, []byte(newHookContent), 0755)
	require.NoError(t, err, "写入新 Hook 脚本不应该失败")

	// 创建 VERSION 文件（这是测试的核心）
	versionContent := "version=1.0.0\n"
	err = os.WriteFile(versionFilePath, []byte(versionContent), 0644)
	require.NoError(t, err, "写入 VERSION 文件不应该失败")

	// 删除旧版本的 Hook（模拟更新过程）
	oldHookPath := filepath.Join(projectPath, ".claude", "hooks", "pushover-notify.py")
	err = os.Remove(oldHookPath)
	if err != nil && !os.IsNotExist(err) {
		t.Logf("删除旧 Hook 失败（可能不存在）: %v", err)
	}

	// 7. 验证 VERSION 文件已创建
	versionData, err := os.ReadFile(versionFilePath)
	require.NoError(t, err, "VERSION 文件应该存在")
	versionFileContent := strings.TrimSpace(string(versionData))
	assert.Contains(t, versionFileContent, "version=1.0.0", "VERSION 文件应该包含正确的版本号")

	// 8. 验证新版本号可以正确读取
	newVersion, err := checker.GetHookVersion()
	require.NoError(t, err, "新版本应该能正确读取版本号")
	assert.Equal(t, "1.0.0", newVersion, "版本号应该是 1.0.0")

	// 9. 验证 Hook 仍然正常工作
	status, err := checker.GetStatus("1.0.0")
	require.NoError(t, err, "获取状态不应该失败")
	assert.True(t, status.Installed, "Hook 应该仍然已安装")
	assert.Equal(t, "1.0.0", status.Version, "状态中的版本号应该是 1.0.0")
	assert.False(t, status.UpdateAvailable, "版本相同不应该有可用更新")
}

// TestPushoverVersionComparison 测试版本比较功能
func TestPushoverVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"相等版本", "1.0.0", "1.0.0", 0},
		{"v1 小于 v2", "1.0.0", "1.0.1", -1},
		{"v1 大于 v2", "1.0.1", "1.0.0", 1},
		{"主版本号比较", "2.0.0", "1.9.9", 1},
		{"次版本号比较", "1.2.0", "1.1.9", 1},
		{"补丁版本比较", "1.0.1", "1.0.0", 1},
		{"带预发布标签", "1.0.0-alpha", "1.0.0", -1},
		{"预发布标签比较", "1.0.0-alpha", "1.0.0-beta", -1},
		{"未知版本处理", "unknown", "1.0.0", -1},
		{"空版本处理", "", "1.0.0", -1},
		{"两个未知版本", "unknown", "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pushover.CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result, "版本比较结果不正确")
		})
	}
}

// TestPushoverStatusChecker 测试状态检测器
func TestPushoverStatusChecker(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	// 创建测试项目
	setupTestProject(t, projectPath)

	checker := pushover.NewStatusChecker(projectPath)

	// 1. 测试未安装状态
	installed := checker.CheckInstalled()
	assert.False(t, installed, "未安装时应该返回 false")

	status, err := checker.GetStatus("1.0.0")
	require.NoError(t, err, "获取状态不应该失败")
	assert.False(t, status.Installed, "未安装时状态应该是 false")

	// 2. 安装 Hook
	installHookWithVersion(t, projectPath, "1.0.0")

	// 3. 测试已安装状态
	installed = checker.CheckInstalled()
	assert.True(t, installed, "安装后应该返回 true")

	status, err = checker.GetStatus("1.0.0")
	require.NoError(t, err)
	assert.True(t, status.Installed, "安装后状态应该是 true")
	assert.Equal(t, "1.0.0", status.Version, "版本号应该是 1.0.0")
	assert.NotNil(t, status.InstalledAt, "安装时间不应该为空")
	assert.False(t, status.UpdateAvailable, "版本相同不应该有可用更新")

	// 4. 测试版本比较功能 - 旧版本 Hook
	status, err = checker.GetStatus("1.2.0")
	require.NoError(t, err)
	assert.True(t, status.UpdateAvailable, "Hook 版本旧于扩展版本时应该有可用更新")

	// 5. 测试版本比较功能 - 新版本 Hook
	installHookWithVersion(t, projectPath, "2.0.0")
	status, err = checker.GetStatus("1.2.0")
	require.NoError(t, err)
	assert.False(t, status.UpdateAvailable, "Hook 版本新于扩展版本时不应该有可用更新")
}

// TestPushoverNotificationModes 测试通知模式
func TestPushoverNotificationModes(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")

	// 创建测试项目
	setupTestProject(t, projectPath)

	extensionsPath := filepath.Join(tempDir, "extensions")
	setupMockExtension(t, extensionsPath, "1.0.0")

	service := pushover.NewService(extensionsPath)

	// 安装 Hook
	installHookWithVersion(t, projectPath, "1.0.0")

	// 测试各种通知模式
	modes := []struct {
		mode                     pushover.NotificationMode
		expectedNoPushoverExists bool
		expectedNoWindowsExists  bool
	}{
		{pushover.ModeEnabled, false, false},
		{pushover.ModePushoverOnly, false, true},
		{pushover.ModeWindowsOnly, true, false},
		{pushover.ModeDisabled, true, true},
	}

	for _, tt := range modes {
		t.Run(string(tt.mode), func(t *testing.T) {
			// 设置模式
			err := service.SetNotificationMode(projectPath, tt.mode)
			require.NoError(t, err, "设置通知模式不应该失败")

			// 验证模式
			checker := pushover.NewStatusChecker(projectPath)
			currentMode := checker.GetNotificationMode()
			assert.Equal(t, tt.mode, currentMode, "通知模式应该匹配")

			// 验证标记文件
			noPushoverPath := filepath.Join(projectPath, ".no-pushover")
			noWindowsPath := filepath.Join(projectPath, ".no-windows")

			_, noPushoverErr := os.Stat(noPushoverPath)
			_, noWindowsErr := os.Stat(noWindowsPath)

			noPushoverExists := !os.IsNotExist(noPushoverErr)
			noWindowsExists := !os.IsNotExist(noWindowsErr)

			assert.Equal(t, tt.expectedNoPushoverExists, noPushoverExists, ".no-pushover 文件状态不匹配")
			assert.Equal(t, tt.expectedNoWindowsExists, noWindowsExists, ".no-windows 文件状态不匹配")
		})
	}
}

// setupTestProject 设置测试项目
func setupTestProject(t *testing.T, projectPath string) {
	t.Helper()

	// 创建项目目录
	err := os.MkdirAll(projectPath, 0755)
	require.NoError(t, err, "创建项目目录不应该失败")

	// 初始化 Git 仓库
	helpers.RunGitCmd(t, projectPath, "init")
	helpers.RunGitCmd(t, projectPath, "config", "user.name", "Test User")
	helpers.RunGitCmd(t, projectPath, "config", "user.email", "test@example.com")

	// 创建初始提交
	helpers.WriteFile(t, projectPath, "README.md", "# Test Project\n")
	helpers.RunGitCmd(t, projectPath, "add", ".")
	helpers.RunGitCmd(t, projectPath, "commit", "-m", "initial commit")
}

// installLegacyHook 安装旧版本的 Hook（没有 VERSION 文件）
// 模拟旧版本的安装位置：.claude/hooks/pushover-notify.py
func installLegacyHook(t *testing.T, projectPath string) {
	t.Helper()

	hookDir := filepath.Join(projectPath, ".claude", "hooks")
	err := os.MkdirAll(hookDir, 0755)
	require.NoError(t, err, "创建 hook 目录不应该失败")

	// 创建旧版本的 hook 脚本
	hookContent := `#!/usr/bin/env python3
# Legacy cc-pushover-hook (no VERSION file)
print("Legacy hook installed")
`
	hookPath := filepath.Join(hookDir, "pushover-notify.py")
	err = os.WriteFile(hookPath, []byte(hookContent), 0755)
	require.NoError(t, err, "写入旧版本 hook 脚本不应该失败")
}

// installHookWithVersion 安装带版本号的 Hook（新版本）
func installHookWithVersion(t *testing.T, projectPath, version string) {
	t.Helper()

	hookDir := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	err := os.MkdirAll(hookDir, 0755)
	require.NoError(t, err, "创建 hook 目录不应该失败")

	// 创建新版本的 hook 脚本
	hookContent := `#!/usr/bin/env python3
# cc-pushover-hook version ` + version + `
print("Hook version ` + version + ` installed")
`
	hookPath := filepath.Join(hookDir, "pushover-notify.py")
	err = os.WriteFile(hookPath, []byte(hookContent), 0755)
	require.NoError(t, err, "写入 hook 脚本不应该失败")

	// 创建 VERSION 文件
	versionContent := "version=" + version + "\n"
	versionPath := filepath.Join(hookDir, "VERSION")
	err = os.WriteFile(versionPath, []byte(versionContent), 0644)
	require.NoError(t, err, "写入 VERSION 文件不应该失败")
}

// setupMockExtension 设置模拟的扩展目录
// 包含 install.py 脚本，用于测试更新流程
func setupMockExtension(t *testing.T, extensionsPath, version string) {
	t.Helper()

	extensionPath := filepath.Join(extensionsPath, "cc-pushover-hook")
	err := os.MkdirAll(extensionPath, 0755)
	require.NoError(t, err, "创建扩展目录不应该失败")

	// 创建模拟的 install.py 脚本
	// 这个脚本会：
	// 1. 创建新版本的 hook 目录结构
	// 2. 创建 VERSION 文件
	// 3. 输出 JSON 格式的安装结果
	installScript := `#!/usr/bin/env python3
import sys
import json
import os

# 模拟安装逻辑
target_dir = None
force = False

# 解析命令行参数
args = sys.argv[1:]
i = 0
while i < len(args):
    if args[i] == "--target-dir" and i + 1 < len(args):
        target_dir = args[i + 1]
        i += 2
    elif args[i] == "--force":
        force = True
        i += 1
    elif args[i] == "--non-interactive":
        i += 1
    else:
        i += 1

if not target_dir:
    print(json.dumps({
        "success": False,
        "message": "Missing --target-dir argument"
    }))
    sys.exit(1)

# 创建新的 hook 目录结构（新版本）
hook_dir = os.path.join(target_dir, ".claude", "hooks", "pushover-hook")
os.makedirs(hook_dir, exist_ok=True)

# 创建 hook 脚本
hook_path = os.path.join(hook_dir, "pushover-notify.py")
with open(hook_path, "w") as f:
    f.write("#!/usr/bin/env python3\\n")
    f.write("# cc-pushover-hook version ` + version + `\\n")
    f.write("print('Hook version ` + version + ` installed')\\n")

# 创建 VERSION 文件
version_path = os.path.join(hook_dir, "VERSION")
with open(version_path, "w") as f:
    f.write("version=" + str(` + version + `) + "\\n")

# 输出 JSON 结果
print(json.dumps({
    "success": True,
    "message": "Hook installed successfully",
    "hook_path": hook_dir,
    "version": "` + version + `"
}))
`

	installScriptPath := filepath.Join(extensionPath, "install.py")
	err = os.WriteFile(installScriptPath, []byte(installScript), 0755)
	require.NoError(t, err, "写入 install.py 脚本不应该失败")
}

// cleanupTestProject 清理测试项目
func cleanupTestProject(t *testing.T, projectPath string) {
	t.Helper()
	err := os.RemoveAll(projectPath)
	if err != nil {
		t.Logf("清理测试项目失败: %v", err)
	}
}

// runCommand 运行命令并返回输出
func runCommand(t *testing.T, dir, command string, args ...string) string {
	t.Helper()

	cmd := exec.Command(command, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("命令 %s %v 失败: %v\n输出: %s", command, args, err, string(output))
	}

	return strings.TrimSpace(string(output))
}
