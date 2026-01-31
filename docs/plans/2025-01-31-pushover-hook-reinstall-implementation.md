# Pushover Hook 重装功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 为每个项目添加"重装 Hook"功能，使用 `install.py --reinstall` 参数重新安装项目的 Pushover Hook，同时保留用户的通知配置（`.no-pushover` 和 `.no-windows` 文件）。

**架构：**
- 后端采用分层架构：Installer 层（执行安装脚本） → Service 层（业务逻辑） → App 层（Wails API）
- 前端使用 Vue 3 + Pinia，通过 Wails 绑定调用后端 API
- 重装前保存通知配置，重装后恢复配置，确保用户设置不丢失

**技术栈：**
- 后端：Go 1.21+ + Wails v2
- 前端：Vue 3 + TypeScript + Pinia
- 测试：Go testing + Vue Test Utils

---

## Task 1: 后端 - Installer 层实现

**Files:**
- Modify: `pkg/pushover/installer.go`
- Test: `pkg/pushover/installer_test.go`

### Step 1: 添加辅助方法和类型

在 `installer.go` 中添加以下代码（插入到合适位置，建议在 `findPython` 方法之后）：

```go
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
		if err := os.WriteFile(noPushoverPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("恢复 .no-pushover 失败: %w", err)
		}
	} else {
		os.Remove(noPushoverPath)
	}

	// 恢复 .no-windows
	if config.NoWindowsFile {
		if err := os.WriteFile(noWindowsPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("恢复 .no-windows 失败: %w", err)
		}
	} else {
		os.Remove(noWindowsPath)
	}

	return nil
}
```

### Step 2: 重构现有的解析逻辑

将 `Install` 和 `Update` 方法中的重复解析逻辑提取为独立方法。在 `installer.go` 中添加：

```go
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
```

### Step 3: 实现 Reinstall 方法

在 `installer.go` 中添加 `Reinstall` 方法（插入到 `Update` 方法之后）：

```go
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
	fmt.Fprintf(os.Stderr, "[DEBUG] Reinstalling Hook: %s %v\n", pythonCmd, args)
	fmt.Fprintf(os.Stderr, "[DEBUG] Working dir: %s\n", in.extensionPath)
	fmt.Fprintf(os.Stderr, "[DEBUG] Config: NoPushover=%v, NoWindows=%v\n", config.NoPushoverFile, config.NoWindowsFile)

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
		fmt.Fprintf(os.Stderr, "[WARN] 恢复通知配置失败: %v\n", restoreErr)
	}

	// 解析并返回结果
	hookPath := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook")
	return in.parseInstallResult(output, hookPath)
}
```

### Step 4: 编写单元测试

创建或修改 `pkg/pushover/installer_test.go`，添加测试：

```go
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
	if err := os.WriteFile(noPushoverPath, []byte(""), 0644); err != nil {
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
	if err := os.WriteFile(noWindowsPath, []byte(""), 0644); err != nil {
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
```

### Step 5: 运行测试验证

```bash
cd .worktrees/pushover-reinstall
go test ./pkg/pushover -v -run "TestRead|TestRestore|TestFile"
```

预期输出：
```
=== RUN   TestReadNotificationConfig
--- PASS: TestReadNotificationConfig (0.00s)
=== RUN   TestRestoreNotificationConfig
--- PASS: TestRestoreNotificationConfig (0.00s)
=== RUN   TestFileExists
--- PASS: TestFileExists (0.00s)
PASS
```

### Step 6: 提交后端 Installer 层

```bash
cd .worktrees/pushover-reinstall
git add pkg/pushover/installer.go pkg/pushover/installer_test.go
git commit -m "feat(pushover): 添加 Reinstall 方法和配置保留逻辑

- 添加 readNotificationConfig 方法读取通知配置
- 添加 restoreNotificationConfig 方法恢复通知配置
- 添加 Reinstall 方法支持使用 --reinstall 参数重装
- 添加 parseInstallResult 方法统一解析安装结果
- 添加单元测试验证配置读写逻辑"
```

---

## Task 2: 后端 - Service 层实现

**Files:**
- Modify: `pkg/pushover/service.go`

### Step 1: 添加 ReinstallHook 方法

在 `service.go` 的 `UpdateHook` 方法之后添加：

```go
// ReinstallHook 重装项目的 Hook（保留用户配置）
func (s *Service) ReinstallHook(projectPath string) (*InstallResult, error) {
	// 检查项目路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	// 检查扩展是否已下载
	if !s.repoManager.IsCloned() {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	// 检查项目是否已安装 Hook
	if !s.CheckHookInstalled(projectPath) {
		return nil, fmt.Errorf("项目未安装 Pushover Hook，请先安装")
	}

	return s.installer.Reinstall(projectPath)
}
```

