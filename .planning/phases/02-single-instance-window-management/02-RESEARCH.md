# Phase 02: Single Instance & Window Management - Research

**Researched:** 2026-02-06
**Domain:** Wails v2 Desktop Application (Go + Vue3)
**Confidence:** HIGH

## Summary

本阶段研究了如何在 Wails v2 应用中实现单实例锁定机制和窗口状态持久化。通过查阅 Wails 官方文档、Context7 代码示例和行业最佳实践,确定了标准的实现方案。

**关键发现:**
1. Wails v2 内置的 `SingleInstanceLock` 机制是实现单实例锁定的标准方式,无需手动实现互斥量或文件锁
2. 使用 `OnSecondInstanceLaunch` 回调处理第二个实例启动,通过 `runtime.WindowShow` 和 `runtime.WindowUnminimise` 激活现有窗口
3. 窗口状态持久化需要在 `OnBeforeClose` 生命周期钩子中保存窗口位置、大小和最大化状态到 SQLite 数据库
4. 在 `OnStartup` 中读取保存的状态并应用,但需验证数据有效性(如位置是否在屏幕范围内)

**主要建议:**
- **单实例锁定**: 使用 Wails 内置的 `SingleInstanceLock` 选项,配置 UUID 作为唯一标识符
- **窗口激活**: 在 `OnSecondInstanceLaunch` 回调中调用 `runtime.WindowShow` + `runtime.WindowUnminimise`
- **状态持久化**: 在 `OnBeforeClose` 中保存窗口状态,在 `OnStartup` 中恢复状态
- **数据验证**: 恢复窗口状态前验证位置是否在屏幕范围内,避免窗口"丢失"
- **数据库设计**: 创建专门的 `window_state` 表存储窗口配置(key-value 结构或专用字段)

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Wails v2 SingleInstanceLock | v2.10+ | 单实例锁定机制 | Wails 官方内置功能,跨平台支持(Windows/macOS/Linux),无需第三方依赖 |
| Wails v2 Runtime APIs | v2.10+ | 窗口管理(WindowShow, WindowUnminimise, WindowSetSize, WindowSetPosition) | 官方运行时 API,提供完整的窗口控制能力 |
| GORM | Latest (项目已使用) | 窗口状态持久化到 SQLite | 项目已集成 GORM + SQLite,无需额外依赖 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/uuid | Latest | 生成 SingleInstanceLock 的 UniqueId | 确保应用 ID 全局唯一,避免冲突 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Wails SingleInstanceLock | 手动实现互斥量(Mutex)/文件锁 | 自定义实现复杂且容易出错,需要处理多平台差异(Windows: CreateMutex, macOS: BSD mutex, Linux: flock) |
| SQLite + GORM | JSON 文件/YAML 配置文件 | 文件格式需要手动处理序列化和并发写入,GORM 提供事务支持和类型安全 |
| runtime APIs | WebView2 平台特定 API | 平台特定 API 降低代码可移植性,runtime APIs 提供跨平台抽象 |

**Installation:**

无需安装额外依赖。项目已集成:
- Wails v2 (包含 SingleInstanceLock 和 runtime APIs)
- GORM + SQLite (用于窗口状态持久化)

如需生成 UUID:
```bash
go get github.com/google/uuid
```

## Architecture Patterns

### Recommended Project Structure

```
pkg/
├── models/
│   └── window_state.go          # WindowState 模型定义
├── repository/
│   ├── window_state_repository.go  # 窗口状态数据访问层
│   └── migration.go               # 数据库迁移(创建 window_state 表)
└── service/
    └── window_service.go         # 窗口管理业务逻辑

app.go                            # 修改: 添加 SingleInstanceLock 配置和窗口状态管理
main.go                           # 修改: 配置 SingleInstanceLock 选项
```

### Pattern 1: Single Instance Lock Implementation

**What:** 使用 Wails 内置的 `SingleInstanceLock` 防止多实例运行,并在检测到第二个实例时激活现有窗口。

**When to use:** 所有需要单实例的桌面应用,特别是需要防止用户意外启动多个实例的场景。

**Example:**

