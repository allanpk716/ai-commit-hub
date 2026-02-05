# 代码库结构

**分析日期：** 2024-02-05

## 目录布局

```
ai-commit-hub/
├── app.go                    # Wails App 结构体和导出方法
├── main.go                   # Wails 应用入口点
├── tray_icon.go             # 系统托盘图标处理
├── go.mod                   # Go 模块定义
├── wails.json               # Wails 配置文件
├── BUILD.md                 # 构建说明
├── README.md                # 项目说明
├── CLAUDE.md                # Claude Code 指导文件
├── Makefile                 # 构建脚本
├── dev-*.bat                # 开发启动脚本
├── rsrc_windows_*.syso     # Windows 资源文件
├── pkg/                     # Go 后端包
│   ├── ai/                  # AI 客户端抽象
│   ├── aicommit/            # AI Commit 特定逻辑
│   ├── config/              # 配置管理
│   ├── git/                 # Git 命令封装
│   ├── models/              # 数据模型定义
│   ├── prompt/              # Prompt 模板
│   ├── provider/            # AI Provider 实现
│   │   ├── anthropic/       # Anthropic Claude
│   │   ├── deepseek/        # DeepSeek
│   │   ├── google/          # Google Gemini
│   │   ├── ollama/          # Ollama 本地模型
│   │   ├── openai/          # OpenAI GPT
│   │   ├── openai_compat/   # OpenAI 兼容接口
│   │   ├── openrouter/      # OpenRouter 聚合
│   │   ├── phind/           # Phind 代码搜索
│   │   └── registry/        # Provider 注册管理
│   ├── pushover/            # Pushover Hook 管理
│   ├── repository/          # 数据访问层
│   │   ├── db.go            # 数据库初始化
│   │   ├── git_project_repository.go     # Git 项目 CRUD
│   │   ├── commit_history_repository.go   # 历史记录 CRUD
│   │   └── migration.go     # 数据库迁移
│   ├── service/             # 业务逻辑层
│   │   ├── commit_service.go         # Commit 生成服务
│   │   ├── config_service.go        # 配置管理服务
│   │   ├── error_service.go         # 错误处理服务
│   │   ├── project_config_service.go # 项目配置服务
│   │   ├── startup_service.go       # 启动服务
│   │   ├── update_service.go        # 更新服务
│   │   └── *_test.go               # 单元测试
│   ├── update/               # 应用更新逻辑
│   └── version/             # 版本信息
├── frontend/                # Vue3 前端应用
│   ├── src/
│   │   ├── assets/          # 静态资源
│   │   ├── components/      # Vue 组件
│   │   │   ├── __tests__/           # 组件测试
│   │   │   ├── BackendApiTest.vue   # API 测试组件
│   │   │   ├── CommitPanel.vue      # Commit 生成面板
│   │   │   ├── ProjectList.vue      # 项目列表
│   │   │   ├── ProjectStatusHeader.vue # 项目状态头部
│   │   │   ├── SettingsDialog.vue   # 设置对话框
│   │   │   ├── ConfirmDialog.vue     # 确认对话框
│   │   │   ├── DiffViewer.vue       # 差异查看器
│   │   │   ├── StagedList.vue        # 已暂存文件列表
│   │   │   ├── UnstagedList.vue      # 未暂存文件列表
│   │   │   ├── StagingArea.vue       # 暂存区域
│   │   │   ├── PushoverStatusRow.vue # Pushover 状态行
│   │   │   ├── ExtensionInfoDialog.vue # 扩展信息对话框
│   │   │   ├── ErrorToast.vue       # 错误提示
│   │   │   ├── SplashScreen.vue      # 启动画面
│   │   │   ├── FileContextMenu.vue   # 文件上下文菜单
│   │   │   ├── ExcludeDialog.vue    # 排除对话框
│   │   │   ├── UpdateDialog.vue      # 更新对话框
│   │   │   ├── UpdateNotification.vue # 更新通知
│   │   │   ├── StatusSkeleton.vue   # 状态加载骨架
│   │   │   └── HelloWorld.vue        # 示例组件
│   │   ├── stores/          # Pinia 状态管理
│   │   │   ├── __tests__/           # Store 测试
│   │   │   ├── commitStore.ts        # Commit 状态
│   │   │   ├── projectStore.ts       # 项目状态
│   │   │   ├── statusCache.ts        # 状态缓存管理
│   │   │   ├── pushoverStore.ts     # Pushover 状态
│   │   │   ├── errorStore.ts         # 错误状态
│   │   │   ├── startupStore.ts       # 启动状态
│   │   │   └── updateStore.ts        # 更新状态
│   │   ├── types/           # TypeScript 类型定义
│   │   │   ├── index.ts              # 基础类型
│   │   │   └── status.ts             # 状态相关类型
│   │   ├── utils/           # 工具函数
│   │   ├── App.vue          # 主应用组件
│   │   ├── main.ts          # 应用入口
│   │   └── style.css        # 全局样式
│   ├── tests/               # 前端测试
│   │   ├── integration/              # 集成测试
│   │   ├── unit/                     # 单元测试
│   │   │   ├── components/           # 组件测试
│   │   │   └── stores/               # Store 测试
│   │   └── playwright.config.ts       # Playwright 配置
│   ├── package.json         # 前端依赖
│   ├── vite.config.ts       # Vite 配置
│   ├── tsconfig.json        # TypeScript 配置
│   └── wailsjs/             # Wails 生成的绑定
│       ├── go/                   # Go 方法绑定
│       └── runtime/               # Wails 运行时
├── assets/                  # 应用资源
│   └── icons/               # 应用图标
├── build/                   # 构建输出
│   └── windows/             # Windows 特定资源
│       └── icon.ico          # Windows 图标
├── docs/                    # 文档目录
│   ├── lessons-learned/     # 经验总结
│   ├── fixes/               # 修复记录
│   ├── plans/               # 项目计划
│   └── features/            # 功能文档
├── scripts/                 # 脚本文件
├── tests/                   # 集成测试
├── tmp/                     # 临时文件
├── .ai-commit-hub/          # 应用配置目录（运行时创建）
├── .planning/               # 规划文档
│   └── codebase/            # 架构文档
└── .github/                 # GitHub 配置
    └── workflows/           # CI/CD 工作流
```

