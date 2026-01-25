# Pushover Hook Version Management and Auto-Update Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement automatic version detection and update management for cc-pushover-hook extension, including startup auto-download, version comparison, and UI update prompts.

**Architecture:**
- Backend: Go service manages local git cache of cc-pushover-hook repository
- Version Detection: Read VERSION file in project, fallback to "unknown" for legacy installs
- UI: Vue3 components display version status and update prompts
- Update Flow: Git pull for cache, re-run install.py for project updates

**Tech Stack:**
- Go 1.21+ with git commands via os/exec
- Vue 3 + TypeScript + Pinia
- Wails v2 for Go-JS bridge
- Python install.py in cc-pushover-hook repo

---

## Part 1: cc-pushover-hook Repository Improvements

### Task 1: Modify install.py to Dynamically Get Version

**Files:**
- Modify: `C:\WorkSpace\agent\cc-pushover-hook\install.py:37`
- Test: `C:\WorkSpace\agent\cc-pushover-hook\test\test_install.py` (create)

**Step 1: Add version detection function**

```python
def get_version_from_git(self) -> str:
    """Get version from git tag or commit hash."""
    try:
        import subprocess
        result = subprocess.run(
            ["git", "describe", "--tags", "--always"],
            cwd=self.script_dir,
            capture_output=True,
            text=True,
            check=False
        )
        if result.returncode == 0:
            return result.stdout.strip().lstrip('v')
        # Fallback to commit hash
        result = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            cwd=self.script_dir,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()
    except Exception:
        return self.VERSION  # Fallback to hardcoded version
```

**Step 2: Update VERSION constant to be computed**

```python
class Installer:
    VERSION = "1.0.0"  # Fallback version

    def __init__(self, args=None):
        self.platform = system()
        self.script_dir = Path(__file__).parent.resolve()
        self.target_dir = None
        self.hook_dir = None
        self.args = args
        self.version = self.get_version_from_git()  # Dynamic version
```

**Step 3: Test version detection**

Run: `cd C:\WorkSpace\agent\cc-pushover-hook && python -c "from install import Installer; i = Installer([]); print(i.get_version_from_git())"`
Expected: `1.3.0` (or current tag)

**Step 4: Update JSON output to use dynamic version**

Modify line 579 in `run()` method:
```python
result = {
    "status": "success",
    "hook_path": str(self.hook_dir),
    "version": self.version  # Use dynamic version
}
```

**Step 5: Commit cc-pushover-hook changes**

```bash
cd C:\WorkSpace\agent\cc-pushover-hook
git add install.py
git commit -m "feat: dynamically get version from git tags

- Add get_version_from_git() method
- VERSION constant as fallback
- Use dynamic version in install output
"
```

---

### Task 2: Create VERSION File on Install

**Files:**
- Modify: `C:\WorkSpace\agent\cc-pushover-hook\install.py:223-290` (copy_hook_files method)

**Step 1: Add create_version_file() method**

After `copy_hook_files()` method, add:

```python
def create_version_file(self) -> None:
    """Create VERSION file in hook directory."""
    from datetime import datetime
    import subprocess

    try:
        # Get git commit hash
        result = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=self.script_dir,
            capture_output=True,
            text=True,
            check=True
        )
        git_commit = result.stdout.strip()

        version_file = self.hook_dir / "VERSION"
        installed_at = datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")

        content = f"version={self.version}\n"
        content += f"installed_at={installed_at}\n"
        content += f"git_commit={git_commit}\n"

        with open(version_file, 'w', encoding='utf-8') as f:
            f.write(content)

        self.print_info(f"[OK] Created: VERSION file")
    except Exception as e:
        self.print_info(f"[WARN] Failed to create VERSION file: {e}")
```

**Step 2: Call create_version_file in run() method**

Modify `run()` method after `copy_hook_files()`:

```python
def run(self) -> None:
    """Run the full installation process."""
    try:
        self.print_banner()
        self.target_dir = self.get_target_directory()
        self.create_hook_directory()
        self.copy_hook_files()
        self.create_version_file()  # ADD THIS LINE
        self.generate_settings_json()
        # ... rest of method
```

**Step 3: Test VERSION file creation**

