# Pushover Hook 重装功能设计

**日期：** 2025-01-31
**状态：** 设计已确认

---

## 1. 功能概述

为每个项目添加"重装 Hook"功能，使用 `install.py --reinstall` 参数重新安装项目的 Pushover Hook，同时保留用户的通知配置（`.no-pushover` 和 `.no-windows` 文件）。

## 2. 需求背景

### 用户需求
- 手动触发单个项目重装：在项目列表中为每个项目单独提供重装按钮
- 覆盖安装时保留用户配置：重装时保留 `.no-pushover` 和 `.no-windows` 等用户配置

### 使用场景
- Hook 运行异常需要修复时
- 需要确保项目使用最新版本的 Hook 文件时
- 不想改变通知设置的重新安装

## 3. UI 设计

### 按钮位置
在 `PushoverStatusRow.vue` 中，与"更新 Hook"按钮并列显示。

### 显示逻辑
- 只在 Hook 已安装时显示
- 与"更新 Hook"按钮互斥显示：
  - **有更新可用**：显示"更新 Hook"按钮
  - **已是最新版本**：显示"重装 Hook"按钮
- 加载状态显示"重装中..."

### 交互流程
1. 用户点击"重装 Hook"按钮
2. 弹出确认对话框，说明：
   - 将使用最新版本的 Hook 文件覆盖当前安装
   - 保留通知配置（`.no-pushover`、`.no-windows`）
3. 用户确认后执行重装
4. 显示操作结果（成功/失败）

### 确认对话框内容
```
重装 Pushover Hook

这将重新安装 Pushover Hook 到该项目：
• 使用最新版本的 Hook 文件覆盖当前安装
• 保留您的通知配置（Pushover/Windows 通知设置）

确定要重装吗？
```

按钮：[取消] [确定重装]

## 4. 后端实现

### 4.1 新增类型定义（pkg/pushover/types.go）

```go
// NotificationConfig 通知配置
type NotificationConfig struct {
    NoPushoverFile bool
    NoWindowsFile  bool
}
```

### 4.2 Installer 层（pkg/pushover/installer.go）

