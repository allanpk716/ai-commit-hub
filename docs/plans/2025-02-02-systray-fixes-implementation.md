# 系统托盘问题修复实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 修复系统托盘的两个关键问题:图标不可见和菜单失效

**架构:** 通过添加 systray 健康检查实现自动重启,并使用 ICO 格式图标替换 PNG 格式

**技术栈:** Wails v2, github.com/getlantern/systray v1.2.2, atomic.Bool, ImageMagick/在线工具

---

## Task 1: 添加 systray 运行状态字段

**文件:**
- Modify: `app.go`

**Step 1: 导入 sync/atomic 包**

在 `app.go` 顶部的 import 区域添加:

```go
"sync/atomic"
```

**Step 2: 在 App struct 中添加新字段**

找到 `type App struct` 定义,在现有的 systray 相关字段之后添加:

```go
// 系统托盘相关字段
systrayReady    chan struct{}   // systray 就绪信号
systrayExit     *sync.Once      // 确保只退出一次
windowVisible   bool            // 窗口可见状态
windowMutex     sync.RWMutex    // 保护 windowVisible
systrayRunning atomic.Bool       // systray 运行状态 (新增)
```

**Step 3: 验证编译**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\.worktrees\fix\systray-issues
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功,无错误

**Step 4: 提交修改**

```bash
git add app.go
git commit -m "feat(tray): 添加 systrayRunning 状态追踪字段"
```

---

## Task 2: 修改 runSystray() 设置运行状态

**文件:**
- Modify: `app.go`

**Step 1: 修改 runSystray() 方法**

找到 `func (a *App) runSystray()` 方法,修改为:

```go
// runSystray 启动系统托盘 (在单独的 goroutine 中运行)
func (a *App) runSystray() {
	// 延迟初始化,避免与 Wails 启动冲突
	time.Sleep(500 * time.Millisecond)

	logger.Info("正在初始化系统托盘...")

	// 标记 systray 开始运行
	a.systrayRunning.Store(true)

	systray.Run(
		a.onSystrayReady,
		func() {
			// systray 退出时的清理
			a.systrayRunning.Store(false)
			a.onSystrayExit()
		},
	)
}
```

**关键变更:**
- 在调用 `systray.Run()` 之前设置 `a.systrayRunning.Store(true)`
- 在退出回调中设置 `a.systrayRunning.Store(false)`

**Step 2: 验证编译**

```bash
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功

**Step 3: 提交修改**

```bash
git add app.go
git commit -m "feat(tray): 在 runSystray 中添加运行状态管理"
```

---

## Task 3: 在 showWindow() 中添加健康检查和重启逻辑

**文件:**
- Modify: `app.go`

**Step 1: 修改 showWindow() 方法**

找到 `func (a *App) showWindow()` 方法,在现有逻辑之后添加健康检查:

```go
// showWindow 显示窗口
func (a *App) showWindow() {
	if a.ctx == nil {
		logger.Warn("showWindow: context 未初始化")
		return
	}

	a.windowMutex.Lock()
	defer a.windowMutex.Unlock()

	if a.windowVisible {
		logger.Debug("窗口已可见,跳过显示")
		return
	}

	logger.Info("显示窗口")
	runtime.WindowShow(a.ctx)
	a.windowVisible = true

	// 发送事件到前端
	runtime.EventsEmit(a.ctx, "window-shown", map[string]interface{}{
		"timestamp": time.Now(),
	})

	// === 健康检查和自动重启 ===
	// 检查 systray 是否还在运行
	if !a.systrayRunning.Load() {
		logger.Warn("检测到 systray 已停止,重新启动...")
		go a.runSystray()

		// 等待 systray 重新初始化完成
		time.Sleep(1 * time.Second)
		logger.Info("systray 重新启动完成")
	}
}
```

**关键变更:**
- 添加 systray 运行状态检查: `if !a.systrayRunning.Load()`
- 在 goroutine 中重启 systray: `go a.runSystray()`
- 等待 1 秒确保重新初始化完成

**Step 2: 验证编译**

```bash
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功

**Step 3: 提交修改**

```bash
git add app.go
git commit -m "feat(tray): 添加 systray 健康检查和自动重启机制"
```

---

## Task 4: 准备 ICO 图标文件

**外部操作 (不涉及代码):**

**步骤 1: 转换 PNG 为 ICO**

**方法 A: 使用在线工具 (推荐)**