### Step 2: 编写集成测试

创建或修改 `tests/integration/pushover_reinstall_test.go`：

```go
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
)

func TestReinstallHook(t *testing.T) {
	// 获取扩展路径（假设扩展已下载）
	extensionPath := filepath.Join(os.Getenv("APPDATA"), "ai-commit-hub", "extensions", "cc-pushover-hook")

	// 创建测试服务
	service := pushover.NewService(filepath.Join(os.Getenv("APPDATA"), "ai-commit-hub"))

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

	// 初始化 Git 仓库（如果需要）
	// 这里简化处理，实际可能需要 git init

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
	service := pushover.NewService(filepath.Join(os.Getenv("APPDATA"), "ai-commit-hub"))

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
```

### Step 3: 运行集成测试

```bash
cd .worktrees/pushover-reinstall
# 注意：集成测试需要扩展已下载，可能需要先运行克隆操作
go test ./tests/integration -v -run "TestReinstall"
```

预期输出（如果扩展已下载）：
```
=== RUN   TestReinstallHook
--- PASS: TestReinstallHook (0.XXs)
    reinstall_test.go:XX: 重装测试通过，配置已正确保留
=== RUN   TestReinstallHook_NotInstalled
--- PASS: TestReinstallHook_NotInstalled (0.00s)
PASS
```

### Step 4: 提交 Service 层

```bash
cd .worktrees/pushover-reinstall
git add pkg/pushover/service.go tests/integration/pushover_reinstall_test.go
git commit -m "feat(pushover): Service 层添加 ReinstallHook 方法

- 添加 ReinstallHook 方法封装重装逻辑
- 添加项目安装状态检查
- 添加集成测试验证重装和配置保留
- 添加错误场景测试"
```

---

## Task 3: 后端 - App 层实现

**Files:**
- Modify: `app.go`

### Step 1: 添加 ReinstallPushoverHook 方法

在 `app.go` 中找到 `UpdatePushoverHook` 方法，在其后添加：

```go
// ReinstallPushoverHook 重装项目的 Pushover Hook
func (a *App) ReinstallPushoverHook(projectPath string) (*pushover.InstallResult, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	result, err := a.pushoverService.ReinstallHook(projectPath)
	if err != nil {
		logger.Errorf("重装 Pushover Hook 失败: %v", err)
		return nil, err
	}

	if result.Success {
		logger.Infof("成功重装 Pushover Hook 到项目: %s", projectPath)
	}

	return result, nil
}
```

### Step 2: 重新生成 Wails 绑定

```bash
cd .worktrees/pushover-reinstall
wails dev
```

等待启动后，按 `Ctrl+C` 停止。这会生成前端的绑定文件。

### Step 3: 验证绑定已生成

检查 `frontend/wailsjs/go/main/App.js`（或 `.ts`）中是否包含 `ReinstallPushoverHook` 函数：

```bash
cd .worktrees/pushover-reinstall
grep -n "ReinstallPushoverHook" frontend/wailsjs/go/main/*.js
```

预期输出：
```
XX:  export function ReinstallPushoverHook(projectPath) {
```

### Step 4: 提交 App 层

```bash
cd .worktrees/pushover-reinstall
git add app.go frontend/wailsjs/
git commit -m "feat(pushover): App 层添加 ReinstallPushoverHook API

- 导出 ReinstallPushoverHook 方法给前端调用
- 添加错误处理和日志记录
- 重新生成 Wails 绑定"
```

---

## Task 4: 前端 - 类型定义

**Files:**
- Modify: `frontend/src/types/index.ts` 或 `frontend/src/types/pushover.ts`

### Step 1: 确保 InstallResult 类型定义

检查 `frontend/src/types/index.ts` 中是否有 `InstallResult` 定义，如果没有则添加：

```typescript
export interface InstallResult {
  success: boolean
  message: string
  hook_path?: string
}
```

### Step 2: 提交类型定义

```bash
cd .worktrees/pushover-reinstall
git add frontend/src/types/
git commit -m "feat(pushover): 添加 InstallResult 类型定义

- 定义 InstallResult 接口用于重装结果"
```

---

## Task 5: 前端 - PushoverStore 实现

**Files:**
- Modify: `frontend/src/stores/pushoverStore.ts`

### Step 1: 添加 reinstallHook 方法

找到 `pushoverStore.ts` 中的 `updateHook` 方法（如果存在），在其后添加：

