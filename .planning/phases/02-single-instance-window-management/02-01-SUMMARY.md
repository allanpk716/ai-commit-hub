---
phase: 02-single-instance-window-management
plan: 01
subsystem: window-management
tags: wails, single-instance, window-activation

# Dependency graph
requires:
  - phase: 01-ci-cd-pipeline
    provides: Wails 应用基础架构和构建流程
provides:
  - Wails SingleInstanceLock 机制防止多实例运行
  - 窗口激活回调实现
  - 静默窗口激活到前台
affects: 02-02-window-state-persistence, 02-03-window-state-restore

# Tech tracking
tech-stack:
  added: []
  patterns:
    - SingleInstanceLock 使用唯一 ID 标识应用实例
    - 静默窗口激活机制（无通知/对话框）
    - 窗口状态标志同步（windowVisible 与实际状态一致）

key-files:
  created: []
  modified:
    - main.go: 添加 SingleInstanceLock 配置
    - app.go: 实现 onSecondInstanceLaunch 回调和窗口状态同步

key-decisions:
  - "使用固定 UUID 'e3984e08-28dc-4e3d-b70a-45e961589cdc' 作为 SingleInstanceLock 唯一标识"
  - "静默激活窗口策略：恢复最小化 + 显示到前台，不显示任何通知"
  - "使用 showWindow() 方法替代直接调用 runtime.WindowShow() 以保持状态一致"

patterns-established:
  - "Pattern 1: SingleInstanceLock 防止多实例冲突"
  - "Pattern 2: defer recover 捕获窗口激活 panic"
  - "Pattern 3: 窗口状态标志同步必须通过封装方法"

# Metrics
duration: 22min
completed: 2026-02-06
---

# Phase 2 Plan 1: 单实例锁定机制 Summary

**使用 Wails SingleInstanceLock 实现单实例应用，检测到第二个实例启动时自动激活现有窗口到前台**

## Performance

- **Duration:** 22 min
- **Started:** 2025-02-06T03:52:00Z
- **Completed:** 2025-02-06T04:14:00Z
- **Tasks:** 2
- **Files modified:** 2 (main.go, app.go)

## Accomplishments

- 配置 Wails SingleInstanceLock 使用固定 UUID 防止多实例运行
- 实现第二个实例启动时自动激活现有窗口到前台的回调机制
- 确保窗口激活后状态标志正确同步，关闭按钮正常工作

## Task Commits

每个任务都已原子性提交：

1. **Task 1: 配置 SingleInstanceLock 选项** - `9756c40` (feat)
2. **Task 2: 实现 onSecondInstanceLaunch 回调** - `211ba09` (feat)
3. **修复: 添加缺失的 options 包导入** - `a38ea18` (fix)
4. **修复: 单实例激活后关闭按钮失效问题** - `b6ce5d4` (fix)

**Plan metadata:** 待创建 (docs: complete plan)

## Files Created/Modified

### 修改的文件

- `main.go` - 添加 SingleInstanceLock 配置
  - 在 wails.Run 选项中添加 SingleInstanceLock 配置块
  - 使用 UUID 'e3984e08-28dc-4e3d-b70a-45e961589cdc' 作为 UniqueId
  - 配置 OnSecondInstanceLaunch 回调为 app.onSecondInstanceLaunch

- `app.go` - 实现单实例窗口激活逻辑
  - 添加 onSecondInstanceLaunch 方法处理第二个实例启动事件
  - 调用 runtime.WindowUnminimise 恢复最小化窗口
  - 调用 a.showWindow() 显示并激活窗口到前台（确保状态同步）
  - 使用 defer recover 捕获 panic 并记录错误日志
  - 添加 github.com/wailsapp/wails/v2/pkg/options 导入

## Decisions Made

1. **使用固定 UUID 作为 SingleInstanceLock 唯一标识**
   - 原因：简单直接，无需动态生成，避免导入额外依赖
   - 格式：'e3984e08-28dc-4e3d-b70a-45e961589cdc'
   - 影响：确保同一应用的所有实例使用相同的唯一 ID

2. **静默激活窗口策略**
   - 原因：用户期望双击应用图标时看到现有窗口，而非错误提示
   - 实现：恢复最小化 + 显示到前台，不显示任何通知或对话框
   - 影响：提供流畅的用户体验，符合 CONTEXT.md 中的决策

3. **使用 showWindow() 方法替代直接调用 runtime.WindowShow()**
   - 原因：确保窗口状态标志（windowVisible）与实际窗口状态同步
   - 实现：onSecondInstanceLaunch 调用 a.showWindow() 而非 runtime.WindowShow()
   - 影响：避免关闭按钮失效的问题（详见 Deviations 部分）

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 添加缺失的 options 包导入**

