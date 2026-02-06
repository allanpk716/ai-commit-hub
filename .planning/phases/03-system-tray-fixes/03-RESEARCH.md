# Phase 3: System Tray Fixes - Research

**Researched:** 2026-02-06
**Domain:** Go system tray libraries (systray), Wails desktop application lifecycle
**Confidence:** HIGH

## Summary

Phase 3 需要修复和增强系统托盘功能。现有代码使用 `getlantern/systray v1.2.2`，该版本不支持托盘图标的双击事件。项目已有丰富的托盘实现经验（详见 `docs/lessons-learned/windows-tray-icon-implementation-guide.md`），但需要升级到支持双击功能的 fork 版本。

研究发现了两个主要的 systray fork 库支持双击功能：
1. **lutischan-ferenc/systray** - 支持 `SetOnDClick()`、`SetOnClick()`、`SetOnRClick()` API
2. **energye/systray** - 支持相同 API，移除 GTK 依赖，更现代的实现

两个库的 API 几乎相同，都使用回调函数替代原有的 `ClickedCh` 通道模式。

**Primary recommendation:** 使用 `energye/systray` 作为主选方案（活跃维护、移除 GTK 依赖），`lutischan-ferenc/systray` 作为备选方案。

## Standard Stack

### Core Libraries

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **energye/systray** | v1.0.3+ (latest) | System tray with click/double-click support | ✅ Active fork of getlantern/systray<br>✅ Removes GTK dependency (better cross-platform)<br>✅ Supports Windows, macOS, Linux<br>✅ Double-click API: `SetOnDClick(fn func())` |
| **lutischan-ferenc/systray** | v1.3.0 (backup) | Alternative systray implementation | ✅ Proven in production (used in previous fixes)<br>✅ API-compatible with energye/systray<br>✅ Good fallback if energye has issues |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **sync/atomic** | Go stdlib | `atomic.Bool` for quitting flag | ✅ Prevent race conditions in exit logic |
| **sync** | Go stdlib | `sync.Once` for one-time operations | ✅ Ensure systray exits only once |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| energye/systray | lutischan-ferenc/systray | energye removes GTK dependency (cleaner), lutischan has longer track record |
| Callback API | Channel API (getlantern v1.2.2) | Callback is more modern, Channel is deprecated pattern |
| Double-click event | Single-click with toggle | Double-click is standard UX for tray apps, single-click may trigger accidentally |

**Installation:**
```bash
# Primary choice
go get github.com/energye/systray@latest
go mod tidy

# Alternative (if needed)
go get github.com/lutischan-ferenc/systray@v1.3.0
go mod tidy
```

## Architecture Patterns

### Recommended Project Structure

Existing code is already well-organized. Changes will be localized to:

```
app.go                          # Systray integration (lines ~21, 260-420)
├── runSystray()                # Initialize systray goroutine
├── onSystrayReady()            # Set icon, menu, click handlers
├── onSystrayExit()             # Cleanup callback
├── showWindow()                # Window state management (reuse existing)
├── hideWindow()                # Window state management (reuse existing)
└── quitApplication()           # Exit logic (reuse existing)

tray_icon.go                    # Icon utilities (no changes needed)
├── windowsICOFromPNGOnce()     # Generate multi-size ICO
├── icoSquareSizes()            # Validate ICO format
└── generateRedBoxICO()         # Fallback icon
```

### Pattern 1: Systray Library Upgrade Path

**What:** Upgrade from getlantern/systray v1.2.2 to energye/systray v1.0.3+

**When to use:** Phase 3 implementation requires double-click tray icon functionality

**Migration steps:**

1. **Update go.mod:**
```go
// Remove (or keep as indirect)
// github.com/getlantern/systray v1.2.2

// Add
github.com/energye/systray v1.0.3
```

2. **Update import (app.go:21):**
```go
// Old
import "github.com/getlantern/systray"

// New
import "github.com/energye/systray"
```

