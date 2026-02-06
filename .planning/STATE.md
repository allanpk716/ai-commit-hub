# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-06)

**Core value:** 简化 Git 工作流 - 自动生成高质量 commit 消息
**Current focus:** Phase 2 - Single Instance & Window Management

## Current Position

Phase: 2 of 5 (Single Instance & Window Management)
Plan: 2 of 3 in current phase
Status: In progress
Last activity: 2026-02-06 — Completed 02-01-SUMMARY.md (单实例锁定机制)

Progress: [██████░░░] 40%

## Performance Metrics

**Velocity:**
- Total plans completed: 6
- Average duration: 4 min
- Total execution time: 0.4 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-ci-cd-pipeline | 3 | 3 | 2 min |
| 02-single-instance-window-management | 3 | 3 | 4 min |

**Recent Trend:**
- Last 5 plans: 4 min, 2 min, 1 min, 2 min, 2 min
- Trend: Stable (2.2 min per plan)

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

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-06 04:14 UTC
Stopped at: Completed 02-01-SUMMARY.md (Single instance lock mechanism completed)
Resume file: None
