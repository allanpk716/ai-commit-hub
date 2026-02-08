# AI Commit Hub - Claude Code 配置

这个目录包含 AI Commit Hub 项目特定的 Claude Code 配置文件，包括技能和斜杠命令。所有文件都被 Git 跟踪，便于团队协作。

## 目录结构

```
docs/.claude/
├── skills/
│   └── wails-debug-helper/
│       └── SKILL.md            # Wails 调试技能
├── commands/
│   └── wails-debug.md          # Wails 调试斜杠命令
└── README.md                   # 本文件
```

## 可用命令

### `/wails-debug` - Wails 开发调试助手

快速诊断和解决 AI Commit Hub 项目的前后端问题。

**用法**:
```bash
/wails-debug                    # 交互式调试流程
/wails-debug --backend          # 后端调试
/wails-debug --frontend         # 前端调试
/wails-debug --log              # 查看日志
/wails-debug --help             # 显示帮助
```

**功能**:
- 🔍 后端调试：启动开发服务器、分析 Go 日志
- 🌐 前端调试：浏览器自动化测试
- 📋 日志查看：快速查看错误和警告
- 🔄 端到端调试：并行分析前后端问题

详见：[commands/wails-debug.md](commands/wails-debug.md)

## 可用技能

### `wails-debug-helper` - Wails 调试助手

专门用于 AI Commit Hub 项目的开发调试技能。

**触发关键词**:
- "调试"、"debug"、"查看日志"
- "启动开发服务器"、"运行开发环境"
- "检查前端问题"、"测试前端"
- "查看错误信息"、"排查问题"

**功能**:
- 使用子代理执行调试任务，避免影响主会话
- 后端日志分析（`logs/` 目录）
- 前端浏览器自动化测试
- 捕获和分析开发者工具信息
- 定位和诊断前后端问题

详见：[skills/wails-debug-helper/SKILL.md](skills/wails-debug-helper/SKILL.md)

## 快速开始

### 1. 验证配置

检查文件是否被 Git 跟踪：

```bash
git ls-files docs/.claude/
```

应该看到：
- `docs/.claude/skills/wails-debug-helper/SKILL.md`
- `docs/.claude/commands/wails-debug.md`
- `docs/.claude/README.md`

### 2. 使用斜杠命令

在 Claude Code 中直接调用：

```bash
# 查看最新的错误日志
/wails-debug --log

# 启动后端调试
/wails-debug --backend

# 前端功能测试
/wails-debug --frontend
```

### 3. 技能自动触发

在对话中自然提到调试相关关键词：

```
用户: 调试一下为什么点击按钮没反应
Claude: [自动触发 wails-debug-helper 技能]
```

## 开发指南

### 添加新技能

1. 在 `skills/` 目录下创建新技能文件夹
2. 创建 `SKILL.md` 文件，遵循技能规范
3. 更新本 README 文档

**技能模板**:
```markdown
---
name: skill-name
description: "简短描述"
---

# 技能名称

## 触发条件
- 关键词 1
- 关键词 2

## 技能描述
详细说明技能功能...

## 执行策略
具体的执行步骤...
```

### 添加新斜杠命令

1. 在 `commands/` 目录下创建新的 `.md` 文件
2. 添加 YAML frontmatter
3. 编写命令说明和使用示例

**命令模板**:
```markdown
---
description: 命令描述
argument-hint: [--arg1] [--arg2]
allowed-tools: Tool1, Tool2, Tool3
---

# 命令名称

## 快速开始
...

## 参数说明
...
```

### 技能 vs 斜杠命令

**技能（Skills）**:
- 自动触发（基于关键词）
- 被动响应
- 适合复杂的多步骤流程

**斜杠命令（Commands）**:
- 手动调用
- 主动执行
- 适合明确的任务和参数化操作

## 相关文档

### 项目文档
- [CLAUDE.md](../../CLAUDE.md) - 项目主文档
- [Wails 开发规范](../../development/wails-development-standards.md)
- [日志输出规范](../../development/logging-standards.md)

### Claude Code 文档
- [Claude Code 官方文档](https://claude.ai/code)
- [技能开发指南](https://github.com/anthropics/claude-code/blob/main/docs/skills.md)
- [斜杠命令开发指南](https://github.com/anthropics/claude-code/blob/main/docs/commands.md)

## 常见问题

### Q: 为什么把配置放在 `docs/.claude/` 而不是根目录的 `.claude/`？

**A**: 根目录的 `.claude/` 被 `.gitignore` 排除（第 44 行），无法被 Git 跟踪。将项目特定的配置放在 `docs/.claude/` 可以：
- ✅ 被 Git 跟踪和版本控制
- ✅ 与其他开发者共享
- ✅ 与项目文档放在一起
- ✅ 不影响用户个人的 `.claude/` 配置

### Q: 技能文件和斜杠命令文件有什么区别？

**A**:
- **技能文件（SKILL.md）**：定义自动触发的行为，包含触发条件、执行策略、子代理配置
- **斜杠命令文件**：定义手动调用的命令，包含参数说明、使用示例、帮助文本

两者可以相互引用，形成完整的调试工具链。

### Q: 如何在本地测试新添加的技能或命令？

**A**:
1. **技能测试**：在对话中提到触发关键词，观察是否自动触发
2. **命令测试**：直接输入 `/command-name` 验证功能
3. **文件检查**：确认 YAML 格式正确（使用 linter 或在线验证器）

### Q: 团队成员如何使用这些配置？

**A**:
1. 拉取最新代码：`git pull`
2. Claude Code 会自动识别 `docs/.claude/` 下的配置
3. 直接使用 `/wails-debug` 命令
4. 或在对话中自然触发技能

### Q: 如何贡献新的技能或命令？

**A**:
1. 遵循现有模板和规范
2. 添加完整的文档和示例
3. 测试功能是否正常
4. 提交 PR 并描述改动

## 维护指南

### 定期检查
- [ ] 验证所有技能和命令正常工作
- [ ] 更新文档和示例
- [ ] 清理过时的内容
- [ ] 收集用户反馈

### 版本兼容性
- 测试新版本 Claude Code 的兼容性
- 更新 YAML frontmatter 格式
- 调整工具权限设置

## 贡献者

- Allan PK716 <allanpk716@gmail.com>

## 许可证

与 AI Commit Hub 项目保持一致。