```go
// Source: https://wails.io/docs/v2.10/guides/single-instance-lock

package main

import (
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

// startup 在应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// onSecondInstanceLaunch 在第二个实例尝试启动时调用
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	// 记录日志(仅失败时,成功时不记录)
	runtime.LogInfo(a.ctx, "检测到第二个实例启动")
	runtime.LogInfof(a.ctx, "工作目录: %s", secondInstanceData.WorkingDirectory)
	runtime.LogInfof(a.ctx, "参数: %v", secondInstanceData.Args)

	// 激活现有窗口到前台
	// 1. 如果窗口最小化,先恢复
	runtime.WindowUnminimise(a.ctx)
	// 2. 显示窗口
	runtime.WindowShow(a.ctx)
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "AI Commit Hub",
		Width:  1280,
		Height: 800,
		OnStartup: app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			// 使用 UUID 确保唯一性
			UniqueId: "e3984e08-28dc-4e3d-b70a-45e961589cdc", // 建议使用 UUID 生成器
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

**Key Points:**
- `UniqueId` 必须全局唯一,建议使用 UUID 生成器(如 `github.com/google/uuid`)
- `OnSecondInstanceLaunch` 回调接收 `SecondInstanceData`,包含第二个实例的参数和工作目录
- 必须在回调中手动调用 `runtime.WindowShow` 和 `runtime.WindowUnminimise`,Wails 不会自动激活窗口
- **静默激活**: 成功时不显示任何通知或对话框,符合用户体验预期

### Pattern 2: Window State Persistence

**What:** 在窗口关闭时保存窗口状态(位置、大小、最大化状态),下次启动时恢复。

**When to use:** 需要记住用户窗口偏好的桌面应用。

**Example:**

```go
// Source: Wails Runtime API 文档 + GORM 最佳实践

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// WindowState 窗口状态模型
type WindowState struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null"` // 配置键(如 "window.main")
	X     int    `gorm:"not null"`             // 窗口 X 坐标
	Y     int    `gorm:"not null"`             // 窗口 Y 坐标
	Width int    `gorm:"not null"`             // 窗口宽度
	Height int   `gorm:"not null"`             // 窗口高度
	Maximized bool   `gorm:"default:false"`    // 是否最大化
	MonitorID string `gorm:"size:50"`           // 显示器编号(可选,用于多显示器)
}

type App struct {
	ctx              context.Context
	windowStateRepo *WindowStateRepository
}

// startup 在应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 尝试恢复窗口状态
	if err := a.restoreWindowState(); err != nil {
		// 记录错误但继续启动,使用默认窗口位置
		runtime.LogWarning(a.ctx, "窗口状态恢复失败,使用默认位置: "+err.Error())
	}
}

// restoreWindowState 恢复窗口状态
func (a *App) restoreWindowState() error {
	state, err := a.windowStateRepo.GetByKey("window.main")
	if err != nil {
		return err // 无记录或数据库错误
	}

	// 验证窗口位置是否在屏幕范围内
	if !a.isPositionValid(state.X, state.Y, state.Width, state.Height) {
		runtime.LogWarning(a.ctx, "保存的窗口位置无效,使用默认位置")
		return nil
	}

	// 在窗口显示前设置位置和大小
	runtime.WindowSetPosition(a.ctx, state.X, state.Y)
	runtime.WindowSetSize(a.ctx, state.Width, state.Height)

	// 如果之前是最大化状态,在窗口显示后最大化
	if state.Maximized {
		runtime.WindowMaximise(a.ctx)
	}

	runtime.LogInfo(a.ctx, "窗口状态已恢复")
	return nil
}

// isPositionValid 验证窗口位置是否在屏幕范围内
func (a *App) isPositionValid(x, y, width, height int) bool {
	// 简单验证: 窗口至少部分在屏幕内
	// TODO: 可以使用 runtime.Screen API 获取实际屏幕尺寸进行更精确的验证
	// 当前使用保守策略: 只要位置为正数且尺寸合理就认为有效
	const (
		minWidth  = 400
		minHeight = 300
		maxCoord  = 10000 // 防止异常值
	)

	return x >= 0 && y >= 0 && width >= minWidth && height >= minHeight &&
		x < maxCoord && y < maxCoord
}

// onBeforeClose 在窗口关闭前调用
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	// 保存窗口状态
	if err := a.saveWindowState(); err != nil {
		runtime.LogError(a.ctx, "保存窗口状态失败: "+err.Error())
		// 即使保存失败也允许关闭
	}

	return false // 允许窗口关闭
}