3. **Update API calls (app.go:377-397):**
```go
// Old API (Channel-based)
menu := systray.AddMenuItem("显示窗口", "显示主窗口")
go func() {
    for range menu.ClickedCh {  // ❌ Deprecated
        a.showWindow()
    }
}()

// New API (Callback-based)
menu := systray.AddMenuItem("显示窗口", "显示主窗口")
menu.Click(func() {  // ✅ Modern callback
    a.showWindow()
})
```

4. **Add double-click handler (app.go:375):**
```go
func (a *App) onSystrayReady() {
    // ... set icon, tooltip ...

    // NEW: Double-click to show window
    systray.SetOnDClick(func() {
        logger.Info("托盘图标双击，显示窗口")
        a.showWindow()
    })

    // NEW: Right-click to show menu
    systray.SetOnRClick(func(menu systray.IMenu) {
        menu.ShowMenu()
    })

    // Menu items...
}
```

### Pattern 2: Race Condition Prevention

**What:** Use `sync.Once` and `atomic.Bool` to prevent concurrent systray operations

**When to use:** Already implemented in app.go, keep and verify during upgrade

**Example:**
```go
type App struct {
    systrayExit    *sync.Once     // ✅ Ensures quit only executes once
    systrayRunning atomic.Bool    // ✅ Thread-safe running state
    quitting       atomic.Bool    // ✅ Thread-safe exit flag
}

func (a *App) quitApplication() {
    a.systrayExit.Do(func() {  // ✅ Guarantees single execution
        a.quitting.Store(true)
        a.showWindow()
        systray.Quit()
    })
}
```

### Pattern 3: Menu Handler Isolation

**What:** Each menu item gets its own goroutine for click handling

**When to use:** Prevent blocking in menu click handlers

**Example:**
```go
// GOOD: Isolated goroutines
menu := systray.AddMenuItem("显示窗口", "显示主窗口")
menu.Click(func() {
    a.showWindow()  // Runs in its own context
})

quitMenu := systray.AddMenuItem("退出应用", "完全退出应用")
quitMenu.Click(func() {
    a.quitApplication()  // Isolated from showWindow
})
```

### Anti-Patterns to Avoid

- **Blocking in onSystrayReady:** Don't perform long-running operations in onSystrayReady
  ```go
  // ❌ BAD
  func (a *App) onSystrayReady() {
      time.Sleep(5 * time.Second)  // Blocks systray init
      systray.SetIcon(icon)
  }

  // ✅ GOOD
  func (a *App) onSystrayReady() {
      systray.SetIcon(icon)  // Fast, non-blocking
  }
  ```

- **Direct runtime calls without state sync:** Don't call runtime.WindowShow() directly
  ```go
  // ❌ BAD
  systray.SetOnDClick(func() {
      runtime.WindowShow(a.ctx)  // Breaks windowVisible sync
  })

  // ✅ GOOD
  systray.SetOnDClick(func() {
      a.showWindow()  // Maintains windowVisible state
  })
  ```

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Double-click detection | Time-based click counting | `systray.SetOnDClick(fn)` | OS handles double-click timing correctly (configurable by user) |
| Thread-safe exit flags | Custom mutex-based flags | `atomic.Bool` | Lock-free, faster, guaranteed atomic operations |
| Single execution guarantee | Custom "executed" bool | `sync.Once` | Built-in, thread-safe, idiomatic Go |
| Icon format validation | Manual byte parsing | `looksLikeICO()`, `icoSquareSizes()` (existing) | Already implemented in tray_icon.go |

**Key insight:** System tray interactions are OS-specific and complex. Always use library-provided APIs rather than reimplementing OS behavior.

## Common Pitfalls

### Pitfall 1: Import Path Mismatch After Upgrade

**What goes wrong:** Upgrade go.mod but forget to update import statements in code

**Why it happens:** Go module cache may have both old and new versions

**How to avoid:**
```bash
# After updating go.mod
go mod tidy  # Cleans up unused imports
go build -o build/bin/ai-commit-hub.exe .  # Verify compilation
```

**Warning signs:**
- Compile error: "cannot find package"
- IDE shows red underline on import
- `go build` fails with module resolution error

### Pitfall 2: Channel-to-Callback Migration Bugs

