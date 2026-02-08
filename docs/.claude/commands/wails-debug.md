---
description: Wails 开发调试助手 - 后端日志分析和前端浏览器自动化测试
argument-hint: [--backend] [--frontend] [--log] [--help]
allowed-tools: Bash, Read, Skill, Task, TaskCreate, TaskUpdate, TaskList, TaskGet
---

# Wails Debug 命令

AI Commit Hub 项目的 Wails 开发调试助手，帮助开发者快速诊断和解决前后端问题。

## 快速开始

```bash
# 交互式调试流程（推荐）
/wails-debug

# 快速查看最新错误日志
/wails-debug --log

# 后端调试：启动开发服务器和分析日志
/wails-debug --backend

# 前端调试：浏览器自动化测试
/wails-debug --frontend

# 显示帮助信息
/wails-debug --help
```

## 功能特性

### 🔍 后端调试 (`--backend`)
- 启动和管理 `wails dev` 进程
- 分析 Go 后端日志（`logs/` 目录）
- 检测端口占用和进程冲突
- 实时监控错误和警告日志

### 🌐 前端调试 (`--frontend`)
- 浏览器自动化测试（使用 dev-Browser 技能）
- 捕获浏览器控制台输出
- 截图和记录 UI 状态
- 测试前端功能和交互

### 📋 日志查看 (`--log`)
- 快速查看最新的错误和警告
- 过滤特定时间段的日志
- 分析堆栈跟踪信息

### 🔄 交互式调试（无参数）
- 智能诊断流程
- 并行启动前后端调试代理
- 生成完整的诊断报告

## 使用场景

### 场景 1: 启动开发环境
```bash
/wails-debug --backend
```

**执行内容**:
1. 检查端口 34115 占用情况
2. 清理旧日志（可选）
3. 启动 `wails dev`（后台模式）
4. 监控启动日志，确认成功
5. 返回服务 URL 和状态

### 场景 2: 查看错误日志
```bash
/wails-debug --log
```

**执行内容**:
1. 列出最新的日志文件
2. 读取最后 100 行
3. 过滤 ERROR/WARN 级别日志
4. 提取错误消息和堆栈信息
5. 生成问题摘要

### 场景 3: 前端功能测试
```bash
/wails-debug --frontend
```

**执行内容**:
1. 确认 `wails dev` 已启动
2. 使用浏览器自动化打开 `http://localhost:34115`
3. 执行用户指定的测试流程
4. 捕获控制台错误和网络请求
5. 截图记录关键状态

### 场景 4: 端到端调试
```bash
/wails-debug
```

**执行内容**:
1. 并行启动后端和前端调试代理
2. 前端代理执行问题操作
3. 后端代理记录对应时间的日志
4. 综合分析前后端信息
5. 生成诊断报告

## 参数说明

| 参数 | 说明 | 子代理类型 |
|------|------|-----------|
| `--backend` | 启动后端调试子代理（日志分析、进程管理） | Bash |
| `--frontend` | 启动前端测试子代理（浏览器自动化） | general-purpose |
| `--log` | 快速查看最新的错误和警告日志 | Bash |
| `--help` | 显示此帮助信息 | - |
| 无参数 | 交互式调试流程 | Bash + general-purpose |

## 执行策略

### 核心原则
1. **使用子代理执行具体调试任务** - 避免影响主会话上下文
2. **并行处理独立任务** - 提高调试效率
3. **避免读取大文件** - 使用 Bash 工具过滤日志
4. **生成结构化报告** - 便于问题追踪

### 子代理分工

#### 后端调试子代理
- **类型**: Bash
- **职责**:
  - 启动 `wails dev` 进程
  - 监控和过滤日志输出
  - 查找错误和警告信息
  - 分析 Go 后端问题
- **工具**: Bash, Read

#### 前端测试子代理
- **类型**: general-purpose
- **职责**:
  - 使用 dev-Browser 技能打开测试页面
  - 自动化测试前端功能
  - 捕获浏览器控制台输出
  - 截图和记录 UI 状态
- **工具**: Skill, Task, Bash

## 常用命令参考

### 后端日志分析
```bash
# 查看最新日志文件
ls -lt logs/*.log | head -1

# 读取最后 50 行
tail -50 logs/logger--*.log

# 过滤错误日志
grep -i "error\|warn\|panic" logs/logger--*.log

# 实时监控日志
tail -f logs/logger--*.log
```

### 进程管理
```bash
# 检查端口占用
netstat -ano | findstr :34115

# 终止指定进程
taskkill /PID <pid> /F

# 查找 wails dev 进程
tasklist | findstr wails
```

### 前端测试
```bash
# 前端单元测试
cd frontend && npm run test:run

# 前端类型检查
cd frontend && npm run type-check

# 前端构建检查
cd frontend && npm run build
```

## 调试报告模板

```markdown
## 调试报告

### 时间
{{datetime}}

### 任务
{{user_request}}

### 后端状态
- Wails Dev 运行状态: ✅/❌
- 日志文件: {{log_file}}
- 错误数量: {{error_count}}

### 前端状态
- 页面加载: ✅/❌
- Console 错误: {{console_errors}}
- 网络请求失败: {{failed_requests}}

### 发现的问题
1. **[问题标题]**
   - 位置: {{file:line}}
   - 错误信息: {{error_message}}
   - 可能原因: {{possible_cause}}
   - 建议: {{suggestion}}

### 下一步操作
- [ ] {{action_item_1}}
- [ ] {{action_item_2}}
```

## 常见问题

### Q: Wails Dev 启动失败？
**A**: 检查端口 34115 是否被占用：
```bash
netstat -ano | findstr :34115
taskkill /PID <pid> /F
```

### Q: 前端编译错误？
**A**: 重新安装前端依赖：
```bash
cd frontend && npm install
```

### Q: 日志文件过大？
**A**: 使用 grep 过滤或清理旧日志：
```bash
grep -i "error" logs/logger--*.log
```

### Q: 浏览器无法连接？
**A**: 确认 `wails dev` 已启动，检查防火墙设置

## 相关文档

- 技能详情: `docs/.claude/skills/wails-debug-helper/SKILL.md`
- Wails 开发规范: `docs/development/wails-development-standards.md`
- 日志输出规范: `docs/development/logging-standards.md`

## 技能触发关键词

此命令关联的技能会在以下情况自动触发：
- 提到"调试"、"debug"、"查看日志"
- 提到"启动开发服务器"、"运行开发环境"
- 提到"检查前端问题"、"测试前端"
- 提到"查看错误信息"、"排查问题"
- 提到"浏览器调试"、"自动化测试"

## 开发者信息

- **项目**: AI Commit Hub
- **技术栈**: Wails (Go + Vue3)
- **开发环境**: Windows
- **日志位置**: `logs/`
- **前端 URL**: `http://localhost:34115`
