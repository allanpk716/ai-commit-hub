# AI Commit Hub

## What This Is

AI Commit Hub 是一个基于 Wails (Go + Vue3) 的轻量级 Git 辅助桌面应用，用于为多个 Git 项目自动生成 AI 驱动的 commit 消息。专注简化 Git 常用操作，复杂操作仍使用 TortoiseGit 等专业工具。

## Core Value

简化 Git 工作流 - 自动生成高质量 commit 消息，让开发者专注于代码而非提交说明编写。

## Requirements

### Validated

<!-- 已实现且验证的核心功能 -->

- ✓ **AI commit 生成** — v1.0
  - 基于已暂存文件自动生成 commit 消息（支持多个 AI provider）
  - AI Provider 抽象层（OpenAI、Anthropic、Google Gemini、Ollama）

- ✓ **多项目支持** — v1.0
  - 管理多个 Git 仓库，独立配置
  - SQLite 持久化

- ✓ **通知功能** — v1.0
  - Pushover + Windows 系统通知集成
  - cc-pushover-hook 扩展支持

- ✓ **基础 Git 操作** — v1.0
  - commit、push、pull、status 等常用操作
  - pkg/git 封装

- ✓ **系统托盘集成** — v1.0
  - 最小化到托盘，后台运行
  - 双击恢复窗口
  - 右键菜单（显示/隐藏、检查更新、退出）

- ✓ **CI/CD Pipeline** — v1.0
  - GitHub Actions 自动化构建
  - 自动发布到 GitHub Releases

- ✓ **单实例锁定** — v1.0
  - Wails SingleInstanceLock 机制
  - 多实例检测和窗口激活

- ✓ **窗口状态管理** — v1.0
  - 窗口位置和大小保存/恢复
  - 最大化状态持久化

- ✓ **自动更新功能** — v1.0
  - GitHub Releases 版本检测
  - 后台下载更新文件（断点续传）
  - 外部更新器程序（嵌入主程序）
  - 文件备份和回滚机制
  - 更新完成后自动重启

- ✓ **代码质量** — v1.0
  - 0 编译错误
  - 0 测试失败

### Active

<!-- 当前需要修复/实现的功能 -->

- [ ] **日志格式修复** — 使用 `github.com/WQGroup/logger` 的 `withField` 格式器
- [ ] **日志输出路径修复** — 输出到程序根目录的 logs 文件夹
- [ ] **自动更新检测失败修复** — 调查并修复 GitHub Releases 版本检测问题

### Out of Scope

<!-- 明确不在范围内的功能 -->

- **复杂 Git 操作** — merge、rebase、cherry-pick 等使用 TortoiseGit
- **完整 Git GUI** — 不替代 TortoiseGit、GitKraken 等专业工具
- **跨平台同步** — 配置文件不跨设备同步
- **团队协作** — 单人开发者工具，无协作功能
- **移动端支持** — Windows 主要平台（macOS/Linux 次要）

## Current State

**Shipped Version:** v1.0.0 (2026-02-07)

**Codebase State:**
- LOC: 7,115 lines (Go + Vue + TypeScript)
- Files: 38 files created/modified in v1.0
- Build Status: ✅ All compilation errors fixed
- Test Status: ✅ All tests passing (0 failures)

**Infrastructure:**
- CI/CD: ✅ Fully automated (GitHub Actions)
- Release: ✅ v1.0.0 tagged and pushed
- Documentation: ✅ Complete (MILESTONE-v1-COMPLETE.md, RELEASE_NOTES.md)

**Technical Architecture:**
- 分层架构：App → Service → Repository → Database
- 事件驱动：Wails Events 实现前后端通信
- 状态管理：Pinia stores + StatusCache (TTL 30s)
- 外部更新器：独立进程更新模式（Windows 文件锁定解决方案）

**Quality Metrics:**
- 编译通过率：100% ✅
- 测试通过率：100% ✅
- 代码覆盖率：核心功能全覆盖

## Current Milestone: v1.0.1 日志系统修复

**Goal:** 修复日志格式和输出路径问题，确保日志正确输出到程序根目录 logs 文件夹

**Target features:**
- 修复日志格式为标准格式：`2025-12-18 18:32:07.379 - [INFO]: 消息内容`
- 配置日志输出到程序根目录的 logs 文件夹
- 修复自动更新检测失败问题

## Context

**技术栈：**
- 后端：Go 1.24+ + Wails v2.11.0
- 前端：Vue 3.5 + TypeScript + Vite + Pinia
- 数据库：SQLite + GORM
- 日志：github.com/WQGroup/logger
- 版本比较：golang.org/x/mod/semver
- 系统托盘：lutischan-ferenc/systray v1.3.0

**架构特点：**
- Wails Events 实现流式输出
- Provider 模式抽象 AI 服务
- Repository 模式数据持久化
- StatusCache 状态缓存优化
- 外部更新器模式（Windows 文件锁定解决方案）

**已知问题（已解决）：**
- ✅ 系统托盘双击失效 - 升级 systray 库并实现双击支持
- ✅ 无单实例锁定机制 - 实现 Wails SingleInstanceLock
- ✅ 自动更新功能界面入口缺失 - 完整的 UI 集成
- ✅ GitHub Actions workflow 配置错误 - 修复构建流程
- ✅ 编译错误（10 处 logger 调用） - 全部修复
- ✅ 测试失败（2 个测试） - 全部修复

## Constraints

- **平台**: Windows 主要平台（macOS/Linux 次要）
- **性能**: 快速启动，低资源占用
- **部署**: 单文件 exe，无需安装
- **更新**: 自动检测并替换更新
- **实例**: 单实例运行
- **兼容性**: 配置文件向后兼容

## Key Decisions

<!-- 项目过程中的重要决策 -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 单实例策略 | 多实例导致配置冲突和资源浪费 | ✓ 激活现有窗口到前台而非静默退出 |
| 自动更新方式 | 用户期望无感知更新 | ✓ 下载新版 exe 后自动替换程序 |
| 功能边界 | 避免与 TortoiseGit 竞争 | ✓ 只集成常用操作，复杂操作外部处理 |
| SingleInstanceLock UUID | 固定 ID 确保唯一性 | ✓ 使用 'e3984e08-28dc-4e3d-b70a-45e961589cdc' |
| 窗口激活策略 | 无打扰用户体验 | ✓ 静默激活（恢复最小化 + 显示到前台，无通知） |
| 外部更新器模式 | Windows 无法替换正在运行的 exe | ✓ 独立进程等待主程序退出后替换文件 |
| 更新器嵌入策略 | 简化部署，单文件发布 | ✓ 使用 embed.FS 嵌入 updater.exe (3.2MB) |
| 版本比较库 | 标准语义化版本比较 | ✓ 使用 golang.org/x/mod/semver |
| systray 库升级 | 旧库不支持双击 | ✓ lutischan-ferenc/systray v1.3.0 |
| 文件回滚机制 | 确保系统稳定 | ✓ 更新失败时自动回滚到备份版本 |

---
*Last updated: 2026-02-08 after starting v1.0.1 milestone*
