# AI Commit Hub

一个桌面应用，用于为多个 Git 项目生成 AI 驱动的 commit 消息。

## 功能特性

- 支持多种 AI Provider (OpenAI, Anthropic, DeepSeek, Ollama)
- 实时流式生成 commit 消息
- 管理多个 Git 项目
- 支持自定义 Prompt 模板
- 一键复制或直接提交到本地
- 历史记录功能
- 支持中文和英文 commit 消息

## 技术栈

- **后端**: Go 1.21+, Wails v2
- **前端**: Vue 3, TypeScript, Pinia
- **数据库**: SQLite (GORM)
- **AI 集成**: 基于 [ai-commit](https://github.com/renatogalera/ai-commit) 核心包

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- Node.js 18 或更高版本
- Wails CLI v2

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 安装依赖

```bash
# 克隆仓库
git clone https://github.com/allanpk716/ai-commit-hub.git
cd ai-commit-hub

# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend && npm install && cd ..
```

### 配置

1. 创建配置目录并复制配置文件：

```bash
# Windows
mkdir %USERPROFILE%\.ai-commit-hub
copy .ai-commit-hub\config.yaml %USERPROFILE%\.ai-commit-hub\

# macOS/Linux
mkdir -p ~/.ai-commit-hub
cp .ai-commit-hub/config.yaml ~/.ai-commit-hub/
```

2. 编辑配置文件，填入您的 API Key：

```yaml
providers:
  openai:
    apiKey: your-openai-api-key-here
```

### 运行

```bash
# 开发模式
wails dev

# 生产构建
wails build
```

构建完成后，可执行文件位于 `build/bin/` 目录。

## 使用指南

### 基本工作流

1. **添加项目**: 点击 "+ 添加项目" 按钮，选择 Git 仓库目录
2. **暂存变更**: 在 Git 客户端/IDE 中使用 `git add` 暂存文件
3. **选择项目**: 在左侧列表中选择项目
4. **生成消息**:
   - 选择 AI Provider 和语言
   - 点击 "生成 Commit 消息"
   - 查看 AI 实时生成结果
5. **操作**:
   - 点击 "复制" 复制到剪贴板
   - 或点击 "提交到本地" 直接执行 git commit
6. **查看历史**: 历史记录面板显示之前的生成结果

### 项目管理

- **删除项目**: 点击项目右侧的删除按钮
- **排序**: 拖拽项目卡片重新排序
- **移动**: 使用上下箭头按钮微调顺序

### 自定义 Prompt 模板

1. 在配置目录创建 `prompts/` 子目录
2. 在其中创建文本文件（如 `my-custom-prompt.txt`）
3. 在配置文件中引用：

```yaml
prompts:
  commitMessage: my-custom-prompt.txt
```

4. 支持的占位符：
   - `{DIFF}`: Git diff 内容
   - `{LANGUAGE}`: 目标语言 (zh/en)

## 配置说明

### Provider 配置

**OpenAI:**
```yaml
providers:
  openai:
    apiKey: sk-...
    model: gpt-4
    baseURL: https://api.openai.com/v1
```

**Anthropic (Claude):**
```yaml
providers:
  anthropic:
    apiKey: sk-ant-...
    model: claude-3-sonnet-20240229
```

**DeepSeek:**
```yaml
providers:
  deepseek:
    apiKey: sk-...
    model: deepseek-chat
```

**Ollama (本地):**
```yaml
providers:
  ollama:
    baseURL: http://localhost:11434
    model: llama2
```

### 语言设置

- `zh`: 生成中文 commit 消息
- `en`: 生成英文 commit 消息

## 开发

### 运行测试

```bash
# Go 后端测试
go test ./... -v

# 前端测试 (如果配置了)
cd frontend && npm test
```

### 项目结构

```
ai-commit-hub/
├── app.go                 # Wails 应用入口
├── main.go                # 主入口
├── pkg/
│   ├── ai/               # AI 客户端接口
│   ├── provider/         # AI Provider 实现
│   ├── config/           # 配置管理
│   ├── git/              # Git 操作
│   ├── prompt/           # Prompt 构建
│   ├── models/           # 数据模型
│   ├── repository/       # 数据库层
│   └── service/          # 业务逻辑层
├── frontend/
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   ├── stores/       # Pinia 状态管理
│   │   └── types/        # TypeScript 类型定义
│   └── wailsjs/          # Wails 自动生成的绑定
└── .ai-commit-hub/       # 默认配置模板
```

## 常见问题

### Q: 如何切换 AI Provider?

A: 在 Commit Panel 的 "AI 设置" 部分选择不同的 Provider。

### Q: 支持哪些 Git 操作?

A: 当前支持查看暂存区状态和执行 `git commit`。push/pull 等操作请使用 Git 客户端。

### Q: 配置文件在哪里?

A:
- Windows: `C:\Users\<username>\.ai-commit-hub\`
- macOS/Linux: `~/.ai-commit-hub/`

### Q: 如何只使用本地模型 (Ollama)?

A: 安装 [Ollama](https://ollama.ai/) 后，在配置文件中启用 ollama provider 并设置 `provider: ollama`。

### Q: 生成消息很慢怎么办?

A: 可以尝试：
1. 切换到更快的模型（如 gpt-3.5-turbo）
2. 使用本地 Ollama 模型
3. 检查网络连接

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 致谢

- [Wails](https://wails.io/) - 桌面应用框架
- [ai-commit](https://github.com/renatogalera/ai-commit) - AI commit 核心功能
- [Vue.js](https://vuejs.org/) - 前端框架
