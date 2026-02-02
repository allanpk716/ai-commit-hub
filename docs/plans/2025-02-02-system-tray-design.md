# 系统托盘功能设计方案

**日期:** 2025-02-02
**状态:** 设计完成
**预计工时:** 2 小时

## 概述

为 AI Commit Hub 添加系统托盘功能,允许用户关闭窗口后应用继续在后台运行,并通过系统托盘图标重新打开或完全退出应用。

## 需求

### 功能需求

1. **关闭行为**: 用户点击窗口关闭按钮时,应用隐藏到系统托盘而非退出
2. **托盘菜单**: 系统托盘提供两个菜单项
   - "显示窗口" - 恢复显示主窗口
   - "退出应用" - 完全退出应用
3. **启动行为**: 应用启动时显示主窗口(非最小化到托盘)
4. **托盘图标**: 复用现有的应用图标(app-icon.png)

### 非功能需求

- **稳定性**: 防止多次退出导致的竞态条件
- **性能**: 托盘功能内存开销 < 10MB
- **响应速度**: 窗口显示/隐藏响应 < 200ms
- **跨平台**: 优先支持 Windows,兼容 macOS 和 Linux

## 技术方案

### 技术选型

- **框架**: Wails v2 (当前版本,保持稳定)
- **托盘库**: `github.com/getlantern/systray`
  - 成熟稳定的跨平台系统托盘库
  - 已在 to_icalendar 项目中验证可行性
  - API 简洁,文档完善

### 选型理由

✅ **稳定可靠**: systray 库在生产环境中广泛使用,经过充分测试
✅ **跨平台**: 同时支持 Windows、macOS、Linux
✅ **已有经验**: to_icalendar 项目已验证此方案可行
✅ **低风险**: 不需要升级 Wails 版本,不影响现有功能
✅ **实现简单**: API 清晰,代码量少(约 100-150 行)

⚠️ **为什么不选择 Wails v3**:
- Wails v3 目前处于 Alpha 阶段,生产就绪度不足
- v3 是完全重写,迁移成本高(1-4 小时)
- 需要重构大量现有代码
- 可能存在未知问题