// saveWindowState 保存窗口状态
func (a *App) saveWindowState() error {
	// 获取当前窗口状态
	x, y := runtime.WindowGetPosition(a.ctx)
	width, height := runtime.WindowGetSize(a.ctx)
	isMaximized := runtime.WindowIsMaximised(a.ctx)

	state := &WindowState{
		Key:   "window.main",
		X:     x,
		Y:     y,
		Width: width,
		Height: height,
		Maximized: isMaximized,
	}

	// 保存到数据库(使用 Upsert 更新或插入)
	if err := a.windowStateRepo.Save(state); err != nil {
		return err
	}

	runtime.LogInfo(a.ctx, "窗口状态已保存")
	return nil
}
```

**Key Points:**
- **保存时机**: 在 `OnBeforeClose` 回调中保存,确保捕获最后的状态
- **恢复时机**: 在 `OnStartup` 中恢复,在窗口显示前设置位置和大小(避免闪烁)
- **数据验证**: 恢复前验证位置有效性,防止窗口"丢失"在屏幕外
- **最大化状态**: 最大化状态要在设置位置和大小后再应用
- **错误处理**: 保存失败时记录错误但不阻塞关闭,恢复失败时使用默认位置

### Pattern 3: WindowState Repository Pattern

**What:** 使用 Repository 模式封装窗口状态的数据访问逻辑。

**When to use:** 需要解耦数据访问层和业务逻辑的场景。

**Example:**

```go
// Source: GORM 最佳实践

package repository

import (
	"gorm.io/gorm"
)

type WindowState struct {
	gorm.Model
	Key       string `gorm:"uniqueIndex;not null"`
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool
	MonitorID string
}

type WindowStateRepository struct {
	db *gorm.DB
}

func NewWindowStateRepository() *WindowStateRepository {
	return &WindowStateRepository{
		db: GetDB(),
	}
}

