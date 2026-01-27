# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

AI Commit Hub 是一个基于 Wails (Go + Vue3) 的桌面应用，用于为多个 Git 项目自动生成 AI 驱动的 commit 消息。

## 开发环境

- **操作系统**: Windows
- **后端**: Go 1.21+ + Wails v2
- **前端**: Vue 3 + TypeScript + Vite + Pinia
- **数据库**: SQLite + GORM

## 常用命令

### 开发命令
```bash
# 启动开发服务器
wails dev

# 生产构建
wails build

# 仅构建 Go 后端（绕过 Wails 绑定问题）
go build -o build/bin/ai-commit-hub.exe .

# 前端开发
cd frontend && npm run dev
```

## 功能特性

### Commit 生成和提交流程

AI Commit Hub 支持完整的 Git 工作流：

1. **生成 Commit 消息**: AI 根据暂存区更改自动生成规范的 commit 消息
2. **提交到本地**: 将生成的消息提交到本地 Git 仓库
3. **推送到远程**: 一键推送更改到远程仓库

### 推送功能说明

**使用方式：**
- 推送按钮只在本地提交成功后可用
- 推送到当前分支的同名远程分支
- 推送成功后按钮自动禁用，避免重复推送
- 切换项目或刷新状态时重置推送按钮

**错误处理：**
- 无远程仓库：显示"未配置远程仓库"错误
- 认证失败：显示"认证失败"错误
- 推送冲突：显示冲突提示，需手动解决
- 网络错误：显示具体的网络错误信息

### 测试命令
```bash
# Go 后端测试
go test ./... -v

# 运行单个包测试
go test ./pkg/repository -v

# 前端测试
cd frontend && npm test
```

### 依赖管理
```bash
# Go 依赖
go mod tidy

# 前端依赖
cd frontend && npm install
```

## 代码架构

### 分层架构

项目采用经典的分层架构，后端 Go 代码与前端 Vue 代码通过 Wails 绑定进行通信：

```
Frontend (Vue3)  ←→  Wails Bindings  ←→  Backend (Go)
     ↓                                            ↓
  Pinia Stores                          App API Methods
     ↓                                            ↓
  Components                            Service Layer
                                               ↓
                                          Repository Layer
                                               ↓
                                          SQLite (GORM)
```

### 后端架构 (Go)

**App 层 (`app.go`)**: Wails 应用的入口，包含所有导出给前端的 API 方法
- 13 个公开方法：项目管理、Commit 生成、Git 操作、历史记录
- 持有 `context.Context` 用于调用 Wails runtime 方法
- 初始化时创建数据库连接和所有 Repository

**Service 层 (`pkg/service/`)**: 业务逻辑层
- `ConfigService`: AI Provider 配置管理
- `CommitService`: Commit 消息生成逻辑，使用 Wails Events 实现流式输出

**Repository 层 (`pkg/repository/`)**: 数据访问层
- `GitProjectRepository`: Git 项目 CRUD 操作
- `CommitHistoryRepository`: Commit 历史记录操作
- `db.go`: 数据库初始化和连接管理

**AI 集成 (`pkg/ai/`, `pkg/provider/`)**: AI Provider 抽象层
- `ai.AIClient` 接口定义（支持流式和非流式）
- Provider Registry 模式：动态注册和获取 AI Provider
- 支持 OpenAI、Anthropic、DeepSeek、Ollama、Google、Phind

**Git 操作 (`pkg/git/`)**: Git 命令封装
- `status.go`: 获取暂存区状态、当前分支
- `git.go`: 执行 git commit、获取 diff

**配置管理 (`pkg/config/`)**: YAML 配置文件解析
- Provider 配置、语言设置、自定义 Prompt 模板

### 前端架构 (Vue3)

**主应用 (`App.vue`)**: 布局容器
- 顶部工具栏（添加项目、设置按钮）
- 左右分栏内容区（项目列表 + Commit 面板）

**组件 (`components/`)**:
- `ProjectList.vue`: 可拖拽排序的项目列表，支持搜索过滤
- `CommitPanel.vue`: Commit 生成面板，显示项目状态、AI 设置、流式输出、历史记录

**状态管理 (`stores/`)**:
- `projectStore.ts`: 项目列表状态、CRUD 操作、排序
- `commitStore.ts`: Commit 生成状态、流式消息监听（Wails Events）

**类型定义 (`types/index.ts`)**: TypeScript 类型与 Go 结构体同步

### Wails 事件流

流式 Commit 生成使用 Wails Events 实现：

```
Frontend          Backend
   |                 |
 generateCommit()   |
   |---------------->|
   |                 Create CommitService
   |                 Start goroutine
   |                 |
   |     commit-delta event (streaming)
   |<----------------|
   |     update UI
   |                 |
   |<----------------|  repeat...
   |                 |
   |     commit-complete event
   |<----------------|
```

## 开发规则

### 通用规则

1. **使用中文回答问题和编写文档**

2. **BAT 脚本规范**: BAT 脚本中不要使用中文（避免编码问题）

3. **临时文件位置**: 临时测试代码、数据统一放在项目根目录的 `tmp/` 文件夹中

4. **脚本修复原则**: 修复脚本时优先在原脚本上修改，非必需不要新建脚本

5. **图片处理**: 使用截图 MCP 或 Agent 能力前，确保图片尺寸小于 1000x1000

6. **计划文件**: 项目计划文件统一放在 `docs/plans/` 目录

7. **日志库使用**: 统一使用 `github.com/WQGroup/logger` 日志库
   - 基本日志级别：`Debug()`、`Info()`、`Warn()`、`Error()`
   - 格式化版本：`Debugf()`、`Infof()`、`Warnf()`、`Errorf()`
   - 支持多种格式器：JSON、文本、结构化日志
   - 支持 YAML 配置文件配置日志行为
   - 支持日志轮转（时间/大小）和自动清理
   - 线程安全，支持并发写入

### Wails 开发规范

1. **调试流程**:
   - 使用 `wails dev` 启动开发服务器
   - 使用浏览器技能（dev-browser）进行前后端通信和交互测试
   - 修改 Go 代码后需要重启 Wails
   - 修改前端代码支持热更新

2. **API 方法命名**: 导出的方法使用大写开头（Go 导出规则），会自动生成驼峰命名的 JavaScript 绑定

3. **错误处理**: 所有 API 方法应检查 `a.initError`，如果数据库初始化失败应返回错误

4. **Wails Events**: 流式输出使用 `runtime.EventsEmit()` 发送事件，前端使用 `EventsOn()` 监听

### Git 提交规范

- 使用 Conventional Commits 格式
- 中文提交消息
- 示例: `feat: 添加项目拖拽排序功能`

## 配置文件位置

- **Windows**: `C:\Users\<username>\.ai-commit-hub\`
- **macOS/Linux**: `~/.ai-commit-hub/`

### 配置文件
- `config.yaml`: AI Provider 配置、语言设置
- `ai-commit-hub.db`: SQLite 数据库
- `prompts/`: 自定义 Prompt 模板目录

## 常见问题

### Wails 绑定生成错误

Windows 上可能出现 `wailsbindings.exe: %1 is not a valid Win32 application` 错误。

**解决方案**:
1. 删除临时目录下的 wbindings 文件
2. 重新运行 `wails dev`
3. 或使用已有的绑定文件，直接 `go build`

### 前后端类型同步

Go 结构体修改后需要重新生成 Wails 绑定。确保 `frontend/src/types/index.ts` 中的 TypeScript 类型与 Go 结构体保持一致。