## 目录用途

### 根目录 (`./`)

**核心文件：**
- `app.go`: Wails App 结构体，包含所有导出 API 方法
- `main.go`: Wails 应用启动入口
- `go.mod/go.sum`: Go 依赖管理
- `wails.json`: Wails 构建和运行配置

**构建相关：**
- `Makefile`: 构建脚本
- `rsrc_windows_*.syso`: Windows 资源文件（图标、版本信息）
- `dev-*.bat`: Windows 开发启动脚本

**文档：**
- `README.md`: 项目说明（GitHub 首页展示）
- `CLAUDE.md`: Claude Code 工作指导
- `docs/`: 文档目录
  - `build/`: 构建说明文档
  - `development/`: 开发规范（Wails、日志等）
  - `history/`: 历史文档（实现总结、最终报告）
  - `lessons-learned/`: 技术经验总结
  - `fixes/`: 问题修复记录
  - `plans/`: 项目开发计划
  - `archive/`: 归档文档（过时设计、安装文档等）

### 后端核心 (`pkg/`)

**服务层 (`pkg/service/`)**:
- `commit_service.go`: Commit 消息生成（支持流式）
- `config_service.go`: AI Provider 配置管理
- `startup_service.go`: 应用启动状态预加载
- `project_config_service.go`: 项目特定配置管理

**数据层 (`pkg/repository/`)**:
- `git_project_repository.go`: Git 项目数据访问
- `commit_history_repository.go`: 历史记录数据访问
- `db.go`: 数据库连接和初始化
- `migration.go`: 数据库迁移逻辑

**AI 层 (`pkg/ai/`, `pkg/provider/`)**:
- `ai/ai.go`: AI 客户端接口定义
- `provider/*/`: 各个 AI Provider 实现
- `provider/registry/`: Provider 注册管理

**工具层 (`pkg/git/`, `pkg/config/`, `pkg/pushover/`)**:
- `git/`: Git 命令封装
- `config/`: 配置文件解析
- `pushover/`: Pushover Hook 管理

### 前端核心 (`frontend/`)

**组件 (`frontend/src/components/`)**:
- `ProjectList.vue`: 项目列表（可拖拽、搜索）
- `CommitPanel.vue`: Commit 生成界面
- `SettingsDialog.vue`: 设置界面
- `ProjectStatusHeader.vue`: 项目状态显示
- 暂存相关：`StagedList.vue`、`UnstagedList.vue`、`StagingArea.vue`

**状态管理 (`frontend/src/stores/`)**:
- `statusCache.ts`: 状态缓存管理（核心）
- `commitStore.ts`: Commit 生成状态
- `projectStore.ts`: 项目列表状态
- `pushoverStore.ts`: Pushover 扩展状态