Run: `python install.py --target-dir /tmp/test-project --non-interactive --skip-diagnostics`
Check: `cat /tmp/test-project/.claude/hooks/pushover-hook/VERSION`
Expected:
```
version=1.3.0
installed_at=2026-01-25T10:30:00Z
git_commit=abc123def456
```

**Step 4: Commit changes**

```bash
cd C:\WorkSpace\agent\cc-pushover-hook
git add install.py
git commit -m "feat: create VERSION file on installation

- Add create_version_file() method
- Write version, install time, git commit to VERSION
- Call from run() after copying files
"
```

---

## Part 2: ai-commit-hub Backend Improvements

### Task 3: Auto-Download Extension on Startup

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go:104-110`

**Step 1: Add auto-clone logic in startup()**

Replace existing pushover initialization with:

```go
// Initialize pushover service
execPath, err := os.Executable()
if err != nil {
	fmt.Printf("获取可执行文件路径失败: %v\n", err)
} else {
	appPath := filepath.Dir(execPath)
	a.pushoverService = pushover.NewService(appPath)

	// Auto-clone extension if not exists
	if a.pushoverService != nil && !a.pushoverService.IsExtensionDownloaded() {
		fmt.Println("cc-pushover-hook 扩展未下载，正在自动下载...")
		if err := a.pushoverService.CloneExtension(); err != nil {
			fmt.Printf("自动下载扩展失败: %v\n", err)
			fmt.Println("请稍后在设置中手动下载扩展")
		} else {
			fmt.Println("cc-pushover-hook 扩展下载成功")
		}
	}
}
```

**Step 2: Test auto-clone functionality**

Rename extension directory temporarily:
```bash
mv "C:\Users\allan716\.ai-commit-hub\extensions\cc-pushover-hook" "C:\Users\allan716\.ai-commit-hub\extensions\cc-pushover-hook.bak"
```

Run application and check log:
Expected:
```
cc-pushover-hook 扩展未下载，正在自动下载...
cc-pushover-hook 扩展下载成功
```

**Step 3: Restore and test with existing extension**

```bash
mv "C:\Users\allan716\.ai-commit-hub\extensions\cc-pushover-hook.bak" "C:\Users\allan716\.ai-commit-hub\extensions\cc-pushover-hook"
```

**Step 4: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add app.go
git commit -m "feat: auto-download cc-pushover-hook extension on startup

- Check if extension exists on startup
- Auto-clone from GitHub if not found
- Log success or error message
"
```

---

### Task 4: Read VERSION File in Status Checker

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\pushover\status.go:78-120`

**Step 1: Replace GetHookVersion() implementation**

Replace entire method with:

```go
// GetHookVersion 获取 Hook 版本
func (sc *StatusChecker) GetHookVersion() (string, error) {
	// Try to read VERSION file first
	versionFile := filepath.Join(sc.projectPath, ".claude", "hooks", "pushover-hook", "VERSION")
	if data, err := os.ReadFile(versionFile); err == nil {
		// Parse VERSION file
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "version=") {
				version := strings.TrimPrefix(line, "version=")
				return version, nil
			}
		}
	}

	// VERSION file doesn't exist - legacy installation
	// Return special marker to indicate unknown version
	fmt.Printf("[DEBUG] VERSION file not found, likely legacy installation\n")
	return "", fmt.Errorf("VERSION file not found (legacy installation)")
}
```

**Step 2: Update GetStatus() to handle unknown version**

Modify method around line 152:

```go
mode := sc.GetNotificationMode()
version, err := sc.GetHookVersion()
installedAt, _ := sc.GetInstalledAt()

// Handle legacy installations without VERSION file
if err != nil {
	// Legacy installation - version is unknown
	version = "unknown"
}

