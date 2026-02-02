# 系统托盘功能实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 为 AI Commit Hub 添加系统托盘功能,允许用户关闭窗口后应用继续在后台运行,并通过系统托盘图标重新打开或完全退出应用。

**架构:** 使用 Wails v2 的生命周期钩子拦截窗口关闭事件,结合 github.com/getlantern/systray 库管理系统托盘图标和菜单。通过 sync.Once 和 sync.WaitGroup 确保安全退出,防止竞态条件。

**技术栈:**
- Wails v2 (当前框架)
- github.com/getlantern/systray (跨平台系统托盘库)
- Vue 3 (前端事件监听,可选)

---

## Task 1: 安装依赖并准备图标资源

**文件:**
- Modify: `go.mod`
- Modify: `go.sum`
- Verify: `frontend/src/assets/app-icon.png`

**Step 1: 安装 systray 依赖**

```bash
go get github.com/getlantern/systray
go mod tidy
```

**Step 2: 验证应用图标存在**

检查: `frontend/src/assets/app-icon.png`
- 如果不存在,需要准备一个 256x256 PNG 图标
- 图标将用于托盘显示

**Step 3: 提交依赖更新**

```bash
git add go.mod go.sum
git commit -m "feat(tray): 添加 systray 依赖"
```

---

## Task 2: 修改 main.go 启动 systray

**文件:**
- Modify: `main.go`

**Step 1: 添加 systray 图标嵌入**

在 `main.go` 顶部,现有 embed 指令之后添加:

```go
//go:embed frontend/src/assets/app-icon.png
var appIcon []byte
```

**Step 2: 修改 main() 函数启动 systray**

找到 `main()` 函数,在 `wails.Run()` 之前添加 systray 启动:

```go
func main() {
	// 初始化 logger
	initLogger()

	// Create an instance of the app structure
	app := NewApp()

	// 启动系统托盘 (在 Wails 启动前)
	go app.runSystray()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "AI Commit Hub",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.onBeforeClose,  // 新增: 拦截关闭
		OnShutdown:       app.shutdown,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		logger.Errorf("Error: %v", err)
	}
}
```

**关键变更:**
- 添加 `go app.runSystray()` 启动托盘 goroutine
- 添加 `OnBeforeClose: app.onBeforeClose` 拦截窗口关闭

**Step 3: 提交 main.go 修改**

```bash
git add main.go
git commit -m "feat(tray): 在 main.go 中启动 systray 并添加关闭拦截"
```

---

## Task 3: 在 App 结构中添加托盘相关字段

**文件:**
- Modify: `app.go`

**Step 1: 导入 systray 包**

在 `app.go` 顶部的 import 区域添加:

```go
import (
	"context"
	// ... 现有 imports
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)
```

**Step 2: 扩展 App 结构体**

在 `app.go` 中找到 `type App struct` 定义,添加托盘相关字段:

```go
// App struct
type App struct {
	ctx                  context.Context
	dbPath               string
	gitProjectRepo       *repository.GitProjectRepository
	commitHistoryRepo    *repository.CommitHistoryRepository
	configService        *service.ConfigService
	projectConfigService *service.ProjectConfigService
	pushoverService      *pushover.Service
	errorService         *service.ErrorService
	initError            error

	// 系统托盘相关字段
	systrayReady    chan struct{}   // systray 就绪信号
	systrayExit     *sync.Once      // 确保只退出一次
	windowVisible   bool            // 窗口可见状态
	windowMutex     sync.RWMutex    // 保护 windowVisible
}
```

**Step 3: 初始化新字段**

在 `NewApp()` 函数中初始化新字段:

```go
// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		systrayReady:  make(chan struct{}),
		systrayExit:   &sync.Once{},
		windowVisible: true, // 启动时窗口可见
	}
}
```

**Step 4: 提交结构修改**

```bash
git add app.go
git commit -m "feat(tray): 添加托盘相关字段到 App 结构"
```

---

## Task 4: 实现 runSystray 方法