// GetByKey 根据 key 获取窗口状态
func (r *WindowStateRepository) GetByKey(key string) (*WindowState, error) {
	var state WindowState
	err := r.db.Where("key = ?", key).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// Save 保存或更新窗口状态(Upsert)
func (r *WindowStateRepository) Save(state *WindowState) error {
	// 使用 GORM 的 Save 方法(自动处理 Insert 或 Update)
	return r.db.Save(state).Error
}

// DeleteByKey 删除指定 key 的窗口状态
func (r *WindowStateRepository) DeleteByKey(key string) error {
	return r.db.Where("key = ?", key).Delete(&WindowState{}).Error
}
```

**Key Points:**
- 使用 `gorm:"uniqueIndex"` 确保 key 唯一性
- 使用 GORM 的 `Save` 方法自动处理 Upsert 逻辑
- 封装数据访问逻辑,便于测试和维护

### Anti-Patterns to Avoid

- **手动实现锁机制**: 不要手动使用 Windows API (CreateMutex) 或文件锁,Wails 已提供跨平台抽象
- **在 OnSecondInstanceLaunch 中显示通知**: 不要显示"已有实例运行"的对话框,应静默激活窗口
- **忽略数据验证**: 不要直接应用保存的窗口位置,必须验证有效性(防止窗口在屏幕外)
- **在 OnShutdown 中保存状态**: 不要在 `OnShutdown` 中保存,使用 `OnBeforeClose` 确保"隐藏到托盘"时也保存状态
- **保存失败时阻止关闭**: 保存窗口状态失败不应阻止应用关闭,应记录错误并继续

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 单实例锁定 | 手动实现 Windows Mutex/macOS BSD mutex/Linux flock | Wails SingleInstanceLock | 跨平台兼容性,自动处理进程间通信 |
| 进程间通信 | 手动使用 Windows SendMessage/Linux DBus | SingleInstanceLock 的 OnSecondInstanceLaunch 回调 | Wails 自动传递第二个实例的参数和工作目录 |
| 窗口状态存储 | 手动序列化 JSON/YAML 到文件 | GORM + SQLite | 并发安全、事务支持、类型安全、项目已集成 |
| UUID 生成 | 手动拼接字符串或使用时间戳 | github.com/google/uuid | 避免碰撞,符合 UUID 标准 |

**Key insight:** Wails 已为常见桌面应用需求提供内置解决方案,手动实现这些功能会增加维护负担且容易引入跨平台兼容性问题。

## Common Pitfalls

### Pitfall 1: OnSecondInstanceLaunch 不自动激活窗口

**What goes wrong:** 配置了 `SingleInstanceLock`,但第二个实例启动时现有窗口没有激活到前台。

**Why it happens:** Wails 的 `OnSecondInstanceLaunch` 回调仅用于通知,不会自动执行窗口激活操作。开发者必须在回调中手动调用 `runtime.WindowShow` 和 `runtime.WindowUnminimise`。

**How to avoid:**
```go
func (a *App) onSecondInstanceLaunch(data options.SecondInstanceData) {
	// 必须手动调用这两个方法
	runtime.WindowUnminimise(a.ctx) // 先恢复最小化窗口
	runtime.WindowShow(a.ctx)        // 再显示并激活到前台
}
```

**Warning signs:** 第二个实例启动后无任何反应,或现有窗口仍保持最小化/隐藏状态。

### Pitfall 2: 窗口位置在屏幕外导致窗口"丢失"

**What goes wrong:** 用户在多显示器环境下使用后,拔掉显示器,下次启动时窗口出现在不存在的显示器位置,用户看不到窗口。

**Why it happens:** 保存的窗口坐标可能指向已断开的显示器,或者坐标超出当前屏幕范围。

**How to avoid:**
```go
func (a *App) isPositionValid(x, y, width, height int) bool {
	// 简单验证: 窗口至少部分在主屏幕内
	// TODO: 可以使用 runtime.ScreenGetAll() 获取所有显示器信息进行更精确验证
	const (
		minWidth  = 400
		minHeight = 300
		maxCoord  = 10000
	)

	return x >= 0 && y >= 0 && width >= minWidth && height >= minHeight &&
		x < maxCoord && y < maxCoord
}

func (a *App) restoreWindowState() error {
	state, err := a.windowStateRepo.GetByKey("window.main")
	if err != nil {
		return err
	}

	// 关键: 恢复前验证位置
	if !a.isPositionValid(state.X, state.Y, state.Width, state.Height) {
		runtime.LogWarning(a.ctx, "保存的窗口位置无效,使用默认位置")
		return nil // 不应用无效的位置
	}

	runtime.WindowSetPosition(a.ctx, state.X, state.Y)
	runtime.WindowSetSize(a.ctx, state.Width, state.Height)
	// ...
}
```

**Warning signs:** 用户反馈"窗口不见了"或"应用启动后看不到界面"。

### Pitfall 3: 在 OnShutdown 而非 OnBeforeClose 中保存状态

**What goes wrong:** 应用"隐藏到托盘"时窗口状态未保存,下次启动时窗口位置不正确。

**Why it happens:** `OnShutdown` 只在应用完全退出时调用,`OnBeforeClose` 在每次窗口关闭时调用(包括隐藏到托盘)。如果应用实现了"隐藏到托盘"功能(如本项目的 `onBeforeClose` 返回 `true` 阻止关闭),则 `OnShutdown` 不会被调用。

**How to avoid:**
```go
// 错误做法: 在 OnShutdown 中保存
func (a *App) shutdown(ctx context.Context) {
	// 这只在应用完全退出时调用,隐藏到托盘时不会执行
	a.saveWindowState()
}

// 正确做法: 在 OnBeforeClose 中保存
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	// 每次窗口关闭时都会执行,包括隐藏到托盘
	a.saveWindowState()

	// 如果要隐藏到托盘,返回 true
	if !a.quitting.Load() {
		a.hideWindow()
		return true // 阻止关闭
	}

	return false // 允许关闭
}
```

**Warning signs:** 窗口状态只在完全退出应用时保存,隐藏到托盘后重启应用时状态丢失。

### Pitfall 4: UniqueId 碰撞导致多个应用无法同时运行

**What goes wrong:** 使用简单的应用名称(如 "ai-commit-hub")作为 `UniqueId`,如果其他应用也使用相同的 ID,会导致两者无法同时运行。

**Why it happens:** Wails 使用 `UniqueId` 生成操作系统级别的锁名称(Windows: Mutex, macOS: Mutex, Linux: DBus name)。如果两个应用使用相同的 `UniqueId`,它们会互相冲突。

**How to avoid:**
```go
import "github.com/google/uuid"

