# 自动更新功能测试指南

## 🧪 测试模式说明

测试模式是为了解决"开发中无法测试更新功能"的问题而设计的。在测试模式下，应用会模拟检测到新版本，允许您反复测试下载、进度显示等功能。

## 🚀 快速开始

### 方法 1：使用测试脚本（推荐）

直接双击运行 `test-update.bat` 脚本：

```bash
./test-update.bat
```

脚本会自动：
- ✅ 设置测试模式环境变量
- ✅ 启动 `wails dev`
- ✅ 显示测试信息

### 方法 2：手动设置环境变量

在命令行中执行：

```bash
# Windows PowerShell
$env:AI_COMMIT_HUB_TEST_MODE="true"
wails dev

# Windows CMD
set AI_COMMIT_HUB_TEST_MODE=true
wails dev

# Linux/macOS
AI_COMMIT_HUB_TEST_MODE=true wails dev
```

## 📋 测试步骤

### 1. 启动应用

使用测试模式启动应用后，您会在日志中看到：

```
🧪 测试模式已启用
🧪 测试模式：返回测试更新信息
```

### 2. 触发检查更新

通过以下方式之一触发更新检查：
- 托盘菜单 → "检查更新"
- 设置页面 → "检查更新" 按钮
- 应用启动时会自动检查（后台）

### 3. 观察更新对话框

应该会显示：
```
发现新版本 v1.0.0-alpha.1

当前版本：v0.0.0-dev
最新版本：v1.0.0-alpha.1

[查看更新说明] [稍后] [立即更新]
```

### 4. 点击"立即更新"

开始下载后，会显示下载进度对话框：

```
正在下载更新
━━━━━━━━━━━━━━━━━━━ 45.2%

12.5 MB / 27.6 MB
下载速度：5.2 MB/s
预计剩余：00:03

[取消]
```

### 5. 测试各项功能

#### ✅ 进度显示
- 观察进度条是否平滑更新
- 百分比是否准确（每 100ms 更新）
- 已下载/总大小是否正确

#### ✅ 速度计算
- 下载速度是否显示合理（MB/s）
- 剩余时间是否准确

#### ✅ 取消功能
- 点击"取消"按钮
- 确认下载停止
- 临时文件被清理

#### ✅ 断点续传（可选）
- 下载过程中断开网络
- 恢复网络后观察是否继续下载
- 进度应该从断点继续

#### ✅ 重试机制（可选）
- 模拟网络错误
- 观察是否自动重试（最多 3 次）

## 🎯 测试数据说明

测试模式使用的更新信息：

| 字段 | 值 |
|------|-----|
| 版本号 | v1.0.0-alpha.1 |
| 下载 URL | GitHub Releases 真实链接 |
| 文件名 | ai-commit-hub-windows-amd64-v1.0.0-alpha.1.zip |
| 文件大小 | ~60MB（估算） |
| 是否预发布 | 是 (alpha) |

## 🛠️ 如何关闭测试模式

### 临时关闭
不设置环境变量即可：
```bash
wails dev  # 不设置 AI_COMMIT_HUB_TEST_MODE
```

### 永久关闭
删除或重命名测试脚本：
```bash
rm test-update.bat
```

## 📝 代码说明

测试模式在 `pkg/service/update_service.go` 中实现：

```go
// 检查环境变量
testMode := os.Getenv("AI_COMMIT_HUB_TEST_MODE") == "true"

// CheckForUpdates 方法中
if s.testMode {
    return s.getTestUpdateInfo()  // 返回测试数据
}
```

## 🐛 常见问题

### Q: 测试模式下会真的下载文件吗？
A: 是的，会从 GitHub Releases 下载真实的文件。这保证了下载功能完全正常工作。

### Q: 下载的文件会替换应用吗？
A: 不会。Wave 2 只实现了下载功能，安装替换功能在 Wave 3-4。

### Q: 可以自定义测试 URL 吗？
A: 目前测试模式使用固定的 v1.0.0-alpha.1 Release。如需自定义，可以修改 `getTestUpdateInfo()` 方法。

### Q: 如何测试不同的版本？
A: 修改 `getTestUpdateInfo()` 中的 `testVersion` 和 `testURL`。

## 📚 下一步

测试通过后，可以继续：
- Wave 3: 实现外部更新器程序
- Wave 4: 实现更新替换和自动重启

完整功能将在 Phase 4 完成后可用。

## 🎉 完成测试

测试通过后，在检查点回复 `approved`，我们将继续执行剩余的 Wave。
