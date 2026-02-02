# 系统托盘问题修复方案

**日期:** 2025-02-02
**状态:** 设计完成
**优先级:** 高 (影响可用性)

## 问题描述

### 问题 1: 托盘图标不可见
**现象:** Windows 系统托盘区域看不到 AI Commit Hub 的图标
**影响:** 用户无法在托盘中找到应用
**根因:** 使用了 PNG 格式图标,但 Windows 系统托盘需要 ICO 格式

### 问题 2: 托盘菜单失效无法恢复窗口
**现象:**
1. 关闭窗口到托盘 → 托盘图标存在
2. 从托盘恢复窗口 → 成功
3. 再次关闭窗口 → 托盘图标存在
4. 尝试右键点击托盘图标 → 无菜单弹出
5. 无法恢复窗口,应用"被困住"

**影响:** 严重可用性问题,用户需要强制杀进程
**根因:** `systray.Run()` 在某些情况下退出了,导致托盘功能完全失效

## 解决方案设计

### 修复 1: Systray 健康检查和自动重启

**核心策略:**
在 `showWindow()` 方法中添加 systray 健康检查,如果发现 systray 已停止运行,则自动重新启动。

#### 实现细节

**1. 添加运行状态字段**

在 `App` 结构中添加原子布尔值用于追踪 systray 运行状态:

```go
type App struct {
    // ... 现有字段

    // 系统托盘相关字段
    systrayReady    chan struct{}
    systrayExit     *sync.Once
    windowVisible   bool
    windowMutex     sync.RWMutex

    // 新增: systray 运行状态 (原子操作)
    systrayRunning atomic.Bool
}
```

**2. 修改 runSystray() 设置运行状态**

在 systray 启动时设置 `true`,退出时设置 `false`:

```go
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

**3. 在 showWindow() 中添加健康检查和重启逻辑**

```go
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

    // === 新增: 健康检查和自动重启 ===
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

**设计考虑:**
- 使用 `atomic.Bool` 而非普通 bool,确保线程安全
- 在 goroutine 中重启 systray,避免阻塞窗口显示
- 1 秒延迟确保 systray 初始化完成
- 日志记录帮助调试

**预期行为:**
- 每次显示窗口时自动检查 systray 状态
- 如果 systray 失效,自动重新启动
- 用户无感知的自动修复过程

---

### 修复 2: 转换图标为 ICO 格式

**核心策略:**
将现有的 PNG 图标转换为包含多个尺寸的 ICO 文件,以支持 Windows 系统托盘的正确显示。

#### 实施步骤

**步骤 1: 图标转换**

**方法 A: 在线工具 (推荐)**
1. 访问: https://convertio.co/png-ico
2. 上传 `frontend/src/assets/app-icon.png` (512x512)
3. 选择生成多尺寸 ICO: 16x16, 32x32, 48x48
4. 下载转换后的 `app-icon.ico`

**方法 B: 本地命令行 (需要 ImageMagick)**
```bash
# 如果已安装 ImageMagick
magick convert frontend/src/assets/app-icon.png \
  -define icon:auto-resize=256,128,64,48,32,16 \
  frontend/src/assets/app-icon.ico
```

**步骤 2: 放置 ICO 文件**

将转换后的文件保存到:
```
frontend/src/assets/app-icon.ico
```

**步骤 3: 修改 main.go 嵌入两个图标**

```go
//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/src/assets/app-icon.png
var appIconPNG []byte

//go:embed frontend/src/assets/app-icon.ico
var appIconICO []byte
```

**步骤 4: 添加平台特定图标选择方法**

在 `app.go` 中添加:

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

**步骤 5: 修改 onSystrayReady() 使用平台特定图标**

```go
func (a *App) onSystrayReady() {
    logger.Info("系统托盘初始化成功")

    // 使用平台特定图标
    systray.SetIcon(a.getTrayIcon())
    systray.SetTitle("AI Commit Hub")
    systray.SetTooltip("AI Commit Hub - 点击显示窗口")

    // 创建菜单
    // ... 其余代码不变
}
```

---

## 测试计划

### 测试场景 1: 图标显示验证

**步骤:**
1. 清理旧的构建产物: `wails build -clean`
2. 重新构建应用: `wails build`
3. 启动应用
4. 检查 Windows 系统托盘区域