**文件:**
- Modify: `app.go`

**Step 1: 实现 runSystray() 方法**

在 `app.go` 中添加以下方法(在 `shutdown()` 方法之后):

```go
// runSystray 启动系统托盘 (在单独的 goroutine 中运行)
func (a *App) runSystray() {
	// 延迟初始化,避免与 Wails 启动冲突
	time.Sleep(500 * time.Millisecond)

	logger.Info("正在初始化系统托盘...")

	systray.Run(
		a.onSystrayReady,
		a.onSystrayExit,
	)
}

// onSystrayReady 在 systray 就绪时调用
func (a *App) onSystrayReady() {
	logger.Info("系统托盘初始化成功")

	// 设置托盘图标
	systray.SetIcon(appIcon)
	systray.SetTitle("AI Commit Hub")
	systray.SetTooltip("AI Commit Hub - 点击显示窗口")

	// 创建托盘菜单
	// 菜单项将在下一个 task 中实现

	// 通知 systray 已就绪
	close(a.systrayReady)
}

// onSystrayExit 在 systray 退出时调用
func (a *App) onSystrayExit() {
	logger.Info("系统托盘已退出")
}
```

**Step 2: 提交 runSystray 实现**

```bash
git add app.go
git commit -m "feat(tray): 实现 runSystray 方法"
```

---

## Task 5: 实现托盘菜单项

**文件:**
- Modify: `app.go`

**Step 1: 在 onSystrayReady 中创建菜单**

修改 `onSystrayReady()` 方法,添加菜单项:

```go
// onSystrayReady 在 systray 就绪时调用
func (a *App) onSystrayReady() {
	logger.Info("系统托盘初始化成功")

	// 设置托盘图标
	systray.SetIcon(appIcon)
	systray.SetTitle("AI Commit Hub")
	systray.SetTooltip("AI Commit Hub - 点击显示窗口")

	// 创建菜单
	menu := systray.AddMenuItem("显示窗口", "显示主窗口")
	go func() {
		for range menu.ClickedCh {
			a.showWindow()
		}
	}()

	// 添加分隔线
	systray.AddSeparator()

	// 退出菜单项
	quitMenu := systray.AddMenuItem("退出应用", "完全退出应用")
	go func() {
		for range quitMenu.ClickedCh {
			a.quitApplication()
		}
	}()

	// 通知 systray 已就绪
	close(a.systrayReady)
}
```

**Step 2: 提交菜单实现**

```bash
git add app.go
git commit -m "feat(tray): 实现托盘菜单项 (显示窗口、退出应用)"
```

---

## Task 6: 实现窗口控制方法

**文件:**
- Modify: `app.go`