1. 访问: https://convertio.co/png-ico
2. 上传文件: `frontend/src/assets/app-icon.png`
3. 配置输出:
   - 格式: ICO
   - 尺寸: 16x16, 32x32, 48x48, 256x256
   - 颜色深度: 32-bit (True Color + Alpha channel)
4. 点击"转换"并下载

**方法 B: 使用 ImageMagick (本地)**

```bash
# 如果安装了 ImageMagick
magick convert frontend/src/assets/app-icon.png \
  -define icon:auto-resize=256,128,64,48,32,16 \
  frontend/src/assets/app-icon.ico
```

**步骤 2: 验证 ICO 文件**

```bash
ls -lh frontend/src/assets/app-icon.ico
```

**预期输出:** 文件大小约 10-50KB (包含多个尺寸)

**步骤 3: 提交 ICO 文件**

```bash
git add frontend/src/assets/app-icon.ico
git commit -m "feat(tray): 添加 Windows 托盘图标 (ICO 格式)"
```

---

## Task 5: 修改 main.go 嵌入 ICO 图标

**文件:**
- Modify: `main.go`

**Step 1: 添加 ICO embed 指令**

在 `main.go` 中,找到现有的 PNG embed 指令,在其后添加:

```go
//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/src/assets/app-icon.png
var appIconPNG []byte

//go:embed frontend/src/assets/app-icon.ico
var appIconICO []byte
```

**Step 2: 验证编译**

```bash
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功,警告信息可以忽略

**Step 3: 提交修改**

```bash
git add main.go
git commit -m "feat(tray): 嵌入 ICO 格式托盘图标"
```

---

## Task 6: 在 app.go 中添加平台特定图标选择

**文件:**
- Modify: `app.go`

**Step 1: 添加 getTrayIcon() 方法**

在 `app.go` 中,在 shutdown() 方法之后添加:

```go
// getTrayIcon 根据平台返回合适的图标
func (a *App) getTrayIcon() []byte {
	if stdruntime.GOOS == "windows" {
		return appIconICO
	}
	// macOS 和 Linux 可以使用 PNG
	return appIconPNG
}
```

**Step 2: 验证编译**

```bash
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功

**Step 3: 提交修改**

```bash
git add app.go
git commit -m "feat(tray): 添加平台特定图标选择逻辑"
```

---

## Task 7: 修改 onSystrayReady() 使用平台特定图标

**文件:**
- Modify: `app.go`

**Step 1: 修改 onSystrayReady() 方法**

找到 `func (a *App) onSystrayReady()` 方法,修改 `systray.SetIcon()` 调用:

**原始代码:**
```go
systray.SetIcon(appIcon)
```

**修改为:**
```go
systray.SetIcon(a.getTrayIcon())
```

**Step 2: 验证编译**

```bash
go build -o build/bin/ai-commit-hub.exe .
```

**预期结果:** 编译成功

**Step 3: 提交修改**

```bash
git add app.go
git commit -m "feat(tray): 使用平台特定图标设置托盘"
```

---

## Task 8: 功能测试 - 托盘菜单持久性

**测试目标:** 验证修复后的托盘菜单在多次关闭/打开后仍然可用

**步骤 1: 构建并启动应用**

```bash
wails dev
```

**步骤 2: 执行 10 次关闭→打开循环**

对于 i = 1 to 10:
1. 点击窗口关闭按钮 (X)
2. 等待窗口隐藏
3. 右键点击托盘图标
4. **验证:** 菜单必须弹出
5. 点击"显示窗口"
6. **验证:** 窗口必须显示
7. **验证:** 日志应显示"显示窗口"

**成功标准:**
- ✅ 所有 10 次循环中,托盘菜单都能正常弹出
- ✅ "显示窗口"和"退出应用"菜单项始终可见
- ✅ 窗口可以正常恢复
- ✅ 如有 systray 重启,日志应记录"检测到 systray 已停止,重新启动..."

**步骤 3: 记录测试结果**

如果测试失败,记录:
- 失败发生在第几次循环
- 日志中的错误信息
- 托盘图标是否仍然可见

---

## Task 9: 功能测试 - 图标显示验证

**测试目标:** 验证托盘图标在 Windows 系统托盘中正确显示

**步骤 1: 启动应用**

```bash
wails build
# 或 wails dev
```

**步骤 2: 检查系统托盘**

1. 查看 Windows 任务栏右下角的系统托盘区域
2. **验证:** AI Commit Hub 图标必须可见
3. **验证:** 图标应该清晰,不模糊
4. **验证:** 图标应该有正确的颜色和形状

**成功标准:**
- ✅ 托盘图标在系统托盘区域清晰可见
- ✅ 图标与应用图标(app-icon.png)一致
- ✅ 图标在不同 DPI 设置下显示正常