return &HookStatus{
	Installed:   true,
	Mode:        mode,
	Version:     version,
	InstalledAt: installedAt,
}, nil
```

**Step 3: Test with new installation**

Run install.py on a test project:
```bash
cd C:\WorkSpace\agent\cc-pushover-hook
python install.py --target-dir /tmp/test-new --non-interactive --skip-diagnostics
```

Check version reading:
Expected: VERSION file exists and returns correct version

**Step 4: Test with legacy installation**

Create a project without VERSION file:
```bash
mkdir -p /tmp/test-legacy/.claude/hooks/pushover-hook
echo "# legacy file" > /tmp/test-legacy/.claude/hooks/pushover-hook/pushover-notify.py
```

Check version reading:
Expected: Returns "unknown" without error

**Step 5: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add pkg/pushover/status.go
git commit -m "feat: read VERSION file for hook version detection

- Parse VERSION file to get installed version
- Return 'unknown' for legacy installations
- Handle missing VERSION file gracefully
"
```

---

### Task 5: Add Version Comparison to Extension Info

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\pushover\service.go`
- Create: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\pushover\version.go`

**Step 1: Create version.go for comparison logic**

```go
package pushover

import (
	"fmt"
	"strconv"
	"strings"
)

// CompareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	// Handle unknown versions
	if v1 == "unknown" {
		return -1, nil // Unknown is treated as older
	}
	if v2 == "unknown" {
		return 1, nil
	}

	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split by dots
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int

		if i < len(parts1) {
			n, err := strconv.Atoi(parts1[i])
			if err != nil {
				return 0, fmt.Errorf("invalid version format: %s", v1)
			}
			num1 = n
		}

		if i < len(parts2) {
			n, err := strconv.Atoi(parts2[i])
			if err != nil {
				return 0, fmt.Errorf("invalid version format: %s", v2)
			}
			num2 = n
		}

		if num1 < num2 {
			return -1, nil
		}
		if num1 > num2 {
			return 1, nil
		}
	}

	return 0, nil
}
```

**Step 2: Add CheckForUpdates method to Service**

In `pkg/pushover/service.go`, add:

```go
// CheckForUpdates 检查项目中的 Hook 是否有更新可用
func (s *Service) CheckForUpdates(projectPath string) (bool, string, string, error) {
	// Get installed hook version
	checker := NewStatusChecker(projectPath)
	status, err := checker.GetStatus()
	if err != nil {
		return false, "", "", fmt.Errorf("获取 Hook 状态失败: %w", err)
	}

	if !status.Installed {
		return false, "", "", fmt.Errorf("Hook 未安装")
	}

	// Get latest version from cache
	latestVersion, err := s.repoManager.GetVersion()
	if err != nil {
		return false, "", "", fmt.Errorf("获取最新版本失败: %w", err)
	}

	// Compare versions
	cmp, err := CompareVersions(status.Version, latestVersion)
	if err != nil {
		return false, "", "", fmt.Errorf("版本比较失败: %w", err)
	}

	updateAvailable := cmp < 0
	return updateAvailable, status.Version, latestVersion, nil
}
```

**Step 3: Add App API method**

In `app.go`, add:

```go
// CheckPushoverUpdates 检查项目的 Pushover Hook 更新
func (a *App) CheckPushoverUpdates(projectPath string) (map[string]interface{}, error) {
	if a.initError != nil {
		return nil, a.initError
	}
	if a.pushoverService == nil {
		return nil, fmt.Errorf("pushover service 未初始化")
	}

	updateAvailable, currentVersion, latestVersion, err := a.pushoverService.CheckForUpdates(projectPath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"update_available": updateAvailable,
		"current_version":  currentVersion,
		"latest_version":   latestVersion,
	}, nil
}
```

**Step 4: Test version comparison**

Run: `go test ./pkg/pushover -v -run TestCompareVersions` (need to create test)

**Step 5: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add pkg/pushover/version.go pkg/pushover/service.go app.go
git commit -m "feat: add version comparison and update checking

- Add CompareVersions() utility
- Add CheckForUpdates() to Service
- Add CheckPushoverUpdates() API method
- Handle 'unknown' version for legacy installs
"
```

---

## Part 3: Frontend UI Improvements

### Task 6: Add Update Check to Pushover Store

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\stores\pushoverStore.ts`

**Step 1: Add checkForUpdates action**

```typescript
/**
 * 检查项目的 Hook 更新
 */
async function checkForUpdates(projectPath: string): Promise<{
  updateAvailable: boolean
  currentVersion: string
  latestVersion: string
} | null> {
  try {
    const result = await CheckPushoverUpdates(projectPath)
    return {
      updateAvailable: result.update_available as boolean,
      currentVersion: result.current_version as string,
      latestVersion: result.latest_version as string
    }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '检查更新失败'
    error.value = `检查更新失败: ${message}`
    return null
  }
}
```

