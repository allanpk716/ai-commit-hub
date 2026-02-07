# Roadmap: AI Commit Hub

## Overview

AI Commit Hub 的 v1 里程碑聚焦于构建稳定可靠的应用基础设施。从 CI/CD 流水线开始，建立自动化发布能力；然后实现单实例锁定和窗口管理，奠定应用基础；接着修复系统托盘交互，确保用户体验；最后实现自动更新系统，让用户能够无缝获取最新版本。整个旅程从构建基础设施开始，以交付完整的桌面应用体验结束。

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3, 4, 5): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: CI/CD Pipeline** - 建立自动化构建和发布流程 ✅ 2026-02-06
- [x] **Phase 2: Single Instance & Window Management** - 实现单实例锁定和窗口状态管理 ✅ 2026-02-06
- [x] **Phase 3: System Tray Fixes** - 修复托盘双击和升级依赖 ✅ 2026-02-06
- [x] **Phase 4: Auto Update System** - 实现完整的自动更新功能 ✅ 2026-02-07
- [x] **Phase 5: Code Quality & Polish** - 修复编译错误和完善代码质量 ✅ 2026-02-07

## Phase Details

### Phase 1: CI/CD Pipeline ✅

**Goal**: 建立自动化构建和发布流程，确保代码能够自动编译、测试并发布到 GitHub Releases

**Completed**: 2026-02-06

**Requirements**: CI-01, CI-02, CI-03, CI-04, CI-05

**Success Criteria** (what must be TRUE):
1. ✓ Push tag to GitHub 时自动触发构建流程
2. ✓ 构建流程生成 Windows 平台可执行文件
3. ✓ 构建产物自动上传到 GitHub Releases
4. ✓ 资源文件命名遵循平台检测规范（ai-commit-hub-windows-amd64-v{version}.zip）

**Plans**: 3 plans complete

Plans:
- [x] 01-01-PLAN.md — 创建 GitHub Actions 基础工作流，配置 Wails 构建和版本注入 ✅
- [x] 01-02-PLAN.md — 实现产物打包（exe + 文档）和校验和生成 ✅
- [x] 01-03-PLAN.md — 配置自动发布到 GitHub Releases，支持预发布版本检测 ✅

### Phase 2: Single Instance & Window Management ✅

**Goal**: 实现单实例锁定机制，防止多实例运行，并支持窗口状态持久化

**Completed**: 2026-02-06

**Depends on**: Nothing

**Requirements**: SI-01, SI-02, SI-03, SI-04

**Success Criteria** (what must be TRUE):
1. ✓ 应用启动时自动检测是否已有实例运行
2. ✓ 检测到多实例时，自动激活现有窗口到前台
3. ✓ 窗口位置和大小在下次启动时自动恢复
4. ✓ 使用 Wails 内置 SingleInstanceLock 机制

**Plans**: 3 plans complete

Plans:
- [x] 02-01-PLAN.md — 实现单实例锁定和窗口激活 ✅
- [x] 02-02-PLAN.md — 创建窗口状态数据层(模型、Repository、迁移) ✅
- [x] 02-03-PLAN.md — 实现窗口状态保存和恢复逻辑 ✅

### Phase 3: System Tray Fixes ✅

**Goal**: 修复系统托盘双击功能，升级依赖库，优化托盘交互体验

**Completed**: 2026-02-06

**Depends on**: Phase 2

**Requirements**: TRAY-01, TRAY-02, TRAY-03, TRAY-04, TRAY-05

**Success Criteria** (what must be TRUE):
1. ✓ 双击托盘图标能够恢复并激活主界面到前台
2. ✓ 右键菜单显示"显示/隐藏"、"检查更新"、"退出"选项
3. ✓ 使用 sync.Once 和 atomic.Bool 防止托盘竞态条件
4. ✓ 区分"最小化到托盘"和"退出应用"行为

**Plans**: 2 plans complete

Plans:
- [x] 03-01-PLAN.md — 升级 systray 库到 lutischan-ferenc/systray v1.3.0 并实现双击支持 ✅
- [x] 03-02-PLAN.md — 修复托盘竞态条件和优化退出行为 ✅

### Phase 4: Auto Update System ✅

**Goal**: 实现完整的自动更新系统，包括版本检测、下载和替换更新

**Completed**: 2026-02-07

**Depends on**: Phase 1, Phase 3

**Requirements**: UPD-01, UPD-02, UPD-03, UPD-04, UPD-05, UPD-06, UPD-07, UPD-08

**Success Criteria** (what must be TRUE):
1. ✓ 应用启动时后台检查 GitHub Releases 最新版本
2. ✓ 主界面提供"检查更新"按钮，设置页面显示版本信息
3. ✓ 后台下载更新文件并通过 Wails Events 流式显示进度
4. ✓ 使用外部更新器程序替换主应用，避免文件锁定
5. ✓ 更新完成后自动重启应用

**Plans**: 4 plans complete

Plans:
- [x] 04-01-PLAN.md — 实现版本检测和 UI 集成（支持预发布版本）✅
- [x] 04-02-PLAN.md — 实现后台下载和进度显示（支持断点续传）✅
- [x] 04-03-PLAN.md — 实现外部更新器程序（嵌入主程序）✅
- [x] 04-04-PLAN.md — 实现更新替换和自动重启 ✅

### Phase 5: Code Quality & Polish ✅

**Goal**: 修复编译错误和测试失败，确保代码质量和可维护性

**Completed**: 2026-02-07

**Depends on**: Phase 4

**Requirements**: Q-01, Q-02, Q-03, Q-04, Q-05

**Success Criteria** (what must be TRUE):
1. ✓ 项目能够成功编译，无编译错误
2. ✓ 所有测试通过，无重复函数和类型错误
3. ✓ app.go:969 logger.Errorf 使用正确的格式字符串
4. ✓ error_service_test.go 无重复函数声明

**Plans**: 2 plans complete

Plans:
- [x] 05-01: 修复编译错误 ✅
- [x] 05-02: 修复测试错误 ✅

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. CI/CD Pipeline | 3/3 | ✓ Complete | 2026-02-06 |
| 2. Single Instance & Window Management | 3/3 | ✓ Complete | 2026-02-06 |
| 3. System Tray Fixes | 2/2 | ✓ Complete | 2026-02-06 |
| 4. Auto Update System | 4/4 | ✓ Complete | 2026-02-07 |
| 5. Code Quality & Polish | 2/2 | ✓ Complete | 2026-02-07 |

**Total Progress:** 14/14 plans complete (100%) 🎉
