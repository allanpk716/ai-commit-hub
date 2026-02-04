## 现状定位
- 托盘功能由 [app.go](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/app.go) 使用 `github.com/lutischan-ferenc/systray` 启动，入口在 [main.go](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/main.go) 的 `go app.runSystray()`。
- Windows 下图标字节来自 `//go:embed frontend/src/assets/app-icon.ico`（[main.go:L19-L23](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/main.go#L19-L23)），并在 [app.go:L296-L300](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/app.go#L296-L300) 调用 `systray.SetIcon(...)`。

## 高概率根因（按代码现状推断）
- ICO 资源本身不符合托盘期望（缺少 16x16/32x32 层、全透明、或格式异常），Windows 托盘会出现“可点击但图标空白”。
- 线程/消息循环时序问题：当前在 goroutine 中 `systray.Run`，且通过 `Sleep(500ms)` 做“碰运气式”延迟（[app.go:L272-L290](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/app.go#L272-L290)），可能导致托盘初始化不稳定或图标设置未生效。
- 依赖混杂：`go.mod` 同时直接依赖了 `getlantern/systray` 与 `lutischan-ferenc/systray`（[go.mod:L7-L25](file:///c:/WorkSpace/Go2Hell/src/github.com/allanpk716/ai-commit-hub/go.mod#L7-L25)），未来很容易误 import 或产生平台行为差异。

## 计划改动（最小、可回滚、以定位+稳态为目标）
1. 增加“图标字节自检”与清晰日志
   - 在设置图标前检查 `len(appIconICO/appIconPNG)`，为 0 直接报错并回退到默认行为。
   - 在 Windows 下实现一个轻量的 ICO 目录解析（不引入新依赖）：读取 ICO header + entries，打印是否包含 16/24/32/48/256 尺寸。
   - 若缺少 16x16 或 32x32，日志明确提示“需要包含小尺寸层，否则托盘可能空白”。
2. 稳定 systray 线程模型
   - 在 `runSystray()` 开头对 systray goroutine 调用 `runtime.LockOSThread()`（用别名避免与 Wails `runtime` 冲突），减少 Windows 侧消息循环/线程切换导致的随机问题。
   - 移除或替换 `Sleep(500ms)`：优先改为“立即启动”，若仍需延迟则改成更明确的条件（例如等待 Wails `startup` 完成信号），避免时间魔法。
3. 依赖收敛
   - 移除未使用的 `github.com/getlantern/systray` 直依赖，统一只保留当前 import 的 `lutischan-ferenc/systray`。
   - 执行依赖整理（等价于 `go mod tidy`），确保依赖树干净、可预测。

## 验证方式
- 编译校验：按你的规则先跑 `go build -v` 确认可编译。
- 运行校验：启动后确认日志包含“系统托盘初始化成功”以及“ICO 尺寸自检结果”。
- 现象校验：托盘区图标应可见且非空白；若仍空白，依据自检输出可直接判定是否 ICO 层缺失，进而只需替换 `frontend/src/assets/app-icon.ico` 为包含 16/32/48/256 的标准 ICO 即可。

## 预期产出
- 即使问题最终是图标文件本身不对，也能通过自检日志在一次运行内把根因钉死（而不是反复猜）。
- 若问题来自线程/时序，`LockOSThread` + 去掉时间魔法将显著提升稳定性。