**Step 2: Export new action**

In the return statement, add:
```typescript
  return {
    // ... existing exports
    checkForUpdates
  }
```

**Step 3: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add frontend/src/stores/pushoverStore.ts
git commit -m "feat: add checkForUpdates to pushover store

- Add checkForUpdates action
- Call CheckPushoverUpdates API
- Return update status and versions
"
```

---

### Task 7: Update PushoverStatusCard Component

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\components\PushoverStatusCard.vue`

**Step 1: Add update check state and UI**

Add after `error` ref:
```typescript
const updateInfo = ref<{
  updateAvailable: boolean
  currentVersion: string
  latestVersion: string
} | null>(null)
```

**Step 2: Add check for updates on mount**

```typescript
onMounted(async () => {
  if (props.projectPath) {
    const info = await pushoverStore.checkForUpdates(props.projectPath)
    if (info) {
      updateInfo.value = info
    }
  }
})
```

**Step 3: Add update prompt UI**

In template, after mode selector, add:
```vue
<div v-if="updateInfo?.updateAvailable" class="update-prompt">
  <p class="update-message">
    ⚠️ 有新版本可用 ({{ updateInfo.currentVersion }} → {{ updateInfo.latestVersion }})
  </p>
  <button
    class="btn btn-update"
    :disabled="loading"
    @click="handleUpdateHook"
  >
    更新到 {{ updateInfo.latestVersion }}
  </button>
</div>

<div v-else-if="status?.version === 'unknown'" class="update-prompt">
  <p class="update-message">
    ⚠️ 版本未知（可能是旧版本）
  </p>
  <button
    class="btn btn-update"
    :disabled="loading"
    @click="handleUpdateHook"
  >
    更新到最新版本
  </button>
</div>
```

**Step 4: Add styles**

```css
.update-prompt {
  margin-top: var(--space-md);
  padding: var(--space-sm);
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: var(--radius-sm);
}

.update-message {
  margin: 0 0 var(--space-sm) 0;
  color: #fbbf24;
  font-size: 13px;
}

.btn-update {
  background: #fbbf24;
  color: #000;
}

.btn-update:hover:not(:disabled) {
  background: #f59e0b;
}
```

**Step 5: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add frontend/src/components/PushoverStatusCard.vue
git commit -m "feat: show update prompt in PushoverStatusCard

- Display update available message
- Show current and latest versions
- Add update button for legacy installs
"
```

---

### Task 8: Update PushoverManagementPanel Component

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\components\PushoverManagementPanel.vue`

**Step 1: Add extension version display**

In the extension status section, add:
```vue
<div class="extension-info">
  <div class="info-row">
    <span class="info-label">当前版本:</span>
    <span class="info-value">{{ extensionInfo.version || '未知' }}</span>
  </div>
  <div class="info-row">
    <span class="info-label">最新版本:</span>
    <span class="info-value">{{ extensionInfo.latest_version || '未知' }}</span>
  </div>
  <div v-if="extensionInfo.update_available" class="update-badge">
    有更新可用
  </div>
</div>
```

**Step 2: Add update buttons**

```vue
<div class="extension-actions">
  <button
    class="btn btn-secondary"
    :disabled="loading"
    @click="handleCheckUpdates"
  >
    检查更新
  </button>
  <button
    v-if="extensionInfo.update_available"
    class="btn btn-primary"
    :disabled="loading"
    @click="handleUpdateExtension"
  >
    更新扩展
  </button>
</div>
```

**Step 3: Add handler methods**

```typescript
async function handleCheckUpdates() {
  await pushoverStore.checkExtensionStatus()
}

async function handleUpdateExtension() {
  try {
    await pushoverStore.updateExtension()
    await pushoverStore.checkExtensionStatus()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '更新失败'
    error.value = `更新扩展失败: ${message}`
  }
}
```

**Step 4: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add frontend/src/components/PushoverManagementPanel.vue
git commit -m "feat: show extension version and update controls