**类型定义 (`frontend/src/types/`)**:
- `index.ts`: 与 Go 结构体同步的基础类型
- `status.ts`: StatusCache 相关类型

### 构建和资源

**构建输出 (`build/`)**:
- `windows/icon.ico`: Windows 托盘图标

**前端资源 (`frontend/dist/`)**:
- 构建后的静态文件（HTML、CSS、JS）

## 关键文件位置

### 入口点

**应用入口：**
- `main.go`: Wails 应用启动
- `app.go`: 后端 API 导出
- `frontend/src/main.ts`: Vue 应用启动
- `frontend/src/App.vue`: 主应用组件

### 配置文件

**应用配置：**
- `wails.json`: Wails 构建配置
- `frontend/package.json`: 前端依赖配置
- `frontend/vite.config.ts`: Vite 构建配置

**运行时配置：**
- `~/.ai-commit-hub/config.yaml`: 用户配置文件
- `~/.ai-commit-hub/ai-commit-hub.db`: SQLite 数据库

### 核心业务逻辑

**后端：**
- `pkg/service/commit_service.go`: AI Commit 生成
- `pkg/repository/git_project_repository.go`: Git 项目管理
- `pkg/ai/ai.go`: AI 接口抽象

**前端：**
- `frontend/src/stores/statusCache.ts`: 状态缓存管理
- `frontend/src/components/CommitPanel.vue`: Commit 界面
- `frontend/src/components/ProjectList.vue`: 项目管理界面

## 命名约定

### Go 代码

**文件命名：**
- 小写 + 下划线：`git_project_repository.go`
- 驼峰式导出：`ProjectConfigService`

**函数命名：**
- 公开方法：大写开头 `GetProjects()`
- 私有方法：小写开头 `getProjectPath()`
- 接口：`AIClient`、`Repository`

**变量命名：**
- 结构体：大写开头 `GitProject`
- 私有变量：小写开头 `projectPath`
- 接口接收者：简短名称 `g *GitProject`

### Vue/TypeScript 代码

**文件命名：**
- PascalCase：`ProjectList.vue`、`StatusCache.ts`

**组件命名：**
- PascalCase：`ProjectList`、`CommitPanel`

**状态管理：**
- Store 文件：`projectStore.ts`
- State 属性：`projects`、`loading`
- Action 方法：`fetchProjects`、`addProject`

**事件命名：**
- Wails Events：`startup-complete`、`commit-delta`
- 组件事件：`project-selected`、`commit-generated`

## 新增代码指南

### 新增 Git 功能

**后端位置：**
- 逻辑：`pkg/service/` 新增服务或修改现有服务
- 数据访问：`pkg/repository/git_project_repository.go`
- API 导出：`app.go` 添加新方法

**前端位置：**
- 状态管理：`frontend/src/stores/statusCache.ts`
- 组件：`frontend/src/components/` 新增或修改组件
- 类型定义：`frontend/src/types/index.ts`

### 新增 AI Provider

**位置：**
- 实现：`pkg/provider/[name]/` 新建目录
- 注册：自动通过匿名导入注册
- 测试：`pkg/service/commit_service_test.go`

**步骤：**
1. 在 `pkg/provider/` 创建新目录
2. 实现 `AIClient` 接口
3. 匿名导入触发 `init()` 注册

### 新增前端组件

**位置：**
- 组件：`frontend/src/components/[Name].vue`
- 测试：`frontend/src/components/__tests__/[Name].spec.ts`
- 样式：组件内部 scoped style

**结构：**
```vue
<template>
  <!-- 模板 -->
</template>

<script setup lang="ts">
import { ref } from 'vue'
// 逻辑
</script>

<style scoped>
/* 样式 */
</style>
```

## 特殊目录说明

### `.planning/codebase/`
架构分析和规划文档目录，包含：
- `ARCHITECTURE.md`: 架构模式、层次、数据流
- `STRUCTURE.md`: 目录布局、关键位置
- `CONVENTIONS.md`: 编码规范
- `TESTING.md`: 测试模式
- `CONCERNS.md`: 技术债务

### `docs/lessons-learned/`
经验总结和技术文档：
- Windows 托盘图标实现指南
- 双击功能修复记录
- 退出问题修复

### `tmp/`
临时测试文件目录（应在 .gitignore 中忽略）：
- 临时测试数据和脚本
- 快速原型验证
- 调试信息和日志
- **注意**: 此目录内容会被 git ignore，不应提交到仓库

---

*结构分析：2024-02-05*