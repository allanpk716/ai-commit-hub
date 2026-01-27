# cc-pushover-hook 版本检查修复设计

## 问题

`GetLatestVersion()` 使用 `git describe --tags --abbrev=0 origin/main` 获取远程版本，但只读取本地缓存的 `origin/main` 引用。远程发布新 tag 后（如 v1.4.1），本地不知道，导致显示"已是最新"。

## 根本原因

`origin/main` 引用仅在 `git fetch` 或 `git pull` 时更新。当前实现没有先 fetch 远程引用。

## 解决方案

### 后端修改

**文件：** `pkg/pushover/repository.go`

**新增方法：** `fetchRemoteTags()`

```go
// fetchRemoteTags 获取远程 tags 和分支更新
func (rm *RepositoryManager) fetchRemoteTags() error {
    if !rm.IsCloned() {
        return fmt.Errorf("扩展不存在")
    }

    extensionPath := rm.GetExtensionPath()

    cmd := Command("git", "fetch", "origin", "main", "--tags")
    cmd.Dir = extensionPath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("fetch 远程更新失败: %v\n输出: %s", err, string(output))
    }

    return nil
}
```

**修改：** `GetLatestVersion()`

在获取版本前调用 `fetchRemoteTags()`：

```go
func (rm *RepositoryManager) GetLatestVersion() (string, error) {
    if !rm.IsCloned() {
        return "", fmt.Errorf("扩展不存在")
    }

    // 先获取远程更新
    if err := rm.fetchRemoteTags(); err != nil {
        return "", err
    }

    // 原有获取版本逻辑...
}
```

### 前端修改

**文件：** `frontend/src/stores/pushoverStore.ts`

添加错误处理：

```typescript
updateCheckError: string | null = null

async checkExtensionUpdates() {
  this.updateCheckError = null
  this.isCheckingUpdate = true

  try {
    const result = await CheckForUpdates()
    this.updateAvailable = result.needsUpdate
    this.currentVersion = result.currentVersion
    this.latestVersion = result.latestVersion
  } catch (error) {
    this.updateCheckError = error.message || '检查更新失败'
    console.error('检查扩展更新失败:', error)
  } finally {
    this.isCheckingUpdate = false
  }
}
```

**文件：** `frontend/src/components/PushoverStatusRow.vue`

显示错误状态和重试按钮。

## 错误处理策略

采用严格模式：fetch 失败时返回错误，UI 显示明确的失败提示。

**边界情况：**
- 首次安装后无网络：显示检查失败
- 远程仓库访问受限：返回明确错误
- fetch 超时：返回超时错误
- git 命令不存在：返回友好提示

## 调用链

```
UI 启动/用户点击
  → pushoverStore.checkExtensionUpdates()
  → CheckForUpdates() [Wails binding]
  → Service.CheckForUpdates()
  → RepositoryManager.GetLatestVersion()
  → fetchRemoteTags() [新增]
```

## 测试计划

### 单元测试

1. `fetchRemoteTags()` 成功场景
2. `fetchRemoteTags()` 失败场景（无网络、超时）
3. `GetLatestVersion()` fetch 后获取新版本
4. `GetLatestVersion()` fetch 失败时错误传播

### 集成测试

1. 本地 v1.4.0，远程 v1.4.1 → 检测到更新
2. 已是最新版本 → needsUpdate=false
3. 网络断开 → 返回错误

## 验证步骤

1. 修改后端代码
2. 修改前端 Store 和组件
3. 启动应用，应检测到 v1.4.1
4. 断网重试，应显示错误
5. 验证手动"更新扩展"功能正常