- **Found during:** Task 2 (实现 onSecondInstanceLaunch 回调)
- **Issue:** 编译失败，报错 `undefined: options`
  - 原因：使用了 options.SecondInstanceData 类型但未导入相应包
  - 错误信息：`./app.go:76:58: undefined: options`
- **Fix:**
  - 在 app.go 中添加 `github.com/wailsapp/wails/v2/pkg/options` 导入
  - 验证编译成功
- **Files modified:** app.go
- **Verification:** `wails build` 编译成功，无错误
- **Committed in:** a38ea18 (独立 fix 提交)

---

**2. [Rule 1 - Bug] 修复单实例激活后关闭按钮失效问题**

- **Found during:** 验证测试（用户反馈）
- **Issue:** 通过第二个实例激活窗口后，点击关闭按钮没反应
  - 原因分析：
    - onSecondInstanceLaunch 直接调用 runtime.WindowShow()
    - 但没有更新 windowVisible 标志（初始值为 false）
    - hideWindow() 检查到 windowVisible == false，直接返回，不执行隐藏
  - 根本原因：窗口状态标志与实际状态不一致
- **Fix:**
  - 使用 a.showWindow() 替代直接调用 runtime.WindowShow()
  - showWindow() 方法会正确设置 windowVisible = true
  - 确保窗口状态标志与实际状态同步
- **Files modified:** app.go
  - 修改 onSecondInstanceLaunch 方法实现
  - 从 `runtime.WindowShow(a.ctx)` 改为 `a.showWindow()`
- **Verification:**
  - 启动应用第一个实例
  - 再次双击 exe 启动第二个实例
  - 现有窗口被激活到前台
  - 点击关闭按钮，应用正常退出
- **Committed in:** b6ce5d4 (独立 fix 提交)

代码对比：

```go
// 修复前（直接调用 runtime API）
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
    defer func() {
        if r := recover(); r != nil {
            logger.Errorf("激活窗口失败: %v", r)
        }
    }()

    runtime.WindowUnminimise(a.ctx)
    runtime.WindowShow(a.ctx)  // ❌ 没有更新 windowVisible 标志
}

// 修复后（使用封装方法）
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
    defer func() {
        if r := recover(); r != nil {
            logger.Errorf("激活窗口失败: %v", r)
        }
    }()

    runtime.WindowUnminimise(a.ctx)
    a.showWindow()  // ✅ 正确设置 windowVisible = true
}
```

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)

**Impact on plan:** 两个自动修复都是必要的，确保了功能正确性。导入缺失是开发过程正常问题，关闭按钮失效是状态同步设计的遗漏。没有引入范围蔓延。

## Issues Encountered

### 编译错误：undefined: options

**问题：** 实现 onSecondInstanceLaunch 回调时，编译报错 `undefined: options`

**解决：** 在 app.go 中添加 `github.com/wailsapp/wails/v2/pkg/options` 导入

**教训：** 使用新的类型时，确保导入相应的包

### 功能缺陷：关闭按钮失效

**问题：** 单实例激活窗口后，点击关闭按钮无反应

**调试过程：**
1. 检查 hideWindow() 方法，发现第一行就检查 `if !a.windowVisible { return }`
2. 检查 windowVisible 初始化值，默认为 false
3. 检查窗口显示逻辑，发现 onSecondInstanceLaunch 直接调用 runtime.WindowShow()，没有设置 windowVisible
4. 对比 showWindow() 方法，发现它会设置 windowVisible = true

**解决：** 将 runtime.WindowShow() 替换为 a.showWindow()

**教训：** 窗口状态标志必须与实际窗口状态保持同步，所有窗口操作都应通过封装方法进行

## User Setup Required

无 - 不需要外部服务配置

## Verification Results

测试执行步骤：

1. ✅ 构建: `wails build`
2. ✅ 启动应用第一个实例
3. ✅ 再次双击 exe 启动第二个实例
4. ✅ 观察现有窗口被激活到前台（而非打开新窗口）
5. ✅ 确认没有任何错误对话框或通知显示
6. ✅ 点击关闭按钮，应用正常退出

**测试结论：** 所有功能符合预期，单实例锁定机制正常工作

## Next Phase Readiness

### 已完成

- ✅ SingleInstanceLock 配置完成
- ✅ 窗口激活回调实现
- ✅ 窗口状态同步机制修复
- ✅ 所有测试通过

### 下一阶段准备

**计划 02-02: 创建窗口状态数据层** 需要以下准备：

- ✅ 单实例锁定机制已就绪，不会影响状态持久化
- ✅ 窗口状态标志（windowVisible）已正确同步
- ✅ showWindow/hideWindow 方法状态管理逻辑清晰

**建议：** 下一计划应关注窗口位置、大小等状态的持久化，与当前的状态标志机制保持一致

### 阻塞问题

无

---
*Phase: 02-single-instance-window-management*
*Plan: 01*
*Completed: 2025-02-06*
