# 项目级别 AI 配置功能

## 概述

AI Commit Hub 支持为不同的 Git 项目配置独立的 AI Provider 和语言设置。

## 功能特性

### 默认配置

新添加的项目默认使用配置文件中的全局设置：
- Provider: 配置文件中 `provider` 字段
- Language: 配置文件中 `language` 字段

### 自定义配置

可以为特定项目设置独立的 AI 配置：
- 打开项目详情
- 在 AI 配置区域选择 Provider 和 Language
- 配置会立即保存到数据库

### 恢复默认

点击"恢复默认"按钮可将项目配置重置为全局默认值。

### 配置验证

当项目的配置与配置文件不一致时（如 Provider 已删除），会显示警告提示用户确认重置。

## 使用示例

### 场景 1: 开源项目用英文

1. 选择开源项目
2. 将 Language 设置为 "English"
3. 后续所有 commit 消息都使用英文

### 场景 2: 个人项目用中文

1. 选择个人项目
2. 将 Language 设置为 "中文"
3. 后续所有 commit 消息都使用中文

### 场景 3: 不同项目使用不同 Provider

1. 项目 A 使用 DeepSeek（快速）
2. 项目 B 使用 OpenAI（质量更高）
3. 每个项目独立配置，互不影响

## 配置文件位置

- Windows: `C:\Users\<username>\.ai-commit-hub\config.yaml`
- macOS/Linux: `~/.ai-commit-hub/config.yaml`

## 技术实现

### 数据库存储

项目配置存储在 SQLite 数据库的 `git_projects` 表中：
- `provider`: AI 服务商（可为 null 表示使用默认）
- `language`: 语言设置（可为 null 表示使用默认）
- `model`: 模型名称（可为 null 表示使用默认）
- `use_default`: 是否使用默认配置的标识

### 配置优先级

1. 如果 `use_default = true` 或配置字段为 null，使用配置文件默认值
2. 否则使用数据库中存储的项目配置

### 配置验证

切换项目时会验证配置的有效性：
- 检查 Provider 是否在配置文件中存在
- 检查 Language 是否有效（zh/en/chinese/english）
- 如果配置无效，显示警告并提供重置选项
