# AI Commit Hub - Bug Fix & Enhancement

## What This Is

基于 Wails（Go + Vue3）的桌面应用，为多个 Git 项目自动生成 AI 驱动的 commit 消息。当前版本正在修复三个关键问题并添加自动更新功能，以提升 Windows 平台用户体验。

## Core Value

用户能够通过 AI 辅助快速生成规范的 Git commit 消息，同时享受流畅的桌面应用体验（托盘交互、单实例、自动更新）。

## Requirements

### Validated

✓ **AI Commit 生成** - 支持多个 AI Provider（OpenAI、Anthropic、Google、Ollama）生成 commit 消息 — existing
✓ **Git 操作集成** - Git status、diff、commit、push 操作 — existing
✓ **多项目管理** - 管理多个 Git 项目的配置 — existing
✓ **配置持久化** - YAML 配置文件 + SQLite 数据库 — existing
✓ **系统托盘集成** - Windows 系统托盘图标和右键菜单 — existing
✓ **Vue3 用户界面** - 响应式前端界面，支持 commit 生成、暂存管理、设置配置 — existing
✓ **错误处理和日志** - 统一的错误处理和结构化日志系统 — existing

### Active

- [ ] **SYSTRAY-01**: Windows 托盘图标双击时显示主窗口（从最小化/隐藏状态恢复）
- [ ] **SINGLE-01**: 单实例运行 - 检测到已有实例时激活现有窗口，不启动新实例
- [ ] **UPDATE-01**: GitHub Actions 配置 - 在打 tag 时自动构建并发布 Windows 版本
- [ ] **UPDATE-02**: 启动时检查更新 - 程序启动时自动检查 GitHub Releases 是否有新版本
- [ ] **UPDATE-03**: 手动检查更新按钮 - 在设置界面提供"检查更新"按钮
- [ ] **UPDATE-04**: 主界面版本显示 - 在主界面显示当前版本号和更新状态提示
- [ ] **UPDATE-05**: 下载更新提示 - 下载完成后提示用户确认安装
- [ ] **UPDATE-06**: 自动更新安装 - 用户确认后自动安装并重启程序

### Out of Scope

- **跨平台构建** - GitHub Actions 只构建 Windows 版本（macOS/Linux 暂不需要）
- **托盘图标状态显示** - 不在托盘图标上显示未读提交数量或其他状态
- **后台定期检查更新** - 只在启动时检查和手动检查，不进行后台定期检查
- **托盘菜单退出功能** - 不添加"退出程序"到托盘右键菜单（当前已有其他退出方式）
- **关闭最小化到托盘** - 不改变当前窗口关闭行为

## Context

**技术栈：**
- 后端：Go 1.24 + Wails v2.11 + SQLite + GORM
- 前端：Vue 3.5 + TypeScript 5.9 + Pinia 3.0 + Vite 7.2
- 系统集成：github.com/getlantern/systray v1.2.2
- 日志：github.com/WQGroup/logger v0.0.16

**现有架构：**
- 事件驱动架构，使用 Wails Events 进行实时更新
- Provider 抽象层支持多个 AI 服务
- Repository 模式处理数据持久化
- Service 层封装业务逻辑

**已知问题：**
- systray 状态管理复杂（app.go:69-70, 294-305），存在潜在的竞态条件
- app.go 文件过大（1942 行），需要模块化
- 已有 update_service.go 但可能需要扩展功能

**开发平台：**
- 主要开发平台：Windows
- 配置目录：`C:\Users\<username>\.ai-commit-hub\`

## Constraints

- **平台**: 仅 Windows 平台（GitHub Actions 只构建 Windows 版本）
- **版本号**: 使用语义化版本号（v1.0.0, v1.0.1, v2.0.0）
- **时间**: 正常进度，质量优先
- **技术栈**: 必须使用现有的 Wails + Go + Vue3 技术栈
- **兼容性**: 保持现有功能不受影响

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 优先级顺序：托盘双击 → 单实例 → 自动更新 | 托盘功能直接影响用户体验，单实例防止资源浪费，自动更新是增强功能 | — Pending |
| GitHub Actions 只构建 Windows 版本 | 用户明确表示只需要 Windows 版本，减少构建复杂度和时间 | — Pending |
| 自动更新采用下载后确认模式 | 平衡自动化和用户控制权，避免意外更新导致工作中断 | — Pending |
| 版本信息显示在主界面而非设置页 | 用户更容易看到更新提示，提高更新率 | — Pending |

---
*Last updated: 2025-02-05 after initialization*