**Step 1: 实现 showWindow() 方法**

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
}
```

**Step 2: 实现 hideWindow() 方法**

```go
// hideWindow 隐藏窗口
func (a *App) hideWindow() {
	if a.ctx == nil {
		logger.Warn("hideWindow: context 未初始化")
		return
	}

	a.windowMutex.Lock()
	defer a.windowMutex.Unlock()

	if !a.windowVisible {
		logger.Debug("窗口已隐藏,跳过隐藏")
		return
	}

	logger.Info("隐藏窗口到托盘")
	runtime.WindowHide(a.ctx)
	a.windowVisible = false

	// 发送事件到前端
	runtime.EventsEmit(a.ctx, "window-hidden", map[string]interface{}{
		"timestamp": time.Now(),
	})
}
```

**Step 3: 实现 quitApplication() 方法**

```go
// quitApplication 完全退出应用
func (a *App) quitApplication() {
	// 使用 sync.Once 确保只执行一次
	a.systrayExit.Do(func() {
		logger.Info("应用正在退出...")

		if a.ctx != nil {
			runtime.Quit(a.ctx)
		} else {
			// 如果 context 未初始化,强制退出
			logger.Warn("context 未初始化,使用 os.Exit")
			os.Exit(0)
		}
	})
}
```

**Step 4: 提交窗口控制方法**

```bash
git add app.go
git commit -m "feat(tray): 实现窗口控制方法 (show/hide/quit)"
```

---

## Task 7: 实现 onBeforeClose 拦截窗口关闭

**文件:**
- Modify: `app.go`

**Step 1: 实现 onBeforeClose() 方法**

```go
// onBeforeClose 拦截窗口关闭事件,隐藏到托盘而非退出
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	logger.Info("窗口关闭事件被触发,将隐藏到托盘")

	// 隐藏窗口而非退出
	a.hideWindow()

	// 返回 true 阻止窗口关闭
	return true
}
```

**Step 2: 更新 shutdown() 方法确保 systray 正确退出**

找到现有的 `shutdown()` 方法,修改为:

```go
// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	logger.Info("AI Commit Hub shutting down...")

	// 退出系统托盘
	systray.Quit()
}
```

**Step 3: 提交关闭拦截实现**

```bash
git add app.go
git commit -m "feat(tray): 实现 onBeforeClose 拦截窗口关闭"
```

---

## Task 8: 添加前端事件监听 (可选)

**文件:**
- Modify: `frontend/src/App.vue`

**Step 1: 在 App.vue 中添加窗口状态监听**

在 `<script setup lang="ts">` 的 `onMounted` 中添加事件监听:

```typescript
onMounted(async () => {
  console.log('[App] onMounted 开始')

  // 现有的初始化代码...

  // 2. 监听窗口可见性事件 (系统托盘相关)
  EventsOn('window-shown', (data: { timestamp: string }) => {
    console.log('[App] 窗口已从托盘恢复', data.timestamp)
  })

  EventsOn('window-hidden', (data: { timestamp: string }) => {
    console.log('[App] 窗口已隐藏到托盘', data.timestamp)
  })

  // 现有的事件监听...

  console.log('[App] onMounted 完成')
})
```

**Step 2: (可选) 添加首次使用提示**

在 `App.vue` 的 `<script setup>` 中添加提示逻辑:

```typescript
// 检查是否首次使用托盘功能
const showTrayTip = ref(false)

onMounted(async () => {
  // ... 现有代码

  EventsOn('window-hidden', () => {
    console.log('[App] 窗口已隐藏到托盘')

    // 首次隐藏时显示提示
    if (!localStorage.getItem('tray-tip-shown')) {
      showTrayTip.value = true
      localStorage.setItem('tray-tip-shown', 'true')

      // 3秒后自动关闭提示
      setTimeout(() => {
        showTrayTip.value = false
      }, 3000)
    }
  })
})
```

在 `<template>` 中添加提示组件:

```vue
<template>
  <!-- 现有内容 -->

  <!-- 托盘提示 (可选) -->
  <transition name="fade">
    <div v-if="showTrayTip" class="tray-tip">
      <span class="icon">💡</span>
      <span>应用已最小化到系统托盘,可以通过托盘图标重新打开</span>
    </div>
  </transition>
</template>

<style scoped>
.tray-tip {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: var(--accent-primary);
  color: white;
  padding: var(--space-md) var(--space-lg);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  z-index: var(--z-modal);
  animation: slide-up 0.3s ease-out;
}

.tray-tip .icon {
  font-size: 18px;
}