```typescript
// 在文件顶部确保导入了正确的类型
import type { InstallResult } from '../types'

async reinstallHook(projectPath: string): Promise<InstallResult> {
  try {
    const result = await ReinstallPushoverHook(projectPath)

    if (result.success) {
      // 刷新项目状态
      await this.refreshProjectStatus(projectPath)
    }

    return result
  } catch (error) {
    console.error('重装 Hook 失败:', error)
    throw error
  }
}

async refreshProjectStatus(projectPath: string) {
  // 刷新 StatusCache 中的项目状态
  const statusCache = useStatusCache()
  await statusCache.refresh(projectPath, { force: true })
}
```

### Step 2: 确保导入了 Wails 生成的函数

检查 `pushoverStore.ts` 顶部是否有正确的导入：

```typescript
import { ReinstallPushoverHook } from '../../wailsjs/go/main/App'
```

如果没有，添加这一行。

### Step 3: 提交 Store 层

```bash
cd .worktrees/pushover-reinstall
git add frontend/src/stores/pushoverStore.ts
git commit -m "feat(pushover): 添加 reinstallHook 方法

- 添加 reinstallHook 方法调用后端 API
- 添加 refreshProjectStatus 方法刷新状态缓存
- 重装成功后自动刷新项目状态"
```

---

## Task 6: 前端 - PushoverStatusRow 组件实现

**Files:**
- Modify: `frontend/src/components/PushoverStatusRow.vue`

### Step 1: 修改模板

找到 `status-right` 部分的按钮组，修改为：

```vue
<div class="status-right">
  <span v-if="isLatest && status?.installed" class="latest-badge">已是最新</span>

  <button
    v-else-if="!status?.installed"
    class="action-btn btn-primary"
    :disabled="loading"
    @click="handleInstall"
  >
    {{ loading ? '处理中...' : '安装 Hook' }}
  </button>

  <!-- 有更新时显示更新按钮 -->
  <button
    v-else-if="needsUpdate"
    class="action-btn btn-update"
    :disabled="loading"
    @click="handleUpdate"
  >
    {{ loading ? '更新中...' : '更新 Hook' }}
  </button>

  <!-- 已是最新版本时显示重装按钮 -->
  <button
    v-else
    class="action-btn btn-reinstall"
    :disabled="loading"
    @click="handleReinstall"
  >
    {{ loading ? '重装中...' : '重装 Hook' }}
  </button>
</div>

<!-- 重装确认对话框 -->
<div v-if="showReinstallDialog" class="dialog-overlay" @click="closeReinstallDialog">
  <div class="dialog-content" @click.stop>
    <h3>重装 Pushover Hook</h3>
    <p class="dialog-description">
      这将重新安装 Pushover Hook 到该项目：
    </p>
    <ul class="dialog-list">
      <li>使用最新版本的 Hook 文件覆盖当前安装</li>
      <li>保留您的通知配置（Pushover/Windows 通知设置）</li>
    </ul>
    <div class="dialog-actions">
      <button
        class="dialog-btn btn-cancel"
        @click="closeReinstallDialog"
      >
        取消
      </button>
      <button
        class="dialog-btn btn-confirm"
        :disabled="loading"
        @click="confirmReinstall"
      >
        确定重装
      </button>
    </div>
  </div>
</div>
```

### Step 2: 修改脚本

在 `<script setup>` 部分添加：

```typescript
const showReinstallDialog = ref(false)

function handleReinstall() {
  showReinstallDialog.value = true
}

function closeReinstallDialog() {
  showReinstallDialog.value = false
}

async function confirmReinstall() {
  if (localLoading.value) return

  localLoading.value = true
  try {
    const result = await pushoverStore.reinstallHook(props.projectPath)

    if (result.success) {
      // 关闭对话框
      closeReinstallDialog()
      // 可选：显示成功提示
      console.log('重装成功:', result.message)
    } else {
      // 显示错误信息
      console.error('重装失败:', result.message)
    }
  } catch (e) {
    console.error('重装 Hook 失败:', e)
  } finally {
    localLoading.value = false
  }
}
```

### Step 3: 添加样式

在 `<style scoped>` 部分添加：

```css
/* 重装按钮样式 */
.btn-reinstall {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.btn-reinstall:hover:not(:disabled) {
  background: rgba(6, 182, 212, 0.25);
}

/* 对话框样式 */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  max-width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.dialog-content h3 {
  margin: 0 0 var(--space-md) 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.dialog-description {
  margin: 0 0 var(--space-sm) 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-list {
  margin: 0 0 var(--space-md) 0;
  padding-left: var(--space-lg);
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-list li {
  margin-bottom: var(--space-xs);
}

.dialog-actions {
  display: flex;
  gap: var(--space-sm);
  justify-content: flex-end;
}

.dialog-btn {
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-fast);
}

.dialog-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.btn-confirm {
  background: var(--accent-primary);
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  background: var(--accent-secondary);
}
```