```go
// Reinstall 重装 Hook（保留用户配置）
func (in *Installer) Reinstall(projectPath string) (*InstallResult, error) {
    // 1. 检查扩展目录是否存在
    if _, err := os.Stat(in.extensionPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
    }

    // 2. 检查 install.py 是否存在
    installScript := filepath.Join(in.extensionPath, "install.py")
    if _, err := os.Stat(installScript); os.IsNotExist(err) {
        return nil, fmt.Errorf("install.py 不存在，请确保 cc-pushover-hook 扩展完整")
    }

    // 3. 检查 Python 是否可用
    pythonCmd, err := in.findPython()
    if err != nil {
        return nil, err
    }

    // 4. 读取当前通知配置
    config := in.readNotificationConfig(projectPath)

    // 5. 调用 install.py --reinstall
    args := []string{
        installScript,
        "--target-dir", projectPath,
        "--non-interactive",
        "--reinstall",
    }

    cmd := exec.Command(pythonCmd, args...)
    cmd.Dir = in.extensionPath
    output, err := cmd.CombinedOutput()

    if err != nil {
        return &InstallResult{
            Success: false,
            Message: fmt.Sprintf("重装失败: %v\n输出: %s", err, string(output)),
        }, nil
    }

    // 6. 恢复通知配置
    if err := in.restoreNotificationConfig(projectPath, config); err != nil {
        // 配置恢复失败记录警告，但不影响重装结果
        fmt.Fprintf(os.Stderr, "[WARN] 恢复通知配置失败: %v\n", err)
    }

    // 7. 解析并返回结果
    return in.parseInstallResult(output)
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

// parseInstallResult 解析安装结果（提取自 Install 和 Update 方法）
func (in *Installer) parseInstallResult(output []byte) (*InstallResult, error) {
    outputStr := string(output)
    lines := strings.Split(strings.TrimSpace(outputStr), "\n")
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
            Success: true,
            Message: "操作成功",
            HookPath: "", // 需要在调用处指定
        }, nil
    }

    return &InstallResult{
        Success: false,
        Message: fmt.Sprintf("无法解析安装结果: %s", outputStr),
    }, nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

### 4.3 Service 层（pkg/pushover/service.go）

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

### 4.4 App 层（app.go）

```go
// ReinstallPushoverHook 重装项目的 Pushover Hook
func (a *App) ReinstallPushoverHook(projectPath string) (*InstallResult, error) {
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

## 5. 前端实现

### 5.1 PushoverStatusRow.vue 修改

```vue
<template>
  <div class="pushover-status-row">
    <!-- 现有的状态显示部分保持不变 -->
    <!-- ... -->

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

      <!-- 修改：有更新时显示更新按钮 -->
      <button
        v-else-if="needsUpdate"
        class="action-btn btn-update"
        :disabled="loading"
        @click="handleUpdate"
      >
        {{ loading ? '更新中...' : '更新 Hook' }}
      </button>

      <!-- 新增：已是最新版本时显示重装按钮 -->
      <button
        v-else
        class="action-btn btn-reinstall"
        :disabled="loading"
        @click="handleReinstall"
      >
        {{ loading ? '重装中...' : '重装 Hook' }}
      </button>
    </div>

    <!-- 新增：重装确认对话框 -->
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
  </div>
</template>

<script setup lang="ts">
// 现有的导入保持不变
// ...

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
</script>

<style scoped>
/* 新增对话框样式 */
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

/* 重装按钮样式 */
.btn-reinstall {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.btn-reinstall:hover:not(:disabled) {
  background: rgba(6, 182, 212, 0.25);
}
</style>
```

### 5.2 pushoverStore.ts 修改

```typescript
// 在 pushoverStore 中添加新方法

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

### 5.3 类型定义（types/pushover.ts 或 frontend/src/types/index.ts）

确保 `InstallResult` 类型已定义：

```typescript
export interface InstallResult {
  success: boolean
  message: string
  hookPath?: string
}
```

### 5.4 Wails 绑定生成

修改后运行以下命令重新生成绑定：

```bash
wails dev
```

## 6. 数据流

```
用户点击"重装 Hook"按钮
    ↓
弹出确认对话框
    ↓
用户确认
    ↓
pushoverStore.reinstallHook(projectPath)
    ↓
ReinstallPushoverHook(projectPath) // Wails 绑定
    ↓
App.ReinstallPushoverHook()
    ↓
Service.ReinstallHook()
    ↓
Installer.Reinstall()
    ↓
├─ 读取通知配置 (.no-pushover, .no-windows)
├─ 调用 install.py --reinstall
├─ 恢复通知配置
└─ 解析并返回安装结果
    ↓
返回安装结果到前端
    ↓
刷新项目状态缓存
    ↓
更新 UI
```

## 7. 错误处理

| 错误场景 | 错误信息 | 处理方式 |
|---------|---------|---------|
| Python 未安装 | "未找到 Python，请安装 Python 3.6+" | 显示错误 Toast，禁用按钮 |
| 扩展未下载 | "cc-pushover-hook 扩展未下载" | 显示错误 Toast，禁用按钮 |
| 项目路径不存在 | "项目路径不存在: {path}" | 显示错误 Toast |
| 项目未安装 Hook | "项目未安装 Pushover Hook，请先安装" | 显示错误 Toast |
| 安装脚本执行失败 | "重装失败: {error}" | 显示错误 Toast，包含脚本输出 |
| 配置恢复失败 | 记录警告日志 | 不影响重装结果，后台记录警告 |

## 8. 测试计划

### 8.1 单元测试

**pkg/pushover/installer_test.go**

```go
func TestReinstall(t *testing.T) {
    // 测试用例：
    // 1. 正常重装（保留配置）
    // 2. 配置恢复失败时的处理
    // 3. 扩展未下载
    // 4. Python 未安装
}
```

### 8.2 集成测试

**tests/integration/pushover_reinstall_test.go**

```go
func TestReinstallIntegration(t *testing.T) {
    // 1. 创建测试项目
    // 2. 安装 Hook
    // 3. 修改通知配置
    // 4. 重装 Hook
    // 5. 验证配置已保留
    // 6. 验证 Hook 文件已更新
}
```

### 8.3 手动测试

1. **正常重装流程**
   - 安装 Hook 到测试项目
   - 修改通知模式（如禁用 Pushover）
   - 点击"重装 Hook"
   - 确认配置已保留

2. **异常场景测试**
   - 扩展未下载时点击重装
   - Python 未安装时点击重装
   - Hook 未安装时点击重装

3. **UI 交互测试**
   - 确认对话框的显示和关闭
   - 加载状态的显示
   - 错误提示的显示

## 9. 后续优化

- [ ] 批量重装：在管理面板中添加"全部重装"功能
- [ ] 重装历史：记录重装操作历史，支持回滚
- [ ] 配置备份：重装前自动备份配置到 `.claude/hooks/pushover-hook/config.backup`
- [ ] 进度提示：对于大型项目，显示重装进度条

## 10. 实现清单

### 后端
- [ ] `pkg/pushover/installer.go`: 添加 `Reinstall()` 方法
- [ ] `pkg/pushover/installer.go`: 添加 `readNotificationConfig()` 方法
- [ ] `pkg/pushover/installer.go`: 添加 `restoreNotificationConfig()` 方法
- [ ] `pkg/pushover/installer.go`: 添加 `parseInstallResult()` 方法（重构现有代码）
- [ ] `pkg/pushover/service.go`: 添加 `ReinstallHook()` 方法
- [ ] `app.go`: 添加 `ReinstallPushoverHook()` 方法
- [ ] `pkg/pushover/installer_test.go`: 添加单元测试
- [ ] `tests/integration/pushover_reinstall_test.go`: 添加集成测试

### 前端
- [ ] `frontend/src/components/PushoverStatusRow.vue`: 修改模板和脚本
- [ ] `frontend/src/stores/pushoverStore.ts`: 添加 `reinstallHook()` 方法
- [ ] `frontend/src/types/pushover.ts`: 确保 `InstallResult` 类型定义
- [ ] 运行 `wails dev` 重新生成绑定

---

**文档版本：** 1.0
**最后更新：** 2025-01-31