参考: [Wails v3 Migration Guide](https://v3alpha.wails.io/migration/v2-to-v3/)

## 架构设计

### 应用生命周期

```
应用启动
    ↓
[启动阶段]
    ├─ 初始化数据库和服务
    ├─ 启动 systray goroutine (后台)
    └─ 显示主窗口 (WindowStartState: Normal)
    ↓
[运行阶段]
    ├─ 用户正常使用主窗口
    ├─ 点击关闭按钮 → 触发 OnBeforeClose → 隐藏窗口 → 保留托盘
    └─ 点击托盘菜单 → 显示/隐藏窗口
    ↓
[退出阶段]
    └─ 托盘菜单选择"退出" → 调用 runtime.Quit → 真正退出
```

### 核心组件

#### 1. Systray Manager (`app.go`)

负责系统托盘的初始化和生命周期管理:

- 托盘图标加载和显示
- 托盘菜单创建和管理
- 菜单项事件处理
- 通过 sync.WaitGroup 确保优雅退出

#### 2. Window Controller (集成在 `App` 结构)

管理窗口显示/隐藏状态:

- 拦截 `OnBeforeClose` 事件
- 窗口显示/隐藏方法
- 通过 Wails Events 通知前端窗口状态变化
- 安全退出机制(防止多次退出)

#### 3. 前端状态同步 (Vue, 可选)

- 监听 `window-visibility-changed` 事件
- 可选:在窗口隐藏时暂停一些耗资源操作
- 可选:添加"最小化到托盘"的提示信息

### 线程模型

```
主线程 (Go)
    ├─ Wails 事件循环
    └─ App 方法调用

Systray Goroutine
    ├─ 托盘图标管理
    ├─ 菜单事件监听
    └─ 通过 channel 与主线程通信

前端 (Vue)
    ├─ UI 渲染
    └─ 监听 Wails Events
```

## 数据结构设计

### App 结构扩展

```go
type App struct {
    ctx             context.Context
    // ... 现有字段

    // 新增托盘相关字段
    systrayReady    chan struct{}     // systray 就绪信号
    systrayExit     *sync.Once        // 确保只退出一次
    windowVisible   bool              // 窗口可见状态
    windowMutex     sync.RWMutex      // 保护 windowVisible
}
```

### 关键方法签名

```go
// main.go
func main() {
    app := NewApp()

    // 在 Wails 启动前,先启动 systray
    go app.runSystray()

    // Wails 配置
    err := wails.Run(&options.App{
        // ... 现有配置
        OnStartup:     app.startup,
        OnBeforeClose: app.onBeforeClose,  // 拦截关闭
        OnShutdown:    app.shutdown,
    })
}

// app.go 新增方法
func (a *App) runSystray()
func (a *App) onBeforeClose(ctx context.Context) (prevent bool)
func (a *App) showWindow()
func (a *App) hideWindow()
func (a *App) quitApplication()
```

## 事件流设计

### 关闭窗口流程

```
用户点击关闭按钮
    ↓
OnBeforeClose 被调用
    ↓
prevent = true (阻止关闭)
    ↓
调用 hideWindow()
    ↓
runtime.WindowHide(ctx)
    ↓
EventsEmit("window-hidden")
    ↓
前端接收事件 (可选处理)
```

### 显示窗口流程

```
用户点击托盘"显示窗口"
    ↓
systray 菜单回调触发
    ↓
调用 showWindow()
    ↓
runtime.WindowShow(ctx)
    ↓
EventsEmit("window-shown")
    ↓
前端接收事件 (可选处理)
```

### 退出应用流程

```
用户点击托盘"退出应用"
    ↓
systray 菜单回调触发
    ↓
调用 quitApplication()
    ↓
sync.Once 确保只执行一次
    ↓
runtime.Quit(ctx)
    ↓
OnShutdown → systray.Quit()
```

## 错误处理和边界情况

### 关键边界场景

#### 1. 多次退出保护

```go
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        logger.Info("应用正在退出...")
        runtime.Quit(a.ctx)
    })
}
```

使用 `sync.Once` 确保退出逻辑只执行一次,防止用户快速点击"退出"导致的竞态条件。

#### 2. Systray 初始化延迟

```go
func (a *App) runSystray() {
    // 延迟初始化,避免与 Wails 启动冲突
    time.Sleep(500 * time.Millisecond)

    systray.Run(
        a.onSystrayReady,
        a.onSystrayExit,
    )
}
```

避免 systray 与 Wails WebView 初始化冲突,参考 to_icalendar 的经验 (commit d3387fd)。

#### 3. 窗口状态同步

```go
func (a *App) hideWindow() {
    a.windowMutex.Lock()
    defer a.windowMutex.Unlock()

    if !a.windowVisible {
        return // 已经隐藏,避免重复操作
    }

    runtime.WindowHide(a.ctx)
    a.windowVisible = false

    runtime.EventsEmit(a.ctx, "window-hidden", map[string]interface{}{
        "timestamp": time.Now(),
    })
}
```

#### 4. 平台兼容性处理

```go
func (a *App) getTrayIcon() []byte {
    // Windows 需要 .ico 格式
    // macOS 需要 PDF/PNG
    // Linux 需要 PNG

    if runtime.GOOS == "windows" {
        return appIcon // 已在 main.go 中嵌入
    }
    return iconBytes
}
```

### 错误处理策略

| 场景 | 处理方式 |
|------|----------|
| Systray 初始化失败 | 记录错误日志,但继续运行应用(降级为普通窗口模式) |
| 窗口显示/隐藏失败 | 记录警告,下次操作时重试 |
| 图标加载失败 | 使用默认占位图标,确保应用仍可运行 |
| 退出时清理失败 | 强制退出 (os.Exit(0)) 作为最后手段 |

### 日志记录

```go
logger.Info("系统托盘初始化成功")
logger.Debugf("窗口状态变更: visible=%v", visible)
logger.Warn("窗口隐藏失败,将重试")
logger.Error("系统托盘初始化失败:", err)
```

## 前端集成(可选)

### 状态监听 (App.vue)

```typescript
onMounted(() => {
  // 监听窗口隐藏事件
  EventsOn('window-hidden', (data: { timestamp: string }) => {
    console.log('[App] 窗口已隐藏到托盘', data.timestamp)

    // 可选:暂停一些耗资源操作
    stopAutoRefresh()
  })

  // 监听窗口显示事件
  EventsOn('window-shown', (data: { timestamp: string }) => {
    console.log('[App] 窗口已从托盘恢复', data.timestamp)

    // 可选:恢复自动刷新
    startAutoRefresh()
  })
})
```

### 用户体验增强

#### 1. 首次关闭提示

```typescript
// 首次点击关闭按钮时显示提示
if (!localStorage.getItem('tray-tip-shown')) {
  showInfoToast('应用已最小化到系统托盘,可以通过托盘图标重新打开')
  localStorage.setItem('tray-tip-shown', 'true')
}
```

#### 2. 窗口状态指示器 (可选)

```vue
<template>
  <div v-if="!windowVisible" class="tray-indicator">
    <span class="icon">🔽</span>
    <span>应用已在托盘运行</span>
  </div>
</template>
```

### 前端改动范围

- ✅ 无需大规模重构
- ✅ 只需添加可选的事件监听和用户提示
- ✅ 不影响现有功能

## 测试计划

### 测试场景

| 场景 | 测试步骤 | 预期结果 |
|------|----------|----------|
| **基本启动** | 启动应用 | 主窗口显示,托盘图标出现 |
| **关闭到托盘** | 点击窗口关闭按钮 | 窗口隐藏,托盘图标保留 |
| **从托盘恢复** | 点击托盘"显示窗口" | 窗口重新显示 |
| **托盘退出** | 点击托盘"退出应用" | 应用完全退出,托盘图标消失 |
| **重复关闭保护** | 快速多次点击关闭按钮 | 窗口只隐藏一次,无错误 |
| **重复退出保护** | 快速多次点击托盘"退出" | 应用只退出一次,无竞态 |
| **长时间运行** | 应用运行 24 小时 | 托盘响应正常,无内存泄漏 |

### 测试环境

- Windows 10/11 (主要目标平台)
- macOS (次要)
- Linux (可选)

### 性能指标

- 托盘初始化时间 < 1 秒
- 窗口显示/隐藏响应 < 200ms
- 内存占用增加 < 10MB (systray 开销)

## 实施步骤

### 步骤 1: 准备工作 (15 分钟)

- [ ] 在 `frontend/src/assets/` 下准备托盘图标文件
- [ ] 确认当前应用图标路径 (`app-icon.png`)
- [ ] 安装依赖: `go get github.com/getlantern/systray`

### 步骤 2: 修改 main.go (20 分钟)

- [ ] 添加 systray 图标的 embed 指令
- [ ] 在 `main()` 中启动 systray goroutine
- [ ] 添加 `OnBeforeClose` 钩子配置

### 步骤 3: 扩展 App 结构 (30 分钟)

- [ ] 在 `app.go` 中添加托盘相关字段
- [ ] 实现 `runSystray()` 方法
- [ ] 实现 `onSystrayReady()` 回调(创建菜单)
- [ ] 实现 `onSystrayExit()` 回调
- [ ] 实现 `onBeforeClose()` 拦截关闭
- [ ] 实现 `showWindow()` / `hideWindow()` 方法
- [ ] 实现 `quitApplication()` 安全退出

### 步骤 4: 前端集成 (可选,15 分钟)

- [ ] 在 `App.vue` 中添加窗口事件监听
- [ ] 添加首次使用提示(可选)

### 步骤 5: 测试验证 (30 分钟)

- [ ] 功能测试(参照测试场景表格)
- [ ] Windows 平台测试
- [ ] 长时间运行测试
- [ ] 性能基准测试

### 步骤 6: 文档更新 (15 分钟)

- [ ] 更新 `CLAUDE.md` 添加系统托盘说明
- [ ] 添加用户使用说明

**总工时估算:** 约 2 小时

## 风险评估

### 风险等级

⚠️ **低风险**

### 已知风险

1. **初始化时序问题**
   - 可能需要调整 systray 初始化延迟时间
   - 参考: to_icalendar 项目遇到过类似问题(commit d3387fd)
   - 缓解措施:使用 500ms 延迟初始化

2. **平台差异**
   - 不同平台的托盘图标格式要求不同
   - 缓解措施:在 `getTrayIcon()` 中根据平台选择合适格式

3. **现有功能影响**
   - `OnBeforeClose` 可能影响现有的窗口关闭逻辑
   - 缓解措施:充分测试所有关闭场景

### 回滚方案

如果出现问题,可以快速回滚:

1. 移除 `main.go` 中的 `OnBeforeClose` 钩子配置
2. 注释掉 `runSystray()` 的调用
3. 应用恢复原有行为(正常关闭退出)

## 参考资料

### 相关项目

- **to_icalendar** (参考实现)
  - Commit: `93df45d` - 系统托盘功能实现
  - Commit: `d3387fd` - 添加 `runtime.LockOSThread` 防止崩溃
  - 文件: `cmd/to_icalendar_tray/main.go`, `app.go`

### 官方文档

- [Wails v2 Documentation](https://wails.io/docs/v2.0/introduction)
- [Wails v3 System Tray](https://v3alpha.wails.io/features/menus/systray/)
- [Wails v2 to v3 Migration](https://v3alpha.wails.io/migration/v2-to-v3/)

### 第三方库

- [github.com/getlantern/systray](https://github.com/getlantern/systray)
  - 跨平台系统托盘库
  - 支持 Windows, macOS, Linux
  - API 简洁,文档完善

## 总结

本设计方案提供了一个稳定、可靠、低风险的系统托盘功能实现方案。通过使用成熟的 `github.com/getlantern/systray` 库,并结合 Wails v2 的生命周期钩子,可以在不影响现有功能的前提下,为应用添加后台运行能力。

### 核心特性

✅ 点击关闭按钮隐藏到托盘
✅ 托盘菜单提供"显示窗口"和"退出应用"
✅ 启动时显示主窗口
✅ 复用现有应用图标
✅ 安全退出机制(防止竞态)
✅ 跨平台支持(Windows 优先)

### 关键优势

- **稳定性**: 基于成熟的第三方库,已有成功案例
- **低风险**: 不需要升级框架,改动范围可控
- **易维护**: 代码量少,逻辑清晰
- **用户友好**: 符合 Windows 用户习惯的托盘交互

---

**设计者:** Claude Code
**审批状态:** 待实施
**文档版本:** 1.0