**What goes wrong:** Forget to update `ClickedCh` loops to `Click()` callbacks

**Why it happens:** Muscle memory from old API pattern

**How to avoid:**
```go
// ❌ WRONG: Old pattern with new library
menu := systray.AddMenuItem("Show", "Show window")
go func() {
    for range menu.ClickedCh {  // ❌ Field doesn't exist in new API
        a.showWindow()
    }
}()

// ✅ CORRECT: New callback pattern
menu := systray.AddMenuItem("Show", "Show window")
menu.Click(func() {  // ✅ Method exists in new API
    a.showWindow()
})
```

**Warning signs:**
- Compile error: "menu.ClickedCh undefined"
- Runtime panic: "no field or method ClickedCh"

### Pitfall 3: Double-Click Not Firing on Windows

**What goes wrong:** Double-click handler registered but never triggered

**Why it happens:** Some systray implementations require `SetMenuNil()` or special initialization

**How to avoid:**
```go
func (a *App) onSystrayReady() {
    // Set double-click BEFORE adding menu items
    systray.SetOnDClick(func() {
        logger.Info("托盘图标双击")
        a.showWindow()
    })

    // Then add menu items
    menu := systray.AddMenuItem("显示窗口", "显示主窗口")
    menu.Click(func() {
        a.showWindow()
    })
}
```

**Warning signs:**
- Double-click does nothing
- Single-click works fine
- Menu items work correctly

### Pitfall 4: Race Condition in Exit Logic

**What goes wrong:** `onBeforeClose` intercepts quit event, creating infinite loop

**Why it happens:** Exiting without setting `quitting` flag

**How to avoid:**
```go
// ✅ CORRECT: Always set quitting flag first
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        a.quitting.Store(true)  // ✅ MUST be first
        a.showWindow()
        systray.Quit()
    })
}

func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    if a.quitting.Load() {  // ✅ Check flag
        return false  // Allow close
    }
    a.hideWindow()
    return true  // Prevent close (minimize to tray)
}
```

**Warning signs:**
- Clicking "退出应用" doesn't close app
- App shows window then immediately hides it
- Log shows repeated "应用正在退出" messages

### Pitfall 5: goroutine Leak in Click Handlers

**What goes wrong:** Click handler spawns goroutines that never exit

**Why it happens:** Uncleaned resources or infinite loops

**How to avoid:**
```go
// ✅ GOOD: Callbacks are self-cleaning
menu.Click(func() {
    a.showWindow()  // Synchronous, no goroutine spawn
})

// ❌ BAD: Spawns untracked goroutine
menu.Click(func() {
    go func() {  // ❌ Untracked goroutine
        for {
            time.Sleep(1 * time.Second)
            // Do something...
        }
    }()
})
```

**Warning signs:**
- Memory usage increases over time
- Goroutine count grows (use runtime.NumGoroutine())
- App becomes sluggish after many clicks

## Code Examples

Verified patterns from official sources:

### Example 1: Basic Double-Click Setup