**步骤 3: 测试不同 DPI 设置 (可选)**

1. 右键点击桌面 → 显示设置 → 显示
2. 修改缩放比例: 100%, 125%, 150%, 175%
3. 验证图标在每种比例下都清晰可见

---

## Task 10: 边界情况测试

**测试 A: 快速连续点击**

**步骤:**
1. 启动应用
2. 快速连续点击关闭按钮 5 次
3. **验证:** 只有最后一次关闭生效
4. **验证:** 窗口正确隐藏
5. 从托盘恢复窗口
6. **验证:** 窗口正常显示,托盘菜单可用

**成功标准:**
- ✅ 无错误或崩溃
- ✅ 托盘菜单始终可用
- ✅ 窗口状态正确

**测试 B: 窗口最小化后关闭**

**步骤:**
1. 启动应用
2. 点击窗口最小化按钮 (不是关闭)
3. 再点击关闭按钮 (X)
4. **验证:** 窗口正确隐藏到托盘
5. 从托盘恢复窗口
6. **验证:** 功能正常

**成功标准:**
- ✅ 窗口正确隐藏
- ✅ 托盘菜单可用
- ✅ 无应用崩溃

**测试 C: 长时间运行测试**

**步骤:**
1. 启动应用
2. 每 5 分钟执行一次关闭→打开操作
3. 持续 30 分钟
4. **验证:** 所有操作中托盘菜单始终可用

**成功标准:**
- ✅ 无性能退化
- ✅ 无内存泄漏
- ✅ 托盘功能始终稳定

---

## Task 11: 文档更新

**文件:**
- Modify: `docs/plans/2025-02-02-systray-fixes-design.md`

**步骤 1: 更新设计文档状态**

在文档顶部的状态信息中更新为:

```markdown
**状态:** 已实施
**实施日期:** 2025-02-02
**测试结果:** 通过
```

**步骤 2: 添加实施总结**

在文档末尾添加实施总结:

```markdown
## 实施总结

### 实际修改
- 添加 `systrayRunning atomic.Bool` 状态追踪
- 修改 `runSystray()` 添加运行状态管理
- 修改 `showWindow()` 添加健康检查和自动重启
- 转换图标为 ICO 格式
- 添加 `getTrayIcon()` 平台选择方法
- 修改 `onSystrayReady()` 使用平台特定图标

### 测试结果
- ✅ 托盘图标在 Windows 系统托盘中正确显示
- ✅ 托盘菜单在 10 次关闭/打开循环后始终可用
- ✅ 快速连续操作无问题
- ✅ 长时间运行稳定

### 已知问题
- 无
```

**步骤 3: 提交文档更新**

```bash
git add docs/plans/2025-02-02-systray-fixes-design.md
git commit -m "docs: 更新系统托盘修复文档 - 已实施"
```

---

## Task 12: 最终检查和合并

**步骤 1: 运行完整测试**

```bash
# 回到主分支目录
cd ../..

# 运行 Go 测试 (如果项目有测试)
go test ./...

# 构建应用
wails build
```

**步骤 2: 检查代码质量**

- 确认无编译错误
- 确认无运行时警告
- 确认日志输出清晰

**步骤 3: 切换到 main 分支**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
git checkout main
```

**步骤 4: 合并修复分支**

```bash
git merge fix/systray-issues --no-edit
```

**步骤 5: 推送到远程 (可选)**

```bash
git push origin main
```

**步骤 6: 清理 worktree**

```bash
git worktree remove .worktrees/fix/systray-issues
git branch -d fix/systray-issues
```

---

## 完成标准

✅ 所有 12 个任务完成
✅ 托盘图标在 Windows 系统托盘中可见
✅ 托盘菜单在多次关闭/打开后始终可用
✅ showWindow() 自动检测并重启失效的 systray
✅ 图标格式正确 (ICO for Windows, PNG for others)
✅ 所有测试场景通过
✅ 文档已更新
✅ 代码已合并到 main 分支

---

## 回滚方案

如果修复后出现新问题:

```bash
# 回滚到修复前的 commit
git log --oneline -3
git reset --hard fa5a2b7  # 修复设计文档的 commit

# 或者回滚特定修改
git revert HEAD~1  # 回滚最后一次提交
git revert HEAD~2  # 回滚最后两次提交
```

---

**实施计划创建时间:** 2025-02-02
**预计完成时间:** 65 分钟
**难度等级:** 中等
**依赖任务:** 无
