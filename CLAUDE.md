# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

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
wails dev              # 启动开发服务器
wails build           # 生产构建
go build -o build/bin/ai-commit-hub.exe .  # 仅构建后端
cd frontend && npm run dev  # 前端开发
```

### 测试命令
```bash
go test ./... -v       # Go 后端测试
cd frontend && npm run test:run  # 前端测试
```

### 依赖管理
```bash
go mod tidy           # Go 依赖
cd frontend && npm install  # 前端依赖
```

## 代码架构

### 后端 (Go)
```
app.go                          # Wails 应用入口
pkg/service/                    # 业务逻辑层
pkg/repository/                 # 数据访问层
pkg/ai/, pkg/provider/          # AI Provider 抽象层
pkg/git/                        # Git 命令封装
pkg/config/                     # YAML 配置解析
```

### 前端 (Vue3)
```
frontend/src/App.vue            # 主应用布局
frontend/src/components/        # Vue 组件
frontend/src/stores/            # Pinia 状态管理
frontend/src/types/             # TypeScript 类型定义
```

## 开发规则

### 核心规则
1. **使用中文**回答问题和编写文档
2. **BAT 脚本**中不要使用中文（避免编码问题）
3. **临时文件**统一放在 `tmp/` 文件夹
4. **图片处理**前确保尺寸小于 1000x1000
5. **计划文件**统一放在 `docs/plans/` 目录

### Wails 开发规范
详见：`docs/development/wails-development-standards.md`

- 使用自定义 `Command()` 函数隐藏 Windows 控制台窗口
- 统一使用 `github.com/WQGroup/logger` 日志库
- API 方法需检查 `a.initError`
- 使用 Wails Events 实现流式输出

### 日志输出规范
详见：`docs/development/logging-standards.md`

- 统一使用 `github.com/WQGroup/logger` 日志库
- 支持多种日志级别和格式（JSON、文本）
- 支持日志轮转和自动清理

### Git 提交规范
- 使用 Conventional Commits 格式
- 中文提交消息
- 示例：`feat: 添加项目拖拽排序功能`

## 配置文件位置

- **Windows**: `C:\Users\<username>\.ai-commit-hub\`
- **macOS/Linux**: `~/.ai-commit-hub/`

配置文件：`config.yaml`, `ai-commit-hub.db`, `prompts/`

## 常见问题

### Wails 绑定生成错误
Windows 上可能出现 `wailsbindings.exe: %1 is not a valid Win32 application` 错误。

**解决方案**:
1. 删除临时目录下的 wbindings 文件
2. 重新运行 `wails dev`
3. 或使用已有的绑定文件，直接 `go build`

## Claude Code 技能

项目特定的 Claude Code 技能存放在 `.claude/skills/` 目录：

- **wails-debug-helper**: AI Commit Hub 项目调试助手
  - 位置：`.claude/skills/wails-debug-helper/SKILL.md`
  - 用途：使用子代理进行后端日志分析和前端浏览器自动化测试，避免影响主会话上下文
  - 触发词：调试、debug、查看日志、启动开发服务器、检查前端问题、浏览器调试等

## 详细文档导航

### Claude Code 配置
- Wails 调试助手：`docs/.claude/README.md`
- 调试技能：`.claude/skills/wails-debug-helper/SKILL.md`（可被 Claude Code 自动加载）
- 调试技能（文档参考）：`docs/.claude/skills/wails-debug-helper/SKILL.md`
- 调试命令：`docs/.claude/commands/wails-debug.md`

### 开发规范
- Wails 开发规范：`docs/development/wails-development-standards.md`
- 日志输出规范：`docs/development/logging-standards.md`

### 经验总结
- Windows 控制台窗口隐藏：`docs/lessons-learned/windows-console-hidden-fix.md`
- 系统托盘实现指南：`docs/lessons-learned/windows-tray-icon-implementation-guide.md`
- 双击功能修复：`docs/fixes/tray-icon-doubleclick-fix.md`
- 退出应用修复：`docs/fixes/systray-exit-fix.md`

### 状态管理
- StatusCache 实现：`frontend/src/stores/statusCache.ts`
- StatusCache 测试：`frontend/src/stores/__tests__/statusCache.spec.ts`

### 启动流程
- 后端启动：`app.go` 中的 `startup()` 方法
- 前端启动：`frontend/src/App.vue` 和 `frontend/src/main.ts`