### Step 4: 提交前端组件

```bash
cd .worktrees/pushover-reinstall
git add frontend/src/components/PushoverStatusRow.vue
git commit -m "feat(pushover): 添加重装 Hook 按钮和确认对话框

- 添加重装按钮（仅在已是最新版本时显示）
- 添加确认对话框说明重装操作
- 保留用户通知配置
- 添加加载状态和错误处理"
```

---

## Task 7: 前端测试和验证

**Files:**
- Test: 手动测试

### Step 1: 启动开发服务器

```bash
cd .worktrees/pushover-reinstall
wails dev
```

### Step 2: 手动测试流程

1. **准备测试环境**
   - 确保 cc-pushover-hook 扩展已下载
   - 选择一个已安装 Hook 的项目

2. **测试重装功能**
   - 打开项目列表
   - 找到 Pushover Hook 状态行
   - 确认显示"重装 Hook"按钮（已是最新版本时）
   - 修改通知配置（如禁用 Pushover）
   - 点击"重装 Hook"按钮

3. **验证对话框**
   - 确认对话框显示正确
   - 检查文案是否清晰
   - 点击"取消"关闭对话框
   - 再次点击"重装 Hook"
   - 点击"确定重装"

4. **验证结果**
   - 按钮显示"重装中..."
   - 操作完成后恢复为"重装 Hook"
   - 通知配置保持不变
   - Hook 文件已更新

5. **测试错误场景**
   - 扩展未下载时的错误提示
   - Python 未安装时的错误提示
   - 网络错误时的处理

### Step 3: 记录测试结果

如果测试通过，继续；如果发现问题，修复并重新测试。

### Step 4: 最终提交

```bash
cd .worktrees/pushover-reinstall
git add .
git commit -m "test(pushover): 完成重装功能手动测试

- 验证重装流程正常
- 验证配置保留功能正常
- 验证错误处理正常"
```

---

## Task 8: 文档和清理

**Files:**
- Create: `tmp/重装功能测试.md`（可选）

### Step 1: 更新 CLAUDE.md（如需要）

如果有重要的使用说明，在 `CLAUDE.md` 中添加：

```markdown
### Pushover Hook 重装

项目支持重装 Pushover Hook 功能：
- 只在 Hook 已安装且已是最新版本时显示"重装 Hook"按钮
- 重装会保留用户的通知配置（.no-pushover、.no-windows）
- 使用 install.py --reinstall 参数执行重装
```

### Step 2: 清理临时文件

```bash
cd .worktrees/pushover-reinstall
rm -f tmp/*.go
```

### Step 3: 最终验证

```bash
cd .worktrees/pushover-reinstall
go test ./... -v
```

确保所有测试通过。

### Step 4: 提交文档更新

```bash
cd .worktrees/pushover-reinstall
git add CLAUDE.md docs/
git commit -m "docs(pushover): 更新文档说明重装功能

- 添加重装功能使用说明
- 添加配置保留说明"
```

---

## 验收标准

完成所有任务后，以下条件应全部满足：

1. ✅ 后端 `Reinstall` 方法正确实现并保留配置
2. ✅ 前端显示"重装 Hook"按钮（已是最新版本时）
3. ✅ 点击按钮显示确认对话框
4. ✅ 确认后执行重装并保留用户配置
5. ✅ 重装成功后刷新项目状态
6. ✅ 所有单元测试和集成测试通过
7. ✅ 手动测试验证所有功能正常
8. ✅ 代码已提交到 `feature/pushover-hook-reinstall` 分支

---

## 常见问题

### Q1: Wails 绑定未生成

**问题：** 前端调用 `ReinstallPushoverHook` 报错 "function not defined"

**解决：** 运行 `wails dev` 重新生成绑定

### Q2: 配置未保留

**问题：** 重装后通知配置丢失

**解决：** 检查 `readNotificationConfig` 和 `restoreNotificationConfig` 的实现逻辑

### Q3: 测试失败

**问题：** 集成测试失败 "扩展未下载"

**解决：** 先运行 `python install.py` 下载扩展，或跳过集成测试

---

**计划版本：** 1.0
**创建日期：** 2025-01-31
**预计工作量：** 2-3 小时
