# AI Commit Hub

> 基于 AI 的智能 Git Commit 消息生成工具

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/allanpk716/ai-commit-hub)](https://goreportcard.com/report/github.com/allanpk716/ai-commit-hub)

## 特性

- 🤖 **AI 驱动**: 使用多种 AI Provider 生成规范的 commit 消息
- 📦 **多项目管理**: 同时管理多个 Git 项目，支持拖拽排序
- 🔄 **流式输出**: 实时显示 AI 生成的 commit 消息
- 🚀 **一键推送**: 生成、提交、推送一站式完成
- 🔔 **Pushover 集成**: 支持 Pushover 通知和 Hook 管理
- 💾 **离线历史**: 保存 commit 历史记录
- 🎨 **现代化 UI**: 基于 Vue 3 的优雅界面
- 🪟 **系统托盘**: 最小化到托盘，后台运行
- ⚡ **高性能**: 智能缓存和并发优化

## 支持的 AI Provider

- OpenAI (GPT-3.5, GPT-4)
- Anthropic (Claude)
- Google (Gemini)
- DeepSeek
- Ollama (本地模型)
- Phind

## 截图

> 添加应用截图

## 安装

### 从源码构建

**前置要求:**
- Go 1.21+
- Node.js 18+
- Wails CLI

**步骤:**

```bash
# 克隆仓库
git clone https://github.com/allanpk716/ai-commit-hub.git
cd ai-commit-hub

# 安装依赖
go mod tidy
cd frontend && npm install && cd ..

# 构建
wails build
```

### 下载预编译版本

前往 [Releases](https://github.com/allanpk716/ai-commit-hub/releases) 下载最新版本。

## 使用

### 首次使用

1. 启动应用
2. 点击右上角"设置"图标
3. 配置 AI Provider（API Key、模型等）
4. 点击"添加项目"，选择 Git 仓库路径
5. 选择项目，查看暂存区状态
6. 点击"生成 Commit"，AI 将生成 commit 消息
7. 编辑消息（如需要）
8. 点击"提交"
9. 点击"推送"推送到远程仓库

### 配置 AI Provider

**方式 1: UI 设置**
- 点击"设置"按钮
- 选择 Provider
- 输入 API Key（除了 Ollama）
- 选择模型
- 点击"保存"

**方式 2: 配置文件**

编辑 `~/.ai-commit-hub/config.yaml`:

```yaml
provider: openai
api_key: your-api-key
model: gpt-3.5-turbo
language: zh  # commit 消息语言（zh/en）
```

### 自定义 Prompt 模板

在 `~/.ai-commit-hub/prompts/` 目录创建自定义模板：

```
请根据以下 Git diff 生成规范的 commit 消息。

要求：
1. 使用 Conventional Commits 格式
2. 中文描述
3. 简洁明了

Diff:
{{.Diff}}
```

## 开发

### 启动开发服务器

```bash
wails dev
```

### 运行测试

```bash
# Go 后端测试
go test ./... -v

# 前端测试
cd frontend && npm run test

# 集成测试
go test ./tests/integration/... -v

# 基准测试
go test ./tests/benchmark/... -bench=. -benchmem
```

### 代码规范

```bash
# Go 代码格式化
gofumpt -w .

# TypeScript 代码检查
cd frontend && npm run lint
```

## 架构

```
┌─────────────────┐
│   Frontend      │  Vue 3 + TypeScript
│   (Vue 3)       │  - 组件层
│                 │  - Composables
│  ┌───────────┐  │  - Pinia Stores
│  │  Stores   │  │
└──┼───────────┼──┘
   │           │
   │  Wails    │  绑定层
   │  Bindings │
   │           │
┌──┼───────────┼──┐
│  │   App     │  │  Go 后端
│  │  Layer    │  │  - API 方法
│  └───────────┘  │  - Services
│                 │  - Repositories
│   Services     │
│  ┌───────────┐  │
│  │ Repositories│ │
└──┼───────────┼──┘
   │
┌──┼───────────┐  │
│  │   SQLite  │  │  数据库
└──┴───────────┘  │  (GORM)
```

详细架构文档请参考 [docs/architecture/](docs/architecture/)

## 性能优化

项目经过多轮性能优化：

- **后端优化**: 动态并发控制、工作池、连接池
- **前端优化**: StatusCache 缓存、虚拟滚动、防抖节流、计算缓存
- **代码优化**: 模块化重构、接口抽象、统一错误处理

性能指标：
- 启动时间: < 3 秒
- 状态刷新: < 500ms
- 大量项目 (100+): 流畅

详见 [性能优化文档](docs/architecture/performance-optimization.md)

## 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

### 开发流程

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add some amazing feature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

- `feat:` 新功能
- `fix:` Bug 修复
- `refactor:` 重构
- `style:` 代码格式（不影响功能）
- `docs:` 文档更新
- `test:` 测试相关
- `chore:` 构建/工具相关
- `perf:` 性能优化

## 常见问题

### Q: 支持 GitLab/Gitea 等其他 Git 托管服务吗？

A: 是的，只要是标准的 Git 仓库都支持。

### Q: commit 消息支持其他语言吗？

A: 支持，在设置中选择语言（中文/英文）。

### Q: 可以自定义 commit 消息格式吗？

A: 可以，在 `~/.ai-commit-hub/prompts/` 目录创建自定义模板。

### Q: AI Provider 的 API Key 存储在哪里？

A: 存储在本地配置文件 `~/.ai-commit-hub/config.yaml`，不会上传到云端。

### Q: 如何最小化到系统托盘？

A: 点击窗口关闭按钮 (X)，应用将最小化到托盘。右键托盘图标可以恢复窗口或完全退出应用。

### Q: Pushover Hook 是什么？

A: Pushover 是一个 Git Hook，可以在 Git 操作（如 push）时发送通知到移动设备。应用支持自动安装、更新和管理 Pushover Hook。

## 许可证

[MIT License](LICENSE)

## 致谢

- [Wails](https://wails.io/) - 桌面应用框架
- [Vue 3](https://vuejs.org/) - 前端框架
- [GORM](https://gorm.io/) - ORM 库
- [ai-commit](https://github.com/renatogalera/ai-commit) - AI commit 核心功能
- 所有贡献者

## 联系方式

- 作者: allanpk716
- Issues: [GitHub Issues](https://github.com/allanpk716/ai-commit-hub/issues)
- Discussions: [GitHub Discussions](https://github.com/allanpk716/ai-commit-hub/discussions)