@keyframes slide-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
```

**Step 3: 提交前端集成**

```bash
git add frontend/src/App.vue
git commit -m "feat(tray): 添加窗口状态事件监听和用户提示"
```

---

## Task 9: 功能测试

**测试环境:**
- Windows 10/11 (主要目标)

**Step 1: 构建并运行应用**

```bash
wails dev
```

**Step 2: 测试基本功能**

按顺序测试以下场景:

1. **启动测试**
   - ✅ 应用启动,主窗口正常显示
   - ✅ 系统托盘图标出现
   - ✅ 托盘图标正确显示

2. **关闭到托盘测试**
   - ✅ 点击窗口关闭按钮 (X)
   - ✅ 窗口隐藏,托盘图标保留
   - ✅ 应用继续运行 (未退出)

3. **从托盘恢复测试**
   - ✅ 右键点击托盘图标
   - ✅ 菜单显示"显示窗口"和"退出应用"
   - ✅ 点击"显示窗口"
   - ✅ 窗口重新显示

4. **退出应用测试**
   - ✅ 右键点击托盘图标
   - ✅ 点击"退出应用"
   - ✅ 应用完全退出
   - ✅ 托盘图标消失

5. **重复操作测试**
   - ✅ 多次快速点击关闭按钮
   - ✅ 窗口只隐藏一次,无错误
   - ✅ 多次快速点击托盘"退出"
   - ✅ 应用只退出一次,无竞态

**Step 3: 检查日志输出**

查看日志确认:
- `[INFO] 系统托盘初始化成功`
- `[INFO] 显示窗口` / `[INFO] 隐藏窗口到托盘`
- `[INFO] 应用正在退出...`
- `[INFO] 系统托盘已退出`

**Step 4: 性能检查**

- 观察内存占用是否正常 (< 10MB 增加)
- 窗口显示/隐藏响应是否流畅 (< 200ms)

---

## Task 10: 文档更新

**文件:**
- Modify: `CLAUDE.md`

**Step 1: 在 CLAUDE.md 中添加系统托盘说明**

在 `## 功能特性` 章节添加新章节:

```markdown
### 系统托盘功能

**功能说明:**
- 支持将应用最小化到系统托盘,后台运行
- 关闭窗口时应用不退出,继续驻留在托盘
- 通过托盘菜单可以恢复窗口或完全退出应用

**使用方式:**
1. **隐藏到托盘**: 点击窗口关闭按钮 (X)
2. **恢复窗口**: 右键点击托盘图标 → "显示窗口"
3. **退出应用**: 右键点击托盘图标 → "退出应用"

**注意事项:**
- 首次使用时会在关闭窗口后显示提示信息
- 应用启动时默认显示主窗口
- 托盘图标使用应用图标 (app-icon.png)

**技术实现:**
- 使用 `github.com/getlantern/systray` 库
- Wails `OnBeforeClose` 钩子拦截窗口关闭
- `sync.Once` 确保安全退出
```

**Step 2: 提交文档更新**

```bash
git add CLAUDE.md
git commit -m "docs: 添加系统托盘功能说明"
```

---

## Task 11: 最终检查和清理

**Step 1: 运行完整测试套件**

```bash
# Go 后端测试
go test ./... -v

# 前端测试 (如果有)
cd frontend && npm run test:run
```

**Step 2: 检查代码质量**

- 确认没有编译错误
- 确认没有运行时警告
- 确认日志输出清晰

**Step 3: 清理临时文件**

```bash
# 清理 build 产物 (如果有)
wails build -clean
```

**Step 4: 创建最终合并请求**

```bash
# 切换回 main 分支
cd ../..
git checkout main

# 合并 feature 分支
git merge feature/system-tray

# 推送到远程 (如果需要)
git push origin main
```

**Step 5: 创建 Git Tag (可选)**

```bash
git tag -a v1.1.0 -m "添加系统托盘功能"
git push origin v1.1.0
```

---

## 完成标准

✅ 所有 11 个任务完成
✅ 所有测试场景通过
✅ 代码已提交到 feature/system-tray 分支
✅ CLAUDE.md 已更新
✅ 无编译错误和运行时警告
✅ 内存占用增加 < 10MB
✅ 窗口响应 < 200ms

---

## 回滚方案

如果遇到问题需要回滚:

```bash
# 删除 worktree
git worktree remove .worktrees/feature/system-tray

# 删除 feature 分支
git branch -D feature/system-tray

# 恢复到之前的版本
git reset --hard <commit-before-changes>
```

---

**实施计划创建时间:** 2025-02-02
**预计完成时间:** 2 小时
**难度等级:** 中等
**依赖任务:** 无
