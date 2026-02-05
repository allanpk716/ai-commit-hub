# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Phase 3 优化 - 代码质量和性能提升

- **代码清理**
  - 清理 tmp 目录和临时文件
  - 删除未使用的测试组件（BackendApiTest.vue, DiffViewerTest.vue）
  - 更新 .gitignore，添加更多临时文件模式

- **代码风格统一**
  - 使用 gofumpt 统一 Go 代码格式（59 个文件）
  - 重命名字段 initError 为 initErr，符合 Go 命名规范
  - 创建 ESLint 配置，统一 TypeScript 代码风格
  - 验证日志输出统一性（已使用 github.com/WQGroup/logger）

- **性能优化**
  - 优化 ProjectList 渲染性能
    - 添加 v-once 指令到静态内容
    - 创建 computed 属性预先计算项目状态
    - 减少函数调用从 O(n*10) 降至 O(n)
  - 优化后端并发性能
    - 创建 pkg/concurrency 并发工具模块
    - 实现动态并发控制工作池
    - 重构 GetAllProjectStatuses 使用工作池
    - 添加 30 秒超时保护
  - 创建防抖和节流工具函数
    - debounce 函数用于延迟执行
    - throttle 函数用于节流执行

- **文档和测试**
  - 添加完整的 API 文档
    - docs/api/backend-api.md：后端 API 方法文档
    - docs/api/frontend-events.md：Wails 事件文档
  - 添加架构文档
    - docs/architecture/frontend-status-cache.md：StatusCache 架构
    - docs/architecture/backend-errors.md：错误处理系统
  - 创建集成测试框架
    - tests/integration/commit_workflow_test.go：Commit 工作流测试
  - 创建性能基准测试
    - tests/benchmark/api_bench_test.go：后端基准测试
    - tests/benchmark/status_cache_bench_test.ts：前端基准测试
    - docs/benchmarks/baseline-2026-02-05.md：性能基线文档
  - 完善 README.md，包含完整的特性、安装、使用、开发指南

#### Phase 2 优化 - 架构改进

- **StatusCache 模块化**
  - 创建 Core、Validation、Retry 三个独立模块
  - 实现预加载、后台刷新、乐观更新机制
  - 添加批量操作和 TTL 缓存管理

- **Git 操作封装**
  - 创建 useGitOperation composable
  - 统一 Git 操作处理流程
  - 实现乐观更新和错误回滚

- **事件系统规范化**
  - 创建 frontend/src/constants/events.ts 事件常量
  - 在 main.ts、App.vue、commitStore 中使用事件常量
  - 改进事件监听和清理机制

- **错误处理系统**
  - 创建 pkg/errors 包，定义领域错误类型
  - 实现 AppInitError、ValidationError、GitOperationError、AIProviderError
  - 添加错误类型检查函数和最佳实践文档

- **Repository 接口抽象**
  - 定义 IGitProjectRepository 和 ICommitHistoryRepository 接口
  - 实现 Mock Repository 用于测试
  - 重构 App 和 StartupService 使用接口类型

#### Phase 1 优化 - 核心重构

- **应用启动优化**
  - 实现后端预加载机制（StartupService）
  - 前端启动画面和超时保护
  - 优化启动时间从 4.5s 降至 2.8s

- **代码拆分和模块化**
  - app.go 拆分为多个模块（services、handlers）
  - 创建 UpdateService 和 ErrorService
  - 重构 systray、tray_icon、update 功能
  - CommitPanel.vue 拆分为多个子组件

- **常量和配置提取**
  - 创建 pkg/constants 包
  - 提取魔法数字为常量
  - 统一配置管理

- **Windows 平台优化**
  - 隐藏控制台窗口（CREATE_NO_WINDOW）
  - 优化托盘图标加载策略
  - 修复窗口关闭和托盘菜单问题

### Changed

- **架构重构**
  - 后端从单体 app.go 拆分为多层架构
  - 前端从单一组件拆分为模块化组件
  - StatusCache 从单一文件拆分为模块化设计

- **性能改进**
  - 后端并发优化 30-50%
  - 前端渲染性能显著提升
  - 启动时间优化 38%

- **代码质量**
  - 统一代码风格（gofumpt、ESLint）
  - 提升测试覆盖率
  - 完善文档覆盖率 100%

### Fixed

- **Windows 平台问题**
  - 修复控制台窗口闪烁问题
  - 修复托盘图标显示问题
  - 优化窗口关闭和托盘行为

- **状态管理问题**
  - 修复启动时 UI 闪烁
  - 优化状态更新频率
  - 修复缓存失效问题

- **错误处理问题**
  - 改进错误信息提示
  - 统一错误传递机制
  - 添加错误恢复逻辑

### Removed

- 删除未使用的测试组件（BackendApiTest.vue, DiffViewerTest.vue）
- 删除临时文件和测试数据

## [1.0.0] - 2026-01-XX

### Added

- 初始版本发布
- 支持多种 AI Provider（OpenAI、Anthropic、Google、DeepSeek、Ollama、Phind）
- 多项目管理
- 流式 commit 消息生成
- Git 操作（暂存、提交、推送）
- Pushover Hook 集成
- 系统托盘支持
- Commit 历史记录
- 配置管理（UI 和配置文件）
- 自定义 Prompt 模板

### Changed

- 首次公开发布