- Display current and latest versions
- Add check updates button
- Add update extension button
- Show update badge when available
"
```

---

## Part 4: Update Hook Implementation

### Task 9: Implement Update Hook Functionality

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\pushover\installer.go`

**Step 1: Add Update method to Installer**

```go
// Update 更新项目中已安装的 Hook
func (i *Installer) Update(projectPath string) (*InstallResult, error) {
	// Check if extension is downloaded
	if !i.repoManager.IsCloned() {
		return nil, fmt.Errorf("扩展未下载，请先下载扩展")
	}

	// Get extension path
	extensionPath := i.repoManager.GetExtensionPath()
	installScript := filepath.Join(extensionPath, "install.py")

	// Check if install script exists
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return nil, fmt.Errorf("install.py 不存在: %s", installScript)
	}

	// Run install.py with --force flag to overwrite
	cmd := exec.Command("python", "install.py",
		"--target-dir", projectPath,
		"--force",
		"--non-interactive",
		"--skip-diagnostics",
		"--quiet")

	cmd.Dir = extensionPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("更新失败: %v\n输出: %s", err, string(output)),
		}, nil
	}

	// Parse JSON output
	var result InstallResult
	if err := json.Unmarshal(output, &result); err != nil {
		return &InstallResult{
			Success: true,
			Message: "更新完成（无法解析版本信息）",
		}, nil
	}

	return &result, nil
}
```

**Step 2: Add Service method**

In `service.go`, add:
```go
// UpdateHook 更新项目的 Hook
func (s *Service) UpdateHook(projectPath string) (*InstallResult, error) {
	return s.installer.Update(projectPath)
}
```

**Step 3: Add App API method**

In `app.go`, add:
```go
// UpdatePushoverHook 更新项目的 Pushover Hook
func (a *App) UpdatePushoverHook(projectPath string) (*pushover.InstallResult, error) {
	if a.initError != nil {
		return &pushover.InstallResult{Success: false, Message: a.initError.Error()}, nil
	}
	if a.pushoverService == nil {
		return &pushover.InstallResult{Success: false, Message: "pushover service 未初始化"}, nil
	}

	result, err := a.pushoverService.UpdateHook(projectPath)
	if err != nil {
		return result, err
	}

	// Update successful - sync database status
	if result.Success {
		if syncErr := a.syncProjectHookStatusByPath(projectPath); syncErr != nil {
			fmt.Printf("同步 Hook 状态失败: %v\n", syncErr)
		}
	}

	return result, nil
}
```

**Step 4: Test update functionality**

Run: Test updating a legacy installation

**Step 5: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add pkg/pushover/installer.go pkg/pushover/service.go app.go
git commit -m "feat: implement update hook functionality

- Add Update() method to Installer
- Run install.py with --force flag
- Update database after successful update
- Add UpdatePushoverHook API
"
```

---

### Task 10: Frontend Update Hook Handler

**Files:**
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\stores\pushoverStore.ts`
- Modify: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\components\PushoverStatusCard.vue`

**Step 1: Add updateHook to pushoverStore**

```typescript
/**
 * 更新项目的 Hook
 */
async function updateHook(projectPath: string): Promise<InstallResult> {
  loading.value = true
  error.value = null

  try {
    const result = await UpdatePushoverHook(projectPath)
    if (result && result.success) {
      // Refresh project status
      await getProjectHookStatus(projectPath)
    }
    return result || { success: false, message: '更新失败' }
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    error.value = `更新 Hook 失败: ${message}`
    return { success: false, message }
  } finally {
    loading.value = false
  }
}
```

**Step 2: Export updateHook**

```typescript
  return {
    // ... existing exports
    updateHook
  }
