# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-06)

**Core value:** 简化 Git 工作流 - 自动生成高质量 commit 消息
**Current focus:** Phase 2 - Single Instance & Window Management

## Current Position

Phase: 4 of 5 (Auto Update System)
Plan: 1 of 4 in current phase
Status: In progress
Last activity: 2026-02-07 — Completed 04-01 (版本检测和 UI 集成)

Progress: [██████████░] 61%

## Performance Metrics

**Velocity:**
- Total plans completed: 9
- Average duration: 3 min
- Total execution time: 0.5 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-ci-cd-pipeline | 3 | 6 min | 2 min |
| 02-single-instance-window-management | 3 | 11 min | 4 min |
| 03-system-tray-fixes | 2 | 7 min | 4 min |
| 04-auto-update-system | 1 | 8 min | 8 min |

**Recent Trend:**
- Last 5 plans: 4 min, 2 min, 1 min, 4 min, 8 min
- Trend: Stable (3.8 min per plan)

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Initial]: 单实例策略 - 激活现有窗口到前台而非静默退出
- [Initial]: 自动更新方式 - 下载新版 exe 后自动替换程序
- [Initial]: 功能边界 - 只集成常用操作，复杂操作外部处理
- [Initial]: 实施顺序 - CI/CD 优先，建立发布流程后再实现其他功能
- [01-01]: Windows amd64 only - 排除 32-bit (386) 构建因 WebView2 崩溃问题
- [01-01]: NODE_OPTIONS=--max-old-space-size=4096 防止前端构建 OOM
- [01-01]: 版本注入方式 - 通过 ldflags 注入到 pkg/version 包
- [01-01]: CGO_ENABLED=1 - Windows SQLite 驱动必需
- [01-02]: 打包格式 - 使用 7z 创建 ZIP 归档 (Windows 兼容性好)
- [01-02]: 包内容 - exe + README.md + config.yaml (用户友好)
- [01-02]: 双重校验和 - SHA256 (安全) + MD5 (兼容性)
- [01-02]: Artifact 保留期 - 30 天 (手动下载测试)
- [01-02]: 命名规范 - ai-commit-hub-windows-amd64-v{version}.zip (支持平台检测)
- [01-03]: Job 分离 - Build (Windows) + Release (Ubuntu) 降低成本
- [01-03]: Job outputs - VERSION 和 PRERELEASE 跨 Job 共享
- [01-03]: 自动发布 - 使用 softprops/action-gh-release 创建 Release
- [01-03]: 自动 Release Notes - GitHub 自动生成 changelog
- [01-03]: Pre-release 检测 - 正则匹配 alpha/beta/rc/pre 标签
- [02-01]: SingleInstanceLock UUID - 使用固定 UUID 'e3984e08-28dc-4e3d-b70a-45e961589cdc'
- [02-01]: 窗口激活策略 - 静默激活（恢复最小化 + 显示到前台，无通知）
- [02-01]: 窗口状态同步 - 必须使用封装方法（showWindow/hideWindow）而非直接调用 runtime API
- [02-03]: 生命周期集成 - 在 startup 恢复窗口状态，在 onBeforeClose 保存窗口状态
- [02-03]: 位置验证 - 使用边界检查防止窗口"丢失"在屏幕外（minWidth 400, minHeight 300, maxCoord 10000）
- [02-03]: 启动时序优化 - main.go 预读取数据库设置窗口大小，startup() 中恢复位置和最大化状态，避免闪烁
- [02-03]: Upsert 策略 - 使用 clause.OnConflict 处理窗口状态保存的 UNIQUE constraint 错误
- [04-01]: 版本比较库 - 使用 golang.org/x/mod/semver 实现标准语义化版本比较
- [04-01]: 版本检测端点 - 使用 /releases 而非 /releases/latest 支持预发布版本
- [04-01]: 缓存机制 - 24 小时 TTL + 速率限制时返回缓存
- [04-01]: 后台检查 - 每 24 小时自动检查更新

### Pending Todos