func main() {
	// 生成或使用固定的 UUID
	appID := "e3984e08-28dc-4e3d-b70a-45e961589cdc" // 建议使用 uuid.New().String() 生成

	err := wails.Run(&options.App{
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: appID, // 使用 UUID 确保全局唯一
			// ...
		},
	})
}
```

**Warning signs:** 用户反馈"安装了你的应用后,另一个应用无法启动了"或"两个应用互相干扰"。

### Pitfall 5: 保存失败时阻塞应用关闭

**What goes wrong:** 窗口状态保存失败(如磁盘满、数据库锁定),导致应用无法关闭。

**Why it happens:** 在 `OnBeforeClose` 中如果保存逻辑抛出异常或返回错误,可能会影响窗口关闭流程。

**How to avoid:**
```go
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	// 使用 defer recover 确保不会因为保存失败而崩溃
	defer func() {
		if r := recover(); r != nil {
			runtime.LogError(a.ctx, "保存窗口状态时发生 panic: "+fmt.Sprint(r))
		}
	}()

	if err := a.saveWindowState(); err != nil {
		// 记录错误但不阻塞关闭
		runtime.LogError(a.ctx, "保存窗口状态失败: "+err.Error())
		// 不返回错误,继续关闭流程
	}

	// 其他关闭逻辑...
	return false // 允许关闭
}
```

**Warning signs:** 用户反馈"应用关闭时卡住"或"点击关闭按钮没反应"。

## Code Examples

Verified patterns from official sources:

### Example 1: Complete SingleInstanceLock Configuration

```go
// Source: Wails v2.10 官方文档
// URL: https://wails.io/docs/v2.10/guides/single-instance-lock

package main

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	// 静默激活窗口,不显示通知
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)

	// 如果需要处理第二个实例的参数,可以发送事件到前端
	// runtime.EventsEmit(a.ctx, "second-instance-args", secondInstanceData.Args)
}

func main() {
	app := &App{}

	err := wails.Run(&options.App{
		Title:  "My App",
		Width:  1024,
		Height: 768,
		OnStartup: app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "my-unique-app-id-12345", // 使用 UUID
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{app},
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
```

### Example 2: Window State Save and Restore

```go
// Source: Wails Runtime API 文档 + GNOME Developer Documentation 最佳实践
// URL: https://wails.io/docs/v2.10/reference/runtime/window
// URL: https://developer.gnome.org/documentation/tutorials/save-state.html

package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	// 保存窗口状态
	x, y := runtime.WindowGetPosition(a.ctx)
	width, height := runtime.WindowGetSize(a.ctx)
	isMaximized := runtime.WindowIsMaximised(a.ctx)

	state := &WindowState{
		Key:       "window.main",
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximized: isMaximized,
	}

	if err := a.windowStateRepo.Save(state); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("保存窗口状态失败: %v", err))
		// 不阻止关闭
	}

	return false
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 尝试恢复窗口状态
	state, err := a.windowStateRepo.GetByKey("window.main")
	if err != nil {
		runtime.LogInfo(a.ctx, "首次启动,使用默认窗口位置")
		return
	}

	// 验证位置有效性
	if !a.isPositionValid(state.X, state.Y, state.Width, state.Height) {
		runtime.LogWarning(a.ctx, "保存的窗口位置无效,使用默认位置")
		return
	}

	// 恢复窗口状态
	runtime.WindowSetPosition(a.ctx, state.X, state.Y)
	runtime.WindowSetSize(a.ctx, state.Width, state.Height)

	if state.Maximized {
		runtime.WindowMaximise(a.ctx)
	}

	runtime.LogInfo(a.ctx, "窗口状态已恢复")
}
```

### Example 3: Database Migration for WindowState Table

```go
// Source: GORM AutoMigrate 最佳实践
// URL: https://gorm.io/docs/create.html

package repository

import (
	"gorm.io/gorm"
)

type WindowState struct {
	gorm.Model
	Key       string `gorm:"uniqueIndex;not null;size:100"`
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool
	MonitorID string `gorm:"size:50"`
}

// MigrateWindowStateTable 创建或更新 window_states 表
func MigrateWindowStateTable(db *gorm.DB) error {
	return db.AutoMigrate(&WindowState{})
}