```

**Step 3: Add handler in PushoverStatusCard**

```typescript
async function handleUpdateHook() {
  error.value = null
  loading.value = true

  try {
    const result = await pushoverStore.updateHook(props.projectPath)
    if (!result.success) {
      error.value = result.message || '更新失败'
    } else {
      // Refresh update info
      const info = await pushoverStore.checkForUpdates(props.projectPath)
      if (info) {
        updateInfo.value = info
      }
    }
  } catch (e: unknown) {
    error.value = '更新失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
```

**Step 4: Add Wailsjs binding**

Run: `wails generate module`

**Step 5: Test update flow**

1. Create legacy installation without VERSION
2. Open app, see update prompt
3. Click update button
4. Verify VERSION file created

**Step 6: Commit changes**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git add frontend/src/stores/pushoverStore.ts frontend/src/components/PushoverStatusCard.vue
wails generate module
git add wailsjs
git commit -m "feat: add update hook functionality to UI

- Add updateHook to pushoverStore
- Add handleUpdateHook in PushoverStatusCard
- Refresh status after update
- Regenerate Wails bindings
"
```

---

## Part 5: Testing and Documentation

### Task 11: Integration Testing

**Files:**
- Create: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\tests\integration\pushover_update_test.go`

**Step 1: Create integration test**

```go
package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPushoverUpdateFlow(t *testing.T) {
	// Test: legacy installation → update → VERSION file exists
	projectPath := setupTestProject(t)
	defer cleanupTestProject(t, projectPath)

	// 1. Verify legacy installation (no VERSION)
	versionFile := filepath.Join(projectPath, ".claude", "hooks", "pushover-hook", "VERSION")
	if _, err := os.Stat(versionFile); !os.IsNotExist(err) {
		t.Fatal("VERSION file should not exist in legacy installation")
	}

	// 2. Run update
	result, err := pushoverService.UpdateHook(projectPath)
	if err != nil {
		t.Fatalf("UpdateHook failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("UpdateHook returned success=false: %s", result.Message)
	}

	// 3. Verify VERSION file created
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		t.Fatal("VERSION file should exist after update")
	}

	// 4. Verify version is detected
	checker := pushover.NewStatusChecker(projectPath)
	version, err := checker.GetHookVersion()
	if err != nil {
		t.Fatalf("GetHookVersion failed: %v", err)
	}
	if version == "" || version == "unknown" {
		t.Fatal("Version should be valid after update")
	}
}
```

**Step 2: Run tests**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
go test ./tests/integration -v -run TestPushoverUpdateFlow
```

**Step 3: Commit test**

```bash
git add tests/integration/pushover_update_test.go
git commit -m "test: add integration test for update flow

- Test legacy installation update
- Verify VERSION file creation
- Verify version detection after update
"
```

---

### Task 12: Update Documentation

**Files:**
- Create: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\pushover-version-management.md`

**Step 1: Create documentation**

```markdown
# Pushover Hook Version Management

## Overview

The application automatically manages cc-pushover-hook extension versions, including:

- **Auto-download**: Extension is automatically downloaded on first startup
- **Version detection**: Reads VERSION file in project for installed version
- **Legacy support**: Handles installations without VERSION file
- **Update prompts**: Notifies users when updates are available
- **One-click updates**: Update hooks directly from UI

## Version File Format

Installed hooks store version in `.claude/hooks/pushover-hook/VERSION`:

```
version=v1.3.0
installed_at=2026-01-25T10:30:00Z
git_commit=abc123def456
```

## Update Flow

1. **Check for updates**: Compares installed version with cached extension
2. **Update extension**: Runs `git pull` on cached repository
3. **Update project hook**: Re-runs install.py with `--force` flag
4. **Verify**: VERSION file updated to latest version

## API Methods

### CheckPushoverUpdates(projectPath)
Returns update availability and versions.

### UpdatePushoverHook(projectPath)
Updates project hook to latest version.

### UpdatePushoverExtension()
Updates cached extension repository.
```

**Step 2: Commit documentation**

```bash
git add docs/pushover-version-management.md
git commit -m "docs: add Pushover version management documentation

- Document auto-download feature
- Explain VERSION file format
- Describe update flow
- List API methods
"
```

---

## Summary

This plan implements complete version management for cc-pushover-hook:

1. ✅ Dynamic version detection from git tags
2. ✅ VERSION file creation on installation
3. ✅ Auto-download extension on startup
4. ✅ Legacy installation support (no VERSION file)
5. ✅ Version comparison and update detection
6. ✅ UI components showing version status
7. ✅ One-click update functionality
8. ✅ Integration tests
9. ✅ Documentation

**Total commits**: ~12 focused commits following TDD principles
**Estimated time**: 2-3 hours for full implementation