**Source:** [energye/systray GitHub README](https://github.com/energye/systray)

```go
package main

import (
    "fmt"
    "github.com/energye/systray"
)

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {
    // Set icon, tooltip
    systray.SetIcon(iconData)
    systray.SetTooltip("Awesome App")

    // Double-click handler
    systray.SetOnDClick(func() {
        fmt.Println("Double-clicked!")
    })

    // Right-click handler (show menu)
    systray.SetOnRClick(func(menu systray.IMenu) {
        menu.ShowMenu()
    })

    // Menu items
    mQuit := systray.AddMenuItem("Quit", "Quit the app")
    mQuit.Click(func() {
        systray.Quit()
    })
}

func onExit() {
    // Cleanup
}
```

### Example 2: Integration with Existing Window State

**Source:** Project's app.go (lines 414-450)

```go
// showWindow 显示窗口（正确维护状态）
func (a *App) showWindow() {
    if a.ctx == nil {
        logger.Warn("showWindow: context 未初始化")
        return
    }

    // 使用 WindowUnminimise 恢复最小化的窗口
    runtime.WindowUnminimise(a.ctx)

    // 显示窗口
    runtime.WindowShow(a.ctx)

    // 更新窗口可见状态（重要：避免关闭按钮失效）
    a.windowMutex.Lock()
    a.windowVisible = true
    a.windowMutex.Unlock()

    // 激活窗口到前台
    runtime.WindowShow(a.ctx)
}

// hideWindow 隐藏窗口（正确维护状态）
func (a *App) hideWindow() {
    if a.ctx == nil {
        logger.Warn("hideWindow: context 未初始化")
        return
    }

    // 隐藏窗口
    runtime.WindowHide(a.ctx)

    // 更新窗口可见状态
    a.windowMutex.Lock()
    a.windowVisible = false
    a.windowMutex.Unlock()
}
```

### Example 3: Thread-Safe Exit Logic

**Source:** Project's app.go (lines 452-468)

```go
// quitApplication 退出应用（使用 sync.Once 和 atomic.Bool）
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        logger.Info("应用正在退出...")

        // 设置退出标志，防止 onBeforeClose 拦截
        a.quitting.Store(true)

        // 先显示窗口（如果当前隐藏），避免用户看到应用"卡住"
        a.showWindow()

        // 调用 systray.Quit() 触发退出流程
        systray.Quit()
    })
}

// onBeforeClose 关闭拦截逻辑
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    // 如果应用正在退出，不拦截关闭事件
    if a.quitting.Load() {
        logger.Info("应用正在退出,允许窗口关闭")
        return false  // ✅ 允许关闭
    }

    logger.Info("窗口关闭事件被触发,将隐藏到托盘")

    // 隐藏窗口而非退出
    a.hideWindow()

    // 返回 true 阻止窗口关闭
    return true
}
```

### Example 4: Icon Multi-Level Fallback

**Source:** Project's app.go (lines 295-321) and tray_icon.go

```go
// getTrayIcon 根据平台返回合适的图标（多级回退策略）
func (a *App) getTrayIcon() []byte {
    if stdruntime.GOOS == "windows" {
        // Level 1: 优先使用 Wails 生成的嵌入 ICO
        if len(appIconICO) > 0 {
            if sizes, err := icoSquareSizes(appIconICO); err == nil {
                logger.Info("使用 Wails 生成的托盘图标 ICO", "sizes", sizes)
                return appIconICO
            } else {
                logger.Warnf("嵌入 ICO 格式验证失败: %v", err)
            }
        }

        // Level 2: 从 PNG 动态生成 ICO
        if len(appIconPNG) > 0 {
            if generated, err := windowsICOFromPNGOnce(appIconPNG); err == nil && len(generated) > 0 {
                logger.Info("从 PNG 生成的托盘图标", "bytes", len(generated))
                return generated
            } else if err != nil {
                logger.Errorf("从 PNG 生成 ICO 失败: %v", err)
            }
        }

        // Level 3: 红色占位 ICO
        logger.Warn("无法获取托盘图标，使用红色占位图标")
        return generateRedBoxICO()
    }

    return appIconPNG
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| getlantern/systray v1.2.2 (Channel API) | energye/systray v1.0.3+ (Callback API) | Phase 3 (upcoming) | Enables double-click, modern Go patterns |
| Single-click to show menu | Double-click to show window | Phase 3 (upcoming) | Improved UX, faster access to main window |
| No native double-click support | OS-native double-click handling | Phase 3 (upcoming) | Consistent with platform conventions |

**Deprecated/outdated:**
- **ClickedCh channel pattern**: Replaced by `Click(fn func())` callback
- **getlantern/systray v1.2.2**: Lacks double-click support, superseded by forks
- **GTK dependency (Linux)**: energye/systray removes GTK, uses pure DBus

## Open Questions

### 1. Which systray fork to use? (RESOLVED)

**Question:** energye/systray vs lutischan-ferenc/systray?

**Answer:**
- **Primary choice:** `energye/systray` (latest, removes GTK dependency)
- **Fallback:** `lutischan-ferenc/systray` (proven in docs/fixes/tray-icon-doubleclick-fix.md)

Both have identical APIs. Try energye first, fall back to lutischan if issues arise.

**Recommendation:** Start with energye/systray, document any issues for future reference.

### 2. Double-click timing configuration (LOW PRIORITY)

**What we know:** lutischan-ferenc/systray provides `SetDClickTimeMinInterval(value int64)` to configure double-click detection window (default 500ms)

**What's unclear:** Does energye/systray also support this API? (Not documented in README)

**Recommendation:** If double-click feels unresponsive, test `SetDClickTimeMinInterval(300)` for faster detection. Otherwise, use default 500ms.

### 3. Cross-platform double-click behavior (VERIFIED)

**What we know:**
- Windows: ✅ Double-click supported (both libraries)
- macOS: ✅ Double-click supported (both libraries)
- Linux: ⚠️ May vary by desktop environment

**Recommendation:** Test on target platforms, document any Linux-specific issues.

## Sources

### Primary (HIGH confidence)

- [energye/systray GitHub Repository](https://github.com/energye/systray) - Official README, API documentation
- [lutischan-ferenc/systray pkg.go.dev](https://pkg.go.dev/github.com/lutischan-ferenc/systray) - Full API reference
- [getlantern/systray GitHub](https://github.com/getlantern/systray) - Original library (for context)
- [Project docs/lessons-learned/windows-tray-icon-implementation-guide.md](C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\lessons-learned\windows-tray-icon-implementation-guide.md) - Comprehensive project experience (1290 lines)
- [Project docs/fixes/tray-icon-doubleclick-fix.md](C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\fixes\tray-icon-doubleclick-fix.md) - Previous double-click implementation attempt
- [Project docs/fixes/systray-exit-fix.md](C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\fixes\systray-exit-fix.md) - Exit logic implementation details

### Secondary (MEDIUM confidence)

- [getlantern/systray Issue #30 - On click tray event](https://github.com/getlantern/systray/issues/30) - Original feature request for click events
- [Wails Issue #1521 - Support Tray Menus](https://github.com/wailsapp/wails/issues/1521) - Discussion of systray forks with double-click support
- [Go System Tray Implementation Guide](https://dev.to/osuka42/building-a-simple-system-tray-app-with-go-899) - Tutorial on systray patterns
- [Building a Tray/GUI Desktop Application in Go](https://owenmoore.hashnode.dev/build-tray-gui-desktop-application-go) - Best practices for tray apps

### Tertiary (LOW confidence)

- [CSDN Blog: go 使用systray 实现托盘和程序退出](https://blog.csdn.net/weixin_42094764/article/details/132622482) - Chinese blog post (community pattern)
- [iTYing Forum: golang跨平台通知区域图标和菜单插件库systray的使用](https://bbs.itying.com/topic/6876cd694715aa0088487c86) - Chinese forum discussion
- [Reddit: Toolkit-agnostic system tray in Go](https://www.reddit.com/r/golang/comments/t8perr/toolkitagnostic_system_tray_in_go_based_on/) - Community discussion

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Multiple official sources and forks verified
- Architecture: HIGH - Project has existing implementation (Phase 2 complete), patterns proven in production
- Pitfalls: HIGH - Comprehensive project documentation (1290-line implementation guide)
- Migration path: HIGH - API differences clearly documented, backward-compatible

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (30 days - systray libraries are stable)

**Dependencies on other phases:**
- **Phase 2 (Single Instance & Window Management):** ✅ Complete (verified 2026-02-06)
- Phase 3 will reuse window state management methods (showWindow/hideWindow) from Phase 2
- No additional dependencies identified

**Key recommendations for planner:**
1. Create tasks for systray library upgrade (go.mod + import changes)
2. Create tasks for API migration (ClickedCh → Click callback)
3. Create tasks for double-click handler implementation
4. Reuse existing exit logic (sync.Once + atomic.Bool) - already proven
5. Reuse existing icon fallback logic - already proven
6. Test on all target platforms (Windows, macOS if applicable)
7. Verify double-click timing feels responsive (adjust if needed)
