# Provider 过滤与错误处理功能测试指南

## 测试前准备

1. **启动应用**
   ```bash
   wails dev
   ```

2. **打开配置文件**
   - Windows: `C:\Users\<username>\.ai-commit-hub\config.yaml`
   - macOS/Linux: `~/.ai-commit-hub/config.yaml`

---

## 测试场景

### 场景 1：只配置一个 Provider

**步骤：**
1. 编辑 `config.yaml`，只保留一个 provider 配置：
```yaml
providers:
  openai:
    apiKey: "sk-test-key"
    model: "gpt-4"
    baseURL: "https://api.openai.com/v1"
```

2. 保存文件，重启应用

3. 打开应用，选择一个项目

4. 查看 Provider 下拉框

**预期结果：**
- OpenAI 显示为可选中状态
- Anthropic、DeepSeek、Ollama、Google、Phind 显示但禁用
- 禁用的 providers 旁边显示 "(未配置: 未在配置文件中添加)"

---

### 场景 2：配置多个 Providers

**步骤：**
1. 编辑 `config.yaml`，添加多个 providers：
```yaml
providers:
  openai:
    apiKey: "sk-test-key"
    model: "gpt-4"
  anthropic:
    apiKey: "sk-ant-test-key"
    model: "claude-3-opus-20240229"
  ollama:
    model: "llama2"
    baseURL: "http://localhost:11434"
```

2. 保存文件，重启应用

3. 打开应用，查看 Provider 下拉框

**预期结果：**
- OpenAI、Anthropic、Ollama 显示为可选中状态
- DeepSeek、Google、Phind 显示但禁用
- 禁用的 providers 旁边显示 "(未配置: 未在配置文件中添加)"

---

### 场景 3：缺少 API Key

**步骤：**
1. 编辑 `config.yaml`，添加一个没有 API Key 的 provider：
```yaml
providers:
  openai:
    apiKey: "sk-test-key"
    model: "gpt-4"
  anthropic:
    model: "claude-3-opus-20240229"
    # 缺少 apiKey
```

2. 保存文件，重启应用

3. 查看 Provider 下拉框

**预期结果：**
- OpenAI 显示为可选中状态
- Anthropic 显示但禁用，旁边显示 "(未配置: 缺少 API Key)"

---

### 场景 4：尝试使用未配置的 Provider（如果 UI 允许）

**步骤：**
1. 准备一个有 staged 文件的 Git 项目
2. 选择一个未配置的 provider（如果 UI 允许选择）
3. 点击"生成 Commit 消息"按钮

**预期结果：**
- 界面右下角显示错误横幅
- 错误消息：Provider 'xxx' 未配置，请先在配置文件中添加
- 生成状态重置

---

### 场景 5：配置动态加载

**步骤：**
1. 初始状态：`config.yaml` 只有 openai 配置
2. 打开应用，确认只有 openai 可选
3. 关闭应用
4. 添加 anthropic 配置到 `config.yaml`
5. 重新打开应用
6. 查看 Provider 下拉框

**预期结果：**
- OpenAI 和 Anthropic 都显示为可选中状态
- 其他 providers 显示但禁用

---

## 测试检查清单

- [ ] 场景 1：只配置一个 provider 时，下拉框正确显示
- [ ] 场景 2：配置多个 providers 时，下拉框正确显示
- [ ] 场景 3：缺少 API Key 时，正确显示禁用状态和原因
- [ ] 场景 4：使用未配置 provider 时，显示错误提示
- [ ] 场景 5：配置变更后，重启应用正确更新

---

## 已知问题

### Minor（次要问题）

1. **错误日志语言不一致**
   - `commitStore.ts` 中使用英文 `'Failed to load providers:'`
   - 其他地方使用中文
   - 影响：不影响功能，仅日志风格不一致
   - 优先级：低

2. **CSS 动画键名重复**
   - `CommitPanel.vue` 中定义了两个 `pulse` 动画
   - 影响：当前无实际影响，但可能导致混淆
   - 优先级：低

---

## 相关 Commit

- `ebc28b0` - 添加 ProviderInfo 模型类型
- `cef2e3b` - ConfigService 添加 GetConfiguredProviders 方法
- `bc9d5d9` - 修复 GetConfiguredProviders 的 BaseURL 检查逻辑
- `bb0ed69` - 添加 GetConfiguredProviders API 方法
- `8eef60a` - 增强 CommitService 错误处理
- `993034a` - 前端添加 ProviderInfo 类型定义
- `e9fafe5` - commitStore 添加 provider 列表加载功能
- `532e3de` - CommitPanel 改造 provider 下拉框为动态渲染
