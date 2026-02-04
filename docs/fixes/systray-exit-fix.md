# 系统托盘退出修复

**日期**: 2026-02-02
**问题**: 托盘菜单"退出应用"无效
**解决方案**: 修复退出逻辑，统一使用 PNG 图标

---

## 问题描述

### 1. 托盘菜单"退出应用"无效

**现象**:
- 右键点击托盘图标 → 点击"退出应用" → 窗口显示，但应用不退出
- 窗口关闭按钮正常，应用最小化到托盘

**根本原因**:
```go
// 旧的退出逻辑
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        runtime.Quit(a.ctx)  // ❌ 触发 onBeforeClose
    })
}

func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    a.hideWindow()
    return true  // ❌ 拦截所有关闭事件
}
```

退出流程：
1. `quitApplication()` → `runtime.Quit()`
2. 触发 `onBeforeClose()` → 拦截关闭 → `hideWindow()`
3. 窗口显示后隐藏，应用未退出

### 2. ICO 图标在 Windows 上显示异常

**现象**:
- Windows 托盘图标显示模糊/缩放效果差
- ICO 格式在高 DPI 显示器上效果不佳

**根本原因**:
```go
// 旧代码：Windows 使用 ICO
func (a *App) getTrayIcon() []byte {
    if stdruntime.GOOS == "windows" {
        return appIconICO  // ❌ ICO 在缩放上不如 PNG
    }
    return appIconPNG
}
```

Windows Vista+ 系统托盘原生支持 PNG，PNG 在缩放和透明度处理上更好。

---

## 解决方案

### 方案 A（已采用）：修复退出逻辑 + 统一使用 PNG

**实现要点**:

1. **添加退出标志位**：
```go
type App struct {
    // ...
    quitting atomic.Bool  // 应用是否正在退出
}
```

2. **修复退出流程**：
```go
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        a.quitting.Store(true)      // ✅ 设置退出标志
        a.showWindow()               // ✅ 显示窗口（避免"卡住"）
        systray.Quit()               // ✅ 直接退出 systray
    })
}
```

3. **修改关闭拦截逻辑**：
```go
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    if a.quitting.Load() {          // ✅ 检查退出标志
        return false                // ✅ 允许关闭
    }
    a.hideWindow()
    return true
}
```

4. **统一使用 PNG 图标**：
```go
func (a *App) getTrayIcon() []byte {
    return appIconPNG               // ✅ 所有平台统一使用 PNG
}
```

**退出流程**：
1. 点击"退出应用" → 设置 `quitting = true`
2. 显示窗口（如果当前隐藏）
3. 调用 `systray.Quit()` → 触发 `onSystrayExit()`
4. Wails 调用 `shutdown()` → 应用退出
5. 窗口关闭事件 → `onBeforeClose()` 检查 `quitting` → 允许关闭 ✅

---

## 修改文件

| 文件 | 修改内容 |
|------|----------|
| `app.go:70` | 添加 `quitting atomic.Bool` 字段 |
| `app.go:261-267` | 修改 `getTrayIcon()` 统一返回 PNG |
| `app.go:387-400` | 重写 `quitApplication()` 退出逻辑 |
| `app.go:405-414` | 修改 `onBeforeClose()` 检查退出标志 |

---

## 测试验证

### 编译测试
```bash
go build -o build/bin/ai-commit-hub.exe .
```
✅ 编译成功

### 功能测试清单

- [ ] **托盘菜单"退出应用"**
  - 右键托盘图标 → 点击"退出应用"
  - 预期：应用立即退出，窗口正常关闭

- [ ] **窗口关闭按钮**
  - 点击窗口关闭按钮 (X)
  - 预期：应用隐藏到托盘，不退出

- [ ] **托盘图标显示**
  - 观察托盘图标清晰度
  - 预期：图标清晰，无模糊/锯齿

- [ ] **托盘菜单"显示窗口"**
  - 隐藏窗口后 → 右键托盘 → 点击"显示窗口"
  - 预期：窗口正常显示

---

## 技术说明

### 为什么显示窗口再退出？

`quitApplication()` 中调用 `showWindow()` 是为了避免用户看到应用"卡住"：

```go
a.showWindow()    // 如果窗口隐藏，先显示
systray.Quit()    // 然后退出
```

**用户体验**：
- ❌ 不显示：用户点击"退出"后，托盘图标消失，但什么都没发生（几秒后才退出）
- ✅ 显示窗口：用户看到窗口显示，然后关闭，知道应用正在退出

### 为什么 PNG 比 ICO 好？

| 特性 | ICO | PNG |
|------|-----|-----|
| 缩放质量 | 固定尺寸，缩放差 | 矢量友好，缩放好 |
| 透明度 | 1-bit 透明度 | 8-bit alpha 通道 |
| 文件大小 | 大（多尺寸嵌入） | 小（单尺寸） |
| 系统支持 | 旧系统 | Windows Vista+ |

**Windows Vista+ 的 systray 原生支持 PNG**，不需要使用旧格式的 ICO。

---

## 遗留问题

无

---

## 参考

- [Windows systray PNG 支持](https://docs.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shell_notifyiconw)
- [Wails onBeforeClose 文档](https://wails.io/docs/reference/runtime/events)