// 在主应用初始化时调用迁移
func InitializeDatabase() {
	// ...
	db := repository.GetDB()

	// 迁移窗口状态表
	if err := repository.MigrateWindowStateTable(db); err != nil {
		logger.Errorf("窗口状态表迁移失败: %v", err)
		// 不阻塞启动,窗口状态持久化是可选功能
	}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 手动实现互斥量/文件锁 | Wails SingleInstanceLock | Wails v2.0+ | 大幅简化代码,提升跨平台兼容性 |
| OnShutdown 保存状态 | OnBeforeClose 保存状态 | 隐藏到托盘功能普及后 | 确保隐藏到托盘时也保存状态 |
| JSON 文件存储窗口状态 | GORM + SQLite | 项目已集成 GORM | 利用现有基础设施,类型安全,事务支持 |
| 位置验证依赖启发式算法 | runtime.ScreenGetAll() 精确验证 | Wails v2.10+ | 支持多显示器场景,更准确(可选实现) |

**Deprecated/outdated:**
- 手动实现 Windows API 调用(CreateMutex, SendMessage): Wails 提供的抽象层已足够
- 使用 OnShutdown 保存窗口状态: 在隐藏到托盘场景下不会触发
- 保存到 JSON/YAML 文件: 项目已使用 GORM,应统一数据存储方案

## Open Questions

1. **窗口状态验证的精确阈值**
   - **What we know:** 需要验证窗口位置是否在屏幕范围内,防止窗口"丢失"
   - **What's unclear:** 使用什么阈值判断"超出屏幕"(如负坐标、超出屏幕边界的像素数)
   - **Recommendation:** 实现阶段使用保守策略(坐标 >= 0,尺寸 >= 最小窗口尺寸),未来可使用 `runtime.ScreenGetAll()` 获取实际显示器信息进行精确验证

2. **多显示器场景的处理策略**
   - **What we know:** 用户可能在多显示器环境下使用应用,拔掉显示器后窗口可能"丢失"
   - **What's unclear:** 如果原显示器不存在,是移到主显示器还是使用系统默认位置
   - **Recommendation:** 遵循 CONTEXT.md 的决策:"如果原显示器不存在,使用系统默认位置(不尝试移到主显示器)"

3. **错误对话框的具体文案和样式**
   - **What we know:** CONTEXT.md 规定"激活失败:显示错误对话框,包含技术详情(错误信息)"
   - **What's unclear:** 对话框的具体文案(中英文)、按钮、图标样式
   - **Recommendation:** 实现阶段由开发者根据项目风格确定,建议使用 `runtime.MessageDialog` 显示标准错误对话框,包含技术错误信息和用户友好的描述

4. **数据库表结构的字段命名**
   - **What we know:** 需要存储窗口位置(X,Y)、大小(宽高)、最大化状态、显示器编号
   - **What's unclear:** 字段命名(snake_case vs camelCase)、数据类型(int vs int64)、是否需要索引
   - **Recommendation:** 遵循 GORM 惯例:
     - 表名: `window_states` (复数形式)
     - 字段名: snake_case(x, y, width, height, maximized, monitor_id)
     - 数据类型: int (窗口坐标和尺寸不会超过 int 范围)
     - 索引: key 字段添加 `uniqueIndex` 确保查询性能

## Sources

### Primary (HIGH confidence)
- **Context7: /wailsapp/wails** - SingleInstanceLock configuration, OnSecondInstanceLaunch callback, Window Runtime APIs (WindowShow, WindowUnminimise, WindowSetSize, WindowSetPosition, WindowGetSize, WindowGetPosition, WindowIsMaximised)
- **Wails Official Documentation** - https://wails.io/docs/v2.10/guides/single-instance-lock
- **Wails API Reference** - https://wails.io/docs/v2.10/reference/options/ (SingleInstanceLock options)
- **Wails Runtime API** - https://wails.io/docs/v2.10/reference/runtime/window/

### Secondary (MEDIUM confidence)
- **GNOME Developer Documentation** - https://developer.gnome.org/documentation/tutorials/save-state.html (Window state persistence best practices)
- **GORM Documentation** - https://gorm.io/docs/create.html (AutoMigrate and model definitions)
- **GORM Best Practices** - https://www.pingcap.com/article/building-robust-go-applications-with-gorm-best-practices/ (Repository pattern and data validation)

### Tertiary (LOW confidence)
- **Microsoft Learn - WPF Samples** - https://learn.microsoft.com/en-us/samples/microsoft/wpf-samples/save-window-placement-state-sample/ (Window state persistence patterns,非 Wails 特定)
- **Go Packages - settingstore** - https://pkg.go.dev/github.com/gouniverse/settingstore (Key-value settings storage in SQL databases)

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - Wails 官方文档和 Context7 代码示例明确支持 SingleInstanceLock 和 Runtime APIs
- Architecture: **HIGH** - 基于 Wails 官方推荐模式和项目现有架构(GORM + SQLite)
- Pitfalls: **HIGH** - 基于 Wails 官方文档警告和 GNOME Developer Documentation 最佳实践

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (30 days - Wails v2 是稳定版本,API 不会频繁变更)
