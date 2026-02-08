---
name: wails-debug-helper
description: "AI Commit Hub 项目调试助手 - 使用子代理进行后端日志分析和前端浏览器自动化测试，避免影响主会话上下文"
---

# Wails Debug Helper

本地调试助手，专门用于 AI Commit Hub 项目的开发调试。使用子代理（subagent）执行调试任务，避免影响主会话上下文。

## 触发条件

当用户提到以下内容时使用此技能：
- "调试"、"debug"、"查看日志"
- "启动开发服务器"、"运行开发环境"
- "检查前端问题"、"测试前端"
- "查看错误信息"、"排查问题"
- "浏览器调试"、"自动化测试"

## 技能描述

AI Commit Hub 项目调试助手，协助开发者：
- 启动和管理 Wails 开发环境
- 分析后端 Go 日志（`logs/` 目录）
- 使用浏览器自动化测试前端功能
- 捕获和分析浏览器开发者工具信息
- 定位和诊断前后端问题

## 上下文设置

### 项目根目录
```
C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub
```

### 日志位置
```
logs/*.log  # Go 后端日志
```

### 开发命令
```bash
wails dev              # 启动开发服务器（默认端口 34115）
wails dev -debug       # 启动调试模式（可附加 Go 调试器）
```

### 前端 URL
```
http://localhost:34115  # Wails 开发服务器
```

## 执行策略

### 核心原则
1. **优先使用 Task 工具启动子代理**执行具体调试任务
2. **避免在主会话中执行长时间运行的调试操作**
3. **使用 Bash 工具检查日志文件**（不要用 Read 工具读取大日志）
4. **使用 dev-Browser 技能进行浏览器自动化测试**

### 子代理分工

#### 后端调试子代理
```yaml
subagent_type: Bash
职责:
  - 启动 wails dev 进程
  - 监控和过滤日志输出
  - 查找错误和警告信息
  - 分析 Go 后端问题
工具:
  - Bash: 运行 wails 命令、过滤日志
  - Read: 读取小配置文件
```

#### 前端测试子代理
```yaml
subagent_type: general-purpose
职责:
  - 使用 dev-Browser 技能打开测试页面
  - 自动化测试前端功能
  - 捕获浏览器控制台输出
  - 截图和记录 UI 状态
工具:
  - Skill: 调用 dev-Browser
  - Task: 启动子任务
```

## 调试工作流

### 场景 1: 启动开发环境

**用户请求**: "启动开发服务器" / "开始调试"

**执行步骤**:
1. 检查端口占用（`netstat -ano | findstr :34115`）
2. 清理旧日志（可选）
3. 启动 wails dev（后台运行）
4. 监控启动日志，确认成功

**子代理提示词**:
```
你是后端调试助手。任务：
1. 检查端口 34115 是否被占用，如果被占用则终止占用进程
2. 切换到项目根目录
3. 启动 wails dev（后台模式）
4. 监控 logs/ 目录下的日志输出
5. 等待启动成功消息（"Server started" 或 "frontend connected"）
6. 返回启动状态和 URL

使用 Bash 工具执行所有命令。
```

### 场景 2: 查看后端日志

**用户请求**: "查看日志" / "有什么错误"

**执行步骤**:
1. 列出最新的日志文件
2. 读取最后 N 行（使用 tail）
3. 过滤错误/警告级别日志
4. 分析和总结关键问题

**子代理提示词**:
```
你是日志分析助手。任务：
1. 列出 logs/ 目录下的所有日志文件（按时间排序）
2. 读取最新日志文件的最后 100 行
3. 过滤并高亮显示包含 "ERROR"、"WARN"、"panic"、"fatal" 的行
4. 提取最近的错误消息和堆栈信息
5. 总结发现的问题（中文）

使用 Bash 工具运行 tail、grep 等命令。
```

### 场景 3: 前端功能测试

**用户请求**: "测试前端" / "检查 [功能] 是否正常"

**执行步骤**:
1. 确认 wails dev 已启动
2. 使用 dev-Browser 技能打开页面
3. 自动化执行用户操作流程
4. 捕获控制台输出和网络请求
5. 截图记录关键状态
6. 检查 UI 是否符合预期

**子代理提示词**:
```
你是前端测试助手。任务：
1. 使用 dev-Browser 技能访问 http://localhost:34115
2. 等待页面加载完成
3. 执行以下测试：
   - 打开浏览器开发者工具（Console、Network 标签）
   - 测试 [用户指定的功能]
   - 捕获所有控制台错误和警告
   - 截图记录关键状态
   - 检查网络请求状态码
4. 返回测试结果和发现的问题

使用 Skill 工具调用 dev-Browser。
```

### 场景 4: 端到端调试

**用户请求**: "调试 [具体问题]" / "为什么 [功能] 不工作"

**执行步骤**:
1. **并行启动两个子代理**：
   - 后端代理：监控日志
   - 前端代理：复现问题
2. 前端代理执行问题操作
3. 后端代理记录对应时间的日志
4. 综合分析前后端信息
5. 生成诊断报告

**双子代理提示词**:
```
后端代理:
监控 logs/ 目录，实时追踪日志输出，特别关注时间窗口 [开始时间] 到 [结束时间] 的日志。
过滤 ERROR/WARN 级别日志，提取相关 API 调用和错误堆栈。

前端代理:
使用 dev-Browser 访问应用，执行以下步骤：
1. 登录/初始化
2. 触发 [问题功能]
3. 打开开发者工具 Console
4. 记录所有错误消息
5. 截图当前状态
6. 检查 Network 标签的失败请求

返回完整的诊断报告（中文）。
```

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

# 查找特定时间段的日志
grep "2025-02-08 14:" logs/logger--*.log
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

## 输出格式

### 调试报告模板
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

## 错误处理

### 常见问题及解决方案

#### Wails Dev 启动失败
```
错误: "bind: address already in use"
解决: 终止占用 34115 端口的进程
```

#### 前端编译错误
```
错误: "Cannot find module"
解决: cd frontend && npm install
```

#### 日志文件过大
```
错误: 日志文件超过 100MB
解决: 清理旧日志或使用 grep 过滤
```

#### 浏览器无法连接
```
错误: "ERR_CONNECTION_REFUSED"
解决: 检查 wails dev 是否启动，确认防火墙设置
```

## 限制和注意事项

1. **仅在开发环境使用**，不要在生产环境运行
2. **日志文件可能很大**，避免使用 Read 工具读取完整文件
3. **浏览器自动化可能需要等待**，某些操作需要页面加载完成
4. **子代理有超时限制**，长时间运行的任务需要分步执行
5. **Windows 路径处理**，使用正斜杠或转义反斜杠

## 示例对话

**用户**: 调试一下为什么点击"检查更新"按钮没反应

**助手**:
我来帮你调试这个问题。让我启动两个调试代理来分析前后端。

<启动后端日志监控代理>
<启动前端测试代理>

等待调试结果...

**调试报告**:
- 前端按钮点击事件已触发 ✅
- Console 无错误 ✅
- Network 标签显示请求已发送 ✅
- 后端日志显示 API 收到请求 ✅
- **问题**: `updateStore.ts:58` 抛出 "Cannot read property 'hasUpdate' of undefined"
- **原因**: API 返回的数据结构与前端期望不匹配
- **建议**: 检查 `CheckForUpdates` Go 函数的返回值结构