**预期结果:**
- ✅ 托盘图标清晰可见
- ✅ 图标在不同 DPI 设置下正确显示
- ✅ 图标颜色和形状正确

### 测试场景 2: 托盘菜单持久性测试

**步骤:**
1. 启动应用
2. 执行 10 次循环:
   - 点击关闭按钮 (X) → 窗口隐藏
   - 右键点击托盘图标 → 验证菜单弹出
   - 点击"显示窗口" → 窗口显示
3. 记录每次操作是否成功

**预期结果:**
- ✅ 所有 10 次循环中,托盘菜单都能正常弹出
- ✅ "显示窗口"和"退出应用"菜单项始终可用
- ✅ 窗口可以正常恢复
- ✅ 日志中如有 systray 重启,应记录相应信息

### 测试场景 3: 边界情况测试

**测试 A: 快速连续操作**
- 快速连续点击关闭按钮 5 次
- 验证: 只有最后一次生效,无错误

**测试 B: 窗口最小化后关闭**
- 先点击最小化按钮
- 再点击关闭按钮
- 验证: 窗口正确隐藏到托盘

**测试 C: 长时间运行**
- 应用运行 30 分钟
- 每 5 分钟执行一次关闭→打开操作
- 验证: 托盘菜单始终可用

### 测试场景 4: 退出功能验证

**步骤:**
1. 从托盘菜单选择"退出应用"
2. 验证应用完全退出
3. 验证托盘图标消失
4. 重新启动应用
5. 验证托盘功能正常

**预期结果:**
- ✅ 应用干净退出,无残留进程
- ✅ 托盘图标消失
- ✅ 重启后功能正常

---

## 实施顺序

### 阶段 1: 修复托盘菜单问题 (高优先级)

**预计时间:** 30 分钟

**文件修改:**
1. `app.go` - 添加 `systrayRunning atomic.Bool`
2. `app.go` - 修改 `runSystray()` 设置状态
3. `app.go` - 修改 `showWindow()` 添加健康检查

**测试:** 验证关闭→打开→关闭→打开流程

### 阶段 2: 修复图标显示问题 (中优先级)

**预计时间:** 15 分钟

**准备工作:**
1. 转换 PNG → ICO (在线工具)
2. 保存 `app-icon.ico` 到 assets 目录

**文件修改:**
1. `main.go` - 添加 ICO embed 指令
2. `app.go` - 添加 `getTrayIcon()` 方法
3. `app.go` - 修改 `onSystrayReady()` 调用 `getTrayIcon()`

**测试:** 验证托盘图标可见

### 阶段 3: 完整测试验证

**预计时间:** 20 分钟

执行所有测试场景,确保两个问题都已修复。

---

## 风险评估

### 低风险
- ✅ 添加 atomic.Bool 不影响现有逻辑
- ✅ 图标转换是静态操作,不涉及代码逻辑
- ✅ 平台特定图标选择已有成熟方案

### 中风险
- ⚠️ systray 重启可能有短暂延迟(1秒)
- ⚠️ 需要准备 ICO 文件(外部工具转换)
- ⚠️ 修改 embed 指令需要重新构建

### 缓解措施
- 重启逻辑在后台 goroutine 中执行,不阻塞窗口显示
- 提供在线工具和命令行两种转换方式
- 详细日志记录便于调试

---

## 成功标准

### 必须满足
- ✅ 托盘图标在 Windows 系统托盘中可见
- ✅ 托盘菜单在多次关闭/打开后始终可用
- ✅ 窗口可以从托盘恢复,即使 systray 曾失效
- ✅ 应用可以正常退出

### 期望满足
- ✅ 日志清晰记录 systray 重启事件
- ✅ 不同 DPI 设置下图标显示正常
- ✅ 无性能退化

---

## 回滚方案

如果修复后出现问题:

```bash
# 回滚到修复前的 commit
git log --oneline -5  # 找到修复前的 commit
git reset --hard <commit-sha>

# 或者
git revert HEAD
```

当前 main 分支的最新提交是 `054e3a7`,修复将在新的 commit 中进行。

---

**设计者:** Claude Code
**创建时间:** 2025-02-02
**预计实施时间:** 65 分钟
**难度等级:** 中等
