# AI Commit Hub

## What This Is

AI Commit Hub 是一个基于 Wails (Go + Vue3) 的轻量级 Git 辅助桌面应用，用于为多个 Git 项目自动生成 AI 驱动的 commit 消息。专注简化 Git 常用操作，复杂操作仍使用 TortoiseGit 等专业工具。

## Core Value

简化 Git 工作流 - 自动生成高质量 commit 消息，让开发者专注于代码而非提交说明编写。

## Requirements

### Validated

<!-- 已实现且验证的核心功能 -->

- ✓ **AI commit 生成** - 基于已暂存文件自动生成 commit 消息（支持多个 AI provider）
- ✓ **多项目支持** - 管理多个 Git 仓库，独立配置
- ✓ **通知功能** - Pushover + Windows 系统通知集成
- ✓ **基础 Git 操作** - commit、push、pull、status 等常用操作
- ✓ **AI Provider 抽象** - 支持 OpenAI、Anthropic、Google Gemini、Ollama
- ✓ **系统托盘集成** - 最小化到托盘，后台运行

### Active

<!-- 当前需要修复/实现的功能 -->

- [ ] **托盘双击恢复** - 双击托盘图标应恢复并激活主界面到前台
- [ ] **单实例锁定** - 防止多实例运行，新启动时激活现有窗口
- [ ] **自动更新功能** - 检测 GitHub releases 版本、显示更新提示、自动下载并替换程序
- [ ] **CI/CD 修复** - 修复 GitHub Actions 自动化构建流程
- [ ] **编译错误修复** - 修复 app.go:969 logger.Errorf 使用问题
- [ ] **测试文件修复** - 修复 error_service_test.go 重复函数和类型错误
- [ ] **全面代码检查** - 检查并修复其他潜在问题

### Out of Scope

<!-- 明确不在范围内的功能 -->

- **复杂 Git 操作** - merge、rebase、cherry-pick 等使用 TortoiseGit
- **完整 Git GUI** - 不替代 TortoiseGit、GitKraken 等专业工具
- **跨平台同步** - 配置文件不跨设备同步
- **团队协作** - 单人开发者工具，无协作功能

## Context

**技术栈：**
- 后端：Go 1.24+ + Wails v2.11.0
- 前端：Vue 3.5 + TypeScript + Vite + Pinia
- 数据库：SQLite + GORM
- 日志：github.com/WQGroup/logger

**架构特点：**
- Wails Events 实现流式输出
- Provider 模式抽象 AI 服务
- Repository 模式数据持久化
- 系统托盘集成（systray）

**已知问题：**
- 系统托盘双击失效（之前文档称已实现，但测试不工作）
- 无单实例锁定机制
- 自动更新功能代码存在但界面入口缺失
- GitHub Actions workflow 配置错误
- 部分代码存在编译和测试错误

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
| 单实例策略 | 多实例导致配置冲突和资源浪费 | 激活现有窗口到前台而非静默退出 |
| 自动更新方式 | 用户期望无感知更新 | 下载新版 exe 后自动替换程序 |
| 功能边界 | 避免与 TortoiseGit 竞争 | 只集成常用操作，复杂操作外部处理 |

---
*Last updated: 2026-02-06 after initialization*