**Phase 2 已完成并验证:**
- ✅ 窗口状态保存和恢复功能已实现并验证
- ✅ 最大化状态恢复正常工作
- ✅ 窗口位置恢复无闪烁
- ✅ UNIQUE constraint 错误已修复

### Blockers/Concerns

**无阻塞问题**

**Phase 4-01 完成总结:**
- 版本比较支持预发布版本完成 (04-01)
- UpdateService 重构完成，使用 /releases 端点 (04-01)
- 24 小时后台自动检查已实现 (04-01)
- 托盘菜单"检查更新"已连接真实 API (04-01)
- UpdateInfo 模型增强，包含预发布标识 (04-01)

**可以继续 Phase 4-02 (Download and Install)**

**Phase 2 完成总结:**
- 单例锁定和窗口激活功能完成 (02-01)
- 窗口状态持久化数据层完成 (02-02)
- 窗口状态持久化应用层集成完成 (02-03)
- 所有关键问题已修复并验证：
  - 窗口启动时序问题（闪烁）- 通过 main.go 预读取解决
  - UNIQUE constraint 错误 - 通过 Upsert 策略解决
  - 最大化状态恢复问题 - 通过延迟 100ms 解决

**可以继续 Phase 3 (System Tray Fixes)**

## Session Continuity

Last session: 2026-02-07
Stopped at: Completed 04-01 - Version detection and UI integration
Resume file: None

## Phase 2 完成总结

**完成日期**: 2026-02-06

**已完成计划**:
- 02-01: 单实例锁定和窗口激活
- 02-02: 窗口状态数据层
- 02-03: 窗口状态保存和恢复

**验证结果**: Passed (4/4 must-haves verified)

**关键决策**:
- SingleInstanceLock UUID - 使用固定 UUID 'e3984e08-28dc-4e3d-b70a-45e961589cdc'
- 窗口激活策略 - 静默激活（恢复最小化 + 显示到前台，无通知）
- 窗口状态同步 - 必须使用封装方法（showWindow/hideWindow）而非直接调用 runtime API
- 生命周期集成 - 在 startup 恢复窗口状态，在 onBeforeClose 保存窗口状态
- 位置验证 - 使用边界检查防止窗口"丢失"在屏幕外（minWidth 400, minHeight 300, maxCoord 10000）
- 启动时序优化 - main.go 预读取数据库设置窗口大小，startup() 中恢复位置和最大化状态，避免闪烁
- Upsert 策略 - 使用 clause.OnConflict 处理窗口状态保存的 UNIQUE constraint 错误

**下一步**: Phase 4 - Auto Update System

## Phase 3 完成总结

**完成日期**: 2026-02-06

**已完成计划**:
- 03-01: 升级 systray 库并实现双击功能
- 03-02: 修复托盘竞态条件和优化退出行为

**验证结果**: Passed (4/4 must-haves verified)

**关键决策**:
- systray 库升级 - 使用 lutischan-ferenc/systray v1.3.0 替代 getlantern/systray v1.2.2
- API 迁移 - 从 Channel API 迁移到 Callback API，简化代码
- 双击实现 - 使用 SetOnDClick() 实现托盘图标双击恢复窗口
- 竞态条件防护 - 使用 sync.Once 和 atomic.Bool 防止重复退出
- 退出/最小化分离 - 通过 quitting atomic.Bool 区分两种行为
- 菜单结构 - 实现三项菜单（显示窗口、检查更新 stub、退出应用）

## Phase 4-01 完成总结

**完成日期**: 2026-02-07

**已完成计划**:
- 04-01: 版本检测和 UI 集成

**关键决策**:
- 版本比较库 - 使用 golang.org/x/mod/semver v0.32.0 实现标准语义化版本比较
- 版本检测端点 - 使用 /releases 而非 /releases/latest 支持预发布版本
- 缓存机制 - 24 小时 TTL + 速率限制时返回缓存提高可靠性
- 后台检查 - 使用 time.NewTicker(24*time.Hour) 实现定时检查
- 精确平台匹配 - 从 strings.Contains(asset.Name, "windows") 改为精确匹配 windows-amd64

**下一步**: Phase 4-02 - Download and Install

